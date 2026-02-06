package remote

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	// ErrCircuitBreakerOpen is returned when circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrMaxRetriesExceeded is returned when max retries exceeded
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
)

// RetryStrategy defines the retry strategy type
type RetryStrategy string

const (
	// StrategyExponential uses exponential backoff with jitter
	StrategyExponential RetryStrategy = "exponential"
	// StrategyLinear uses linear backoff
	StrategyLinear RetryStrategy = "linear"
	// StrategyFixed uses fixed delay
	StrategyFixed RetryStrategy = "fixed"
)

// RequestPriority defines priority for requests
type RequestPriority int

const (
	// PriorityLow for non-critical requests
	PriorityLow RequestPriority = 0
	// PriorityNormal for standard requests
	PriorityNormal RequestPriority = 1
	// PriorityHigh for important requests
	PriorityHigh RequestPriority = 2
	// PriorityCritical for critical requests
	PriorityCritical RequestPriority = 3
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	// Strategy is the backoff strategy
	Strategy RetryStrategy
	// MaxRetries is maximum number of retry attempts
	MaxRetries int
	// InitialDelay is initial delay before first retry
	InitialDelay time.Duration
	// MaxDelay is maximum delay between retries
	MaxDelay time.Duration
	// JitterEnabled adds randomness to backoff
	JitterEnabled bool
	// JitterFactor is jitter factor (0.0-1.0)
	JitterFactor float64
	// RetryableStatusCodes are HTTP status codes that trigger retry
	RetryableStatusCodes []int
	// RetryableErrors are errors that trigger retry
	RetryableErrors []error
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		Strategy:             StrategyExponential,
		MaxRetries:           3,
		InitialDelay:         1 * time.Second,
		MaxDelay:             30 * time.Second,
		JitterEnabled:        true,
		JitterFactor:         0.2,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
		RetryableErrors:      []error{ErrRateLimited},
	}
}

// CircuitBreakerState represents circuit breaker state
type CircuitBreakerState string

const (
	// StateClosed is normal operation
	StateClosed CircuitBreakerState = "closed"
	// StateOpen is when circuit breaker is tripped
	StateOpen CircuitBreakerState = "open"
	// StateHalfOpen is testing if service recovered
	StateHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	// FailureThreshold is failures before opening
	FailureThreshold int
	// SuccessThreshold is successes to close after half-open
	SuccessThreshold int
	// Timeout is duration to stay open
	Timeout time.Duration
	// HalfOpenRequests is requests allowed in half-open state
	HalfOpenRequests int
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          1 * time.Minute,
		HalfOpenRequests: 3,
	}
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	config            CircuitBreakerConfig
	mu                sync.RWMutex
	state             CircuitBreakerState
	failures          int
	lastFailureTime   time.Time
	halfOpenSuccesses int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// CanExecute checks if request can be executed
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.halfOpenSuccesses = 0
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited requests
		return cb.halfOpenSuccesses < cb.config.HalfOpenRequests
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	if cb.failures >= cb.config.FailureThreshold {
		cb.state = StateOpen
	}
}

// GetState returns current state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// RetryMetrics tracks retry statistics
type RetryMetrics struct {
	mu               sync.RWMutex
	totalAttempts    int
	totalSuccesses   int
	totalFailures    int
	totalRetries     int
	backoffTimeTotal time.Duration
	priorityStats    map[RequestPriority]int
}

// NewRetryMetrics creates new retry metrics
func NewRetryMetrics() *RetryMetrics {
	return &RetryMetrics{
		priorityStats: make(map[RequestPriority]int),
	}
}

// RecordAttempt records an attempt
func (rm *RetryMetrics) RecordAttempt(priority RequestPriority) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.totalAttempts++
	rm.priorityStats[priority]++
}

// RecordSuccess records a successful request
func (rm *RetryMetrics) RecordSuccess() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.totalSuccesses++
}

// RecordFailure records a failed request
func (rm *RetryMetrics) RecordFailure() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.totalFailures++
}

// RecordRetry records a retry attempt
func (rm *RetryMetrics) RecordRetry(backoffTime time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.totalRetries++
	rm.backoffTimeTotal += backoffTime
}

// GetSuccessRate returns success rate
func (rm *RetryMetrics) GetSuccessRate() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if rm.totalAttempts == 0 {
		return 0
	}
	return float64(rm.totalSuccesses) / float64(rm.totalAttempts)
}

