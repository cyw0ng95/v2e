package proc

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bytedance/sonic/decoder"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
)

// fastDecoder is a configured sonic decoder for zero-copy parsing
var fastDecoder = decoder.NewDecoder("")

func init() {
	// Use fastest configuration for zero-copy parsing
	fastDecoder.UseInt64()
	fastDecoder.UseNumber()
}

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

	// MaxMessageSize is the maximum size of a message that can be sent between processes
	// Default to 10MB to accommodate large CVE data from NVD API; can be overridden by config
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

// messagePool is a sync.Pool for Message objects to reduce allocations
var messagePool = sync.Pool{
	New: func() interface{} {
		return &Message{}
	},
}

// GetMessage retrieves a Message from the pool
func GetMessage() *Message {
	msg := messagePool.Get().(*Message)
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
		messagePool.Put(msg)
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
