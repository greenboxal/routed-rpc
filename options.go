package routedrpc

import (
	"github.com/hashicorp/memberlist"
	"time"
)

type Options struct {
	*memberlist.Config

	Handler Handler

	ForwardingLimit int

	ArpTimeout   time.Duration
	ArpCacheSize int
}
