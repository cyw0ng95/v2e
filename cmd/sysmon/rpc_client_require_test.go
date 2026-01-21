package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestSysmonRPCClient_InvokeRPC_ContextCanceled(t *testing.T) {
	sp := subprocess.New("test-sysmon")
	// Ensure sending doesn't write to real stdout and disables batching
	sp.SetOutput(&bytes.Buffer{})

	client := NewRPCClient(sp)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediate cancel

	resp, err := client.InvokeRPC(ctx, "broker", "RPCTestMethod", nil)
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, resp)
	require.Len(t, client.pendingRequests, 0)
}
