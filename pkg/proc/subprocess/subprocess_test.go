package subprocess

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bytedance/sonic"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	sp := New("test-subprocess")
	if sp.ID != "test-subprocess" {
		t.Errorf("Expected ID to be 'test-subprocess', got '%s'", sp.ID)
	}
	if sp.handlers == nil {
		t.Error("Expected handlers map to be initialized")
	}
}

func TestRegisterHandler(t *testing.T) {
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
}

func TestSendMessage(t *testing.T) {
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
}

func TestSendResponse(t *testing.T) {
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
}

func TestSendEvent(t *testing.T) {
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
}

func TestHandleMessage(t *testing.T) {
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
			Payload: sonic.RawMessage(`{"result": "success"}`),
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
}

func TestRunWithMessages(t *testing.T) {
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
}

func TestUnmarshalPayload(t *testing.T) {
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
}

func TestUnmarshalPayload_NoPayload(t *testing.T) {
	msg := &Message{
		Type: MessageTypeEvent,
		ID:   "test",
	}

	var result map[string]string
	err := UnmarshalPayload(msg, &result)
	if err == nil {
		t.Error("Expected error when unmarshaling message with no payload")
	}
}
