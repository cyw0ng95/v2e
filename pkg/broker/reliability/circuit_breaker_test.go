package routing

import (
	"fmt"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// Test 1: CircuitBreaker - Create Circuit Breaker
func TestCircuitBreaker_New(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewCircuitBreaker", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 2,
			Timeout:          30 * time.Second,
		})

		if cb == nil {
			t.Fatal("Circuit breaker is nil")
		}

		if cb.state != CircuitClosed {
			t.Errorf("Initial state = %s, want CLOSED", cb.state)
		}

		if cb.failureThreshold != 5 {
			t.Errorf("Failure threshold = %d, want 5", cb.failureThreshold)
		}
	})
}

// Test 2: CircuitBreaker - Default Configuration
func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	testutils.Run(t, testutils.Level1, "DefaultConfig", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{})

		if cb.failureThreshold != 5 {
			t.Errorf("Default failure threshold = %d, want 5", cb.failureThreshold)
		}

		if cb.successThreshold != 2 {
			t.Errorf("Default success threshold = %d, want 2", cb.successThreshold)
		}

		if cb.timeout != 30*time.Second {
			t.Errorf("Default timeout = %v, want 30s", cb.timeout)
		}
	})
}

// Test 3: CircuitBreaker - Allow Request in Closed State
func TestCircuitBreaker_AllowRequest_Closed(t *testing.T) {
	testutils.Run(t, testutils.Level1, "AllowRequestClosed", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{})

		allowed := cb.AllowRequest()
		if !allowed {
			t.Error("Circuit breaker should allow requests in CLOSED state")
		}
	})
}

// Test 4: CircuitBreaker - Record Success
func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecordSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{})

		cb.RecordSuccess()

		if cb.successCount != 1 {
			t.Errorf("Success count = %d, want 1", cb.successCount)
		}

		if cb.failureCount != 0 {
			t.Errorf("Failure count = %d, want 0 (reset on success)", cb.failureCount)
		}
	})
}

// Test 5: CircuitBreaker - Record Failure
func TestCircuitBreaker_RecordFailure(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecordFailure", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{})

		cb.RecordFailure()

		if cb.failureCount != 1 {
			t.Errorf("Failure count = %d, want 1", cb.failureCount)
		}

		if cb.successCount != 0 {
			t.Errorf("Success count = %d, want 0 (reset on failure)", cb.successCount)
		}
	})
}

// Test 6: CircuitBreaker - Transition to Open on Failure Threshold
func TestCircuitBreaker_TransitionToOpen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TransitionToOpen", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 3,
		})

		// Record 3 failures to trigger open state
		cb.RecordFailure()
		cb.RecordFailure()
		cb.RecordFailure()

		if cb.state != CircuitOpen {
			t.Errorf("State = %s, want OPEN", cb.state)
		}
	})
}

// Test 7: CircuitBreaker - Block Requests in Open State
func TestCircuitBreaker_AllowRequest_Open(t *testing.T) {
	testutils.Run(t, testutils.Level1, "AllowRequestOpen", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 2,
			Timeout:          1 * time.Second,
		})

		// Trigger open state
		cb.RecordFailure()
		cb.RecordFailure()

		// Should block requests
		allowed := cb.AllowRequest()
		if allowed {
			t.Error("Circuit breaker should block requests in OPEN state")
		}
	})
}

// Test 8: CircuitBreaker - Transition to Half-Open After Timeout
func TestCircuitBreaker_TransitionToHalfOpen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TransitionToHalfOpen", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 2,
			Timeout:          100 * time.Millisecond,
		})

		// Trigger open state
		cb.RecordFailure()
		cb.RecordFailure()

		if cb.state != CircuitOpen {
			t.Fatalf("State should be OPEN after failures")
		}

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Should transition to HALF_OPEN
		allowed := cb.AllowRequest()
		if !allowed {
			t.Error("Circuit breaker should allow requests after timeout (HALF_OPEN)")
		}

		if cb.state != CircuitHalfOpen {
			t.Errorf("State = %s, want HALF_OPEN", cb.state)
		}
	})
}

// Test 9: CircuitBreaker - Close from Half-Open on Success
func TestCircuitBreaker_HalfOpenToClosedOnSuccess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "HalfOpenToClosed", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 2,
			SuccessThreshold: 2,
			Timeout:          100 * time.Millisecond,
		})

		// Open the circuit
		cb.RecordFailure()
		cb.RecordFailure()

		// Wait for timeout -> HALF_OPEN
		time.Sleep(150 * time.Millisecond)
		cb.AllowRequest()

		// Record successes to close circuit
		cb.RecordSuccess()
		cb.RecordSuccess()

		if cb.state != CircuitClosed {
			t.Errorf("State = %s, want CLOSED", cb.state)
		}
	})
}

