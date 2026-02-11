package fsm

import (
	"fmt"
	"time"
)

// MemoryState represents the unified FSM state for notes and memory cards
type MemoryState string

const (
	// MemoryStateDraft is the initial editing state for notes
	MemoryStateDraft MemoryState = "draft"
	// MemoryStateNew is the initial state for memory cards
	MemoryStateNew MemoryState = "new"
	// MemoryStateLearning indicates active learning in progress
	MemoryStateLearning MemoryState = "learning"
	// MemoryStateReviewed indicates post-first-review state for cards
	MemoryStateReviewed MemoryState = "reviewed"
	// MemoryStateLearned indicates a note marked as complete
	MemoryStateLearned MemoryState = "learned"
	// MemoryStateMastered indicates a card is fully learned
	MemoryStateMastered MemoryState = "mastered"
	// MemoryStateArchived indicates the item is archived
	MemoryStateArchived MemoryState = "archived"
)

// Valid transitions for MemoryFSM
var validMemoryTransitions = map[MemoryState]map[MemoryState]bool{
	MemoryStateDraft: {
		MemoryStateLearned:  true,
		MemoryStateArchived: true,
	},
	MemoryStateNew: {
		MemoryStateLearning: true,
		MemoryStateArchived: true,
	},
	MemoryStateLearning: {
		MemoryStateReviewed: true,
		MemoryStateMastered: true,
		MemoryStateArchived: true,
	},
	MemoryStateReviewed: {
		MemoryStateLearning: true,
		MemoryStateMastered: true,
		MemoryStateArchived: true,
	},
	MemoryStateLearned: {
		// Learned notes can remain editable while maintaining status
		MemoryStateLearned:  true,
		MemoryStateArchived: true,
	},
	MemoryStateMastered: {
		MemoryStateArchived: true,
	},
	MemoryStateArchived: {},
}

// StateHistory records each state transition
type StateHistory struct {
	FromState MemoryState `json:"from_state"`
	ToState   MemoryState `json:"to_state"`
	Timestamp time.Time   `json:"timestamp"`
	Reason    string      `json:"reason"`
	UserID    string      `json:"user_id,omitempty"`
}

// MemoryObject is the interface for all learning objects (notes and memory cards)
type MemoryObject interface {
	GetURN() string
	GetMemoryFSMState() MemoryState
	SetMemoryFSMState(state MemoryState) error
}

// ValidateMemoryTransition checks if a memory state transition is valid
func ValidateMemoryTransition(from, to MemoryState) error {
	if from == to {
		return nil // Same state is always valid
	}

	nextStates, ok := validMemoryTransitions[from]
	if !ok {
		return fmt.Errorf("unknown memory state: %s", from)
	}

	if !nextStates[to] {
		return fmt.Errorf("invalid memory state transition: %s -> %s", from, to)
	}

	return nil
}

// ParseMemoryState parses a string into a MemoryState
func ParseMemoryState(s string) (MemoryState, error) {
	switch s {
	case string(MemoryStateDraft):
		return MemoryStateDraft, nil
	case string(MemoryStateNew):
		return MemoryStateNew, nil
	case string(MemoryStateLearning):
		return MemoryStateLearning, nil
	case string(MemoryStateReviewed):
		return MemoryStateReviewed, nil
	case string(MemoryStateLearned):
		return MemoryStateLearned, nil
	case string(MemoryStateMastered):
		return MemoryStateMastered, nil
	case string(MemoryStateArchived):
		return MemoryStateArchived, nil
	default:
		return "", fmt.Errorf("unknown memory state: %q", s)
	}
}

// MemoryFSMState represents the persisted FSM state for an object
type MemoryFSMState struct {
	URN          string         `json:"urn"`
	State        MemoryState    `json:"state"`
	StateHistory []StateHistory `json:"state_history"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
