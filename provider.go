package routedrpc

// Provider should implement network level details for the cluster
type Provider interface {
	// Returns information about the current node
	Self() Node

	// Returns all nodes in the cluster (including itself)
	Members() []Node

	// Returns node by ID
	GetMember(id interface{}) (Node, bool)

	// Shutdowns the current node
	Shutdown() error

	// Sends a message to all nodes (excluding itself
	Broadcast(msg interface{}) error

	// Used to inject a reference to the cluster instance attached to this provider
	SetCluster(cluster *Cluster)
}
