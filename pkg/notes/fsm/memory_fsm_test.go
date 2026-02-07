package fsm

import (
	"fmt"
	"testing"
	"time"
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

func TestBaseMemoryFSM_CanTransition(t *testing.T) {
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

	// Valid transition
	if !fsm.CanTransition(MemoryStateLearned) {
		t.Error("expected draft to learned to be valid")
	}

	// Same state
	if !fsm.CanTransition(MemoryStateDraft) {
		t.Error("expected same state to be valid")
	}

	// Invalid transition
	if fsm.CanTransition(MemoryStateLearning) {
		t.Error("expected draft to learning to be invalid")
	}

	// Transition to learned
	fsm.Transition(MemoryStateLearned, "test", "user")

	// Check new state is valid
	if !fsm.CanTransition(MemoryStateArchived) {
		t.Error("expected learned to archived to be valid")
	}
}

func TestBaseMemoryFSM_NilStorage(t *testing.T) {
	obj := &testMemoryObject{urn: "test-urn", state: MemoryStateDraft}
	fsm, err := NewBaseMemoryFSM(obj, MemoryStateDraft, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Operations should work without storage
	err = fsm.Transition(MemoryStateLearned, "test", "user")
	if err != nil {
		t.Fatal(err)
	}

	// GetState should work
	state := fsm.GetState()
	if state.State != MemoryStateLearned {
		t.Errorf("expected state learned, got %s", state.State)
	}

	// GetHistory should work
	history := fsm.GetHistory()
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}

	// LoadState should fail
	err = fsm.LoadState()
	if err == nil {
		t.Error("expected error for LoadState with nil storage")
	}
}

func TestBaseMemoryFSM_TransitionWithNilStorage(t *testing.T) {
	obj := &testMemoryObject{urn: "test-urn", state: MemoryStateDraft}
	fsm, err := NewBaseMemoryFSM(obj, MemoryStateDraft, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Multiple transitions should work
	err = fsm.Transition(MemoryStateLearned, "learned", "user1")
	if err != nil {
		t.Fatal(err)
	}

	err = fsm.Transition(MemoryStateArchived, "archived", "user2")
	if err != nil {
		t.Fatal(err)
	}

	if fsm.GetStateValue() != MemoryStateArchived {
		t.Errorf("expected state archived, got %s", fsm.GetStateValue())
	}
}

func TestBaseMemoryFSM_StateObject(t *testing.T) {
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

	state := fsm.GetState()

	// Check URN
	if state.URN != "test-urn" {
		t.Errorf("expected URN test-urn, got %s", state.URN)
	}

	// Check initial state
	if state.State != MemoryStateDraft {
		t.Errorf("expected state draft, got %s", state.State)
	}

	// Check timestamps
	if state.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}

	if state.UpdatedAt.IsZero() {
		t.Error("expected non-zero updated_at")
	}
}

func TestBaseMemoryFSM_HistoryIsImmutable(t *testing.T) {
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

	// Perform transition
	fsm.Transition(MemoryStateLearned, "test", "user")

	// Get history
	history1 := fsm.GetHistory()
	initialLen := len(history1)

	// Modify returned history
	history1 = append(history1, StateHistory{})

	// Get history again
	history2 := fsm.GetHistory()

	// Should not be affected
	if len(history2) != initialLen {
		t.Errorf("history should be immutable, got %d entries", len(history2))
	}
}

func TestBaseMemoryFSM_TransitionUpdatesUpdatedAt(t *testing.T) {
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

	state1 := fsm.GetState()
	initialUpdatedAt := state1.UpdatedAt

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Perform transition
	fsm.Transition(MemoryStateLearned, "test", "user")

	state2 := fsm.GetState()

	if !state2.UpdatedAt.After(initialUpdatedAt) {
		t.Error("expected updated_at to increase after transition")
	}
}

func TestBaseMemoryFSM_InvalidTransitionRollback(t *testing.T) {
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

	// Invalid transition
	err = fsm.Transition(MemoryStateLearning, "invalid", "user")
	if err == nil {
		t.Error("expected error for invalid transition")
	}

	// State should not change
	if fsm.GetStateValue() != MemoryStateDraft {
		t.Errorf("expected state to remain draft, got %s", fsm.GetStateValue())
	}

	// History should not be updated
	if len(fsm.GetHistory()) != 0 {
		t.Error("expected no history entries for invalid transition")
	}
}

func TestBaseMemoryFSM_SameStateTransition(t *testing.T) {
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

	// Transition to same state
	err = fsm.Transition(MemoryStateDraft, "no change", "user")
	if err != nil {
		t.Fatal(err)
	}

	// History should record same state transition
	history := fsm.GetHistory()
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}

	if history[0].FromState != MemoryStateDraft || history[0].ToState != MemoryStateDraft {
		t.Error("expected same state transition in history")
	}
}

