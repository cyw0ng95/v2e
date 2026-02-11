package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

func setupTestStore(t *testing.T) (*Store, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := common.NewLogger(os.Stderr, "[TEST] ", common.DebugLevel)

	store, err := NewStore(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	return store, dbPath
}

func TestNewStore(t *testing.T) {
	testutils.Run(t, testutils.Level2, "NewStore", nil, func(t *testing.T, tx *gorm.DB) {
		store, _ := setupTestStore(t)
		defer store.Close()

		// Verify all buckets are created
		err := store.db.View(func(tx *bolt.Tx) error {
			buckets := [][]byte{
				BucketSessions,
				BucketFSMStates,
				BucketProviderStates,
				BucketCheckpoints,
				BucketPermits,
			}
			for _, bucketName := range buckets {
				if tx.Bucket(bucketName) == nil {
					t.Errorf("Bucket %s not created", bucketName)
				}
			}
			return nil
		})
		if err != nil {
			t.Errorf("Failed to verify buckets: %v", err)
		}
	})
}

func TestMacroState(t *testing.T) {
	testutils.Run(t, testutils.Level2, "MacroState", nil, func(t *testing.T, tx *gorm.DB) {
		store, _ := setupTestStore(t)
		defer store.Close()

		// Test save
		state := &MacroFSMState{
			ID:        "macro-1",
			State:     MacroBootstrapping,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  map[string]interface{}{"key": "value"},
		}

		err := store.SaveMacroState(state)
		if err != nil {
			t.Fatalf("Failed to save macro state: %v", err)
		}

		// Test get
		retrieved, err := store.GetMacroState("macro-1")
		if err != nil {
			t.Fatalf("Failed to get macro state: %v", err)
		}

		if retrieved.ID != state.ID {
			t.Errorf("ID mismatch: got %v, want %v", retrieved.ID, state.ID)
		}
		if retrieved.State != state.State {
			t.Errorf("State mismatch: got %v, want %v", retrieved.State, state.State)
		}

		// Test update
		state.State = MacroOrchestrating
		state.UpdatedAt = time.Now()
		err = store.SaveMacroState(state)
		if err != nil {
			t.Fatalf("Failed to update macro state: %v", err)
		}

		retrieved, err = store.GetMacroState("macro-1")
		if err != nil {
			t.Fatalf("Failed to get updated macro state: %v", err)
		}
		if retrieved.State != MacroOrchestrating {
			t.Errorf("State not updated: got %v, want %v", retrieved.State, MacroOrchestrating)
		}
	})
}

func TestProviderState(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ProviderState", nil, func(t *testing.T, tx *gorm.DB) {
		store, _ := setupTestStore(t)
		defer store.Close()

		// Test save
		state := &ProviderFSMState{
			ID:             "provider-cve-1",
			ProviderType:   "cve",
			State:          ProviderIdle,
			LastCheckpoint: "v2e::nvd::cve::CVE-2024-12233",
			ProcessedCount: 100,
			ErrorCount:     2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := store.SaveProviderState(state)
		if err != nil {
			t.Fatalf("Failed to save provider state: %v", err)
		}

		// Test get
		retrieved, err := store.GetProviderState("provider-cve-1")
		if err != nil {
			t.Fatalf("Failed to get provider state: %v", err)
		}

		if retrieved.ID != state.ID {
			t.Errorf("ID mismatch: got %v, want %v", retrieved.ID, state.ID)
		}
		if retrieved.State != state.State {
			t.Errorf("State mismatch: got %v, want %v", retrieved.State, state.State)
		}
		if retrieved.ProcessedCount != state.ProcessedCount {
			t.Errorf("ProcessedCount mismatch: got %v, want %v", retrieved.ProcessedCount, state.ProcessedCount)
		}

		// Test list
		states, err := store.ListProviderStates()
		if err != nil {
			t.Fatalf("Failed to list provider states: %v", err)
		}
		if len(states) != 1 {
			t.Errorf("Expected 1 provider state, got %d", len(states))
		}

		// Test delete
		err = store.DeleteProviderState("provider-cve-1")
		if err != nil {
			t.Fatalf("Failed to delete provider state: %v", err)
		}

		states, err = store.ListProviderStates()
		if err != nil {
			t.Fatalf("Failed to list provider states after delete: %v", err)
		}
		if len(states) != 0 {
			t.Errorf("Expected 0 provider states after delete, got %d", len(states))
		}
	})
}

