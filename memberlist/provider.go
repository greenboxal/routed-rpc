package memberlist

import (
	"encoding/gob"
	"fmt"
	"net"
	"sync"

	"github.com/greenboxal/routed-rpc"
	mb "github.com/hashicorp/memberlist"
)

type Memberlist struct {
	config   *Config
	mutex    sync.RWMutex
	rpc      *routedrpc.RPC
	mb       *mb.Memberlist
	local    *node
	members  map[string]*node
	listener net.Listener
}

func Create(config *Config) (*Memberlist, error) {
	result := &Memberlist{
		members: make(map[string]*node),
		config:  config,
	}

	c := mb.DefaultLANConfig()

	c.Name = config.Name
	c.BindAddr = config.BindAddr
	c.BindPort = config.WhispBindPort
	c.AdvertiseAddr = config.AdvertiseAddr
	c.AdvertisePort = config.WhispBindPort
	c.Delegate = result
	c.Events = result

	memberlist, err := mb.Create(c)

	if err != nil {
		return nil, err
	}

	result.mb = memberlist

	result.local = newNode(result, memberlist.LocalNode().Name)
	result.local.update(memberlist.LocalNode())
	result.members = append(result.memers, result.local)

	result.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", config.BindAddr, config.RpcBindPort))

	if err != nil {
		return nil, err
	}

	go result.handleListener()

	return result, nil
}

func (m *Memberlist) Join(others []string) (int, error) {
	return m.mb.Join(others)
}

func (m *Memberlist) Self() routedrpc.Node {
	return m.local
}

func (m *Memberlist) Members() []routedrpc.Node {
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

func (m *Memberlist) GetMember(id interface{}) (routedrpc.Node, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	node, found := m.members[id.(string)]

	return node, found
}

func (m *Memberlist) Broadcast(msg interface{}) error {
	data, err := encode(msg)

	if err != nil {
		return err
	}

	for _, v := range m.members {
		v.sendRaw(data)
	}

	return err
}

func (m *Memberlist) SetRpc(rpc *routedrpc.RPC) {
	m.rpc = rpc
}

func (m *Memberlist) Shutdown() error {
	if err := m.mb.Shutdown(); err != nil {
		return err
	}

	return m.listener.Close()
}

func (m *Memberlist) NodeMeta(limit int) []byte {
	addr := net.TCPAddr{
		IP:   net.ParseIP(m.config.AdvertiseAddr),
		Port: m.config.RpcAdvertisePort,
	}

	data, err := encode(addr)

	if err != nil {
		return nil
	}

	return data
}

func (m *Memberlist) NotifyMsg(data []byte) {
	msg, err := decode(data)

	if err != nil {
		return
	}

	m.rpc.ProcessRPCMessage(msg)
}

func (m *Memberlist) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

func (m *Memberlist) LocalState(join bool) []byte {
	return []byte{}
}

func (m *Memberlist) MergeRemoteState(buf []byte, join bool) {
}

func (m *Memberlist) NotifyJoin(n *mb.Node) {
	m.NotifyUpdate(n)
}

func (m *Memberlist) NotifyLeave(n *mb.Node) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, found := m.members[n.Name]

	if !found {
		return
	}

	item.offline = true
	delete(m.members, n.Name)
}

func (m *Memberlist) NotifyUpdate(n *mb.Node) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, found := m.members[n.Name]

	if !found {
		item = newNode(m, n.Name)
		m.members[n.Name] = item
	}

	item.update(n)
}

func (m *Memberlist) handleListener() {
	for true {
		conn, err := m.listener.Accept()

		if err != nil {
			return
		}

		go m.handleConnection(conn)
	}
}

func (m *Memberlist) handleConnection(conn net.Conn) {
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
