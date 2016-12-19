package routedrpc

type Call struct {
	arrived chan *Call

	Done  chan *Call
	Reply interface{}
	Error error
}
