package taskflow

import "testing"

func TestRunStore_GetActiveRun_None(t *testing.T) {
	rs := NewTempRunStore(t)
	active, err := rs.GetActiveRun()
	if err != nil {
		t.Fatalf("GetActiveRun returned error: %v", err)
	}
	if active != nil {
		t.Fatalf("expected no active run, got %v", active)
	}
}

func TestRunStore_UpdateProgress_Increments(t *testing.T) {
	rs := NewTempRunStore(t)
	_, _ = rs.CreateRun("r1", 0, 1, DataTypeCVE)
	if err := rs.UpdateProgress("r1", 2, 3, 4); err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}
	run, _ := rs.GetRun("r1")
	if run.FetchedCount != 2 || run.StoredCount != 3 || run.ErrorCount != 4 {
		t.Fatalf("unexpected counters: %+v", run)
	}
}

func TestRunStore_DeleteRun_RemovesEntry(t *testing.T) {
	rs := NewTempRunStore(t)
	_, _ = rs.CreateRun("r2", 0, 1, DataTypeCVE)
	if err := rs.DeleteRun("r2"); err != nil {
		t.Fatalf("DeleteRun failed: %v", err)
	}
	if _, err := rs.GetRun("r2"); err == nil {
		t.Fatalf("expected error retrieving deleted run")
	}
}

func TestRunStore_GetLatestRun_PicksMostRecent(t *testing.T) {
	rs := NewTempRunStore(t)
	_, _ = rs.CreateRun("old", 0, 1, DataTypeCVE)
	// update state to bump UpdatedAt
	_ = rs.UpdateState("old", StateRunning)
	_, _ = rs.CreateRun("new", 0, 1, DataTypeCVE)
	latest, err := rs.GetLatestRun()
	if err != nil {
		t.Fatalf("GetLatestRun failed: %v", err)
	}
	if latest.ID != "new" {
		t.Fatalf("expected latest run 'new', got %s", latest.ID)
	}
}
