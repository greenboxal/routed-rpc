package routedrpc

import (
	"container/list"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

type cache struct {
	mutex     sync.Mutex
	entries   *lru.ARCCache
	waitSlots map[Address]*list.List
}

func newCache(size int) *cache {
	c, _ := lru.NewARC(size)

	return &cache{
		entries:   c,
		waitSlots: make(map[Address]*list.List),
	}
}

func (c *cache) Add(addr Address, name Node, multiple bool) error {
	c.mutex.Lock()

	if e, found := c.entries.Get(addr); found {
		switch v := e.(type) {
		case *cacheEntry:
			v.Add(name)
		case Node:
			if name.ID() != v.ID() {
				if !multiple {
					c.mutex.Unlock()
					return fmt.Errorf("multiple instances of '%v' detected", addr)
				}

				entry := newCacheEntry()
				entry.Add(v)
				entry.Add(name)

				c.entries.Remove(addr)
				c.entries.Add(addr, entry)
			}
		}
	} else {
		c.entries.Add(addr, name)
	}

	slot, found := c.waitSlots[addr]
	if found {
		delete(c.waitSlots, addr)
	}

	c.mutex.Unlock()

	if found {
		e := slot.Front()

		for e != nil {
			ch := e.Value.(chan Node)

			ch <- name

			e = e.Next()
		}
	}

	return nil
}

func (c *cache) Get(addr Address) (Node, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value, found := c.entries.Get(addr)

	if !found {
		return nil, false
	}

	return value.(Node), true
}

func (c *cache) WaitAndGet(addr Address, timeout time.Duration) (Node, bool) {
	c.mutex.Lock()

	value, found := c.entries.Get(addr)

	if found {
		c.mutex.Unlock()

		return value.(Node), true
	}

	ch := make(chan Node)

	slot, found := c.waitSlots[addr]

	if !found {
		slot = list.New()
		c.waitSlots[addr] = slot
	}

	e := slot.PushBack(ch)

	// We unlock now as we enter the wait state
	c.mutex.Unlock()

	// Wait either for a value or for the timeout signal
	select {
	case v := <-ch:
		value = v
		found = true
	case <-time.After(timeout):
		value = nil
		found = false
	}

	// Cleanup the wait slot
	c.mutex.Lock()
	slot.Remove(e)

	if slot.Len() == 0 {
		delete(c.waitSlots, addr)
	}
	c.mutex.Unlock()

	if !found {
		return nil, found
	}

	return value.(Node), true
}
