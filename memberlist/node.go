package memberlist

import (
	mb "github.com/hashicorp/memberlist"
)

type node struct {
	*mb.Node
	m *Memberlist
}

func (n *node) ID() interface{} {
	return n.Name
}

func (n *node) Send(msg interface{}) error {
	data, err := encode(msg)

	if err != nil {
		return err
	}

	return n.m.SendToTCP(n.Node, data)
}
