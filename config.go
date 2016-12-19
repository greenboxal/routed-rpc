package routedrpc

import (
	"time"
)

// Config holds the configuration for Rpc
type Config struct {
	Provider Provider
	Handler  Handler

	ForwardingLimit int

	CallTimeout time.Duration

	ArpTimeout   time.Duration
	ArpCacheSize int
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