// Test 10: CircuitBreaker - Reopen from Half-Open on Failure
func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	testutils.Run(t, testutils.Level1, "HalfOpenToOpen", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 2,
			Timeout:          100 * time.Millisecond,
		})

		// Open the circuit
		cb.RecordFailure()
		cb.RecordFailure()

		// Wait for timeout -> HALF_OPEN
		time.Sleep(150 * time.Millisecond)
		cb.AllowRequest()

		// Record failure -> should immediately reopen
		cb.RecordFailure()

		if cb.state != CircuitOpen {
			t.Errorf("State = %s, want OPEN (any failure in HALF_OPEN reopens)", cb.state)
		}
	})
}

// Test 11: CircuitBreaker - Call with Success
func TestCircuitBreaker_Call_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CallSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{})

		executed := false
		err := cb.Call(func() error {
			executed = true
			return nil
		})

		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}

		if !executed {
			t.Error("Function was not executed")
		}

		if cb.successCount != 1 {
			t.Errorf("Success count = %d, want 1", cb.successCount)
		}
	})
}

// Test 12: CircuitBreaker - Call with Failure
func TestCircuitBreaker_Call_Failure(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CallFailure", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{})

		err := cb.Call(func() error {
			return fmt.Errorf("test error")
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if cb.failureCount != 1 {
			t.Errorf("Failure count = %d, want 1", cb.failureCount)
		}
	})
}

// Test 13: CircuitBreaker - Call Blocked in Open State
func TestCircuitBreaker_Call_Blocked(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CallBlocked", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 1,
		})

		// Open the circuit
		cb.RecordFailure()

		executed := false
		err := cb.Call(func() error {
			executed = true
			return nil
		})

		if err == nil {
			t.Error("Expected error for blocked call, got nil")
		}

		if executed {
			t.Error("Function should not execute when circuit is open")
		}
	})
}

// Test 14: CircuitBreaker - Reset
func TestCircuitBreaker_Reset(t *testing.T) {
	testutils.Run(t, testutils.Level1, "Reset", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 2,
		})

		// Open the circuit
		cb.RecordFailure()
		cb.RecordFailure()

		if cb.state != CircuitOpen {
			t.Fatalf("State should be OPEN")
		}

		// Reset
		cb.Reset()

		if cb.state != CircuitClosed {
			t.Errorf("State after reset = %s, want CLOSED", cb.state)
		}

		if cb.failureCount != 0 {
			t.Errorf("Failure count after reset = %d, want 0", cb.failureCount)
		}

		if cb.successCount != 0 {
			t.Errorf("Success count after reset = %d, want 0", cb.successCount)
		}
	})
}

// Test 15: CircuitBreaker - Get Stats
func TestCircuitBreaker_GetStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetStats", nil, func(t *testing.T, tx *gorm.DB) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          60 * time.Second,
		})

		cb.RecordFailure()
		cb.RecordSuccess()

		stats := cb.GetStats()

		if stats["state"] != "CLOSED" {
			t.Errorf("State = %s, want CLOSED", stats["state"])
		}

		if stats["failure_count"] != 0 {
			t.Errorf("Failure count = %d, want 0 (reset on success)", stats["failure_count"])
		}

		if stats["success_count"] != 1 {
			t.Errorf("Success count = %d, want 1", stats["success_count"])
		}

		if stats["failure_threshold"] != 5 {
			t.Errorf("Failure threshold = %d, want 5", stats["failure_threshold"])
		}
	})
}

// Test 16: CircuitBreakerManager - Create Manager
func TestCircuitBreakerManager_New(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewManager", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 3,
		})

		if manager == nil {
			t.Fatal("Manager is nil")
		}

		if len(manager.breakers) != 0 {
			t.Errorf("Breakers count = %d, want 0", len(manager.breakers))
		}
	})
}

// Test 17: CircuitBreakerManager - Get or Create Breaker
func TestCircuitBreakerManager_GetOrCreateBreaker(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetOrCreateBreaker", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{})

		// Create first breaker
		breaker1 := manager.GetOrCreateBreaker("service-1")
		if breaker1 == nil {
			t.Fatal("Breaker is nil")
		}

		// Get same breaker
		breaker2 := manager.GetOrCreateBreaker("service-1")
		if breaker1 != breaker2 {
			t.Error("Should return same breaker instance")
		}

		// Create different breaker
		breaker3 := manager.GetOrCreateBreaker("service-2")
		if breaker1 == breaker3 {
			t.Error("Should create different breaker for different service")
		}
	})
}

