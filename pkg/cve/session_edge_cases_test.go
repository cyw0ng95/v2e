package cve

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
)

// TestSessionManager_CreateSession tests creating sessions with various parameters
func TestSessionManager_CreateSession(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Test creating a valid session
	sessionID := "test-session-1"
	sess, err := manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if sess.ID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, sess.ID)
	}
	if sess.State != session.StateIdle {
		t.Errorf("Expected session state %s, got %s", session.StateIdle, sess.State)
	}
	if sess.StartIndex != 0 {
		t.Errorf("Expected start index 0, got %d", sess.StartIndex)
	}
	if sess.ResultsPerBatch != 100 {
		t.Errorf("Expected results per batch 100, got %d", sess.ResultsPerBatch)
	}

	// Test creating another session (should fail)
	_, err = manager.CreateSession("another-session", 0, 100)
	if err != session.ErrSessionExists {
		t.Errorf("Expected ErrSessionExists, got %v", err)
	}

	// Test creating session with different parameters
	manager2, err := session.NewManager(filepath.Join(tmpDir, "session_test2.db"), logger)
	if err != nil {
		t.Fatalf("Failed to create second session manager: %v", err)
	}
	defer manager2.Close()

	sess2, err := manager2.CreateSession("different-session", 50, 200)
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}

	if sess2.StartIndex != 50 {
		t.Errorf("Expected start index 50, got %d", sess2.StartIndex)
	}
	if sess2.ResultsPerBatch != 200 {
		t.Errorf("Expected results per batch 200, got %d", sess2.ResultsPerBatch)
	}
}

// TestSessionManager_GetSession tests retrieving sessions
func TestSessionManager_GetSession(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_get_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_GET_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Test getting non-existent session
	_, err = manager.GetSession()
	if err != session.ErrNoSession {
		t.Errorf("Expected ErrNoSession, got %v", err)
	}

	// Create a session
	sessionID := "get-test-session"
	createdSess, err := manager.CreateSession(sessionID, 10, 50)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get the session
	retrievedSess, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSess.ID != createdSess.ID {
		t.Errorf("Session ID mismatch: expected %s, got %s", createdSess.ID, retrievedSess.ID)
	}
	if retrievedSess.StartIndex != createdSess.StartIndex {
		t.Errorf("Start index mismatch: expected %d, got %d", createdSess.StartIndex, retrievedSess.StartIndex)
	}
	if retrievedSess.ResultsPerBatch != createdSess.ResultsPerBatch {
		t.Errorf("Results per batch mismatch: expected %d, got %d", createdSess.ResultsPerBatch, retrievedSess.ResultsPerBatch)
	}
}

// TestSessionManager_UpdateState tests updating session state
func TestSessionManager_UpdateState(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_state_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_STATE_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	sessionID := "state-test-session"
	_, err = manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test updating to running state
	err = manager.UpdateState(session.StateRunning)
	if err != nil {
		t.Fatalf("Failed to update state to running: %v", err)
	}

	// Verify the state was updated
	sess, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if sess.State != session.StateRunning {
		t.Errorf("Expected state %s, got %s", session.StateRunning, sess.State)
	}

	// Test updating to paused state
	err = manager.UpdateState(session.StatePaused)
	if err != nil {
		t.Fatalf("Failed to update state to paused: %v", err)
	}

	// Verify the state was updated
	sess, err = manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if sess.State != session.StatePaused {
		t.Errorf("Expected state %s, got %s", session.StatePaused, sess.State)
	}

	// Test updating to stopped state
	err = manager.UpdateState(session.StateStopped)
	if err != nil {
		t.Fatalf("Failed to update state to stopped: %v", err)
	}

	// Verify the state was updated
	sess, err = manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if sess.State != session.StateStopped {
		t.Errorf("Expected state %s, got %s", session.StateStopped, sess.State)
	}
}

// TestSessionManager_UpdateProgress tests updating session progress
func TestSessionManager_UpdateProgress(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_progress_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_PROGRESS_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	sessionID := "progress-test-session"
	_, err = manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test initial progress values
	sess, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if sess.FetchedCount != 0 {
		t.Errorf("Expected fetched count 0, got %d", sess.FetchedCount)
	}
	if sess.StoredCount != 0 {
		t.Errorf("Expected stored count 0, got %d", sess.StoredCount)
	}
	if sess.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", sess.ErrorCount)
	}

	// Update progress with positive values
	err = manager.UpdateProgress(10, 8, 2)
	if err != nil {
		t.Fatalf("Failed to update progress: %v", err)
	}

	// Verify progress was updated
	sess, err = manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if sess.FetchedCount != 10 {
		t.Errorf("Expected fetched count 10, got %d", sess.FetchedCount)
	}
	if sess.StoredCount != 8 {
		t.Errorf("Expected stored count 8, got %d", sess.StoredCount)
	}
	if sess.ErrorCount != 2 {
		t.Errorf("Expected error count 2, got %d", sess.ErrorCount)
	}

	// Update progress again to test cumulative updates
	err = manager.UpdateProgress(5, 4, 1)
	if err != nil {
		t.Fatalf("Failed to update progress again: %v", err)
	}

	// Verify cumulative progress
	sess, err = manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if sess.FetchedCount != 15 { // 10 + 5
		t.Errorf("Expected fetched count 15, got %d", sess.FetchedCount)
	}
	if sess.StoredCount != 12 { // 8 + 4
		t.Errorf("Expected stored count 12, got %d", sess.StoredCount)
	}
	if sess.ErrorCount != 3 { // 2 + 1
		t.Errorf("Expected error count 3, got %d", sess.ErrorCount)
	}
}

