package routedrpc

// A Handler exposes functions for handling incoming messages and address resolution
type Handler interface {
	// Should return true is _target_ can be handled by this node
	//
	// Multiple should be true if this address can be subject
	// to high availability setups
	HasTarget(target Address) (ok bool, multiple bool)

	// Handles calls for Cluster.Cast
	HandleCast(sender, target Address, message interface{}) error

	// Handles calls for Cluster.Call and Rpc.Go
	HandleCall(sender, target Address, message interface{}) (interface{}, error)
}
