package memberlist

import (
	"encoding/gob"
	"errors"
	"net"
	"sync"
	"time"

	mb "github.com/hashicorp/memberlist"
)

var connectionTimeout = 30 * time.Second

type node struct {
	mutex sync.Mutex
	id    string

	addr net.TCPAddr
	conn *nodeConnection

	current  *mb.Node
	provider *Provider

	offline bool
}

type nodeConnection struct {
	conn    net.Conn
	timer   *time.Timer
	encoder *gob.Encoder
	closer  chan interface{}
}

func init() {
	gob.Register(net.TCPAddr{})
}

func newNode(m *Provider, id string) *node {
	return &node{
		id:       id,
		provider: m,
		offline:  true,
	}
}

func (n *node) ID() interface{} {
	return n.id
}

func (n *node) Send(msg interface{}) error {
	if n.offline {
		return errors.New("member is offline")
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	conn, err := n.requestConnection()

	if err != nil {
		return err
	}

	conn.timer.Reset(connectionTimeout)

	return conn.encoder.Encode(&msg)
}

func (n *node) Online() bool {
	return !n.offline
}

func (n *node) sendRaw(data []byte) error {
	if n.offline {
		return errors.New("member is offline")
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	conn, err := n.requestConnection()

	if err != nil {
		return err
	}

	conn.timer.Reset(connectionTimeout)
	_, err = n.conn.conn.Write(data)

	return err
}

func (n *node) update(node *mb.Node) error {
	return n.updateWithConnection(node, nil)
}

func (n *node) updateWithConnection(node *mb.Node, conn net.Conn) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	newAddressInt, err := decode(node.Meta)

	if err != nil {
		return err
	}

	newAddress := newAddressInt.(net.TCPAddr)

	if newAddress.IP.Equal(n.addr.IP) && newAddress.Port == n.addr.Port {
		return nil
	}

	n.addr = newAddress
	n.offline = false

	if n.conn != nil {
		n.conn.closer <- node
		n.conn = nil
	}

	if conn != nil {
		n.setConnection(conn)
	}

	return nil
}

func (n *node) requestConnection() (*nodeConnection, error) {
	if n.conn != nil {
		return n.conn, nil
	}

	conn, err := net.DialTCP("tcp", nil, &n.addr)

	if err != nil {
		return nil, err
	}

	return n.setConnection(conn)
}

func (n *node) setConnection(conn net.Conn) (*nodeConnection, error) {
	nodeConn := &nodeConnection{
		closer:  make(chan interface{}),
		conn:    conn,
		encoder: gob.NewEncoder(conn),
		timer:   time.NewTimer(connectionTimeout),
	}

	n.conn = nodeConn

	go n.connectionWatcher(nodeConn)
	go n.connectionHandler(nodeConn)

	return nodeConn, nil
}

func (n *node) connectionWatcher(conn *nodeConnection) {
	select {
	case <-conn.timer.C:
		conn.conn.Close()
	case <-conn.closer:
		conn.conn.Close()
	}
}

func (n *node) connectionHandler(conn *nodeConnection) {
	var value interface{}

	d := gob.NewDecoder(conn.conn)

	for true {
		err := d.Decode(&value)

		if err != nil {
			conn.conn.Close()
			break
		}

		n.provider.cluster.ProcessRPCMessage(value)
	}
}