func TestCheckpoint(t *testing.T) {
	testutils.Run(t, testutils.Level2, "Checkpoint", nil, func(t *testing.T, tx *gorm.DB) {
		store, _ := setupTestStore(t)
		defer store.Close()

		// Test save with valid URN
		urnStr := "v2e::nvd::cve::CVE-2024-12233"
		checkpoint := &Checkpoint{
			URN:         urnStr,
			ProviderID:  "provider-cve-1",
			ProcessedAt: time.Now(),
			Success:     true,
		}

		err := store.SaveCheckpoint(checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		// Test get
		retrieved, err := store.GetCheckpoint(urnStr)
		if err != nil {
			t.Fatalf("Failed to get checkpoint: %v", err)
		}

		if retrieved.URN != checkpoint.URN {
			t.Errorf("URN mismatch: got %v, want %v", retrieved.URN, checkpoint.URN)
		}
		if retrieved.Success != checkpoint.Success {
			t.Errorf("Success mismatch: got %v, want %v", retrieved.Success, checkpoint.Success)
		}

		// Test save with invalid URN
		badCheckpoint := &Checkpoint{
			URN:         "invalid-urn",
			ProviderID:  "provider-cve-1",
			ProcessedAt: time.Now(),
		}
		err = store.SaveCheckpoint(badCheckpoint)
		if err == nil {
			t.Error("Expected error for invalid URN, got nil")
		}

		// Test list by provider
		urn2 := urn.MustNew(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-99999")
		checkpoint2 := &Checkpoint{
			URN:         urn2.String(),
			ProviderID:  "provider-cve-1",
			ProcessedAt: time.Now(),
			Success:     true,
		}
		err = store.SaveCheckpoint(checkpoint2)
		if err != nil {
			t.Fatalf("Failed to save second checkpoint: %v", err)
		}

		checkpoints, err := store.ListCheckpointsByProvider("provider-cve-1")
		if err != nil {
			t.Fatalf("Failed to list checkpoints: %v", err)
		}
		if len(checkpoints) != 2 {
			t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
		}
	})
}

func TestPermitAllocation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "PermitAllocation", nil, func(t *testing.T, tx *gorm.DB) {
		store, _ := setupTestStore(t)
		defer store.Close()

		// Test save
		now := time.Now()
		allocation := &PermitAllocation{
			ProviderID:  "provider-cve-1",
			PermitCount: 5,
			AllocatedAt: now,
		}

		err := store.SavePermitAllocation(allocation)
		if err != nil {
			t.Fatalf("Failed to save permit allocation: %v", err)
		}

		// Test get
		retrieved, err := store.GetPermitAllocation("provider-cve-1")
		if err != nil {
			t.Fatalf("Failed to get permit allocation: %v", err)
		}

		if retrieved.ProviderID != allocation.ProviderID {
			t.Errorf("ProviderID mismatch: got %v, want %v", retrieved.ProviderID, allocation.ProviderID)
		}
		if retrieved.PermitCount != allocation.PermitCount {
			t.Errorf("PermitCount mismatch: got %v, want %v", retrieved.PermitCount, allocation.PermitCount)
		}

		// Test update with release time
		releasedAt := time.Now()
		allocation.ReleasedAt = &releasedAt
		err = store.SavePermitAllocation(allocation)
		if err != nil {
			t.Fatalf("Failed to update permit allocation: %v", err)
		}

		retrieved, err = store.GetPermitAllocation("provider-cve-1")
		if err != nil {
			t.Fatalf("Failed to get updated permit allocation: %v", err)
		}
		if retrieved.ReleasedAt == nil {
			t.Error("ReleasedAt should not be nil after update")
		}
	})
}

func TestProviderStateTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ProviderStateTransitions", nil, func(t *testing.T, tx *gorm.DB) {
		store, _ := setupTestStore(t)
		defer store.Close()

		state := &ProviderFSMState{
			ID:           "provider-cve-1",
			ProviderType: "cve",
			State:        ProviderIdle,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Test state transitions
		transitions := []ProviderState{
			ProviderAcquiring,
			ProviderRunning,
			ProviderWaitingQuota,
			ProviderRunning,
			ProviderPaused,
			ProviderRunning,
			ProviderTerminated,
		}

		err := store.SaveProviderState(state)
		if err != nil {
			t.Fatalf("Failed to save initial state: %v", err)
		}

		for _, nextState := range transitions {
			state.State = nextState
			state.UpdatedAt = time.Now()
			err := store.SaveProviderState(state)
			if err != nil {
				t.Fatalf("Failed to transition to %v: %v", nextState, err)
			}

			retrieved, err := store.GetProviderState("provider-cve-1")
			if err != nil {
				t.Fatalf("Failed to get state after transition: %v", err)
			}
			if retrieved.State != nextState {
				t.Errorf("State transition failed: expected %v, got %v", nextState, retrieved.State)
			}
		}
	})
}
