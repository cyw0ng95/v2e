package reliability

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed - Normal operation, requests pass through
	CircuitClosed CircuitState = iota
	// CircuitOpen - Circuit is open, requests are blocked
	CircuitOpen
	// CircuitHalfOpen - Testing if service has recovered
	CircuitHalfOpen
)

// String returns string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements circuit breaker pattern for subprocess health
// Implements Requirement 10: Broker Circuit Breakers
type CircuitBreaker struct {
	mu                  sync.RWMutex
	state               CircuitState
	failureCount        int
	successCount        int
	failureThreshold    int           // Number of consecutive failures to open circuit
	successThreshold    int           // Number of consecutive successes to close circuit
	timeout             time.Duration // Time to wait before transitioning from Open to HalfOpen
	lastFailureTime     time.Time
	lastStateChangeTime time.Time
	openUntil           time.Time
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           // Default: 5
	SuccessThreshold int           // Default: 2
	Timeout          time.Duration // Default: 30s
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 2
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	return &CircuitBreaker{
		state:               CircuitClosed,
		failureThreshold:    config.FailureThreshold,
		successThreshold:    config.SuccessThreshold,
		timeout:             config.Timeout,
		lastStateChangeTime: time.Now(),
	}
}

// Call attempts to execute a request through the circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.AllowRequest() {
		return fmt.Errorf("circuit breaker is OPEN")
	}

	err := fn()

	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// AllowRequest checks if a request should be allowed through
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if timeout has expired
		if time.Now().After(cb.openUntil) {
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			cb.failureCount = 0
			cb.lastStateChangeTime = time.Now()
			return true
		}
		return false
	case CircuitHalfOpen:
		// Allow limited requests in half-open state
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0 // Reset failure count
	cb.successCount++

	switch cb.state {
	case CircuitHalfOpen:
		// If enough successes, close the circuit
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.successCount = 0
			cb.lastStateChangeTime = time.Now()
		}
	case CircuitClosed:
		// Already closed, nothing to do
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount = 0 // Reset success count
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitClosed:
		// If enough failures, open the circuit
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitOpen
			cb.openUntil = time.Now().Add(cb.timeout)
			cb.lastStateChangeTime = time.Now()
		}
	case CircuitHalfOpen:
		// Any failure in half-open state immediately opens circuit
		cb.state = CircuitOpen
		cb.openUntil = time.Now().Add(cb.timeout)
		cb.lastStateChangeTime = time.Now()
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChangeTime = time.Now()
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":                  cb.state.String(),
		"failure_count":          cb.failureCount,
		"success_count":          cb.successCount,
		"failure_threshold":      cb.failureThreshold,
		"success_threshold":      cb.successThreshold,
		"timeout_seconds":        cb.timeout.Seconds(),
		"last_failure_time":      cb.lastFailureTime,
		"last_state_change_time": cb.lastStateChangeTime,
		"open_until":             cb.openUntil,
	}
}

// CircuitBreakerManager manages circuit breakers for multiple targets
type CircuitBreakerManager struct {
	mu            sync.RWMutex
	breakers      map[string]*CircuitBreaker
	defaultConfig CircuitBreakerConfig
	onStateChange func(target string, oldState, newState CircuitState)
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers:      make(map[string]*CircuitBreaker),
		defaultConfig: config,
	}
}

// GetOrCreateBreaker gets or creates a circuit breaker for a target
func (m *CircuitBreakerManager) GetOrCreateBreaker(target string) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	breaker, exists := m.breakers[target]
	if !exists {
		breaker = NewCircuitBreaker(m.defaultConfig)
		m.breakers[target] = breaker
	}

	return breaker
}

// Call executes a request through the circuit breaker for a target
func (m *CircuitBreakerManager) Call(target string, fn func() error) error {
	breaker := m.GetOrCreateBreaker(target)

	oldState := breaker.GetState()
	err := breaker.Call(fn)
	newState := breaker.GetState()

	// Notify state change if configured
	if oldState != newState && m.onStateChange != nil {
		m.onStateChange(target, oldState, newState)
	}

	return err
}

// AllowRequest checks if a request to target should be allowed
func (m *CircuitBreakerManager) AllowRequest(target string) bool {
	breaker := m.GetOrCreateBreaker(target)
	return breaker.AllowRequest()
}

// RecordSuccess records a success for a target
func (m *CircuitBreakerManager) RecordSuccess(target string) {
	breaker := m.GetOrCreateBreaker(target)
	oldState := breaker.GetState()
	breaker.RecordSuccess()
	newState := breaker.GetState()

	if oldState != newState && m.onStateChange != nil {
		m.onStateChange(target, oldState, newState)
	}
}

// RecordFailure records a failure for a target
func (m *CircuitBreakerManager) RecordFailure(target string) {
	breaker := m.GetOrCreateBreaker(target)
	oldState := breaker.GetState()
	breaker.RecordFailure()
	newState := breaker.GetState()

	if oldState != newState && m.onStateChange != nil {
		m.onStateChange(target, oldState, newState)
	}
}

// Reset resets the circuit breaker for a target
func (m *CircuitBreakerManager) Reset(target string) {
	breaker := m.GetOrCreateBreaker(target)
	breaker.Reset()
}

// GetState returns the circuit state for a target
func (m *CircuitBreakerManager) GetState(target string) CircuitState {
	m.mu.RLock()
	breaker, exists := m.breakers[target]
	m.mu.RUnlock()

	if !exists {
		return CircuitClosed
	}

	return breaker.GetState()
}

// GetAllStates returns circuit states for all targets
func (m *CircuitBreakerManager) GetAllStates() map[string]CircuitState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]CircuitState)
	for target, breaker := range m.breakers {
		states[target] = breaker.GetState()
	}

	return states
}

// GetStats returns statistics for all circuit breakers
func (m *CircuitBreakerManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	for target, breaker := range m.breakers {
		stats[target] = breaker.GetStats()
	}

	return stats
}

// SetStateChangeCallback sets a callback for state changes
func (m *CircuitBreakerManager) SetStateChangeCallback(callback func(target string, oldState, newState CircuitState)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onStateChange = callback
}
