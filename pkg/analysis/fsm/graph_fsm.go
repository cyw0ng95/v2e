package fsm

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// OperationType represents the type of operation being executed
type OperationType string

const (
	OpBuild    OperationType = "BUILD"
	OpAnalysis OperationType = "ANALYSIS"
	OpPersist  OperationType = "PERSIST"
)

// BaseGraphFSM provides base implementation of GraphFSM
type BaseGraphFSM struct {
	mu              sync.RWMutex
	state           GraphState
	eventHandler     func(*Event) error
	logger          *common.Logger
	lastError       error
	retryConfig     RetryConfig
	lastOperation   OperationType
	retryCount      int
	lastFailedState GraphState

	// Enhanced error handling fields
	transitionHistory *TransitionHistory
	rollbackManager  *RollbackManager
	lastSnapshot    StateSnapshot
}

// NewGraphFSM creates a new GraphFSM instance
func NewGraphFSM(logger *common.Logger) GraphFSM {
	return &BaseGraphFSM{
		state:            GraphIdle,
		logger:           logger,
		retryConfig:      DefaultRetryConfig(),
		transitionHistory: NewTransitionHistory(100),
		rollbackManager:  NewRollbackManager(5),
	}
}

// SetRetryConfig sets retry configuration for transient errors
func (g *BaseGraphFSM) SetRetryConfig(config RetryConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.retryConfig = config
}

// GetRetryConfig returns current retry configuration
func (g *BaseGraphFSM) GetRetryConfig() RetryConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.retryConfig
}

// GetState returns current graph state
func (g *BaseGraphFSM) GetState() GraphState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state
}

// Transition attempts to transition to a new state with rollback support
func (g *BaseGraphFSM) Transition(newState GraphState) error {
	startTime := time.Now()
	g.mu.Lock()

	currentState := g.state

	// Save snapshot for potential rollback
	g.lastSnapshot = g.rollbackManager.SaveSnapshot(string(currentState), currentState)

	// Validate the transition
	if err := ValidateGraphTransition(currentState, newState); err != nil {
		g.mu.Unlock()

		// Create a detailed transition error
		transErr := NewTransitionError(ErrInvalidTransition, string(currentState), string(newState), err)
		transErr.CanRecover = false // Invalid transitions are not recoverable

		// Record the failed transition
		g.transitionHistory.Record(string(currentState), string(newState), false, transErr, time.Since(startTime))

		// Log the error
		if g.logger != nil {
			g.logger.Error("GraphFSM transition validation failed: %s -> %s: %v", currentState, newState, err)
		}

		return transErr
	}

	// Check if we're transitioning from ERROR state - reset retry count
	if currentState == GraphError {
		g.retryCount = 0
		g.lastOperation = ""
		g.lastFailedState = ""
	}

	oldState := g.state
	g.state = newState

	if g.logger != nil {
		g.logger.Info("GraphFSM state transition: %s -> %s", oldState, newState)
	}

	g.mu.Unlock()

	// Record successful transition
	g.transitionHistory.Record(string(oldState), string(newState), true, nil, time.Since(startTime))

	return nil
}

