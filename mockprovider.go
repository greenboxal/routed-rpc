package routedrpc

type MockHub struct {
	nodes map[int]*MockProvider
}

type MockProvider struct {
	id      int
	hub     *MockHub
	cluster *Cluster
	online  bool
}

func NewMockHub() *MockHub {
	return &MockHub{
		nodes: make(map[int]*MockProvider),
	}
}

func (h *MockHub) CreateClient(id int) *MockProvider {
	p := &MockProvider{
		id:     id,
		hub:    h,
		online: true,
	}

	h.nodes[id] = p

	return p
}

func (m *MockProvider) GetMember(id interface{}) (Node, bool) {
	n, found := m.hub.nodes[id.(int)]

	return n, found
}

func (m *MockProvider) Shutdown() error {
	m.online = false
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

func (m *MockProvider) Online() bool {
	return m.online
}

func (m *MockProvider) SetOnline(online bool) {
	m.online = online
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

func (m *MockProvider) SetCluster(cluster *Cluster) {
	m.cluster = cluster
}

func (m *MockProvider) Send(msg interface{}) error {
	go m.cluster.ProcessRPCMessage(msg)

	return nil
}
