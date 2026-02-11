package subprocess

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestNew(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNew", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-subprocess")
		if sp.ID != "test-subprocess" {
			t.Errorf("Expected ID to be 'test-subprocess', got '%s'", sp.ID)
		}
		if sp.handlers == nil {
			t.Error("Expected handlers map to be initialized")
		}
	})

}

func TestNewWithFDs(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNewWithFDs", nil, func(t *testing.T, tx *gorm.DB) {
		t.Skip("Skipping FD pipe test - UDS-only transport")
	})

}

func TestRegisterHandler(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRegisterHandler", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		handler := func(ctx context.Context, msg *Message) (*Message, error) {
			return nil, nil
		}

		sp.RegisterHandler("test-pattern", handler)

		sp.mu.RLock()
		_, exists := sp.handlers["test-pattern"]
		sp.mu.RUnlock()

		if !exists {
			t.Error("Expected handler to be registered")
		}
	})

}

func TestSendMessage(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendMessage", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		msg := &Message{
			Type: MessageTypeEvent,
			ID:   "test-event",
		}

		if err := sp.SendMessage(msg); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Check that the message was written
		result := output.String()
		if result == "" {
			t.Error("Expected message to be written to output")
		}

		// Verify it's valid JSON
		var parsed Message
		// Remove trailing newline
		result = strings.TrimSpace(result)
		if err := sonic.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("Failed to parse output as JSON: %v", err)
		}

		if parsed.Type != MessageTypeEvent {
			t.Errorf("Expected type to be %s, got %s", MessageTypeEvent, parsed.Type)
		}
		if parsed.ID != "test-event" {
			t.Errorf("Expected ID to be 'test-event', got '%s'", parsed.ID)
		}
	})

}

func TestSendResponse(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendResponse", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		payload := map[string]string{"status": "ok"}
		if err := sp.SendResponse("resp-1", payload); err != nil {
			t.Fatalf("Failed to send response: %v", err)
		}

		// Parse the output
		result := strings.TrimSpace(output.String())
		var msg Message
		if err := sonic.Unmarshal([]byte(result), &msg); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if msg.Type != MessageTypeResponse {
			t.Errorf("Expected type to be response, got %s", msg.Type)
		}

		var receivedPayload map[string]string
		if err := UnmarshalPayload(&msg, &receivedPayload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		if receivedPayload["status"] != "ok" {
			t.Errorf("Expected status to be 'ok', got '%s'", receivedPayload["status"])
		}
	})

}

func TestSendEvent(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendEvent", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		payload := map[string]interface{}{"event": "started"}
		if err := sp.SendEvent("evt-1", payload); err != nil {
			t.Fatalf("Failed to send event: %v", err)
		}

		result := strings.TrimSpace(output.String())
		var msg Message
		if err := sonic.Unmarshal([]byte(result), &msg); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if msg.Type != MessageTypeEvent {
			t.Errorf("Expected type to be event, got %s", msg.Type)
		}
	})

}

func TestHandleMessage(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestHandleMessage", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Register a handler
		handlerCalled := false
		sp.RegisterHandler("request", func(ctx context.Context, msg *Message) (*Message, error) {
			handlerCalled = true
			return &Message{
				Type:    MessageTypeResponse,
				ID:      msg.ID,
				Payload: json.RawMessage(`{"result": "success"}`),
			}, nil
		})

		// Create a request message
		msg := &Message{
			Type: MessageTypeRequest,
			ID:   "req-1",
		}

		// Handle the message (simulate the Add that would be done in Run)
		sp.wg.Add(1)
		sp.handleMessage(msg)
		sp.wg.Wait()

		if !handlerCalled {
			t.Error("Expected handler to be called")
		}

		// Check that a response was sent
		result := output.String()
		if result == "" {
			t.Error("Expected response to be written")
		}
	})

}

func TestRunWithMessages(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRunWithMessages", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")

		// Create input with test messages
		input := `{"type":"request","id":"req-1"}
	{"type":"event","id":"evt-1"}
	`
		sp.SetInput(strings.NewReader(input))

		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Register handlers
		requestReceived := false
		sp.RegisterHandler("request", func(ctx context.Context, msg *Message) (*Message, error) {
			requestReceived = true
			return &Message{
				Type: MessageTypeResponse,
				ID:   msg.ID,
			}, nil
		})

		// Run in a goroutine with timeout
		done := make(chan error, 1)
		go func() {
			done <- sp.Run()
		}()

		// Wait for processing or timeout
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			sp.Stop()
			t.Fatal("Run timed out")
		}

		if !requestReceived {
			t.Error("Expected request to be received and handled")
		}
	})

}

func TestUnmarshalPayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshalPayload", nil, func(t *testing.T, tx *gorm.DB) {
		payload := map[string]string{"key": "value"}
		data, _ := sonic.Marshal(payload)

		msg := &Message{
			Type:    MessageTypeRequest,
			ID:      "test",
			Payload: data,
		}

		var result map[string]string
		if err := UnmarshalPayload(msg, &result); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		if result["key"] != "value" {
			t.Errorf("Expected 'value', got '%s'", result["key"])
		}
	})

}

func TestUnmarshalPayload_NoPayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshalPayload_NoPayload", nil, func(t *testing.T, tx *gorm.DB) {
		msg := &Message{
			Type: MessageTypeEvent,
			ID:   "test",
		}

		var result map[string]string
		err := UnmarshalPayload(msg, &result)
		if err == nil {
			t.Error("Expected error when unmarshaling message with no payload")
		}
	})

}

