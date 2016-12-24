package routedrpc

import (
	"time"

	"github.com/Sirupsen/logrus"
)

// Config holds the configuration for Rpc
type Config struct {
	// Distributed system provider
	Provider Provider

	// The handler used
	Handler Handler

	// How many times a message can be forwarded before being discarded
	ForwardingLimit int

	// Default call timeout
	CallTimeout time.Duration

	// Timeout for waiting ARP responses
	ArpTimeout time.Duration

	// ARP cache size
	ArpCacheSize int

	// Logger
	Log *logrus.Entry
}

// DefaultConfig returns sane config defaults
func DefaultConfig() *Config {
	return &Config{
		Provider:        nil,
		Handler:         NewStandardHandler(),
		ForwardingLimit: 5,
		CallTimeout:     2 * time.Second,
		ArpTimeout:      1 * time.Second,
		ArpCacheSize:    1000000,
	}
}
