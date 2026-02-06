package fsm

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Storage interface for FSM state persistence
type Storage interface {
	SaveMemoryFSMState(urn string, state *MemoryFSMState) error
	LoadMemoryFSMState(urn string) (*MemoryFSMState, error)
	SaveLearningFSMState(state *LearningFSMState) error
	LoadLearningFSMState() (*LearningFSMState, error)
}

// BoltDB buckets for FSM state storage
var (
	BucketMemoryFSMStates  = []byte("memory_fsm_states")
	BucketLearningFSMState = []byte("learning_fsm_state")
)

// BoltDBStorage implements FSM state persistence using BoltDB
type BoltDBStorage struct {
	db *bolt.DB
	mu sync.RWMutex
}

// NewBoltDBStorage creates a new BoltDB-based FSM storage
func NewBoltDBStorage(dbPath string) (*BoltDBStorage, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt db: %w", err)
	}

	// Create buckets
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(BucketMemoryFSMStates); err != nil {
			return fmt.Errorf("create bucket %s: %w", BucketMemoryFSMStates, err)
		}
		if _, err := tx.CreateBucketIfNotExists(BucketLearningFSMState); err != nil {
			return fmt.Errorf("create bucket %s: %w", BucketLearningFSMState, err)
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create buckets: %w", err)
	}

	return &BoltDBStorage{db: db}, nil
}

// Close closes the BoltDB database
func (s *BoltDBStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

// SaveMemoryFSMState saves an object's FSM state
func (s *BoltDBStorage) SaveMemoryFSMState(urn string, state *MemoryFSMState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMemoryFSMStates)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketMemoryFSMStates)
		}

		data, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("marshal state: %w", err)
		}

		return b.Put([]byte(urn), data)
	})
}

// LoadMemoryFSMState loads an object's FSM state
func (s *BoltDBStorage) LoadMemoryFSMState(urn string) (*MemoryFSMState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var state MemoryFSMState

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMemoryFSMStates)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketMemoryFSMStates)
		}

		data := b.Get([]byte(urn))
		if data == nil {
			return fmt.Errorf("state not found for URN: %s", urn)
		}

		return json.Unmarshal(data, &state)
	})

	if err != nil {
		return nil, err
	}

	return &state, nil
}

// DeleteMemoryFSMState deletes an object's FSM state
func (s *BoltDBStorage) DeleteMemoryFSMState(urn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMemoryFSMStates)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketMemoryFSMStates)
		}

		return b.Delete([]byte(urn))
	})
}

// SaveLearningFSMState saves the user's learning session state
func (s *BoltDBStorage) SaveLearningFSMState(state *LearningFSMState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketLearningFSMState)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketLearningFSMState)
		}

		data, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("marshal state: %w", err)
		}

		// Use a single key for learning state (one per user/session)
		return b.Put([]byte("current"), data)
	})
}

// LoadLearningFSMState restores the user's session state
func (s *BoltDBStorage) LoadLearningFSMState() (*LearningFSMState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var state LearningFSMState

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketLearningFSMState)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketLearningFSMState)
		}

		data := b.Get([]byte("current"))
		if data == nil {
			return fmt.Errorf("no learning state found")
		}

		return json.Unmarshal(data, &state)
	})

	if err != nil {
		return nil, err
	}

	return &state, nil
}

// ClearLearningFSMState clears the current learning session state
func (s *BoltDBStorage) ClearLearningFSMState() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketLearningFSMState)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketLearningFSMState)
		}

		return b.Delete([]byte("current"))
	})
}

// Backup creates a backup of the FSM states to a file
func (s *BoltDBStorage) Backup(backupPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(backupPath, 0600)
	})
}