// TransitionWithHandler attempts a transition with a handler function that can be rolled back
func (g *BaseGraphFSM) TransitionWithHandler(newState GraphState, handler func() error) error {
	startTime := time.Now()
	g.mu.Lock()

	currentState := g.state

	// Save snapshot for potential rollback
	g.lastSnapshot = g.rollbackManager.SaveSnapshot(string(currentState), currentState)

	// Validate the transition first
	if err := ValidateGraphTransition(currentState, newState); err != nil {
		g.mu.Unlock()

		transErr := NewTransitionError(ErrInvalidTransition, string(currentState), string(newState), err)
		transErr.CanRecover = false

		g.transitionHistory.Record(string(currentState), string(newState), false, transErr, time.Since(startTime))

		if g.logger != nil {
			g.logger.Error("GraphFSM transition validation failed: %s -> %s: %v", currentState, newState, err)
		}

		return transErr
	}

	oldState := g.state

	// Attempt state change
	g.state = newState

	// Unlock before calling handler to avoid deadlock
	g.mu.Unlock()

	// Execute the handler
	var handlerErr error
	if handler != nil {
		handlerErr = handler()
	}

	// If handler failed, attempt rollback
	if handlerErr != nil {
		g.mu.Lock()
		// Rollback to previous state
		g.state = oldState
		g.mu.Unlock()

		// Create transition error
		transErr := NewTransitionError(ErrTransitionFailed, string(oldState), string(newState), handlerErr)
		transErr.RolledBack = true
		transErr.CanRecover = true

		g.transitionHistory.Record(string(oldState), string(newState), false, transErr, time.Since(startTime))

		if g.logger != nil {
			g.logger.Error("GraphFSM transition handler failed, rolled back: %s -> %s: %v", oldState, newState, handlerErr)
		}

		return transErr
	}

	// Check if we're transitioning from ERROR state - reset retry count
	if oldState == GraphError {
		g.mu.Lock()
		g.retryCount = 0
		g.lastOperation = ""
		g.lastFailedState = ""
		g.mu.Unlock()
	}

	if g.logger != nil {
		g.logger.Info("GraphFSM state transition: %s -> %s", oldState, newState)
	}

	// Record successful transition
	g.transitionHistory.Record(string(oldState), string(newState), true, nil, time.Since(startTime))

	return nil
}

// RollbackToState rolls back to a previous state
func (g *BaseGraphFSM) RollbackToState(targetState GraphState) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	currentState := g.state

	// Check if rollback is valid
	if err := ValidateGraphTransition(currentState, targetState); err != nil {
		// For rollback, we may need to force the transition
		// Log a warning but proceed
		if g.logger != nil {
			g.logger.Warn("Forcing rollback transition: %s -> %s (validation would fail: %v)", currentState, targetState, err)
		}
	}

	// Check if we have a snapshot for the target state
	if snapshot, exists := g.rollbackManager.GetLatestSnapshot(string(targetState)); exists {
		g.state = targetState
		g.lastSnapshot = snapshot

		if g.logger != nil {
			g.logger.Info("GraphFSM rolled back to state: %s (snapshot from %s)", targetState, snapshot.Timestamp.Format(time.RFC3339))
		}

		return nil
	}

	// No snapshot found, just change state
	g.state = targetState

	if g.logger != nil {
		g.logger.Info("GraphFSM rolled back to state: %s (no snapshot available)", targetState)
	}

	return nil
}

// canRetry checks if retry is allowed based on error type and retry count
func (g *BaseGraphFSM) canRetry() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Check if we have an FSMError that is transient
	if g.lastError != nil {
		var fsmErr *FSMError
		if IsFSMError(g.lastError) && errors.As(g.lastError, &fsmErr) {
			if fsmErr.IsTransient() && g.retryCount < g.retryConfig.MaxRetries {
				return true
			}
		}
	}
	return g.retryCount < g.retryConfig.MaxRetries
}

// calculateBackoffDelay calculates exponential backoff delay
func (g *BaseGraphFSM) calculateBackoffDelay() time.Duration {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.retryConfig.BackoffFactor <= 0 {
		return g.retryConfig.BaseDelay
	}

	// Calculate exponential delay: base * (backoffFactor ^ retryCount)
	delay := float64(g.retryConfig.BaseDelay) * math.Pow(g.retryConfig.BackoffFactor, float64(g.retryCount))

	// Cap at max delay
	if delay > float64(g.retryConfig.MaxDelay) {
		delay = float64(g.retryConfig.MaxDelay)
	}

	return time.Duration(delay)
}

// StartBuild initiates graph building
func (g *BaseGraphFSM) StartBuild() error {
	// If we're in ERROR state and retry is allowed, attempt recovery
	if g.GetState() == GraphError {
		if g.canRetry() {
			// Reset retry count for new operation
			g.mu.Lock()
			g.retryCount = 0
			g.mu.Unlock()
		} else {
			return fmt.Errorf("cannot start build from ERROR state: %w", g.lastError)
		}
	}

	if err := g.Transition(GraphBuilding); err != nil {
		return err
	}

	g.mu.Lock()
	g.lastOperation = OpBuild
	g.mu.Unlock()

	event := NewEvent(EventGraphBuildStarted)
	return g.emitEvent(event)
}

