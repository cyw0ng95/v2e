package session

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// Create temporary database file
	dbPath := filepath.Join(t.TempDir(), "test_session.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	if manager.db == nil {
		t.Error("Database not initialized")
	}
}

func TestCreateSession(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_create_session.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	session, err := manager.CreateSession("test-session-1", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID != "test-session-1" {
		t.Errorf("Expected session ID 'test-session-1', got '%s'", session.ID)
	}

	if session.State != StateIdle {
		t.Errorf("Expected state 'idle', got '%s'", session.State)
	}

	if session.StartIndex != 0 {
		t.Errorf("Expected start index 0, got %d", session.StartIndex)
	}

	if session.ResultsPerBatch != 100 {
		t.Errorf("Expected results per batch 100, got %d", session.ResultsPerBatch)
	}
}

func TestSingleSessionEnforcement(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_single_session.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create first session
	_, err = manager.CreateSession("session-1", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create first session: %v", err)
	}

	// Try to create second session (should fail)
	_, err = manager.CreateSession("session-2", 0, 100)
	if err != ErrSessionExists {
		t.Errorf("Expected ErrSessionExists, got %v", err)
	}
}

func TestGetSession(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_get_session.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Try to get session when none exists
	_, err = manager.GetSession()
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got %v", err)
	}

	// Create a session
	created, err := manager.CreateSession("test-session", 0, 50)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get the session
	retrieved, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected session ID '%s', got '%s'", created.ID, retrieved.ID)
	}
}

func TestUpdateState(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_update_state.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	_, err = manager.CreateSession("test-session", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Update state to running
	err = manager.UpdateState(StateRunning)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	// Verify state
	session, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.State != StateRunning {
		t.Errorf("Expected state 'running', got '%s'", session.State)
	}
}

func TestUpdateProgress(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_update_progress.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	_, err = manager.CreateSession("test-session", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Update progress
	err = manager.UpdateProgress(10, 8, 2)
	if err != nil {
		t.Fatalf("Failed to update progress: %v", err)
	}

	// Verify progress
	session, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.FetchedCount != 10 {
		t.Errorf("Expected fetched count 10, got %d", session.FetchedCount)
	}

	if session.StoredCount != 8 {
		t.Errorf("Expected stored count 8, got %d", session.StoredCount)
	}

	if session.ErrorCount != 2 {
		t.Errorf("Expected error count 2, got %d", session.ErrorCount)
	}

	// Update progress again (should accumulate)
	err = manager.UpdateProgress(5, 5, 0)
	if err != nil {
		t.Fatalf("Failed to update progress second time: %v", err)
	}

	session, err = manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.FetchedCount != 15 {
		t.Errorf("Expected fetched count 15, got %d", session.FetchedCount)
	}

	if session.StoredCount != 13 {
		t.Errorf("Expected stored count 13, got %d", session.StoredCount)
	}

	if session.ErrorCount != 2 {
		t.Errorf("Expected error count 2, got %d", session.ErrorCount)
	}
}

func TestDeleteSession(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_delete_session.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	_, err = manager.CreateSession("test-session", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Delete the session
	err = manager.DeleteSession()
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session is deleted
	_, err = manager.GetSession()
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession after deletion, got %v", err)
	}
}

func TestSessionPersistence(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_session_persistence.db")

	// Create session in first manager
	manager1, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create first manager: %v", err)
	}

	_, err = manager1.CreateSession("persistent-session", 100, 200)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	err = manager1.UpdateState(StateRunning)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	manager1.Close()

	// Open database with second manager
	manager2, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}
	defer manager2.Close()

	// Verify session persists
	session, err := manager2.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session from second manager: %v", err)
	}

	if session.ID != "persistent-session" {
		t.Errorf("Expected session ID 'persistent-session', got '%s'", session.ID)
	}

	if session.State != StateRunning {
		t.Errorf("Expected state 'running', got '%s'", session.State)
	}

	if session.StartIndex != 100 {
		t.Errorf("Expected start index 100, got %d", session.StartIndex)
	}

	if session.ResultsPerBatch != 200 {
		t.Errorf("Expected results per batch 200, got %d", session.ResultsPerBatch)
	}
}

func TestUpdateTimestamps(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_update_timestamps.db")

	manager, err := NewManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Create a session
	session1, err := manager.CreateSession("test-session", 0, 100)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	originalUpdatedAt := session1.UpdatedAt

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Update state
	err = manager.UpdateState(StateRunning)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	// Get updated session
	session2, err := manager.GetSession()
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	// Verify UpdatedAt changed
	if !session2.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt timestamp should be updated")
	}
}
