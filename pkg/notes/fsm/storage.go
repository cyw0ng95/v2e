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
