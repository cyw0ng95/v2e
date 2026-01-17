package proc

import (
	"encoding/json"
	"errors"
	"github.com/bytedance/sonic"
	"testing"
)

func TestMessageType_Constants(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{MessageTypeRequest, "request"},
		{MessageTypeResponse, "response"},
		{MessageTypeEvent, "event"},
		{MessageTypeError, "error"},
	}

	for _, tt := range tests {
		t.Run(string(tt.msgType), func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.msgType))
			}
		})
	}
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage(MessageTypeRequest, "test-id")

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}
	if msg.Type != MessageTypeRequest {
		t.Errorf("Expected Type to be MessageTypeRequest, got %s", msg.Type)
	}
	if msg.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", msg.ID)
	}
}

func TestNewRequestMessage(t *testing.T) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, err := NewRequestMessage("req-1", payload)
	if err != nil {
		t.Fatalf("NewRequestMessage failed: %v", err)
	}

	if msg.Type != MessageTypeRequest {
		t.Errorf("Expected Type to be MessageTypeRequest, got %s", msg.Type)
	}
	if msg.ID != "req-1" {
		t.Errorf("Expected ID to be 'req-1', got %s", msg.ID)
	}

	var result TestPayload
	if err := msg.UnmarshalPayload(&result); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if result.Command != payload.Command {
		t.Errorf("Expected Command to be %s, got %s", payload.Command, result.Command)
	}
}

func TestNewRequestMessage_NilPayload(t *testing.T) {
	msg, err := NewRequestMessage("req-1", nil)
	if err != nil {
		t.Fatalf("NewRequestMessage with nil payload failed: %v", err)
	}

	if msg.Payload != nil {
		t.Error("Expected Payload to be nil")
	}
}

func TestNewResponseMessage(t *testing.T) {
	type TestResponse struct {
		Status string `json:"status"`
		Result int    `json:"result"`
	}

	response := TestResponse{
		Status: "success",
		Result: 42,
	}

	msg, err := NewResponseMessage("resp-1", response)
	if err != nil {
		t.Fatalf("NewResponseMessage failed: %v", err)
	}

	if msg.Type != MessageTypeResponse {
		t.Errorf("Expected Type to be MessageTypeResponse, got %s", msg.Type)
	}

	var result TestResponse
	if err := msg.UnmarshalPayload(&result); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if result.Status != response.Status {
		t.Errorf("Expected Status to be %s, got %s", response.Status, result.Status)
	}
	if result.Result != response.Result {
		t.Errorf("Expected Result to be %d, got %d", response.Result, result.Result)
	}
}

func TestNewEventMessage(t *testing.T) {
	type TestEvent struct {
		EventType string `json:"event_type"`
		Data      string `json:"data"`
	}

	event := TestEvent{
		EventType: "process_started",
		Data:      "pid:12345",
	}

	msg, err := NewEventMessage("evt-1", event)
	if err != nil {
		t.Fatalf("NewEventMessage failed: %v", err)
	}

	if msg.Type != MessageTypeEvent {
		t.Errorf("Expected Type to be MessageTypeEvent, got %s", msg.Type)
	}

	var result TestEvent
	if err := msg.UnmarshalPayload(&result); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if result.EventType != event.EventType {
		t.Errorf("Expected EventType to be %s, got %s", event.EventType, result.EventType)
	}
}

func TestNewErrorMessage(t *testing.T) {
	testErr := errors.New("test error occurred")
	msg := NewErrorMessage("err-1", testErr)

	if msg.Type != MessageTypeError {
		t.Errorf("Expected Type to be MessageTypeError, got %s", msg.Type)
	}
	if msg.ID != "err-1" {
		t.Errorf("Expected ID to be 'err-1', got %s", msg.ID)
	}
	if msg.Error != testErr.Error() {
		t.Errorf("Expected Error to be '%s', got '%s'", testErr.Error(), msg.Error)
	}
}

func TestNewErrorMessage_NilError(t *testing.T) {
	msg := NewErrorMessage("err-1", nil)

	if msg.Type != MessageTypeError {
		t.Errorf("Expected Type to be MessageTypeError, got %s", msg.Type)
	}
	if msg.Error != "" {
		t.Errorf("Expected Error to be empty, got '%s'", msg.Error)
	}
}

