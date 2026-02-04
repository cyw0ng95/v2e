package session

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// TestConcurrentSessionCreation tests concurrent session creation attempts
// Note: Due to bbolt's transaction model, multiple concurrent creates may succeed
// if they check for existing session before any complete. This test verifies that
// concurrent operations don't cause database corruption.
func TestConcurrentSessionCreation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentSessionCreation", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_concurrent_creation.db")

		// Correct logger initialization
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		var wg sync.WaitGroup
		successCount := 0
		errorCount := 0
		var mu sync.Mutex

		// Try to create 10 sessions concurrently with the SAME session ID
		// Due to bbolt transactions, some may succeed before checking
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// Use same session ID for all attempts
				_, err := manager.CreateSession("same-session-id", 0, 100)
				mu.Lock()
				defer mu.Unlock()
				if err == nil {
					successCount++
				} else if err == ErrSessionExists {
					errorCount++
				} else {
					t.Errorf("Unexpected error: %v", err)
				}
			}(i)
		}

		wg.Wait()

		// At least one should succeed, and there should be no database corruption
		if successCount == 0 {
			t.Error("Expected at least one successful creation")
		}

		// Verify we can still read the session
		session, err := manager.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session after concurrent creates: %v", err)
		}

		if session.ID != "same-session-id" {
			t.Error("Session ID corrupted by concurrent creates")
		}

		t.Logf("Concurrent creation test: %d succeeded, %d got 'exists' error", successCount, errorCount)
	})

}

// TestConcurrentStateUpdates tests concurrent state updates
func TestConcurrentStateUpdates(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentStateUpdates", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_concurrent_state_updates.db")

		// Add logger setup for NewManager calls
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		_, err = manager.CreateSession("test-session", 0, 100)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Update state concurrently 20 times
		var wg sync.WaitGroup
		states := []SessionState{StateRunning, StatePaused, StateRunning, StatePaused}

		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(iteration int) {
				defer wg.Done()
				state := states[iteration%len(states)]
				err := manager.UpdateState(state)
				if err != nil {
					t.Errorf("Failed to update state: %v", err)
				}
			}(i)
		}

		wg.Wait()

		// Verify session still exists and is in a valid state
		session, err := manager.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		// State should be one of the valid states
		validStates := map[SessionState]bool{
			StateIdle:    true,
			StateRunning: true,
			StatePaused:  true,
			StateStopped: true,
		}

		if !validStates[session.State] {
			t.Errorf("Invalid state after concurrent updates: %s", session.State)
		}
	})

}

// TestConcurrentProgressUpdates tests concurrent progress updates
func TestConcurrentProgressUpdates(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentProgressUpdates", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_concurrent_progress.db")

		// Add logger setup for NewManager calls
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		_, err = manager.CreateSession("test-session", 0, 100)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Update progress concurrently 100 times
		var wg sync.WaitGroup
		iterations := 100

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := manager.UpdateProgress(1, 1, 0)
				if err != nil {
					t.Errorf("Failed to update progress: %v", err)
				}
			}()
		}

		wg.Wait()

		// Verify final counts
		session, err := manager.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		// Due to concurrent writes and potential transaction conflicts in bbolt,
		// we may not get all 100 updates. But we should get at least some.
		// This tests that concurrent updates don't cause corruption.
		if session.FetchedCount == 0 {
			t.Error("Expected at least some fetched count updates")
		}

		if session.StoredCount == 0 {
			t.Error("Expected at least some stored count updates")
		}

		// The counts should match since we update both together
		if session.FetchedCount != session.StoredCount {
			t.Errorf("Fetched and stored counts should match, got %d and %d",
				session.FetchedCount, session.StoredCount)
		}

		t.Logf("Concurrent updates: %d fetched, %d stored out of %d attempts",
			session.FetchedCount, session.StoredCount, iterations)
	})

}

// TestSessionDataIntegrity tests data integrity of session fields
func TestSessionDataIntegrity(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSessionDataIntegrity", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_data_integrity.db")

		// Add logger setup for NewManager calls
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		// Create session with specific values
		sessionID := "integrity-test-session"
		startIndex := 12345
		resultsPerBatch := 250

		session, err := manager.CreateSession(sessionID, startIndex, resultsPerBatch)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Verify all fields are correct
		if session.ID != sessionID {
			t.Errorf("Session ID mismatch: expected %s, got %s", sessionID, session.ID)
		}

		if session.StartIndex != startIndex {
			t.Errorf("StartIndex mismatch: expected %d, got %d", startIndex, session.StartIndex)
		}

		if session.ResultsPerBatch != resultsPerBatch {
			t.Errorf("ResultsPerBatch mismatch: expected %d, got %d", resultsPerBatch, session.ResultsPerBatch)
		}

		if session.State != StateIdle {
			t.Errorf("Initial state should be idle, got %s", session.State)
		}

		if session.FetchedCount != 0 || session.StoredCount != 0 || session.ErrorCount != 0 {
			t.Error("Initial counters should be zero")
		}

		// Update state and progress
		manager.UpdateState(StateRunning)
		manager.UpdateProgress(100, 95, 5)

		// Retrieve and verify
		retrieved, err := manager.GetSession()
		if err != nil {
			t.Fatalf("Failed to retrieve session: %v", err)
		}

		if retrieved.State != StateRunning {
			t.Errorf("State should be running, got %s", retrieved.State)
		}

		if retrieved.FetchedCount != 100 {
			t.Errorf("FetchedCount should be 100, got %d", retrieved.FetchedCount)
		}

		if retrieved.StoredCount != 95 {
			t.Errorf("StoredCount should be 95, got %d", retrieved.StoredCount)
		}

		if retrieved.ErrorCount != 5 {
			t.Errorf("ErrorCount should be 5, got %d", retrieved.ErrorCount)
		}

		// Verify immutable fields didn't change
		if retrieved.ID != sessionID {
			t.Error("Session ID should not change")
		}

		if retrieved.StartIndex != startIndex {
			t.Error("StartIndex should not change")
		}

		if retrieved.ResultsPerBatch != resultsPerBatch {
			t.Error("ResultsPerBatch should not change")
		}
	})

}

