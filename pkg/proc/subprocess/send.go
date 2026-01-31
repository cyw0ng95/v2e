package subprocess

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/jsonutil"
)

// SendMessage sends a message to the broker via stdout
func (s *Subprocess) SendMessage(msg *Message) error {
	return s.sendMessage(msg)
}

// sendMessage is the internal method to send a message
// Uses lock-free channel-based batching for better performance
func (s *Subprocess) sendMessage(msg *Message) error {
	// Use shared fast marshal helper for performance
	data, err := jsonutil.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// If batching is disabled (for tests), write directly
	if s.disableBatching {
		// Zero-copy optimization: if payload is large, avoid extra copies by
		// writing marshaled data directly to the writer buffer and flushing.
		s.writeMu.Lock()
		defer s.writeMu.Unlock()

		// Use pooled bufio.Writer for efficient writes (Principle 13, 14)
		writer := writerPool.Get().(*bufio.Writer)
		writer.Reset(s.output)
		defer writerPool.Put(writer)

		// If the marshaled data is large, write directly (no additional joining)
		if len(data) >= zeroCopyThreshold {
			if _, err := writer.Write(data); err != nil {
				return fmt.Errorf("failed to write message: %w", err)
			}
			if err := writer.WriteByte('\n'); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
			if err := writer.Flush(); err != nil {
				return fmt.Errorf("failed to flush: %w", err)
			}
			return nil
		}

		// Small message path: write as before
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush: %w", err)
		}
		return nil
	}

	// Send to batching channel (lock-free)
	select {
	case s.outChan <- data:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// SendResponse sends a response message
func (s *Subprocess) SendResponse(id string, payload interface{}) error {
	var rawPayload json.RawMessage
	if payload != nil {
		// Use shared fast marshal helper for performance
		data, err := jsonutil.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = data
	}

	msg := &Message{
		Type:    MessageTypeResponse,
		ID:      id,
		Payload: rawPayload,
		Source:  s.ID,
	}
	return s.sendMessage(msg)
}

// SendEvent sends an event message
func (s *Subprocess) SendEvent(id string, payload interface{}) error {
	var rawPayload json.RawMessage
	if payload != nil {
		// Use shared fast marshal helper for performance
		data, err := jsonutil.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = data
	}

	msg := &Message{
		Type:    MessageTypeEvent,
		ID:      id,
		Payload: rawPayload,
		Source:  s.ID,
	}
	return s.sendMessage(msg)
}

// SendError sends an error message
func (s *Subprocess) SendError(id string, err error) error {
	msg := &Message{
		Type:  MessageTypeError,
		ID:    id,
		Error: err.Error(),
		Source:  s.ID,
	}
	return s.sendMessage(msg)
}

// Stop gracefully stops the subprocess
func (s *Subprocess) Stop() error {
	s.cancel()
	close(s.outChan) // Close channel to signal writer goroutine
	s.wg.Wait()
	return nil
}

// Flush ensures all pending messages are written
// Useful for testing or before shutdown
func (s *Subprocess) Flush() {
	// Just wait for the ticker to fire (at most 15ms)
	time.Sleep(defaultFlushInterval * 3)
}
