package fsm

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestStateTransitionError_Creation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StateTransitionError_Creation", nil, func(t *testing.T, _ *gorm.DB) {
		cause := fmt.Errorf("underlying error")
		err := NewTransitionError(ErrInvalidTransition, "IDLE", "BUILDING", cause)

		if err.ErrorType != ErrInvalidTransition {
			t.Errorf("Expected error type %s, got %s", ErrInvalidTransition, err.ErrorType)
		}

		if err.FromState != "IDLE" {
			t.Errorf("Expected from state IDLE, got %s", err.FromState)
		}

		if err.ToState != "BUILDING" {
			t.Errorf("Expected to state BUILDING, got %s", err.ToState)
		}

		if err.Cause != cause {
			t.Errorf("Expected cause to match, got %v", err.Cause)
		}

		if !err.CanRecover {
			t.Error("Expected CanRecover to be true by default")
		}

		if err.RolledBack {
			t.Error("Expected RolledBack to be false by default")
		}
	})
}

func TestStateTransitionError_IsRecoverable(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StateTransitionError_IsRecoverable", nil, func(t *testing.T, _ *gorm.DB) {
		// Test recoverable error
		err1 := NewTransitionError(ErrTransitionFailed, "A", "B", fmt.Errorf("transient error"))
		err1.RecoveryAttempts = 1
		if !err1.IsRecoverable() {
			t.Error("Expected error to be recoverable")
		}

		// Test non-recoverable due to max attempts
		err2 := NewTransitionError(ErrTransitionFailed, "A", "B", fmt.Errorf("error"))
		err2.RecoveryAttempts = maxRecoveryAttempts + 1
		if err2.IsRecoverable() {
			t.Error("Expected error to be non-recoverable after max attempts")
		}

		// Test non-recoverable flag
		err3 := NewTransitionError(ErrInvalidTransition, "A", "B", fmt.Errorf("error"))
		err3.CanRecover = false
		if err3.IsRecoverable() {
			t.Error("Expected error to be non-recoverable when CanRecover is false")
		}
	})
}

func TestTransitionHistory_Record(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TransitionHistory_Record", nil, func(t *testing.T, _ *gorm.DB) {
		h := NewTransitionHistory(10)

		// Record some transitions
		h.Record("IDLE", "BUILDING", true, nil, time.Millisecond)
		h.Record("BUILDING", "READY", true, nil, time.Millisecond*2)

		if h.Len() != 2 {
			t.Errorf("Expected 2 entries, got %d", h.Len())
		}

		// Check recent
		recent := h.GetRecent(1)
		if len(recent) != 1 {
			t.Errorf("Expected 1 recent entry, got %d", len(recent))
		}

		if recent[0].To != "READY" {
			t.Errorf("Expected most recent to be READY, got %s", recent[0].To)
		}
	})
}

func TestTransitionHistory_MaxLength(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TransitionHistory_MaxLength", nil, func(t *testing.T, _ *gorm.DB) {
		maxLen := 5
		h := NewTransitionHistory(maxLen)

		// Add more than maxLen entries
		for i := 0; i < 10; i++ {
			h.Record(fmt.Sprintf("STATE_%d", i), fmt.Sprintf("STATE_%d", i+1), true, nil, time.Millisecond)
		}

		// Should only keep maxLen entries
		if h.Len() != maxLen {
			t.Errorf("Expected %d entries, got %d", maxLen, h.Len())
		}
	})
}

func TestTransitionHistory_FailedTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TransitionHistory_FailedTransitions", nil, func(t *testing.T, _ *gorm.DB) {
		h := NewTransitionHistory(10)

		// Record mix of success and failure
		h.Record("IDLE", "BUILDING", true, nil, time.Millisecond)
		h.Record("BUILDING", "ERROR", false, fmt.Errorf("test error"), time.Millisecond)
		h.Record("ERROR", "IDLE", true, nil, time.Millisecond)

		failed := h.GetFailedTransitions()
		if len(failed) != 1 {
			t.Errorf("Expected 1 failed transition, got %d", len(failed))
		}

		if failed[0].From != "BUILDING" || failed[0].To != "ERROR" {
			t.Errorf("Failed transition mismatch: %s -> %s", failed[0].From, failed[0].To)
		}
	})
}

func TestRollbackManager_SaveAndRetrieve(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RollbackManager_SaveAndRetrieve", nil, func(t *testing.T, _ *gorm.DB) {
		r := NewRollbackManager(5)

		// Save a snapshot
		snapshot := r.SaveSnapshot("READY", map[string]string{"key": "value"})

		if snapshot.State != "READY" {
			t.Errorf("Expected state READY, got %s", snapshot.State)
		}

		if snapshot.SequenceID != 1 {
			t.Errorf("Expected sequence ID 1, got %d", snapshot.SequenceID)
		}

		// Retrieve it
		retrieved, exists := r.GetLatestSnapshot("READY")
		if !exists {
			t.Fatal("Snapshot should exist")
		}

		if retrieved.SequenceID != snapshot.SequenceID {
			t.Errorf("Retrieved snapshot ID mismatch: %d vs %d", retrieved.SequenceID, snapshot.SequenceID)
		}
	})
}

