package routedrpc

type StandardHandler struct {
	handlers map[Address]Handler
}

func NewStandardHandler() *StandardHandler {
	return &StandardHandler{
		handlers: make(map[Address]Handler),
	}
}

func (n *StandardHandler) Register(address Address, handler Handler) {
	n.handlers[address] = handler
}

func (n *StandardHandler) Deregister(address Address) {
	delete(n.handlers, address)
}

func (n *StandardHandler) HasTarget(target Address) bool {
	handler, found := n.handlers[target]

	if !found {
		return false
	}

	return handler.HasTarget(target)
}

func (n *StandardHandler) HandleCast(sender, target Address, message interface{}) error {
	handler, found := n.handlers[target]

	if !found {
		return AddressNotFound{}
	}

	return handler.HandleCast(sender, target, message)
}

func (n *StandardHandler) HandleCall(sender, target Address, message interface{}) (interface{}, error) {
	handler, found := n.handlers[target]

	if !found {
		return nil, AddressNotFound{}
	}

	return handler.HandleCall(sender, target, message)
}
