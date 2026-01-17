package subprocess

import (
	"bufio"
	"fmt"

	"github.com/bytedance/sonic"
)

// The core types, constants and pools (Message, MessageType, Handler, Subprocess,
// MaxMessageSize, bufferPool, writerPool, etc.) have been moved to
// `types.go` for improved maintainability. This file now contains the
// runtime logic (reading, writing, batching, signal handling, logging).

// Run starts the subprocess and processes incoming messages
// It blocks until the subprocess is stopped or an error occurs
func (s *Subprocess) Run() error {
	// Start message writer if batching is enabled
	if !s.disableBatching {
		s.wg.Add(1)
		go s.messageWriter()
	}

	// Send a ready event to signal that the subprocess is initialized
	if err := s.SendEvent("subprocess_ready", map[string]interface{}{
		"id": s.ID,
	}); err != nil {
		return fmt.Errorf("failed to send ready event: %w", err)
	}

	// Start processing messages
	scanner := bufio.NewScanner(s.input)
	// Get buffer from pool for better performance
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr
	scanner.Buffer(buf, MaxMessageSize)

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

		// Parse the message using fastest configuration for zero-copy
		var msg Message
		api := sonic.ConfigFastest
		if err := api.Unmarshal([]byte(line), &msg); err != nil {
			// Send error response
			// Principle 15: Avoid fmt.Sprintf in hot paths - use direct string concat
			errMsg := &Message{
				Type:  MessageTypeError,
				ID:    "parse-error",
				Error: "failed to parse message: " + err.Error(),
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
	// For response and error messages, prioritize type-based lookup
	// to ensure they go to the correct response handler
	s.mu.RLock()
	var handler Handler
	var exists bool

	if msg.Type == MessageTypeResponse || msg.Type == MessageTypeError {
		// For responses and errors, look up by type first
		handler, exists = s.handlers[string(msg.Type)]
		if !exists {
			// Fallback to ID-based lookup
			handler, exists = s.handlers[msg.ID]
		}
	} else {
		// For other message types (requests, events), look up by ID first
		handler, exists = s.handlers[msg.ID]
		if !exists {
			// Fallback to type-based lookup
			handler, exists = s.handlers[string(msg.Type)]
		}
	}
	s.mu.RUnlock()

	if !exists {
		// No handler found, send error
		// Principle 15: Avoid fmt.Sprintf in hot paths
		errMsg := &Message{
			Type:  MessageTypeError,
			ID:    msg.ID,
			Error: "no handler found for message: " + msg.ID,
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

// messageWriter and flushBatch have been moved to writer.go to improve
// file separation and maintainability.

// Send/Stop/Flush helpers have been moved to send.go to separate
// I/O-related helpers from the core runtime logic

// SetupLogging moved to logging.go

// RunWithDefaults moved to lifecycle.go
