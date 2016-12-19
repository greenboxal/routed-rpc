package routedrpc

type whoHasRequest struct {
	Sender  string
	Address Address
}

type whoHasReply struct {
	Who     string
	Address Address
}
