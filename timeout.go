package routedrpc

// Timeout is returned when a RPC call times out
type Timeout struct {
}

// Returns a text representation of the error
func (t Timeout) Error() string {
	return "the function call timeouted"
}
