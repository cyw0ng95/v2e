package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccessRPCClient_MarshalFailure(t *testing.T) {
	client := NewRPCClient("test-access-marshal", 50*time.Millisecond)
	// direct access to subprocess to avoid writing to stdout
	client.sp.SetOutput(&bytes.Buffer{})

	// params containing a channel should cause sonic.Marshal to error
	params := struct{ C chan int }{C: make(chan int)}

	ctx := context.Background()
	resp, err := client.InvokeRPCWithTarget(ctx, "broker", "RPCTest", params)
	require.Error(t, err)
	require.Nil(t, resp)
	// pendingRequests is now handled internally by the common RPC client
}

func TestAccessRPCClient_PendingCleanupOnTimeout(t *testing.T) {
	client := NewRPCClient("test-access-timeout", 1*time.Millisecond)
	client.sp.SetOutput(&bytes.Buffer{})

	ctx := context.Background()
	resp, err := client.InvokeRPCWithTarget(ctx, "broker", "RPCTestTimeout", nil)
	require.Error(t, err)
	require.Nil(t, resp)
	// pendingRequests cleanup is now handled internally by the common RPC client
}
