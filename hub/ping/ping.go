package ping

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
)

//type PingArgs struct {
//	Data []byte
//}
//type PingReply struct {
//	Data []byte
//}
//type PeerInfo struct {
//	Address    string
//	LastUpdate time.Time
//}

type PingService struct {
	relay relayer.Relay
}

func NewPingService(relay relayer.Relay) *PingService {
	return &PingService{relay: relay}
}

func (t *PingService) Ping(_ context.Context, argType ql.BridgeArgs, replyType *ql.BridgeReply) error {
	call := ql.BridgeCall{}
	err := json.Unmarshal(argType.Data, &call)
	if err != nil {
		return err
	}
	data := strings.Trim(string(call.Body), "\\\"")
	log.Printf("Received a Ping call, message: %s\n", data)

	adds := strings.Split(data, "/")
	//t.db[fmt.Sprintf("%s", adds[6])] = PeerInfo{
	//	Address:    fmt.Sprintf("%s", data),
	//	LastUpdate: time.Now(),
	//}
	t.relay.Storage().SavePeer(fmt.Sprintf("%s", data), fmt.Sprintf("%s", adds[6]))

	replyType.Data = []byte("Pong")
	return nil
}
