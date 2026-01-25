// Package drivers implements transport drivers for the broker
package drivers

import (
	"github.com/cyw0ng95/v2e/cmd/broker/transport"
)

// TransportDriver defines the interface for transport drivers
type TransportDriver interface {
	CreateTransport(config interface{}) (transport.Transport, error)
	RegisterProcess(processID string)
	UnregisterProcess(processID string)
}

// FDTransportDriver implements transport driver for file descriptor-based communication
type FDTransportDriver struct {
	transports map[string]transport.Transport
}

// NewFDTransportDriver creates a new file descriptor transport driver
func NewFDTransportDriver() *FDTransportDriver {
	return &FDTransportDriver{
		transports: make(map[string]transport.Transport),
	}
}

// CreateTransport creates a new transport based on the provided configuration
func (d *FDTransportDriver) CreateTransport(config interface{}) (transport.Transport, error) {
	// Extract file descriptor configuration
	cfg, ok := config.(map[string]int)
	if !ok {
		return nil, nil // Return nil for now - this is just a placeholder
	}

	inputFD := cfg["input_fd"]
	outputFD := cfg["output_fd"]

	transport := transport.NewFDPipeTransport(inputFD, outputFD)
	return transport, nil
}

// RegisterProcess registers a process with the driver
func (d *FDTransportDriver) RegisterProcess(processID string) {
	// Implementation will depend on specific transport needs
}

// UnregisterProcess unregisters a process from the driver
func (d *FDTransportDriver) UnregisterProcess(processID string) {
	// Implementation will depend on specific transport needs
}