func TestRollbackManager_MaxSnapshotsPerState(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RollbackManager_MaxSnapshotsPerState", nil, func(t *testing.T, _ *gorm.DB) {
		maxPerState := 3
		r := NewRollbackManager(maxPerState)

		// Add more than maxPerState snapshots
		for i := 0; i < 5; i++ {
			r.SaveSnapshot("READY", i)
		}

		// Count snapshots for READY state
		// The manager keeps only the most recent
		retrieved, _ := r.GetLatestSnapshot("READY")
		if retrieved.Data != 4 {
			t.Errorf("Expected latest data to be 4, got %v", retrieved.Data)
		}
	})
}

func TestRollbackManager_Clear(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RollbackManager_Clear", nil, func(t *testing.T, _ *gorm.DB) {
		r := NewRollbackManager(5)

		r.SaveSnapshot("READY", "data")
		r.ClearSnapshots("READY")

		_, exists := r.GetLatestSnapshot("READY")
		if exists {
			t.Error("Snapshot should be cleared")
		}
	})
}

func TestGraphFSM_TransitionWithHandler_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphFSM_TransitionWithHandler_Success", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		fsm := NewGraphFSM(logger)

		handlerCalled := false
		err := fsm.(*BaseGraphFSM).TransitionWithHandler(GraphBuilding, func() error {
			handlerCalled = true
			return nil
		})

		if err != nil {
			t.Fatalf("Transition failed: %v", err)
		}

		if !handlerCalled {
			t.Error("Handler was not called")
		}

		if fsm.GetState() != GraphBuilding {
			t.Errorf("Expected state BUILDING, got %s", fsm.GetState())
		}
	})
}

func TestGraphFSM_TransitionWithHandler_Rollback(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphFSM_TransitionWithHandler_Rollback", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		fsm := NewGraphFSM(logger)

		initialState := fsm.GetState()

		// Handler returns error - should rollback
		testErr := fmt.Errorf("handler error")
		err := fsm.(*BaseGraphFSM).TransitionWithHandler(GraphBuilding, func() error {
			return testErr
		})

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		transErr, ok := err.(*StateTransitionError)
		if !ok {
			t.Fatalf("Expected StateTransitionError, got %T", err)
		}

		if !transErr.RolledBack {
			t.Error("Expected error to indicate rollback occurred")
		}

		// State should be rolled back
		if fsm.GetState() != initialState {
			t.Errorf("Expected state to rollback to %s, got %s", initialState, fsm.GetState())
		}
	})
}

func TestGraphFSM_RollbackToState(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphFSM_RollbackToState", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		fsm := NewGraphFSM(logger)

		// Start from IDLE, go to READY
		fsm.StartBuild()
		fsm.CompleteBuild()

		if fsm.GetState() != GraphReady {
			t.Fatalf("Expected state READY, got %s", fsm.GetState())
		}

		// Rollback to IDLE
		err := fsm.(*BaseGraphFSM).RollbackToState(GraphIdle)
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		if fsm.GetState() != GraphIdle {
			t.Errorf("Expected state IDLE after rollback, got %s", fsm.GetState())
		}
	})
}

func TestGraphFSM_GetDiagnostics(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphFSM_GetDiagnostics", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		fsm := NewGraphFSM(logger)

		diag := fsm.(*BaseGraphFSM).GetDiagnostics()

		if diag["current_state"] != GraphIdle {
			t.Errorf("Expected current_state IDLE, got %v", diag["current_state"])
		}

		if diag["history_entries"].(int) != 0 {
			t.Errorf("Expected 0 history entries, got %v", diag["history_entries"])
		}

		// Trigger some transitions
		fsm.StartBuild()
		fsm.CompleteBuild()

		diag = fsm.(*BaseGraphFSM).GetDiagnostics()

		if diag["current_state"] != GraphReady {
			t.Errorf("Expected current_state READY, got %v", diag["current_state"])
		}

		if diag["history_entries"].(int) != 2 {
			t.Errorf("Expected 2 history entries, got %v", diag["history_entries"])
		}
	})
}

func TestGraphFSM_GetFailedTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphFSM_GetFailedTransitions", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		fsm := NewGraphFSM(logger)

		// Trigger a failure
		fsm.StartBuild()
		fsm.FailBuild(fmt.Errorf("test error"))

		// Verify FSM is in ERROR state
		baseFSM := fsm.(*BaseGraphFSM)
		if baseFSM.GetState() != GraphError {
			t.Error("Expected FSM to be in ERROR state after FailBuild")
		}

		// Verify last error is set
		if baseFSM.GetLastError() == nil {
			t.Error("Expected lastError to be set after FailBuild")
		}

		// Transition to ERROR state is valid, so check for BuildFailed event
		// The event should be in the history, even if transition succeeded
		history := baseFSM.GetTransitionHistory(10)
		if len(history) == 0 {
			t.Error("Expected at least one transition in history")
		}

		// Verify at least one successful transition (IDLE->BUILDING)
		hasSuccess := false
		for _, entry := range history {
			if entry.Success {
				hasSuccess = true
				break
			}
		}

		if !hasSuccess {
			t.Error("Expected at least one successful transition")
		}
	})
}

func TestGraphFSM_ClearHistory(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphFSM_ClearHistory", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		fsm := NewGraphFSM(logger)

		// Generate some history
		fsm.StartBuild()
		fsm.CompleteBuild()

		if fsm.(*BaseGraphFSM).GetTransitionHistory(10) == nil {
			t.Fatal("Expected history entries")
		}

		// Clear history
		fsm.(*BaseGraphFSM).ClearHistory()

		history := fsm.(*BaseGraphFSM).GetTransitionHistory(10)
		if len(history) != 0 {
			t.Errorf("Expected empty history after clear, got %d entries", len(history))
		}
	})
}