func TestBaseMemoryFSM_MultipleTransitions(t *testing.T) {
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

	// Multiple transitions
	transitions := []MemoryState{
		MemoryStateLearned,
		MemoryStateArchived,
	}

	for _, targetState := range transitions {
		err = fsm.Transition(targetState, "test", "user")
		if err != nil {
			t.Fatalf("transition to %s failed: %v", targetState, err)
		}
	}

	// Check final state
	if fsm.GetStateValue() != MemoryStateArchived {
		t.Errorf("expected final state archived, got %s", fsm.GetStateValue())
	}

	// Check history
	history := fsm.GetHistory()
	if len(history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(history))
	}
}

func TestBaseMemoryFSM_GetStateValue(t *testing.T) {
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

	// Get initial state value
	if fsm.GetStateValue() != MemoryStateDraft {
		t.Errorf("expected state draft, got %s", fsm.GetStateValue())
	}

	// Transition
	fsm.Transition(MemoryStateLearned, "test", "user")

	// Check updated state value
	if fsm.GetStateValue() != MemoryStateLearned {
		t.Errorf("expected state learned, got %s", fsm.GetStateValue())
	}
}

func TestBaseMemoryFSM_ThreadSafe(t *testing.T) {
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

	done := make(chan bool)

	// Concurrent GetStateValue
	for i := 0; i < 10; i++ {
		go func() {
			fsm.GetStateValue()
			done <- true
		}()
	}

	// Concurrent GetState
	for i := 0; i < 10; i++ {
		go func() {
			fsm.GetState()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestBaseMemoryFSM_HistoryOrder(t *testing.T) {
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
	fsm.Transition(MemoryStateLearned, "first", "user1")
	fsm.Transition(MemoryStateArchived, "second", "user2")

	history := fsm.GetHistory()

	// Check order
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}

	if history[0].ToState != MemoryStateLearned {
		t.Errorf("expected first transition to learned, got %s", history[0].ToState)
	}

	if history[1].ToState != MemoryStateArchived {
		t.Errorf("expected second transition to archived, got %s", history[1].ToState)
	}

	// Check reasons
	if history[0].Reason != "first" {
		t.Errorf("expected reason 'first', got %s", history[0].Reason)
	}

	if history[1].Reason != "second" {
		t.Errorf("expected reason 'second', got %s", history[1].Reason)
	}

	// Check user IDs
	if history[0].UserID != "user1" {
		t.Errorf("expected user ID 'user1', got %s", history[0].UserID)
	}

	if history[1].UserID != "user2" {
		t.Errorf("expected user ID 'user2', got %s", history[1].UserID)
	}
}

func TestBaseMemoryFSM_StatePersistenceReload(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"

	// First FSM with storage
	storage1, err := NewBoltDBStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	obj1 := &testMemoryObject{urn: "test-urn", state: MemoryStateDraft}
	fsm1, err := NewBaseMemoryFSM(obj1, MemoryStateDraft, storage1)
	if err != nil {
		t.Fatal(err)
	}

	// Perform multiple transitions
	err = fsm1.Transition(MemoryStateLearned, "learned", "user1")
	if err != nil {
		t.Fatal(err)
	}
	err = fsm1.Transition(MemoryStateArchived, "archived", "user2")
	if err != nil {
		t.Fatal(err)
	}

	// Close storage
	storage1.Close()

	// Reopen storage and load state directly
	storage2, err := NewBoltDBStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer storage2.Close()

	// Load saved state directly from storage
	savedState, err := storage2.LoadMemoryFSMState("test-urn")
	if err != nil {
		t.Fatal(err)
	}

	// Check state was saved correctly
	if savedState.State != MemoryStateArchived {
		t.Errorf("expected state archived, got %s", savedState.State)
	}

	// Check history was saved correctly
	if len(savedState.StateHistory) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(savedState.StateHistory))
	}
}

