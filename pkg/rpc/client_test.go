package rpc

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
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

func BenchmarkGetDefaultTimeout(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetDefaultTimeout()
	}
}
