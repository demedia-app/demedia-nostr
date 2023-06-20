package ql

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
	SpanID        uint64
	TraceID       uint64
}
