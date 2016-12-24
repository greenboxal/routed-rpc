package memberlist

import (
	"encoding/gob"
	"fmt"
	"net"
	"sync"

	"github.com/greenboxal/routed-rpc"
	mb "github.com/hashicorp/memberlist"
)

type Provider struct {
	config   *Config
	mutex    sync.RWMutex
	rpc      *routedrpc.RPC
	mb       *mb.Memberlist
	local    *node
	members  map[string]*node
	listener net.Listener
}

func Create(config *Config) (*Provider, error) {
	provider := &Provider{
		members: make(map[string]*node),
		config:  config,
	}

	if err := provider.initializeRPCServer(); err != nil {
		return nil, err
	}

	if err := provider.initializeMemberlist(); err != nil {
		return nil, err
	}

	// Initialize local node
	result.local = newNode(result, memberlist.LocalNode().Name)
	result.local.update(memberlist.LocalNode())
	result.members = append(result.memers, result.local)

	return provider, nil
}

func (m *Provider) initializeRPCServer() error {
	result.listener, err = net.Listen("tcp", m.config.RpcBindEndpoint)

	if err != nil {
		return nil, err
	}

	go result.handleListener()

	return result, nil
}

func (m *Provider) initializeMemberlist() error {
	c := mb.DefaultLANConfig()

	bind := net.ResolveTCPAddr("tcp", m.config.WhispBindEndpoint)
	advertise := net.ResolveTCPAddr("tcp", m.config.WhispAdvertiseEndpoint)

	c.Name = config.Name
	c.BindAddr = bind.IP.String()
	c.BindPort = bind.Port
	c.AdvertiseAddr = advertise.IP.String()
	c.AdvertisePort = advertise.Port
	c.Delegate = delegate{m}
	c.Events = delegate{m}

	mb, err := mb.Create(c)

	if err != nil {
		return nil, err
	}

	result.mb = mb
}

// Join joins the current node into a cluster
func (m *Provider) Join(others []string) (int, error) {
	return m.mb.Join(others)
}

// Self returns the current node
func (m *Provider) Self() routedrpc.Node {
	return m.local
}

// Members returns all members in the cluster including itself
func (m *Provider) Members() []routedrpc.Node {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	i := 0
	result := make([]routedrpc.Node, len(m.members))

	for _, v := range m.members {
		result[i] = v
		i++
	}

	return result
}

// Returns a member by ID
func (m *Provider) GetMember(id interface{}) (routedrpc.Node, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	node, found := m.members[id.(string)]

	return node, found
}

// Broadcast sends a message to all members
func (m *Provider) Broadcast(msg interface{}) error {
	data, err := encode(msg)

	if err != nil {
		return err
	}

	for _, v := range m.members {
		v.sendRaw(data)
	}

	return err
}

// SetRPC saves the reference to the attached RPC instance
func (m *Provider) SetRPC(rpc *routedrpc.RPC) {
	m.rpc = rpc
}

func (m *Provider) Shutdown() error {
	if err := m.mb.Shutdown(); err != nil {
		return err
	}

	return m.listener.Close()
}

func (m *Provider) removeNode(n *mb.Node) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, found := m.members[n.Name]

	if !found {
		return
	}

	item.offline = true
	delete(m.members, n.Name)
}

func (m *Provider) updateNode(n *mb.Node) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, found := m.members[n.Name]

	if !found {
		item = newNode(m, n.Name)
		m.members[n.Name] = item
	}

	item.update(n)
}

func (m *Provider) handleListener() {
	for true {
		conn, err := m.listener.Accept()

		if err != nil {
			return
		}

		go m.handleConnection(conn)
	}
}

func (m *Provider) handleConnection(conn net.Conn) {
	decoder := gob.NewDecoder(conn)

	for true {
		var msg interface{}

		err := decoder.Decode(&msg)

		if err != nil {
			conn.Close()
			break
		}

		m.rpc.ProcessRPCMessage(msg)
	}
}
