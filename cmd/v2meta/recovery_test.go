package main

import (
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// testLogger creates a logger for tests
func testLogger() *common.Logger {
	return common.NewLogger(os.Stderr, "test", common.InfoLevel)
}

// MockMacroFSM for testing recovery
type MockMacroFSM struct {
	providers map[string]fsm.ProviderFSM
}

func NewMockMacroFSM() *MockMacroFSM {
	return &MockMacroFSM{
		providers: make(map[string]fsm.ProviderFSM),
	}
}

func (m *MockMacroFSM) AddProvider(provider fsm.ProviderFSM) {
	m.providers[provider.GetID()] = provider
}

func (m *MockMacroFSM) GetProviders() []fsm.ProviderFSM {
	providers := make([]fsm.ProviderFSM, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	return providers
}

// MockStorage for testing recovery
type MockStorage struct {
	providerStates []*storage.ProviderFSMState
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		providerStates: make([]*storage.ProviderFSMState, 0),
	}
}

func (m *MockStorage) ListProviderStates() ([]*storage.ProviderFSMState, error) {
	return m.providerStates, nil
}

func (m *MockStorage) AddProviderState(state *storage.ProviderFSMState) {
	m.providerStates = append(m.providerStates, state)
}

// Test 1: RecoveryManager - Create Recovery Manager
func TestRecoveryManager_New(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewRecoveryManager", nil, func(t *testing.T, tx *gorm.DB) {
		storage := &storage.Store{}
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()
		logger := testLogger()

		manager := NewRecoveryManager(storage, executor, macroFSM, logger)

		if manager == nil {
			t.Fatal("RecoveryManager is nil")
		}

		if manager.storage != storage {
			t.Error("Storage not set correctly")
		}

		if manager.executor != executor {
			t.Error("Executor not set correctly")
		}
	})
}

// Test 2: RecoveryManager - Recover No Providers
func TestRecoveryManager_RecoverProviders_Empty(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProvidersEmpty", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// No providers to recover
		if len(executor.activeJobs) != 0 {
			t.Errorf("Active jobs count = %d, want 0", len(executor.activeJobs))
		}
	})
}

// Test 3: RecoveryManager - Recover RUNNING Provider
func TestRecoveryManager_RecoverProvider_Running(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderRunning", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testLogger())
		macroFSM := NewMockMacroFSM()

		// Create provider
		provider := NewMockProvider("provider-1")
		macroFSM.AddProvider(provider)

		// Add RUNNING state
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:          "provider-1",
			State:       storage.ProviderRunning,
			PermitsHeld: 5,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Provider should be restarted
		time.Sleep(50 * time.Millisecond)
		if provider.GetState() != fsm.ProviderRunning {
			t.Errorf("Provider state = %s, want RUNNING", provider.GetState())
		}
	})
}

// Test 4: RecoveryManager - Recover WAITING_QUOTA Provider
func TestRecoveryManager_RecoverProvider_WaitingQuota(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderWaitingQuota", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(3),
		}

		executor := NewPermitExecutor(mockRPC, testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-2")
		macroFSM.AddProvider(provider)

		// Add WAITING_QUOTA state
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:          "provider-2",
			State:       storage.ProviderWaitingQuota,
			PermitsHeld: 2,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Provider should transition to ACQUIRING and then RUNNING
		time.Sleep(50 * time.Millisecond)
		if provider.GetState() != fsm.ProviderRunning {
			t.Errorf("Provider state = %s, want RUNNING", provider.GetState())
		}
	})
}

// Test 5: RecoveryManager - Skip PAUSED Provider
func TestRecoveryManager_RecoverProvider_Paused(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderPaused", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-3")
		provider.Transition(fsm.ProviderPaused)
		macroFSM.AddProvider(provider)

		// Add PAUSED state
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-3",
			State: storage.ProviderPaused,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Provider should remain PAUSED
		if provider.GetState() != fsm.ProviderPaused {
			t.Errorf("Provider state = %s, want PAUSED", provider.GetState())
		}

		// Should not be in active jobs
		if len(executor.activeJobs) != 0 {
			t.Errorf("Active jobs count = %d, want 0 (paused providers not started)", len(executor.activeJobs))
		}
	})
}

