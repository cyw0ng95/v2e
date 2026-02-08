package fsm

import (
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// TestProviderFSM_FullLifecycle tests complete provider lifecycle with state persistence
func TestProviderFSM_FullLifecycle(t *testing.T) {
	// Create in-memory storage for testing
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	store, err := storage.NewStore(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	config := ProviderConfig{
		ID:           "test-provider-lifecycle",
		ProviderType: "cve",
		Storage:      store,
		Executor: func() error {
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	// Test 1: Initial state should be IDLE
	if provider.GetState() != ProviderIdle {
		t.Errorf("Initial state = %v, want %v", provider.GetState(), ProviderIdle)
	}

	// Test 2: Start should transition to ACQUIRING
	eventReceived := false
	var receivedEvent *Event
	provider.SetEventHandler(func(event *Event) error {
		receivedEvent = event
		eventReceived = true
		return nil
	})

	err = provider.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !eventReceived {
		t.Error("Event handler not called on Start")
	}
	if receivedEvent.Type != EventProviderStarted {
		t.Errorf("Event type = %v, want %v", receivedEvent.Type, EventProviderStarted)
	}

	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State after Start = %v, want %v", provider.GetState(), ProviderAcquiring)
	}

	// Test 3: Grant quota to transition to RUNNING
	err = provider.OnQuotaGranted(10)
	if err != nil {
		t.Fatalf("OnQuotaGranted failed: %v", err)
	}

	// Wait for async execution to start
	time.Sleep(200 * time.Millisecond)

	if provider.GetState() != ProviderRunning {
		t.Errorf("State after quota grant = %v, want %v", provider.GetState(), ProviderRunning)
	}

	// Test 4: Pause should transition to PAUSED
	err = provider.Pause()
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}

	if provider.GetState() != ProviderPaused {
		t.Errorf("State after Pause = %v, want %v", provider.GetState(), ProviderPaused)
	}

	// Test 5: Resume should transition back to ACQUIRING
	err = provider.Resume()
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State after Resume = %v, want %v", provider.GetState(), ProviderAcquiring)
	}

	// Test 6: Grant quota again to go to RUNNING
	err = provider.OnQuotaGranted(10)
	if err != nil {
		t.Fatalf("OnQuotaGranted (2nd) failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if provider.GetState() != ProviderRunning {
		t.Errorf("State after second quota grant = %v, want %v", provider.GetState(), ProviderRunning)
	}

	// Test 7: Stop should transition to TERMINATED
	err = provider.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if provider.GetState() != ProviderTerminated {
		t.Errorf("State after Stop = %v, want %v", provider.GetState(), ProviderTerminated)
	}

	// Test 8: State should be persisted
	loadedState, err := store.GetProviderState(provider.GetID())
	if err != nil {
		t.Fatalf("Failed to load persisted state: %v", err)
	}

	if loadedState.State != storage.ProviderState(ProviderTerminated) {
		t.Errorf("Persisted state = %v, want %v", loadedState.State, ProviderTerminated)
	}
}

// TestProviderFSM_QuotaRevocation tests quota revocation handling
func TestProviderFSM_QuotaRevocation(t *testing.T) {
	config := ProviderConfig{
		ID:           "test-quota-revocation",
		ProviderType: "cve",
		Storage:      nil,
		Executor:     func() error { return nil },
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	// Start and grant quota to reach RUNNING state
	provider.Start()
	provider.OnQuotaGranted(10)
	time.Sleep(100 * time.Millisecond)

	if provider.GetState() != ProviderRunning {
		t.Errorf("State before revocation = %v, want %v", provider.GetState(), ProviderRunning)
	}

	// Revoke quota
	err = provider.OnQuotaRevoked(5)
	if err != nil {
		t.Fatalf("OnQuotaRevoked failed: %v", err)
	}

	// Should transition to WAITING_QUOTA
	if provider.GetState() != ProviderWaitingQuota {
		t.Errorf("State after revocation = %v, want %v", provider.GetState(), ProviderWaitingQuota)
	}

	// Grant quota to retry acquisition
	err = provider.OnQuotaGranted(3)
	if err != nil {
		t.Fatalf("OnQuotaGranted after revocation failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Should transition back to ACQUIRING to retry
	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State after quota retry = %v, want %v", provider.GetState(), ProviderAcquiring)
	}
}

// TestProviderFSM_RateLimiting tests rate limit backoff handling
func TestProviderFSM_RateLimiting(t *testing.T) {
	config := ProviderConfig{
		ID:           "test-rate-limit",
		ProviderType: "cve",
		Storage:      nil,
		Executor:     func() error { return nil },
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	// Start and grant quota to reach RUNNING state
	provider.Start()
	provider.OnQuotaGranted(10)
	time.Sleep(100 * time.Millisecond)

	if provider.GetState() != ProviderRunning {
		t.Errorf("State before rate limit = %v, want %v", provider.GetState(), ProviderRunning)
	}

	// Simulate rate limit
	retryAfter := 50 * time.Millisecond
	err = provider.OnRateLimited(retryAfter)
	if err != nil {
		t.Fatalf("OnRateLimited failed: %v", err)
	}

	// Should transition to WAITING_BACKOFF
	if provider.GetState() != ProviderWaitingBackoff {
		t.Errorf("State after rate limit = %v, want %v", provider.GetState(), ProviderWaitingBackoff)
	}

	// Wait for backoff to complete
	time.Sleep(100 * time.Millisecond)

	// Should transition back to ACQUIRING to retry
	if provider.GetState() != ProviderAcquiring {
		t.Errorf("State after backoff = %v, want %v", provider.GetState(), ProviderAcquiring)
	}
}

// TestProviderFSM_CheckpointPersistence tests checkpoint saving and recovery
func TestProviderFSM_CheckpointPersistence(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	store, err := storage.NewStore(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	config := ProviderConfig{
		ID:           "test-checkpoint",
		ProviderType: "cve",
		Storage:      store,
		Executor:     func() error { return nil },
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	// Save multiple checkpoints
	itemURN := urn.MustParse("v2e::nvd::cve::CVE-2024-12233")
	for i := 0; i < 5; i++ {
		err := provider.SaveCheckpoint(itemURN, true, "")
		if err != nil {
			t.Fatalf("SaveCheckpoint failed: %v", err)
		}
	}

	// Verify stats
	stats := provider.GetStats()
	if stats["processed_count"].(int64) != 5 {
		t.Errorf("Processed count = %v, want 5", stats["processed_count"])
	}

	// Verify checkpoint was saved
	checkpoints, err := store.ListCheckpointsByProvider(provider.GetID())
	if err != nil {
		t.Fatalf("Failed to get checkpoints: %v", err)
	}

	if len(checkpoints) != 5 {
		t.Errorf("Checkpoint count = %d, want 5", len(checkpoints))
	}

	// Verify checkpoint data
	for i, checkpoint := range checkpoints {
		if checkpoint.URN != itemURN.Key() {
			t.Errorf("Checkpoint %d URN = %v, want %v", i, checkpoint.URN, itemURN.Key())
		}
		if checkpoint.ProviderID != provider.GetID() {
			t.Errorf("Checkpoint %d ProviderID = %v, want %v", i, checkpoint.ProviderID, provider.GetID())
		}
		if !checkpoint.Success {
			t.Errorf("Checkpoint %d Success = false, want true", i)
		}
	}

	// Test failed checkpoint
	err = provider.SaveCheckpoint(itemURN, false, "Test error")
	if err != nil {
		t.Fatalf("SaveCheckpoint with error failed: %v", err)
	}

	stats = provider.GetStats()
	if stats["error_count"].(int64) != 1 {
		t.Errorf("Error count = %v, want 1", stats["error_count"])
	}

	// Verify failed checkpoint
	checkpoints, err = store.ListCheckpointsByProvider(provider.GetID())
	if err != nil {
		t.Fatalf("Failed to get checkpoints after error: %v", err)
	}

	lastCheckpoint := checkpoints[len(checkpoints)-1]
	if lastCheckpoint.Success {
		t.Error("Last checkpoint should be marked as failed")
	}
	if lastCheckpoint.ErrorMessage != "Test error" {
		t.Errorf("Last checkpoint error message = %v, want 'Test error'", lastCheckpoint.ErrorMessage)
	}
}

// TestProviderFSM_InvalidTransitions tests that invalid transitions are rejected
func TestProviderFSM_InvalidTransitions(t *testing.T) {
	config := ProviderConfig{
		ID:           "test-invalid-transitions",
		ProviderType: "cve",
		Storage:      nil,
		Executor:     func() error { return nil },
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	// Test invalid transition: IDLE -> RUNNING
	err = provider.Transition(ProviderRunning)
	if err == nil {
		t.Error("Expected error for IDLE -> RUNNING transition, got nil")
	}

	// Test invalid transition: IDLE -> PAUSED
	err = provider.Transition(ProviderPaused)
	if err == nil {
		t.Error("Expected error for IDLE -> PAUSED transition, got nil")
	}

	// Test invalid transition: RUNNING -> IDLE
	provider.mu.Lock()
	provider.state = ProviderRunning
	provider.mu.Unlock()

	err = provider.Transition(ProviderIdle)
	if err == nil {
		t.Error("Expected error for RUNNING -> IDLE transition, got nil")
	}
}

// TestProviderFSM_StateRecovery tests state recovery after restart
func TestProviderFSM_StateRecovery(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
	store, err := storage.NewStore(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Create provider and transition to RUNNING
	config := ProviderConfig{
		ID:           "test-recovery",
		ProviderType: "cve",
		Storage:      store,
		Executor:     func() error { return nil },
	}

	provider1, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	provider1.Start()
	provider1.OnQuotaGranted(10)
	time.Sleep(100 * time.Millisecond)

	// Save some checkpoints
	itemURN := urn.MustParse("v2e::nvd::cve::CVE-2024-12233")
	provider1.SaveCheckpoint(itemURN, true, "")

	// Create a new provider instance (simulating restart)
	config2 := ProviderConfig{
		ID:           "test-recovery",
		ProviderType: "cve",
		Storage:      store,
		Executor:     func() error { return nil },
	}

	provider2, err := NewBaseProviderFSM(config2)
	if err != nil {
		t.Fatalf("Failed to create second provider FSM: %v", err)
	}

	// Verify state was recovered
	if provider2.GetState() != ProviderRunning {
		t.Errorf("Recovered state = %v, want %v", provider2.GetState(), ProviderRunning)
	}

	// Verify stats were recovered
	stats2 := provider2.GetStats()
	if stats2["processed_count"].(int64) != 1 {
		t.Errorf("Recovered processed count = %v, want 1", stats2["processed_count"])
	}

	if stats2["last_checkpoint"] != itemURN.Key() {
		t.Errorf("Recovered last checkpoint = %v, want %v", stats2["last_checkpoint"], itemURN.Key())
	}
}

// TestProviderFSM_MultipleProviders tests multiple providers with different states
func TestProviderFSM_MultipleProviders(t *testing.T) {
	providers := make([]*BaseProviderFSM, 0, 4)

	// Create 4 providers with different initial states
	providerTypes := []string{"cve", "cwe", "capec", "attack"}
	for i, pType := range providerTypes {
		config := ProviderConfig{
			ID:           pType + "-provider",
			ProviderType: pType,
			Storage:      nil,
			Executor:     func() error { return nil },
		}

		provider, err := NewBaseProviderFSM(config)
		if err != nil {
			t.Fatalf("Failed to create provider %s: %v", pType, err)
		}

		providers = append(providers, provider)

		// Set different initial states
		switch i {
		case 0:
			provider.Transition(ProviderRunning)
		case 1:
			provider.Transition(ProviderPaused)
		case 2:
			provider.Transition(ProviderWaitingQuota)
		case 3:
			provider.Transition(ProviderWaitingBackoff)
		}
	}

	// Verify each provider has correct state
	expectedStates := []ProviderState{
		ProviderRunning,
		ProviderPaused,
		ProviderWaitingQuota,
		ProviderWaitingBackoff,
	}

	for i, provider := range providers {
		if provider.GetState() != expectedStates[i] {
			t.Errorf("Provider %d state = %v, want %v", i, provider.GetState(), expectedStates[i])
		}

		// Test pause/resume on each
		if provider.GetState() == ProviderRunning {
			err := provider.Pause()
			if err != nil {
				t.Errorf("Failed to pause provider %d: %v", i, err)
			}
			if provider.GetState() != ProviderPaused {
				t.Errorf("Provider %d state after pause = %v, want %v", i, provider.GetState(), ProviderPaused)
			}
		} else if provider.GetState() == ProviderPaused {
			err := provider.Resume()
			if err != nil {
				t.Errorf("Failed to resume provider %d: %v", i, err)
			}
			if provider.GetState() != ProviderAcquiring {
				t.Errorf("Provider %d state after resume = %v, want %v", i, provider.GetState(), ProviderAcquiring)
			}
		}
	}
}

// TestProviderFSM_ConcurrentOperations tests concurrent access to provider FSM
func TestProviderFSM_ConcurrentOperations(t *testing.T) {
	config := ProviderConfig{
		ID:           "test-concurrent",
		ProviderType: "cve",
		Storage:      nil,
		Executor:     func() error { return nil },
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	// Start the provider
	err = provider.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Perform concurrent operations
	done := make(chan bool, 1)

	go func() {
		// Concurrent state reads
		for i := 0; i < 100; i++ {
			_ = provider.GetState()
			_ = provider.GetStats()
		}
		done <- true
	}()

	go func() {
		// Concurrent checkpoint saves
		itemURN := urn.MustParse("v2e::nvd::cve::CVE-2024-12233")
		for i := 0; i < 100; i++ {
			_ = provider.SaveCheckpoint(itemURN, true, "")
		}
		done <- true
	}()

	go func() {
		// Concurrent quota grants
		for i := 0; i < 10; i++ {
			_ = provider.OnQuotaGranted(5)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify final state is valid
	finalState := provider.GetState()
	if finalState != ProviderAcquiring && finalState != ProviderRunning && finalState != ProviderTerminated {
		t.Errorf("Final state %v is not valid after concurrent operations", finalState)
	}

	// Verify stats are consistent
	stats := provider.GetStats()
	if stats["processed_count"].(int64) != 100 {
		t.Errorf("Processed count = %v, want 100", stats["processed_count"])
	}
}

// TestProviderFSM_ContextCancellation tests graceful shutdown on context cancellation
func TestProviderFSM_ContextCancellation(t *testing.T) {
	config := ProviderConfig{
		ID:           "test-context-cancel",
		ProviderType: "cve",
		Storage:      nil,
		Executor: func() error {
			// Simulate long-running operation
			time.Sleep(1 * time.Second)
			return nil
		},
	}

	provider, err := NewBaseProviderFSM(config)
	if err != nil {
		t.Fatalf("Failed to create provider FSM: %v", err)
	}

	provider.Start()
	provider.OnQuotaGranted(10)
	time.Sleep(50 * time.Millisecond)

	// Stop the provider
	err = provider.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify provider is in TERMINATED state
	if provider.GetState() != ProviderTerminated {
		t.Errorf("State after context cancellation = %v, want %v", provider.GetState(), ProviderTerminated)
	}
}
