package transport

import (
	"fmt"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TransportManager manages communication transports for processes
type TransportManager struct {
	transports map[string]Transport
	mu         sync.RWMutex
}

// NewTransportManager creates a new TransportManager
func NewTransportManager() *TransportManager {
	return &TransportManager{
		transports: make(map[string]Transport),
	}
}

// RegisterTransport registers a transport for a process
func (tm *TransportManager) RegisterTransport(processID string, transport Transport) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.transports[processID] = transport
}

// UnregisterTransport removes a transport for a process
func (tm *TransportManager) UnregisterTransport(processID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
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