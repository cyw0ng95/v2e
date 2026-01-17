package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	// ErrSessionExists indicates that a session already exists
	ErrSessionExists = errors.New("session already exists")
	// ErrNoSession indicates that no session exists
	ErrNoSession = errors.New("no session exists")
	// ErrInvalidState indicates an invalid session state
	ErrInvalidState = errors.New("invalid session state")
)

// SessionState represents the state of a session
type SessionState string

const (
	// StateIdle means no jobs are running
	StateIdle SessionState = "idle"
	// StateRunning means jobs are actively running
	StateRunning SessionState = "running"
	// StatePaused means jobs are paused
	StatePaused SessionState = "paused"
	// StateStopped means the session has been stopped
	StateStopped SessionState = "stopped"
)

// Session represents a job execution session
type Session struct {
	ID              string       `json:"id"`
	State           SessionState `json:"state"`
	StartIndex      int          `json:"start_index"`       // Starting index for NVD fetching
	ResultsPerBatch int          `json:"results_per_batch"` // Number of results per batch
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	FetchedCount    int64        `json:"fetched_count"` // Total CVEs fetched
	StoredCount     int64        `json:"stored_count"`  // Total CVEs stored
	ErrorCount      int64        `json:"error_count"`   // Total errors encountered
}

// Manager manages session state in a bbolt database
type Manager struct {
	db         *bolt.DB
	bucketName []byte
}

// NewManager creates a new session manager
func NewManager(dbPath string) (*Manager, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open session database: %w", err)
	}

	bucketName := []byte("sessions")

	// Create bucket if it doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return &Manager{
		db:         db,
		bucketName: bucketName,
	}, nil
}

// CreateSession creates a new session
// Only one session is allowed at a time
func (m *Manager) CreateSession(sessionID string, startIndex, resultsPerBatch int) (*Session, error) {
	// Check if a session already exists
	existing, err := m.GetSession()
	if err != nil && err != ErrNoSession {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSessionExists
	}

	session := &Session{
		ID:              sessionID,
		State:           StateIdle,
		StartIndex:      startIndex,
		ResultsPerBatch: resultsPerBatch,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FetchedCount:    0,
		StoredCount:     0,
		ErrorCount:      0,
	}

	err = m.saveSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves the current session
func (m *Manager) GetSession() (*Session, error) {
	var session *Session

	err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(m.bucketName)
		if b == nil {
			return ErrNoSession
		}

		// Get the first (and only) session
		c := b.Cursor()
		k, v := c.First()
		if k == nil {
			return ErrNoSession
		}

		session = &Session{}
		return json.Unmarshal(v, session)
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

// UpdateState updates the session state
func (m *Manager) UpdateState(state SessionState) error {
	session, err := m.GetSession()
	if err != nil {
		return err
	}

	session.State = state
	session.UpdatedAt = time.Now()

	return m.saveSession(session)
}

// UpdateProgress updates the session progress counters
func (m *Manager) UpdateProgress(fetched, stored, errors int64) error {
	session, err := m.GetSession()
	if err != nil {
		return err
	}

	session.FetchedCount += fetched
	session.StoredCount += stored
	session.ErrorCount += errors
	session.UpdatedAt = time.Now()

	return m.saveSession(session)
}

// DeleteSession deletes the current session
func (m *Manager) DeleteSession() error {
	return m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(m.bucketName)
		if b == nil {
			return ErrNoSession
		}

		// Delete the first (and only) session
		c := b.Cursor()
		k, _ := c.First()
		if k == nil {
			return ErrNoSession
		}

		return b.Delete(k)
	})
}

// saveSession saves the session to the database
func (m *Manager) saveSession(session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(m.bucketName)
		if b == nil {
			return errors.New("bucket not found")
		}

		// Use session ID as key
		return b.Put([]byte(session.ID), data)
	})
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}
