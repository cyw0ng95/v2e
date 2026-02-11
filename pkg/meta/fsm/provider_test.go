package fsm

import (
	"errors"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/urn"
)

func TestNewBaseProviderFSM(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
		Storage:      nil,
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create BaseProviderFSM: %v", err)
	}

	if provider.GetID() != "provider-1" {
		t.Errorf("ID = %v, want provider-1", provider.GetID())
	}

	if provider.GetType() != "cve" {
		t.Errorf("Type = %v, want cve", provider.GetType())
	}

	if provider.GetState() != ProviderIdle {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderIdle)
	}
}

func TestNewBaseProviderFSM_EmptyID(t *testing.T) {
	config := ProviderConfig{
		ID:           "",
		ProviderType: "cve",
	}

	_, err := NewBaseProviderFSM(config)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestNewBaseProviderFSM_EmptyType(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "",
	}

	_, err := NewBaseProviderFSM(config)
	if err == nil {
		t.Error("Expected error for empty type, got nil")
	}
}

func TestProviderFSM_Transition(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	tests := []struct {
		name      string
		fromState ProviderState
		toState   ProviderState
		wantErr   bool
	}{
		{
			name:      "IDLE to ACQUIRING",
			fromState: ProviderIdle,
			toState:   ProviderAcquiring,
			wantErr:   false,
		},
		{
			name:      "ACQUIRING to RUNNING",
			fromState: ProviderAcquiring,
			toState:   ProviderRunning,
			wantErr:   false,
		},
		{
			name:      "RUNNING to PAUSED",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial state
			provider.mu.Lock()
			provider.state = tt.fromState
			provider.mu.Unlock()

			err := provider.Transition(tt.toState)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transition() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && provider.GetState() != tt.toState {
				t.Errorf("State after transition = %v, want %v", provider.GetState(), tt.toState)
			}
		})
	}
}

func TestProviderFSM_Start(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	eventReceived := false
	provider.SetEventHandler(func(event *Event) error {
		if event.Type == EventProviderStarted {
			eventReceived = true
		}
		return nil
	})

	err := provider.Start()
	if err != nil {
		t.Fatalf("Failed to start provider: %v", err)
	}

	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderAcquiring)
	}

	if !eventReceived {
		t.Error("EventProviderStarted not received")
	}
}

func TestProviderFSM_Start_InvalidState(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to RUNNING
	provider.mu.Lock()
	provider.state = ProviderRunning
	provider.mu.Unlock()

	err := provider.Start()
	if err == nil {
		t.Error("Expected error when starting from RUNNING state, got nil")
	}
}

func TestProviderFSM_Pause(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to RUNNING
	provider.mu.Lock()
	provider.state = ProviderRunning
	provider.mu.Unlock()

	err := provider.Pause()
	if err != nil {
		t.Fatalf("Failed to pause provider: %v", err)
	}

	if provider.GetState() != ProviderPaused {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderPaused)
	}
}

func TestProviderFSM_Resume(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to PAUSED
	provider.mu.Lock()
	provider.state = ProviderPaused
	provider.mu.Unlock()

	err := provider.Resume()
	if err != nil {
		t.Fatalf("Failed to resume provider: %v", err)
	}

	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderAcquiring)
	}
}

func TestProviderFSM_Stop(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to RUNNING
	provider.mu.Lock()
	provider.state = ProviderRunning
	provider.mu.Unlock()

	err := provider.Stop()
	if err != nil {
		t.Fatalf("Failed to stop provider: %v", err)
	}

	if provider.GetState() != ProviderTerminated {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderTerminated)
	}
}

func TestProviderFSM_OnQuotaRevoked(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to RUNNING with some permits
	provider.mu.Lock()
	provider.state = ProviderRunning
	provider.permitsHeld = 5
	provider.mu.Unlock()

	err := provider.OnQuotaRevoked(3)
	if err != nil {
		t.Fatalf("Failed to handle quota revocation: %v", err)
	}

	if provider.GetState() != ProviderWaitingQuota {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderWaitingQuota)
	}

	provider.mu.RLock()
	permits := provider.permitsHeld
	provider.mu.RUnlock()

	if permits != 2 {
		t.Errorf("Permits held = %d, want 2", permits)
	}
}

func TestProviderFSM_OnQuotaGranted(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
		Executor: func() error {
			// Mock executor that does nothing
			return nil
		},
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to ACQUIRING
	provider.mu.Lock()
	provider.state = ProviderAcquiring
	provider.mu.Unlock()

	err := provider.OnQuotaGranted(5)
	if err != nil {
		t.Fatalf("Failed to handle quota grant: %v", err)
	}

	// Wait a bit for async state change
	time.Sleep(10 * time.Millisecond)

	if provider.GetState() != ProviderRunning {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderRunning)
	}

	provider.mu.RLock()
	permits := provider.permitsHeld
	provider.mu.RUnlock()

	if permits != 5 {
		t.Errorf("Permits held = %d, want 5", permits)
	}
}