func TestMessage_UnmarshalPayload_NoPayload(t *testing.T) {
	msg := NewMessage(MessageTypeRequest, "test-id")

	var result map[string]interface{}
	err := msg.UnmarshalPayload(&result)

	if err == nil {
		t.Error("Expected error when unmarshaling nil payload")
	}
}

func TestMessage_Marshal(t *testing.T) {
	msg := NewMessage(MessageTypeRequest, "test-id")
	msg.Payload = json.RawMessage(`{"test":"value"}`)

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected marshaled data to be non-empty")
	}

	// Verify it contains the expected fields
	var result map[string]interface{}
	if err := sonic.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["type"] != string(MessageTypeRequest) {
		t.Errorf("Expected type to be %s, got %v", MessageTypeRequest, result["type"])
	}
	if result["id"] != "test-id" {
		t.Errorf("Expected id to be 'test-id', got %v", result["id"])
	}
}

func TestUnmarshal(t *testing.T) {
	data := []byte(`{"type":"request","id":"test-id","payload":{"key":"value"}}`)

	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.Type != MessageTypeRequest {
		t.Errorf("Expected Type to be MessageTypeRequest, got %s", msg.Type)
	}
	if msg.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", msg.ID)
	}

	var payload map[string]string
	if err := msg.UnmarshalPayload(&payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if payload["key"] != "value" {
		t.Errorf("Expected payload key to be 'value', got %s", payload["key"])
	}
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json}`)

	_, err := Unmarshal(data)
	if err == nil {
		t.Error("Expected error when unmarshaling invalid JSON")
	}
}

func TestMessage_MarshalUnmarshal_RoundTrip(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestData{
		Name:  "test",
		Value: 123,
	}

	msg, err := NewRequestMessage("req-1", original)
	if err != nil {
		t.Fatalf("NewRequestMessage failed: %v", err)
	}

	// Marshal the message
	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal it back
	decoded, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify the payload
	var result TestData
	if err := decoded.UnmarshalPayload(&result); err != nil {
		t.Fatalf("UnmarshalPayload failed: %v", err)
	}

	if result.Name != original.Name {
		t.Errorf("Expected Name to be %s, got %s", original.Name, result.Name)
	}
	if result.Value != original.Value {
		t.Errorf("Expected Value to be %d, got %d", original.Value, result.Value)
	}
}

func TestMarshalFast(t *testing.T) {
	msg := &Message{
		Type: MessageTypeRequest,
		ID:   "test-id",
	}

	data, err := msg.MarshalFast()
	if err != nil {
		t.Fatalf("MarshalFast failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalFast returned empty data")
	}

	// Verify it's valid JSON
	var parsed Message
	if err := sonic.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal fast-marshaled data: %v", err)
	}

	if parsed.Type != msg.Type {
		t.Errorf("Expected type %s, got %s", msg.Type, parsed.Type)
	}
}

func TestUnmarshalFast(t *testing.T) {
	original := &Message{
		Type: MessageTypeResponse,
		ID:   "resp-1",
	}

	data, err := sonic.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	msg, err := UnmarshalFast(data)
	if err != nil {
		t.Fatalf("UnmarshalFast failed: %v", err)
	}

	if msg.Type != original.Type {
		t.Errorf("Expected type %s, got %s", original.Type, msg.Type)
	}
	if msg.ID != original.ID {
		t.Errorf("Expected ID %s, got %s", original.ID, msg.ID)
	}

	// Message should be returned to pool
	PutMessage(msg)
}

func TestPutMessage(t *testing.T) {
	msg := GetMessage()
	msg.Type = MessageTypeEvent
	msg.ID = "event-1"

	// PutMessage should reset fields
	PutMessage(msg)

	// Get another message from pool
	msg2 := GetMessage()

	// Fields should be reset
	if msg2.Type != "" {
		t.Errorf("Expected Type to be empty after PutMessage, got %s", msg2.Type)
	}
	if msg2.ID != "" {
		t.Errorf("Expected ID to be empty after PutMessage, got %s", msg2.ID)
	}
}
