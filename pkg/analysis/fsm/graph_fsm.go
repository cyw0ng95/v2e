package fsm

import (
	"fmt"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// BaseGraphFSM provides the base implementation of GraphFSM
type BaseGraphFSM struct {
	mu           sync.RWMutex
	state        GraphState
	eventHandler func(*Event) error
	logger       *common.Logger
	lastError    error
}

// NewGraphFSM creates a new GraphFSM instance
func NewGraphFSM(logger *common.Logger) GraphFSM {
	return &BaseGraphFSM{
		state:  GraphIdle,
		logger: logger,
	}
}

// GetState returns the current graph state
func (g *BaseGraphFSM) GetState() GraphState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state
}

// Transition attempts to transition to a new state
func (g *BaseGraphFSM) Transition(newState GraphState) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if err := ValidateGraphTransition(g.state, newState); err != nil {
		return err
	}

	oldState := g.state
	g.state = newState

	if g.logger != nil {
		g.logger.Info("GraphFSM state transition: %s -> %s", oldState, newState)
	}

	return nil
}

// StartBuild initiates graph building
func (g *BaseGraphFSM) StartBuild() error {
	if err := g.Transition(GraphBuilding); err != nil {
		return err
	}

	event := NewEvent(EventGraphBuildStarted)
	return g.emitEvent(event)
}

// CompleteBuild marks graph building as complete
func (g *BaseGraphFSM) CompleteBuild() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphReady); err != nil {
		return err
	}

	event := NewEvent(EventGraphBuildCompleted)
	return g.emitEvent(event)
}

// FailBuild marks graph building as failed
func (g *BaseGraphFSM) FailBuild(err error) error {
	g.mu.Lock()
	g.lastError = err
	g.mu.Unlock()

	if transErr := g.Transition(GraphError); transErr != nil {
		return transErr
	}

	event := NewEvent(EventGraphBuildFailed)
	event.Data["error"] = err.Error()
	return g.emitEvent(event)
}

// StartAnalysis initiates graph analysis
func (g *BaseGraphFSM) StartAnalysis() error {
	if err := g.Transition(GraphAnalyzing); err != nil {
		return err
	}

	event := NewEvent(EventGraphAnalysisStarted)
	return g.emitEvent(event)
}

// CompleteAnalysis marks analysis as complete
func (g *BaseGraphFSM) CompleteAnalysis() error {
	if err := g.Transition(GraphReady); err != nil {
		return err
	}

	event := NewEvent(EventGraphAnalysisCompleted)
	return g.emitEvent(event)
}

// StartPersist initiates graph persistence
func (g *BaseGraphFSM) StartPersist() error {
	if err := g.Transition(GraphPersisting); err != nil {
		return err
	}

	event := NewEvent(EventGraphPersistStarted)
	return g.emitEvent(event)
}

// CompletePersist marks persistence as complete
func (g *BaseGraphFSM) CompletePersist() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphReady); err != nil {
		return err
	}

	event := NewEvent(EventGraphPersistCompleted)
	return g.emitEvent(event)
}

// FailPersist marks persistence as failed
func (g *BaseGraphFSM) FailPersist(err error) error {
	g.mu.Lock()
	g.lastError = err
	g.mu.Unlock()

	if transErr := g.Transition(GraphError); transErr != nil {
		return transErr
	}

	event := NewEvent(EventGraphPersistFailed)
	event.Data["error"] = err.Error()
	return g.emitEvent(event)
}

// Clear clears the graph
func (g *BaseGraphFSM) Clear() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	if err := g.Transition(GraphIdle); err != nil {
		return err
	}

	event := NewEvent(EventGraphCleared)
	return g.emitEvent(event)
}

// SetEventHandler sets the callback for event bubbling
func (g *BaseGraphFSM) SetEventHandler(handler func(*Event) error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.eventHandler = handler
}

// emitEvent emits an event to the parent FSM
func (g *BaseGraphFSM) emitEvent(event *Event) error {
	g.mu.RLock()
	handler := g.eventHandler
	g.mu.RUnlock()

	if handler != nil {
		return handler(event)
	}

	return nil
}

// GetLastError returns the last error encountered
func (g *BaseGraphFSM) GetLastError() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.lastError
}

// Reset resets the FSM to idle state (used for error recovery)
func (g *BaseGraphFSM) Reset() error {
	g.mu.Lock()
	g.lastError = nil
	g.mu.Unlock()

	currentState := g.GetState()
	if currentState == GraphError {
		return g.Transition(GraphIdle)
	}

	return fmt.Errorf("cannot reset from state: %s", currentState)
}
