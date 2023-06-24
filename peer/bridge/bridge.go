package bridge

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type BridgeService struct {
	relay  relayer.Relay
	tracer trace.Tracer
}

func NewBridgeService(relay relayer.Relay, tc trace.Tracer) *BridgeService {
	return &BridgeService{relay: relay, tracer: tc}
}

func (t *BridgeService) Ql(ctx context.Context, argType ql.BridgeArgs, replyType *ql.BridgeReply) error {
	call := ql.BridgeCall{}
	err := json.Unmarshal(argType.Data, &call)
	if err != nil {
		return err
	}
	ctx = propagation.TraceContext{}.Extract(ctx, call.Carrier)
	ctx, span := t.tracer.Start(ctx, "ql.method")
	defer span.End()
	log := relayer.DefaultLogger()
	log.InfofWithContext(ctx, "Received a Ql call, method: %s", call.Method)
	switch call.Method {
	case "saveEvent":
		var d nostr.Event
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		log.InfofWithContext(ctx, "Received a saveEvent call, event: %s", d.ID)
		return t.relay.Storage().SaveEvent(&d)
	case "queryEvents":
		var d nostr.Filter
		err := json.Unmarshal(call.Body, &d)
		if err != nil {
			return err
		}
		log.InfofWithContext(ctx, "Received a queryEvents call")
		events, err := t.relay.Storage().QueryEvents(&d)
		if err != nil {
			return err
		}
		b, err := json.Marshal(events)
		if err != nil {
			return err
		}
		replyType.Data = b
		log.InfofWithContext(ctx, "Sending a queryEvents reply")
		return nil
	default:
		log.InfofWithContext(ctx, "Received a call, method: %s", call.Method)
		return errors.New("method not found")
	}
}
