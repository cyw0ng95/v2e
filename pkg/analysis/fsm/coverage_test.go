package fsm

import (
	"fmt"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestAnalyzeFSM_CanProcess(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)

	testutils.Run(t, testutils.Level1, "CanProcess_Idle", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if !baseFSM.CanProcess() {
			t.Error("Expected CanProcess to return true in IDLE state")
		}
	})

	testutils.Run(t, testutils.Level1, "CanProcess_Processing", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()

		graphFSM := fsm.GetGraphFSM()
		graphFSM.StartBuild()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if !baseFSM.CanProcess() {
			t.Error("Expected CanProcess to return true in PROCESSING state")
		}
	})

	testutils.Run(t, testutils.Level1, "CanProcess_Paused", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()
		fsm.Pause()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if baseFSM.CanProcess() {
			t.Error("Expected CanProcess to return false in PAUSED state")
		}
	})

	testutils.Run(t, testutils.Level1, "CanProcess_Terminated", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()
		fsm.Stop()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if baseFSM.CanProcess() {
			t.Error("Expected CanProcess to return false in TERMINATED state")
		}
	})
}

func TestAnalyzeFSM_IsHealthy(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)

	testutils.Run(t, testutils.Level1, "IsHealthy_Idle", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if !baseFSM.IsHealthy() {
			t.Error("Expected IsHealthy to return true in IDLE state")
		}
	})

	testutils.Run(t, testutils.Level1, "IsHealthy_Processing", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()

		graphFSM := fsm.GetGraphFSM()
		graphFSM.StartBuild()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if !baseFSM.IsHealthy() {
			t.Error("Expected IsHealthy to return true in PROCESSING state")
		}
	})

	testutils.Run(t, testutils.Level1, "IsHealthy_Draining", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()

		graphFSM := fsm.GetGraphFSM()
		graphFSM.StartBuild()
		fsm.Stop()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if baseFSM.IsHealthy() {
			t.Error("Expected IsHealthy to return false in DRAINING state")
		}
	})

	testutils.Run(t, testutils.Level1, "IsHealthy_Terminated", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewAnalyzeFSM(logger)
		fsm.Start()
		fsm.Stop()

		baseFSM := fsm.(*BaseAnalyzeFSM)
		if baseFSM.IsHealthy() {
			t.Error("Expected IsHealthy to return false in TERMINATED state")
		}
	})
}

func TestGraphFSM_FailPersist(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)

	testutils.Run(t, testutils.Level1, "FailPersist_TransitionsToError", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewGraphFSM(logger)

		fsm.StartBuild()
		fsm.CompleteBuild()
		fsm.StartPersist()

		testErr := fmt.Errorf("persist failed")
		if err := fsm.FailPersist(testErr); err != nil {
			t.Fatalf("FailPersist failed: %v", err)
		}

		if fsm.GetState() != GraphError {
			t.Errorf("Expected ERROR state, got %s", fsm.GetState())
		}
	})
}

func TestGraphFSM_Clear(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)

	testutils.Run(t, testutils.Level1, "Clear_FromReady", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewGraphFSM(logger)

		fsm.StartBuild()
		fsm.CompleteBuild()

		if err := fsm.Clear(); err != nil {
			t.Fatalf("Clear failed: %v", err)
		}

		if fsm.GetState() != GraphIdle {
			t.Errorf("Expected IDLE state, got %s", fsm.GetState())
		}
	})

	testutils.Run(t, testutils.Level1, "Clear_EmitsEvent", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewGraphFSM(logger)
		fsm.StartBuild()
		fsm.CompleteBuild()

		var receivedEvent *Event
		fsm.SetEventHandler(func(event *Event) error {
			receivedEvent = event
			return nil
		})

		fsm.Clear()

		if receivedEvent == nil {
			t.Error("Expected event to be emitted")
		}

		if receivedEvent.Type != EventGraphCleared {
			t.Errorf("Expected event type %s, got %s", EventGraphCleared, receivedEvent.Type)
		}
	})
}

func TestGraphFSM_GetLastError(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)

	testutils.Run(t, testutils.Level1, "GetLastError_AfterFailure", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewGraphFSM(logger)

		testErr := fmt.Errorf("test error")
		fsm.StartBuild()
		fsm.FailBuild(testErr)

		baseFSM := fsm.(*BaseGraphFSM)
		lastErr := baseFSM.GetLastError()

		if lastErr == nil {
			t.Error("Expected last error to be set")
		}
		if lastErr.Error() != "test error" {
			t.Errorf("Expected error message 'test error', got '%s'", lastErr.Error())
		}
	})

	testutils.Run(t, testutils.Level1, "GetLastError_AfterReset", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewGraphFSM(logger)

		testErr := fmt.Errorf("test error")
		fsm.StartBuild()
		fsm.FailBuild(testErr)

		baseFSM := fsm.(*BaseGraphFSM)
		baseFSM.Reset()

		if baseFSM.GetLastError() != nil {
			t.Error("Expected last error to be cleared after reset")
		}
	})
}

func TestGraphFSM_Stats(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)

	testutils.Run(t, testutils.Level1, "Stats_ReturnsCurrentState", nil, func(t *testing.T, tx *gorm.DB) {
		fsm := NewGraphFSM(logger)
		fsm.StartBuild()

		if fsm.GetState() != GraphBuilding {
			t.Errorf("Expected BUILDING state, got %s", fsm.GetState())
		}
	})
}
