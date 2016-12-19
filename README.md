# routedrpc

A distributed RPC system which can handle targets who move between nodes.

# Use cases

This was originally built for handling distributed game servers where players and other objects can move between servers.
One use case is sending a chat message from one player to another and requesting general information about an object.

# Implementation

The system architecture is based on the ARP protocol.
When a message (RPC call) needs to be sent, the system asks to all nodes where that object can be found (based on its address).

When a response is received it's cached and the message forwarded to the correct node.
If a message arrives and the object is not present in that server anymore, it is then forwarded again to the correct server.
This is done automatically as long as the object is still being handled in any server.

# Documentation

Check [godoc](https://godoc.org/github.com/greenboxal/routed-rpc).

# License

MIT.

