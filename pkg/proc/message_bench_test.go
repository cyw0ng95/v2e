package proc

import (
	"encoding/json"
	"errors"
	"testing"
)

// BenchmarkNewMessage benchmarks the creation of a new message
func BenchmarkNewMessage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewMessage(MessageTypeRequest, "test-id")
	}
}

// BenchmarkNewRequestMessage benchmarks creating a request message with payload
func BenchmarkNewRequestMessage(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewRequestMessage("req-1", payload)
	}
}

// BenchmarkNewResponseMessage benchmarks creating a response message with payload
func BenchmarkNewResponseMessage(b *testing.B) {
	type TestResponse struct {
		Status string `json:"status"`
		Result int    `json:"result"`
	}

	response := TestResponse{
		Status: "success",
		Result: 42,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewResponseMessage("resp-1", response)
	}
}

// BenchmarkMessageMarshal benchmarks marshaling a message to JSON
func BenchmarkMessageMarshal(b *testing.B) {
	msg := NewMessage(MessageTypeRequest, "test-id")
	msg.Payload = json.RawMessage(`{"test":"value"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.Marshal()
	}
}

// BenchmarkMessageUnmarshal benchmarks unmarshaling JSON to a message
func BenchmarkMessageUnmarshal(b *testing.B) {
	data := []byte(`{"type":"request","id":"test-id","payload":{"key":"value"}}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Unmarshal(data)
	}
}

// BenchmarkMessageUnmarshalPayload benchmarks unmarshaling the payload from a message
func BenchmarkMessageUnmarshalPayload(b *testing.B) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestData{
		Name:  "test",
		Value: 123,
	}

	msg, _ := NewRequestMessage("req-1", original)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result TestData
		_ = msg.UnmarshalPayload(&result)
	}
}

// BenchmarkMessageMarshalUnmarshalRoundTrip benchmarks a full marshal/unmarshal cycle
func BenchmarkMessageMarshalUnmarshalRoundTrip(b *testing.B) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestData{
		Name:  "test",
		Value: 123,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg, _ := NewRequestMessage("req-1", original)
		data, _ := msg.Marshal()
		decoded, _ := Unmarshal(data)
		var result TestData
		_ = decoded.UnmarshalPayload(&result)
	}
}

// BenchmarkMessageMarshalLargePayload benchmarks marshaling with a large payload
func BenchmarkMessageMarshalLargePayload(b *testing.B) {
	// Create a large payload similar to a CVE response
	largePayload := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		key := string(rune('A'+i%26)) + string(rune('0'+i/26))
		largePayload[key] = map[string]interface{}{
			"id":          i,
			"description": "This is a long description that simulates a CVE entry with detailed information",
			"severity":    "HIGH",
			"published":   "2024-01-01T00:00:00Z",
			"references":  []string{"ref1", "ref2", "ref3", "ref4", "ref5"},
		}
	}

	msg, _ := NewRequestMessage("req-1", largePayload)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.Marshal()
	}
}

// BenchmarkNewErrorMessage benchmarks creating an error message
func BenchmarkNewErrorMessage(b *testing.B) {
	err := errors.New("test error")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewErrorMessage("err-1", err)
	}
}

// BenchmarkMessageWithRouting benchmarks messages with routing information
func BenchmarkMessageWithRouting(b *testing.B) {
	type TestPayload struct {
		Command string `json:"command"`
	}

	payload := TestPayload{
		Command: "test",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg, _ := NewRequestMessage("req-1", payload)
		msg.Source = "process-a"
		msg.Target = "process-b"
		msg.CorrelationID = "corr-123"
		_, _ = msg.Marshal()
	}
}