func TestProviderFSM_OnRateLimited(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set to RUNNING
	provider.mu.Lock()
	provider.state = ProviderRunning
	provider.mu.Unlock()

	err := provider.OnRateLimited(10 * time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to handle rate limiting: %v", err)
	}

	if provider.GetState() != ProviderWaitingBackoff {
		t.Errorf("State = %v, want %v", provider.GetState(), ProviderWaitingBackoff)
	}

	// Wait for backoff to complete
	time.Sleep(50 * time.Millisecond)

	// Should transition back to ACQUIRING after backoff
	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State after backoff = %v, want %v", provider.GetState(), ProviderAcquiring)
	}
}

func TestProviderFSM_SaveCheckpoint(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
		Storage:      nil, // No storage for this test
	}
	provider, _ := NewBaseProviderFSM(config)

	itemURN := urn.MustParse("v2e::nvd::cve::CVE-2024-12233")

	err := provider.SaveCheckpoint(itemURN, true, "")
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	stats := provider.GetStats()
	if stats["last_checkpoint"] != itemURN.Key() {
		t.Errorf("last_checkpoint = %v, want %v", stats["last_checkpoint"], itemURN.Key())
	}

	if stats["processed_count"].(int64) != 1 {
		t.Errorf("processed_count = %v, want 1", stats["processed_count"])
	}
}

func TestProviderFSM_SaveCheckpoint_Nil(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	err := provider.SaveCheckpoint(nil, true, "")
	if err == nil {
		t.Error("Expected error for nil URN, got nil")
	}
}

func TestProviderFSM_SaveCheckpoint_EmitsEvent(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	eventCount := 0
	provider.SetEventHandler(func(event *Event) error {
		if event.Type == EventCheckpoint {
			eventCount++
		}
		return nil
	})

	// Save 100 checkpoints to trigger event
	for i := 0; i < 100; i++ {
		itemURN := urn.MustParse("v2e::nvd::cve::CVE-2024-00001")
		provider.SaveCheckpoint(itemURN, true, "")
	}

	if eventCount != 1 {
		t.Errorf("Checkpoint events = %d, want 1 (emitted at 100 items)", eventCount)
	}
}

func TestProviderFSM_Execute_NoExecutor(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
		Executor:     nil,
	}
	provider, _ := NewBaseProviderFSM(config)

	err := provider.Execute()
	if err == nil {
		t.Error("Expected error when no executor defined, got nil")
	}
}

func TestProviderFSM_Execute_WithExecutor(t *testing.T) {
	executed := false
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
		Executor: func() error {
			executed = true
			return nil
		},
	}
	provider, _ := NewBaseProviderFSM(config)

	err := provider.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !executed {
		t.Error("Executor was not called")
	}
}

func TestProviderFSM_Execute_Error(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
		Executor: func() error {
			return errors.New("execution failed")
		},
	}
	provider, _ := NewBaseProviderFSM(config)

	err := provider.Execute()
	if err == nil {
		t.Error("Expected error from executor, got nil")
	}
}

func TestProviderFSM_GetStats(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	// Set some stats
	provider.mu.Lock()
	provider.processedCount = 50
	provider.errorCount = 2
	provider.permitsHeld = 3
	provider.mu.Unlock()

	stats := provider.GetStats()

	if stats["id"] != "provider-1" {
		t.Errorf("id = %v, want provider-1", stats["id"])
	}

	if stats["provider_type"] != "cve" {
		t.Errorf("provider_type = %v, want cve", stats["provider_type"])
	}

	if stats["processed_count"].(int64) != 50 {
		t.Errorf("processed_count = %v, want 50", stats["processed_count"])
	}

	if stats["error_count"].(int64) != 2 {
		t.Errorf("error_count = %v, want 2", stats["error_count"])
	}

	if stats["permits_held"].(int32) != 3 {
		t.Errorf("permits_held = %v, want 3", stats["permits_held"])
	}
}

func TestProviderFSM_EventHandler(t *testing.T) {
	config := ProviderConfig{
		ID:           "provider-1",
		ProviderType: "cve",
	}
	provider, _ := NewBaseProviderFSM(config)

	var receivedEvent *Event
	provider.SetEventHandler(func(event *Event) error {
		receivedEvent = event
		return nil
	})

	// Trigger an event
	provider.emitEvent(EventProviderStarted)

	if receivedEvent == nil {
		t.Fatal("Event handler was not called")
	}

	if receivedEvent.Type != EventProviderStarted {
		t.Errorf("Event type = %v, want %v", receivedEvent.Type, EventProviderStarted)
	}

	if receivedEvent.ProviderID != "provider-1" {
		t.Errorf("Event provider ID = %v, want provider-1", receivedEvent.ProviderID)
	}
}
