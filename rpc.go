package routedrpc

import (
	"io/ioutil"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// RPC holds state for the current node in the system
type RPC struct {
	cache  *cache
	config *Config

	seq     uint64
	mutex   sync.Mutex
	pending map[uint64]*Call
}

// Create a Rpc instance
func Create(config *Config) *RPC {
	rpc := &RPC{
		config:  config,
		cache:   newCache(config.ArpCacheSize),
		pending: make(map[uint64]*Call),
		seq:     1,
	}

	if config.Log == nil {
		config.Log = log.New(ioutil.Discard, "", 0)
	}

	config.Provider.SetRPC(rpc)

	return rpc
}

// Shutdown stops this node operations
func (r *RPC) Shutdown() error {
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

// WhoHas Returns which nodes can handle the provided address
//
// This information is retried either from the ARP cache or asked to the network
func (r *RPC) WhoHas(target Address) (Node, error) {
	if ok, _ := r.config.Handler.HasTarget(target); ok {
		return r.config.Provider.Self(), nil
	}

	node, found := r.cache.Get(target)

	if !found {
		go r.config.Provider.Broadcast(whoHasRequest{
			Sender:  r.config.Provider.Self().ID(),
			Address: target,
		})

		node, found = r.cache.WaitAndGet(target, r.config.ArpTimeout)

		if !found {
			return nil, AddressNotFound{}
		}
	}

	return node, nil
}

// Cast sends a RPC call without waiting for a response (fire and forget)
func (r *RPC) Cast(target Address, msg interface{}) error {
	return r.sendMessage(&message{
		XID:             0,
		SenderID:        r.config.Provider.Self().ID(),
		Sender:          nil,
		Target:          target,
		ForwardingCount: 0,
		Type:            castMessage,
		Message:         msg,
	})
}

// GoWithTimeout Sends a RPC call to _target_
func (r *RPC) GoWithTimeout(target Address, args interface{}, reply interface{}, done chan *Call, timeout time.Duration) *Call {
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

	replyValue := reflect.ValueOf(reply)

	if replyValue.Kind() == reflect.Ptr {
		replyValue = replyValue.Elem()
	}

	call := &Call{
		arrived: make(chan *Call, 1),
		Done:    done,
	}

	xid := atomic.AddUint64(&r.seq, 1)
	r.pending[xid] = call

	go func() {
		err := r.sendMessage(&message{
			XID:             xid,
			SenderID:        r.config.Provider.Self().ID(),
			Sender:          nil,
			Target:          target,
			ForwardingCount: 0,
			Type:            callMessage,
			Message:         args,
		})

		if err != nil {
			call.Error = err
		} else {
			select {
			case <-call.arrived:
				replyValue.Set(reflect.ValueOf(call.Reply))
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

// Go sends a RPC call to _target_
func (r *RPC) Go(target Address, args interface{}, reply interface{}, done chan *Call) *Call {
	return r.GoWithTimeout(target, args, reply, done, r.config.CallTimeout)
}

// CallWithTimeout sends a RPC call to _target_
func (r *RPC) CallWithTimeout(target Address, args interface{}, reply interface{}, timeout time.Duration) error {
	call := <-r.GoWithTimeout(target, args, reply, make(chan *Call, 1), timeout).Done

	return call.Error
}

// Call sends a RPC call to _target_
func (r *RPC) Call(target Address, args interface{}, reply interface{}) error {
	return r.CallWithTimeout(target, args, reply, r.config.CallTimeout)
}

func (r *RPC) sendMessage(msg *message) error {
	node, err := r.WhoHas(msg.Target)

	if err != nil {
		return err
	}

	r.config.Log.Printf("sending message to %v\n", node.ID())

	return node.Send(msg)
}

func (r *RPC) forwardMessage(msg *message) {
	if msg.ForwardingCount > r.config.ForwardingLimit {
		msg = &message{
			XID:             msg.XID,
			Sender:          msg.Sender,
			SenderID:        msg.SenderID,
			Target:          msg.Target,
			ForwardingCount: 0,
			Type:            errorMessage,
			Message:         AddressNotFound{},
		}
	} else {
		msg.ForwardingCount++
	}

	r.sendMessage(msg)
}

func (r *RPC) processMessage(msg *message) {
	target := msg.Target

	if msg.Type == castMessage || msg.Type == callMessage {
		if ok, _ := r.config.Handler.HasTarget(target); !ok {
			r.forwardMessage(msg)
			return
		}
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

		node, found := r.config.Provider.GetMember(msg.SenderID)

		if !found {
			return
		}

		node.Send(&message{
			XID:             msg.XID,
			Sender:          nil,
			SenderID:        r.config.Provider.Self().ID(),
			Target:          nil,
			ForwardingCount: 0,
			Type:            typ,
			Message:         reply,
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
			if msg.Type == replyMessage {
				call.Reply = msg.Message
			} else {
				call.Error = msg.Message.(error)
			}

			call.arrived <- call
		}
	}
}

// ProcessRPCMessage should be called by the Provider implementation in order to forward system messages
func (r *RPC) ProcessRPCMessage(msg interface{}) {
	switch v := msg.(type) {
	case *whoHasRequest:
		r.processWhoHasRequest(v)
	case whoHasRequest:
		r.processWhoHasRequest(&v)

	case *whoHasReply:
		r.processWhoHasReply(v)
	case whoHasReply:
		r.processWhoHasReply(&v)

	case *message:
		r.processMessage(v)
	case message:
		r.processMessage(&v)
	}
}

func (r *RPC) processWhoHasRequest(req *whoHasRequest) {
	ok, ha := r.config.Handler.HasTarget(req.Address)

	if !ok {
		return
	}

	sender, found := r.config.Provider.GetMember(req.Sender)

	if !found {
		return
	}

	sender.Send(whoHasReply{
		Who:      r.config.Provider.Self().ID(),
		Address:  req.Address,
		Multiple: ha,
	})
}

func (r *RPC) processWhoHasReply(reply *whoHasReply) {
	member, found := r.config.Provider.GetMember(reply.Who)

	if !found {
		r.config.Log.Printf("error adding %v to cache: %v doesn't exist\n", reply.Address, reply.Who)
		return
	}

	err := r.cache.Add(reply.Address, member, reply.Multiple)

	if err != nil {
		r.config.Log.Printf("error adding %v on %v to cache: %s\n", reply.Address, member.ID(), err.Error())
		return
	}

	r.config.Log.Printf("adding %v on %v (ha = %v)", reply.Address, member.ID(), reply.Multiple)
}
