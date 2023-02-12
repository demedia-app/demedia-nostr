package ql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
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
	BridgeReply,
	error,
) {
	body, err := json.Marshal(input)
	if err != nil {
		return BridgeReply{}, fmt.Errorf("QlCall, json marshal input: %w", err)
	}

	ma, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return BridgeReply{}, err
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return BridgeReply{}, err
	}

	err = h.Connect(ctx, *peerInfo)
	if err != nil {
		return BridgeReply{}, fmt.Errorf("QlCall, host connection: \n%w", err)
	}
	rpcClient := rpc.NewClient(h, "/p2p/1.0.0")

	args, err := json.Marshal(BridgeCall{Method: method, Body: body})
	if err != nil {
		return BridgeReply{}, fmt.Errorf("QlCall, json marshal BridgeCall: %w", err)
	}

	var reply BridgeReply

	err = rpcClient.Call(
		peerInfo.ID,
		serviceName,
		serviceMethod,
		BridgeArgs{Data: args},
		&reply,
	)
	if err != nil {
		return BridgeReply{}, fmt.Errorf("QlCall, rpcClient call: %w", err)
	}
	return reply, nil
}
