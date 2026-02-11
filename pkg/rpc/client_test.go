package rpc

import (
	"context"
	"fmt"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
	"io"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestDefaultRPCTimeout(t *testing.T) {
	tests := []struct {
		name     string
		expected time.Duration
	}{
		{
			name:     "default timeout is 30 seconds",
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if DefaultRPCTimeout != tt.expected {
				t.Errorf("DefaultRPCTimeout = %v, want %v", DefaultRPCTimeout, tt.expected)
			}
		})
	}
}

func TestGetDefaultTimeout(t *testing.T) {
	tests := []struct {
		name     string
		expected time.Duration
	}{
		{
			name:     "returns 30 seconds",
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultTimeout()
			if got != tt.expected {
				t.Errorf("GetDefaultTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetDefaultTimeoutConsistency(t *testing.T) {
	// Test that GetDefaultTimeout consistently returns the same value
	testutils.Run(t, testutils.Level2, "consistent return values", nil, func(t *testing.T, tx *gorm.DB) {
		first := GetDefaultTimeout()
		second := GetDefaultTimeout()
		third := GetDefaultTimeout()

		if first != second || second != third {
			t.Errorf("GetDefaultTimeout() returns inconsistent values: %v, %v, %v", first, second, third)
		}
	})
}

func TestGetDefaultTimeoutMatchesConstant(t *testing.T) {
	// Test that GetDefaultTimeout returns the same value as DefaultRPCTimeout constant
	testutils.Run(t, testutils.Level2, "matches constant", nil, func(t *testing.T, tx *gorm.DB) {
		if GetDefaultTimeout() != DefaultRPCTimeout {
			t.Errorf("GetDefaultTimeout() = %v, DefaultRPCTimeout = %v (should be equal)", GetDefaultTimeout(), DefaultRPCTimeout)
		}
	})
}

func TestNewClientWithDefaultTimeout(t *testing.T) {
	// Test that NewClient can be initialized with the default timeout
	testutils.Run(t, testutils.Level2, "client creation with default timeout", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(nil, "", common.InfoLevel)
		sp := subprocess.New("test-service")
		client := NewClient(sp, logger, GetDefaultTimeout())

		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
		// Note: We can't directly access client.rpcTimeout as it's unexported,
		// but we've verified the client was created successfully
	})
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name          string
		setupPending  bool
		correlationID string
		expectSignal  bool
	}{
		{
			name:          "response for pending request",
			setupPending:  true,
			correlationID: "test-correlation-1",
			expectSignal:  true,
		},
		{
			name:          "response for unknown correlation ID",
			setupPending:  false,
			correlationID: "test-correlation-2",
			expectSignal:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := common.NewLogger(io.Discard, "", common.InfoLevel)
			sp := subprocess.New("test-service")
			client := NewClient(sp, logger, GetDefaultTimeout())

			var entry *RequestEntry
			if tt.setupPending {
				entry = &RequestEntry{
					resp: make(chan *subprocess.Message, 1),
				}
				client.pendingRequests[tt.correlationID] = entry
			}

			msg := &subprocess.Message{
				CorrelationID: tt.correlationID,
				Type:          subprocess.MessageTypeResponse,
			}

			ctx := context.Background()
			respMsg, err := client.HandleResponse(ctx, msg)

			if err != nil {
				t.Errorf("HandleResponse() error = %v, want nil", err)
			}

			if respMsg != nil {
				t.Error("HandleResponse() should return nil message")
			}

			if tt.setupPending && tt.expectSignal {
				select {
				case sigMsg := <-entry.resp:
					if sigMsg != msg {
						t.Errorf("Signal received wrong message")
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("HandleResponse() failed to signal pending request")
				}
			}

			if _, exists := client.pendingRequests[tt.correlationID]; exists {
				t.Error("HandleResponse() should have removed pending request from map")
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name          string
		setupPending  bool
		correlationID string
		errorMsg      string
	}{
		{
			name:          "error message for pending request",
			setupPending:  true,
			correlationID: "test-correlation-1",
			errorMsg:      "test error",
		},
		{
			name:          "error message for unknown correlation ID",
			setupPending:  false,
			correlationID: "test-correlation-2",
			errorMsg:      "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := common.NewLogger(io.Discard, "", common.InfoLevel)
			sp := subprocess.New("test-service")
			client := NewClient(sp, logger, GetDefaultTimeout())

			var entry *RequestEntry
			if tt.setupPending {
				entry = &RequestEntry{
					resp: make(chan *subprocess.Message, 1),
				}
				client.pendingRequests[tt.correlationID] = entry
			}

			msg := &subprocess.Message{
				CorrelationID: tt.correlationID,
				Type:          subprocess.MessageTypeError,
				Error:         tt.errorMsg,
			}

			ctx := context.Background()
			respMsg, err := client.HandleError(ctx, msg)

			if err != nil {
				t.Errorf("HandleError() error = %v, want nil", err)
			}

			if respMsg != nil {
				t.Error("HandleError() should return nil message")
			}

			if tt.setupPending {
				select {
				case sigMsg := <-entry.resp:
					if sigMsg != msg {
						t.Errorf("Signal received wrong message")
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("HandleError() failed to signal pending request")
				}
			}
		})
	}
}

func TestInvokeRPC(t *testing.T) {
	tests := []struct {
		name          string
		target        string
		method        string
		params        interface{}
		expectSuccess bool
	}{
		{
			name:          "invoke RPC with nil params",
			target:        "test-service",
			method:        "TestMethod",
			params:        nil,
			expectSuccess: false,
		},
		{
			name:          "invoke RPC with params",
			target:        "test-service",
			method:        "TestMethod",
			params:        map[string]string{"key": "value"},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := common.NewLogger(io.Discard, "", common.InfoLevel)
			sp := subprocess.New("test-service")
			client := NewClient(sp, logger, 100*time.Millisecond)

			ctx := context.Background()
			respMsg, err := client.InvokeRPC(ctx, tt.target, tt.method, tt.params)

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("InvokeRPC() error = %v, want nil", err)
				}
				if respMsg == nil {
					t.Error("InvokeRPC() should return response message")
				}
			} else {
				if err == nil {
					t.Error("InvokeRPC() should return error (no broker connection)")
				}
			}
		})
	}
}

func TestInvokeRPC_ContextCanceled(t *testing.T) {
	logger := common.NewLogger(io.Discard, "", common.InfoLevel)
	sp := subprocess.New("test-service")
	client := NewClient(sp, logger, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.InvokeRPC(ctx, "test-service", "TestMethod", nil)
	if err == nil {
		t.Error("InvokeRPC() should return error for canceled context")
	}
}

func BenchmarkHandleResponse(b *testing.B) {
	logger := common.NewLogger(nil, "", common.InfoLevel)
	sp := subprocess.New("test-service")
	client := NewClient(sp, logger, GetDefaultTimeout())

	entry := &RequestEntry{
		resp: make(chan *subprocess.Message, 1),
	}
	client.pendingRequests["test-correlation"] = entry

	msg := &subprocess.Message{
		CorrelationID: "test-correlation",
		Type:          subprocess.MessageTypeResponse,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.HandleResponse(context.Background(), msg)
	}
}

func BenchmarkHandleError(b *testing.B) {
	logger := common.NewLogger(nil, "", common.InfoLevel)
	sp := subprocess.New("test-service")
	client := NewClient(sp, logger, GetDefaultTimeout())

	entry := &RequestEntry{
		resp: make(chan *subprocess.Message, 1),
	}
	client.pendingRequests["test-correlation"] = entry

	msg := &subprocess.Message{
		CorrelationID: "test-correlation",
		Type:          subprocess.MessageTypeError,
		Error:         "test error",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.HandleError(context.Background(), msg)
	}
}

func BenchmarkGetDefaultTimeout(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetDefaultTimeout()
	}
}

// TestPendingRequestMap_NoLeakOnTimeout verifies that pending requests are properly
// cleaned up from the map when a request times out. This test prevents memory leaks
// by ensuring the map doesn't grow unbounded.
func TestPendingRequestMap_NoLeakOnTimeout(t *testing.T) {
	logger := common.NewLogger(io.Discard, "", common.InfoLevel)
	sp := subprocess.New("test-service")
	client := NewClient(sp, logger, 50*time.Millisecond)

	// Track initial map size
	initialSize := len(client.pendingRequests)

	// Make a request that will timeout (no broker to respond)
	ctx := context.Background()
	_, err := client.InvokeRPC(ctx, "test-service", "TestMethod", map[string]string{"key": "value"})

	// Should get timeout error
	if err == nil {
		t.Error("InvokeRPC() should return timeout error")
	}

	// Wait a bit for defer to complete
	time.Sleep(100 * time.Millisecond)

	// Verify map size hasn't grown (entry was cleaned up)
	currentSize := len(client.pendingRequests)
	if currentSize != initialSize {
		t.Errorf("Pending request map leaked: expected %d entries, got %d", initialSize, currentSize)
	}
}

// TestPendingRequestMap_NoLeakOnContextCancel verifies that pending requests are
// properly cleaned up when the context is canceled.
func TestPendingRequestMap_NoLeakOnContextCancel(t *testing.T) {
	logger := common.NewLogger(io.Discard, "", common.InfoLevel)
	sp := subprocess.New("test-service")
	client := NewClient(sp, logger, 10*time.Second)

	// Track initial map size
	initialSize := len(client.pendingRequests)

	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Make a request with canceled context
	_, err := client.InvokeRPC(ctx, "test-service", "TestMethod", nil)

	// Should get context canceled error
	if err == nil {
		t.Error("InvokeRPC() should return context canceled error")
	}

	// Wait a bit for defer to complete
	time.Sleep(100 * time.Millisecond)

	// Verify map size hasn't grown (entry was cleaned up)
	currentSize := len(client.pendingRequests)
	if currentSize != initialSize {
		t.Errorf("Pending request map leaked on context cancel: expected %d entries, got %d", initialSize, currentSize)
	}
}

// TestPendingRequestMap_NoLeakOnConcurrentRequests verifies that multiple concurrent
// requests don't cause memory leaks in the pending request map.
func TestPendingRequestMap_NoLeakOnConcurrentRequests(t *testing.T) {
	logger := common.NewLogger(io.Discard, "", common.InfoLevel)
	sp := subprocess.New("test-service")
	client := NewClient(sp, logger, 100*time.Millisecond)

	// Track initial map size
	initialSize := len(client.pendingRequests)

	// Launch multiple concurrent requests that will timeout
	const numRequests = 10
	done := make(chan struct{}, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			ctx := context.Background()
			_, _ = client.InvokeRPC(ctx, "test-service", fmt.Sprintf("TestMethod%d", idx), nil)
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for concurrent requests")
		}
	}

	// Wait a bit for all defer functions to complete
	time.Sleep(200 * time.Millisecond)

	// Verify map size hasn't grown (all entries were cleaned up)
	currentSize := len(client.pendingRequests)
	if currentSize != initialSize {
		t.Errorf("Pending request map leaked after %d concurrent requests: expected %d entries, got %d",
			numRequests, initialSize, currentSize)
	}
}

// TestRequestEntry_ConcurrentCloseAndSignal verifies that concurrent calls to
// Close() and Signal() don't cause panics or deadlocks.
func TestRequestEntry_ConcurrentCloseAndSignal(t *testing.T) {
	entry := &RequestEntry{
		resp: make(chan *subprocess.Message, 1),
	}

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		CorrelationID: "test-123",
	}

	// Launch concurrent Close and Signal calls
	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			entry.Close()
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			entry.Signal(msg)
		}
	}()

	// Wait for both goroutines
	select {
	case <-done:
		select {
		case <-done:
			// Both completed
		case <-time.After(time.Second):
			t.Fatal("Timed out waiting for goroutines")
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for goroutines")
	}

	// Verify channel is closed (due to sync.Once, either Close or Signal executed)
	// The channel should be closed regardless of which function ran first
	select {
	case v, ok := <-entry.resp:
		// Channel is closed when ok is false
		// If ok is true, we received the message and channel should still be closed
		if ok && v != nil {
			// Signal() executed first and sent message
			// Channel should now be closed
		}
		// If ok is false, Close() executed first and closed channel
		// Either way, no panic or deadlock occurred
	case <-time.After(100 * time.Millisecond):
		t.Error("Channel should be closed (select should not block)")
	}
}