// Test 6: RecoveryManager - Skip TERMINATED Provider
func TestRecoveryManager_RecoverProvider_Terminated(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderTerminated", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-4")
		provider.Transition(fsm.ProviderTerminated)
		macroFSM.AddProvider(provider)

		// Add TERMINATED state
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-4",
			State: storage.ProviderTerminated,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Provider should remain TERMINATED
		if provider.GetState() != fsm.ProviderTerminated {
			t.Errorf("Provider state = %s, want TERMINATED", provider.GetState())
		}
	})
}

// Test 7: RecoveryManager - Skip IDLE Provider
func TestRecoveryManager_RecoverProvider_Idle(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderIdle", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-5")
		macroFSM.AddProvider(provider)

		// Add IDLE state
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-5",
			State: storage.ProviderIdle,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Provider should remain IDLE
		if provider.GetState() != fsm.ProviderIdle {
			t.Errorf("Provider state = %s, want IDLE", provider.GetState())
		}
	})
}

// Test 8: RecoveryManager - Skip WAITING_BACKOFF Provider
func TestRecoveryManager_RecoverProvider_WaitingBackoff(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderWaitingBackoff", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-6")
		provider.Transition(fsm.ProviderWaitingBackoff)
		macroFSM.AddProvider(provider)

		// Add WAITING_BACKOFF state
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-6",
			State: storage.ProviderWaitingBackoff,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Provider should maintain state
		if provider.GetState() != fsm.ProviderWaitingBackoff {
			t.Errorf("Provider state = %s, want WAITING_BACKOFF", provider.GetState())
		}
	})
}

// Test 9: RecoveryManager - Recover Multiple Providers
func TestRecoveryManager_RecoverMultipleProviders(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverMultipleProviders", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testLogger())
		macroFSM := NewMockMacroFSM()

		// Add 3 providers in different states
		provider1 := NewMockProvider("provider-running")
		provider2 := NewMockProvider("provider-paused")
		provider3 := NewMockProvider("provider-waiting")

		macroFSM.AddProvider(provider1)
		macroFSM.AddProvider(provider2)
		macroFSM.AddProvider(provider3)

		provider2.Transition(fsm.ProviderPaused)

		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-running",
			State: storage.ProviderRunning,
		})

		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-paused",
			State: storage.ProviderPaused,
		})

		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "provider-waiting",
			State: storage.ProviderWaitingQuota,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Running and waiting should be recovered (2 active jobs)
		activeIDs := executor.GetActiveProviders()
		if len(activeIDs) != 2 {
			t.Errorf("Active providers count = %d, want 2", len(activeIDs))
		}
	})
}

// Test 10: RecoveryManager - Provider Not Found in MacroFSM
func TestRecoveryManager_RecoverProvider_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderNotFound", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testLogger())
		macroFSM := NewMockMacroFSM()

		// Add state but no provider in MacroFSM
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:    "missing-provider",
			State: storage.ProviderRunning,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		// Should not fail, but log error for missing provider
		if err != nil {
			t.Fatalf("RecoverProviders should not fail for missing provider: %v", err)
		}
	})
}

// Test 11: RecoveryManager - Checkpoint Restoration
func TestRecoveryManager_RecoverProvider_WithCheckpoint(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderWithCheckpoint", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-7")
		// Set checkpoint before recovery
		provider.SaveCheckpoint("v2e::nvd::cve::CVE-2024-12345")
		macroFSM.AddProvider(provider)

		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:          "provider-7",
			State:       storage.ProviderRunning,
			PermitsHeld: 5,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Checkpoint should be preserved
		checkpoint, _ := provider.GetLastCheckpoint()
		if checkpoint != "v2e::nvd::cve::CVE-2024-12345" {
			t.Errorf("Checkpoint = %s, want v2e::nvd::cve::CVE-2024-12345", checkpoint)
		}
	})
}

