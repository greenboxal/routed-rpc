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
	SenderID        interface{}
	Sender          Address
	Target          Address
	ForwardingCount int
	Type            messageType
	Message         interface{}
}

func init() {
	gob.Register(whoHasRequest{})
	gob.Register(whoHasReply{})
	gob.Register(message{})
}
