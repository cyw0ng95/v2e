package dlq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ErrorCode represents standardized error codes for categorization
type ErrorCode string

const (
	// API error codes
	ErrorCodeAPITimeout     ErrorCode = "API_7000"
	ErrorCodeAPIRateLimit   ErrorCode = "API_7001"
	ErrorCodeAPIServerError ErrorCode = "API_7002"
	ErrorCodeAPIAuthFailed  ErrorCode = "API_7003"
	
	// Data error codes
	ErrorCodeDataCorrupted  ErrorCode = "DATA_8000"
	ErrorCodeDataNotFound   ErrorCode = "DATA_8001"
	ErrorCodeDataInvalid    ErrorCode = "DATA_8002"
	
	// System error codes
	ErrorCodeSystemOverload ErrorCode = "SYS_9000"
	ErrorCodeSystemResource ErrorCode = "SYS_9001"
	ErrorCodeSystemUnknown  ErrorCode = "SYS_9999"
)

// ReplayStrategy defines how failed messages should be retried
type ReplayStrategy int

const (
	// StrategyImmediate retries immediately
	StrategyImmediate ReplayStrategy = iota
	// StrategyExponentialBackoff uses exponential backoff
	StrategyExponentialBackoff
	// StrategyFixed uses fixed delay between retries
	StrategyFixed
	// StrategyManual requires manual intervention
	StrategyManual
	// StrategyDiscard discards the message (no retry)
	StrategyDiscard
)

// FailedMessage represents a message that failed processing
type FailedMessage struct {
	ID              string                 // Unique message ID
	OriginalPayload interface{}            // Original message content
	ErrorCode       ErrorCode              // Standardized error code
	ErrorMessage    string                 // Human-readable error
	FailedAt        time.Time              // When it failed
	RetryCount      int                    // Number of retry attempts
	MaxRetries      int                    // Maximum allowed retries
	Strategy        ReplayStrategy         // Replay strategy to use
	NextRetryAt     time.Time              // When to retry next
	Metadata        map[string]interface{} // Additional context
}

// DeadLetterQueue manages failed messages with intelligent replay
type DeadLetterQueue struct {
	mu       sync.RWMutex
	messages map[string]*FailedMessage // messageID -> FailedMessage
	
	// Strategy mappings: ErrorCode -> ReplayStrategy
	strategyMap map[ErrorCode]ReplayStrategy
	
	// Backoff configuration
	baseDelay     time.Duration // Base delay for exponential backoff
	maxDelay      time.Duration // Maximum delay cap
	backoffFactor float64       // Multiplier for each retry
	
	// Statistics
	stats DLQStats
}

// DLQStats tracks DLQ statistics
type DLQStats struct {
	TotalEnqueued   int64     // Total messages added to DLQ
	TotalReplayed   int64     // Total successful replays
	TotalDiscarded  int64     // Total discarded messages
	CurrentSize     int       // Current queue size
	LastEnqueueTime time.Time // Last enqueue timestamp
	LastReplayTime  time.Time // Last replay timestamp
}

// NewDeadLetterQueue creates a new DLQ with default configuration
func NewDeadLetterQueue() *DeadLetterQueue {
	dlq := &DeadLetterQueue{
		messages:      make(map[string]*FailedMessage),
		strategyMap:   make(map[ErrorCode]ReplayStrategy),
		baseDelay:     1 * time.Second,
		maxDelay:      5 * time.Minute,
		backoffFactor: 2.0,
	}
	
	// Configure default strategies based on error codes
	dlq.ConfigureStrategy(ErrorCodeAPITimeout, StrategyExponentialBackoff)
	dlq.ConfigureStrategy(ErrorCodeAPIRateLimit, StrategyExponentialBackoff)
	dlq.ConfigureStrategy(ErrorCodeAPIServerError, StrategyExponentialBackoff)
	dlq.ConfigureStrategy(ErrorCodeAPIAuthFailed, StrategyManual)
	dlq.ConfigureStrategy(ErrorCodeDataCorrupted, StrategyManual)
	dlq.ConfigureStrategy(ErrorCodeDataNotFound, StrategyFixed)
	dlq.ConfigureStrategy(ErrorCodeDataInvalid, StrategyDiscard)
	dlq.ConfigureStrategy(ErrorCodeSystemOverload, StrategyExponentialBackoff)
	dlq.ConfigureStrategy(ErrorCodeSystemResource, StrategyFixed)
	dlq.ConfigureStrategy(ErrorCodeSystemUnknown, StrategyManual)
	
	return dlq
}

// ConfigureStrategy sets the replay strategy for a specific error code
func (dlq *DeadLetterQueue) ConfigureStrategy(errCode ErrorCode, strategy ReplayStrategy) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	dlq.strategyMap[errCode] = strategy
}

// Enqueue adds a failed message to the DLQ
func (dlq *DeadLetterQueue) Enqueue(msg *FailedMessage) error {
	if msg == nil {
		return fmt.Errorf("cannot enqueue nil message")
	}
	
	if msg.ID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	// Set default values if not specified
	if msg.FailedAt.IsZero() {
		msg.FailedAt = time.Now()
	}
	
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 5 // Default max retries
	}
	
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]interface{})
	}
	
	// Determine strategy based on error code
	if strategy, exists := dlq.strategyMap[msg.ErrorCode]; exists {
		msg.Strategy = strategy
	} else {
		msg.Strategy = StrategyManual // Default to manual if unknown error
	}
	
	// Calculate next retry time based on strategy
	msg.NextRetryAt = dlq.calculateNextRetry(msg)
	
	dlq.messages[msg.ID] = msg
	dlq.stats.TotalEnqueued++
	dlq.stats.CurrentSize = len(dlq.messages)
	dlq.stats.LastEnqueueTime = time.Now()
	
	return nil
}

