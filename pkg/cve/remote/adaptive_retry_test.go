package remote

import (
	"errors"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestAdaptiveRetry_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAdaptiveRetry_Success", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()
		attempts := 0

		err := ar.Execute(func() error {
			attempts++
			return nil
		}, PriorityNormal)

		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})
}

func TestAdaptiveRetry_RetryableError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAdaptiveRetry_RetryableError", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()
		attempts := 0

		err := ar.Execute(func() error {
			attempts++
			if attempts < 2 {
				return ErrRateLimited
			}
			return nil
		}, PriorityNormal)

		if err != nil {
			t.Errorf("Expected success after retry, got error: %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})
}

func TestAdaptiveRetry_MaxRetriesExceeded(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAdaptiveRetry_MaxRetriesExceeded", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()

		err := ar.Execute(func() error {
			return ErrRateLimited
		}, PriorityNormal)

		if !errors.Is(err, ErrMaxRetriesExceeded) {
			t.Errorf("Expected ErrMaxRetriesExceeded, got: %v", err)
		}
	})
}

func TestAdaptiveRetry_NonRetryableError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAdaptiveRetry_NonRetryableError", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()
		attempts := 0
		expectedErr := errors.New("non-retryable error")

		err := ar.Execute(func() error {
			attempts++
			return expectedErr
		}, PriorityNormal)

		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected non-retryable error, got: %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
		}
	})
}

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCircuitBreaker_ClosedToOpen", nil, func(t *testing.T, _ *gorm.DB) {
		config := CircuitBreakerConfig{
			FailureThreshold: 3,
			Timeout:          100 * time.Millisecond,
		}
		cb := NewCircuitBreaker(config)

		// Record failures
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		if cb.GetState() != StateOpen {
			t.Errorf("Expected circuit breaker to be open, got: %s", cb.GetState())
		}

		if cb.CanExecute() {
			t.Error("Expected CanExecute to return false when open")
		}
	})
}

func TestCircuitBreaker_OpenToHalfOpen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCircuitBreaker_OpenToHalfOpen", nil, func(t *testing.T, _ *gorm.DB) {
		config := CircuitBreakerConfig{
			FailureThreshold: 3,
			Timeout:          50 * time.Millisecond,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		// Wait for timeout
		time.Sleep(60 * time.Millisecond)

		// Check if can execute (should be half-open)
		if !cb.CanExecute() {
			t.Error("Expected CanExecute to return true in half-open state")
		}
	})
}

func TestCircuitBreaker_HalfOpenToClosed(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCircuitBreaker_HalfOpenToClosed", nil, func(t *testing.T, _ *gorm.DB) {
		config := CircuitBreakerConfig{
			FailureThreshold: 3,
			SuccessThreshold: 2,
			Timeout:          50 * time.Millisecond,
			HalfOpenRequests: 5,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		// Wait for timeout
		time.Sleep(60 * time.Millisecond)

		// Allow execution
		cb.CanExecute()

		// Record successes
		for i := 0; i < 2; i++ {
			cb.RecordSuccess()
		}

		if cb.GetState() != StateClosed {
			t.Errorf("Expected circuit breaker to be closed, got: %s", cb.GetState())
		}

		if !cb.CanExecute() {
			t.Error("Expected CanExecute to return true in closed state")
		}
	})
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCircuitBreaker_HalfOpenToOpen", nil, func(t *testing.T, _ *gorm.DB) {
		config := CircuitBreakerConfig{
			FailureThreshold: 3,
			SuccessThreshold: 2,
			Timeout:          50 * time.Millisecond,
			HalfOpenRequests: 5,
		}
		cb := NewCircuitBreaker(config)

		// Trigger open state
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		// Wait for timeout
		time.Sleep(60 * time.Millisecond)

		// Allow execution
		cb.CanExecute()

		// Record a failure
		cb.RecordFailure()

		if cb.GetState() != StateOpen {
			t.Errorf("Expected circuit breaker to be open, got: %s", cb.GetState())
		}
	})
}

