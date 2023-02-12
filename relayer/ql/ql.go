package utility

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/sithumonline/demedia-poc/core/config"
	"github.com/sithumonline/demedia-poc/peer/transact/bridge"
)

func QlCall(
	h host.Host,
	ctx context.Context,
	input interface{},
	peerAddr string,
	serviceName string,
	serviceMethod string,
	method string,
) (
	bridge.BridgeReply,
	error,
) {
	body, err := json.Marshal(input)
	if err != nil {
		return bridge.BridgeReply{}, fmt.Errorf("QlCall, json marshal input: %w", err)
	}

	ma, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return bridge.BridgeReply{}, err
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return bridge.BridgeReply{}, err
	}

	err = h.Connect(ctx, *peerInfo)
	if err != nil {
		return bridge.BridgeReply{}, fmt.Errorf("QlCall, host connection: \n%w", err)
		//return bridge.BridgeReply{}, err
	}
	rpcClient := rpc.NewClient(h, config.ProtocolId)

	args, err := json.Marshal(bridge.BridgeCall{Method: method, Body: body})
	if err != nil {
		return bridge.BridgeReply{}, fmt.Errorf("QlCall, json marshal BridgeCall: %w", err)
	}

	var reply bridge.BridgeReply

	err = rpcClient.Call(
		peerInfo.ID,
		serviceName,
		serviceMethod,
		bridge.BridgeArgs{Data: args},
		&reply,
	)
	if err != nil {
		return bridge.BridgeReply{}, fmt.Errorf("QlCall, rpcClient call: %w", err)
	}
	return reply, nil
}
