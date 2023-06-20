package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type BridgeService struct {
	relay relayer.Relay
}

func NewBridgeService(relay relayer.Relay) *BridgeService {
	return &BridgeService{relay: relay}
}

func (t *BridgeService) Ql(ctx context.Context, argType ql.BridgeArgs, replyType *ql.BridgeReply) error {
	call := ql.BridgeCall{}
	err := json.Unmarshal(argType.Data, &call)
	if err != nil {
		return err
	}
	sctx, err := tracer.Extract(call.DDCarrier)
	if err != nil {
		return err
	}
	span := tracer.StartSpan("ql.method", tracer.ChildOf(sctx))
	defer span.Finish()
	log := relayer.DefaultLogger(t.relay.Name(), call.CorrelationId)
	log.InfofWithContext(span.Context(), "Received a Ql call, method: %s", call.Method)
	switch call.Method {
	case "saveEvent":
		var d nostr.Event
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		log.InfofWithContext(span.Context(), "Received a saveEvent call, event: %s", d.ID)
		return t.relay.Storage().SaveEvent(&d)
	case "queryEvents":
		var d nostr.Filter
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		log.InfofWithContext(span.Context(), "Received a queryEvents call")
		events, err := t.relay.Storage().QueryEvents(&d)
		if err != nil {
			return err
		}
		b, err := json.Marshal(events)
		if err != nil {
			return err
		}
		replyType.Data = b
		log.InfofWithContext(span.Context(), "Sending a queryEvents reply")
		return nil
	default:
		log.InfofWithContext(span.Context(), "Received a call, method: %s", call.Method)
		return errors.New("method not found")
	}
}