// GetAllMemoryFSMStates retrieves all memory FSM states (for debugging/export)
func (s *BoltDBStorage) GetAllMemoryFSMStates() (map[string]*MemoryFSMState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	states := make(map[string]*MemoryFSMState)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMemoryFSMStates)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketMemoryFSMStates)
		}

		return b.ForEach(func(k, v []byte) error {
			var state MemoryFSMState
			if err := json.Unmarshal(v, &state); err != nil {
				return fmt.Errorf("unmarshal state for %s: %w", string(k), err)
			}
			states[string(k)] = &state
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return states, nil
}

// ValidateMemoryFSMState validates the integrity of a single memory FSM state
func (s *BoltDBStorage) ValidateMemoryFSMState(urn string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMemoryFSMStates)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketMemoryFSMStates)
		}

		data := b.Get([]byte(urn))
		if data == nil {
			return fmt.Errorf("state not found for URN: %s", urn)
		}

		var state MemoryFSMState
		if err := json.Unmarshal(data, &state); err != nil {
			return fmt.Errorf("failed to unmarshal state: %w", err)
		}

		// Validate URN matches
		if state.URN != urn {
			return fmt.Errorf("state URN mismatch: expected %s, got %s", urn, state.URN)
		}

		// Validate state is a known memory state
		if !isValidMemoryState(state.State) {
			return fmt.Errorf("invalid memory state: %s", state.State)
		}

		// Validate timestamps
		if state.CreatedAt.IsZero() {
			return fmt.Errorf("created_at timestamp is zero")
		}
		if state.UpdatedAt.IsZero() {
			return fmt.Errorf("updated_at timestamp is zero")
		}

		// Validate state history is ordered by timestamp
		for i := 1; i < len(state.StateHistory); i++ {
			if state.StateHistory[i].Timestamp.Before(state.StateHistory[i-1].Timestamp) {
				return fmt.Errorf("state history not ordered: entry %d timestamp before entry %d", i, i-1)
			}
		}

		return nil
	})
}

// ValidateLearningFSMState validates the integrity of the learning FSM state
func (s *BoltDBStorage) ValidateLearningFSMState() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketLearningFSMState)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketLearningFSMState)
		}

		data := b.Get([]byte("current"))
		if data == nil {
			return fmt.Errorf("no learning state found")
		}

		var state LearningFSMState
		if err := json.Unmarshal(data, &state); err != nil {
			return fmt.Errorf("failed to unmarshal state: %w", err)
		}

		// Validate strategy is valid
		if !isValidLearningStrategy(state.CurrentStrategy) {
			return fmt.Errorf("invalid learning strategy: %s", state.CurrentStrategy)
		}

		// Validate timestamps
		if state.SessionStart.IsZero() {
			return fmt.Errorf("session_start timestamp is zero")
		}

		// Validate current item URN is valid if set
		if state.CurrentItemURN != "" {
			// URN should be non-empty and have valid format
			if len(state.CurrentItemURN) < 10 {
				return fmt.Errorf("current_item_urn appears invalid: %s", state.CurrentItemURN)
			}
		}

		// Validate viewed items URNs are not empty
		for _, urn := range state.ViewedItems {
			if urn == "" {
				return fmt.Errorf("viewed items contains empty URN")
			}
		}

		return nil
	})
}

// ValidateAllMemoryFSMStates validates all stored memory FSM states
func (s *BoltDBStorage) ValidateAllMemoryFSMStates() (map[string]error, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errors := make(map[string]error)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMemoryFSMStates)
		if b == nil {
			return fmt.Errorf("bucket %s not found", BucketMemoryFSMStates)
		}

		return b.ForEach(func(k, v []byte) error {
			urn := string(k)
			data := v

			var state MemoryFSMState
			if err := json.Unmarshal(data, &state); err != nil {
				errors[urn] = fmt.Errorf("failed to unmarshal state: %w", err)
				return nil
			}

			// Validate state fields
			if state.URN != urn {
				errors[urn] = fmt.Errorf("state URN mismatch: expected %s, got %s", urn, state.URN)
				return nil
			}

			if !isValidMemoryState(state.State) {
				errors[urn] = fmt.Errorf("invalid memory state: %s", state.State)
				return nil
			}

			if state.CreatedAt.IsZero() {
				errors[urn] = fmt.Errorf("created_at timestamp is zero")
				return nil
			}

			if state.UpdatedAt.IsZero() {
				errors[urn] = fmt.Errorf("updated_at timestamp is zero")
				return nil
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return errors, nil
}

// isValidMemoryState checks if a memory state string is valid
func isValidMemoryState(state MemoryState) bool {
	switch state {
	case MemoryStateDraft, MemoryStateNew, MemoryStateLearning, MemoryStateReviewed, MemoryStateLearned, MemoryStateMastered, MemoryStateArchived:
		return true
	default:
		return false
	}
}

// isValidLearningStrategy checks if a learning strategy string is valid
func isValidLearningStrategy(strategy string) bool {
	switch strategy {
	case "bfs", "dfs":
		return true
	default:
		return false
	}
}
