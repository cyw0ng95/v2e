package fsm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StateTransition represents a state change with metadata
type StateTransition struct {
	FromState string
	ToState   string
	Timestamp time.Time
	Reason    string
	UserID    string
}

// TransitionValidator defines a function to validate state transitions
type TransitionValidator func(from, to interface{}) error

// FSMState represents a generic FSM state for persistence
type FSMState struct {
	State        interface{}
	StateHistory []StateTransition
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActivity time.Time
}

// BaseFSM provides common FSM functionality for state machines
type BaseFSM struct {
	mu           sync.RWMutex
	state        interface{}
	stateHistory []StateTransition
	createdAt    time.Time
	updatedAt    time.Time
	lastActivity time.Time
	storage      interface{}
	validator    TransitionValidator
}

// NewBaseFSM creates a new base FSM
func NewBaseFSM(initialState interface{}, storage interface{}, validator TransitionValidator) *BaseFSM {
	now := time.Now()
	return &BaseFSM{
		state:        initialState,
		stateHistory: make([]StateTransition, 0),
		createdAt:    now,
		updatedAt:    now,
		lastActivity: now,
		storage:      storage,
		validator:    validator,
	}
}

// GetState returns the current state
func (f *BaseFSM) GetState() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// GetStateValue returns the current state as a string
func (f *BaseFSM) GetStateValue() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return fmt.Sprintf("%v", f.state)
}

// SetState sets the state directly without validation (use with caution)
func (f *BaseFSM) SetState(newState interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = newState
	f.updatedAt = time.Now()
	f.lastActivity = time.Now()
}

// Transition validates and executes a state change
func (f *BaseFSM) Transition(newState interface{}, reason, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	oldState := f.state

	if f.validator != nil {
		if err := f.validator(oldState, newState); err != nil {
			return fmt.Errorf("invalid transition from %v to %v: %w", oldState, newState, err)
		}
	}

	transition := StateTransition{
		FromState: fmt.Sprintf("%v", oldState),
		ToState:   fmt.Sprintf("%v", newState),
		Timestamp: time.Now(),
		Reason:    reason,
		UserID:    userID,
	}

	f.state = newState
	f.stateHistory = append(f.stateHistory, transition)
	f.updatedAt = time.Now()
	f.lastActivity = time.Now()

	return nil
}

// GetHistory returns state transition history
func (f *BaseFSM) GetHistory() []StateTransition {
	f.mu.RLock()
	defer f.mu.RUnlock()

	history := make([]StateTransition, len(f.stateHistory))
	copy(history, f.stateHistory)
	return history
}

// UpdateActivity updates the last activity timestamp
func (f *BaseFSM) UpdateActivity() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastActivity = time.Now()
}

// GetLastActivity returns the last activity timestamp
func (f *BaseFSM) GetLastActivity() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lastActivity
}

// GetCreatedAt returns the creation timestamp
func (f *BaseFSM) GetCreatedAt() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.createdAt
}

// GetUpdatedAt returns the last updated timestamp
func (f *BaseFSM) GetUpdatedAt() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.updatedAt
}

// SaveStateWithContext persists FSM state with context timeout
func (f *BaseFSM) SaveStateWithContext(ctx context.Context, saveFunc func(*FSMState) error) error {
	if saveFunc == nil {
		return nil
	}

	f.mu.RLock()

	state := &FSMState{
		State:        f.state,
		StateHistory: make([]StateTransition, len(f.stateHistory)),
		CreatedAt:    f.createdAt,
		UpdatedAt:    f.updatedAt,
		LastActivity: f.lastActivity,
	}
	copy(state.StateHistory, f.stateHistory)

	f.mu.RUnlock()

	type result struct {
		err error
	}
	resultCh := make(chan result, 1)

	go func() {
		resultCh <- result{saveFunc(state)}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("save state timeout: %w", ctx.Err())
	case res := <-resultCh:
		return res.err
	}
}

// SaveState persists FSM state
func (f *BaseFSM) SaveState(saveFunc func(*FSMState) error) error {
	return f.SaveStateWithContext(context.Background(), saveFunc)
}

// LoadStateWithContext loads FSM state with context timeout
func (f *BaseFSM) LoadStateWithContext(ctx context.Context, loadFunc func() (*FSMState, error)) error {
	if loadFunc == nil {
		return fmt.Errorf("no load function provided")
	}

	type result struct {
		state *FSMState
		err   error
	}
	resultCh := make(chan result, 1)

	go func() {
		state, err := loadFunc()
		resultCh <- result{state, err}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("load state timeout: %w", ctx.Err())
	case res := <-resultCh:
		if res.err != nil {
			return res.err
		}

		f.mu.Lock()
		f.state = res.state.State
		f.stateHistory = make([]StateTransition, len(res.state.StateHistory))
		copy(f.stateHistory, res.state.StateHistory)
		f.createdAt = res.state.CreatedAt
		f.updatedAt = res.state.UpdatedAt
		f.lastActivity = res.state.LastActivity
		f.mu.Unlock()

		return nil
	}
}

// LoadState loads FSM state
func (f *BaseFSM) LoadState(loadFunc func() (*FSMState, error)) error {
	return f.LoadStateWithContext(context.Background(), loadFunc)
}

// IsStateEqual checks if current state equals the given state
func (f *BaseFSM) IsStateEqual(state interface{}) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == state
}