func TestRetryMetrics(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRetryMetrics", nil, func(t *testing.T, _ *gorm.DB) {
		rm := NewRetryMetrics()

		// Record some data
		rm.RecordAttempt(PriorityNormal)
		rm.RecordAttempt(PriorityHigh)
		rm.RecordSuccess()
		rm.RecordFailure()
		rm.RecordRetry(1 * time.Second)

		stats := rm.GetStats()

		if stats["total_attempts"].(int) != 2 {
			t.Errorf("Expected 2 attempts, got: %v", stats["total_attempts"])
		}
		if stats["total_successes"].(int) != 1 {
			t.Errorf("Expected 1 success, got: %v", stats["total_successes"])
		}
		if stats["total_failures"].(int) != 1 {
			t.Errorf("Expected 1 failure, got: %v", stats["total_failures"])
		}
		if stats["total_retries"].(int) != 1 {
			t.Errorf("Expected 1 retry, got: %v", stats["total_retries"])
		}
	})
}

func TestRetryMetrics_SuccessRate(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRetryMetrics_SuccessRate", nil, func(t *testing.T, _ *gorm.DB) {
		rm := NewRetryMetrics()

		rm.RecordAttempt(PriorityNormal)
		rm.RecordAttempt(PriorityNormal)
		rm.RecordSuccess()

		rate := rm.GetSuccessRate()
		expected := 0.5
		if rate != expected {
			t.Errorf("Expected success rate %f, got %f", expected, rate)
		}
	})
}

func TestRetryMetrics_RetryRate(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRetryMetrics_RetryRate", nil, func(t *testing.T, _ *gorm.DB) {
		rm := NewRetryMetrics()

		rm.RecordAttempt(PriorityNormal)
		rm.RecordAttempt(PriorityNormal)
		rm.RecordRetry(1 * time.Second)

		rate := rm.GetRetryRate()
		expected := 0.5
		if rate != expected {
			t.Errorf("Expected retry rate %f, got %f", expected, rate)
		}
	})
}

func TestRetryMetrics_AverageBackoffTime(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRetryMetrics_AverageBackoffTime", nil, func(t *testing.T, _ *gorm.DB) {
		rm := NewRetryMetrics()

		rm.RecordRetry(1 * time.Second)
		rm.RecordRetry(2 * time.Second)
		rm.RecordRetry(3 * time.Second)

		avg := rm.GetAverageBackoffTime()
		expected := 2 * time.Second
		if avg != expected {
			t.Errorf("Expected average backoff time %v, got %v", expected, avg)
		}
	})
}

func TestExponentialBackoff(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExponentialBackoff", nil, func(t *testing.T, _ *gorm.DB) {
		config := DefaultRetryConfig()
		config.JitterEnabled = false
		ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

		backoff0 := ar.calculateBackoff(0)
		backoff1 := ar.calculateBackoff(1)
		backoff2 := ar.calculateBackoff(2)

		if backoff0 != ar.config.InitialDelay {
			t.Errorf("Expected backoff0 %v, got %v", ar.config.InitialDelay, backoff0)
		}
		if backoff1 != ar.config.InitialDelay*2 {
			t.Errorf("Expected backoff1 %v, got %v", ar.config.InitialDelay*2, backoff1)
		}
		if backoff2 != ar.config.InitialDelay*4 {
			t.Errorf("Expected backoff2 %v, got %v", ar.config.InitialDelay*4, backoff2)
		}
	})
}

func TestExponentialBackoffWithJitter(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExponentialBackoffWithJitter", nil, func(t *testing.T, _ *gorm.DB) {
		config := DefaultRetryConfig()
		config.JitterEnabled = true
		config.JitterFactor = 0.5
		ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

		backoffs := make(map[time.Duration]bool)
		for i := 0; i < 10; i++ {
			backoff := ar.calculateBackoff(1)
			backoffs[backoff] = true
		}

		if len(backoffs) == 1 {
			t.Error("Expected jitter to produce varying backoff times")
		}
	})
}

