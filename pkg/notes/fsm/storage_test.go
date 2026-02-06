package fsm

import (
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
