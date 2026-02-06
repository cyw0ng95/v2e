package fsm

import (
	"fmt"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func TestGraphFSM_StateTransitions(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	fsm := NewGraphFSM(logger)

	// Initial state should be IDLE
	if fsm.GetState() != GraphIdle {
		t.Errorf("Expected initial state IDLE, got %s", fsm.GetState())
	}

	// Test successful build flow
	if err := fsm.StartBuild(); err != nil {
		t.Errorf("StartBuild failed: %v", err)
	}
	if fsm.GetState() != GraphBuilding {
		t.Errorf("Expected BUILDING state, got %s", fsm.GetState())
	}

	if err := fsm.CompleteBuild(); err != nil {
		t.Errorf("CompleteBuild failed: %v", err)
	}
	if fsm.GetState() != GraphReady {
		t.Errorf("Expected READY state, got %s", fsm.GetState())
	}

	// Test analysis flow
	if err := fsm.StartAnalysis(); err != nil {
		t.Errorf("StartAnalysis failed: %v", err)
	}
	if fsm.GetState() != GraphAnalyzing {
		t.Errorf("Expected ANALYZING state, got %s", fsm.GetState())
	}

	if err := fsm.CompleteAnalysis(); err != nil {
		t.Errorf("CompleteAnalysis failed: %v", err)
	}
	if fsm.GetState() != GraphReady {
		t.Errorf("Expected READY state, got %s", fsm.GetState())
	}

	// Test persistence flow
	if err := fsm.StartPersist(); err != nil {
		t.Errorf("StartPersist failed: %v", err)
	}
	if fsm.GetState() != GraphPersisting {
		t.Errorf("Expected PERSISTING state, got %s", fsm.GetState())
	}

	if err := fsm.CompletePersist(); err != nil {
		t.Errorf("CompletePersist failed: %v", err)
	}
	if fsm.GetState() != GraphReady {
		t.Errorf("Expected READY state, got %s", fsm.GetState())
	}
}

func TestGraphFSM_ErrorHandling(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	fsm := NewGraphFSM(logger)

	// Start build
	if err := fsm.StartBuild(); err != nil {
		t.Fatalf("StartBuild failed: %v", err)
	}

	// Test build failure
	testErr := fmt.Errorf("test error")
	if err := fsm.FailBuild(testErr); err != nil {
		t.Errorf("FailBuild failed: %v", err)
	}
	if fsm.GetState() != GraphError {
		t.Errorf("Expected ERROR state, got %s", fsm.GetState())
	}

	// Reset to idle
	baseGraphFSM := fsm.(*BaseGraphFSM)
	if err := baseGraphFSM.Reset(); err != nil {
		t.Errorf("Reset failed: %v", err)
	}
	if fsm.GetState() != GraphIdle {
		t.Errorf("Expected IDLE state after reset, got %s", fsm.GetState())
	}
}

func TestGraphFSM_EventEmission(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	fsm := NewGraphFSM(logger)

	// Track emitted events
	var receivedEvents []*Event
	fsm.SetEventHandler(func(event *Event) error {
		receivedEvents = append(receivedEvents, event)
		return nil
	})

	// Trigger some state transitions
	fsm.StartBuild()
	fsm.CompleteBuild()
	fsm.Clear()

	// Verify events were emitted
	if len(receivedEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedEvents))
	}

	expectedTypes := []EventType{
		EventGraphBuildStarted,
		EventGraphBuildCompleted,
		EventGraphCleared,
	}

	for i, event := range receivedEvents {
		if event.Type != expectedTypes[i] {
			t.Errorf("Event %d: expected type %s, got %s", i, expectedTypes[i], event.Type)
		}
	}
}

func TestAnalyzeFSM_Lifecycle(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	fsm := NewAnalyzeFSM(logger)

	// Initial state should be BOOTSTRAPPING
	if fsm.GetState() != AnalyzeBootstrapping {
		t.Errorf("Expected initial state BOOTSTRAPPING, got %s", fsm.GetState())
	}

	// Start service
	if err := fsm.Start(); err != nil {
		t.Errorf("Start failed: %v", err)
	}
	if fsm.GetState() != AnalyzeIdle {
		t.Errorf("Expected IDLE state, got %s", fsm.GetState())
	}

	// Pause service
	if err := fsm.Pause(); err != nil {
		t.Errorf("Pause failed: %v", err)
	}
	if fsm.GetState() != AnalyzePaused {
		t.Errorf("Expected PAUSED state, got %s", fsm.GetState())
	}

	// Resume service
	if err := fsm.Resume(); err != nil {
		t.Errorf("Resume failed: %v", err)
	}
	if fsm.GetState() != AnalyzeIdle {
		t.Errorf("Expected IDLE state, got %s", fsm.GetState())
	}

	// Stop service
	if err := fsm.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	if fsm.GetState() != AnalyzeTerminated {
		t.Errorf("Expected TERMINATED state, got %s", fsm.GetState())
	}
}

func TestAnalyzeFSM_GraphEventHandling(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	fsm := NewAnalyzeFSM(logger)

	// Start service
	fsm.Start()

	// Get the graph FSM
	graphFSM := fsm.GetGraphFSM()

	// Start a build operation (should transition AnalyzeFSM to PROCESSING)
	if err := graphFSM.StartBuild(); err != nil {
		t.Fatalf("StartBuild failed: %v", err)
	}

	if fsm.GetState() != AnalyzeProcessing {
		t.Errorf("Expected PROCESSING state, got %s", fsm.GetState())
	}

	// Complete the build (should transition back to IDLE)
	if err := graphFSM.CompleteBuild(); err != nil {
		t.Fatalf("CompleteBuild failed: %v", err)
	}

	if fsm.GetState() != AnalyzeIdle {
		t.Errorf("Expected IDLE state, got %s", fsm.GetState())
	}
}

func TestAnalyzeFSM_ResourceConstrained(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	fsm := NewAnalyzeFSM(logger)

	fsm.Start()
	
	// Simulate resource constraint while processing
	graphFSM := fsm.GetGraphFSM()
	graphFSM.StartBuild()

	if err := fsm.OnResourceConstrained("test constraint"); err != nil {
		t.Errorf("OnResourceConstrained failed: %v", err)
	}

	if fsm.GetState() != AnalyzePaused {
		t.Errorf("Expected PAUSED state after resource constraint, got %s", fsm.GetState())
	}
}
