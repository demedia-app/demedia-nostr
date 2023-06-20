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
	DDCarrier     map[string]string
}
