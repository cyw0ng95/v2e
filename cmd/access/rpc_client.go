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
	// We'll create the subprocess separately and pass it in
	// For now, we'll create it here as before but eventually it should come from standard startup
	var sp *subprocess.Subprocess

	// Create subprocess based on build-time default communication type
	switch subprocess.DefaultProcCommType() {
	case "fd":
		sp = subprocess.NewWithFDs(processID, subprocess.DefaultBuildRPCInputFD(), subprocess.DefaultBuildRPCOutputFD())
	case "uds":
		// Use deterministic UDS path based on build-time base path and process ID
		socketPath := fmt.Sprintf("%s_%s.sock", subprocess.DefaultProcUDSBasePath(), processID)
		sp = subprocess.NewWithUDS(processID, socketPath)
	default:
		sp = subprocess.New(processID)
	}

	// Create a dummy logger for the common RPC client
	logger := common.NewLogger(os.Stderr, "[ACCESS] ", common.InfoLevel)

	client := &RPCClient{
		sp:         sp,
		client:     rpc.NewClient(sp, logger, rpcTimeout),
		rpcTimeout: rpcTimeout,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)

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

// handleResponse handles response messages from the broker
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Delegate to the common client's response handler
	return c.client.HandleResponse(ctx, msg)
}

// handleError handles error messages from the broker (treat them as responses)
func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Delegate to the common client's error handler
	return c.client.HandleError(ctx, msg)
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
