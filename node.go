package routedrpc

type Node interface {
	ID() uint64
	Send(data interface{}) error
}