// Test 18: CircuitBreakerManager - Call Success
func TestCircuitBreakerManager_Call_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ManagerCallSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{})

		executed := false
		err := manager.Call("service-1", func() error {
			executed = true
			return nil
		})

		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}

		if !executed {
			t.Error("Function was not executed")
		}

		state := manager.GetState("service-1")
		if state != CircuitClosed {
			t.Errorf("State = %s, want CLOSED", state)
		}
	})
}

// Test 19: CircuitBreakerManager - Record Success
func TestCircuitBreakerManager_RecordSuccess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ManagerRecordSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{})

		manager.RecordSuccess("service-1")

		breaker := manager.GetOrCreateBreaker("service-1")
		if breaker.successCount != 1 {
			t.Errorf("Success count = %d, want 1", breaker.successCount)
		}
	})
}

// Test 20: CircuitBreakerManager - Record Failure
func TestCircuitBreakerManager_RecordFailure(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ManagerRecordFailure", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{})

		manager.RecordFailure("service-1")

		breaker := manager.GetOrCreateBreaker("service-1")
		if breaker.failureCount != 1 {
			t.Errorf("Failure count = %d, want 1", breaker.failureCount)
		}
	})
}

// Test 21: CircuitBreakerManager - State Change Callback
func TestCircuitBreakerManager_StateChangeCallback(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StateChangeCallback", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 2,
		})

		var callbackTarget string
		var callbackOldState, callbackNewState CircuitState

		manager.SetStateChangeCallback(func(target string, oldState, newState CircuitState) {
			callbackTarget = target
			callbackOldState = oldState
			callbackNewState = newState
		})

		// Trigger state change
		manager.RecordFailure("service-1")
		manager.RecordFailure("service-1") // Should open circuit

		if callbackTarget != "service-1" {
			t.Errorf("Callback target = %s, want service-1", callbackTarget)
		}

		if callbackOldState != CircuitClosed {
			t.Errorf("Callback old state = %s, want CLOSED", callbackOldState)
		}

		if callbackNewState != CircuitOpen {
			t.Errorf("Callback new state = %s, want OPEN", callbackNewState)
		}
	})
}

// Test 22: CircuitBreakerManager - Get All States
func TestCircuitBreakerManager_GetAllStates(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetAllStates", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 2,
		})

		// Create multiple breakers with different states
		manager.RecordSuccess("service-1") // CLOSED
		manager.RecordFailure("service-2")
		manager.RecordFailure("service-2") // OPEN
		manager.RecordSuccess("service-3") // CLOSED

		states := manager.GetAllStates()

		if len(states) != 3 {
			t.Errorf("States count = %d, want 3", len(states))
		}

		if states["service-1"] != CircuitClosed {
			t.Errorf("Service-1 state = %s, want CLOSED", states["service-1"])
		}

		if states["service-2"] != CircuitOpen {
			t.Errorf("Service-2 state = %s, want OPEN", states["service-2"])
		}
	})
}

// Test 23: CircuitBreakerManager - Reset Breaker
func TestCircuitBreakerManager_Reset(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ManagerReset", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 2,
		})

		// Open circuit
		manager.RecordFailure("service-1")
		manager.RecordFailure("service-1")

		if manager.GetState("service-1") != CircuitOpen {
			t.Fatalf("Circuit should be open")
		}

		// Reset
		manager.Reset("service-1")

		if manager.GetState("service-1") != CircuitClosed {
			t.Errorf("State after reset = %s, want CLOSED", manager.GetState("service-1"))
		}
	})
}

// Test 24: CircuitBreakerManager - Allow Request
func TestCircuitBreakerManager_AllowRequest(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ManagerAllowRequest", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{
			FailureThreshold: 2,
		})

		// Initially should allow
		if !manager.AllowRequest("service-1") {
			t.Error("Should allow requests initially")
		}

		// Open circuit
		manager.RecordFailure("service-1")
		manager.RecordFailure("service-1")

		// Should not allow
		if manager.AllowRequest("service-1") {
			t.Error("Should not allow requests when circuit is open")
		}
	})
}

// Test 25: CircuitBreakerManager - Get Stats
func TestCircuitBreakerManager_GetStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ManagerGetStats", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewCircuitBreakerManager(CircuitBreakerConfig{})

		manager.RecordSuccess("service-1")
		manager.RecordFailure("service-2")

		stats := manager.GetStats()

		if len(stats) != 2 {
			t.Errorf("Stats count = %d, want 2", len(stats))
		}

		if stats["service-1"] == nil {
			t.Error("Service-1 stats should exist")
		}

		if stats["service-2"] == nil {
			t.Error("Service-2 stats should exist")
		}
	})
}
