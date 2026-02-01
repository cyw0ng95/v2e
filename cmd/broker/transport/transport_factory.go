package transport

import (
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// TransportType represents the type of transport to use
type TransportType string

const (
	TransportTypeFD   TransportType = "fd"
	TransportTypeUDS  TransportType = "uds"
	TransportTypeAuto TransportType = "auto" // Auto-detect based on configuration or environment
)

// TransportConfig holds configuration for transport creation
type TransportConfig struct {
	Type       TransportType
	InputFD    int
	OutputFD   int
	SocketPath string
	IsServer   bool
	BasePath   string
	UDSEnabled bool // Flag to enable UDS transport
	FDEnabled  bool // Flag to enable FD transport
}

// NewTransport creates a new transport based on the configuration
func NewTransport(config *TransportConfig) (Transport, error) {
	// Use configuration flags to determine which transport to enable
	if config.Type == TransportTypeAuto {
		// Auto-detect based on build-time default communicaton type
		if subprocess.DefaultProcCommType() == "uds" {
			config.Type = TransportTypeUDS
		} else {
			config.Type = TransportTypeFD
		}
	}

	// Override based on config flags if set
	if config.UDSEnabled && !config.FDEnabled {
		config.Type = TransportTypeUDS
	} else if config.FDEnabled && !config.UDSEnabled {
		config.Type = TransportTypeFD
	} else if config.UDSEnabled && config.FDEnabled {
		// Both enabled: prefer build-time default
		if subprocess.DefaultProcCommType() == "uds" {
			config.Type = TransportTypeUDS
		} else {
			config.Type = TransportTypeFD
		}
	}

	switch config.Type {
	case TransportTypeFD:
		return NewFDPipeTransport(config.InputFD, config.OutputFD), nil
	case TransportTypeUDS:
		if config.SocketPath == "" {
			// Generate socket path if not provided
			if config.BasePath == "" {
				config.BasePath = subprocess.DefaultProcUDSBasePath()
			}
			config.SocketPath = fmt.Sprintf("%s_%s.sock", config.BasePath, "default")
		}
		transport := NewUDSTransport(config.SocketPath, config.IsServer)
		// Reconnection options remain default unless explicitly configured
		return transport, nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", config.Type)
	}
}

// MultiTransport provides an abstraction layer that can handle both transport types
type MultiTransport struct {
	primary   Transport
	secondary Transport
	active    Transport
	config    *TransportConfig
}

// NewMultiTransport creates a new multi-transport instance
func NewMultiTransport(primaryConfig, secondaryConfig *TransportConfig) (*MultiTransport, error) {
	primary, err := NewTransport(primaryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary transport: %w", err)
	}

	var secondary Transport
	if secondaryConfig != nil {
		secondary, err = NewTransport(secondaryConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create secondary transport: %w", err)
		}
	}

	return &MultiTransport{
		primary:   primary,
		secondary: secondary,
		active:    primary, // Start with primary
		config:    primaryConfig,
	}, nil
}

// Connect connects the active transport
func (mt *MultiTransport) Connect() error {
	return mt.active.Connect()
}

// Send sends a message through the active transport
func (mt *MultiTransport) Send(msg *proc.Message) error {
	return mt.active.Send(msg)
}

// Receive receives a message from the active transport
func (mt *MultiTransport) Receive() (*proc.Message, error) {
	return mt.active.Receive()
}

// Close closes the active transport
func (mt *MultiTransport) Close() error {
	return mt.active.Close()
}

// SwitchTransport switches to a different transport type
func (mt *MultiTransport) SwitchTransport(transportType TransportType) error {
	var newTransport Transport
	var err error

	config := &TransportConfig{
		Type:     transportType,
		InputFD:  mt.config.InputFD,
		OutputFD: mt.config.OutputFD,
		IsServer: mt.config.IsServer,
		BasePath: mt.config.BasePath,
	}

	newTransport, err = NewTransport(config)
	if err != nil {
		return fmt.Errorf("failed to create new transport: %w", err)
	}

	// Try to connect the new transport
	if err := newTransport.Connect(); err != nil {
		return fmt.Errorf("failed to connect new transport: %w", err)
	}

	// Close the old active transport
	if mt.active != nil {
		mt.active.Close()
	}

	// Switch to the new transport
	mt.active = newTransport
	mt.config.Type = transportType

	return nil
}

// GetCurrentTransportType returns the type of the currently active transport
func (mt *MultiTransport) GetCurrentTransportType() TransportType {
	return mt.config.Type
}
