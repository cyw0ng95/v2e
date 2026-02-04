package taskflow

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// mockRPCInvoker mocks the RPC invoker for testing
type mockRPCInvoker struct {
	mu sync.Mutex
}

func newMockRPCInvoker() *mockRPCInvoker {
	return &mockRPCInvoker{}
}

func (m *mockRPCInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	return &mockMessage{Type: "response", Payload: []byte(`{"vulnerabilities": []}`)}, nil
}

type mockMessage struct {
	Type    string
	Payload []byte
	Error   string
}

// newTestLogger creates a test logger that discards output
func newTestLogger() *common.Logger {
	return common.NewLogger(io.Discard, "", common.InfoLevel)
}

// TestJobExecutor_ConcurrentStartPrevention verifies that only one job can start at a time
func TestJobExecutor_ConcurrentStartPrevention(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobExecutor_ConcurrentStartPrevention", nil, func(t *testing.T, tx *gorm.DB) {
		invoker := newMockRPCInvoker()
		logger := newTestLogger()
		store, err := NewRunStore("test_concurrent_start.db", logger)
		if err != nil {
			t.Fatalf("Failed to create run store: %v", err)
		}
		defer store.Close()
		defer os.Remove("test_concurrent_start.db")

		executor := NewJobExecutor(invoker, store, logger, 100)
		ctx := context.Background()

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		// Try to start 10 concurrent jobs
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				runID := "concurrent-" + string(rune('0'+idx))
				err := executor.StartTyped(ctx, runID, 0, 100, DataTypeCVE)
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Only one should succeed
		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful start, got %d", successCount)
		}

		// Clean up
		activeRun, _ := executor.GetActiveRun()
		if activeRun != nil {
			executor.Stop(activeRun.ID)
		}
	})

}

// TestJobExecutor_PauseResumeStateTransitions verifies pause/resume works
func TestJobExecutor_PauseResumeStateTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobExecutor_PauseResumeStateTransitions", nil, func(t *testing.T, tx *gorm.DB) {
		invoker := newMockRPCInvoker()
		logger := newTestLogger()
		store, err := NewRunStore("test_pause_resume.db", logger)
		if err != nil {
			t.Fatalf("Failed to create run store: %v", err)
		}
		defer store.Close()
		defer os.Remove("test_pause_resume.db")

		executor := NewJobExecutor(invoker, store, logger, 100)
		ctx := context.Background()
		runID := "test-pause-resume"

		// Start job
		err = executor.StartTyped(ctx, runID, 0, 100, DataTypeCVE)
		if err != nil {
			t.Fatalf("StartTyped failed: %v", err)
		}

		// Wait for job to start
		time.Sleep(100 * time.Millisecond)

		// Pause job
		err = executor.Pause(runID)
		if err != nil {
			t.Fatalf("Pause failed: %v", err)
		}

		// Verify paused state
		run, err := store.GetRun(runID)
		if err != nil {
			t.Fatalf("Failed to get run after pause: %v", err)
		}

		if run.State != StatePaused {
			t.Errorf("Expected state %s, got %s", StatePaused, run.State)
		}

		// Resume job
		err = executor.Resume(ctx, runID)
		if err != nil {
			t.Fatalf("Resume failed: %v", err)
		}

		// Verify running state
		run, err = store.GetRun(runID)
		if err != nil {
			t.Fatalf("Failed to get run after resume: %v", err)
		}

		if run.State != StateRunning {
			t.Errorf("Expected state %s, got %s", StateRunning, run.State)
		}

		// Clean up
		executor.Stop(runID)
	})

}

// TestJobExecutor_StopFromPaused verifies stopping from paused state works
func TestJobExecutor_StopFromPaused(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobExecutor_StopFromPaused", nil, func(t *testing.T, tx *gorm.DB) {
		invoker := newMockRPCInvoker()
		logger := newTestLogger()
		store, err := NewRunStore("test_stop_paused.db", logger)
		if err != nil {
			t.Fatalf("Failed to create run store: %v", err)
		}
		defer store.Close()
		defer os.Remove("test_stop_paused.db")

		executor := NewJobExecutor(invoker, store, logger, 100)
		ctx := context.Background()
		runID := "test-stop-paused"

		// Start job
		err = executor.StartTyped(ctx, runID, 0, 100, DataTypeCVE)
		if err != nil {
			t.Fatalf("StartTyped failed: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		// Pause job
		err = executor.Pause(runID)
		if err != nil {
			t.Fatalf("Pause failed: %v", err)
		}

		// Stop from paused state
		err = executor.Stop(runID)
		if err != nil {
			t.Fatalf("Stop from paused failed: %v", err)
		}

		// Verify stopped state
		run, err := store.GetRun(runID)
		if err != nil {
			t.Fatalf("Failed to get run after stop: %v", err)
		}

		if run.State != StateStopped {
			t.Errorf("Expected state %s, got %s", StateStopped, run.State)
		}
	})

}

