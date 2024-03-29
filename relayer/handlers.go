package relayer

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/imroc/req/v3"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/nbd-wtf/go-nostr/nip42"
	"github.com/sithumonline/demedia-nostr/relayer/hashutil"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slices"
)

// TODO: consider moving these to Server as config params
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = pongWait / 2

	// Maximum message size allowed from peer.
	maxMessageSize = 512000
)

// TODO: consider moving these to Server as config params
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "handleWebsocket")
	span.SetAttributes(attribute.String("span_id", span.SpanContext().SpanID().String()))
	defer span.End()
	s.Log.InfofWithContext(ctx, "handling websocket request from %s", r.RemoteAddr)
	store := s.relay.Storage()
	advancedDeleter, _ := store.(AdvancedDeleter)
	advancedQuerier, _ := store.(AdvancedQuerier)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Log.ErrorfWithContext(ctx, "failed to upgrade websocket: %v", err)
		return
	}
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[conn] = struct{}{}
	ticker := time.NewTicker(pingPeriod)

	// NIP-42 challenge
	challenge := make([]byte, 8)
	rand.Read(challenge)

	ws := &WebSocket{
		conn:      conn,
		challenge: hex.EncodeToString(challenge),
	}
	s.Log.InfofWithContext(ctx, "challenge: %s", ws.challenge)
	// reader
	go func() {
		ctx, span := s.tracer.Start(ctx, "handleWebsocket.reader")
		span.SetAttributes(attribute.String("span_id", span.SpanContext().SpanID().String()))
		defer span.End()
		defer func() {
			ticker.Stop()
			s.clientsMu.Lock()
			if _, ok := s.clients[conn]; ok {
				conn.Close()
				delete(s.clients, conn)
				removeListener(ws)
			}
			s.clientsMu.Unlock()
		}()

		conn.SetReadLimit(maxMessageSize)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		// NIP-42 auth challenge
		if _, ok := s.relay.(Auther); ok {
			ws.WriteJSON([]interface{}{"AUTH", ws.challenge})
		}
		s.Log.InfofWithContext(ctx, "auth challenge sent")
		for {
			ctx, span := s.tracer.Start(ctx, "handleWebsocket.reader.for")
			span.SetAttributes(attribute.String("span_id", span.SpanContext().SpanID().String()))
			s.Log.InfofWithContext(ctx, "inside for loop and waiting for message")
			typ, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,        // 1001
					websocket.CloseNoStatusReceived, // 1005
					websocket.CloseAbnormalClosure,  // 1006
				) {
					s.Log.WarningfWithContext(ctx, "unexpected close error from %s: %v", r.Header.Get("X-Forwarded-For"), err)
				}
				break
			}

			if typ == websocket.PingMessage {
				ws.WriteMessage(websocket.PongMessage, nil)
				continue
			}

			go func(message []byte) {
				ctx, span := s.tracer.Start(ctx, "handleWebsocket.reader.for.go")
				span.SetAttributes(attribute.String("span_id", span.SpanContext().SpanID().String()))
				defer span.End()
				s.Log.InfofWithContext(ctx, "initializing go routine for message")
				var notice string
				defer func() {
					if notice != "" {
						ws.WriteJSON([]interface{}{"NOTICE", notice})
					}
				}()

				var request []json.RawMessage
				if err := json.Unmarshal(message, &request); err != nil {
					// stop silently
					return
				}

				if len(request) < 2 {
					notice = "request has less than 2 parameters"
					return
				}

				var typ string
				json.Unmarshal(request[0], &typ)

				switch typ {
				case "EVENT":
					// it's a new event
					var evt nostr.Event
					if err := json.Unmarshal(request[1], &evt); err != nil {
						notice = "failed to decode event: " + err.Error()
						return
					}

					// check serialization
					serialized := evt.Serialize()

					// assign ID
					hash := sha256.Sum256(serialized)
					evt.ID = hex.EncodeToString(hash[:])

					// check signature (requires the ID to be set)
					if ok, err := evt.CheckSignature(); err != nil {
						ws.WriteJSON([]interface{}{"OK", evt.ID, false, "error: failed to verify signature"})
						return
					} else if !ok && evt.Kind != 5 {
						ws.WriteJSON([]interface{}{"OK", evt.ID, false, "invalid: signature is invalid"})
						return
					}

					if evt.Kind == 5 {
						// event deletion -- nip09
						for _, tag := range evt.Tags {
							if len(tag) >= 2 && tag[0] == "e" {
								if advancedDeleter != nil {
									advancedDeleter.BeforeDelete(tag[1], evt.PubKey)
								}

								s.Log.InfofWithContext(ctx, "delete event id: %s, key: %s", tag[1], evt.PubKey)
								if s.host != nil {
									if err := DeleteEvent(s.relay, nostr.Event{ID: tag[1], PubKey: evt.PubKey}, s.host, ctx, span); err != nil {
										ws.WriteJSON([]interface{}{"OK", evt.ID, false, fmt.Sprintf("error: %s", err.Error())})
										return
									}
								} else {
									if err := store.DeleteEvent(tag[1], evt.PubKey); err != nil {
										ws.WriteJSON([]interface{}{"OK", evt.ID, false, fmt.Sprintf("error: %s", err.Error())})
										return
									}
								}

								if advancedDeleter != nil {
									advancedDeleter.AfterDelete(tag[1], evt.PubKey)
								}
							}
							if tag[0] == "audio" && s.ipfs != nil {
								urlChunks := strings.Split(tag[1], "/")
								if len(urlChunks) == 5 {
									s.Log.InfofWithContext(ctx, "delete cid: %s", urlChunks[4])
									if err = s.ipfs.DeleteFile(urlChunks[4]); err != nil {
										ws.WriteJSON([]interface{}{"OK", evt.ID, false, fmt.Sprintf("error: %s", err.Error())})
										return
									}
								}
							}
						}
						return
					}

					isEvtChanged := false
					if evt.Kind == 1 && (s.blob != nil || s.ipfs != nil) {
						for _, tag := range evt.Tags {
							if len(tag) != 2 {
								continue
							}
							if tag[0] != "audio" {
								continue
							}

							s.Log.InfofWithContext(ctx, "media tag: %s url: %s", tag[0], tag[1])

							reqClient := req.C()        // Use C() to create a client.
							resp, err := reqClient.R(). // Use R() to create a request.
											Get(tag[1])
							if err != nil {
								s.Log.ErrorfWithContext(ctx, "failed to get file from url: %v", err)
								continue
							}
							defer resp.Body.Close()

							fileBytes, err := io.ReadAll(resp.Body)
							if err != nil {
								s.Log.ErrorfWithContext(ctx, "failed to h io read: %v", err)
								continue
							}

							var u = ""
							if s.blob != nil {
								fileSplit := strings.Split(tag[1], "/")
								fileName := fmt.Sprintf("hub_%s", fileSplit[len(fileSplit)-1])
								err = s.blob.SaveFile(fileName, fileBytes)
								if err != nil {
									s.Log.ErrorfWithContext(ctx, "failed to h save file to blob: %v", err)
									continue
								}

								u, err = s.blob.GetFileURL(fileName)
								if err != nil {
									s.Log.ErrorfWithContext(ctx, "failed to h get url: %v", err)
									continue
								}
							} else if s.ipfs != nil {
								u, err = s.ipfs.UploadFile(fileBytes)
								if err != nil {
									s.Log.ErrorfWithContext(ctx, "failed to h save file to ipfs: %v", err)
									continue
								}
								s.Log.InfofWithContext(ctx, "ipfs file saved url: %s", u)
							}

							tag[1] = u
							isEvtChanged = true
							s.Log.InfofWithContext(ctx, "audio url changed to: %s", u)
						}
					}

					for _, tag := range evt.Tags {
						if len(tag) >= 2 {
							if tag[0] == "p" || tag[0] == "e" {
								isEvtChanged = true
							}
						}
					}
					s.Log.InfofWithContext(ctx, fmt.Sprintf("before sig content evt.kind: %d isEcdsaPvtKey: %t isEvtChanged: %t len.evt.Tags: %d", evt.Kind, s.ecdsaPvtKey != nil, isEvtChanged, len(evt.Tags)))
					isHashAdded := false
					if evt.Kind == 1 && s.ecdsaPvtKey != nil && (isEvtChanged || len(evt.Tags) == 0) {
						sig, err := hashutil.GetSing(hashutil.GetSha256([]byte(evt.Content)), s.ecdsaPvtKey)
						if err != nil {
							s.Log.ErrorfWithContext(ctx, "failed to calculate sig: %v", err)
						} else {
							evt.Tags = evt.Tags.AppendUnique([]string{"hash", sig, "true"})
							isHashAdded = true
						}
					}

					p := ""
					if isEvtChanged || isHashAdded {
						p = hashutil.StringifyEvent(&evt)
					}

					if p != "" {
						// gen hash for event as audio url changed
						bs := hashutil.GetSha256([]byte(p))
						evt.ID = fmt.Sprintf("%x", bs)
					}

					if s.host != nil {
						s.Log.InfofWithContext(ctx, "initializing send event to peer")
						ok, message := SendEvent(s.relay, evt, s.host, ctx, span)
						s.Log.InfofWithContext(ctx, "completed send event to peer")
						ws.WriteJSON([]interface{}{"OK", evt.ID, ok, message})
					} else {
						ok, message := AddEvent(s.relay, evt)
						ws.WriteJSON([]interface{}{"OK", evt.ID, ok, message})
					}

				case "REQ":
					var id string
					json.Unmarshal(request[1], &id)
					if id == "" {
						notice = "REQ has no <id>"
						return
					}

					filters := make(nostr.Filters, len(request)-2)
					for i, filterReq := range request[2:] {
						if err := json.Unmarshal(
							filterReq,
							&filters[i],
						); err != nil {
							notice = "failed to decode filter"
							return
						}

						filter := &filters[i]

						// prevent kind-4 events from being returned to unauthed users,
						//   only when authentication is a thing
						if _, ok := s.relay.(Auther); ok {
							if slices.Contains(filter.Kinds, 4) {
								senders := filter.Authors
								receivers, _ := filter.Tags["p"]
								switch {
								case ws.authed == "":
									// not authenticated
									notice = "restricted: this relay does not serve kind-4 to unauthenticated users, does your client implement NIP-42?"
									return
								case len(senders) == 1 && len(receivers) < 2 && (senders[0] == ws.authed):
									// allowed filter: ws.authed is sole sender (filter specifies one or all receivers)
								case len(receivers) == 1 && len(senders) < 2 && (receivers[0] == ws.authed):
									// allowed filter: ws.authed is sole receiver (filter specifies one or all senders)
								default:
									// restricted filter: do not return any events,
									//   even if other elements in filters array were not restricted).
									//   client should know better.
									notice = "restricted: authenticated user does not have authorization for requested filters."
									return
								}
							}
						}

						if advancedQuerier != nil {
							advancedQuerier.BeforeQuery(filter)
						}

						var events []nostr.Event
						if s.host != nil {
							senders := filter.Authors
							receivers, _ := filter.Tags["p"]
							var pubKey string
							if len(senders) != 0 {
								pubKey = senders[0]
							} else if len(receivers) != 0 {
								pubKey = receivers[0]
							}
							s.Log.InfofWithContext(ctx, "fetching events from peer ID: %s", id)
							events, err = FetchEvent(pubKey, filter, s.relay, s.host, ctx, span)
							s.Log.InfofWithContext(ctx, "completed fetching events from peer ID: %s", id)
						} else {
							events, err = store.QueryEvents(filter)
						}

						if err != nil {
							s.Log.Errorf("store: %v", err)
							continue
						}

						if advancedQuerier != nil {
							advancedQuerier.AfterQuery(events, filter)
						}

						// this block should not trigger if the SQL query accounts for filter.Limit
						// other implementations may be broken, and this ensures the client
						// won't be bombarded.
						if filter.Limit > 0 && len(events) > filter.Limit {
							events = events[0:filter.Limit]
						}

						for _, event := range events {
							if event.Kind == 1 && s.ecdsaPvtKey != nil && len(event.Tags) > 0 {
								for _, tag := range event.Tags {
									if len(tag) >= 2 {
										if tag[0] == "hash" {
											b, err := hashutil.GetVerification(tag[1], hashutil.GetSha256([]byte(event.Content)), &s.ecdsaPvtKey.PublicKey)
											if err != nil {
												s.Log.ErrorfWithContext(ctx, "failed to verify sig: %v", err)
											} else {
												tag[2] = strconv.FormatBool(b)
												bs := hashutil.GetSha256([]byte(hashutil.StringifyEvent(&event)))
												event.ID = fmt.Sprintf("%x", bs)
											}
										}
									}
								}
							}

							s.Log.InfofWithContext(ctx, "sending EVENT ID: %s", id)
							ws.WriteJSON([]interface{}{"EVENT", id, event})
						}
					}
					// moved EOSE out of for loop.
					// otherwise subscriptions may be cancelled too early
					s.Log.InfofWithContext(ctx, "sending EOSE ID: %s", id)
					ws.WriteJSON([]interface{}{"EOSE", id})
					setListener(id, ws, filters)
				case "CLOSE":
					var id string
					json.Unmarshal(request[1], &id)
					if id == "" {
						notice = "CLOSE has no <id>"
						return
					}

					removeListenerId(ws, id)
				case "AUTH":
					if auther, ok := s.relay.(Auther); ok {
						var evt nostr.Event
						if err := json.Unmarshal(request[1], &evt); err != nil {
							notice = "failed to decode auth event: " + err.Error()
							return
						}
						if pubkey, ok := nip42.ValidateAuthEvent(&evt, ws.challenge, auther.ServiceURL()); ok {
							ws.authed = pubkey
							ws.WriteJSON([]interface{}{"OK", evt.ID, true, "authentication success"})
						} else {
							ws.WriteJSON([]interface{}{"OK", evt.ID, false, "error: failed to authenticate"})
						}
					}
				default:
					if cwh, ok := s.relay.(CustomWebSocketHandler); ok {
						cwh.HandleUnknownType(ws, typ, request)
					} else {
						notice = "unknown message type " + typ
					}
				}
			}(message)
			span.End()
		}
	}()

	// writer
	go func() {
		defer func() {
			ticker.Stop()
			conn.Close()
		}()

		for {
			select {
			case <-ticker.C:
				err := ws.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					s.Log.ErrorfWithContext(ctx, "error writing ping: %v; closing websocket %v", err)
					return
				}
			}
		}
	}()
}

func (s *Server) handleNIP11(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	supportedNIPs := []int{9, 11, 12, 15, 16, 20}
	if _, ok := s.relay.(Auther); ok {
		supportedNIPs = append(supportedNIPs, 42)
	}

	info := nip11.RelayInformationDocument{
		Name:          s.relay.Name(),
		Description:   "relay powered by the relayer framework",
		PubKey:        "~",
		Contact:       "~",
		SupportedNIPs: supportedNIPs,
		Software:      "https://github.com/sithumonline/demedia-nostr/relayer",
		Version:       "~",
	}

	if ifmer, ok := s.relay.(Informationer); ok {
		info = ifmer.GetNIP11InformationDocument()
	}

	json.NewEncoder(w).Encode(info)
}
