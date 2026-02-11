package remote

import (
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// BenchmarkAdaptiveRetry_Success benchmarks successful requests without retry
func BenchmarkAdaptiveRetry_Success(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ar := NewAdaptiveRetryWithDefaults()
		err := ar.Execute(func() error {
			return nil
		}, PriorityNormal)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAdaptiveRetry_WithRetry benchmarks requests that need retry
func BenchmarkAdaptiveRetry_WithRetry(b *testing.B) {
	b.ReportAllocs()
	attempts := 0
	for i := 0; i < b.N; i++ {
		ar := NewAdaptiveRetryWithDefaults()
		err := ar.Execute(func() error {
			attempts++
			if attempts%2 == 0 {
				return ErrRateLimited
			}
			return nil
		}, PriorityNormal)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCircuitBreaker_Closed benchmarks circuit breaker in closed state
func BenchmarkCircuitBreaker_Closed(b *testing.B) {
	config := CircuitBreakerConfig{
		FailureThreshold: 100,
	}
	cb := NewCircuitBreaker(config)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.CanExecute()
	}
}

// BenchmarkRetryMetrics_Record benchmarks metrics recording
func BenchmarkRetryMetrics_Record(b *testing.B) {
	rm := NewRetryMetrics()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.RecordAttempt(PriorityNormal)
		rm.RecordSuccess()
	}
}

// BenchmarkRetryMetrics_GetStats benchmarks statistics retrieval
func BenchmarkRetryMetrics_GetStats(b *testing.B) {
	rm := NewRetryMetrics()
	for i := 0; i < 1000; i++ {
		rm.RecordAttempt(PriorityNormal)
		rm.RecordSuccess()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rm.GetStats()
	}
}

// BenchmarkAdaptiveRetry_WithCircuitBreaker benchmarks retry with circuit breaker
func BenchmarkAdaptiveRetry_WithCircuitBreaker(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ar := NewAdaptiveRetryWithDefaults()
		err := ar.Execute(func() error {
			return nil
		}, PriorityNormal)
		if err != nil {
			if !errors.Is(err, ErrCircuitBreakerOpen) {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkBackoff_Calculation benchmarks backoff calculation
func BenchmarkBackoff_Calculation(b *testing.B) {
	config := DefaultRetryConfig()
	ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for attempt := 0; attempt < 10; attempt++ {
			_ = ar.calculateBackoff(attempt, nil)
		}
	}
}

// BenchmarkExponentialBackoff benchmarks exponential backoff strategy
func BenchmarkExponentialBackoff(b *testing.B) {
	config := DefaultRetryConfig()
	config.Strategy = StrategyExponential
	config.JitterEnabled = false
	ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for attempt := 0; attempt < 10; attempt++ {
			_ = ar.calculateBackoff(attempt, nil)
		}
	}
}

// BenchmarkLinearBackoff benchmarks linear backoff strategy
func BenchmarkLinearBackoff(b *testing.B) {
	config := DefaultRetryConfig()
	config.Strategy = StrategyLinear
	config.JitterEnabled = false
	ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for attempt := 0; attempt < 10; attempt++ {
			_ = ar.calculateBackoff(attempt, nil)
		}
	}
}

// BenchmarkFixedBackoff benchmarks fixed backoff strategy
func BenchmarkFixedBackoff(b *testing.B) {
	config := DefaultRetryConfig()
	config.Strategy = StrategyFixed
	config.JitterEnabled = false
	ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for attempt := 0; attempt < 10; attempt++ {
			_ = ar.calculateBackoff(attempt, nil)
		}
	}
}

// BenchmarkBackoff_WithJitter benchmarks backoff calculation with jitter
func BenchmarkBackoff_WithJitter(b *testing.B) {
	config := DefaultRetryConfig()
	config.JitterEnabled = true
	ar := NewAdaptiveRetry(config, DefaultCircuitBreakerConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for attempt := 0; attempt < 10; attempt++ {
			_ = ar.calculateBackoff(attempt, nil)
		}
	}
}

// TestAdaptiveRetry_SuccessRate benchmarks and measures success rate
func TestAdaptiveRetry_SuccessRate(t *testing.T) {
	testutils.Run(t, testutils.Level3, "TestAdaptiveRetry_SuccessRate", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()

		// Execute 100 requests with 10% failure rate
		for i := 0; i < 100; i++ {
			_ = ar.Execute(func() error {
				if i%10 == 0 {
					return ErrRateLimited
				}
				return nil
			}, PriorityNormal)
		}

		successRate := ar.GetMetrics().GetSuccessRate()
		t.Logf("Success rate: %.2f%%", successRate*100)

		// Expected success rate around 90%
		if successRate < 0.85 || successRate > 0.95 {
			t.Errorf("Expected success rate around 0.9, got %.2f", successRate)
		}
	})
}

