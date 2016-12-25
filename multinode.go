package routedrpc

import (
	"container/list"
	"sync"
)

// LoadBalancedNodeAddress represents a virtual node that actually load balances between multiple nodes
var LoadBalancedNodeAddress = "b99389a0-57ad-4feb-8de0-b455b8263f2b"

type multiNode struct {
	mutex   sync.Mutex
	nodes   *list.List
	current *list.Element
}

func newCacheEntry() *multiNode {
	return &multiNode{
		nodes: list.New(),
	}
}

func (e *multiNode) ID() interface{} {
	return LoadBalancedNodeAddress
}

func (e *multiNode) Add(node Node) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	item := e.nodes.Front()
	for item != nil {
		n := item.Value.(Node)

		if n.ID() == node.ID() {
			return
		}

		item = item.Next()
	}

	e.nodes.PushBack(node)
}

func (n *multiNode) getNext() (Node, *list.Element) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	current := n.current

	if current == nil {
		current = n.nodes.Front()
	}

	for current != nil {
		e := current
		node := e.Value.(Node)

		current = current.Next()

		if !node.Online() {
			n.removeNodeFromList(e)
			continue
		}

		n.current = current

		return node, e
	}

	return nil, nil
}

func (n *multiNode) removeNodeFromList(e *list.Element) {
	if n.current == e {
		n.current = n.current.Next()
	}

	n.nodes.Remove(e)
}

func (n *multiNode) Send(msg interface{}) error {
	for true {
		node, e := n.getNext()

		if node == nil {
			break
		}

		err := node.Send(msg)

		if err != nil {
			n.removeNodeFromList(e)
			continue
		}

		return nil
	}

	return AddressNotFound{}
}

func (n *multiNode) Online() bool {
	e := n.nodes.Front()

	for e != nil {
		node := e.Value.(Node)

		if node.Online() {
			return true
		}

		e = e.Next()
	}

	return false
}
