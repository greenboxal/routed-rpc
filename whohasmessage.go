package routedrpc

type whoHasRequest struct {
	Sender  uint64
	Address Address
}

type whoHasReply struct {
	Who     uint64
	Address Address
}
