package fsm

import (
	"context"
	"testing"
	"time"
)

// MockProviderFSM is a mock implementation of ProviderFSM for testing
type MockProviderFSM struct {
	id           string
	providerType string
	state        ProviderState
	eventHandler func(*Event) error
}

func NewMockProviderFSM(id, providerType string) *MockProviderFSM {
	return &MockProviderFSM{
		id:           id,
		providerType: providerType,
		state:        ProviderIdle,
	}
}

func (m *MockProviderFSM) GetID() string                                { return m.id }
func (m *MockProviderFSM) GetType() string                              { return m.providerType }
func (m *MockProviderFSM) GetState() ProviderState                      { return m.state }
func (m *MockProviderFSM) Transition(newState ProviderState) error      { m.state = newState; return nil }
func (m *MockProviderFSM) Start() error                                 { return m.Transition(ProviderRunning) }
func (m *MockProviderFSM) Pause() error                                 { return m.Transition(ProviderPaused) }
func (m *MockProviderFSM) Resume() error                                { return m.Transition(ProviderRunning) }
func (m *MockProviderFSM) Stop() error                                  { return m.Transition(ProviderTerminated) }
func (m *MockProviderFSM) OnQuotaRevoked(count int) error               { return nil }
func (m *MockProviderFSM) OnQuotaGranted(count int) error               { return nil }
func (m *MockProviderFSM) OnRateLimited(retryAfter time.Duration) error { return nil }
func (m *MockProviderFSM) Execute() error                               { return nil }
func (m *MockProviderFSM) SetEventHandler(handler func(*Event) error)   { m.eventHandler = handler }
func (m *MockProviderFSM) Initialize(ctx context.Context) error          { return nil }
func (m *MockProviderFSM) GetStats() map[string]interface{}            { return map[string]interface{}{"id": m.id, "type": m.providerType} }
func (m *MockProviderFSM) GetDependencies() []string                    { return nil }

func (m *MockProviderFSM) EmitEvent(eventType EventType) error {
	if m.eventHandler == nil {
		return nil
	}
	event := NewEvent(eventType, m.id)
	return m.eventHandler(event)
}

func TestNewMacroFSMManager(t *testing.T) {
	manager, err := NewMacroFSMManager("test-macro", nil)
	if err != nil {
		t.Fatalf("Failed to create MacroFSMManager: %v", err)
	}
	defer manager.Stop()

	if manager.GetState() != MacroBootstrapping {
		t.Errorf("Initial state = %v, want %v", manager.GetState(), MacroBootstrapping)
	}

	if manager.GetID() != "test-macro" {
		t.Errorf("ID = %v, want test-macro", manager.GetID())
	}
}

func TestNewMacroFSMManager_EmptyID(t *testing.T) {
	_, err := NewMacroFSMManager("", nil)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestMacroFSMManager_Transition(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial state
			manager.mu.Lock()
			manager.state = tt.fromState
			manager.mu.Unlock()

			err := manager.Transition(tt.toState)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transition() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && manager.GetState() != tt.toState {
				t.Errorf("State after transition = %v, want %v", manager.GetState(), tt.toState)
			}
		})
	}
}

func TestMacroFSMManager_AddProvider(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	provider := NewMockProviderFSM("provider-1", "cve")

	err := manager.AddProvider(provider)
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}

	if manager.GetProviderCount() != 1 {
		t.Errorf("Provider count = %d, want 1", manager.GetProviderCount())
	}

	// Try adding same provider again
	err = manager.AddProvider(provider)
	if err == nil {
		t.Error("Expected error when adding duplicate provider, got nil")
	}
}

func TestMacroFSMManager_AddProvider_Nil(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	err := manager.AddProvider(nil)
	if err == nil {
		t.Error("Expected error for nil provider, got nil")
	}
}

func TestMacroFSMManager_RemoveProvider(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	provider := NewMockProviderFSM("provider-1", "cve")
	manager.AddProvider(provider)

	err := manager.RemoveProvider("provider-1")
	if err != nil {
		t.Fatalf("Failed to remove provider: %v", err)
	}

	if manager.GetProviderCount() != 0 {
		t.Errorf("Provider count = %d, want 0", manager.GetProviderCount())
	}

	// Try removing non-existent provider
	err = manager.RemoveProvider("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent provider, got nil")
	}
}

