package routedrpc

// Node represents a node in the RPC system
type Node interface {
	// Returns the node ID
	ID() interface{}

	// Sends a message for a node
	Send(msg interface{}) error

	// Indicates that this node is online
	Online() bool
}
