package memberlist

import (
	mb "github.com/hashicorp/memberlist"
	"net"
)

type delegate struct {
	*Provider
}

func (m *delegate) NodeMeta(limit int) []byte {
	endpoint := net.ResolveTCPAddr("tcp", m.config.RpcAdvertiseEndpoint)
	data, err := encode(endpoint)

	if err != nil {
		return nil
	}

	return data
}

func (m *delegate) NotifyMsg(data []byte) {
}

func (m *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

func (m *delegate) LocalState(join bool) []byte {
	return []byte{}
}

func (m *delegate) MergeRemoteState(buf []byte, join bool) {
}

func (m *delegate) NotifyJoin(n *mb.Node) {
	m.updateNode(n)
}

func (m *delegate) NotifyLeave(n *mb.Node) {
	m.removeNode(n)
}

func (m *delegate) NotifyUpdate(n *mb.Node) {
	m.updateNode(n)
}
