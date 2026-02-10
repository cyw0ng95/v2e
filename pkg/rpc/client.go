package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// DefaultRPCTimeout is defined in pkg/common/defaults.go
// This package variable provides backward compatibility
var DefaultRPCTimeout = common.DefaultRPCTimeout

// GetDefaultTimeout returns the default RPC timeout duration from pkg/common
func GetDefaultTimeout() time.Duration {
	return common.DefaultRPCTimeout
}

// RequestEntry represents a pending request in the RPC client
type RequestEntry struct {
	resp chan *subprocess.Message
	once sync.Once
}

// Signal signals the request entry with a message
func (e *RequestEntry) Signal(m *subprocess.Message) {
	e.once.Do(func() {
		e.resp <- m
		close(e.resp)
	})
}

// Close closes the request entry
func (e *RequestEntry) Close() {
	e.once.Do(func() {
		close(e.resp)
	})
}

// Client handles RPC communication with other services through the broker
type Client struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]*RequestEntry
	mu              sync.RWMutex
	correlationSeq  uint64
	rpcTimeout      time.Duration
	logger          *common.Logger
}

// NewClient creates a new RPC client for inter-service communication
func NewClient(sp *subprocess.Subprocess, logger *common.Logger, rpcTimeout time.Duration) *Client {
	client := &Client{
		sp:              sp,
		pendingRequests: make(map[string]*RequestEntry),
		rpcTimeout:      rpcTimeout,
		logger:          logger,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.HandleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.HandleError)

	return client
}

// HandleResponse handles response messages from other services
func (c *Client) HandleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Look up the pending request entry and remove it while holding the lock
	c.mu.Lock()
	entry := c.pendingRequests[msg.CorrelationID]
	if entry != nil {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()

	if entry != nil {
		c.logger.Debug("Found pending request for correlation ID: %s, signaling response", msg.CorrelationID)
		entry.Signal(msg)
	} else {
		c.logger.Warn("Received response for unknown correlation ID: %s, type=%s, target=%s", msg.CorrelationID, msg.Type, msg.Target)
	}

	return nil, nil // Don't send another response
}

// HandleError handles error messages from other services (treat them as responses)
func (c *Client) HandleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Error messages are also valid responses
	c.logger.Debug("Handling error message: correlationID=%s, error=%s, target=%s", msg.CorrelationID, msg.Error, msg.Target)
	return c.HandleResponse(ctx, msg)
}

// InvokeRPC invokes an RPC method on another service through the broker
func (c *Client) InvokeRPC(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	// Generate correlation ID
	c.mu.Lock()
	c.correlationSeq++
	correlationID := fmt.Sprintf("rpc-%s-%d-%d", c.sp.ID, time.Now().UnixNano(), c.correlationSeq)
	c.mu.Unlock()

	// Create response channel and entry
	resp := make(chan *subprocess.Message, 1)
	entry := &RequestEntry{resp: resp}

	// Register pending request
	c.mu.Lock()
	c.pendingRequests[correlationID] = entry
	c.mu.Unlock()

	// Clean up on exit: remove from map and close entry
	defer func() {
		c.mu.Lock()
		if _, exists := c.pendingRequests[correlationID]; exists {
			delete(c.pendingRequests, correlationID)
		}
		c.mu.Unlock()
		entry.Close()
	}()

	// Create request message
	var payload []byte
	if params != nil {
		data, err := subprocess.MarshalFast(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		payload = data
	}

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            method,
		Payload:       payload,
		Target:        target,
		CorrelationID: correlationID,
		Source:        c.sp.ID,
	}

	c.logger.Debug("Sending RPC request: method=%s, target=%s, correlationID=%s", method, target, correlationID)

	// Send request to broker (which will route to target)
	if err := c.sp.SendMessage(msg); err != nil {
		c.logger.Error("Failed to send RPC request: %v", err)
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}

	// Wait for response with timeout
	c.logger.Debug("Waiting for RPC response: method=%s, target=%s, correlationID=%s", method, target, correlationID)
	select {
	case response := <-resp:
		c.logger.Debug("Received RPC response: correlationID=%s, type=%s", correlationID, response.Type)
		return response, nil
	case <-time.After(c.rpcTimeout):
		c.logger.Warn("RPC timeout waiting for response: method=%s, target=%s, correlationID=%s", method, target, correlationID)
		return nil, fmt.Errorf("RPC timeout waiting for response from %s", target)
	case <-ctx.Done():
		err := ctx.Err()
		c.logger.Warn("RPC call context canceled while waiting for response: method=%s, target=%s, correlationID=%s, error: %v", method, target, correlationID, err)
		return nil, err
	}
}
