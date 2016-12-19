package routedrpc

type Node interface {
	ID() interface{}
	Send(msg interface{}) error
}
