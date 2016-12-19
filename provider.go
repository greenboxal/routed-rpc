package routedrpc

type Provider interface {
	Self() Node
	Members() []Node
	Shutdown() error
	Broadcast(msg interface{}) error
	SetRpc(rpc *Rpc)
}