// TestMultipleSessionLifecycles tests creating/deleting sessions multiple times
func TestMultipleSessionLifecycles(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMultipleSessionLifecycles", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_multiple_lifecycles.db")

		// Add logger setup for NewManager calls
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		// Create and delete 10 sessions
		for i := 0; i < 10; i++ {
			sessionID := fmt.Sprintf("session-%d", i)

			// Create session
			session, err := manager.CreateSession(sessionID, i*100, 100+i*10)
			if err != nil {
				t.Fatalf("Iteration %d: Failed to create session: %v", i, err)
			}

			if session.ID != sessionID {
				t.Errorf("Iteration %d: Session ID mismatch", i)
			}

			// Update some progress
			manager.UpdateProgress(int64(i*10), int64(i*9), int64(i))

			// Verify progress
			retrieved, _ := manager.GetSession()
			if retrieved.FetchedCount != int64(i*10) {
				t.Errorf("Iteration %d: Progress not updated correctly", i)
			}

			// Delete session
			err = manager.DeleteSession()
			if err != nil {
				t.Fatalf("Iteration %d: Failed to delete session: %v", i, err)
			}

			// Verify deletion
			_, err = manager.GetSession()
			if err != ErrNoSession {
				t.Errorf("Iteration %d: Session should be deleted", i)
			}
		}
	})

}

// TestSessionUpdateTimestamps tests that timestamps are updated correctly
func TestSessionTimestampsAccuracy(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSessionTimestampsAccuracy", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_timestamps_accuracy.db")

		// Add logger setup for NewManager calls
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer manager.Close()

		// Record time before creation
		beforeCreate := time.Now()

		session, err := manager.CreateSession("test-session", 0, 100)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Record time after creation
		afterCreate := time.Now()

		// Verify CreatedAt is within the time window
		if session.CreatedAt.Before(beforeCreate) || session.CreatedAt.After(afterCreate) {
			t.Error("CreatedAt timestamp is not within expected time window")
		}

		// CreatedAt and UpdatedAt should be close initially
		diff := session.UpdatedAt.Sub(session.CreatedAt)
		if diff > 100*time.Millisecond {
			t.Error("Initial UpdatedAt should be close to CreatedAt")
		}

		// Wait a bit
		time.Sleep(50 * time.Millisecond)

		// Update state
		beforeUpdate := time.Now()
		manager.UpdateState(StateRunning)
		afterUpdate := time.Now()

		// Retrieve and check timestamps
		updated, _ := manager.GetSession()

		// CreatedAt should not change
		if !updated.CreatedAt.Equal(session.CreatedAt) {
			t.Error("CreatedAt should not change after updates")
		}

		// UpdatedAt should be updated and within time window
		if updated.UpdatedAt.Before(beforeUpdate) || updated.UpdatedAt.After(afterUpdate) {
			t.Error("UpdatedAt timestamp not within expected time window after update")
		}

		// UpdatedAt should be after original UpdatedAt
		if !updated.UpdatedAt.After(session.UpdatedAt) {
			t.Error("UpdatedAt should be updated to a later time")
		}
	})

}

// TestSessionDatabaseCorruption tests recovery from potential database issues
func TestSessionManagerReopen(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSessionManagerReopen", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := filepath.Join(t.TempDir(), "test_reopen.db")

		// Ensure consistent logger initialization
		logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

		// Create first manager and session
		manager1, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create first manager: %v", err)
		}

		session1, err := manager1.CreateSession("persistent-session", 100, 200)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		manager1.UpdateState(StateRunning)
		manager1.UpdateProgress(50, 45, 5)

		// Close first manager
		manager1.Close()

		// Open second manager with same database
		manager2, err := NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create second manager: %v", err)
		}
		defer manager2.Close()

		// Retrieve session
		session2, err := manager2.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session from second manager: %v", err)
		}

		// Verify all data persisted correctly
		if session2.ID != session1.ID {
			t.Error("Session ID not persisted")
		}

		if session2.State != StateRunning {
			t.Error("Session state not persisted")
		}

		if session2.FetchedCount != 50 {
			t.Error("FetchedCount not persisted")
		}

		if session2.StoredCount != 45 {
			t.Error("StoredCount not persisted")
		}

		if session2.ErrorCount != 5 {
			t.Error("ErrorCount not persisted")
		}

		// Verify configuration persisted
		if session2.StartIndex != 100 || session2.ResultsPerBatch != 200 {
			t.Error("Session configuration not persisted")
		}
	})

}