// TestAdaptiveRetry_RetryOverhead measures retry overhead
func TestAdaptiveRetry_RetryOverhead(t *testing.T) {
	testutils.Run(t, testutils.Level3, "TestAdaptiveRetry_RetryOverhead", nil, func(t *testing.T, _ *gorm.DB) {
		// Benchmark with fast retry (less sleep)
		ar := NewAdaptiveRetryWithDefaults()

		start := time.Now()
		for i := 0; i < 10; i++ {
			_ = ar.Execute(func() error {
				if i%3 == 0 && i > 0 {
					return ErrRateLimited
				}
				return nil
			}, PriorityNormal)
		}
		durationWithRetry := time.Since(start)

		t.Logf("With retry: %v", durationWithRetry)
		t.Logf("Retry rate: %.2f%%", ar.GetMetrics().GetRetryRate()*100)
	})
}

// TestCircuitBreaker_RecoveryTime measures circuit breaker recovery time
func TestCircuitBreaker_RecoveryTime(t *testing.T) {
	testutils.Run(t, testutils.Level3, "TestCircuitBreaker_RecoveryTime", nil, func(t *testing.T, _ *gorm.DB) {
		config := CircuitBreakerConfig{
			FailureThreshold: 3,
			Timeout:          100 * time.Millisecond,
		}
		ar := NewAdaptiveRetry(DefaultRetryConfig(), config)

		// Trip circuit breaker
		for i := 0; i < 10; i++ {
			_ = ar.Execute(func() error {
				return ErrRateLimited
			}, PriorityNormal)
		}

		// Measure recovery time
		recoveryStart := time.Now()
		for {
			err := ar.Execute(func() error {
				return nil
			}, PriorityNormal)
			if err == nil {
				break
			}
			if time.Since(recoveryStart) > 200*time.Millisecond {
				t.Errorf("Circuit breaker did not recover in expected time")
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		recoveryTime := time.Since(recoveryStart)

		t.Logf("Circuit breaker recovery time: %v", recoveryTime)

		// Recovery should happen within timeout + small margin
		if recoveryTime > 150*time.Millisecond {
			t.Errorf("Recovery time too long: %v", recoveryTime)
		}
	})
}

// TestAdaptiveRetry_PriorityImpact measures priority-based retry performance
func TestAdaptiveRetry_PriorityImpact(t *testing.T) {
	testutils.Run(t, testutils.Level3, "TestAdaptiveRetry_PriorityImpact", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()

		// Execute requests with different priorities
		priorities := []RequestPriority{PriorityLow, PriorityNormal, PriorityHigh, PriorityCritical}
		times := make(map[RequestPriority]time.Duration)

		for _, priority := range priorities {
			start := time.Now()
			for i := 0; i < 10; i++ {
				_ = ar.Execute(func() error {
					return nil
				}, priority)
			}
			times[priority] = time.Since(start)
		}

		for _, priority := range priorities {
			t.Logf("Priority %d: %v total, %v avg", priority, times[priority], times[priority]/10)
		}
	})
}

// TestRetryMetrics_Accuracy tests metrics accuracy
func TestRetryMetrics_Accuracy(t *testing.T) {
	testutils.Run(t, testutils.Level3, "TestRetryMetrics_Accuracy", nil, func(t *testing.T, _ *gorm.DB) {
		ar := NewAdaptiveRetryWithDefaults()

		totalAttempts := 100
		expectedSuccesses := 80

		// Execute requests
		for i := 0; i < totalAttempts; i++ {
			if i >= expectedSuccesses {
				// Use retryable error
				_ = ar.Execute(func() error {
					return ErrRateLimited
				}, PriorityNormal)
			} else {
				_ = ar.Execute(func() error {
					return nil
				}, PriorityNormal)
			}
		}

		stats := ar.GetMetrics()
		metrics := stats.GetStats()

		totalAttemptsMetric := metrics["total_attempts"].(int)
		totalSuccessesMetric := metrics["total_successes"].(int)

		t.Logf("Total attempts: %d (expected: %d)", totalAttemptsMetric, totalAttempts)
		t.Logf("Total successes: %d (expected: %d)", totalSuccessesMetric, expectedSuccesses)
		t.Logf("Total failures: %d", metrics["total_failures"].(int))
		t.Logf("Total retries: %d", metrics["total_retries"].(int))

		if totalAttemptsMetric != totalAttempts {
			t.Errorf("Expected %d attempts, got %d", totalAttempts, totalAttemptsMetric)
		}
		if totalSuccessesMetric != expectedSuccesses {
			t.Errorf("Expected %d successes, got %d", expectedSuccesses, totalSuccessesMetric)
		}
	})
}
