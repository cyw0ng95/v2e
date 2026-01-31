package main

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

// RPCClient handles RPC communication with the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	client          *rpc.Client // Use the common RPC client
	pendingRequests map[string]*requestEntry
	mu              sync.RWMutex
	correlationSeq  uint64
	// per-client RPC timeout (configurable)
	rpcTimeout time.Duration
}

// NewRPCClient creates a new RPC client for broker communication
func NewRPCClient(processID string, rpcTimeout time.Duration) *RPCClient {
	// We'll create the subprocess separately and pass it in
	// For now, we'll create it here as before but eventually it should come from standard startup
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

	// Create a dummy logger for the common RPC client
	logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)

	client := &RPCClient{
		sp:              sp,
		client:          rpc.NewClient(sp, logger, rpcTimeout),
		pendingRequests: make(map[string]*requestEntry),
		rpcTimeout:      rpcTimeout,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)

	return client
}

// NewRPCClientWithSubprocess creates a new RPC client using an existing subprocess instance
func NewRPCClientWithSubprocess(sp *subprocess.Subprocess, logger *common.Logger, rpcTimeout time.Duration) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		client:          rpc.NewClient(sp, logger, rpcTimeout),
		pendingRequests: make(map[string]*requestEntry),
		rpcTimeout:      rpcTimeout,
	}

	// The common rpc.Client already registers its own handlers for response and error messages
	// No need to register additional handlers here

	return client
}

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
	// Use the common client's InvokeRPC method
	return c.client.InvokeRPC(ctx, target, method, params)
}

// Run starts the RPC client message processing
func (c *RPCClient) Run(ctx context.Context) error {
	return c.sp.Run()
}