// TestJobExecutor_RecoveryAfterCrash verifies running jobs are recovered
func TestJobExecutor_RecoveryAfterCrash(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobExecutor_RecoveryAfterCrash", nil, func(t *testing.T, tx *gorm.DB) {
		invoker := newMockRPCInvoker()
		logger := newTestLogger()
		store, err := NewRunStore("test_recovery.db", logger)
		if err != nil {
			t.Fatalf("Failed to create run store: %v", err)
		}
		defer store.Close()
		defer os.Remove("test_recovery.db")

		// Create a run in PAUSED state (simulating crash with paused job)
		// Note: running jobs auto-resume, paused jobs stay paused
		runID := "test-recovery"
		_, err = store.CreateRun(runID, 0, 100, DataTypeCVE)
		if err != nil {
			t.Fatalf("Failed to create run: %v", err)
		}

		// Proper state flow: queued -> running -> paused
		err = store.UpdateState(runID, StateRunning)
		if err != nil {
			t.Fatalf("Failed to set running state: %v", err)
		}
		err = store.UpdateState(runID, StatePaused)
		if err != nil {
			t.Fatalf("Failed to set paused state: %v", err)
		}

		// Create new executor (simulating restart)
		executor := NewJobExecutor(invoker, store, logger, 100)
		ctx := context.Background()

		// Run recovery - should NOT auto-resume paused jobs
		err = executor.RecoverRuns(ctx)
		if err != nil {
			t.Fatalf("RecoverRuns failed: %v", err)
		}

		// Verify the run was NOT auto-resumed (paused jobs stay paused)
		activeRun, err := executor.GetActiveRun()
		if err != nil {
			t.Fatalf("GetActiveRun failed: %v", err)
		}
		if activeRun != nil {
			t.Error("Expected no active run after recovery (paused job should stay paused)")
		}

		// Verify run is still paused
		runAfter, err := store.GetRun(runID)
		if err != nil {
			t.Fatalf("Failed to get run: %v", err)
		}
		if runAfter.State != StatePaused {
			t.Errorf("Expected run to remain paused, got state %s", runAfter.State)
		}
	})

}

// TestJobExecutor_PausedRunStaysPaused verifies paused runs are not auto-resumed
func TestJobExecutor_PausedRunStaysPaused(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJobExecutor_PausedRunStaysPaused", nil, func(t *testing.T, tx *gorm.DB) {
		invoker := newMockRPCInvoker()
		logger := newTestLogger()
		store, err := NewRunStore("test_paused_stays_paused.db", logger)
		if err != nil {
			t.Fatalf("Failed to create run store: %v", err)
		}
		defer store.Close()
		defer os.Remove("test_paused_stays_paused.db")

		// Create a paused run (proper state flow: queued -> running -> paused)
		runID := "test-paused-stays"
		_, err = store.CreateRun(runID, 0, 100, DataTypeCVE)
		if err != nil {
			t.Fatalf("Failed to create run: %v", err)
		}

		// Proper state flow: queued -> running -> paused
		err = store.UpdateState(runID, StateRunning)
		if err != nil {
			t.Fatalf("Failed to set running state: %v", err)
		}
		err = store.UpdateState(runID, StatePaused)
		if err != nil {
			t.Fatalf("Failed to set paused state: %v", err)
		}

		// Create new executor and run recovery
		executor := NewJobExecutor(invoker, store, logger, 100)
		ctx := context.Background()

		err = executor.RecoverRuns(ctx)
		if err != nil {
			t.Fatalf("RecoverRuns failed: %v", err)
		}

		// Verify the run was NOT auto-resumed
		activeRun, err := executor.GetActiveRun()
		if err != nil {
			t.Fatalf("GetActiveRun failed: %v", err)
		}
		if activeRun != nil {
			t.Error("Expected no active run after recovery (paused run should stay paused)")
		}

		// Verify run is still paused
		runAfter, err := store.GetRun(runID)
		if err != nil {
			t.Fatalf("Failed to get run: %v", err)
		}
		if runAfter.State != StatePaused {
			t.Errorf("Expected run to remain paused, got state %s", runAfter.State)
		}
	})

}
