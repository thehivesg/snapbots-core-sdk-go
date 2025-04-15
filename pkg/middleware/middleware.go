package middleware

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Config holds the configuration for the middleware
type Config struct {
	BotID   string
	NATSURL string
	Timeout time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout: 3 * time.Second,
	}
}

// MiddlewareFactory creates and manages middleware instances
type MiddlewareFactory struct {
	config     *Config
	natsClient *nats.Conn
}

// NewMiddlewareFactory creates a new middleware factory
func NewMiddlewareFactory(config *Config) (*MiddlewareFactory, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Connect to NATS
	nc, err := nats.Connect(config.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &MiddlewareFactory{
		config:     config,
		natsClient: nc,
	}, nil
}

// Close closes the NATS connection
func (f *MiddlewareFactory) Close() {
	if f.natsClient != nil {
		f.natsClient.Close()
	}
}
