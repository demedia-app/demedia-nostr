package bridge

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
)

type BridgeService struct {
	relay relayer.Relay
}

func NewBridgeService(relay relayer.Relay) *BridgeService {
	return &BridgeService{relay: relay}
}

func (t *BridgeService) Ql(_ context.Context, argType ql.BridgeArgs, replyType *ql.BridgeReply) error {
	call := ql.BridgeCall{}
	err := json.Unmarshal(argType.Data, &call)
	if err != nil {
		return err
	}

	switch call.Method {
	case "saveEvent":
		var d nostr.Event
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		return t.relay.Storage().SaveEvent(&d)
	default:
		return errors.New("method not found")
	}
}
