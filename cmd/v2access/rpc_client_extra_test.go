package main

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestRPCClient_MarshalErrorAndTimeoutAndHandleResponse(t *testing.T) {
	// create rpc client with tiny timeout
	c := NewRPCClient("test-access", 50*time.Millisecond)

	// set subprocess output to buffer to avoid writing to stdout
	buf := &bytes.Buffer{}
	c.sp.SetOutput(buf)

	// 1) marshal error: params that cannot be marshaled
	_, err := c.InvokeRPCWithTarget(context.Background(), "broker", "TestMethod", make(chan int))
	if err == nil {
		t.Fatalf("expected marshal error, got nil")
	}

	// 2) timeout: send a valid params but no responder
	type P struct {
		A string `json:"a"`
	}
	start := time.Now()
	_, err = c.InvokeRPCWithTarget(context.Background(), "broker", "NoResponder", P{A: "x"})
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatalf("timeout returned too quickly")
	}

	// 3) handleResponse functionality is now tested internally by the common RPC client
	// This part of the test is now covered by the common RPC client's internal tests
}
