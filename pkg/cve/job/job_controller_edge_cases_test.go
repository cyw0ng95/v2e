package job

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	subprocess "github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// MockRPCInvoker is a mock implementation of RPCInvoker for testing
type MockRPCInvoker struct {
	mu            sync.Mutex
	callHistory   []CallRecord
	fetchCVEsFunc func(ctx context.Context, params interface{}) (interface{}, error)
	saveCVEFunc   func(ctx context.Context, params interface{}) (interface{}, error)
}

type CallRecord struct {
	Method string
	Params interface{}
}

func (m *MockRPCInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := CallRecord{Method: method, Params: params}
	m.callHistory = append(m.callHistory, record)

	switch {
	case method == "RPCFetchCVEs":
		if m.fetchCVEsFunc != nil {
			return m.fetchCVEsFunc(ctx, params)
		}
		// Default behavior: return a mock response
		return createMockCVEMessage(), nil
	case method == "RPCSaveCVEByID":
		if m.saveCVEFunc != nil {
			return m.saveCVEFunc(ctx, params)
		}
		// Default behavior: return success
		return createMockResponseMessage(), nil
	default:
		return nil, errors.New("unknown method")
	}
}

func (m *MockRPCInvoker) GetCallHistory() []CallRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	history := make([]CallRecord, len(m.callHistory))
	copy(history, m.callHistory)
	return history
}

func (m *MockRPCInvoker) ResetHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callHistory = nil
}

// createMockCVEMessage creates a mock subprocess.Message for CVE response
func createMockCVEMessage() *subprocess.Message {
	// Create a simple mock CVE response
	mockResponse := `{
		"resultsPerPage": 1,
		"startIndex": 0,
		"totalResults": 1,
		"format": "MITRE",
		"version": "4.0",
		"timestamp": "2023-01-01T00:00:00.000",
		"vulnerabilities": [
			{
				"cve": {
					"id": "CVE-2023-1234",
					"sourceIdentifier": "test@example.com",
					"published": "2023-01-01T00:00:00.000",
					"lastModified": "2023-01-01T00:00:00.000",
					"vulnStatus": "Analyzed",
					"descriptions": [
						{
							"lang": "en",
							"value": "Test vulnerability description"
						}
					]
				}
			}
		]
	}`

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: []byte(mockResponse),
	}
	return msg
}

// createMockResponseMessage creates a mock response message
func createMockResponseMessage() *subprocess.Message {
	msg := &subprocess.Message{
		Type: subprocess.MessageTypeResponse,
	}
	return msg
}

// TestJobController_StartStop tests starting and stopping the job controller
func TestJobController_StartStop(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_StartStop", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_start_stop.db")
		logger := common.NewLogger(os.Stdout, "JOB_START_STOP_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{}
		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("start-stop-test-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Test starting the controller
		ctx := context.Background()
		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Verify it's running
		if !controller.IsRunning() {
			t.Error("Controller should be running after Start")
		}

		// Try to start again (should fail)
		err = controller.Start(ctx)
		if err != ErrJobRunning {
			t.Errorf("Expected ErrJobRunning when starting already running job, got %v", err)
		}

		// Stop the controller
		err = controller.Stop()
		if err != nil {
			t.Fatalf("Failed to stop controller: %v", err)
		}

		// Verify it's not running
		if controller.IsRunning() {
			t.Error("Controller should not be running after Stop")
		}

		// Try to stop again (should fail)
		err = controller.Stop()
		if err != ErrJobNotRunning {
			t.Errorf("Expected ErrJobNotRunning when stopping non-running job, got %v", err)
		}
	})

}