func TestSendError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendError", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		testErr := errors.New("test error message")
		if err := sp.SendError("error-1", testErr); err != nil {
			t.Fatalf("Failed to send error: %v", err)
		}

		result := strings.TrimSpace(output.String())
		var msg Message
		if err := sonic.Unmarshal([]byte(result), &msg); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if msg.Type != MessageTypeError {
			t.Errorf("Expected type to be error, got %s", msg.Type)
		}
		if msg.Error != testErr.Error() {
			t.Errorf("Expected error to be '%s', got '%s'", testErr.Error(), msg.Error)
		}
	})

}

func TestStop(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStop", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")

		// Start the message writer
		sp.wg.Add(1)
		go sp.messageWriter()

		// Stop should not error
		if err := sp.Stop(); err != nil {
			t.Errorf("Stop() returned error: %v", err)
		}

		// Context should be cancelled
		select {
		case <-sp.ctx.Done():
			// Expected
		default:
			t.Error("Context was not cancelled after Stop()")
		}
	})

}

func TestFlush(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFlush", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Send a message
		msg := &Message{
			Type: MessageTypeEvent,
			ID:   "test-event",
		}

		if err := sp.SendMessage(msg); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Flush should ensure the message is written
		sp.Flush()

		result := output.String()
		if result == "" {
			t.Error("Expected message to be written after flush")
		}
	})

}

func TestSetInput(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetInput", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		input := &bytes.Buffer{}

		sp.SetInput(input)

		if sp.input != input {
			t.Error("SetInput did not set the input stream")
		}
	})

}

func TestMessageBatching(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMessageBatching", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.output = output

		// Enable batching by NOT calling SetOutput (which disables it)
		// Start message writer
		sp.wg.Add(1)
		go sp.messageWriter()

		// Send multiple messages
		for i := 0; i < 5; i++ {
			msg := &Message{
				Type: MessageTypeEvent,
				ID:   "test-event",
			}
			if err := sp.sendMessage(msg); err != nil {
				t.Fatalf("Failed to send message %d: %v", i, err)
			}
		}

		// Wait for batching ticker. Small sleeps here are necessary to allow the
		// internal batching ticker to tick; keeping them short but non-zero.
		time.Sleep(15 * time.Millisecond)

		// Stop the subprocess and wait for writer to finish
		// This ensures all messages are flushed and no concurrent access to buffer
		sp.Stop()

		// Now safe to read from output buffer
		result := output.String()
		lines := strings.Split(strings.TrimSpace(result), "\n")

		if len(lines) < 5 {
			t.Errorf("Expected at least 5 messages, got %d", len(lines))
		}
	})

}

// TestNewWithUDS_ConnectionLeak verifies that connection is closed properly
// when all connection attempts fail, preventing file descriptor leaks
func TestNewWithUDS_ConnectionLeak(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewWithUDS_ConnectionLeak", nil, func(t *testing.T, tx *gorm.DB) {
		// Use a non-existent socket path to ensure all connection attempts fail
		nonExistentPath := "/tmp/test_nonexistent_socket_12345.sock"

		// Remove the file if it exists (cleanup from previous runs)
		_ = os.Remove(nonExistentPath)

		// Track original file descriptor count (approximate)
		// In a real scenario, repeated calls without proper cleanup would leak fds
		initialFDs := countOpenFiles(t)

		// Attempt to create subprocess with non-existent socket
		// This should fail after all retries and close the connection
		sp, err := NewWithUDS("test-leak", nonExistentPath)

		// Should return error
		if err == nil {
			t.Error("Expected error when connecting to non-existent socket")
			if sp != nil {
				_ = sp.Stop()
			}
		}

		// Verify error message contains expected information
		if err != nil && sp == nil {
			if !strings.Contains(err.Error(), "failed to connect") {
				t.Errorf("Error message should mention connection failure, got: %v", err)
			}
		}

		// Verify file descriptor count hasn't increased significantly
		// (allowing for some variance due to test infrastructure)
		finalFDs := countOpenFiles(t)
		fdDelta := finalFDs - initialFDs
		if fdDelta > 5 {
			// More than 5 extra fds suggests a leak
			t.Errorf("Potential file descriptor leak: %d extra fds after failed connection attempts", fdDelta)
		}
	})
}

// TestNewWithUDS_ConnectionLeakRepeated tests that repeated failed connection
// attempts don't leak file descriptors
func TestNewWithUDS_ConnectionLeakRepeated(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewWithUDS_ConnectionLeakRepeated", nil, func(t *testing.T, tx *gorm.DB) {
		nonExistentPath := "/tmp/test_repeated_socket_12345.sock"
		_ = os.Remove(nonExistentPath)

		initialFDs := countOpenFiles(t)

		// Attempt multiple failed connections
		for i := 0; i < 10; i++ {
			_, err := NewWithUDS(fmt.Sprintf("test-leak-%d", i), nonExistentPath)
			if err == nil {
				t.Errorf("Expected error on attempt %d", i)
			}
		}

		finalFDs := countOpenFiles(t)
		fdDelta := finalFDs - initialFDs

		// With proper cleanup, we shouldn't accumulate leaked fds
		// Each failed attempt has 3 retries, so 10 attempts = up to 30 connection attempts
		// But each should be properly closed
		if fdDelta > 20 {
			t.Errorf("File descriptor leak detected after repeated failed connections: %d fds accumulated", fdDelta)
		}
	})
}

// countOpenFiles counts the number of open file descriptors for the current process
// This is a best-effort check for connection leak detection
func countOpenFiles(t *testing.T) int {
	// Read /proc/self/fd directory (Linux-specific)
	fdDir := "/proc/self/fd"
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		// If /proc/self/fd is not available (e.g., on macOS during tests),
		// return 0 to skip this check
		return 0
	}
	return len(entries)
}
