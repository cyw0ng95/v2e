package fsm

import (
	"context"
	"fmt"
	"time"
)

// MacroState represents the high-level orchestration state
type MacroState string

const (
	// MacroBootstrapping - Initial state, setting up resources and loading state
	MacroBootstrapping MacroState = "BOOTSTRAPPING"
	// MacroOrchestrating - Active state, coordinating provider FSMs
	MacroOrchestrating MacroState = "ORCHESTRATING"
	// MacroStabilizing - Winding down, waiting for providers to finish
	MacroStabilizing MacroState = "STABILIZING"
	// MacroDraining - Final state, cleaning up resources
	MacroDraining MacroState = "DRAINING"
)

// ProviderState represents the state of a provider FSM
type ProviderState string

const (
	// ProviderIdle - Not executing, waiting for work
	ProviderIdle ProviderState = "IDLE"
	// ProviderAcquiring - Requesting permits from broker
	ProviderAcquiring ProviderState = "ACQUIRING"
	// ProviderRunning - Actively processing items
	ProviderRunning ProviderState = "RUNNING"
	// ProviderWaitingQuota - Paused due to permit revocation by broker
	ProviderWaitingQuota ProviderState = "WAITING_QUOTA"
	// ProviderWaitingBackoff - Paused due to rate limiting (429 errors)
	ProviderWaitingBackoff ProviderState = "WAITING_BACKOFF"
	// ProviderPaused - Paused by user request
	ProviderPaused ProviderState = "PAUSED"
	// ProviderTerminated - Completed or stopped
	ProviderTerminated ProviderState = "TERMINATED"
)

// EventType represents the type of FSM event
type EventType string

const (
	// EventProviderStarted - Provider has started processing
	EventProviderStarted EventType = "PROVIDER_STARTED"
	// EventProviderCompleted - Provider has completed successfully
	EventProviderCompleted EventType = "PROVIDER_COMPLETED"
	// EventProviderFailed - Provider has failed
	EventProviderFailed EventType = "PROVIDER_FAILED"
	// EventProviderPaused - Provider was paused
	EventProviderPaused EventType = "PROVIDER_PAUSED"
	// EventProviderResumed - Provider was resumed
	EventProviderResumed EventType = "PROVIDER_RESUMED"
	// EventQuotaRevoked - Broker revoked permits
	EventQuotaRevoked EventType = "QUOTA_REVOKED"
	// EventQuotaGranted - Broker granted permits
	EventQuotaGranted EventType = "QUOTA_GRANTED"
	// EventRateLimited - Provider hit rate limit
	EventRateLimited EventType = "RATE_LIMITED"
	// EventCheckpoint - Provider reached a checkpoint
	EventCheckpoint EventType = "CHECKPOINT"
)

