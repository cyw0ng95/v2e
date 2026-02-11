package fsm

import (
	"errors"
	"fmt"
	"time"
)

// GraphState represents the state of the graph analysis FSM
type GraphState string

const (
	// GraphIdle - Graph is idle, not processing
	GraphIdle GraphState = "IDLE"
	// GraphBuilding - Graph is being built from data sources
	GraphBuilding GraphState = "BUILDING"
	// GraphAnalyzing - Graph is being analyzed (path finding, metrics, etc.)
	GraphAnalyzing GraphState = "ANALYZING"
	// GraphPersisting - Graph is being persisted to disk
	GraphPersisting GraphState = "PERSISTING"
	// GraphReady - Graph is ready for queries
	GraphReady GraphState = "READY"
	// GraphError - Graph encountered an error
	GraphError GraphState = "ERROR"
)

// AnalyzeState represents the state of the analysis service FSM
type AnalyzeState string

const (
	// AnalyzeBootstrapping - Initial state, setting up resources
	AnalyzeBootstrapping AnalyzeState = "BOOTSTRAPPING"
	// AnalyzeIdle - Service is idle, waiting for requests
	AnalyzeIdle AnalyzeState = "IDLE"
	// AnalyzeProcessing - Service is processing analysis requests
	AnalyzeProcessing AnalyzeState = "PROCESSING"
	// AnalyzePaused - Service is paused by user or broker
	AnalyzePaused AnalyzeState = "PAUSED"
	// AnalyzeDraining - Service is shutting down
	AnalyzeDraining AnalyzeState = "DRAINING"
	// AnalyzeTerminated - Service has terminated
	AnalyzeTerminated AnalyzeState = "TERMINATED"
)

// EventType represents the type of FSM event
type EventType string

const (
	// EventGraphBuildStarted - Graph build has started
	EventGraphBuildStarted EventType = "GRAPH_BUILD_STARTED"
	// EventGraphBuildCompleted - Graph build has completed
	EventGraphBuildCompleted EventType = "GRAPH_BUILD_COMPLETED"
	// EventGraphBuildFailed - Graph build has failed
	EventGraphBuildFailed EventType = "GRAPH_BUILD_FAILED"
	// EventGraphAnalysisStarted - Graph analysis has started
	EventGraphAnalysisStarted EventType = "GRAPH_ANALYSIS_STARTED"
	// EventGraphAnalysisCompleted - Graph analysis has completed
	EventGraphAnalysisCompleted EventType = "GRAPH_ANALYSIS_COMPLETED"
	// EventGraphAnalysisFailed - Graph analysis has failed
	EventGraphAnalysisFailed EventType = "GRAPH_ANALYSIS_FAILED"
	// EventGraphPersistStarted - Graph persistence has started
	EventGraphPersistStarted EventType = "GRAPH_PERSIST_STARTED"
	// EventGraphPersistCompleted - Graph persistence has completed
	EventGraphPersistCompleted EventType = "GRAPH_PERSIST_COMPLETED"
	// EventGraphPersistFailed - Graph persistence has failed
	EventGraphPersistFailed EventType = "GRAPH_PERSIST_FAILED"
	// EventGraphCleared - Graph has been cleared
	EventGraphCleared EventType = "GRAPH_CLEARED"
	// EventAnalysisPaused - Analysis service paused
	EventAnalysisPaused EventType = "ANALYSIS_PAUSED"
	// EventAnalysisResumed - Analysis service resumed
	EventAnalysisResumed EventType = "ANALYSIS_RESUMED"
	// EventResourceConstrained - Resource constraint detected
	EventResourceConstrained EventType = "RESOURCE_CONSTRAINED"
)

// Event represents an FSM state transition event
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewEvent creates a new FSM event
func NewEvent(eventType EventType) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}
}

// GraphFSM defines the interface for graph state machine
type GraphFSM interface {
	// GetState returns the current graph state
	GetState() GraphState

	// Transition attempts to transition to a new state
	Transition(newState GraphState) error

	// StartBuild initiates graph building
	StartBuild() error

	// CompleteBuild marks graph building as complete
	CompleteBuild() error

	// FailBuild marks graph building as failed
	FailBuild(err error) error

	// StartAnalysis initiates graph analysis
	StartAnalysis() error

	// CompleteAnalysis marks analysis as complete
	CompleteAnalysis() error

	// FailAnalysis marks analysis as failed
	FailAnalysis(err error) error

	// StartPersist initiates graph persistence
	StartPersist() error

	// CompletePersist marks persistence as complete
	CompletePersist() error

	// FailPersist marks persistence as failed
	FailPersist(err error) error

	// Clear clears the graph
	Clear() error

	// SetEventHandler sets the callback for event bubbling
	SetEventHandler(handler func(*Event) error)
}

// AnalyzeFSM defines the interface for analysis service state machine
type AnalyzeFSM interface {
	// GetState returns the current analysis state
	GetState() AnalyzeState

	// Transition attempts to transition to a new state
	Transition(newState AnalyzeState) error

	// Start starts the analysis service
	Start() error

	// Pause pauses the analysis service
	Pause() error

	// Resume resumes the analysis service
	Resume() error

	// Stop stops the analysis service
	Stop() error

	// HandleEvent processes an event from GraphFSM
	HandleEvent(event *Event) error

	// GetGraphFSM returns the managed GraphFSM
	GetGraphFSM() GraphFSM

	// OnResourceConstrained handles resource constraint notifications
	OnResourceConstrained(reason string) error
}

