package transport

import (
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// TransportType represents the type of transport to use
type TransportType string

const (
	TransportTypeUDS TransportType = "uds"
)

// TransportConfig holds configuration for transport creation
type TransportConfig struct {
	Type       TransportType
	SocketPath string
	IsServer   bool
	BasePath   string
}

// NewTransport creates a new transport based on the configuration
func NewTransport(config *TransportConfig) (Transport, error) {
	// Always use UDS transport
	if config.SocketPath == "" {
		// Generate socket path if not provided
		if config.BasePath == "" {
			config.BasePath = subprocess.DefaultProcUDSBasePath()
		}
		config.SocketPath = fmt.Sprintf("%s_%s.sock", config.BasePath, "default")
	}
	transport := NewUDSTransport(config.SocketPath, config.IsServer)
	return transport, nil
}
