package memberlist

import (
	"github.com/greenboxal/routed-rpc"
	mb "github.com/hashicorp/memberlist"
)

type Memberlist struct {
	*mb.Memberlist

	rpc *routedrpc.RPC
}

func Create(config *mb.Config) (*Memberlist, error) {
	result := &Memberlist{}

	config.Delegate = result

	memberlist, err := mb.Create(config)

	if err != nil {
		return nil, err
	}

	result.Memberlist = memberlist

	return result, nil
}

func (m *Memberlist) Self() routedrpc.Node {
	return &node{Node: m.LocalNode(), m: m}
}

func (m *Memberlist) Members() []routedrpc.Node {
	members := m.Memberlist.Members()
	result := make([]routedrpc.Node, len(members))

	for i, v := range members {
		result[i] = &node{Node: v, m: m}
	}

	return result
}

func (m *Memberlist) Broadcast(msg interface{}) error {
	data, err := encode(msg)

	if err != nil {
		return err
	}

	for _, n := range m.Memberlist.Members() {
		err = m.SendToTCP(n, data)
	}

	return err
}

func (m *Memberlist) SetRpc(rpc *routedrpc.RPC) {
	m.rpc = rpc
}

func (m *Memberlist) NodeMeta(limit int) []byte {
	return []byte{}
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
