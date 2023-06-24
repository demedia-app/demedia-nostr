package ql

import "go.opentelemetry.io/otel/propagation"

type BridgeArgs struct {
	Data []byte
}
type BridgeReply struct {
	Data []byte
}

type BridgeCall struct {
	Body      []byte
	Method    string
	DDCarrier propagation.MapCarrier
}
