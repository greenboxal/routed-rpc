package routedrpc

// MuxHandler multiplexes calls to other handlers based on the target address
type MuxHandler struct {
	handlers map[Address]Handler
}

// NewMuxHandler creates a new _MuxHandler_
func NewMuxHandler() *MuxHandler {
	return &MuxHandler{
		handlers: make(map[Address]Handler),
	}
}

// Register a handler for an address
func (n *MuxHandler) Register(address Address, handler Handler) {
	n.handlers[address] = handler
}

// Deregister a handler
func (n *MuxHandler) Deregister(address Address) {
	delete(n.handlers, address)
}

// HasTarget returns if this target address can be processed in this node
func (n *MuxHandler) HasTarget(target Address) (bool, bool) {
	handler, found := n.handlers[target]

	if !found {
		return false, false
	}

	return handler.HasTarget(target)
}

// HandleCast forwards the call to the correct handler based on the target address
func (n *MuxHandler) HandleCast(sender, target Address, message interface{}) error {
	handler, found := n.handlers[target]

	if !found {
		return AddressNotFound{}
	}

	return handler.HandleCast(sender, target, message)
}

// HandleCall forwards the call to the correct handler based on the target address
func (n *MuxHandler) HandleCall(sender, target Address, message interface{}) (interface{}, error) {
	handler, found := n.handlers[target]

	if !found {
		return nil, AddressNotFound{}
	}

	return handler.HandleCall(sender, target, message)
}
