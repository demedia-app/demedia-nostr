package ping

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
)

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

	data := string(call.Body)
	log.Printf("Received a Ping call, message: %s\n", data)

	adds := strings.Split(data, ";")
	t.relay.Storage().SavePeer(adds[1], adds[0])

	replyType.Data = []byte("Pong")
	return nil
}
