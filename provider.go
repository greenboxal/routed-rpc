package routedrpc

// Provider should implement network level details for the RPC system
type Provider interface {
	// Returns information about the current node
	Self() Node

	// Returns all nodes in the system (including itself)
	Members() []Node

	// Shutdowns the current node
	Shutdown() error

	// Sends a message to all nodes (excluding itself
	Broadcast(msg interface{}) error

	// Used to inject a reference to the Rpc instance attached to this provider
	SetRpc(rpc *RPC)
}
