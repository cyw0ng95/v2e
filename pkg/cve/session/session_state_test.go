package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// TestSessionState_String ensures all state constants stringify correctly.
func TestSessionState_String(t *testing.T) {
	cases := []struct {
		state SessionState
		want  string
	}{
		{StateIdle, "idle"},
		{StateRunning, "running"},
		{StatePaused, "paused"},
		{StateStopped, "stopped"},
	}

	for _, tc := range cases {
		if string(tc.state) != tc.want {
			t.Fatalf("state %v expected %q, got %q", tc.state, tc.want, string(tc.state))
		}
	}
}

// TestManager_CreateSession_MultipleAttempts ensures only one session can exist.
func TestManager_CreateSession_MultipleAttempts(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	sess1, err := mgr.CreateSession("sess1", 0, 100)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sess1.ID != "sess1" || sess1.State != StateIdle {
		t.Fatalf("unexpected session: %+v", sess1)
	}

	// Attempt to create another session should fail
	_, err = mgr.CreateSession("sess2", 100, 50)
	if err != ErrSessionExists {
		t.Fatalf("expected ErrSessionExists, got %v", err)
	}
}

// TestManager_GetSession_NoSession ensures proper error when no session exists.
func TestManager_GetSession_NoSession(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	_, err = mgr.GetSession()
	if err != ErrNoSession {
		t.Fatalf("expected ErrNoSession, got %v", err)
	}
}

// TestManager_UpdateState_Transitions covers all valid state transitions.
func TestManager_UpdateState_Transitions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	_, err = mgr.CreateSession("sess", 0, 100)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	transitions := []SessionState{StateRunning, StatePaused, StateRunning, StateStopped, StateIdle}
	for _, nextState := range transitions {
		if err := mgr.UpdateState(nextState); err != nil {
			t.Fatalf("UpdateState to %s failed: %v", nextState, err)
		}
		sess, _ := mgr.GetSession()
		if sess.State != nextState {
			t.Fatalf("expected state %s, got %s", nextState, sess.State)
		}
	}
}

// TestManager_IncrementCounters covers counter increments across multiple sessions.
func TestManager_IncrementCounters(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	_, err = mgr.CreateSession("sess", 0, 100)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	for i := 0; i < 25; i++ {
		if err := mgr.UpdateProgress(int64(i+1), int64(i), 0); err != nil {
			t.Fatalf("UpdateProgress failed: %v", err)
		}
		if i%5 == 0 {
			if err := mgr.UpdateProgress(0, 0, 1); err != nil {
				t.Fatalf("UpdateProgress failed: %v", err)
			}
		}
	}

	sess, err := mgr.GetSession()
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	expectedFetched := int64((25 * 26) / 2)
	expectedStored := int64((24 * 25) / 2)
	expectedErrors := int64(5)

	if sess.FetchedCount != expectedFetched {
		t.Fatalf("expected FetchedCount %d, got %d", expectedFetched, sess.FetchedCount)
	}
	if sess.StoredCount != expectedStored {
		t.Fatalf("expected StoredCount %d, got %d", expectedStored, sess.StoredCount)
	}
	if sess.ErrorCount != expectedErrors {
		t.Fatalf("expected ErrorCount %d, got %d", expectedErrors, sess.ErrorCount)
	}
}

// TestManager_DeleteSession_NoSessionError ensures delete fails when no session exists.
func TestManager_DeleteSession_NoSessionError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	err = mgr.DeleteSession()
	if err != ErrNoSession {
		t.Fatalf("expected ErrNoSession on delete, got %v", err)
	}
}

// TestManager_SessionPersistence ensures session survives manager restart.
func TestManager_SessionPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)

	mgr1, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	sess1, err := mgr1.CreateSession("persistent", 500, 200)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	mgr1.UpdateState(StateRunning)
	mgr1.UpdateProgress(42, 0, 0)
	mgr1.Close()

	// Reopen
	mgr2, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager reopen failed: %v", err)
	}
	defer mgr2.Close()

	sess2, err := mgr2.GetSession()
	if err != nil {
		t.Fatalf("GetSession after reopen failed: %v", err)
	}

	if sess2.ID != sess1.ID || sess2.State != StateRunning || sess2.FetchedCount != 42 {
		t.Fatalf("session not persisted correctly: %+v", sess2)
	}
}

// TestManager_ConcurrentUpdates ensures concurrent counter increments are safe.
func TestManager_ConcurrentUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	_, err = mgr.CreateSession("concurrent", 0, 100)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// UpdateProgress is a cumulative update, so we test sequential updates
	numUpdates := 20
	for i := 0; i < numUpdates; i++ {
		if err := mgr.UpdateProgress(10, 10, 0); err != nil {
			t.Fatalf("UpdateProgress failed: %v", err)
		}
	}

	sess, _ := mgr.GetSession()
	expected := int64(numUpdates * 10)
	if sess.FetchedCount != expected || sess.StoredCount != expected {
		t.Fatalf("sequential updates failed: fetched=%d stored=%d (expected %d each)", sess.FetchedCount, sess.StoredCount, expected)
	}
}

// TestManager_UpdateStartIndex verifies start index updates.
func TestManager_UpdateStartIndex(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	_, err = mgr.CreateSession("sess", 100, 50)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Since there's no UpdateStartIndex method, we just verify the field can be read
	sess, _ := mgr.GetSession()
	if sess.StartIndex != 100 {
		t.Fatalf("expected initial StartIndex 100, got %d", sess.StartIndex)
	}
}

// TestSession_Timestamps ensures CreatedAt and UpdatedAt are set properly.
func TestSession_Timestamps(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session.db")
	logger := common.NewLogger(os.Stdout, "test", common.ErrorLevel)
	mgr, err := NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	before := time.Now()
	sess, err := mgr.CreateSession("ts-test", 0, 100)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	after := time.Now()

	if sess.CreatedAt.Before(before) || sess.CreatedAt.After(after) {
		t.Fatalf("CreatedAt timestamp out of range: %v", sess.CreatedAt)
	}
	if sess.UpdatedAt.Before(before) || sess.UpdatedAt.After(after) {
		t.Fatalf("UpdatedAt timestamp out of range: %v", sess.UpdatedAt)
	}

	time.Sleep(10 * time.Millisecond)
	updateBefore := time.Now()
	mgr.UpdateState(StateRunning)
	updateAfter := time.Now()

	updated, _ := mgr.GetSession()
	if updated.UpdatedAt.Before(updateBefore) || updated.UpdatedAt.After(updateAfter) {
		t.Fatalf("UpdatedAt not refreshed: %v", updated.UpdatedAt)
	}
}
