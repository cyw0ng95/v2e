package job

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// mockJobStateRPCInvoker simulates RPC calls for testing job state transitions.
type mockJobStateRPCInvoker struct {
	mu              sync.Mutex
	callCount       int
	shouldFail      bool
	failAfterN      int
	returnEmptyData bool
}

func (m *mockJobStateRPCInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail || (m.failAfterN > 0 && m.callCount > m.failAfterN) {
		return nil, errors.New("mock RPC failure")
	}

	// Handle different RPC methods appropriately
	switch method {
	case "RPCSaveCVEByID":
		// Return success response for save operations
		savePayload := []byte(`{"success": true}`)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: savePayload}, nil
	case "RPCFetchCVEs":
		if m.returnEmptyData {
			return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: []byte(`{"vulnerabilities":[]}`)}, nil
		}
		// Return valid CVE data to keep job running
		cveData := []byte(`{"vulnerabilities":[{"cve":{"id":"CVE-2024-0001","sourceIdentifier":"nvd@nist.gov","published":"2024-01-01T00:00:00.000Z","lastModified":"2024-01-01T00:00:00.000Z","vulnStatus":"Pending"}}]}`)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: cveData}, nil
	default:
		return nil, errors.New("unknown RPC method: " + method)
	}
}

// TestController_StartStop ensures basic start/stop lifecycle.
func TestController_StartStop(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_StartStop", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		if ctrl.IsRunning() {
			t.Fatalf("controller should not be running initially")
		}

		if err := ctrl.Start(context.Background()); err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		if !ctrl.IsRunning() {
			t.Fatalf("controller should be running after Start")
		}

		// Attempt to start again should fail
		if err := ctrl.Start(context.Background()); err != ErrJobRunning {
			t.Fatalf("expected ErrJobRunning, got %v", err)
		}

		if err := ctrl.Stop(); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		if ctrl.IsRunning() {
			t.Fatalf("controller should not be running after Stop")
		}
	})

}

// TestController_PauseResume ensures pause/resume transitions work correctly.
func TestController_PauseResume(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_PauseResume", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		ctrl.Start(context.Background())

		if err := ctrl.Pause(); err != nil {
			t.Fatalf("Pause failed: %v", err)
		}

		if ctrl.IsRunning() {
			t.Fatalf("controller should not be running after Pause")
		}

		sess, _ := sessMgr.GetSession()
		if sess.State != session.StatePaused {
			t.Fatalf("session state should be paused, got %s", sess.State)
		}

		if err := ctrl.Resume(context.Background()); err != nil {
			t.Fatalf("Resume failed: %v", err)
		}

		if !ctrl.IsRunning() {
			t.Fatalf("controller should be running after Resume")
		}

		ctrl.Stop()
	})

}

// TestController_StopNotRunning ensures Stop fails when not running.
func TestController_StopNotRunning(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_StopNotRunning", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		if err := ctrl.Stop(); err != ErrJobNotRunning {
			t.Fatalf("expected ErrJobNotRunning, got %v", err)
		}
	})

}

// TestController_PauseNotRunning ensures Pause fails when not running.
func TestController_PauseNotRunning(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_PauseNotRunning", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		if err := ctrl.Pause(); err != ErrJobNotRunning {
			t.Fatalf("expected ErrJobNotRunning, got %v", err)
		}
	})

}

// TestController_ResumeNotPaused ensures Resume fails when session is not paused.
func TestController_ResumeNotPaused(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_ResumeNotPaused", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)
		sessMgr.UpdateState(session.StateIdle)

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		err := ctrl.Resume(context.Background())
		if err == nil || err.Error() != "session is not paused (current state: idle)" {
			t.Fatalf("expected state mismatch error, got %v", err)
		}
	})

}

// TestController_StartWithNoSession ensures Start fails when no session exists.
func TestController_StartWithNoSession(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_StartWithNoSession", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		err := ctrl.Start(context.Background())
		if err == nil || !errors.Is(err, session.ErrNoSession) {
			t.Fatalf("expected no session error, got %v", err)
		}

		if ctrl.IsRunning() {
			t.Fatalf("controller should not be running after failed Start")
		}
	})

}

// TestController_ConcurrentStartAttempts ensures only one Start succeeds.
func TestController_ConcurrentStartAttempts(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_ConcurrentStartAttempts", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)

		invoker := &mockJobStateRPCInvoker{}
		ctrl := NewController(invoker, sessMgr, logger)

		numGoroutines := 20
		var wg sync.WaitGroup
		successes := 0
		var mu sync.Mutex

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				err := ctrl.Start(context.Background())
				if err == nil {
					mu.Lock()
					successes++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		mu.Lock()
		defer mu.Unlock()
		if successes != 1 {
			t.Fatalf("expected exactly 1 Start to succeed, got %d", successes)
		}

		ctrl.Stop()
	})

}

// TestController_JobContextCancellation ensures job stops when context is cancelled.
func TestController_JobContextCancellation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestController_JobContextCancellation", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDir := t.TempDir()
		logger := common.NewLogger(testWriter{t}, "test", common.ErrorLevel)
		sessMgr, _ := session.NewManager(tmpDir+"/session.db", logger)
		defer sessMgr.Close()

		sessMgr.CreateSession("sess1", 0, 100)

		invoker := &mockJobStateRPCInvoker{returnEmptyData: true}
		ctrl := NewController(invoker, sessMgr, logger)

		ctx, cancel := context.WithCancel(context.Background())
		ctrl.Start(ctx)
		time.Sleep(50 * time.Millisecond)

		cancel()
		time.Sleep(50 * time.Millisecond)

		if ctrl.IsRunning() {
			t.Fatalf("controller should stop after context cancellation")
		}
	})

}

// testWriter adapts testing.T to io.Writer for logger output.
type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Logf("%s", string(p))
	return len(p), nil
}
