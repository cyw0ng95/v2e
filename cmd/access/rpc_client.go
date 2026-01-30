package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

type requestEntry struct {
	resp chan *subprocess.Message
	once sync.Once
}

func (e *requestEntry) signal(m *subprocess.Message) {
	e.once.Do(func() {
		e.resp <- m
		close(e.resp)
	})
}

func (e *requestEntry) close() {
	e.once.Do(func() {
		close(e.resp)
	})
}

// RPCClient handles RPC communication with the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]*requestEntry
	mu              sync.RWMutex
	correlationSeq  uint64
	// per-client RPC timeout (configurable)
	rpcTimeout time.Duration
}

// NewRPCClient creates a new RPC client for broker communication
func NewRPCClient(processID string, rpcTimeout time.Duration) *RPCClient {
	var sp *subprocess.Subprocess

	// Check if we're running as an RPC subprocess with file descriptors
	if os.Getenv("BROKER_PASSING_RPC_FDS") == "1" {
		// Use file descriptors 3 and 4 for RPC communication
		inputFD := 3
		outputFD := 4

		// Allow environment override for file descriptors
		if val := os.Getenv("RPC_INPUT_FD"); val != "" {
			if fd, err := strconv.Atoi(val); err == nil {
				inputFD = fd
			}
		}
		if val := os.Getenv("RPC_OUTPUT_FD"); val != "" {
			if fd, err := strconv.Atoi(val); err == nil {
				outputFD = fd
			}
		}

		sp = subprocess.NewWithFDs(processID, inputFD, outputFD)
	} else {
		// Use default stdin/stdout for non-RPC mode
		sp = subprocess.New(processID)
	}

	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]*requestEntry),
		rpcTimeout:      rpcTimeout,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)

	return client
}

// handleResponse handles response messages from the broker
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Look up the pending requestEntry
	c.mu.Lock()
	entry := c.pendingRequests[msg.CorrelationID]
	if entry != nil {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()

	if entry != nil {
		common.Debug(LogMsgRPCResponseReceived, msg.CorrelationID, msg.Type)
		common.Debug("Signaling response for correlation ID: %s, message type: %s", msg.CorrelationID, msg.Type)
		entry.signal(msg)
		common.Debug(LogMsgRPCChannelSignal, msg.CorrelationID)
	} else {
		common.Warn("Received response for unknown correlation ID: %s, message type: %s, target: %s", msg.CorrelationID, msg.Type, msg.Target)
	}

	return nil, nil // Don't send another response
}

// handleError handles error messages from the broker (treat them as responses)
func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Error messages are also valid responses
	common.Debug("Handling error message for correlation ID: %s, error: %s", msg.CorrelationID, msg.Error)
	return c.handleResponse(ctx, msg)
}

// InvokeRPC invokes an RPC method on the broker and waits for response
func (c *RPCClient) InvokeRPC(ctx context.Context, method string, params interface{}) (*subprocess.Message, error) {
	return c.InvokeRPCWithTarget(ctx, "broker", method, params)
}

// InvokeRPCWithTarget invokes an RPC method on a specific target process and waits for response
func (c *RPCClient) InvokeRPCWithTarget(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	// Generate correlation ID
	c.mu.Lock()
	c.correlationSeq++
	correlationID := fmt.Sprintf("access-rpc-%d", c.correlationSeq)
	c.mu.Unlock()

	common.Debug("Creating RPC request with correlation ID: %s, method: %s, target: %s", correlationID, method, target)

	// Create response channel and entry
	resp := make(chan *subprocess.Message, 1)
	entry := &requestEntry{resp: resp}

	// Register pending request
	c.mu.Lock()
	c.pendingRequests[correlationID] = entry
	c.mu.Unlock()
	common.Debug(LogMsgRPCPendingRequestAdded, correlationID)

	// Clean up on exit: remove from map and ensure channel is closed exactly once
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		entry.close()
		common.Debug(LogMsgRPCPendingRequestRemoved, correlationID)
	}()

	// Create request message
	var payload json.RawMessage
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
		Source:        c.sp.ID, // Set the source to this service's process ID
	}

	// Send request to broker
	common.Debug("About to send RPC request: %s to target: %s, correlation: %s", method, target, correlationID)
	if err := c.sp.SendMessage(msg); err != nil {
		common.Error("Failed to send RPC request: %s to target: %s, correlation: %s, error: %v", method, target, correlationID, err)
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}
	common.Debug("Successfully sent RPC request: %s to target: %s, correlation: %s", method, target, correlationID)

	// Wait for response with timeout (use configured rpcTimeout)
	common.Debug("Waiting for RPC response: %s to target: %s, correlation: %s", method, target, correlationID)
	select {
	case response := <-resp:
		common.Debug("Received RPC response: %s from target: %s, correlation: %s", method, target, correlationID)
		return response, nil
	case <-time.After(c.rpcTimeout):
		common.Error("RPC timeout waiting for response: %s from target: %s, correlation: %s", method, target, correlationID)
		return nil, fmt.Errorf("RPC timeout waiting for response")
	case <-ctx.Done():
		err := ctx.Err()
		common.Error("RPC call context canceled while waiting for response: %s from target: %s, correlation: %s, error: %v", method, target, correlationID, err)
		return nil, err
	}
}

// Run starts the RPC client message processing
func (c *RPCClient) Run(ctx context.Context) error {
	return c.sp.Run()
}
