package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRPCClientHandleResponse_UnknownCorrelation(t *testing.T) {
	logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
	sp := subprocess.New("meta-test")
	client := NewRPCClient(sp, logger)

	// Response with no pending entry should not panic and should warn only.
	if _, err := client.handleResponse(context.Background(), &subprocess.Message{CorrelationID: "missing"}); err != nil {
		t.Fatalf("handleResponse returned error: %v", err)
	}
}

func TestRPCClientInvokeRPC_SendError(t *testing.T) {
	logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
	sp := subprocess.New("meta-test")
	sp.SetOutput(errorWriter{})
	client := NewRPCClient(sp, logger)

	if _, err := client.InvokeRPC(context.Background(), "target", "Method", nil); err == nil {
		t.Fatalf("expected send error")
	}
}

func TestRPCClientInvokeRPC_Timeout(t *testing.T) {
	logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
	sp := subprocess.New("meta-test")
	// Use a writer that never receives a response to force timeout
	sp.SetOutput(&bytes.Buffer{})
	client := NewRPCClient(sp, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	if _, err := client.InvokeRPC(ctx, "target", "Method", nil); err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestRPCClientInvokeRPC_ContextCanceled(t *testing.T) {
	logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
	sp := subprocess.New("meta-test")
	sp.SetOutput(&bytes.Buffer{})
	client := NewRPCClient(sp, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.InvokeRPC(ctx, "target", "Method", nil)
	if err == nil {
		t.Fatalf("expected context cancellation error")
	}
	client.mu.RLock()
	defer client.mu.RUnlock()
	if len(client.pendingRequests) != 0 {
		t.Fatalf("pendingRequests not cleared on cancel: %d", len(client.pendingRequests))
	}
}

func TestRPCClientInvokeRPC_MarshalError(t *testing.T) {
	logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
	sp := subprocess.New("meta-test")
	sp.SetOutput(&bytes.Buffer{})
	client := NewRPCClient(sp, logger)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Channels are not JSON-serializable; MarshalFast should fail
	_, err := client.InvokeRPC(ctx, "target", "Method", make(chan int))
	if err == nil {
		t.Fatalf("expected marshal error")
	}
	client.mu.RLock()
	defer client.mu.RUnlock()
	if len(client.pendingRequests) != 0 {
		t.Fatalf("pendingRequests not cleared after marshal error: %d", len(client.pendingRequests))
	}
}

// errorWriter always returns an error to trigger SendMessage failure.
type errorWriter struct{}

func (errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

// testWriter adapts testing.T to io.Writer for logger output suppression.
type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Logf("%s", string(p))
	return len(p), nil
}
