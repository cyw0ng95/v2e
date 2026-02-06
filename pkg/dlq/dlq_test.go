package dlq

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeadLetterQueue(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	assert.NotNil(t, dlq)
	assert.NotNil(t, dlq.messages)
	assert.NotNil(t, dlq.strategyMap)
	assert.Equal(t, 1*time.Second, dlq.baseDelay)
	assert.Equal(t, 5*time.Minute, dlq.maxDelay)
	assert.Equal(t, 2.0, dlq.backoffFactor)
	
	// Verify default strategies configured
	assert.Equal(t, StrategyExponentialBackoff, dlq.strategyMap[ErrorCodeAPITimeout])
	assert.Equal(t, StrategyManual, dlq.strategyMap[ErrorCodeAPIAuthFailed])
	assert.Equal(t, StrategyDiscard, dlq.strategyMap[ErrorCodeDataInvalid])
}

func TestEnqueue(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	msg := &FailedMessage{
		ID:           "msg-001",
		ErrorCode:    ErrorCodeAPITimeout,
		ErrorMessage: "Request timeout",
	}
	
	err := dlq.Enqueue(msg)
	require.NoError(t, err)
	
	stats := dlq.GetStats()
	assert.Equal(t, int64(1), stats.TotalEnqueued)
	assert.Equal(t, 1, stats.CurrentSize)
	
	// Verify message was configured with defaults
	assert.Equal(t, 5, msg.MaxRetries)
	assert.Equal(t, StrategyExponentialBackoff, msg.Strategy)
	assert.NotZero(t, msg.FailedAt)
	assert.NotNil(t, msg.Metadata)
}

func TestEnqueueErrors(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	// Nil message
	err := dlq.Enqueue(nil)
	assert.Error(t, err)
	
	// Empty ID
	err = dlq.Enqueue(&FailedMessage{})
	assert.Error(t, err)
}

func TestConfigureStrategy(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	dlq.ConfigureStrategy(ErrorCodeAPITimeout, StrategyFixed)
	
	assert.Equal(t, StrategyFixed, dlq.strategyMap[ErrorCodeAPITimeout])
}

func TestCalculateNextRetry(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	tests := []struct {
		name         string
		strategy     ReplayStrategy
		retryCount   int
		expectDelay  bool
		minDelay     time.Duration
	}{
		{"Immediate", StrategyImmediate, 0, false, 0},
		{"ExponentialBackoff-First", StrategyExponentialBackoff, 0, true, 1 * time.Second},
		{"ExponentialBackoff-Second", StrategyExponentialBackoff, 1, true, 2 * time.Second},
		{"ExponentialBackoff-Third", StrategyExponentialBackoff, 2, true, 4 * time.Second},
		{"Fixed", StrategyFixed, 0, true, 30 * time.Second},
		{"Manual", StrategyManual, 0, false, 0},
		{"Discard", StrategyDiscard, 0, false, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &FailedMessage{
				Strategy:   tt.strategy,
				RetryCount: tt.retryCount,
			}
			
			nextRetry := dlq.calculateNextRetry(msg)
			
			if tt.expectDelay {
				assert.False(t, nextRetry.IsZero())
				assert.True(t, nextRetry.After(time.Now()))
			} else {
				if tt.strategy == StrategyImmediate {
					// Immediate should be now or very close
					assert.True(t, time.Since(nextRetry) < 1*time.Second)
				} else {
					// Manual and Discard should be zero time
					assert.True(t, nextRetry.IsZero())
				}
			}
		})
	}
}

func TestGetReadyForReplay(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	// Override strategy for immediate message
	dlq.ConfigureStrategy(ErrorCodeAPITimeout, StrategyImmediate)
	
	// Add messages with different retry times
	dlq.Enqueue(&FailedMessage{
		ID:        "msg-1",
		ErrorCode: ErrorCodeAPITimeout,
	})
	
	dlq.Enqueue(&FailedMessage{
		ID:        "msg-2",
		ErrorCode: ErrorCodeAPIAuthFailed, // Manual strategy
	})
	
	dlq.Enqueue(&FailedMessage{
		ID:        "msg-3",
		ErrorCode: ErrorCodeDataInvalid, // Discard strategy
	})
	
	// Wait a moment
	time.Sleep(100 * time.Millisecond)
	
	ready := dlq.GetReadyForReplay()
	
	// Only immediate strategy should be ready
	require.Len(t, ready, 1)
	assert.Equal(t, "msg-1", ready[0].ID)
}

func TestMarkReplayed(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	dlq.Enqueue(&FailedMessage{
		ID:        "msg-1",
		ErrorCode: ErrorCodeAPITimeout,
	})
	
	err := dlq.MarkReplayed("msg-1")
	require.NoError(t, err)
	
	stats := dlq.GetStats()
	assert.Equal(t, int64(1), stats.TotalReplayed)
	assert.Equal(t, 0, stats.CurrentSize)
	
	// Try to get the message
	_, err = dlq.GetMessage("msg-1")
	assert.Error(t, err)
}

