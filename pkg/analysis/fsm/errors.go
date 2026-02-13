package fsm

import (
	"fmt"
	"sync"
	"time"
)

// TransitionErrorType classifies state transition errors
type TransitionErrorType string

const (
	// ErrInvalidTransition - transition not allowed by FSM rules
	ErrInvalidTransition TransitionErrorType = "INVALID_TRANSITION"
	// ErrTransitionFailed - transition execution failed
	ErrTransitionFailed TransitionErrorType = "TRANSITION_FAILED"
	// ErrRollbackFailed - rollback to previous state failed
	ErrRollbackFailed TransitionErrorType = "ROLLBACK_FAILED"
	// ErrRecoveryExhausted - all recovery attempts exhausted
	ErrRecoveryExhausted TransitionErrorType = "RECOVERY_EXHAUSTED"
)

// StateTransitionError represents a state transition error with recovery context
type StateTransitionError struct {
	// ErrorType is the category of transition error
	ErrorType TransitionErrorType
	// FromState is the state before transition attempt
	FromState string
	// ToState is the target state that failed
	ToState string
	// Cause is the underlying error
	Cause error
	// Timestamp when error occurred
	Timestamp time.Time
	// RecoveryAttempts made so far
	RecoveryAttempts int
	// CanRecover indicates if recovery is possible
	CanRecover bool
	// RolledBack indicates if state was rolled back
	RolledBack bool
}

// Error returns the error message
func (e *StateTransitionError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("[%s] %s -> %s at %s",
			e.ErrorType, e.FromState, e.ToState,
			e.Timestamp.Format(time.RFC3339))
	}
	return fmt.Sprintf("[%s] %s -> %s: %v (at %s)",
		e.ErrorType, e.FromState, e.ToState, e.Cause,
		e.Timestamp.Format(time.RFC3339))
}

// Unwrap returns the underlying cause
func (e *StateTransitionError) Unwrap() error {
	return e.Cause
}

// IsRecoverable returns true if the error is recoverable
func (e *StateTransitionError) IsRecoverable() bool {
	return e.CanRecover && e.RecoveryAttempts < maxRecoveryAttempts
}

// NewTransitionError creates a new transition error
func NewTransitionError(errorType TransitionErrorType, from, to string, cause error) *StateTransitionError {
	return &StateTransitionError{
		ErrorType:         errorType,
		FromState:         from,
		ToState:           to,
		Cause:             cause,
		Timestamp:         time.Now(),
		RecoveryAttempts:  0,
		CanRecover:        true,
		RolledBack:        false,
	}
}

const (
	// maxRecoveryAttempts is the maximum number of recovery attempts
	maxRecoveryAttempts = 3
)

// TransitionHistory records state transition history for debugging
type TransitionHistory struct {
	mu     sync.RWMutex
	entry  []HistoryEntry
	maxLen int
}

// HistoryEntry represents a single state transition entry
type HistoryEntry struct {
	From      string        `json:"from"`
	To        string        `json:"to"`
	Timestamp time.Time     `json:"timestamp"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// NewTransitionHistory creates a new transition history
func NewTransitionHistory(maxLen int) *TransitionHistory {
	if maxLen <= 0 {
		maxLen = 100 // default max entries
	}
	return &TransitionHistory{
		entry:  make([]HistoryEntry, 0, maxLen),
		maxLen: maxLen,
	}
}

// Record records a state transition
func (h *TransitionHistory) Record(from, to string, success bool, err error, duration time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry := HistoryEntry{
		From:      from,
		To:        to,
		Timestamp: time.Now(),
		Success:   success,
		Duration:  duration,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Add to history
	h.entry = append(h.entry, entry)

	// Trim if exceeding max length
	if len(h.entry) > h.maxLen {
		// Keep only the most recent entries
		h.entry = h.entry[len(h.entry)-h.maxLen:]
	}
}

// GetRecent returns the N most recent transitions
func (h *TransitionHistory) GetRecent(n int) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || n > len(h.entry) {
		n = len(h.entry)
	}

	// Return the most recent n entries
	start := len(h.entry) - n
	if start < 0 {
		start = 0
	}

	result := make([]HistoryEntry, n)
	copy(result, h.entry[start:])
	return result
}

// GetFailedTransitions returns only failed transitions
func (h *TransitionHistory) GetFailedTransitions() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var failed []HistoryEntry
	for _, e := range h.entry {
		if !e.Success {
			failed = append(failed, e)
		}
	}
	return failed
}

// Clear clears all history entries
func (h *TransitionHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entry = make([]HistoryEntry, 0, h.maxLen)
}

// Len returns the number of entries
func (h *TransitionHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entry)
}

// StateSnapshot captures a snapshot of state for rollback
type StateSnapshot struct {
	State      string      `json:"state"`
	Timestamp  time.Time   `json:"timestamp"`
	Data       interface{} `json:"data,omitempty"`
	SequenceID int64       `json:"sequence_id"`
}

// RollbackManager manages state snapshots for rollback
type RollbackManager struct {
	mu         sync.RWMutex
	snapshots  map[string][]StateSnapshot // keyed by state name
	maxPerState int
	seqCounter int64
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(maxPerState int) *RollbackManager {
	if maxPerState <= 0 {
		maxPerState = 5 // default snapshots per state
	}
	return &RollbackManager{
		snapshots:  make(map[string][]StateSnapshot),
		maxPerState: maxPerState,
		seqCounter: 0,
	}
}

// SaveSnapshot saves a state snapshot
func (r *RollbackManager) SaveSnapshot(state string, data interface{}) StateSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seqCounter++
	snapshot := StateSnapshot{
		State:      state,
		Timestamp:  time.Now(),
		Data:       data,
		SequenceID: r.seqCounter,
	}

	// Initialize slice if needed
	if r.snapshots[state] == nil {
		r.snapshots[state] = make([]StateSnapshot, 0, r.maxPerState)
	}

	// Add snapshot
	r.snapshots[state] = append(r.snapshots[state], snapshot)

	// Trim if exceeding max
	if len(r.snapshots[state]) > r.maxPerState {
		// Keep only the most recent
		r.snapshots[state] = r.snapshots[state][len(r.snapshots[state])-r.maxPerState:]
	}

	return snapshot
}

// GetLatestSnapshot returns the most recent snapshot for a state
func (r *RollbackManager) GetLatestSnapshot(state string) (StateSnapshot, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshots, exists := r.snapshots[state]
	if !exists || len(snapshots) == 0 {
		return StateSnapshot{}, false
	}

	return snapshots[len(snapshots)-1], true
}

// ClearSnapshots removes all snapshots for a state
func (r *RollbackManager) ClearSnapshots(state string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.snapshots, state)
}

// ClearAll removes all snapshots
func (r *RollbackManager) ClearAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshots = make(map[string][]StateSnapshot)
}
