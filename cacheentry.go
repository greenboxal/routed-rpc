package routedrpc

import (
	"container/list"
	"sync"
)

// LoadBalancedNodeAddress represents a virtual node that actually load balances between multiple nodes
var LoadBalancedNodeAddress = "b99389a0-57ad-4feb-8de0-b455b8263f2b"

type cacheEntry struct {
	mutex sync.Mutex
	nodes *list.List
}

func newCacheEntry() *cacheEntry {
	return &cacheEntry{
		nodes: list.New(),
	}
}

func (e *cacheEntry) ID() interface{} {
	return LoadBalancedNodeAddress
}

func (e *cacheEntry) Add(node Node) {
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

func (n *cacheEntry) Send(msg interface{}) error {
	for true {
		if n.nodes.Len() == 0 {
			return AddressNotFound{}
		}

		n.mutex.Lock()

		e := n.nodes.Front()
		node := e.Value.(Node)

		n.nodes.Remove(e)
		n.nodes.PushBack(node)

		n.mutex.Unlock()

		err := node.Send(msg)

		if err != nil {
			n.mutex.Lock()
			n.nodes.Remove(e)
			n.mutex.Unlock()
			continue
		}

		break
	}

	return nil
}
