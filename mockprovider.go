package routedrpc

type MockHub struct {
	nodes map[int]*MockProvider
}

type MockProvider struct {
	id  int
	hub *MockHub
	rpc *RPC
}

func NewMockHub() *MockHub {
	return &MockHub{
		nodes: make(map[int]*MockProvider),
	}
}

func (h *MockHub) CreateClient(id int) *MockProvider {
	p := &MockProvider{
		id:  id,
		hub: h,
	}

	h.nodes[id] = p

	return p
}

func (m *MockProvider) Shutdown() error {
	return nil
}

func (m *MockProvider) ID() interface{} {
	return m.id
}

func (m *MockProvider) Self() Node {
	return m
}

func (m *MockProvider) Members() []Node {
	i := 0
	members := m.hub.nodes
	result := make([]Node, len(members))

	for _, v := range members {
		result[i] = v
		i++
	}

	return result
}

func (m *MockProvider) Broadcast(msg interface{}) error {
	for _, n := range m.hub.nodes {
		if n.id == m.id {
			continue
		}

		n.Send(msg)
	}

	return nil
}

func (m *MockProvider) SetRpc(rpc *RPC) {
	m.rpc = rpc
}

func (m *MockProvider) Send(msg interface{}) error {
	go m.rpc.ProcessRPCMessage(msg)

	return nil
}