func TestBaseMemoryFSM_EmptyHistory(t *testing.T) {
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

	// No transitions performed
	history := fsm.GetHistory()
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d entries", len(history))
	}
}

func TestBaseMemoryFSM_StateObjectCopy(t *testing.T) {
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

	state1 := fsm.GetState()
	state2 := fsm.GetState()

	// Modify state1
	state1.State = MemoryStateLearned

	// state2 should be unaffected
	if state2.State == MemoryStateLearned {
		t.Error("state object should be copied, not referenced")
	}
}

func TestParseMemoryState_InvalidInput(t *testing.T) {
	invalidInputs := []string{
		"invalid",
		"UNKNOWN",
		"Draft",
		"LEARNING",
		" ",
		"\t",
		"\n",
	}

	for _, input := range invalidInputs {
		_, err := ParseMemoryState(input)
		if err == nil {
			t.Errorf("expected error for input '%s'", input)
		}
	}
}

func TestValidateMemoryTransition_AllCombinations(t *testing.T) {
	allStates := []MemoryState{
		MemoryStateNew,
		MemoryStateDraft,
		MemoryStateLearning,
		MemoryStateReviewed,
		MemoryStateLearned,
		MemoryStateMastered,
		MemoryStateArchived,
	}

	for _, from := range allStates {
		for _, to := range allStates {
			err := ValidateMemoryTransition(from, to)

			// Same state is always valid
			if from == to {
				if err != nil {
					t.Errorf("same state %s should always be valid", from)
				}
				continue
			}

			// Check if transition is defined
			isValid := validMemoryTransitions[from][to]
			if isValid && err != nil {
				t.Errorf("expected valid transition %s -> %s", from, to)
			}
			if !isValid && err == nil {
				t.Errorf("expected invalid transition %s -> %s", from, to)
			}
		}
	}
}

func TestBaseMemoryFSM_TransitionWithDifferentUsers(t *testing.T) {
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

	// Transition with different users
	users := []string{"alice", "bob", "charlie"}
	for i, user := range users {
		var targetState MemoryState
		if i == 0 {
			targetState = MemoryStateLearned
		} else {
			targetState = MemoryStateArchived
		}
		err := fsm.Transition(targetState, fmt.Sprintf("action %d", i), user)
		if err != nil {
			t.Fatal(err)
		}
	}

	history := fsm.GetHistory()
	if len(history) != len(users) {
		t.Errorf("expected %d history entries, got %d", len(users), len(history))
	}

	// Verify user IDs in history
	for i, user := range users {
		if history[i].UserID != user {
			t.Errorf("expected user ID %s at index %d, got %s", user, i, history[i].UserID)
		}
	}
}

func TestBaseMemoryFSM_Timestamps(t *testing.T) {
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

	state := fsm.GetState()

	// CreatedAt and UpdatedAt should be set initially
	if state.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if state.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	// UpdatedAt should be >= CreatedAt
	if state.UpdatedAt.Before(state.CreatedAt) {
		t.Error("expected UpdatedAt to be >= CreatedAt")
	}
}