// CompleteBuild marks graph building as complete
func (g *BaseGraphFSM) CompleteBuild() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphReady); err != nil {
		return err
	}

	event := NewEvent(EventGraphBuildCompleted)
	return g.emitEvent(event)
}

// FailBuild marks graph building as failed
func (g *BaseGraphFSM) FailBuild(err error) error {
	g.mu.Lock()

	// Wrap error with context if not already an FSMError
	var fsmErr *FSMError
	if !IsFSMError(err) {
		// Classify error as transient by default for build operations
		// Common transient errors: network issues, temporary unavailability
		fsmErr = NewTransientError(string(GraphBuilding), err)
		fsmErr.RetryCount = g.retryCount
	} else {
		fsmErr = err.(*FSMError)
		fsmErr.RetryCount = g.retryCount
	}

	g.lastError = fsmErr
	g.lastFailedState = GraphBuilding
	g.retryCount++

	currentRetryCount := g.retryCount
	maxRetries := g.retryConfig.MaxRetries
	logger := g.logger
	g.mu.Unlock()

	if transErr := g.Transition(GraphError); transErr != nil {
		return fmt.Errorf("failed to transition to ERROR state: %w", transErr)
	}

	// Log error with retry context
	if logger != nil {
		if currentRetryCount <= maxRetries {
			logger.Warn("Build failed (attempt %d/%d): %v", currentRetryCount, maxRetries, fsmErr.Err)
		} else {
			logger.Error("Build failed after %d attempts (max retries exceeded): %v", currentRetryCount, fsmErr.Err)
		}
	}

	event := NewEvent(EventGraphBuildFailed)
	event.Data["error"] = fsmErr.Error()
	event.Data["retry_count"] = currentRetryCount
	event.Data["can_retry"] = currentRetryCount <= maxRetries

	return g.emitEvent(event)
}

// StartAnalysis initiates graph analysis
func (g *BaseGraphFSM) StartAnalysis() error {
	if err := g.Transition(GraphAnalyzing); err != nil {
		return err
	}

	g.mu.Lock()
	g.lastOperation = OpAnalysis
	g.mu.Unlock()

	event := NewEvent(EventGraphAnalysisStarted)
	return g.emitEvent(event)
}

// CompleteAnalysis marks analysis as complete
func (g *BaseGraphFSM) CompleteAnalysis() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphReady); err != nil {
		return err
	}

	event := NewEvent(EventGraphAnalysisCompleted)
	return g.emitEvent(event)
}

// FailAnalysis marks analysis as failed
func (g *BaseGraphFSM) FailAnalysis(err error) error {
	g.mu.Lock()

	// Wrap error with context if not already an FSMError
	var fsmErr *FSMError
	if !IsFSMError(err) {
		fsmErr = NewTransientError(string(GraphAnalyzing), err)
		fsmErr.RetryCount = g.retryCount
	} else {
		fsmErr = err.(*FSMError)
		fsmErr.RetryCount = g.retryCount
	}

	g.lastError = fsmErr
	g.lastFailedState = GraphAnalyzing
	g.retryCount++

	currentRetryCount := g.retryCount
	maxRetries := g.retryConfig.MaxRetries
	logger := g.logger
	g.mu.Unlock()

	if transErr := g.Transition(GraphError); transErr != nil {
		return fmt.Errorf("failed to transition to ERROR state: %w", transErr)
	}

	if logger != nil {
		if currentRetryCount <= maxRetries {
			logger.Warn("Analysis failed (attempt %d/%d): %v", currentRetryCount, maxRetries, fsmErr.Err)
		} else {
			logger.Error("Analysis failed after %d attempts (max retries exceeded): %v", currentRetryCount, fsmErr.Err)
		}
	}

	event := NewEvent(EventGraphAnalysisFailed)
	event.Data["error"] = fsmErr.Error()
	event.Data["retry_count"] = currentRetryCount
	event.Data["can_retry"] = currentRetryCount <= maxRetries

	return g.emitEvent(event)
}

