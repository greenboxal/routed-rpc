package routedrpc

type Call struct {
	arrived chan *Call

	Done  chan *Call
	Args  interface{}
	Reply interface{}
	Error error
}