func TestMacroFSMManager_GetProvider(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	provider := NewMockProviderFSM("provider-1", "cve")
	manager.AddProvider(provider)

	retrieved, err := manager.GetProvider("provider-1")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}

	if retrieved.GetID() != "provider-1" {
		t.Errorf("Retrieved provider ID = %v, want provider-1", retrieved.GetID())
	}

	// Try getting non-existent provider
	_, err = manager.GetProvider("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent provider, got nil")
	}
}

func TestMacroFSMManager_GetProviders(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	provider1 := NewMockProviderFSM("provider-1", "cve")
	provider2 := NewMockProviderFSM("provider-2", "cwe")

	manager.AddProvider(provider1)
	manager.AddProvider(provider2)

	providers := manager.GetProviders()
	if len(providers) != 2 {
		t.Errorf("Providers count = %d, want 2", len(providers))
	}
}

func TestMacroFSMManager_HandleEvent(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	event := NewEvent(EventProviderStarted, "provider-1")

	err := manager.HandleEvent(event)
	if err != nil {
		t.Fatalf("Failed to handle event: %v", err)
	}

	// Wait a bit for async processing
	time.Sleep(200 * time.Millisecond)

	// After EventProviderStarted, should transition to ORCHESTRATING
	if manager.GetState() != MacroOrchestrating {
		t.Errorf("State after EventProviderStarted = %v, want %v",
			manager.GetState(), MacroOrchestrating)
	}
}

func TestMacroFSMManager_HandleEvent_Nil(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	err := manager.HandleEvent(nil)
	if err == nil {
		t.Error("Expected error for nil event, got nil")
	}
}

func TestMacroFSMManager_EventBubbling(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	provider := NewMockProviderFSM("provider-1", "cve")
	manager.AddProvider(provider)

	// Emit event from provider
	err := provider.EmitEvent(EventProviderStarted)
	if err != nil {
		t.Fatalf("Failed to emit event: %v", err)
	}

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Macro should have transitioned to ORCHESTRATING
	if manager.GetState() != MacroOrchestrating {
		t.Errorf("State = %v, want %v", manager.GetState(), MacroOrchestrating)
	}
}

func TestMacroFSMManager_AllProvidersCompleted(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	// Start in orchestrating state
	manager.Transition(MacroOrchestrating)

	provider1 := NewMockProviderFSM("provider-1", "cve")
	provider2 := NewMockProviderFSM("provider-2", "cwe")

	manager.AddProvider(provider1)
	manager.AddProvider(provider2)

	// Mark both as terminated
	provider1.Transition(ProviderTerminated)
	provider2.Transition(ProviderTerminated)

	// Emit completion event
	provider1.EmitEvent(EventProviderCompleted)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Should transition to STABILIZING when all providers complete
	if manager.GetState() != MacroStabilizing {
		t.Errorf("State = %v, want %v", manager.GetState(), MacroStabilizing)
	}
}

func TestMacroFSMManager_GetStats(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)
	defer manager.Stop()

	provider1 := NewMockProviderFSM("provider-1", "cve")
	provider2 := NewMockProviderFSM("provider-2", "cwe")

	manager.AddProvider(provider1)
	manager.AddProvider(provider2)

	provider1.Transition(ProviderRunning)
	provider2.Transition(ProviderPaused)

	stats := manager.GetStats()

	if stats["id"] != "test-macro" {
		t.Errorf("Stats ID = %v, want test-macro", stats["id"])
	}

	if stats["provider_count"] != 2 {
		t.Errorf("Stats provider_count = %v, want 2", stats["provider_count"])
	}

	providerStates := stats["provider_states"].(map[string]int)
	if providerStates["RUNNING"] != 1 {
		t.Errorf("Running providers = %d, want 1", providerStates["RUNNING"])
	}
	if providerStates["PAUSED"] != 1 {
		t.Errorf("Paused providers = %d, want 1", providerStates["PAUSED"])
	}
}

func TestMacroFSMManager_Stop(t *testing.T) {
	manager, _ := NewMacroFSMManager("test-macro", nil)

	err := manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager: %v", err)
	}

	if manager.GetState() != MacroDraining {
		t.Errorf("State after stop = %v, want %v", manager.GetState(), MacroDraining)
	}
}
