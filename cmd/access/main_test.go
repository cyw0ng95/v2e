package main

import (
	"context"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// TODO: Tests disabled - access service is currently a stub
// Access needs to be redesigned to communicate with broker via RPC
// instead of instantiating its own broker.
// See: https://github.com/cyw0ng95/v2e/pull/74

func TestHealthEndpoint(t *testing.T) {
	// Basic test placeholder
	t.Skip("Access service is currently a stub pending redesign")
}

func TestNewRPCClient_Access(t *testing.T) {
	client := NewRPCClient("test-access")
	if client == nil {
		t.Fatal("NewRPCClient returned nil")
	}
	if client.sp == nil {
		t.Fatal("RPCClient subprocess is nil")
	}
	if client.sp.ID != "test-access" {
		t.Fatalf("expected subprocess ID 'test-access', got '%s'", client.sp.ID)
	}
}

func TestRPCClient_HandleResponse_UnknownCorrelation(t *testing.T) {
	client := NewRPCClient("test-access-2")
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "m",
		CorrelationID: "nonexistent-corr",
	}
	ctx := context.Background()
	resp, err := client.handleResponse(ctx, msg)
	if err != nil {
		t.Fatalf("handleResponse returned error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response from handleResponse for unknown correlation, got %v", resp)
	}
}

func TestInvokeRPCWithTarget_ResponseDelivered(t *testing.T) {
	client := NewRPCClient("test-access-3")
	ctx := context.Background()

	// Start InvokeRPCWithTarget in a goroutine so we can deliver a response
	done := make(chan struct{})
	var respMsg *subprocess.Message
	var invokeErr error
	go func() {
		respMsg, invokeErr = client.InvokeRPCWithTarget(ctx, "broker", "RPCTestMethod", map[string]string{"k": "v"})
		close(done)
	}()

	// Wait briefly for the pendingRequests map to be populated by InvokeRPCWithTarget
	time.Sleep(10 * time.Millisecond)

	// Simulate broker sending a response by calling the client's handler directly
	// The correlation ID uses the internal sequence - first call will be access-rpc-1
	sim := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "RPCTestMethod",
		CorrelationID: "access-rpc-1",
		Payload:       []byte(`"ok"`),
	}
	if _, err := client.handleResponse(ctx, sim); err != nil {
		t.Fatalf("handleResponse simulation returned error: %v", err)
	}

	// Wait for InvokeRPCWithTarget to return
	select {
	case <-done:
		// proceed
	case <-time.After(1 * time.Second):
		t.Fatal("InvokeRPCWithTarget did not return in time")
	}

	if invokeErr != nil {
		t.Fatalf("InvokeRPCWithTarget returned error: %v", invokeErr)
	}
	if respMsg == nil {
		t.Fatal("expected response message, got nil")
	}
	if respMsg.CorrelationID != "access-rpc-1" {
		t.Fatalf("expected correlation id access-rpc-1, got %s", respMsg.CorrelationID)
	}
}
