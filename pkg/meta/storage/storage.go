package storage

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// Bucket names for different data types
var (
	BucketSessions       = []byte("sessions")        // Original session data
	BucketFSMStates      = []byte("fsm_states")      // Macro FSM states
	BucketProviderStates = []byte("provider_states") // Provider FSM states
	BucketCheckpoints    = []byte("checkpoints")     // URN-based checkpoints
	BucketPermits        = []byte("permits")         // Permit tracking
)

// MacroState represents the state of the macro FSM
type MacroState string

const (
	MacroBootstrapping MacroState = "BOOTSTRAPPING"
	MacroOrchestrating MacroState = "ORCHESTRATING"
	MacroStabilizing   MacroState = "STABILIZING"
	MacroDraining      MacroState = "DRAINING"
)

// ProviderState represents the state of a provider FSM.
// NOTE: This type must match fsm.ProviderState exactly.
// Both types use the same string values for compatibility.
type ProviderState string

const (
	ProviderIdle           ProviderState = "IDLE"
	ProviderAcquiring      ProviderState = "ACQUIRING"
	ProviderRunning        ProviderState = "RUNNING"
	ProviderWaitingQuota   ProviderState = "WAITING_QUOTA"
	ProviderWaitingBackoff ProviderState = "WAITING_BACKOFF"
	ProviderPaused         ProviderState = "PAUSED"
	ProviderTerminated     ProviderState = "TERMINATED"
)

// MacroFSMState represents the persisted state of the macro FSM
type MacroFSMState struct {
	ID        string                 `json:"id"`
	State     MacroState             `json:"state"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderFSMState represents the persisted state of a provider FSM
type ProviderFSMState struct {
	ID             string                 `json:"id"`
	ProviderType   string                 `json:"provider_type"` // "cve", "cwe", "capec", "attack"
	State          ProviderState          `json:"state"`
	LastCheckpoint string                 `json:"last_checkpoint,omitempty"` // URN of last processed item
	ProcessedCount int64                  `json:"processed_count"`
	ErrorCount     int64                  `json:"error_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Checkpoint represents a URN-based checkpoint
type Checkpoint struct {
	URN          string    `json:"urn"`         // Full URN string
	ProviderID   string    `json:"provider_id"` // Provider FSM ID
	ProcessedAt  time.Time `json:"processed_at"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// PermitAllocation represents a permit allocation to a provider
type PermitAllocation struct {
	ProviderID  string     `json:"provider_id"`
	PermitCount int        `json:"permit_count"`
	AllocatedAt time.Time  `json:"allocated_at"`
	ReleasedAt  *time.Time `json:"released_at,omitempty"`
}

// Store provides enhanced storage capabilities for FSM and ETL data
type Store struct {
	db     *bolt.DB
	logger *common.Logger
}

// NewStore creates a new enhanced store
func NewStore(dbPath string, logger *common.Logger) (*Store, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create all buckets
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			BucketSessions,
			BucketFSMStates,
			BucketProviderStates,
			BucketCheckpoints,
			BucketPermits,
		}
		for _, bucketName := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveMacroState saves the macro FSM state
func (s *Store) SaveMacroState(state *MacroFSMState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal macro state: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketFSMStates)
		return b.Put([]byte(state.ID), data)
	})
}

// GetMacroState retrieves the macro FSM state by ID
func (s *Store) GetMacroState(id string) (*MacroFSMState, error) {
	var state MacroFSMState
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketFSMStates)
		v := b.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("macro state not found: %s", id)
		}
		return json.Unmarshal(v, &state)
	})
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveProviderState saves a provider FSM state
func (s *Store) SaveProviderState(state *ProviderFSMState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal provider state: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketProviderStates)
		return b.Put([]byte(state.ID), data)
	})
}

// GetProviderState retrieves a provider FSM state by ID
func (s *Store) GetProviderState(id string) (*ProviderFSMState, error) {
	var state ProviderFSMState
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketProviderStates)
		v := b.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("provider state not found: %s", id)
		}
		return json.Unmarshal(v, &state)
	})
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// ListProviderStates returns all provider states
func (s *Store) ListProviderStates() ([]*ProviderFSMState, error) {
	var states []*ProviderFSMState
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketProviderStates)
		return b.ForEach(func(k, v []byte) error {
			var state ProviderFSMState
			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}
			states = append(states, &state)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return states, nil
}

// SaveCheckpoint saves a URN-based checkpoint
func (s *Store) SaveCheckpoint(checkpoint *Checkpoint) error {
	// Validate URN format
	if _, err := urn.Parse(checkpoint.URN); err != nil {
		return fmt.Errorf("invalid checkpoint URN: %w", err)
	}

	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpoints)
		// Use URN as key for easy lookup
		return b.Put([]byte(checkpoint.URN), data)
	})
}

// GetCheckpoint retrieves a checkpoint by URN
func (s *Store) GetCheckpoint(urnStr string) (*Checkpoint, error) {
	var checkpoint Checkpoint
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpoints)
		v := b.Get([]byte(urnStr))
		if v == nil {
			return fmt.Errorf("checkpoint not found: %s", urnStr)
		}
		return json.Unmarshal(v, &checkpoint)
	})
	if err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

// ListCheckpointsByProvider returns all checkpoints for a provider
func (s *Store) ListCheckpointsByProvider(providerID string) ([]*Checkpoint, error) {
	var checkpoints []*Checkpoint
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpoints)
		return b.ForEach(func(k, v []byte) error {
			var checkpoint Checkpoint
			if err := json.Unmarshal(v, &checkpoint); err != nil {
				return err
			}
			if checkpoint.ProviderID == providerID {
				checkpoints = append(checkpoints, &checkpoint)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return checkpoints, nil
}

// SavePermitAllocation saves a permit allocation
func (s *Store) SavePermitAllocation(allocation *PermitAllocation) error {
	data, err := json.Marshal(allocation)
	if err != nil {
		return fmt.Errorf("failed to marshal permit allocation: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketPermits)
		return b.Put([]byte(allocation.ProviderID), data)
	})
}

// GetPermitAllocation retrieves a permit allocation by provider ID
func (s *Store) GetPermitAllocation(providerID string) (*PermitAllocation, error) {
	var allocation PermitAllocation
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketPermits)
		v := b.Get([]byte(providerID))
		if v == nil {
			return fmt.Errorf("permit allocation not found: %s", providerID)
		}
		return json.Unmarshal(v, &allocation)
	})
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

// DeleteProviderState deletes a provider state
func (s *Store) DeleteProviderState(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketProviderStates)
		return b.Delete([]byte(id))
	})
}
