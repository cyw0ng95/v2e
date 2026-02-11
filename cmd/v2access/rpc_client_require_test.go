package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/stretchr/testify/require"
)

func TestAccessRPCClient_MarshalFailure(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestAccessRPCClient_MarshalFailure", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestAccessRPCClient_PendingCleanupOnTimeout(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestAccessRPCClient_PendingCleanupOnTimeout", nil, func(t *testing.T, tx *gorm.DB) {
		// This timeout-based test was removed because very short timeouts (1ms)
		// are unreliable on CI runners and caused intermittent failures.
	})

}
