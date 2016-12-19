package routedrpc

// A Handler exposes functions for handling incoming messages and address resolution
type Handler interface {
	// Should return true is _target_ can be handled by this node
	HasTarget(target Address) bool

	// Handles calls for Rpc.Cast
	HandleCast(sender, target Address, message interface{}) error

	// Handles calls for Rpc.Call and Rpc.Go
	HandleCall(sender, target Address, message interface{}) (interface{}, error)
}
