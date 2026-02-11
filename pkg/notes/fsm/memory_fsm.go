package fsm

import (
	"fmt"
	"sync"
	"time"
)

// BaseMemoryFSM implements MemoryFSM for notes and memory cards
type BaseMemoryFSM struct {
	mu           sync.RWMutex
	object       MemoryObject
	state        MemoryState
	stateHistory []StateHistory
	createdAt    time.Time
	updatedAt    time.Time
	storage      Storage // FSM state persistence
}

// NewBaseMemoryFSM creates a new MemoryFSM for a learning object
func NewBaseMemoryFSM(object MemoryObject, initialState MemoryState, storage Storage) (*BaseMemoryFSM, error) {
	now := time.Now()
	fsm := &BaseMemoryFSM{
		object:       object,
		state:        initialState,
		stateHistory: make([]StateHistory, 0),
		createdAt:    now,
		updatedAt:    now,
		storage:      storage,
	}

	// Save initial state
	if storage != nil {
		if err := storage.SaveMemoryFSMState(object.GetURN(), fsm.GetState()); err != nil {
			return nil, fmt.Errorf("failed to save initial FSM state: %w", err)
		}
	}

	return fsm, nil
}

// GetState returns the current FSM state
func (m *BaseMemoryFSM) GetState() *MemoryFSMState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &MemoryFSMState{
		URN:          m.object.GetURN(),
		State:        m.state,
		StateHistory: append([]StateHistory{}, m.stateHistory...),
		CreatedAt:    m.createdAt,
		UpdatedAt:    m.updatedAt,
	}
}

// GetStateValue returns the current state value
func (m *BaseMemoryFSM) GetStateValue() MemoryState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// Transition validates and executes a state change
func (m *BaseMemoryFSM) Transition(newState MemoryState, reason string, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate transition
	if err := ValidateMemoryTransition(m.state, newState); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}

	// Record state change
	history := StateHistory{
		FromState: m.state,
		ToState:   newState,
		Timestamp: time.Now(),
		Reason:    reason,
		UserID:    userID,
	}

	// Update state
	oldState := m.state
	m.state = newState
	m.stateHistory = append(m.stateHistory, history)
	m.updatedAt = time.Now()

	// Update object state
	if err := m.object.SetMemoryFSMState(newState); err != nil {
		// Rollback on error
		m.state = oldState
		m.stateHistory = m.stateHistory[:len(m.stateHistory)-1]
		return fmt.Errorf("failed to update object state: %w", err)
	}

	// Persist state
	if m.storage != nil {
		// Capture state to save (without acquiring lock again)
		stateToSave := &MemoryFSMState{
			URN:          m.object.GetURN(),
			State:        newState,
			StateHistory: m.stateHistory,
			CreatedAt:    m.createdAt,
			UpdatedAt:    m.updatedAt,
		}
		if err := m.storage.SaveMemoryFSMState(m.object.GetURN(), stateToSave); err != nil {
			// Note: We don't rollback here since the object state was updated
			// The storage error should be logged but shouldn't block the transition
			return fmt.Errorf("failed to persist FSM state: %w", err)
		}
	}

	return nil
}

// GetHistory returns the state transition history
func (m *BaseMemoryFSM) GetHistory() []StateHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := make([]StateHistory, len(m.stateHistory))
	copy(history, m.stateHistory)
	return history
}

// CanTransition checks if a transition to the given state is valid
func (m *BaseMemoryFSM) CanTransition(to MemoryState) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state == to || validMemoryTransitions[m.state][to]
}

// LoadState restores FSM state from storage
func (m *BaseMemoryFSM) LoadState() error {
	if m.storage == nil {
		return fmt.Errorf("no storage configured")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.storage.LoadMemoryFSMState(m.object.GetURN())
	if err != nil {
		return fmt.Errorf("failed to load FSM state: %w", err)
	}

	m.state = state.State
	m.stateHistory = state.StateHistory
	m.createdAt = state.CreatedAt
	m.updatedAt = state.UpdatedAt

	return nil
}