// Test 12: RecoveryManager - Should Recover Check
func TestRecoveryManager_ShouldRecover(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ShouldRecover", nil, func(t *testing.T, tx *gorm.DB) {
		manager := NewRecoveryManager(nil, nil, nil, testLogger())

		tests := []struct {
			state    storage.ProviderState
			expected bool
		}{
			{storage.ProviderRunning, true},
			{storage.ProviderWaitingQuota, true},
			{storage.ProviderPaused, false},
			{storage.ProviderTerminated, false},
			{storage.ProviderIdle, false},
			{storage.ProviderWaitingBackoff, false},
		}

		for _, tt := range tests {
			result := manager.shouldRecover(tt.state)
			if result != tt.expected {
				t.Errorf("shouldRecover(%s) = %v, want %v", tt.state, result, tt.expected)
			}
		}
	})
}

// Test 13: RecoveryManager - Get Recovery Stats
func TestRecoveryManager_GetRecoveryStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetRecoveryStats", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()

		mockStorage.AddProviderState(&storage.ProviderFSMState{ID: "p1", State: storage.ProviderRunning})
		mockStorage.AddProviderState(&storage.ProviderFSMState{ID: "p2", State: storage.ProviderWaitingQuota})
		mockStorage.AddProviderState(&storage.ProviderFSMState{ID: "p3", State: storage.ProviderPaused})
		mockStorage.AddProviderState(&storage.ProviderFSMState{ID: "p4", State: storage.ProviderTerminated})
		mockStorage.AddProviderState(&storage.ProviderFSMState{ID: "p5", State: storage.ProviderIdle})

		manager := NewRecoveryManager(mockStorage, nil, nil, testLogger())

		stats, err := manager.GetRecoveryStats()
		if err != nil {
			t.Fatalf("GetRecoveryStats failed: %v", err)
		}

		if stats.TotalProviders != 5 {
			t.Errorf("Total providers = %d, want 5", stats.TotalProviders)
		}

		if stats.RecoveredProviders != 2 {
			t.Errorf("Recovered providers = %d, want 2 (RUNNING + WAITING_QUOTA)", stats.RecoveredProviders)
		}

		if stats.SkippedProviders != 2 {
			t.Errorf("Skipped providers = %d, want 2 (TERMINATED + IDLE)", stats.SkippedProviders)
		}
	})
}

// Test 14: RecoveryManager - Default Permits for Recovery
func TestRecoveryManager_RecoverProvider_DefaultPermits(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderDefaultPermits", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(2),
		}

		executor := NewPermitExecutor(mockRPC, testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-8")
		macroFSM.AddProvider(provider)

		// Add state without permits held (should use default 2)
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:          "provider-8",
			State:       storage.ProviderRunning,
			PermitsHeld: 0, // No permits
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Should request default 2 permits
		if len(mockRPC.permitRequests) == 0 {
			t.Fatal("No permit requests made")
		}

		if mockRPC.permitRequests[0] != 2 {
			t.Errorf("Requested permits = %d, want 2 (default)", mockRPC.permitRequests[0])
		}
	})
}

// Test 15: RecoveryManager - Recovery with Existing Permits
func TestRecoveryManager_RecoverProvider_ExistingPermits(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RecoverProviderExistingPermits", nil, func(t *testing.T, tx *gorm.DB) {
		mockStorage := NewMockStorage()
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(10),
		}

		executor := NewPermitExecutor(mockRPC, testLogger())
		macroFSM := NewMockMacroFSM()

		provider := NewMockProvider("provider-9")
		macroFSM.AddProvider(provider)

		// Add state with 10 permits held
		mockStorage.AddProviderState(&storage.ProviderFSMState{
			ID:          "provider-9",
			State:       storage.ProviderRunning,
			PermitsHeld: 10,
		})

		manager := NewRecoveryManager(mockStorage, executor, macroFSM, testLogger())

		err := manager.RecoverProviders()
		if err != nil {
			t.Fatalf("RecoverProviders failed: %v", err)
		}

		// Should request same number of permits as before
		if len(mockRPC.permitRequests) == 0 {
			t.Fatal("No permit requests made")
		}

		if mockRPC.permitRequests[0] != 10 {
			t.Errorf("Requested permits = %d, want 10 (from state)", mockRPC.permitRequests[0])
		}
	})
}
