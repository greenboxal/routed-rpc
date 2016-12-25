package routedrpc

import "time"

// Client represents a single RPC sender which may call other addresses
type Client struct {
	address Address
	cluster *Cluster
}

// NewClient returns a new Client with a specific name
func NewClient(cluster *Cluster, address Address) *Client {
	return &Client{
		cluster: cluster,
		address: address,
	}
}

// Cluster instance
func (c *Client) Cluster() *Cluster {
	return c.cluster
}

// Cast sends a RPC call without waiting for a response (fire and forget)
func (c *Client) Cast(target Address, msg interface{}) error {
	return c.cluster.Cast(c.address, target, msg)
}

// GoWithTimeout Sends a Client call to _target_
func (c *Client) GoWithTimeout(target Address, args interface{}, reply interface{}, done chan *Call, timeout time.Duration) *Call {
	return c.cluster.GoWithTimeout(c.address, target, args, reply, done, timeout)
}

// Go sends a Client call to _target_
func (c *Client) Go(target Address, args interface{}, reply interface{}, done chan *Call) *Call {
	return c.cluster.Go(c.address, target, args, reply, done)
}

// CallWithTimeout sends a Client call to _target_
func (c *Client) CallWithTimeout(target Address, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.cluster.CallWithTimeout(c.address, target, args, reply, timeout)
}

// Call sends a Client call to _target_
func (c *Client) Call(target Address, args interface{}, reply interface{}) error {
	return c.cluster.Call(c.address, target, args, reply)
}
