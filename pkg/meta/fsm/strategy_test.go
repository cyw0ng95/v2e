package fsm

import (
	"testing"
)

func TestNewProviderTransitionStrategy(t *testing.T) {
	strategy := NewProviderTransitionStrategy()

	if strategy == nil {
		t.Fatal("Expected non-nil strategy")
	}

	if strategy.customTransitions == nil {
		t.Error("Expected customTransitions map to be initialized")
	}
}

func TestProviderTransitionStrategy_Validate(t *testing.T) {
	strategy := NewProviderTransitionStrategy()

	tests := []struct {
		name      string
		fromState ProviderState
		toState   ProviderState
		wantErr   bool
	}{
		{
			name:      "Valid: IDLE to ACQUIRING",
			fromState: ProviderIdle,
			toState:   ProviderAcquiring,
			wantErr:   false,
		},
		{
			name:      "Valid: ACQUIRING to RUNNING",
			fromState: ProviderAcquiring,
			toState:   ProviderRunning,
			wantErr:   false,
		},
		{
			name:      "Valid: RUNNING to PAUSED",
			fromState: ProviderRunning,
			toState:   ProviderPaused,
			wantErr:   false,
		},
		{
			name:      "Invalid: IDLE to RUNNING",
			fromState: ProviderIdle,
			toState:   ProviderRunning,
			wantErr:   true,
		},
		{
			name:      "Same state: RUNNING to RUNNING",
			fromState: ProviderRunning,
			toState:   ProviderRunning,
			wantErr:   false, // Same state is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.Validate(tt.fromState, tt.toState)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProviderTransitionStrategy_CanTransition(t *testing.T) {
	strategy := NewProviderTransitionStrategy()

	// Valid transition
	if !strategy.CanTransition(ProviderIdle, ProviderAcquiring) {
		t.Error("Expected CanTransition to return true for IDLE -> ACQUIRING")
	}

	// Invalid transition
	if strategy.CanTransition(ProviderIdle, ProviderRunning) {
		t.Error("Expected CanTransition to return false for IDLE -> RUNNING")
	}

	// Same state is valid
	if !strategy.CanTransition(ProviderIdle, ProviderIdle) {
		t.Error("Expected CanTransition to return true for same state")
	}
}

func TestProviderTransitionStrategy_AddCustomTransition(t *testing.T) {
	strategy := NewProviderTransitionStrategy()

	// Add a custom transition that's not normally valid
	strategy.AddCustomTransition(ProviderIdle, ProviderRunning)

	// Custom transitions are stored in the map for extensibility
	// The base ValidateProviderTransition still checks base rules
	// Custom transitions can be used by provider implementations
	// to extend the base validation logic
	if len(strategy.customTransitions) != 1 {
		t.Errorf("Expected 1 custom transition, got %d", len(strategy.customTransitions))
	}

	// Check that the custom transition was stored correctly
	key := ProviderStateTransition{From: ProviderIdle, To: ProviderRunning}
	if !strategy.customTransitions[key] {
		t.Error("Expected custom transition to be stored in map")
	}
}

func TestProviderTransitionStrategy_ValidateInvalidTypes(t *testing.T) {
	strategy := NewProviderTransitionStrategy()

	// Test with invalid type for from state
	err := strategy.Validate("invalid", ProviderIdle)
	if err == nil {
		t.Error("Expected error for invalid from state type")
	}

	// Test with invalid type for to state
	err = strategy.Validate(ProviderIdle, 123)
	if err == nil {
		t.Error("Expected error for invalid to state type")
	}
}

func TestNewMacroTransitionStrategy(t *testing.T) {
	strategy := NewMacroTransitionStrategy()

	if strategy == nil {
		t.Fatal("Expected non-nil strategy")
	}

	if strategy.customTransitions == nil {
		t.Error("Expected customTransitions map to be initialized")
	}
}

func TestMacroTransitionStrategy_Validate(t *testing.T) {
	strategy := NewMacroTransitionStrategy()

	tests := []struct {
		name      string
		fromState MacroState
		toState   MacroState
		wantErr   bool
	}{
		{
			name:      "Valid: BOOTSTRAPPING to ORCHESTRATING",
			fromState: MacroBootstrapping,
			toState:   MacroOrchestrating,
			wantErr:   false,
		},
		{
			name:      "Valid: ORCHESTRATING to STABILIZING",
			fromState: MacroOrchestrating,
			toState:   MacroStabilizing,
			wantErr:   false,
		},
		{
			name:      "Valid: STABILIZING to DRAINING",
			fromState: MacroStabilizing,
			toState:   MacroDraining,
			wantErr:   false,
		},
		{
			name:      "Invalid: DRAINING to BOOTSTRAPPING",
			fromState: MacroDraining,
			toState:   MacroBootstrapping,
			wantErr:   true,
		},
		{
			name:      "Same state: ORCHESTRATING to ORCHESTRATING",
			fromState: MacroOrchestrating,
			toState:   MacroOrchestrating,
			wantErr:   false, // Same state is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.Validate(tt.fromState, tt.toState)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMacroTransitionStrategy_CanTransition(t *testing.T) {
	strategy := NewMacroTransitionStrategy()

	// Valid transition
	if !strategy.CanTransition(MacroBootstrapping, MacroOrchestrating) {
		t.Error("Expected CanTransition to return true for BOOTSTRAPPING -> ORCHESTRATING")
	}

	// Invalid transition
	if strategy.CanTransition(MacroDraining, MacroBootstrapping) {
		t.Error("Expected CanTransition to return false for DRAINING -> BOOTSTRAPPING")
	}

	// Same state is valid
	if !strategy.CanTransition(MacroBootstrapping, MacroBootstrapping) {
		t.Error("Expected CanTransition to return true for same state")
	}
}

func TestMacroTransitionStrategy_AddCustomTransition(t *testing.T) {
	strategy := NewMacroTransitionStrategy()

	// Add a custom transition that's not normally valid
	strategy.AddCustomTransition(MacroDraining, MacroBootstrapping)

	// Now it should be in custom transitions
	// (Note: Validate still checks base rules, custom is for extensibility)
	if len(strategy.customTransitions) != 1 {
		t.Errorf("Expected 1 custom transition, got %d", len(strategy.customTransitions))
	}
}

func TestMacroTransitionStrategy_ValidateInvalidTypes(t *testing.T) {
	strategy := NewMacroTransitionStrategy()

	// Test with invalid type for from state
	err := strategy.Validate("invalid", MacroBootstrapping)
	if err == nil {
		t.Error("Expected error for invalid from state type")
	}

	// Test with invalid type for to state
	err = strategy.Validate(MacroBootstrapping, 123)
	if err == nil {
		t.Error("Expected error for invalid to state type")
	}
}

func TestNewTransitionContext(t *testing.T) {
	fsm := "test-fsm"
	oldState := ProviderIdle
	newState := ProviderAcquiring
	trigger := "manual"

	ctx := NewTransitionContext(fsm, oldState, newState, trigger)

	if ctx.FSM != fsm {
		t.Errorf("FSM = %v, want %v", ctx.FSM, fsm)
	}

	if ctx.OldState != oldState {
		t.Errorf("OldState = %v, want %v", ctx.OldState, oldState)
	}

	if ctx.NewState != newState {
		t.Errorf("NewState = %v, want %v", ctx.NewState, newState)
	}

	if ctx.Trigger != "manual" {
		t.Errorf("Trigger = %v, want manual", ctx.Trigger)
	}

	if ctx.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if ctx.Metadata == nil {
		t.Error("Expected Metadata map to be initialized")
	}
}
