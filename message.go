package routedrpc

import (
	"encoding/gob"
)

type messageType int

const (
	_ messageType = iota
	castMessage
	callMessage
	replyMessage
	errorMessage
)

type message struct {
	XID             uint64
	Sender          Address
	Target          Address
	ForwardingCount int
	Type            messageType
	Message         []byte
}

func init() {
	gob.Register(whoHasRequest{})
	gob.Register(whoHasReply{})
	gob.Register(message{})
}
