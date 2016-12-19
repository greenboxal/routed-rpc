package routedrpc

import (
	"bytes"
	"encoding/gob"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/memberlist"
)

type Rpc struct {
	Members *memberlist.Memberlist

	cache   *cache
	options *Options

	seq     uint64
	mutex   sync.Mutex
	pending map[uint64]*Call
}

func Create(options *Options) (*Rpc, error) {
	rpc := &Rpc{
		options: options,
		cache:   newCache(options.ArpCacheSize),
	}

	options.Delegate = &memberlistDelegate{rpc}

	members, err := memberlist.Create(options.Config)

	if err != nil {
		return nil, err
	}

	rpc.Members = members

	return rpc, nil
}

func (r *Rpc) WhoHas(target Address) (*memberlist.Node, error) {
	name, found := r.cache.Get(target)

	if !found {
		go r.broadcast(whoHasRequest{
			Sender:  r.Members.LocalNode().Name,
			Address: target,
		})

		name, found = r.cache.WaitAndGet(target, r.options.ArpTimeout)

		if !found {
			return nil, AddressNotFound{}
		}
	}

	node, found := r.getNode(name)

	if !found {
		return nil, AddressNotFound{}
	}

	return node, nil
}

func (r *Rpc) Cast(target Address, msg interface{}) error {
	return r.sendMessage(&message{
		XID:             0,
		Sender:          r.Members.LocalNode().Name,
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
			Sender:          r.Members.LocalNode().Name,
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
	node, err := r.WhoHas(msg.Target)

	if err != nil {
		return err
	}

	return r.sendRaw(node, msg)
}

func (r *Rpc) sendRaw(n *memberlist.Node, msg interface{}) error {
	buff := bytes.NewBuffer([]byte{})

	err := gob.NewEncoder(buff).Encode(msg)

	if err != nil {
		return err
	}

	return r.Members.SendToTCP(n, buff.Bytes())
}

func (r *Rpc) broadcast(msg interface{}) error {
	buff := bytes.NewBuffer([]byte{})

	err := gob.NewEncoder(buff).Encode(msg)

	if err != nil {
		return err
	}

	for _, n := range r.Members.Members() {
		r.Members.SendToTCP(n, buff.Bytes())
	}

	return nil
}

func (r *Rpc) processMessage(msg *message) {
	target := msg.Target

	if !r.options.Handler.HasTarget(target) {
		if msg.ForwardingCount > r.options.ForwardingLimit {
			// TODO: Notify
			return
		}

		msg.ForwardingCount++

		r.sendMessage(msg)

		return
	}

	switch msg.Type {
	case castMessage:
		r.options.Handler.HandleCast(msg.Sender, target, msg.Message)
	case callMessage:
		res, err := r.options.Handler.HandleCall(msg.Sender, target, msg.Message)

		if err != nil {
			msg.Type = errorMessage
			msg.Message = marshal(err)
		} else {
			msg.Type = replyMessage
			msg.Message = marshal(res)
		}

		msg.Target = msg.Sender
		msg.Sender = target

		r.sendMessage(msg)
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
				call.Reply = unmarshal(msg.Message)
			} else {
				call.Error = unmarshal(msg.Message).(error)
			}

			call.arrived <- call
		}
	}
}

func (r *Rpc) processRpcMessage(data []byte) {
	var msg interface{}

	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&msg)

	if err != nil {
		return
	}

	switch v := msg.(type) {
	case whoHasRequest:
		r.processWhoHasRequest(&v)
	case whoHasReply:
		r.processWhoHasReply(&v)
	case message:
		r.processMessage(&v)
	}
}

func (r *Rpc) processWhoHasRequest(req *whoHasRequest) {
	if !r.options.Handler.HasTarget(req.Address) {
		return
	}

	sender, found := r.getNode(req.Sender)

	if !found {
		return
	}

	r.sendRaw(sender, whoHasReply{})
}

func (r *Rpc) processWhoHasReply(reply *whoHasReply) {
	r.cache.Add(reply.Address, reply.Who)
}

func (r *Rpc) getNode(name string) (*memberlist.Node, bool) {
	for _, n := range r.Members.Members() {
		if n.Name == name {
			return n, true
		}
	}

	return nil, false
}
