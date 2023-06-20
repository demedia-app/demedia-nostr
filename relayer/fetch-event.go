package relayer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
)

func FetchEvent(pubKey string, filter *nostr.Filter, relay Relay, host host.Host, correlationId string, ctx context.Context) (events []nostr.Event, err error) {
	address := relay.Storage().GetPeer(pubKey)
	reply, sandErr := ql.QlCall(host, ctx, filter, address, "BridgeService", "Ql", "queryEvents", correlationId)
	if sandErr != nil {
		return nil, fmt.Errorf("error: failed to fetch: %s", sandErr.Error())
	}

	var d []nostr.Event
	err = json.Unmarshal(reply.Data, &d)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reply data: %v", err)
	}

	return d, nil
}