// Event represents an FSM state transition event
type Event struct {
	Type       EventType              `json:"type"`
	ProviderID string                 `json:"provider_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// NewEvent creates a new FSM event
func NewEvent(eventType EventType, providerID string) *Event {
	return &Event{
		Type:       eventType,
		ProviderID: providerID,
		Timestamp:  time.Now(),
		Data:       make(map[string]interface{}),
	}
}

// MacroFSM defines the interface for the high-level orchestration state machine
type MacroFSM interface {
	// GetState returns the current macro state
	GetState() MacroState

	// Transition attempts to transition to a new state
	Transition(newState MacroState) error

	// HandleEvent processes an event from a provider FSM
	HandleEvent(event *Event) error

	// GetProviders returns all managed provider FSMs
	GetProviders() []ProviderFSM

	// AddProvider adds a provider FSM to be managed
	AddProvider(provider ProviderFSM) error

	// RemoveProvider removes a provider FSM
	RemoveProvider(providerID string) error
}

// ProviderFSM defines the interface for provider-specific state machines
type ProviderFSM interface {
	// GetID returns the unique provider identifier
	GetID() string

	// GetType returns the provider type (cve, cwe, capec, attack)
	GetType() string

	// GetState returns the current provider state
	GetState() ProviderState

	// Transition attempts to transition to a new state
	Transition(newState ProviderState) error

	// Start begins execution (IDLE -> ACQUIRING -> RUNNING)
	Start() error

	// Pause pauses execution (RUNNING -> PAUSED)
	Pause() error

	// Resume resumes execution (PAUSED -> ACQUIRING -> RUNNING)
	Resume() error

	// Stop terminates execution (any state -> TERMINATED)
	Stop() error

	// OnQuotaRevoked handles quota revocation from broker
	OnQuotaRevoked(revokedCount int) error

	// OnQuotaGranted handles quota grant from broker
	OnQuotaGranted(grantedCount int) error

	// OnRateLimited handles rate limiting (429 errors)
	OnRateLimited(retryAfter time.Duration) error

	// Execute performs the actual work (called when in RUNNING state)
	Execute() error

	// SetEventHandler sets the callback for event bubbling to MacroFSM
	SetEventHandler(handler func(*Event) error)

	// Initialize sets up provider context before starting
	Initialize(ctx context.Context) error

	// GetStats returns provider statistics for monitoring
	GetStats() map[string]interface{}

	// GetDependencies returns the list of provider IDs that must complete before this provider
	GetDependencies() []string
}

// StateTransition represents a valid state transition
type StateTransition struct {
	From MacroState
	To   MacroState
}

// ProviderStateTransition represents a valid provider state transition
type ProviderStateTransition struct {
	From ProviderState
	To   ProviderState
}

// Macro FSM valid transitions
var validMacroTransitions = map[StateTransition]bool{
	{MacroBootstrapping, MacroOrchestrating}: true,
	{MacroBootstrapping, MacroDraining}:      true, // Emergency drain during bootstrap
	{MacroOrchestrating, MacroStabilizing}:   true,
	{MacroOrchestrating, MacroDraining}:      true, // Emergency drain
	{MacroStabilizing, MacroDraining}:        true,
	{MacroStabilizing, MacroOrchestrating}:   true, // Restart
}

// Provider FSM valid transitions
var validProviderTransitions = map[ProviderStateTransition]bool{
	{ProviderIdle, ProviderAcquiring}:            true,
	{ProviderAcquiring, ProviderRunning}:         true,
	{ProviderAcquiring, ProviderPaused}:          true, // Pause during acquisition
	{ProviderAcquiring, ProviderTerminated}:      true, // Stop during acquisition
	{ProviderRunning, ProviderWaitingQuota}:      true,
	{ProviderRunning, ProviderWaitingBackoff}:    true,
	{ProviderRunning, ProviderPaused}:            true,
	{ProviderRunning, ProviderTerminated}:        true,
	{ProviderWaitingQuota, ProviderAcquiring}:    true, // Retry acquisition
	{ProviderWaitingQuota, ProviderTerminated}:   true,
	{ProviderWaitingBackoff, ProviderAcquiring}:  true, // Retry after backoff
	{ProviderWaitingBackoff, ProviderTerminated}: true,
	{ProviderPaused, ProviderAcquiring}:          true, // Resume
	{ProviderPaused, ProviderTerminated}:         true,
}

// ValidateMacroTransition checks if a macro state transition is valid
func ValidateMacroTransition(from, to MacroState) error {
	if from == to {
		return nil // Same state is always valid
	}

	transition := StateTransition{From: from, To: to}
	if !validMacroTransitions[transition] {
		return fmt.Errorf("invalid macro state transition: %s -> %s", from, to)
	}

	return nil
}

// ValidateProviderTransition checks if a provider state transition is valid
func ValidateProviderTransition(from, to ProviderState) error {
	if from == to {
		return nil // Same state is always valid
	}

	transition := ProviderStateTransition{From: from, To: to}
	if !validProviderTransitions[transition] {
		return fmt.Errorf("invalid provider state transition: %s -> %s", from, to)
	}

	return nil
}
