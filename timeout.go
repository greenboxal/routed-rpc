package routedrpc

type Timeout struct {
}

func (t Timeout) Error() string {
	return "the function call timeouted"
}
