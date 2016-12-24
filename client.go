package routedrpc

import "time"

// Client represents a single RPC sender which may call other addresses
type Client struct {
	address Address
	rpc     *RPC
}

// NewClient returns a new Client with a specific name
func NewClient(rpc *RPC, address Address) *Client {
	return &Client{
		rpc:     rpc,
		address: address,
	}
}

// RPC instance
func (c *Client) RPC() *RPC {
	return c.rpc
}

// Cast sends a RPC call without waiting for a response (fire and forget)
func (c *Client) Cast(target Address, msg interface{}) error {
	return c.rpc.Cast(c.address, target, msg)
}

// GoWithTimeout Sends a Client call to _target_
func (c *Client) GoWithTimeout(target Address, args interface{}, reply interface{}, done chan *Call, timeout time.Duration) *Call {
	return c.rpc.GoWithTimeout(c.address, target, args, reply, done, timeout)
}

// Go sends a Client call to _target_
func (c *Client) Go(target Address, args interface{}, reply interface{}, done chan *Call) *Call {
	return c.rpc.Go(c.address, target, args, reply, done)
}

// CallWithTimeout sends a Client call to _target_
func (c *Client) CallWithTimeout(target Address, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.rpc.CallWithTimeout(c.address, target, args, reply, timeout)
}

// Call sends a Client call to _target_
func (c *Client) Call(target Address, args interface{}, reply interface{}) error {
	return c.rpc.Call(c.address, target, args, reply)
}