// TestJobController_PauseResume tests pausing and resuming the job controller
func TestJobController_PauseResume(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_PauseResume", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_pause_resume.db")
		logger := common.NewLogger(os.Stdout, "JOB_PAUSE_RESUME_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{}
		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("pause-resume-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Start the controller
		ctx := context.Background()
		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Pause the controller
		err = controller.Pause()
		if err != nil {
			t.Fatalf("Failed to pause controller: %v", err)
		}

		// Verify it's not running
		if controller.IsRunning() {
			t.Error("Controller should not be running after Pause")
		}

		// Verify session state is paused
		sess, err := sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}
		if sess.State != session.StatePaused {
			t.Errorf("Expected session state %s after pause, got %s", session.StatePaused, sess.State)
		}

		// Try to resume
		err = controller.Resume(ctx)
		if err != nil {
			t.Fatalf("Failed to resume controller: %v", err)
		}

		// Verify it's running again
		if !controller.IsRunning() {
			t.Error("Controller should be running after Resume")
		}

		// Verify session state is running
		sess, err = sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}
		if sess.State != session.StateRunning {
			t.Errorf("Expected session state %s after resume, got %s", session.StateRunning, sess.State)
		}

		// Try to resume when already running (should fail)
		err = controller.Resume(ctx)
		if err != ErrJobRunning {
			t.Errorf("Expected ErrJobRunning when resuming running job, got %v", err)
		}
	})

}

// TestJobController_ResumeWithoutPause tests resuming without pausing first
func TestJobController_ResumeWithoutPause(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_ResumeWithoutPause", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_resume_without_pause.db")
		logger := common.NewLogger(os.Stdout, "JOB_RESUME_NO_PAUSE_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{}
		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("resume-no-pause-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Try to resume without pausing (should fail)
		ctx := context.Background()
		err = controller.Resume(ctx)
		if err == nil {
			t.Error("Expected error when resuming non-paused job")
		}
	})

}

// TestJobController_RPCFailureHandling tests handling of RPC failures
func TestJobController_RPCFailureHandling(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_RPCFailureHandling", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_rpc_failure_handling.db")
		logger := common.NewLogger(os.Stdout, "JOB_RPC_FAILURE_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		// Create invoker that returns an error for fetch operations
		mockInvoker := &MockRPCInvoker{
			fetchCVEsFunc: func(ctx context.Context, params interface{}) (interface{}, error) {
				return nil, errors.New("network error")
			},
		}

		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("rpc-failure-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Start the controller
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Give it some time to run and handle errors
		time.Sleep(2 * time.Second)

		// Stop the controller
		controller.Stop()

		// Check that progress was updated to reflect errors
		sess, err := sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		// Errors should have been counted
		if sess.ErrorCount <= 0 {
			t.Errorf("Expected error count > 0, got %d", sess.ErrorCount)
		}
	})

}

// TestJobController_SaveFailureHandling tests handling of save failures
func TestJobController_SaveFailureHandling(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_SaveFailureHandling", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_save_failure_handling.db")
		logger := common.NewLogger(os.Stdout, "JOB_SAVE_FAILURE_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		// Create invoker that returns an error for save operations
		mockInvoker := &MockRPCInvoker{
			saveCVEFunc: func(ctx context.Context, params interface{}) (interface{}, error) {
				return nil, errors.New("database error")
			},
		}

		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("save-failure-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Start the controller
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Give it some time to run and handle errors
		time.Sleep(2 * time.Second)

		// Stop the controller
		controller.Stop()

		// Check that progress was updated to reflect errors
		sess, err := sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		// Errors should have been counted
		if sess.ErrorCount <= 0 {
			t.Errorf("Expected error count > 0 due to save failures, got %d", sess.ErrorCount)
		}
	})

}

// TestJobController_ConcurrentAccess tests concurrent access to the controller
func TestJobController_ConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_ConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_concurrent_access.db")
		logger := common.NewLogger(os.Stdout, "JOB_CONCURRENT_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{}
		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("concurrent-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Test concurrent access to IsRunning
		var wg sync.WaitGroup
		const numGoroutines = 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					_ = controller.IsRunning()
				}
			}()
		}

		wg.Wait()

		// Start the controller
		ctx := context.Background()
		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Test concurrent access while running
		var wg2 sync.WaitGroup
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg2.Add(1)
			go func(id int) {
				defer wg2.Done()
				// Each goroutine tries to stop and start the controller
				// This should result in some errors due to state conflicts
				err := controller.Stop()
				if err != nil && err != ErrJobNotRunning {
					errChan <- err
				}
			}(i)
		}

		wg2.Wait()
		close(errChan)

		// Check for unexpected errors
		for err := range errChan {
			if err != nil && err != ErrJobNotRunning {
				t.Errorf("Unexpected error in concurrent access: %v", err)
			}
		}
	})

}

