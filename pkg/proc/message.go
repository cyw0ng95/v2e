package proc

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
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

// NewMessage creates a new message with the given type and ID
func NewMessage(msgType MessageType, id string) *Message {
	return &Message{
		Type: msgType,
		ID:   id,
	}
}

// NewRequestMessage creates a new request message
func NewRequestMessage(id string, payload interface{}) (*Message, error) {
	msg := NewMessage(MessageTypeRequest, id)
	if payload != nil {
		data, err := sonic.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// NewResponseMessage creates a new response message
func NewResponseMessage(id string, payload interface{}) (*Message, error) {
	msg := NewMessage(MessageTypeResponse, id)
	if payload != nil {
		data, err := sonic.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// NewEventMessage creates a new event message
func NewEventMessage(id string, payload interface{}) (*Message, error) {
	msg := NewMessage(MessageTypeEvent, id)
	if payload != nil {
		data, err := sonic.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		msg.Payload = data
	}
	return msg, nil
}

// NewErrorMessage creates a new error message
func NewErrorMessage(id string, err error) *Message {
	msg := NewMessage(MessageTypeError, id)
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
	return sonic.Unmarshal(m.Payload, v)
}

// Marshal serializes the message to JSON
func (m *Message) Marshal() ([]byte, error) {
	return sonic.Marshal(m)
}

// Unmarshal deserializes a message from JSON
func Unmarshal(data []byte) (*Message, error) {
	var msg Message
	if err := sonic.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &msg, nil
}
