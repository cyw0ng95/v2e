package fsm

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestValidateMemoryFSMState_Valid(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &MemoryFSMState{
		URN:   "v2e::card::1",
		State: MemoryStateNew,
		StateHistory: []StateHistory{
			{
				FromState: MemoryStateNew,
				ToState:   MemoryStateNew,
				Timestamp: time.Now(),
				Reason:    "initial",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storage.SaveMemoryFSMState("v2e::card::1", state); err != nil {
		t.Fatal(err)
	}

	if err := storage.ValidateMemoryFSMState("v2e::card::1"); err != nil {
		t.Fatalf("expected no error for valid state, got: %v", err)
	}
}

func TestValidateMemoryFSMState_InvalidURN(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &MemoryFSMState{
		URN:   "v2e::card::2",
		State: MemoryStateNew,
		StateHistory: []StateHistory{
			{
				FromState: MemoryStateNew,
				ToState:   MemoryStateNew,
				Timestamp: time.Now(),
				Reason:    "initial",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	storage.SaveMemoryFSMState("v2e::card::1", state)

	err = storage.ValidateMemoryFSMState("v2e::card::1")
	if err == nil {
		t.Fatal("expected error for URN mismatch")
	}
	if len(err.Error()) < 18 || err.Error()[:18] != "state URN mismatch" {
		t.Fatalf("expected URN mismatch error, got: %v", err)
	}
}

func TestValidateMemoryFSMState_InvalidState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &MemoryFSMState{
		URN:   "v2e::card::1",
		State: "invalid_state",
		StateHistory: []StateHistory{
			{
				FromState: MemoryStateNew,
				ToState:   MemoryStateNew,
				Timestamp: time.Now(),
				Reason:    "initial",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	storage.SaveMemoryFSMState("v2e::card::1", state)

	err = storage.ValidateMemoryFSMState("v2e::card::1")
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
	expected := "invalid memory state:"
	if len(err.Error()) < len(expected) || err.Error()[:len(expected)] != expected {
		t.Fatalf("expected invalid state error, got: %v", err)
	}
}

func TestValidateLearningFSMState_Valid(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &LearningFSMState{
		State:           LearningStateIdle,
		CurrentStrategy: "bfs",
		CurrentItemURN:  "",
		ViewedItems:     []string{},
		CompletedItems:  []string{},
		PathStack:       []string{},
		SessionStart:    time.Now(),
		LastActivity:    time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := storage.SaveLearningFSMState(state); err != nil {
		t.Fatal(err)
	}

	if err := storage.ValidateLearningFSMState(); err != nil {
		t.Fatalf("expected no error for valid learning state, got: %v", err)
	}
}

func TestValidateLearningFSMState_InvalidStrategy(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &LearningFSMState{
		State:           LearningStateIdle,
		CurrentStrategy: "invalid_strategy",
		CurrentItemURN:  "",
		ViewedItems:     []string{},
		CompletedItems:  []string{},
		PathStack:       []string{},
		SessionStart:    time.Now(),
		LastActivity:    time.Now(),
		UpdatedAt:       time.Now(),
	}

	storage.SaveLearningFSMState(state)

	err = storage.ValidateLearningFSMState()
	if err == nil {
		t.Fatal("expected error for invalid strategy")
	}
	expected := "invalid learning strategy:"
	if len(err.Error()) < len(expected) || err.Error()[:len(expected)] != expected {
		t.Fatalf("expected invalid strategy error, got: %v", err)
	}
}

func TestValidateAllMemoryFSMStates(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	validState := &MemoryFSMState{
		URN:   "v2e::card::1",
		State: MemoryStateNew,
		StateHistory: []StateHistory{
			{
				FromState: MemoryStateNew,
				ToState:   MemoryStateNew,
				Timestamp: time.Now(),
				Reason:    "initial",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	storage.SaveMemoryFSMState("v2e::card::1", validState)

	invalidState := &MemoryFSMState{
		URN:   "v2e::card::2",
		State: MemoryStateNew,
		StateHistory: []StateHistory{
			{
				FromState: MemoryStateNew,
				ToState:   MemoryStateNew,
				Timestamp: time.Now(),
				Reason:    "initial",
			},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Now(),
	}
	storage.SaveMemoryFSMState("v2e::card::2", invalidState)

	errors, err := storage.ValidateAllMemoryFSMStates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	if _, ok := errors["v2e::card::2"]; !ok {
		t.Fatal("expected error for v2e::card::2")
	}
}

func TestBoltDBStorage_NewAndClose(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if storage.db == nil {
		t.Fatal("expected db to be initialized")
	}

	if storage.closed {
		t.Fatal("expected closed to be false")
	}

	// Close should work
	err = storage.Close()
	if err != nil {
		t.Fatal(err)
	}

	if !storage.closed {
		t.Error("expected closed to be true after Close")
	}

	// Double close should be safe
	err = storage.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBoltDBStorage_SaveAndLoadMemoryFSMState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &MemoryFSMState{
		URN:   "v2e::card::test",
		State: MemoryStateLearned,
		StateHistory: []StateHistory{
			{FromState: MemoryStateNew, ToState: MemoryStateLearned, Timestamp: time.Now(), Reason: "test"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = storage.SaveMemoryFSMState("v2e::card::test", state)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := storage.LoadMemoryFSMState("v2e::card::test")
	if err != nil {
		t.Fatal(err)
	}

	if loaded.URN != state.URN {
		t.Errorf("expected URN %s, got %s", state.URN, loaded.URN)
	}

	if loaded.State != state.State {
		t.Errorf("expected state %s, got %s", state.State, loaded.State)
	}
}

func TestBoltDBStorage_LoadNonExistentState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	_, err = storage.LoadMemoryFSMState("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent state")
	}
}

func TestBoltDBStorage_DeleteMemoryFSMState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &MemoryFSMState{
		URN:       "v2e::card::test",
		State:     MemoryStateNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = storage.SaveMemoryFSMState("v2e::card::test", state)
	if err != nil {
		t.Fatal(err)
	}

	// Verify it exists
	_, err = storage.LoadMemoryFSMState("v2e::card::test")
	if err != nil {
		t.Fatal(err)
	}

	// Delete it
	err = storage.DeleteMemoryFSMState("v2e::card::test")
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's gone
	_, err = storage.LoadMemoryFSMState("v2e::card::test")
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestBoltDBStorage_SaveAndLoadLearningFSMState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &LearningFSMState{
		State:           LearningStateBrowsing,
		CurrentStrategy: "bfs",
		CurrentItemURN:  "v2e::cve::1",
		ViewedItems:     []string{"v2e::cve::1"},
		CompletedItems:  []string{},
		PathStack:       []string{},
		SessionStart:    time.Now(),
		LastActivity:    time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = storage.SaveLearningFSMState(state)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := storage.LoadLearningFSMState()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.State != state.State {
		t.Errorf("expected state %s, got %s", state.State, loaded.State)
	}

	if loaded.CurrentStrategy != state.CurrentStrategy {
		t.Errorf("expected strategy %s, got %s", state.CurrentStrategy, loaded.CurrentStrategy)
	}
}

func TestBoltDBStorage_LoadNonExistentLearningState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	_, err = storage.LoadLearningFSMState()
	if err == nil {
		t.Fatal("expected error for non-existent learning state")
	}
}

func TestBoltDBStorage_ClearLearningFSMState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &LearningFSMState{
		State:           LearningStateBrowsing,
		CurrentStrategy: "bfs",
		SessionStart:    time.Now(),
		LastActivity:    time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = storage.SaveLearningFSMState(state)
	if err != nil {
		t.Fatal(err)
	}

	// Verify it exists
	_, err = storage.LoadLearningFSMState()
	if err != nil {
		t.Fatal(err)
	}

	// Clear it
	err = storage.ClearLearningFSMState()
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's gone
	_, err = storage.LoadLearningFSMState()
	if err == nil {
		t.Fatal("expected error after clearing")
	}
}

func TestBoltDBStorage_Backup(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	backupPath := tmpFile.Name() + ".backup"
	defer os.Remove(backupPath)

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	// Save some state
	state := &MemoryFSMState{
		URN:       "v2e::card::test",
		State:     MemoryStateNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	storage.SaveMemoryFSMState("v2e::card::test", state)

	// Create backup
	err = storage.Backup(backupPath)
	if err != nil {
		t.Fatal(err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("backup file should exist")
	}
}

func TestBoltDBStorage_GetAllMemoryFSMStates(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	// Save multiple states
	urns := []string{"v2e::card::1", "v2e::card::2", "v2e::card::3"}
	for i, urn := range urns {
		state := &MemoryFSMState{
			URN:       urn,
			State:     MemoryStateNew,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		storage.SaveMemoryFSMState(urn, state)

		if i == 1 {
			// Leave middle one without CreatedAt for validation test
			state.CreatedAt = time.Time{}
			storage.SaveMemoryFSMState(urns[1], state)
		}
	}

	// Get all states
	allStates, err := storage.GetAllMemoryFSMStates()
	if err != nil {
		t.Fatal(err)
	}

	if len(allStates) != len(urns) {
		t.Errorf("expected %d states, got %d", len(urns), len(allStates))
	}

	// Verify URNs match
	for _, urn := range urns {
		if _, ok := allStates[urn]; !ok {
			t.Errorf("expected to find state for %s", urn)
		}
	}
}

func TestBoltDBStorage_OverwriteState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	// Save initial state
	state1 := &MemoryFSMState{
		URN:       "v2e::card::test",
		State:     MemoryStateNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = storage.SaveMemoryFSMState("v2e::card::test", state1)
	if err != nil {
		t.Fatal(err)
	}

	// Overwrite with new state
	time.Sleep(10 * time.Millisecond)
	state2 := &MemoryFSMState{
		URN:       "v2e::card::test",
		State:     MemoryStateLearned,
		CreatedAt: state1.CreatedAt,
		UpdatedAt: time.Now(),
	}

	err = storage.SaveMemoryFSMState("v2e::card::test", state2)
	if err != nil {
		t.Fatal(err)
	}

	// Load and verify it's the new state
	loaded, err := storage.LoadMemoryFSMState("v2e::card::test")
	if err != nil {
		t.Fatal(err)
	}

	if loaded.State != MemoryStateLearned {
		t.Errorf("expected state learned, got %s", loaded.State)
	}

	if !loaded.UpdatedAt.After(state1.UpdatedAt) {
		t.Error("expected updated_at to be newer")
	}
}

func TestBoltDBStorage_EmptyHistory(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	state := &MemoryFSMState{
		URN:          "v2e::card::test",
		State:        MemoryStateNew,
		StateHistory: []StateHistory{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = storage.SaveMemoryFSMState("v2e::card::test", state)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := storage.LoadMemoryFSMState("v2e::card::test")
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.StateHistory) != 0 {
		t.Errorf("expected empty history, got %d entries", len(loaded.StateHistory))
	}
}

func TestBoltDBStorage_MultipleTransitions(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	urn := "v2e::card::test"

	// Save multiple state transitions
	transitions := []MemoryState{
		MemoryStateNew,
		MemoryStateLearning,
		MemoryStateLearned,
		MemoryStateArchived,
	}

	for i, newState := range transitions {
		state := &MemoryFSMState{
			URN:   urn,
			State: newState,
			StateHistory: []StateHistory{
				{
					FromState: MemoryStateNew,
					ToState:   newState,
					Timestamp: time.Now(),
					Reason:    fmt.Sprintf("transition %d", i),
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = storage.SaveMemoryFSMState(urn, state)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Load final state
	loaded, err := storage.LoadMemoryFSMState(urn)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.State != MemoryStateArchived {
		t.Errorf("expected final state archived, got %s", loaded.State)
	}
}

func TestBoltDBStorage_ThreadSafe(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(idx int) {
			urn := fmt.Sprintf("v2e::card::%d", idx)
			state := &MemoryFSMState{
				URN:       urn,
				State:     MemoryStateNew,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			storage.SaveMemoryFSMState(urn, state)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(idx int) {
			urn := fmt.Sprintf("v2e::card::%d", idx)
			storage.LoadMemoryFSMState(urn)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestBoltDBStorage_LongURN(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	// Create a very long URN
	longURN := "v2e::card::"
	for i := 0; i < 100; i++ {
		longURN += fmt.Sprintf("part-%d-", i)
	}

	state := &MemoryFSMState{
		URN:       longURN,
		State:     MemoryStateNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = storage.SaveMemoryFSMState(longURN, state)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := storage.LoadMemoryFSMState(longURN)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.URN != longURN {
		t.Error("URN mismatch")
	}
}

func TestBoltDBStorage_StateHistoryOrder(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	urn := "v2e::card::test"

	// Create state with ordered history
	baseTime := time.Now()
	history := make([]StateHistory, 5)
	for i := 0; i < 5; i++ {
		history[i] = StateHistory{
			FromState: MemoryStateNew,
			ToState:   MemoryStateLearned,
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			Reason:    fmt.Sprintf("step %d", i),
		}
	}

	state := &MemoryFSMState{
		URN:          urn,
		State:        MemoryStateLearned,
		StateHistory: history,
		CreatedAt:    baseTime,
		UpdatedAt:    baseTime.Add(5 * time.Second),
	}

	err = storage.SaveMemoryFSMState(urn, state)
	if err != nil {
		t.Fatal(err)
	}

	// Validate history is ordered
	err = storage.ValidateMemoryFSMState(urn)
	if err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

// ============================================================================
// BoltDB Storage Failure Scenario Tests
// ============================================================================

func TestBoltDBStorage_CorruptDatabase(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-corrupt-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	err = os.WriteFile(tmpFile.Name(), []byte("corrupt data"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewBoltDBStorage(tmpFile.Name())
	if err == nil {
		t.Error("expected error when opening corrupt database")
	}
}

func TestBoltDBStorage_ConcurrentAccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fsm-concurrent-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewBoltDBStorage(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	urns := []string{
		"v2e::card::1",
		"v2e::card::2",
		"v2e::card::3",
	}

	errChan := make(chan error, len(urns))

	for _, urn := range urns {
		go func(u string) {
			state := &MemoryFSMState{
				URN:   u,
				State: MemoryStateNew,
				StateHistory: []StateHistory{
					{
						FromState: MemoryStateNew,
						ToState:   MemoryStateNew,
						Timestamp: time.Now(),
						Reason:    "test",
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			errChan <- storage.SaveMemoryFSMState(u, state)
		}(urn)
	}

	for i := 0; i < len(urns); i++ {
		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("concurrent write failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent operation timeout")
		}
	}
}