// StartPersist initiates graph persistence
func (g *BaseGraphFSM) StartPersist() error {
	if err := g.Transition(GraphPersisting); err != nil {
		return err
	}

	g.mu.Lock()
	g.lastOperation = OpPersist
	g.mu.Unlock()

	event := NewEvent(EventGraphPersistStarted)
	return g.emitEvent(event)
}

// CompletePersist marks persistence as complete
func (g *BaseGraphFSM) CompletePersist() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphReady); err != nil {
		return err
	}

	event := NewEvent(EventGraphPersistCompleted)
	return g.emitEvent(event)
}

// FailPersist marks persistence as failed
func (g *BaseGraphFSM) FailPersist(err error) error {
	g.mu.Lock()

	// Wrap error with context if not already an FSMError
	var fsmErr *FSMError
	if !IsFSMError(err) {
		// Classify error - persist errors are often transient (disk full, lock contention)
		fsmErr = NewTransientError(string(GraphPersisting), err)
		fsmErr.RetryCount = g.retryCount
	} else {
		fsmErr = err.(*FSMError)
		fsmErr.RetryCount = g.retryCount
	}

	g.lastError = fsmErr
	g.lastFailedState = GraphPersisting
	g.retryCount++

	currentRetryCount := g.retryCount
	maxRetries := g.retryConfig.MaxRetries
	logger := g.logger
	g.mu.Unlock()

	if transErr := g.Transition(GraphError); transErr != nil {
		return fmt.Errorf("failed to transition to ERROR state: %w", transErr)
	}

	// Log error with retry context
	if logger != nil {
		if currentRetryCount <= maxRetries {
			logger.Warn("Persist failed (attempt %d/%d): %v", currentRetryCount, maxRetries, fsmErr.Err)
		} else {
			logger.Error("Persist failed after %d attempts (max retries exceeded): %v", currentRetryCount, fsmErr.Err)
		}
	}

	event := NewEvent(EventGraphPersistFailed)
	event.Data["error"] = fsmErr.Error()
	event.Data["retry_count"] = currentRetryCount
	event.Data["can_retry"] = currentRetryCount <= maxRetries

	return g.emitEvent(event)
}

// Clear clears graph
func (g *BaseGraphFSM) Clear() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphIdle); err != nil {
		return err
	}

	event := NewEvent(EventGraphCleared)
	return g.emitEvent(event)
}

// SetEventHandler sets callback for event bubbling
func (g *BaseGraphFSM) SetEventHandler(handler func(*Event) error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.eventHandler = handler
}

// emitEvent emits an event to parent FSM
func (g *BaseGraphFSM) emitEvent(event *Event) error {
	g.mu.RLock()
	handler := g.eventHandler
	g.mu.RUnlock()

	if handler != nil {
		// Handle event handler errors gracefully
		if err := handler(event); err != nil {
			if g.logger != nil {
				g.logger.Error("Event handler failed for event %s: %v", event.Type, err)
			}
			// Don't fail the state transition if event handler fails,
			// but log the error for investigation
		}
	}

	return nil
}

// GetLastError returns last error encountered
func (g *BaseGraphFSM) GetLastError() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.lastError
}

// Reset resets FSM to idle state (used for error recovery)
func (g *BaseGraphFSM) Reset() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	currentState := g.GetState()
	if currentState == GraphError {
		return g.Transition(GraphIdle)
	}

	return fmt.Errorf("cannot reset from state: %s", currentState)
}