func TestLinearBackoff(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLinearBackoff", nil, func(t *testing.T, _ *gorm.DB) {
		config := DefaultRetryConfig()
		config.Strategy = StrategyLinear
		config.JitterEnabled = false
		ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

		backoff0 := ar.calculateBackoff(0)
		backoff1 := ar.calculateBackoff(1)
		backoff2 := ar.calculateBackoff(2)

		if backoff0 != ar.config.InitialDelay {
			t.Errorf("Expected backoff0 %v, got %v", ar.config.InitialDelay, backoff0)
		}
		if backoff1 != ar.config.InitialDelay*2 {
			t.Errorf("Expected backoff1 %v, got %v", ar.config.InitialDelay*2, backoff1)
		}
		if backoff2 != ar.config.InitialDelay*3 {
			t.Errorf("Expected backoff2 %v, got %v", ar.config.InitialDelay*3, backoff2)
		}
	})
}

func TestFixedBackoff(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestFixedBackoff", nil, func(t *testing.T, _ *gorm.DB) {
		config := DefaultRetryConfig()
		config.Strategy = StrategyFixed
		config.JitterEnabled = false
		ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

		backoff0 := ar.calculateBackoff(0)
		backoff1 := ar.calculateBackoff(1)
		backoff2 := ar.calculateBackoff(2)

		if backoff0 != ar.config.InitialDelay {
			t.Errorf("Expected backoff0 %v, got %v", ar.config.InitialDelay, backoff0)
		}
		if backoff1 != ar.config.InitialDelay {
			t.Errorf("Expected backoff1 %v, got %v", ar.config.InitialDelay, backoff1)
		}
		if backoff2 != ar.config.InitialDelay {
			t.Errorf("Expected backoff2 %v, got %v", ar.config.InitialDelay, backoff2)
		}
	})
}

func TestPriorityBasedRetry(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPriorityBasedRetry", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()

		// Execute requests with different priorities
		ar.Execute(func() error { return nil }, PriorityLow)
		ar.Execute(func() error { return nil }, PriorityNormal)
		ar.Execute(func() error { return nil }, PriorityHigh)
		ar.Execute(func() error { return nil }, PriorityCritical)

		stats := ar.GetMetrics().GetStats()
		priorityBreakdown := stats["priority_breakdown"].(map[RequestPriority]int)

		if priorityBreakdown[PriorityLow] != 1 {
			t.Errorf("Expected 1 low priority request, got %d", priorityBreakdown[PriorityLow])
		}
		if priorityBreakdown[PriorityNormal] != 1 {
			t.Errorf("Expected 1 normal priority request, got %d", priorityBreakdown[PriorityNormal])
		}
		if priorityBreakdown[PriorityHigh] != 1 {
			t.Errorf("Expected 1 high priority request, got %d", priorityBreakdown[PriorityHigh])
		}
		if priorityBreakdown[PriorityCritical] != 1 {
			t.Errorf("Expected 1 critical priority request, got %d", priorityBreakdown[PriorityCritical])
		}
	})
}

func TestAdaptiveRetry_WithCircuitBreakerOpen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAdaptiveRetry_WithCircuitBreakerOpen", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()

		// Trip circuit breaker
		for i := 0; i < 10; i++ {
			ar.Execute(func() error {
				return ErrRateLimited
			}, PriorityNormal)
		}

		// Next request should fail with circuit breaker error
		err := ar.Execute(func() error {
			return nil
		}, PriorityNormal)

		if !errors.Is(err, ErrCircuitBreakerOpen) {
			t.Errorf("Expected ErrCircuitBreakerOpen, got: %v", err)
		}
	})
}
