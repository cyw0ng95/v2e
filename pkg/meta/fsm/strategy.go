package fsm

import (
	"fmt"
	"time"
)

// TransitionStrategy defines the interface for state transition strategies.
// This pattern allows different FSM types to share common transition logic
// while customizing specific behaviors (validation, logging, persistence).
//
// The strategy pattern reduces code duplication between BaseProviderFSM and
// MacroFSMManager by extracting common transition behaviors.
type TransitionStrategy interface {
	// Validate checks if a transition is valid
	Validate(from, to interface{}) error

	// CanTransition checks if a transition is allowed without modifying state
	CanTransition(from, to interface{}) bool
}

// ProviderTransitionStrategy implements transition logic for provider FSMs.
// It encapsulates validation rules for provider state transitions.
type ProviderTransitionStrategy struct {
	// Custom transitions map for provider-specific overrides
	customTransitions map[ProviderStateTransition]bool
}

// NewProviderTransitionStrategy creates a new provider transition strategy.
// Storage parameter is kept for API compatibility but persistence is handled
// inline in BaseProviderFSM.Transition for performance.
func NewProviderTransitionStrategy() *ProviderTransitionStrategy {
	return &ProviderTransitionStrategy{
		customTransitions: make(map[ProviderStateTransition]bool),
	}
}

// Validate checks if a provider state transition is valid.
// Returns an error if the transition is not allowed.
func (s *ProviderTransitionStrategy) Validate(from, to interface{}) error {
	fromState, ok := from.(ProviderState)
	if !ok {
		return fmt.Errorf("invalid from state type: %T", from)
	}

	toState, ok := to.(ProviderState)
	if !ok {
		return fmt.Errorf("invalid to state type: %T", to)
	}

	return ValidateProviderTransition(fromState, toState)
}

// CanTransition checks if a provider state transition is allowed.
// Returns true if the transition is valid, false otherwise.
func (s *ProviderTransitionStrategy) CanTransition(from, to interface{}) bool {
	return s.Validate(from, to) == nil
}

// AddCustomTransition adds a custom valid transition for this provider instance.
// This allows providers to extend the base transition rules.
func (s *ProviderTransitionStrategy) AddCustomTransition(from, to ProviderState) {
	s.customTransitions[ProviderStateTransition{From: from, To: to}] = true
}

// MacroTransitionStrategy implements transition logic for macro FSM.
// It encapsulates validation rules for macro state transitions.
type MacroTransitionStrategy struct {
	// Custom transitions map for macro-specific overrides
	customTransitions map[StateTransition]bool
}

// NewMacroTransitionStrategy creates a new macro transition strategy.
func NewMacroTransitionStrategy() *MacroTransitionStrategy {
	return &MacroTransitionStrategy{
		customTransitions: make(map[StateTransition]bool),
	}
}

// Validate checks if a macro state transition is valid.
// Returns an error if the transition is not allowed.
func (s *MacroTransitionStrategy) Validate(from, to interface{}) error {
	fromState, ok := from.(MacroState)
	if !ok {
		return fmt.Errorf("invalid from state type: %T", from)
	}

	toState, ok := to.(MacroState)
	if !ok {
		return fmt.Errorf("invalid to state type: %T", to)
	}

	return ValidateMacroTransition(fromState, toState)
}

// CanTransition checks if a macro state transition is allowed.
// Returns true if the transition is valid, false otherwise.
func (s *MacroTransitionStrategy) CanTransition(from, to interface{}) bool {
	return s.Validate(from, to) == nil
}

// AddCustomTransition adds a custom valid transition for this macro instance.
// This allows macros to extend the base transition rules.
func (s *MacroTransitionStrategy) AddCustomTransition(from, to MacroState) {
	s.customTransitions[StateTransition{From: from, To: to}] = true
}

// TransitionContext holds information about a state transition.
// This is used by transition strategies to make decisions about
// how to handle specific transitions.
type TransitionContext struct {
	// FSM is the FSM instance performing the transition
	FSM interface{}

	// OldState is the state before transition
	OldState interface{}

	// NewState is the target state
	NewState interface{}

	// Timestamp when the transition occurred
	Timestamp time.Time

	// Trigger is what caused the transition (e.g., "manual", "event", "quota_granted")
	Trigger string

	// Metadata for strategy-specific information
	Metadata map[string]interface{}
}

// NewTransitionContext creates a new transition context with current timestamp.
func NewTransitionContext(fsm, oldState, newState, trigger interface{}) *TransitionContext {
	return &TransitionContext{
		FSM:      fsm,
		OldState:  oldState,
		NewState:  newState,
		Trigger:    fmt.Sprintf("%v", trigger),
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
}