func TestMarkFailed(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	dlq.Enqueue(&FailedMessage{
		ID:         "msg-1",
		ErrorCode:  ErrorCodeAPITimeout,
		MaxRetries: 3,
	})
	
	// Mark as failed once
	err := dlq.MarkFailed("msg-1", errors.New("retry failed"))
	require.NoError(t, err)
	
	msg, _ := dlq.GetMessage("msg-1")
	assert.Equal(t, 1, msg.RetryCount)
	assert.Equal(t, "retry failed", msg.ErrorMessage)
	
	// Mark as failed multiple times to exceed max retries
	dlq.MarkFailed("msg-1", errors.New("retry failed 2"))
	dlq.MarkFailed("msg-1", errors.New("retry failed 3"))
	
	msg, _ = dlq.GetMessage("msg-1")
	assert.Equal(t, 3, msg.RetryCount)
	assert.Equal(t, StrategyManual, msg.Strategy) // Should switch to manual
	assert.True(t, msg.NextRetryAt.IsZero())
}

func TestDiscard(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	dlq.Enqueue(&FailedMessage{
		ID:        "msg-1",
		ErrorCode: ErrorCodeAPITimeout,
	})
	
	err := dlq.Discard("msg-1")
	require.NoError(t, err)
	
	stats := dlq.GetStats()
	assert.Equal(t, int64(1), stats.TotalDiscarded)
	assert.Equal(t, 0, stats.CurrentSize)
}

func TestListMessages(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	dlq.Enqueue(&FailedMessage{ID: "msg-1", ErrorCode: ErrorCodeAPITimeout})
	dlq.Enqueue(&FailedMessage{ID: "msg-2", ErrorCode: ErrorCodeDataNotFound})
	dlq.Enqueue(&FailedMessage{ID: "msg-3", ErrorCode: ErrorCodeSystemOverload})
	
	messages := dlq.ListMessages()
	assert.Len(t, messages, 3)
}

func TestStartAutoReplay(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	// Configure immediate strategy
	dlq.ConfigureStrategy(ErrorCodeAPITimeout, StrategyImmediate)
	
	// Add a message that should be replayed immediately
	dlq.Enqueue(&FailedMessage{
		ID:        "msg-1",
		ErrorCode: ErrorCodeAPITimeout,
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	
	handlerCalled := false
	var handlerMu sync.Mutex
	handler := func(msg *FailedMessage) error {
		handlerMu.Lock()
		defer handlerMu.Unlock()
		handlerCalled = true
		assert.Equal(t, "msg-1", msg.ID)
		return nil
	}
	
	// Start auto replay with short interval for testing
	go dlq.StartAutoReplay(ctx, handler, 100*time.Millisecond)
	
	// Wait for replay
	time.Sleep(300 * time.Millisecond)
	
	// Verify handler was called and message was replayed
	handlerMu.Lock()
	called := handlerCalled
	handlerMu.Unlock()
	assert.True(t, called)
	
	stats := dlq.GetStats()
	assert.Equal(t, int64(1), stats.TotalReplayed)
}

func TestExponentialBackoffCap(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	msg := &FailedMessage{
		Strategy:   StrategyExponentialBackoff,
		RetryCount: 100, // Very high retry count
	}
	
	nextRetry := dlq.calculateNextRetry(msg)
	delay := nextRetry.Sub(time.Now())
	
	// Should not exceed max delay
	assert.LessOrEqual(t, delay, dlq.maxDelay+1*time.Second)
}

func TestPowFunction(t *testing.T) {
	tests := []struct {
		base     float64
		exp      int
		expected float64
	}{
		{2.0, 0, 1.0},
		{2.0, 1, 2.0},
		{2.0, 2, 4.0},
		{2.0, 3, 8.0},
		{3.0, 2, 9.0},
	}
	
	for _, tt := range tests {
		result := pow(tt.base, tt.exp)
		assert.Equal(t, tt.expected, result)
	}
}

func TestErrorCodeStrategyMapping(t *testing.T) {
	dlq := NewDeadLetterQueue()
	
	tests := []struct {
		errorCode        ErrorCode
		expectedStrategy ReplayStrategy
	}{
		{ErrorCodeAPITimeout, StrategyExponentialBackoff},
		{ErrorCodeAPIRateLimit, StrategyExponentialBackoff},
		{ErrorCodeAPIServerError, StrategyExponentialBackoff},
		{ErrorCodeAPIAuthFailed, StrategyManual},
		{ErrorCodeDataCorrupted, StrategyManual},
		{ErrorCodeDataNotFound, StrategyFixed},
		{ErrorCodeDataInvalid, StrategyDiscard},
		{ErrorCodeSystemOverload, StrategyExponentialBackoff},
		{ErrorCodeSystemResource, StrategyFixed},
		{ErrorCodeSystemUnknown, StrategyManual},
	}
	
	for _, tt := range tests {
		msg := &FailedMessage{
			ID:        "test",
			ErrorCode: tt.errorCode,
		}
		
		dlq.Enqueue(msg)
		
		retrieved, _ := dlq.GetMessage("test")
		assert.Equal(t, tt.expectedStrategy, retrieved.Strategy,
			"Error code %s should map to strategy %v", tt.errorCode, tt.expectedStrategy)
		
		dlq.Discard("test")
	}
}

func BenchmarkEnqueue(b *testing.B) {
	dlq := NewDeadLetterQueue()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		msg := &FailedMessage{
			ID:        string(rune(i)),
			ErrorCode: ErrorCodeAPITimeout,
		}
		dlq.Enqueue(msg)
	}
}

func BenchmarkGetReadyForReplay(b *testing.B) {
	dlq := NewDeadLetterQueue()
	
	// Pre-populate with messages
	for i := 0; i < 1000; i++ {
		dlq.Enqueue(&FailedMessage{
			ID:        string(rune(i)),
			ErrorCode: ErrorCodeAPITimeout,
			Strategy:  StrategyImmediate,
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = dlq.GetReadyForReplay()
	}
}
