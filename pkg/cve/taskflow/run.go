package taskflow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	bolt "go.etcd.io/bbolt"
)

// JobRun represents a single job execution instance with full state
type JobRun struct {
	ID              string    `json:"id"`
	State           JobState  `json:"state"`
	StartIndex      int       `json:"start_index"`
	ResultsPerBatch int       `json:"results_per_batch"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	FetchedCount    int64     `json:"fetched_count"`
	StoredCount     int64     `json:"stored_count"`
	ErrorCount      int64     `json:"error_count"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

// RunStore manages persistent storage of job runs using BoltDB
type RunStore struct {
	db         *bolt.DB
	bucketName []byte
	logger     *common.Logger
}

// NewRunStore creates a new run store backed by BoltDB
func NewRunStore(dbPath string, logger *common.Logger) (*RunStore, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open run database: %w", err)
	}

	bucketName := []byte("job_runs")

	// Create bucket if it doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return &RunStore{
		db:         db,
		bucketName: bucketName,
		logger:     logger,
	}, nil
}

// CreateRun creates a new job run
func (s *RunStore) CreateRun(runID string, startIndex, resultsPerBatch int) (*JobRun, error) {
	run := &JobRun{
		ID:              runID,
		State:           StateQueued,
		StartIndex:      startIndex,
		ResultsPerBatch: resultsPerBatch,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FetchedCount:    0,
		StoredCount:     0,
		ErrorCount:      0,
	}

	if err := s.saveRun(run); err != nil {
		return nil, err
	}

	return run, nil
}

// GetRun retrieves a job run by ID
func (s *RunStore) GetRun(runID string) (*JobRun, error) {
	var run *JobRun

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		data := b.Get([]byte(runID))
		if data == nil {
			return fmt.Errorf("run not found: %s", runID)
		}

		run = &JobRun{}
		return json.Unmarshal(data, run)
	})

	if err != nil {
		return nil, err
	}

	return run, nil
}

// GetActiveRun retrieves the currently active run (running or paused)
// Returns nil if no active run exists
func (s *RunStore) GetActiveRun() (*JobRun, error) {
	var activeRun *JobRun

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		if b == nil {
			return nil
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var run JobRun
			if err := json.Unmarshal(v, &run); err != nil {
				continue
			}

			// Only running or paused runs are considered active
			if run.State == StateRunning || run.State == StatePaused {
				activeRun = &run
				return nil
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return activeRun, nil
}

// UpdateState updates the run state
func (s *RunStore) UpdateState(runID string, state JobState) error {
	run, err := s.GetRun(runID)
	if err != nil {
		return err
	}

	if !run.State.CanTransitionTo(state) {
		return fmt.Errorf("invalid state transition: %s -> %s", run.State, state)
	}

	run.State = state
	run.UpdatedAt = time.Now()
	s.logger.Debug("Updating run %s state to: %s", runID, state)

	return s.saveRun(run)
}

// UpdateProgress updates the run progress counters
func (s *RunStore) UpdateProgress(runID string, fetched, stored, errors int64) error {
	run, err := s.GetRun(runID)
	if err != nil {
		return err
	}

	run.FetchedCount += fetched
	run.StoredCount += stored
	run.ErrorCount += errors
	run.UpdatedAt = time.Now()

	return s.saveRun(run)
}

// SetError marks the run as failed with an error message
func (s *RunStore) SetError(runID string, errMsg string) error {
	run, err := s.GetRun(runID)
	if err != nil {
		return err
	}

	run.State = StateFailed
	run.ErrorMessage = errMsg
	run.UpdatedAt = time.Now()

	return s.saveRun(run)
}

// DeleteRun deletes a job run
func (s *RunStore) DeleteRun(runID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		return b.Delete([]byte(runID))
	})
}

// saveRun saves the run to the database
func (s *RunStore) saveRun(run *JobRun) error {
	data, err := json.Marshal(run)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		return b.Put([]byte(run.ID), data)
	})
}

// Close closes the database connection
func (s *RunStore) Close() error {
	return s.db.Close()
}
