package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

// RPCClient handles RPC communication with the broker
type RPCClient struct {
	sp     *subprocess.Subprocess
	client *rpc.Client // Use the common RPC client
	// per-client RPC timeout (configurable)
	rpcTimeout time.Duration
}

// NewRPCClient creates a new RPC client for broker communication
func NewRPCClient(processID string, rpcTimeout time.Duration) *RPCClient {
	// Create a dummy logger for the common RPC client
	logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)

	// Use deterministic UDS path based on build-time base path and process ID
	socketPath := fmt.Sprintf("%s_%s.sock", subprocess.DefaultProcUDSBasePath(), processID)
	var sp *subprocess.Subprocess

	// Only attempt to create a UDS-backed subprocess if the socket exists
	// to avoid forcing process exit when no broker listener is present
	if _, err := os.Stat(socketPath); err == nil {
		var err error
		sp, err = subprocess.NewWithUDS(processID, socketPath)
		if err != nil {
			logger.Warn("Failed to create UDS subprocess, falling back to stdio:", err)
			sp = subprocess.New(processID)
		}
	} else {
		// Fallback to stdio subprocess (useful for tests that don't set up a socket)
		sp = subprocess.New(processID)
	}

	client := &RPCClient{
		sp:         sp,
		client:     rpc.NewClient(sp, logger, rpcTimeout),
		rpcTimeout: rpcTimeout,
	}

	// rpc.Client already registers its own Response and Error handlers internally
	// No need to register duplicate handlers here

	return client
}

// NewRPCClientWithSubprocess creates a new RPC client using an existing subprocess instance
func NewRPCClientWithSubprocess(sp *subprocess.Subprocess, logger *common.Logger, rpcTimeout time.Duration) *RPCClient {
	client := &RPCClient{
		sp:         sp,
		client:     rpc.NewClient(sp, logger, rpcTimeout),
		rpcTimeout: rpcTimeout,
	}

	// The common rpc.Client already registers its own handlers for response and error messages
	// No need to register additional handlers here

	return client
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

// handleResponse handles response messages (for test compatibility)
// This delegates to the common RPC client's HandleResponse method
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	return c.client.HandleResponse(ctx, msg)
}

// handleError handles error messages (for test compatibility)
// This delegates to the common RPC client's HandleError method
func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	return c.client.HandleError(ctx, msg)
}