// RetryFailedOperation attempts to retry the last failed operation
func (g *BaseGraphFSM) RetryFailedOperation() error {
	g.mu.Lock()
	lastFailed := g.lastFailedState
	lastErr := g.lastError
	retryCount := g.retryCount
	maxRetries := g.retryConfig.MaxRetries
	g.mu.Unlock()

	if lastFailed == "" {
		return fmt.Errorf("no failed operation to retry")
	}

	if retryCount > maxRetries {
		return fmt.Errorf("max retries (%d) exceeded for operation %s: %w", maxRetries, lastFailed, lastErr)
	}

	// Check if error is retryable
	var fsmErr *FSMError
	if IsFSMError(lastErr) && errors.As(lastErr, &fsmErr) {
		if fsmErr.IsPermanent() {
			return fmt.Errorf("cannot retry permanent error: %w", lastErr)
		}
	}

	// Apply exponential backoff before retry
	delay := g.calculateBackoffDelay()
	if g.logger != nil {
		g.logger.Info("Retrying operation %s after %v delay (attempt %d/%d)", lastFailed, delay, retryCount+1, maxRetries+1)
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	// Retry the operation based on what failed
	switch lastFailed {
	case GraphBuilding:
		return g.StartBuild()
	case GraphAnalyzing:
		return g.StartAnalysis()
	case GraphPersisting:
		return g.StartPersist()
	default:
		return fmt.Errorf("unknown operation type for retry: %s", lastFailed)
	}
}

// CanRecover returns true if FSM can recover from current state
func (g *BaseGraphFSM) CanRecover() bool {
	state := g.GetState()

	// Can always recover from ERROR state
	if state == GraphError {
		g.mu.RLock()
		canRetry := g.retryCount <= g.retryConfig.MaxRetries
		hasPermanentErr := false

		if g.lastError != nil {
			var fsmErr *FSMError
			if IsFSMError(g.lastError) && errors.As(g.lastError, &fsmErr) {
				hasPermanentErr = fsmErr.IsPermanent()
			}
		}
		g.mu.RUnlock()

		// Can recover if retries available or no permanent error
		return canRetry || !hasPermanentErr
	}

	// Can reset from IDLE or READY
	return state == GraphIdle || state == GraphReady
}

// GetTransitionHistory returns the N most recent state transitions
func (g *BaseGraphFSM) GetTransitionHistory(n int) []HistoryEntry {
	return g.transitionHistory.GetRecent(n)
}

// GetFailedTransitions returns all failed transitions
func (g *BaseGraphFSM) GetFailedTransitions() []HistoryEntry {
	return g.transitionHistory.GetFailedTransitions()
}

// GetDiagnostics returns diagnostic information about the FSM
func (g *BaseGraphFSM) GetDiagnostics() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	diag := map[string]interface{}{
		"current_state":         g.state,
		"last_error":           g.lastError,
		"retry_count":          g.retryCount,
		"last_operation":       g.lastOperation,
		"last_failed_state":     g.lastFailedState,
		"can_recover":          g.canRetryNoLock(),
		"history_entries":       g.transitionHistory.Len(),
		"failed_transitions":    len(g.transitionHistory.GetFailedTransitions()),
	}

	// Add retry config
	diag["max_retries"] = g.retryConfig.MaxRetries
	diag["base_delay_ms"] = g.retryConfig.BaseDelay.Milliseconds()
	diag["max_delay_ms"] = g.retryConfig.MaxDelay.Milliseconds()

	return diag
}

// canRetryNoLock checks if retry is allowed (must be called with lock held)
func (g *BaseGraphFSM) canRetryNoLock() bool {
	// Check if we have an FSMError that is transient
	if g.lastError != nil {
		var fsmErr *FSMError
		if IsFSMError(g.lastError) && errors.As(g.lastError, &fsmErr) {
			if fsmErr.IsTransient() && g.retryCount < g.retryConfig.MaxRetries {
				return true
			}
		}
	}
	return g.retryCount < g.retryConfig.MaxRetries
}

// ClearHistory clears the transition history
func (g *BaseGraphFSM) ClearHistory() {
	g.transitionHistory.Clear()
}

// ClearRollbackSnapshots clears all rollback snapshots
func (g *BaseGraphFSM) ClearRollbackSnapshots() {
	g.rollbackManager.ClearAll()
}
