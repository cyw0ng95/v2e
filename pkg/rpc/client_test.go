package rpc

import (
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
	t.Run("consistent return values", func(t *testing.T) {
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
	t.Run("matches constant", func(t *testing.T) {
		if GetDefaultTimeout() != DefaultRPCTimeout {
			t.Errorf("GetDefaultTimeout() = %v, DefaultRPCTimeout = %v (should be equal)", GetDefaultTimeout(), DefaultRPCTimeout)
		}
	})
}

func TestNewClientWithDefaultTimeout(t *testing.T) {
	// Test that NewClient can be initialized with the default timeout
	t.Run("client creation with default timeout", func(t *testing.T) {
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

func BenchmarkGetDefaultTimeout(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetDefaultTimeout()
	}
}
