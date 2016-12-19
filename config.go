package routedrpc

import (
	"time"
)

type Config struct {
	Provider Provider
	Handler  Handler

	ForwardingLimit int

	ArpTimeout   time.Duration
	ArpCacheSize int
}
