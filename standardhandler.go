package routedrpc

// StandardHandler multiplexes calls to other handlers based on the target address
type StandardHandler struct {
	handlers map[Address]Handler
}

// NewStandardHandler creates a new _StandardHandler_
func NewStandardHandler() *StandardHandler {
	return &StandardHandler{
		handlers: make(map[Address]Handler),
	}
}

// Register a handler for an address
func (n *StandardHandler) Register(address Address, handler Handler) {
	n.handlers[address] = handler
}

// Deregister a handler
func (n *StandardHandler) Deregister(address Address) {
	delete(n.handlers, address)
}

// HasTarget returns if this target address can be processed in this node
func (n *StandardHandler) HasTarget(target Address) (bool, bool) {
	handler, found := n.handlers[target]

	if !found {
		return false, false
	}

	return handler.HasTarget(target)
}

// HandleCast forwards the call to the correct handler based on the target address
func (n *StandardHandler) HandleCast(sender, target Address, message interface{}) error {
	handler, found := n.handlers[target]

	if !found {
		return AddressNotFound{}
	}

	return handler.HandleCast(sender, target, message)
}

// HandleCall forwards the call to the correct handler based on the target address
func (n *StandardHandler) HandleCall(sender, target Address, message interface{}) (interface{}, error) {
	handler, found := n.handlers[target]

	if !found {
		return nil, AddressNotFound{}
	}

	return handler.HandleCall(sender, target, message)
}
