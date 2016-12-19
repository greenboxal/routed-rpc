package routedrpc

type memberlistDelegate struct {
	r *Rpc
}

func (d *memberlistDelegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *memberlistDelegate) NotifyMsg(msg []byte) {
	d.r.processRpcMessage(msg)
}

func (d *memberlistDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

func (d *memberlistDelegate) LocalState(join bool) []byte {
	return []byte{}
}

func (d *memberlistDelegate) MergeRemoteState(buf []byte, join bool) {
}
