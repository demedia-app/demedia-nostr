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
	log := relayer.DefaultLogger(t.relay.Name(), call.CorrelationId)
	log.Infof("Received a Ql call, method: %s\n", call.Method)
	switch call.Method {
	case "saveEvent":
		var d nostr.Event
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		log.Infof("Received a saveEvent call, event: %s\n", d.ID)
		return t.relay.Storage().SaveEvent(&d)
	case "queryEvents":
		var d nostr.Filter
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		log.Infof("Received a queryEvents call")
		events, err := t.relay.Storage().QueryEvents(&d)
		if err != nil {
			return err
		}
		b, err := json.Marshal(events)
		if err != nil {
			return err
		}
		replyType.Data = b
		log.Infof("Sending a queryEvents reply")
		return nil
	default:
		log.Infof("Received a call, method: %s\n", call.Method)
		return errors.New("method not found")
	}
}