// TestRequestEntry_DoubleSignal verifies that calling Signal() multiple times
// only sends the message once (due to sync.Once).
func TestRequestEntry_DoubleSignal(t *testing.T) {
	entry := &RequestEntry{
		resp: make(chan *subprocess.Message, 2), // Buffered for 2
	}

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		CorrelationID: "test-123",
	}

	// Signal twice
	entry.Signal(msg)
	entry.Signal(msg)

	// Should only receive one message
	select {
	case receivedMsg := <-entry.resp:
		// First message received
		if receivedMsg != msg {
			t.Error("Received wrong message")
		}
	default:
		t.Error("Expected to receive first message")
	}

	// Channel should be closed now (Signal closes it after sending)
	// Try to receive again - should get zero value immediately due to closed channel
	receivedMsg, ok := <-entry.resp
	if ok {
		t.Error("Channel should be closed after Signal, but received another message")
	}
	if receivedMsg != nil {
		t.Error("Should not receive a second message from closed channel")
	}
}

// TestRequestEntry_DoubleClose verifies that calling Close() multiple times
// doesn't panic (idempotent operation).
func TestRequestEntry_DoubleClose(t *testing.T) {
	entry := &RequestEntry{
		resp: make(chan *subprocess.Message, 1),
	}

	// Close twice - should not panic
	entry.Close()
	entry.Close()

	// Verify channel is closed
	_, ok := <-entry.resp
	if ok {
		t.Error("Response channel should be closed")
	}
}
