package fsm

import (
	"fmt"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// BaseAnalyzeFSM provides the base implementation of AnalyzeFSM
type BaseAnalyzeFSM struct {
	mu       sync.RWMutex
	state    AnalyzeState
	graphFSM GraphFSM
	logger   *common.Logger
}

// NewAnalyzeFSM creates a new AnalyzeFSM instance
func NewAnalyzeFSM(logger *common.Logger) AnalyzeFSM {
	graphFSM := NewGraphFSM(logger)

	fsm := &BaseAnalyzeFSM{
		state:    AnalyzeBootstrapping,
		graphFSM: graphFSM,
		logger:   logger,
	}

	// Set up event handler for GraphFSM events
	graphFSM.SetEventHandler(fsm.HandleEvent)

	return fsm
}

// GetState returns the current analysis state
func (a *BaseAnalyzeFSM) GetState() AnalyzeState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Transition attempts to transition to a new state
func (a *BaseAnalyzeFSM) Transition(newState AnalyzeState) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := ValidateAnalyzeTransition(a.state, newState); err != nil {
		return err
	}

	oldState := a.state
	a.state = newState

	if a.logger != nil {
		a.logger.Info("AnalyzeFSM state transition: %s -> %s", oldState, newState)
	}

	return nil
}

// Start starts the analysis service
func (a *BaseAnalyzeFSM) Start() error {
	currentState := a.GetState()

	if currentState == AnalyzeBootstrapping {
		return a.Transition(AnalyzeIdle)
	}

	return nil
}

// Pause pauses the analysis service
func (a *BaseAnalyzeFSM) Pause() error {
	if err := a.Transition(AnalyzePaused); err != nil {
		return err
	}

	event := NewEvent(EventAnalysisPaused)
	return a.emitEvent(event)
}

// Resume resumes the analysis service
func (a *BaseAnalyzeFSM) Resume() error {
	currentState := a.GetState()
	if currentState != AnalyzePaused {
		return nil // Already running
	}

	if err := a.Transition(AnalyzeIdle); err != nil {
		return err
	}

	event := NewEvent(EventAnalysisResumed)
	return a.emitEvent(event)
}

// Stop stops the analysis service
func (a *BaseAnalyzeFSM) Stop() error {
	// First transition to draining
	if err := a.Transition(AnalyzeDraining); err != nil {
		return err
	}

	// Then transition to terminated
	return a.Transition(AnalyzeTerminated)
}

// HandleEvent processes an event from GraphFSM
func (a *BaseAnalyzeFSM) HandleEvent(event *Event) error {
	if a.logger != nil {
		a.logger.Debug("AnalyzeFSM received event: %s", event.Type)
	}

	// Handle state transitions based on graph events
	switch event.Type {
	case EventGraphBuildStarted:
		// Transition to processing when graph build starts
		if a.GetState() == AnalyzeIdle {
			return a.Transition(AnalyzeProcessing)
		}

	case EventGraphBuildCompleted, EventGraphAnalysisCompleted, EventGraphPersistCompleted:
		// Transition back to idle when operations complete
		if a.GetState() == AnalyzeProcessing {
			return a.Transition(AnalyzeIdle)
		}

	case EventGraphBuildFailed, EventGraphPersistFailed:
		// Remain in processing state on errors (allows retry)
		if a.logger != nil {
			if errMsg, ok := event.Data["error"].(string); ok {
				a.logger.Warn("Graph operation failed: %s", errMsg)
			}
		}

	case EventGraphCleared:
		// Transition to idle when graph is cleared
		if a.GetState() == AnalyzeProcessing {
			return a.Transition(AnalyzeIdle)
		}
	}

	return nil
}

// GetGraphFSM returns the managed GraphFSM
func (a *BaseAnalyzeFSM) GetGraphFSM() GraphFSM {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.graphFSM
}

// OnResourceConstrained handles resource constraint notifications
func (a *BaseAnalyzeFSM) OnResourceConstrained(reason string) error {
	if a.logger != nil {
		a.logger.Warn("Resource constraint detected: %s", reason)
	}

	// Pause the service if it's currently processing
	currentState := a.GetState()
	if currentState == AnalyzeProcessing {
		return a.Pause()
	}

	return nil
}

// emitEvent emits an event (can be extended for event bubbling to broker)
func (a *BaseAnalyzeFSM) emitEvent(event *Event) error {
	// This can be extended to notify broker/perf optimizer
	if a.logger != nil {
		a.logger.Debug("Emitting event: %s", event.Type)
	}
	return nil
}

// CanProcess returns true if the service can process new requests
func (a *BaseAnalyzeFSM) CanProcess() bool {
	state := a.GetState()
	return state == AnalyzeIdle || state == AnalyzeProcessing
}

// IsHealthy returns true if the service is in a healthy state
func (a *BaseAnalyzeFSM) IsHealthy() bool {
	state := a.GetState()
	return state != AnalyzeTerminated && state != AnalyzeDraining
}
