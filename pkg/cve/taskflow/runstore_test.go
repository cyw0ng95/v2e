package taskflow

import (
	"testing"
)

func TestRunStore_UpdateState_ValidAndInvalid(t *testing.T) {
	rs := NewTempRunStore(t)

	runID := "run-1"
	_, err := rs.CreateRun(runID, 0, 10, DataTypeCVE)
	if err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}

	// queued -> running (valid)
	if err := rs.UpdateState(runID, StateRunning); err != nil {
		t.Fatalf("expected queued->running to succeed: %v", err)
	}

	r, err := rs.GetRun(runID)
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}
	if r.State != StateRunning {
		t.Fatalf("expected state running, got %s", r.State)
	}

	// running -> queued (invalid)
	if err := rs.UpdateState(runID, StateQueued); err == nil {
		t.Fatalf("expected running->queued to fail")
	}
}

func TestRunStore_SetError_SetsFailedStateAndMessage(t *testing.T) {
	rs := NewTempRunStore(t)

	runID := "run-err"
	_, err := rs.CreateRun(runID, 0, 5, DataTypeCVE)
	if err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}

	if err := rs.SetError(runID, "something bad"); err != nil {
		t.Fatalf("SetError failed: %v", err)
	}

	r, err := rs.GetRun(runID)
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}

	if r.State != StateFailed {
		t.Fatalf("expected state failed, got %s", r.State)
	}
	if r.ErrorMessage != "something bad" {
		t.Fatalf("unexpected error message: %s", r.ErrorMessage)
	}
}

func TestRunStore_StoppedFromQueuedAndPaused(t *testing.T) {
	rs := NewTempRunStore(t)

	// queued -> stopped
	run1 := "run-stop-1"
	if _, err := rs.CreateRun(run1, 0, 1, DataTypeCVE); err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}
	if err := rs.UpdateState(run1, StateStopped); err != nil {
		t.Fatalf("queued->stopped failed: %v", err)
	}

	// queued -> running -> paused -> stopped
	run2 := "run-stop-2"
	if _, err := rs.CreateRun(run2, 0, 1, DataTypeCVE); err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}
	if err := rs.UpdateState(run2, StateRunning); err != nil {
		t.Fatalf("queued->running failed: %v", err)
	}
	if err := rs.UpdateState(run2, StatePaused); err != nil {
		t.Fatalf("running->paused failed: %v", err)
	}
	if err := rs.UpdateState(run2, StateStopped); err != nil {
		t.Fatalf("paused->stopped failed: %v", err)
	}
}
