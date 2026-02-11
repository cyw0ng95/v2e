package transport

import (
	"fmt"
	"log"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TransportManager manages communication transports for processes
type TransportManager struct {
	transports            map[string]Transport
	udsBasePath           string // Base path for UDS sockets
	transportErrorHandler func(error)
	mu                    sync.RWMutex
}

// NewTransportManager creates a new TransportManager
func NewTransportManager() *TransportManager {
	return &TransportManager{
		transports:  make(map[string]Transport),
		udsBasePath: buildUDSBasePathValue(),
	}
}

// RegisterTransport registers a transport for a process
func (tm *TransportManager) RegisterTransport(processID string, transport Transport) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.transports[processID] = transport
}

// UnregisterTransport removes a transport for a process and closes it
func (tm *TransportManager) UnregisterTransport(processID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if transport, exists := tm.transports[processID]; exists {
		// Close the transport before removing it from the map
		// This ensures the listener is closed and acceptLoop goroutines can exit cleanly
		if err := transport.Close(); err != nil {
			log.Printf("[TransportManager] Error closing transport for process '%s': %v", processID, err)
		}
	}
	delete(tm.transports, processID)
}

// GetTransport gets the transport for a process
func (tm *TransportManager) GetTransport(processID string) (Transport, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transport, exists := tm.transports[processID]
	if !exists {
		return nil, fmt.Errorf("transport for process '%s' not found", processID)
	}

	return transport, nil
}

// SendToProcess sends a message to a process via its transport
func (tm *TransportManager) SendToProcess(processID string, msg *proc.Message) error {
	transport, err := tm.GetTransport(processID)
	if err != nil {
		return err
	}

	return transport.Send(msg)
}

// SetTransportErrorHandler sets a global error handler for all created transports
func (tm *TransportManager) SetTransportErrorHandler(handler func(error)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.transportErrorHandler = handler
}

// RegisterUDSTransport creates and registers a UDS transport for a process
// Returns the socket path created for the transport.
func (tm *TransportManager) RegisterUDSTransport(processID string, isServer bool) (string, error) {
	socketPath := fmt.Sprintf("%s_%s.sock", tm.udsBasePath, processID)
	transport := NewUDSTransport(socketPath, isServer)

	tm.mu.RLock()
	handler := tm.transportErrorHandler
	tm.mu.RUnlock()

	if handler != nil {
		transport.SetErrorHandler(handler)
	}

	if err := transport.Connect(); err != nil {
		return "", fmt.Errorf("failed to connect UDS transport for process %s: %w", processID, err)
	}

	tm.RegisterTransport(processID, transport)
	return socketPath, nil
}

// SetUdsBasePath sets the base path for UDS socket files
func (tm *TransportManager) SetUdsBasePath(path string) {
	tm.udsBasePath = path
}

// CloseAll closes all registered transports. This should be called during broker shutdown.
// Returns an error if any transport close failed, with all errors aggregated.
func (tm *TransportManager) CloseAll() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var closeErrors []error
	for processID, transport := range tm.transports {
		if err := transport.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("process '%s': %w", processID, err))
			log.Printf("[TransportManager] Error closing transport for %s: %v", processID, err)
		}
		delete(tm.transports, processID)
	}

	if len(closeErrors) > 0 {
		log.Printf("[TransportManager] Completed CloseAll with %d error(s)", len(closeErrors))
		return fmt.Errorf("failed to close %d transport(s): %w", len(closeErrors), joinErrors(closeErrors))
	}

	return nil
}

// joinErrors combines multiple errors into a single error message
func joinErrors(errs []error) error {
	var errMsg string
	for i, err := range errs {
		if i > 0 {
			errMsg += "; "
		}
		errMsg += err.Error()
	}
	return fmt.Errorf("%s", errMsg)
}

// IsTransportConnected checks if a transport is connected for a given process.
// Returns true if the transport exists and is connected, false otherwise.
func (tm *TransportManager) IsTransportConnected(processID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transport, exists := tm.transports[processID]
	if !exists {
		return false
	}

	// Check if transport is connected by attempting to send a ping
	// For simplicity, we just check if the transport exists in the map
	// A more thorough check would send a ping message
	return transport != nil
}
