package relayer

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func SendEvent(relay Relay, evt nostr.Event, host host.Host, correlationId string, ctx context.Context, span tracer.Span) (accepted bool, message string) {
	store := relay.Storage()

	if !relay.AcceptEvent(&evt) {
		return false, "blocked: event blocked by relay"
	}

	if 20000 <= evt.Kind && evt.Kind < 30000 {
		// do not store ephemeral events
	} else {
		address := store.GetPeer(evt.PubKey)
		_, sandErr := ql.QlCall(host, ctx, evt, address, "BridgeService", "Ql", "saveEvent", correlationId, span)
		if sandErr != nil {
			return false, fmt.Sprintf("error: failed to sand: %s", sandErr.Error())
		}
	}

	notifyListeners(&evt)

	return true, ""
}
