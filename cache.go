package routedrpc

import (
	"container/list"
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

func (c *cache) Add(addr Address, name interface{}) {
	c.mutex.Lock()

	c.entries.Add(addr, name)

	slot, found := c.waitSlots[addr]
	if found {
		delete(c.waitSlots, addr)
	}

	c.mutex.Unlock()

	if found {
		e := slot.Front()

		for e != nil {
			ch := e.Value.(chan interface{})

			ch <- name

			e = e.Next()
		}
	}
}

func (c *cache) Get(addr Address) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value, found := c.entries.Get(addr)

	if !found {
		return nil, false
	}

	return value.(interface{}), true
}

func (c *cache) WaitAndGet(addr Address, timeout time.Duration) (interface{}, bool) {
	c.mutex.Lock()

	value, found := c.entries.Get(addr)

	if found {
		c.mutex.Unlock()

		return value.(interface{}), true
	}

	ch := make(chan interface{})

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

	return value.(interface{}), true
}
