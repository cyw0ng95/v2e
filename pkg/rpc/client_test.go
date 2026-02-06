package rpc

import (
	"context"
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
