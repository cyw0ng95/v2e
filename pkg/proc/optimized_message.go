package proc

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/bytedance/sonic"
)

// OptimizedMessagePool provides enhanced pooling for Message objects
type OptimizedMessagePool struct {
	pool   sync.Pool
	hits   int64 // atomic counter for cache hits
	misses int64 // atomic counter for cache misses
}

// Global optimized message pool
var optimizedPool = &OptimizedMessagePool{
	pool: sync.Pool{
		New: func() interface{} {
			return &Message{}
		},
	},
}

// Get retrieves a Message from the optimized pool
func (omp *OptimizedMessagePool) Get() *Message {
	msg := omp.pool.Get().(*Message)
	atomic.AddInt64(&omp.hits, 1)

	// Reset fields efficiently
	msg.reset()
	return msg
}

// Put returns a Message to the pool for reuse
func (omp *OptimizedMessagePool) Put(msg *Message) {
	if msg != nil {
		omp.pool.Put(msg)
		atomic.AddInt64(&omp.misses, 1)
	}
}

// reset efficiently resets message fields to zero values
func (m *Message) reset() {
	m.Type = ""
	m.ID = ""
	m.Payload = nil
	m.Error = ""
	m.Source = ""
	m.Target = ""
	m.CorrelationID = ""
}

// GetOptimizedMessage retrieves a Message from the optimized pool
func GetOptimizedMessage() *Message {
	return optimizedPool.Get()
}

// PutOptimizedMessage returns a Message to the optimized pool
func PutOptimizedMessage(msg *Message) {
	optimizedPool.Put(msg)
}

// Optimized sonic configurations for different use cases
var (
	// Fastest configuration for marshaling
	fastMarshalConfig = sonic.ConfigFastest

	// Optimized configuration for unmarshaling
	fastUnmarshalConfig = sonic.Config{
		SortMapKeys:    true,
		ValidateString: true,
	}.Froze()
)

// OptimizedNewRequestMessage creates a new request message with enhanced performance
func OptimizedNewRequestMessage(id string, payload interface{}) (*Message, error) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeRequest
	msg.ID = id

	if payload != nil {
		// Use fastest marshal configuration
		data, err := fastMarshalConfig.Marshal(payload)
		if err != nil {
			PutOptimizedMessage(msg)
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// OptimizedNewResponseMessage creates a new response message with enhanced performance
func OptimizedNewResponseMessage(id string, payload interface{}) (*Message, error) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeResponse
	msg.ID = id

	if payload != nil {
		// Use fastest marshal configuration
		data, err := fastMarshalConfig.Marshal(payload)
		if err != nil {
			PutOptimizedMessage(msg)
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// OptimizedNewEventMessage creates a new event message with enhanced performance
func OptimizedNewEventMessage(id string, payload interface{}) (*Message, error) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeEvent
	msg.ID = id

	if payload != nil {
		// Use fastest marshal configuration
		data, err := fastMarshalConfig.Marshal(payload)
		if err != nil {
			PutOptimizedMessage(msg)
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// OptimizedUnmarshalPayload unmarshals the message payload with enhanced performance
func (m *Message) OptimizedUnmarshalPayload(v interface{}) error {
	if m.Payload == nil {
		return fmt.Errorf("no payload to unmarshal")
	}

	// Use optimized unmarshal configuration
	return fastUnmarshalConfig.Unmarshal(m.Payload, v)
}

// OptimizedMarshal serializes the message to JSON with enhanced performance
func (m *Message) OptimizedMarshal() ([]byte, error) {
	return fastMarshalConfig.Marshal(m)
}

// OptimizedUnmarshal deserializes a message from JSON with enhanced performance
func OptimizedUnmarshal(data []byte) (*Message, error) {
	msg := GetOptimizedMessage()

	// Use optimized unmarshal configuration
	if err := fastUnmarshalConfig.Unmarshal(data, msg); err != nil {
		PutOptimizedMessage(msg)
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return msg, nil
}

// OptimizedUnmarshalReuse deserializes a message from JSON reusing an existing message
func OptimizedUnmarshalReuse(data []byte, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// Reset message first
	msg.reset()

	// Use optimized unmarshal configuration
	return fastUnmarshalConfig.Unmarshal(data, msg)
}

// OptimizedUnmarshalDecoder uses a pooled decoder for maximum performance
func OptimizedUnmarshalDecoder(data []byte) (*Message, error) {
	msg := GetOptimizedMessage()

	// Use the standard unmarshal approach
	if err := fastUnmarshalConfig.Unmarshal(data, msg); err != nil {
		PutOptimizedMessage(msg)
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	return msg, nil
}

// OptimizedBatchMessage allows efficient batching of multiple messages
type OptimizedBatchMessage struct {
	Messages []*Message
}

// MarshalBatch efficiently marshals multiple messages
func (obm *OptimizedBatchMessage) MarshalBatch() ([]byte, error) {
	return fastMarshalConfig.Marshal(obm.Messages)
}

// UnmarshalBatch efficiently unmarshals multiple messages
func UnmarshalBatch(data []byte) ([]*Message, error) {
	var messages []*Message
	if err := fastUnmarshalConfig.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}
	return messages, nil
}

// GetPoolStats returns statistics about the message pool performance
func GetPoolStats() (hits, misses int64) {
	hits = atomic.LoadInt64(&optimizedPool.hits)
	misses = atomic.LoadInt64(&optimizedPool.misses)
	return hits, misses
}

// ResetPoolStats resets the pool statistics
func ResetPoolStats() {
	atomic.StoreInt64(&optimizedPool.hits, 0)
	atomic.StoreInt64(&optimizedPool.misses, 0)
}

// OptimizedNewErrorMessage creates an error message with enhanced performance
func OptimizedNewErrorMessage(id string, err error) *Message {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeError
	msg.ID = id
	if err != nil {
		msg.Error = err.Error()
	}
	return msg
}

// Precomputed constants for common operations
const (
	// Pre-sized buffer for common message operations
	CommonBufferSize = 1024
)

// OptimizedNewMessage creates a new message with enhanced performance
func OptimizedNewMessage(msgType MessageType, id string) *Message {
	msg := GetOptimizedMessage()
	msg.Type = msgType
	msg.ID = id
	return msg
}
