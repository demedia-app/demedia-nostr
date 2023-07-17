package relayer

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
	"go.opentelemetry.io/otel/trace"
)

func DeleteEvent(relay Relay, evt nostr.Event, host host.Host, ctx context.Context, span trace.Span) error {
	store := relay.Storage()
	address := store.GetPeer(evt.PubKey)
	_, sandErr := ql.QlCall(host, ctx, evt, address, "BridgeService", "Ql", "deleteEvent", span)
	if sandErr != nil {
		return fmt.Errorf("error: failed to delete: %s", sandErr.Error())
	}
	return nil
}
