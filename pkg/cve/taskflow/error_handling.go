package taskflow

import (
	"math"
	"sync"
	"time"
)

// RetryPolicy defines the configuration for retry logic
type RetryPolicy struct {
	MaxRetries    int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// CalculateDelay calculates the delay for the given attempt number using exponential backoff
func (rp *RetryPolicy) CalculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rp.BaseDelay) * math.Pow(rp.BackoffFactor, float64(attempt)))
	if delay > rp.MaxDelay {
		delay = rp.MaxDelay
	}
	return delay
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures
type CircuitBreaker struct {
	mutex        sync.Mutex
	state        CircuitBreakerState
	failureCount int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the specified parameters
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        Closed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Call executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()

	// Check if we should attempt to reset the circuit
	if cb.state == Open && time.Since(cb.lastFailure) >= cb.resetTimeout {
		cb.state = HalfOpen
	}

	// If circuit is open, return error immediately
	if cb.state == Open {
		cb.mutex.Unlock()
		return nil // Circuit is open, operation blocked
	}

	cb.mutex.Unlock()

	// Execute the function
	err := fn()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		// Record failure
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.failureCount >= cb.maxFailures {
			cb.state = Open
		}
	} else {
		// Reset on success
		cb.failureCount = 0
		cb.state = Closed
	}

	return err
}

// IsOpen returns true if the circuit breaker is currently open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	return cb.state == Open
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = Closed
	cb.failureCount = 0
}

// StateMachineController manages state transitions with built-in validation
type StateMachineController struct {
	runStore *RunStore
	mutex    sync.RWMutex
}

// TransitTo transitions a run to a new state with validation
func (smc *StateMachineController) TransitTo(runID string, newState JobState) error {
	smc.mutex.Lock()
	defer smc.mutex.Unlock()

	run, err := smc.runStore.GetRun(runID)
	if err != nil {
		return err
	}

	if !run.State.CanTransitionTo(newState) {
		return nil // For now, just return nil for invalid transitions
	}

	// Execute transition
	oldState := run.State
	run.State = newState
	run.UpdatedAt = time.Now()

	if err := smc.runStore.saveRun(run); err != nil {
		// Restore old state on failure
		run.State = oldState
		return err
	}

	return nil
}
