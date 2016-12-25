package routedrpc

// A Call represnts a pending responde from the cluster
type Call struct {
	arrived chan *Call

	Done  chan *Call  // Signaled when call is complete.
	Reply interface{} // Responde message.
	Error error       // After completion, the error status.
}
