package routedrpc

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Rpc struct {
	cache  *cache
	config *Config

	seq     uint64
	mutex   sync.Mutex
	pending map[uint64]*Call
}

func Create(config *Config) *Rpc {
	return &Rpc{
		config:  config,
		cache:   newCache(config.ArpCacheSize),
		pending: make(map[uint64]*Call),
	}
}

func (r *Rpc) Shutdown() error {
	// Cleanup pending calls
	r.mutex.Lock()
	pending := r.pending
	r.pending = make(map[uint64]*Call)
	r.mutex.Unlock()

	// Force all pending calls to timeout
	for _, c := range pending {
		c.Error = Timeout{}
		c.arrived <- c
	}

	// Shutdown the node
	return r.config.Provider.Shutdown()
}

func (r *Rpc) WhoHas(target Address) (Node, error) {
	if r.config.Handler.HasTarget(target) {
		return r.config.Provider.Self(), nil
	}

	name, found := r.cache.Get(target)

	log.Printf("[DEBUG] routed-rpc: Looking address %+v\n", target)

	if !found {
		log.Printf("[DEBUG] routed-rpc: Broadcasting who has request for %+v\n", target)

		go r.config.Provider.Broadcast(whoHasRequest{
			Sender:  r.config.Provider.Self().ID(),
			Address: target,
		})

		name, found = r.cache.WaitAndGet(target, r.config.ArpTimeout)

		if !found {
			return nil, AddressNotFound{}
		}
	}

	log.Printf("[DEBUG] routed-rpc: Found %+v on %s\n", target, name)

	node, found := r.getNode(name)

	if !found {
		return nil, AddressNotFound{}
	}

	return node, nil
}

func (r *Rpc) Cast(target Address, msg interface{}) error {
	return r.sendMessage(&message{
		XID:             0,
		Sender:          r.config.Provider.Self().ID,
		Target:          target,
		ForwardingCount: 0,
		Type:            castMessage,
		Message:         marshal(msg),
	})
}

func (r *Rpc) GoWithTimeout(target Address, args interface{}, reply interface{}, done chan *Call, timeout time.Duration) *Call {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if done == nil {
		done = make(chan *Call, 10)
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel.  If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			panic("done channel is unbuffered")
		}
	}

	call := &Call{
		arrived: make(chan *Call, 1),

		Args:  args,
		Reply: reply,
		Done:  done,
	}

	xid := atomic.AddUint64(&r.seq, 1)
	r.pending[xid] = call

	go func() {
		err := r.sendMessage(&message{
			XID:             xid,
			Sender:          r.config.Provider.Self().ID,
			Target:          target,
			ForwardingCount: 0,
			Type:            callMessage,
			Message:         marshal(args),
		})

		if err != nil {
			call.Error = err
		} else {
			select {
			case <-call.arrived:
				// Handle normally
			case <-time.After(timeout):
				call.Error = Timeout{}
			}
		}

		r.mutex.Lock()
		delete(r.pending, xid)
		r.mutex.Unlock()

		call.Done <- call
	}()

	return call
}

func (r *Rpc) Go(target Address, args interface{}, reply interface{}, done chan *Call) *Call {
	return r.GoWithTimeout(target, args, reply, done, 10*time.Second)
}

func (r *Rpc) CallWithTimeout(target Address, args interface{}, reply interface{}, timeout time.Duration) error {
	call := <-r.GoWithTimeout(target, args, reply, make(chan *Call, 1), timeout).Done

	return call.Error
}

func (r *Rpc) Call(target Address, args interface{}, reply interface{}) error {
	call := <-r.Go(target, args, reply, make(chan *Call, 1)).Done

	return call.Error
}

func (r *Rpc) sendMessage(msg *message) error {
	log.Printf("[DEBUG] routed-rpc: Sending message (xid = %d) to %+v\n", msg.XID, msg.Target)

	node, err := r.WhoHas(msg.Target)

	if err != nil {
		return err
	}

	return node.Send(msg)
}

func (r *Rpc) forwardMessage(msg *message) {
	if msg.ForwardingCount > r.config.ForwardingLimit {
		msg = &message{
			XID:             msg.XID,
			Sender:          r.config.Provider.Self().ID,
			Target:          msg.Sender,
			ForwardingCount: 0,
			Type:            errorMessage,
			Message:         marshal(AddressNotFound{}),
		}
	} else {
		msg.ForwardingCount++
	}

	r.sendMessage(msg)
}

func (r *Rpc) processMessage(msg *message) {
	target := msg.Target

	if !r.config.Handler.HasTarget(target) {
		r.forwardMessage(msg)
		return
	}

	switch msg.Type {
	case castMessage:
		err := r.config.Handler.HandleCast(msg.Sender, target, msg.Message)

		if _, ok := err.(AddressNotFound); ok {
			r.forwardMessage(msg)
			return
		}
	case callMessage:
		var typ messageType
		var reply interface{}

		res, err := r.config.Handler.HandleCall(msg.Sender, target, msg.Message)

		if err != nil {
			if _, ok := err.(AddressNotFound); ok {
				r.forwardMessage(msg)
				return
			}

			typ = errorMessage
			reply = err
		} else {
			typ = replyMessage
			reply = res
		}

		r.sendMessage(&message{
			XID:             msg.XID,
			Sender:          msg.Target,
			Target:          msg.Sender,
			ForwardingCount: 0,
			Type:            typ,
			Message:         marshal(reply),
		})
	case replyMessage:
		fallthrough
	case errorMessage:
		r.mutex.Lock()

		call, found := r.pending[msg.XID]

		if found {
			delete(r.pending, msg.XID)
		}

		r.mutex.Unlock()

		if found {
			reply := unmarshal(msg.Message)

			if msg.Type == replyMessage {
				call.Reply = reply
			} else {
				call.Error = reply.(error)
			}

			call.arrived <- call
		}
	}
}

func (r *Rpc) ProcessRpcMessage(msg interface{}) {
	switch v := msg.(type) {
	case whoHasRequest:
		r.processWhoHasRequest(&v)
	case whoHasReply:
		r.processWhoHasReply(&v)
	case message:
		r.processMessage(&v)
	default:
		log.Printf("[ERROR] routed-rpc: Invalid message: %+v\n", v)
	}
}

func (r *Rpc) processWhoHasRequest(req *whoHasRequest) {
	log.Printf("[DEBUG] routed-rpc: Requested advice for %+v\n", req.Address)

	if !r.config.Handler.HasTarget(req.Address) {
		return
	}

	sender, found := r.getNode(req.Sender)

	if !found {
		return
	}

	sender.Send(whoHasReply{
		Address: req.Address,
	})
}

func (r *Rpc) processWhoHasReply(reply *whoHasReply) {
	log.Printf("[DEBUG] routed-rpc: Got advice that %+v is in %s\n", reply.Address, reply.Who)

	r.cache.Add(reply.Address, reply.Who)
}

func (r *Rpc) getNode(id uint64) (Node, bool) {
	for _, n := range r.config.Provider.Members() {
		if n.ID() == id {
			return n, true
		}
	}

	return nil, false
}
