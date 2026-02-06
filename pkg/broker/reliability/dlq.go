package reliability

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"go.etcd.io/bbolt"
)

// DeadLetterQueue handles failed RPC messages
// Implements Requirement 9: RPC Dead Letter Queue
type DeadLetterQueue struct {
	mu       sync.RWMutex
	db       *bbolt.DB
	maxSize  int
	bucketName string
}

// DLQMessage represents a failed message in the dead letter queue
type DLQMessage struct {
	MessageID     string    `json:"message_id"`
	OriginalMsg   string    `json:"original_msg"` // JSON serialized proc.Message
	FailureReason string    `json:"failure_reason"`
	FailureTime   time.Time `json:"failure_time"`
	RetryCount    int       `json:"retry_count"`
	LastRetryTime time.Time `json:"last_retry_time,omitempty"`
}

// NewDeadLetterQueue creates a new dead letter queue
func NewDeadLetterQueue(dbPath string, maxSize int) (*DeadLetterQueue, error) {
	if maxSize <= 0 {
		maxSize = 10000 // Default max size
	}

	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open DLQ database: %w", err)
	}

	dlq := &DeadLetterQueue{
		db:         db,
		maxSize:    maxSize,
		bucketName: "dead_letters",
	}

	// Initialize bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(dlq.bucketName))
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create DLQ bucket: %w", err)
	}

	return dlq, nil
}

// Add adds a failed message to the dead letter queue
func (dlq *DeadLetterQueue) Add(msg *proc.Message, reason string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// Serialize original message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	dlqMsg := &DLQMessage{
		MessageID:     msg.ID,
		OriginalMsg:   string(msgBytes),
		FailureReason: reason,
		FailureTime:   time.Now(),
		RetryCount:    0,
	}

	// Serialize DLQ message
	dlqMsgBytes, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("failed to serialize DLQ message: %w", err)
	}

	// Save to database
	err = dlq.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		// Check size limit
		stats := bucket.Stats()
		if stats.KeyN >= dlq.maxSize {
			return fmt.Errorf("DLQ is full (size: %d)", stats.KeyN)
		}

		return bucket.Put([]byte(msg.ID), dlqMsgBytes)
	})

	return err
}

// Get retrieves a message from the dead letter queue
func (dlq *DeadLetterQueue) Get(messageID string) (*DLQMessage, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	var dlqMsg *DLQMessage

	err := dlq.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		data := bucket.Get([]byte(messageID))
		if data == nil {
			return fmt.Errorf("message not found in DLQ")
		}

		dlqMsg = &DLQMessage{}
		return json.Unmarshal(data, dlqMsg)
	})

	if err != nil {
		return nil, err
	}

	return dlqMsg, nil
}

// List returns all messages in the dead letter queue
func (dlq *DeadLetterQueue) List(offset, limit int) ([]*DLQMessage, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	messages := make([]*DLQMessage, 0, limit)

	err := dlq.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		cursor := bucket.Cursor()
		count := 0
		skipped := 0

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			// Skip offset
			if skipped < offset {
				skipped++
				continue
			}

			// Check limit
			if count >= limit {
				break
			}

			var dlqMsg DLQMessage
			if err := json.Unmarshal(v, &dlqMsg); err != nil {
				continue // Skip invalid entries
			}

			messages = append(messages, &dlqMsg)
			count++
		}

		return nil
	})

	return messages, err
}

// Replay attempts to replay a message from the DLQ
func (dlq *DeadLetterQueue) Replay(messageID string) (*proc.Message, error) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	var originalMsg *proc.Message

	err := dlq.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		data := bucket.Get([]byte(messageID))
		if data == nil {
			return fmt.Errorf("message not found in DLQ")
		}

		var dlqMsg DLQMessage
		if err := json.Unmarshal(data, &dlqMsg); err != nil {
			return fmt.Errorf("failed to unmarshal DLQ message: %w", err)
		}

		// Deserialize original message
		originalMsg = &proc.Message{}
		if err := json.Unmarshal([]byte(dlqMsg.OriginalMsg), originalMsg); err != nil {
			return fmt.Errorf("failed to unmarshal original message: %w", err)
		}

		// Update retry count
		dlqMsg.RetryCount++
		dlqMsg.LastRetryTime = time.Now()

		// Save updated DLQ message
		updatedData, err := json.Marshal(dlqMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal updated DLQ message: %w", err)
		}

		return bucket.Put([]byte(messageID), updatedData)
	})

	if err != nil {
		return nil, err
	}

	return originalMsg, nil
}

// Remove removes a message from the dead letter queue
func (dlq *DeadLetterQueue) Remove(messageID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	return dlq.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		return bucket.Delete([]byte(messageID))
	})
}

// Count returns the number of messages in the DLQ
func (dlq *DeadLetterQueue) Count() (int, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	var count int

	err := dlq.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		stats := bucket.Stats()
		count = stats.KeyN
		return nil
	})

	return count, err
}

// Clear removes all messages from the DLQ
func (dlq *DeadLetterQueue) Clear() error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	return dlq.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket([]byte(dlq.bucketName)); err != nil {
			return err
		}

		_, err := tx.CreateBucket([]byte(dlq.bucketName))
		return err
	})
}

// Close closes the dead letter queue database
func (dlq *DeadLetterQueue) Close() error {
	return dlq.db.Close()
}

// Stats returns statistics about the DLQ
type DLQStats struct {
	TotalMessages  int       `json:"total_messages"`
	OldestMessage  time.Time `json:"oldest_message,omitempty"`
	NewestMessage  time.Time `json:"newest_message,omitempty"`
	MaxRetries     int       `json:"max_retries"`
	AverageRetries float64   `json:"average_retries"`
}

// GetStats returns statistics about the DLQ
func (dlq *DeadLetterQueue) GetStats() (*DLQStats, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	stats := &DLQStats{}

	err := dlq.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dlq.bucketName))
		if bucket == nil {
			return fmt.Errorf("DLQ bucket not found")
		}

		bucketStats := bucket.Stats()
		stats.TotalMessages = bucketStats.KeyN

		if bucketStats.KeyN == 0 {
			return nil
		}

		cursor := bucket.Cursor()
		totalRetries := 0

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var dlqMsg DLQMessage
			if err := json.Unmarshal(v, &dlqMsg); err != nil {
				continue
			}

			// Track oldest/newest
			if stats.OldestMessage.IsZero() || dlqMsg.FailureTime.Before(stats.OldestMessage) {
				stats.OldestMessage = dlqMsg.FailureTime
			}
			if stats.NewestMessage.IsZero() || dlqMsg.FailureTime.After(stats.NewestMessage) {
				stats.NewestMessage = dlqMsg.FailureTime
			}

			// Track max retries
			if dlqMsg.RetryCount > stats.MaxRetries {
				stats.MaxRetries = dlqMsg.RetryCount
			}

			totalRetries += dlqMsg.RetryCount
		}

		if stats.TotalMessages > 0 {
			stats.AverageRetries = float64(totalRetries) / float64(stats.TotalMessages)
		}

		return nil
	})

	return stats, err
}
