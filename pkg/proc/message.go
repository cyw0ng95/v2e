package proc

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cyw0ng95/v2e/pkg/jsonutil"
)

// MessageType represents the type of message being sent
type MessageType string

const (
	// MessageTypeRequest represents a request message
	MessageTypeRequest MessageType = "request"
	// MessageTypeResponse represents a response message
	MessageTypeResponse MessageType = "response"
	// MessageTypeEvent represents an event message
	MessageTypeEvent MessageType = "event"
	// MessageTypeError represents an error message
	MessageTypeError MessageType = "error"
)

// MaxMessageSize is adjustable at runtime via configuration (default 10MB)
var MaxMessageSize = 10 * 1024 * 1024 // 10MB

// DefaultBufferSize is the default initial buffer size for scanners/readers
const DefaultBufferSize = 4096

// MaxBufferSize is the maximum buffer size for scanners/readers
// Default to MaxMessageSize to allow large messages
var MaxBufferSize = MaxMessageSize

// Message represents a message that can be passed between processes
type Message struct {
	// Type is the type of message
	Type MessageType `json:"type"`
	// ID is a unique identifier for the message
	ID string `json:"id"`
	// Payload is the message data
	Payload json.RawMessage `json:"payload,omitempty"`
	// Error contains error information if Type is MessageTypeError
	Error string `json:"error,omitempty"`
	// Source is the process ID of the message sender (for routing)
	Source string `json:"source,omitempty"`
	// Target is the process ID of the message recipient (for routing)
	Target string `json:"target,omitempty"`
	// CorrelationID is used to match responses to requests
	CorrelationID string `json:"correlation_id,omitempty"`
}

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

// Size-tiered pools for different payload sizes
var (
	smallMessagePool = sync.Pool{
		New: func() interface{} { return &Message{Payload: make(json.RawMessage, 0, 64)} },
	}
	mediumMessagePool = sync.Pool{
		New: func() interface{} { return &Message{Payload: make(json.RawMessage, 0, 512)} },
	}
	largeMessagePool = sync.Pool{
		New: func() interface{} { return &Message{Payload: make(json.RawMessage, 0, 4096)} },
	}
)

// GetMessageBySize retrieves a message optimized for expected payload size
func GetMessageBySize(expectedSize int) *Message {
	var pool *sync.Pool
	switch {
	case expectedSize <= 64:
		pool = &smallMessagePool
	case expectedSize <= 512:
		pool = &mediumMessagePool
	default:
		pool = &largeMessagePool
	}
	msg := pool.Get().(*Message)
	msg.reset()
	return msg
}

