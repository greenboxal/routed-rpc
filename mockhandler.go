package routedrpc

type MockHandler struct {
	Name string

	ErrorToReturnOnCast error
	ErrorToReturnOnCall error

	LastCastSender  Address
	LastCastTarget  Address
	LastCastMessage interface{}
	CastSignal      chan interface{}

	LastCallSender  Address
	LastCallTarget  Address
	LastCallMessage interface{}
}

func NewMockHandler(name string) *MockHandler {
	return &MockHandler{
		Name:       name,
		CastSignal: make(chan interface{}, 1),
	}
}

func (n *MockHandler) WaitCast() {
	<-n.CastSignal
}

func (n *MockHandler) HasTarget(target Address) (bool, bool) {
	return n.Name == target, true
}

func (n *MockHandler) HandleCast(sender, target Address, message interface{}) error {
	n.LastCastSender = sender
	n.LastCastTarget = target
	n.LastCastMessage = message

	n.CastSignal <- n

	return n.ErrorToReturnOnCast
}

func (n *MockHandler) HandleCall(sender, target Address, message interface{}) (interface{}, error) {
	n.LastCallSender = sender
	n.LastCallTarget = target
	n.LastCallMessage = message

	if n.ErrorToReturnOnCall != nil {
		return nil, n.ErrorToReturnOnCall
	}

	return n.Name, nil
}