// GetRetryRate returns retry rate
func (rm *RetryMetrics) GetRetryRate() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if rm.totalAttempts == 0 {
		return 0
	}
	return float64(rm.totalRetries) / float64(rm.totalAttempts)
}

// GetAverageBackoffTime returns average backoff time
func (rm *RetryMetrics) GetAverageBackoffTime() time.Duration {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if rm.totalRetries == 0 {
		return 0
	}
	return rm.backoffTimeTotal / time.Duration(rm.totalRetries)
}

// GetStats returns all statistics
func (rm *RetryMetrics) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return map[string]interface{}{
		"total_attempts":     rm.totalAttempts,
		"total_successes":    rm.totalSuccesses,
		"total_failures":     rm.totalFailures,
		"total_retries":      rm.totalRetries,
		"success_rate":       rm.GetSuccessRate(),
		"retry_rate":         rm.GetRetryRate(),
		"avg_backoff_time":   rm.GetAverageBackoffTime().String(),
		"priority_breakdown": rm.priorityStats,
	}
}

// AdaptiveRetry manages retry logic with exponential backoff and circuit breaker
type AdaptiveRetry struct {
	config         RetryConfig
	circuitBreaker *CircuitBreaker
	metrics        *RetryMetrics
	rand           *rand.Rand
}

// NewAdaptiveRetry creates a new adaptive retry manager
func NewAdaptiveRetry(retryConfig RetryConfig, cbConfig CircuitBreakerConfig) *AdaptiveRetry {
	return &AdaptiveRetry{
		config:         retryConfig,
		circuitBreaker: NewCircuitBreaker(cbConfig),
		metrics:        NewRetryMetrics(),
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewAdaptiveRetryWithDefaults creates retry manager with default configurations
func NewAdaptiveRetryWithDefaults() *AdaptiveRetry {
	return NewAdaptiveRetry(DefaultRetryConfig(), DefaultCircuitBreakerConfig())
}

// Execute runs a function with retry logic
func (ar *AdaptiveRetry) Execute(fn func() error, priority RequestPriority) error {
	ar.metrics.RecordAttempt(priority)

	var lastErr error

	// Check circuit breaker
	if !ar.circuitBreaker.CanExecute() {
		return fmt.Errorf("%w: circuit breaker state: %s", ErrCircuitBreakerOpen, ar.circuitBreaker.GetState())
	}

	for attempt := 0; attempt <= ar.config.MaxRetries; attempt++ {
		// Execute function
		err := fn()
		if err == nil {
			ar.circuitBreaker.RecordSuccess()
			ar.metrics.RecordSuccess()
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !ar.isRetryable(err) {
			ar.circuitBreaker.RecordFailure()
			ar.metrics.RecordFailure()
			return err
		}

		// If max retries exceeded
		if attempt >= ar.config.MaxRetries {
			ar.circuitBreaker.RecordFailure()
			ar.metrics.RecordFailure()
			return fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
		}

		// Calculate backoff
		backoff := ar.calculateBackoff(attempt)
		ar.metrics.RecordRetry(backoff)

		time.Sleep(backoff)
	}

	return lastErr
}

// isRetryable checks if error should trigger retry
func (ar *AdaptiveRetry) isRetryable(err error) bool {
	// Check for specific errors
	for _, retryableErr := range ar.config.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// calculateBackoff calculates backoff delay
func (ar *AdaptiveRetry) calculateBackoff(attempt int) time.Duration {
	var delay time.Duration

	switch ar.config.Strategy {
	case StrategyExponential:
		delay = ar.config.InitialDelay * time.Duration(1<<uint(attempt))
	case StrategyLinear:
		delay = ar.config.InitialDelay * time.Duration(attempt+1)
	case StrategyFixed:
		delay = ar.config.InitialDelay
	default:
		delay = ar.config.InitialDelay
	}

	// Cap at max delay
	if delay > ar.config.MaxDelay {
		delay = ar.config.MaxDelay
	}

	// Add jitter if enabled
	if ar.config.JitterEnabled {
		jitter := ar.rand.Float64() * ar.config.JitterFactor * float64(delay)
		delay = time.Duration(float64(delay) - jitter)
	}

	return delay
}

// GetMetrics returns retry metrics
func (ar *AdaptiveRetry) GetMetrics() *RetryMetrics {
	return ar.metrics
}

// GetCircuitBreakerState returns circuit breaker state
func (ar *AdaptiveRetry) GetCircuitBreakerState() CircuitBreakerState {
	return ar.circuitBreaker.GetState()
}