// GraphStateTransition represents a valid graph state transition
type GraphStateTransition struct {
	From GraphState
	To   GraphState
}

// AnalyzeStateTransition represents a valid analyze state transition
type AnalyzeStateTransition struct {
	From AnalyzeState
	To   AnalyzeState
}

// Valid graph state transitions
var validGraphTransitions = map[GraphStateTransition]bool{
	{GraphIdle, GraphBuilding}:    true,
	{GraphBuilding, GraphReady}:   true,
	{GraphBuilding, GraphError}:   true,
	{GraphReady, GraphAnalyzing}:  true,
	{GraphReady, GraphPersisting}: true,
	{GraphReady, GraphBuilding}:   true, // Rebuild
	{GraphReady, GraphIdle}:       true, // Clear
	{GraphAnalyzing, GraphReady}:  true,
	{GraphAnalyzing, GraphError}:  true,
	{GraphPersisting, GraphReady}: true,
	{GraphPersisting, GraphError}: true,
	{GraphError, GraphIdle}:       true, // Reset after error
	{GraphError, GraphBuilding}:   true, // Retry after error
}

// Valid analysis state transitions
var validAnalyzeTransitions = map[AnalyzeStateTransition]bool{
	{AnalyzeBootstrapping, AnalyzeIdle}:       true,
	{AnalyzeBootstrapping, AnalyzeTerminated}: true, // Emergency shutdown
	{AnalyzeIdle, AnalyzeProcessing}:          true,
	{AnalyzeIdle, AnalyzePaused}:              true,
	{AnalyzeIdle, AnalyzeDraining}:            true,
	{AnalyzeProcessing, AnalyzeIdle}:          true,
	{AnalyzeProcessing, AnalyzePaused}:        true,
	{AnalyzeProcessing, AnalyzeDraining}:      true,
	{AnalyzePaused, AnalyzeIdle}:              true, // Resume
	{AnalyzePaused, AnalyzeProcessing}:        true, // Resume with work
	{AnalyzePaused, AnalyzeDraining}:          true,
	{AnalyzeDraining, AnalyzeTerminated}:      true,
}

// ValidateGraphTransition checks if a graph state transition is valid
func ValidateGraphTransition(from, to GraphState) error {
	if from == to {
		return nil // Same state is always valid
	}

	transition := GraphStateTransition{From: from, To: to}
	if !validGraphTransitions[transition] {
		return fmt.Errorf("invalid graph state transition: %s -> %s", from, to)
	}

	return nil
}

// ValidateAnalyzeTransition checks if an analyze state transition is valid
func ValidateAnalyzeTransition(from, to AnalyzeState) error {
	if from == to {
		return nil // Same state is always valid
	}

	transition := AnalyzeStateTransition{From: from, To: to}
	if !validAnalyzeTransitions[transition] {
		return fmt.Errorf("invalid analyze state transition: %s -> %s", from, to)
	}

	return nil
}

// FSMErrorType represents the category of FSM error
type FSMErrorType string

const (
	// ErrorTypeTransient indicates a temporary error that can be retried
	ErrorTypeTransient FSMErrorType = "TRANSIENT"
	// ErrorTypePermanent indicates a permanent error that should not be retried
	ErrorTypePermanent FSMErrorType = "PERMANENT"
)

// FSMError represents an error that occurs during FSM operation
type FSMError struct {
	// Type is the error category
	Type FSMErrorType
	// State is the FSM state when the error occurred
	State string
	// Err is the underlying error
	Err error
	// RetryCount is the number of retry attempts made
	RetryCount int
	// Message provides additional context
	Message string
}

// Error returns the error message
func (e *FSMError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] state=%s: %v", e.Type, e.State, e.Err)
}

// Unwrap returns the underlying error
func (e *FSMError) Unwrap() error {
	return e.Err
}

// IsTransient returns true if this is a transient error
func (e *FSMError) IsTransient() bool {
	return e.Type == ErrorTypeTransient
}

// IsPermanent returns true if this is a permanent error
func (e *FSMError) IsPermanent() bool {
	return e.Type == ErrorTypePermanent
}

// IsFSMError checks if an error is an FSMError
func IsFSMError(err error) bool {
	var fsmErr *FSMError
	return errors.As(err, &fsmErr)
}

// NewTransientError creates a new transient FSM error
func NewTransientError(state string, err error) *FSMError {
	return &FSMError{
		Type:    ErrorTypeTransient,
		State:   state,
		Err:     err,
		Message: "",
	}
}

// NewPermanentError creates a new permanent FSM error
func NewPermanentError(state string, err error) *FSMError {
	return &FSMError{
		Type:    ErrorTypePermanent,
		State:   state,
		Err:     err,
		Message: "",
	}
}

// RetryConfig defines retry behavior for transient errors
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// BaseDelay is the initial delay before first retry
	BaseDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// BackoffFactor is the multiplier for exponential backoff
	BackoffFactor float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		BaseDelay:      100 * time.Millisecond,
		MaxDelay:       30 * time.Second,
		BackoffFactor: 2.0,
	}
}
