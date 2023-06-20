package ql

import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

type BridgeArgs struct {
	Data []byte
}
type BridgeReply struct {
	Data []byte
}

type BridgeCall struct {
	Body          []byte
	Method        string
	CorrelationId string
	DDCarrier     tracer.TextMapCarrier
}