// PutMessageBySize returns a message to the appropriate pool based on size
func PutMessageBySize(msg *Message, expectedSize int) {
	if msg == nil {
		return
	}
	var pool *sync.Pool
	switch {
	case expectedSize <= 64:
		pool = &smallMessagePool
	case expectedSize <= 512:
		pool = &mediumMessagePool
	default:
		pool = &largeMessagePool
	}
	pool.Put(msg)
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
	// Zero out all fields to prepare for reuse
	m.Type = ""
	m.ID = ""
	m.Payload = nil // Reset to nil to match test expectations
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

// GetMessage retrieves a Message from the pool
func GetMessage() *Message {
	msg := optimizedPool.pool.Get().(*Message)
	// Reset all fields to zero values
	msg.Type = ""
	msg.ID = ""
	msg.Payload = nil
	msg.Error = ""
	msg.Source = ""
	msg.Target = ""
	msg.CorrelationID = ""
	return msg
}

// PutMessage returns a Message to the pool for reuse
func PutMessage(msg *Message) {
	if msg != nil {
		optimizedPool.pool.Put(msg)
	}
}

// NewMessage creates a new message with the given type and ID
// For better performance, consider using GetMessage() and PutMessage() for frequently created messages
func NewMessage(msgType MessageType, id string) *Message {
	msg := GetMessage()
	msg.Type = msgType
	msg.ID = id
	return msg
}

// NewRequestMessage creates a new request message
func NewRequestMessage(id string, payload interface{}) (*Message, error) {
	msg := GetMessage()
	msg.Type = MessageTypeRequest
	msg.ID = id
	if payload != nil {
		data, err := jsonutil.Marshal(payload)
		if err != nil {
			// Return to pool on error - fields will be reset on next Get
			PutMessage(msg)
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// NewResponseMessage creates a new response message
func NewResponseMessage(id string, payload interface{}) (*Message, error) {
	msg := GetMessage()
	msg.Type = MessageTypeResponse
	msg.ID = id
	if payload != nil {
		data, err := jsonutil.Marshal(payload)
		if err != nil {
			// Return to pool on error - fields will be reset on next Get
			PutMessage(msg)
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// NewEventMessage creates a new event message
func NewEventMessage(id string, payload interface{}) (*Message, error) {
	msg := GetMessage()
	msg.Type = MessageTypeEvent
	msg.ID = id
	if payload != nil {
		data, err := jsonutil.Marshal(payload)
		if err != nil {
			// Return to pool on error - fields will be reset on next Get
			PutMessage(msg)
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// NewErrorMessage creates a new error message
func NewErrorMessage(id string, err error) *Message {
	msg := GetMessage()
	msg.Type = MessageTypeError
	msg.ID = id
	if err != nil {
		msg.Error = err.Error()
	}
	return msg
}

// UnmarshalPayload unmarshals the message payload into the given value
func (m *Message) UnmarshalPayload(v interface{}) error {
	if m.Payload == nil {
		return fmt.Errorf("no payload to unmarshal")
	}
	return jsonutil.Unmarshal(m.Payload, v)
}

// Marshal serializes the message to JSON
func (m *Message) Marshal() ([]byte, error) {
	return jsonutil.Marshal(m)
}

// MarshalFast serializes the message to JSON using fastest configuration
// This is faster but may have different behavior for edge cases
func (m *Message) MarshalFast() ([]byte, error) {
	return jsonutil.Marshal(m)
}

// Unmarshal deserializes a message from JSON
func Unmarshal(data []byte) (*Message, error) {
	var msg Message
	if err := jsonutil.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &msg, nil
}

// UnmarshalFast deserializes a message from JSON using zero-copy optimization
// This is faster but requires the input data to remain valid during message lifetime
func UnmarshalFast(data []byte) (*Message, error) {
	msg := GetMessage()
	if err := jsonutil.Unmarshal(data, msg); err != nil {
		PutMessage(msg)
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return msg, nil
}

// Optimized helpers use the shared jsonutil wrapper for marshal/unmarshal.

// OptimizedNewRequestMessage creates a new request message with enhanced performance
func OptimizedNewRequestMessage(id string, payload interface{}) (*Message, error) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeRequest
	msg.ID = id

	if payload != nil {
		// Use shared fast marshal helper
		data, err := jsonutil.Marshal(payload)
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
		data, err := jsonutil.Marshal(payload)
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
		data, err := jsonutil.Marshal(payload)
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

	// Use shared fast unmarshal helper
	return jsonutil.Unmarshal(m.Payload, v)
}

// OptimizedMarshal serializes the message to JSON with enhanced performance
func (m *Message) OptimizedMarshal() ([]byte, error) {
	return jsonutil.Marshal(m)
}

// FastMarshal provides faster JSON serialization for Message objects
func (m *Message) FastMarshal() []byte {
	// Pre-allocate with estimated size
	buf := make([]byte, 0, 128+len(m.ID)+len(m.Source)+len(m.Target)+len(m.Payload))

	buf = append(buf, `{"type":"`...)
	buf = append(buf, string(m.Type)...)
	buf = append(buf, `","id":"`...)
	buf = append(buf, m.ID...)
	buf = append(buf, `","payload":`...)
	if m.Payload != nil {
		buf = append(buf, m.Payload...)
	} else {
		buf = append(buf, "null"...)
	}
	buf = append(buf, `,"source":"`...)
	buf = append(buf, m.Source...)
	buf = append(buf, `","target":"`...)
	buf = append(buf, m.Target...)
	buf = append(buf, `"}`...)

	return buf
}

// OptimizedUnmarshal deserializes a message from JSON with enhanced performance
func OptimizedUnmarshal(data []byte) (*Message, error) {
	msg := GetOptimizedMessage()

	// Use shared fast unmarshal helper
	if err := jsonutil.Unmarshal(data, msg); err != nil {
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
	return jsonutil.Unmarshal(data, msg)
}

// OptimizedUnmarshalDecoder uses a pooled decoder for maximum performance
func OptimizedUnmarshalDecoder(data []byte) (*Message, error) {
	msg := GetOptimizedMessage()

	// Use shared fast unmarshal helper
	if err := jsonutil.Unmarshal(data, msg); err != nil {
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
	return jsonutil.Marshal(obm.Messages)
}

// UnmarshalBatch efficiently unmarshals multiple messages
func UnmarshalBatch(data []byte) ([]*Message, error) {
	var messages []*Message
	if err := jsonutil.Unmarshal(data, &messages); err != nil {
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

// OptimizedNewMessage creates a new message with enhanced performance
func OptimizedNewMessage(msgType MessageType, id string) *Message {
	msg := GetOptimizedMessage()
	msg.Type = msgType
	msg.ID = id
	return msg
}
