package fsm

import (
	"testing"
)

func TestValidateMemoryTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     MemoryState
		to       MemoryState
		expected bool
	}{
		{"draft to learned", MemoryStateDraft, MemoryStateLearned, true},
		{"draft to archived", MemoryStateDraft, MemoryStateArchived, true},
		{"draft to learning (invalid)", MemoryStateDraft, MemoryStateLearning, false},
		{"new to learning", MemoryStateNew, MemoryStateLearning, true},
		{"new to archived", MemoryStateNew, MemoryStateArchived, true},
		{"new to reviewed (invalid)", MemoryStateNew, MemoryStateReviewed, false},
		{"learning to reviewed", MemoryStateLearning, MemoryStateReviewed, true},
		{"learning to mastered", MemoryStateLearning, MemoryStateMastered, true},
		{"learning to archived", MemoryStateLearning, MemoryStateArchived, true},
		{"reviewed to learning", MemoryStateReviewed, MemoryStateLearning, true},
		{"reviewed to mastered", MemoryStateReviewed, MemoryStateMastered, true},
		{"reviewed to archived", MemoryStateReviewed, MemoryStateArchived, true},
		{"mastered to archived", MemoryStateMastered, MemoryStateArchived, true},
		{"mastered to learning (invalid)", MemoryStateMastered, MemoryStateLearning, false},
		{"archived to new (invalid)", MemoryStateArchived, MemoryStateNew, false},
		{"same state", MemoryStateDraft, MemoryStateDraft, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMemoryTransition(tt.from, tt.to)
			if tt.expected && err != nil {
				t.Errorf("expected transition %s -> %s to be valid, got error: %v", tt.from, tt.to, err)
			}
			if !tt.expected && err == nil {
				t.Errorf("expected transition %s -> %s to be invalid, got no error", tt.from, tt.to)
			}
		})
	}
}

func TestParseMemoryState(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  MemoryState
		wantError bool
	}{
		{"draft", "draft", MemoryStateDraft, false},
		{"new", "new", MemoryStateNew, false},
		{"learning", "learning", MemoryStateLearning, false},
		{"reviewed", "reviewed", MemoryStateReviewed, false},
		{"learned", "learned", MemoryStateLearned, false},
		{"mastered", "mastered", MemoryStateMastered, false},
		{"archived", "archived", MemoryStateArchived, false},
		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMemoryState(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for input %s, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBaseMemoryFSM_Transition(t *testing.T) {
	tests := []struct {
		name         string
		initialState MemoryState
		targetState  MemoryState
		wantError    bool
	}{
		{"draft to learned", MemoryStateDraft, MemoryStateLearned, false},
		{"new to learning", MemoryStateNew, MemoryStateLearning, false},
		{"learning to reviewed", MemoryStateLearning, MemoryStateReviewed, false},
		{"reviewed to mastered", MemoryStateReviewed, MemoryStateMastered, false},
		{"invalid transition", MemoryStateArchived, MemoryStateNew, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
			if err != nil {
				t.Fatal(err)
			}
			defer storage.Close()

			obj := &testMemoryObject{urn: "test-urn", state: tt.initialState}
			fsm, err := NewBaseMemoryFSM(obj, tt.initialState, storage)
			if err != nil {
				t.Fatal(err)
			}

			err = fsm.Transition(tt.targetState, "test reason", "test-user")
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for transition %s -> %s", tt.initialState, tt.targetState)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if obj.state != tt.targetState {
					t.Errorf("expected state %s, got %s", tt.targetState, obj.state)
				}
			}
		})
	}
}

func TestBaseMemoryFSM_StateHistory(t *testing.T) {
	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	obj := &testMemoryObject{urn: "test-urn", state: MemoryStateDraft}
	fsm, err := NewBaseMemoryFSM(obj, MemoryStateDraft, storage)
	if err != nil {
		t.Fatal(err)
	}

	// Perform transitions
	err = fsm.Transition(MemoryStateLearned, "first transition", "user1")
	if err != nil {
		t.Fatal(err)
	}

	err = fsm.Transition(MemoryStateArchived, "second transition", "user1")
	if err != nil {
		t.Fatal(err)
	}

	state := fsm.GetState()
	if len(state.StateHistory) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(state.StateHistory))
	}

	// Check first history entry
	if state.StateHistory[0].FromState != MemoryStateDraft {
		t.Errorf("expected from_state %s, got %s", MemoryStateDraft, state.StateHistory[0].FromState)
	}
	if state.StateHistory[0].ToState != MemoryStateLearned {
		t.Errorf("expected to_state %s, got %s", MemoryStateLearned, state.StateHistory[0].ToState)
	}

	// Check second history entry
	if state.StateHistory[1].FromState != MemoryStateLearned {
		t.Errorf("expected from_state %s, got %s", MemoryStateLearned, state.StateHistory[1].FromState)
	}
	if state.StateHistory[1].ToState != MemoryStateArchived {
		t.Errorf("expected to_state %s, got %s", MemoryStateArchived, state.StateHistory[1].ToState)
	}
}

func TestBaseMemoryFSM_StatePersistence(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	storage1, err := NewBoltDBStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	obj := &testMemoryObject{urn: "test-urn", state: MemoryStateDraft}
	fsm1, err := NewBaseMemoryFSM(obj, MemoryStateDraft, storage1)
	if err != nil {
		t.Fatal(err)
	}

	// Transition to learned
	err = fsm1.Transition(MemoryStateLearned, "test transition", "user1")
	if err != nil {
		t.Fatal(err)
	}

	// Close and reopen storage
	storage1.Close()

	storage2, err := NewBoltDBStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer storage2.Close()

	// Load saved state
	savedState, err := storage2.LoadMemoryFSMState("test-urn")
	if err != nil {
		t.Fatal(err)
	}

	if savedState.State != MemoryStateLearned {
		t.Errorf("expected state %s, got %s", MemoryStateLearned, savedState.State)
	}
	if len(savedState.StateHistory) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(savedState.StateHistory))
	}
}

// testMemoryObject is a test implementation of MemoryObject
type testMemoryObject struct {
	urn   string
	state MemoryState
}

func (o *testMemoryObject) GetURN() string {
	return o.urn
}

func (o *testMemoryObject) GetMemoryFSMState() MemoryState {
	return o.state
}

func (o *testMemoryObject) SetMemoryFSMState(state MemoryState) error {
	o.state = state
	return nil
}
