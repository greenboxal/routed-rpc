package routedrpc

type Handler interface {
	HasTarget(target Address) bool
	HandleCast(sender, target Address, message interface{})
	HandleCall(sender, target Address, message interface{}) (interface{}, error)
}