// TestJobController_ContextCancellation tests cancellation via context
func TestJobController_ContextCancellation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_ContextCancellation", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_context_cancellation.db")
		logger := common.NewLogger(os.Stdout, "JOB_CANCEL_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{
			// Simulate slow operations to allow for cancellation
			fetchCVEsFunc: func(ctx context.Context, params interface{}) (interface{}, error) {
				select {
				case <-time.After(100 * time.Millisecond):
					return createMockCVEMessage(), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}

		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("cancel-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Create a cancellable context
		ctx, cancel := context.WithCancel(context.Background())

		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Give it a moment to start
		time.Sleep(50 * time.Millisecond)

		// Cancel the context
		cancel()

		// Give time for cancellation to propagate
		time.Sleep(200 * time.Millisecond)

		// Controller should no longer be running
		if controller.IsRunning() {
			t.Error("Controller should not be running after context cancellation")
		}
	})

}

// TestJobController_SessionManagement tests proper session management
func TestJobController_SessionManagement(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_SessionManagement", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_session_management.db")
		logger := common.NewLogger(os.Stdout, "JOB_SESSION_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{}
		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		sessionID := "session-management-test"
		_, err = sessionMgr.CreateSession(sessionID, 50, 25)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Start the controller
		ctx := context.Background()
		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Verify session state is updated to Running
		sess, err := sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}
		if sess.State != session.StateRunning {
			t.Errorf("Expected session state %s after start, got %s", session.StateRunning, sess.State)
		}

		// Stop the controller
		err = controller.Stop()
		if err != nil {
			t.Fatalf("Failed to stop controller: %v", err)
		}

		// Verify session state is updated to Stopped
		sess, err = sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}
		if sess.State != session.StateStopped {
			t.Errorf("Expected session state %s after stop, got %s", session.StateStopped, sess.State)
		}
	})

}

// TestJobController_NoSession tests behavior when no session exists
func TestJobController_NoSession(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_NoSession", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_no_session.db")
		logger := common.NewLogger(os.Stdout, "JOB_NO_SESSION_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		mockInvoker := &MockRPCInvoker{}
		controller := NewController(mockInvoker, sessionMgr, logger)

		// Try to start without creating a session first (should fail)
		ctx := context.Background()
		err = controller.Start(ctx)
		if err == nil {
			t.Error("Expected error when starting without session")
		}
	})

}

// TestJobController_RPCResponseTypes tests handling of different RPC response types
func TestJobController_RPCResponseTypes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobController_RPCResponseTypes", nil, func(t *testing.T, tx *gorm.DB) {
		// Create mocks with a unique temp file for each test to avoid conflicts
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_job_controller_rpc_response_types.db")
		logger := common.NewLogger(os.Stdout, "JOB_RPC_TYPES_TEST", common.DebugLevel)
		sessionMgr, err := session.NewManager(dbPath, logger)
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}
		defer sessionMgr.Close()

		// Test with error message response
		mockInvoker := &MockRPCInvoker{
			fetchCVEsFunc: func(ctx context.Context, params interface{}) (interface{}, error) {
				// Return an error message
				msg := &subprocess.Message{
					Type:  subprocess.MessageTypeError,
					Error: "API rate limit exceeded",
				}
				return msg, nil
			},
		}

		controller := NewController(mockInvoker, sessionMgr, logger)

		// Create a session first
		_, err = sessionMgr.CreateSession("rpc-types-session", 0, 10)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Start the controller
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = controller.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start controller: %v", err)
		}

		// Give it time to process the error response
		time.Sleep(1 * time.Second)

		// Stop the controller
		controller.Stop()

		// Check that errors were tracked
		sess, err := sessionMgr.GetSession()
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		if sess.ErrorCount <= 0 {
			t.Errorf("Expected error count > 0 after receiving error response, got %d", sess.ErrorCount)
		}
	})

}
