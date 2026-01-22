package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
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

	// 3) handleResponse should deliver message to pending entry
	// create a pending entry manually
	respCh := make(chan *subprocess.Message, 1)
	entry := &requestEntry{resp: respCh}
	correlationID := "test-corr-1"
	c.mu.Lock()
	c.pendingRequests[correlationID] = entry
	c.mu.Unlock()

	// call handleResponse with a message that matches correlation id
	msg := &subprocess.Message{CorrelationID: correlationID}
	_, hErr := c.handleResponse(context.Background(), msg)
	if hErr != nil {
		t.Fatalf("handleResponse returned error: %v", hErr)
	}
	select {
	case m := <-respCh:
		if m == nil {
			t.Fatalf("expected non-nil message on channel")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for entry signal")
	}
}