// TestSessionManager_DeleteSession tests deleting sessions
func TestSessionManager_DeleteSession(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_delete_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_DELETE_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Try to delete non-existent session
	err = manager.DeleteSession()
	if err != session.ErrNoSession {
		t.Errorf("Expected ErrNoSession when deleting non-existent session, got %v", err)
	}

	// Create a session
	sessionID := "delete-test-session"
	_, err = manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify session exists
	_, err = manager.GetSession()
	if err != nil {
		t.Fatalf("Session should exist after creation: %v", err)
	}

	// Delete the session
	err = manager.DeleteSession()
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session no longer exists
	_, err = manager.GetSession()
	if err != session.ErrNoSession {
		t.Errorf("Expected ErrNoSession after deletion, got %v", err)
	}
}

// TestSessionManager_ConcurrentAccess tests concurrent access to session manager
func TestSessionManager_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_concurrent_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_CONCURRENT_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Create a session first
	_, err = manager.CreateSession("concurrent-test-session", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	const numGoroutines = 20
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*3) // For multiple operations

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Perform multiple operations in each goroutine
			for j := 0; j < 10; j++ {
				// Get session
				_, err := manager.GetSession()
				if err != nil && err != session.ErrNoSession {
					errChan <- err
					return
				}

				// Update state
				states := []session.SessionState{session.StateRunning, session.StatePaused, session.StateStopped}
				state := states[(id+j)%len(states)]
				err = manager.UpdateState(state)
				if err != nil {
					errChan <- err
					return
				}

				// Update progress
				err = manager.UpdateProgress(1, 1, 0)
				if err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		if err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}

	// Final verification
	finalSess, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get final session: %v", err)
	}

	if finalSess.FetchedCount != numGoroutines*10 {
		t.Errorf("Expected fetched count %d, got %d", numGoroutines*10, finalSess.FetchedCount)
	}
}

// TestSessionManager_CacheBehavior tests the caching behavior of the session manager
func TestSessionManager_CacheBehavior(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_cache_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_CACHE_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	sessionID := "cache-test-session"
	_, err = manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get session multiple times - should hit cache
	for i := 0; i < 5; i++ {
		sess, err := manager.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session %d: %v", i, err)
		}
		if sess.ID != sessionID {
			t.Errorf("Session ID mismatch on iteration %d", i)
		}
	}

	// Update state and verify it's reflected in subsequent gets
	err = manager.UpdateState(session.StateRunning)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	sess, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}
	if sess.State != session.StateRunning {
		t.Errorf("Expected state %s after update, got %s", session.StateRunning, sess.State)
	}
}

// TestSessionManager_Close tests closing the session manager
func TestSessionManager_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_close_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_CLOSE_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Create a session
	_, err = manager.CreateSession("close-test-session", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Close the manager
	err = manager.Close()
	if err != nil {
		t.Errorf("Failed to close manager: %v", err)
	}

	// Attempting to use the closed manager should fail gracefully
	// Note: This depends on bbolt's behavior when database is closed
	_, err = manager.GetSession()
	// We don't assert on error type here as it depends on underlying db implementation
}

// TestSessionManager_SessionStates tests all session states
func TestSessionManager_SessionStates(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_states_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_STATES_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	sessionID := "states-test-session"
	_, err = manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	states := []session.SessionState{
		session.StateIdle,
		session.StateRunning,
		session.StatePaused,
		session.StateStopped,
	}

	for _, expectedState := range states {
		err = manager.UpdateState(expectedState)
		if err != nil {
			t.Fatalf("Failed to update to state %s: %v", expectedState, err)
		}

		sess, err := manager.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session in state %s: %v", expectedState, err)
		}

		if sess.State != expectedState {
			t.Errorf("Expected state %s, got %s", expectedState, sess.State)
		}
	}
}

// TestSessionManager_TimeStamps tests the CreatedAt and UpdatedAt timestamps
func TestSessionManager_TimeStamps(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "session_timestamp_test.db")

	logger := common.NewLogger(os.Stdout, "SESSION_TIMESTAMP_TEST", common.DebugLevel)
	manager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer manager.Close()

	// Record time before session creation
	beforeCreate := time.Now()

	// Create a session
	sessionID := "timestamp-test-session"
	sess, err := manager.CreateSession(sessionID, 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Record time after session creation
	afterCreate := time.Now()

	// Verify timestamps are within expected range
	if sess.CreatedAt.Before(beforeCreate) || sess.CreatedAt.After(afterCreate) {
		t.Errorf("CreatedAt timestamp is outside expected range: %v (should be between %v and %v)", 
			sess.CreatedAt, beforeCreate, afterCreate)
	}

	if sess.UpdatedAt.Before(beforeCreate) || sess.UpdatedAt.After(afterCreate) {
		t.Errorf("UpdatedAt timestamp is outside expected range: %v (should be between %v and %v)", 
			sess.UpdatedAt, beforeCreate, afterCreate)
	}

	// Update state and check that UpdatedAt changes
	time.Sleep(10 * time.Millisecond) // Ensure some time passes
	beforeUpdate := time.Now()

	err = manager.UpdateState(session.StateRunning)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	afterUpdate := time.Now()
	updatedSess, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	// CreatedAt should remain unchanged
	if !updatedSess.CreatedAt.Equal(sess.CreatedAt) {
		t.Errorf("CreatedAt changed after state update: was %v, now %v", sess.CreatedAt, updatedSess.CreatedAt)
	}

	// UpdatedAt should have changed
	if updatedSess.UpdatedAt.Before(beforeUpdate) || updatedSess.UpdatedAt.After(afterUpdate) {
		t.Errorf("UpdatedAt timestamp is outside expected range after update: %v (should be between %v and %v)", 
			updatedSess.UpdatedAt, beforeUpdate, afterUpdate)
	}
}