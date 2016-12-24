package routedrpc

type whoHasRequest struct {
	Sender  interface{}
	Address Address
}

type whoHasReply struct {
	Who      interface{}
	Address  Address
	Multiple bool
}