// calculateNextRetry determines when to retry based on strategy
func (dlq *DeadLetterQueue) calculateNextRetry(msg *FailedMessage) time.Time {
	now := time.Now()
	
	switch msg.Strategy {
	case StrategyImmediate:
		return now
		
	case StrategyExponentialBackoff:
		// Calculate exponential backoff: baseDelay * (factor ^ retryCount)
		delay := time.Duration(float64(dlq.baseDelay) * pow(dlq.backoffFactor, msg.RetryCount))
		if delay > dlq.maxDelay {
			delay = dlq.maxDelay
		}
		return now.Add(delay)
		
	case StrategyFixed:
		// Fixed 30 second delay
		return now.Add(30 * time.Second)
		
	case StrategyManual:
		// Never auto-retry, requires manual intervention
		return time.Time{} // Zero time indicates no auto-retry
		
	case StrategyDiscard:
		// Discard immediately
		return time.Time{}
		
	default:
		return now.Add(1 * time.Minute) // Default delay
	}
}

// GetReadyForReplay returns messages ready to be replayed
func (dlq *DeadLetterQueue) GetReadyForReplay() []*FailedMessage {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()
	
	now := time.Now()
	ready := make([]*FailedMessage, 0)
	
	for _, msg := range dlq.messages {
		// Skip if max retries exceeded
		if msg.RetryCount >= msg.MaxRetries {
			continue
		}
		
		// Skip manual and discard strategies
		if msg.Strategy == StrategyManual || msg.Strategy == StrategyDiscard {
			continue
		}
		
		// Check if ready to retry
		if !msg.NextRetryAt.IsZero() && (now.After(msg.NextRetryAt) || now.Equal(msg.NextRetryAt)) {
			ready = append(ready, msg)
		}
	}
	
	return ready
}

// MarkReplayed marks a message as successfully replayed and removes it from DLQ
func (dlq *DeadLetterQueue) MarkReplayed(messageID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	if _, exists := dlq.messages[messageID]; !exists {
		return fmt.Errorf("message %s not found in DLQ", messageID)
	}
	
	delete(dlq.messages, messageID)
	dlq.stats.TotalReplayed++
	dlq.stats.CurrentSize = len(dlq.messages)
	dlq.stats.LastReplayTime = time.Now()
	
	return nil
}

// MarkFailed updates a message after a failed replay attempt
func (dlq *DeadLetterQueue) MarkFailed(messageID string, newError error) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	msg, exists := dlq.messages[messageID]
	if !exists {
		return fmt.Errorf("message %s not found in DLQ", messageID)
	}
	
	msg.RetryCount++
	msg.ErrorMessage = newError.Error()
	
	// Check if max retries exceeded
	if msg.RetryCount >= msg.MaxRetries {
		msg.Strategy = StrategyManual // Switch to manual after exhausting retries
		msg.NextRetryAt = time.Time{}
	} else {
		// Recalculate next retry time
		msg.NextRetryAt = dlq.calculateNextRetry(msg)
	}
	
	return nil
}

// Discard removes a message from DLQ without replay
func (dlq *DeadLetterQueue) Discard(messageID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	
	if _, exists := dlq.messages[messageID]; !exists {
		return fmt.Errorf("message %s not found in DLQ", messageID)
	}
	
	delete(dlq.messages, messageID)
	dlq.stats.TotalDiscarded++
	dlq.stats.CurrentSize = len(dlq.messages)
	
	return nil
}

// GetStats returns current DLQ statistics
func (dlq *DeadLetterQueue) GetStats() DLQStats {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()
	
	return dlq.stats
}

// GetMessage retrieves a specific message from DLQ
func (dlq *DeadLetterQueue) GetMessage(messageID string) (*FailedMessage, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()
	
	msg, exists := dlq.messages[messageID]
	if !exists {
		return nil, fmt.Errorf("message %s not found in DLQ", messageID)
	}
	
	return msg, nil
}

// ListMessages returns all messages in DLQ
func (dlq *DeadLetterQueue) ListMessages() []*FailedMessage {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()
	
	messages := make([]*FailedMessage, 0, len(dlq.messages))
	for _, msg := range dlq.messages {
		messages = append(messages, msg)
	}
	
	return messages
}

// StartAutoReplay starts an automatic replay worker
func (dlq *DeadLetterQueue) StartAutoReplay(ctx context.Context, handler func(*FailedMessage) error, tickInterval time.Duration) {
	if tickInterval == 0 {
		tickInterval = 10 * time.Second // Default
	}
	
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
			
		case <-ticker.C:
			ready := dlq.GetReadyForReplay()
			
			for _, msg := range ready {
				if err := handler(msg); err != nil {
					dlq.MarkFailed(msg.ID, err)
				} else {
					dlq.MarkReplayed(msg.ID)
				}
			}
		}
	}
}

// pow calculates base^exp for float64
func pow(base float64, exp int) float64 {
	result := 1.0
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
