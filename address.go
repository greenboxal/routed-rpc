package routedrpc

// Address represents an address inside the cluster
// Can be any serialized object as long as both sides can handle it
type Address interface{}

// InvalidAddress represents an invalid and non routable address inside the system
var InvalidAddress Address = "b99389a0-57ad-4feb-8de0-b455b8263f2b"
