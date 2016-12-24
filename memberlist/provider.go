package memberlist

import (
	"encoding/gob"
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

	if err := provider.sanitizeConfig(); err != nil {
		return nil, err
	}

	if err := provider.initializeRPCServer(); err != nil {
		return nil, err
	}

	if err := provider.initializeMemberlist(); err != nil {
		return nil, err
	}

	// Initialize local node
	local := provider.mb.LocalNode()
	provider.local = newNode(provider, local.Name)
	provider.local.update(local)
	provider.members[local.Name] = provider.local

	return provider, nil
}

func (m *Provider) sanitizeConfig() error {
	c := mb.DefaultLANConfig()

	whispBind, err := net.ResolveTCPAddr("tcp", m.config.WhispBindEndpoint)

	if err != nil {
		return err
	}

	whispAdvertise, err := net.ResolveTCPAddr("tcp", m.config.WhispAdvertiseEndpoint)

	if err != nil {
		return err
	}

	rpcBind, err := net.ResolveTCPAddr("tcp", m.config.WhispBindEndpoint)

	if err != nil {
		return err
	}

	rpcAdvertise, err := net.ResolveTCPAddr("tcp", m.config.WhispAdvertiseEndpoint)

	if err != nil {
		return err
	}

	if whispBind.IP == nil {
		whispBind.IP = net.ParseIP("0.0.0.0")
		m.config.WhispBindEndpoint = whispBind.String()
	}

	if rpcBind.IP == nil {
		rpcBind.IP = net.ParseIP("0.0.0.0")
		m.config.RPCBindEndpoint = rpcBind.String()
	}

	if whispAdvertise.IP == nil {
		whispAdvertise.IP = firstPublicAddress()
		m.config.WhispAdvertiseEndpoint = whispAdvertise.String()
	}

	if rpcAdvertise.IP == nil {
		rpcAdvertise.IP = firstPublicAddress()
		m.config.RPCAdvertiseEndpoint = rpcAdvertise.String()
	}

	return nil
}

func (m *Provider) initializeRPCServer() error {
	listener, err := net.Listen("tcp", m.config.RPCBindEndpoint)

	if err != nil {
		return err
	}

	m.listener = listener

	go m.handleListener()

	return nil
}

func (m *Provider) initializeMemberlist() error {
	c := mb.DefaultLANConfig()

	bind, err := net.ResolveTCPAddr("tcp", m.config.WhispBindEndpoint)

	if err != nil {
		return err
	}

	advertise, err := net.ResolveTCPAddr("tcp", m.config.WhispAdvertiseEndpoint)

	if err != nil {
		return err
	}

	c.Name = m.config.Name
	c.BindAddr = bind.IP.String()
	c.BindPort = bind.Port
	c.AdvertiseAddr = advertise.IP.String()
	c.AdvertisePort = advertise.Port
	c.Delegate = &delegate{m}
	c.Events = &delegate{m}

	mb, err := mb.Create(c)

	if err != nil {
		return err
	}

	m.mb = mb

	return nil
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

func firstPublicAddress() net.IP {
	addresses, err := net.InterfaceAddrs()

	if err != nil {
		return nil
	}

	for _, rawAddr := range addresses {
		var ip net.IP

		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}

		if ip.To4() == nil {
			continue
		}

		if !mb.IsPrivateIP(ip.String()) {
			continue
		}

		return ip
	}

	return nil
}
