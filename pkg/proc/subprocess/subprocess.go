package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
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
// This is a copy to avoid depending on the broker package
type Message struct {
	// Type is the type of message
	Type MessageType `json:"type"`
	// ID is a unique identifier for the message
	ID string `json:"id"`
	// Payload is the message data
	Payload json.RawMessage `json:"payload,omitempty"`
	// Error contains error information if Type is MessageTypeError
	Error string `json:"error,omitempty"`
}

// Handler is a function that handles incoming messages
type Handler func(ctx context.Context, msg *Message) (*Message, error)

// Subprocess represents a subprocess with a message-driven lifecycle
type Subprocess struct {
	// ID is the unique identifier for this subprocess
	ID string

	// handlers maps message IDs or patterns to handler functions
	handlers map[string]Handler

	// input is the input stream (typically stdin)
	input io.Reader

	// output is the output stream (typically stdout)
	output io.Writer

	// ctx is the context for the subprocess
	ctx context.Context

	// cancel is the cancel function for the context
	cancel context.CancelFunc

	// wg is the wait group for goroutines
	wg sync.WaitGroup

	// mu protects concurrent access
	mu sync.RWMutex
}

// New creates a new Subprocess instance
func New(id string) *Subprocess {
	ctx, cancel := context.WithCancel(context.Background())
	return &Subprocess{
		ID:       id,
		handlers: make(map[string]Handler),
		input:    os.Stdin,
		output:   os.Stdout,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SetInput sets the input stream for the subprocess
func (s *Subprocess) SetInput(r io.Reader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.input = r
}

// SetOutput sets the output stream for the subprocess
func (s *Subprocess) SetOutput(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = w
}

// RegisterHandler registers a handler for a specific message type or pattern
func (s *Subprocess) RegisterHandler(pattern string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[pattern] = handler
}

// Run starts the subprocess and processes incoming messages
// It blocks until the subprocess is stopped or an error occurs
func (s *Subprocess) Run() error {
	// Send a ready event to signal that the subprocess is initialized
	if err := s.SendEvent("subprocess_ready", map[string]interface{}{
		"id": s.ID,
	}); err != nil {
		return fmt.Errorf("failed to send ready event: %w", err)
	}

	// Start processing messages
	scanner := bufio.NewScanner(s.input)
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Send error response
			errMsg := &Message{
				Type:  MessageTypeError,
				ID:    "parse-error",
				Error: fmt.Sprintf("failed to parse message: %v", err),
			}
			_ = s.sendMessage(errMsg)
			continue
		}

		// Process the message
		s.wg.Add(1)
		go s.handleMessage(&msg)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	s.wg.Wait()
	return nil
}

// handleMessage processes a single message
func (s *Subprocess) handleMessage(msg *Message) {
	defer s.wg.Done()

	// Find the appropriate handler
	s.mu.RLock()
	handler, exists := s.handlers[msg.ID]
	if !exists {
		// Try to find a handler for the message type
		handler, exists = s.handlers[string(msg.Type)]
	}
	s.mu.RUnlock()

	if !exists {
		// No handler found, send error
		errMsg := &Message{
			Type:  MessageTypeError,
			ID:    msg.ID,
			Error: fmt.Sprintf("no handler found for message: %s", msg.ID),
		}
		_ = s.sendMessage(errMsg)
		return
	}

	// Call the handler
	response, err := handler(s.ctx, msg)
	if err != nil {
		// Send error response
		errMsg := &Message{
			Type:  MessageTypeError,
			ID:    msg.ID,
			Error: err.Error(),
		}
		_ = s.sendMessage(errMsg)
		return
	}

	// Send the response if provided
	if response != nil {
		_ = s.sendMessage(response)
	}
}

// SendMessage sends a message to the broker via stdout
func (s *Subprocess) SendMessage(msg *Message) error {
	return s.sendMessage(msg)
}

// sendMessage is the internal method to send a message
func (s *Subprocess) sendMessage(msg *Message) error {
	s.mu.RLock()
	output := s.output
	s.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write the message as a single line
	if _, err := fmt.Fprintf(output, "%s\n", string(data)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// SendResponse sends a response message
func (s *Subprocess) SendResponse(id string, payload interface{}) error {
	var rawPayload json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = data
	}

	msg := &Message{
		Type:    MessageTypeResponse,
		ID:      id,
		Payload: rawPayload,
	}
	return s.sendMessage(msg)
}

// SendEvent sends an event message
func (s *Subprocess) SendEvent(id string, payload interface{}) error {
	var rawPayload json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = data
	}

	msg := &Message{
		Type:    MessageTypeEvent,
		ID:      id,
		Payload: rawPayload,
	}
	return s.sendMessage(msg)
}

// SendError sends an error message
func (s *Subprocess) SendError(id string, err error) error {
	msg := &Message{
		Type:  MessageTypeError,
		ID:    id,
		Error: err.Error(),
	}
	return s.sendMessage(msg)
}

// Stop gracefully stops the subprocess
func (s *Subprocess) Stop() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

// UnmarshalPayload is a helper to unmarshal message payload
func UnmarshalPayload(msg *Message, v interface{}) error {
	if msg.Payload == nil {
		return fmt.Errorf("no payload to unmarshal")
	}
	return json.Unmarshal(msg.Payload, v)
}
