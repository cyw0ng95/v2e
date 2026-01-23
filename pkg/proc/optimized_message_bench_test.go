package proc

import (
	"encoding/json"
	"errors"
	"testing"
)

// BenchmarkGetOptimizedMessage benchmarks the optimized message pool
func BenchmarkGetOptimizedMessage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetOptimizedMessage()
	}
}

// BenchmarkOptimizedNewRequestMessage benchmarks the optimized request message creation
func BenchmarkOptimizedNewRequestMessage(b *testing.B) {
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
		_, _ = OptimizedNewRequestMessage("req-1", payload)
	}
}

// BenchmarkOptimizedNewResponseMessage benchmarks the optimized response message creation
func BenchmarkOptimizedNewResponseMessage(b *testing.B) {
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
		_, _ = OptimizedNewResponseMessage("resp-1", response)
	}
}

// BenchmarkOptimizedMessageMarshal benchmarks the optimized marshaling
func BenchmarkOptimizedMessageMarshal(b *testing.B) {
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeRequest
	msg.ID = "test-id"
	msg.Payload = json.RawMessage(`{"test":"value"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.OptimizedMarshal()
	}
	
	// Clean up
	PutOptimizedMessage(msg)
}

// BenchmarkOptimizedMessageUnmarshal benchmarks the optimized unmarshaling
func BenchmarkOptimizedMessageUnmarshal(b *testing.B) {
	data := []byte(`{"type":"request","id":"test-id","payload":{"key":"value"}}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = OptimizedUnmarshal(data)
	}
}

// BenchmarkOptimizedMessageUnmarshalPayload benchmarks the optimized payload unmarshaling
func BenchmarkOptimizedMessageUnmarshalPayload(b *testing.B) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestData{
		Name:  "test",
		Value: 123,
	}

	msg, _ := OptimizedNewRequestMessage("req-1", original)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result TestData
		_ = msg.OptimizedUnmarshalPayload(&result)
	}
}

// BenchmarkOptimizedMessageMarshalUnmarshalRoundTrip benchmarks a full optimized marshal/unmarshal cycle
func BenchmarkOptimizedMessageMarshalUnmarshalRoundTrip(b *testing.B) {
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
		msg, _ := OptimizedNewRequestMessage("req-1", original)
		data, _ := msg.OptimizedMarshal()
		decoded, _ := OptimizedUnmarshal(data)
		var result TestData
		_ = decoded.OptimizedUnmarshalPayload(&result)
		
		// Clean up
		PutOptimizedMessage(msg)
		PutOptimizedMessage(decoded)
	}
}

// BenchmarkOptimizedMessageMarshalLargePayload benchmarks optimized marshaling with a large payload
func BenchmarkOptimizedMessageMarshalLargePayload(b *testing.B) {
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

	msg, _ := OptimizedNewRequestMessage("req-1", largePayload)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.OptimizedMarshal()
	}
	
	// Clean up
	PutOptimizedMessage(msg)
}

// BenchmarkOptimizedNewErrorMessage benchmarks the optimized error message creation
func BenchmarkOptimizedNewErrorMessage(b *testing.B) {
	err := errors.New("test error")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = OptimizedNewErrorMessage("err-1", err)
	}
}

// BenchmarkOptimizedMessageWithRouting benchmarks optimized messages with routing information
func BenchmarkOptimizedMessageWithRouting(b *testing.B) {
	type TestPayload struct {
		Command string `json:"command"`
	}

	payload := TestPayload{
		Command: "test",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg, _ := OptimizedNewRequestMessage("req-1", payload)
		msg.Source = "process-a"
		msg.Target = "process-b"
		msg.CorrelationID = "corr-123"
		_, _ = msg.OptimizedMarshal()
		
		// Clean up
		PutOptimizedMessage(msg)
	}
}

// BenchmarkComparisonOriginalVsOptimized benchmarks original vs optimized message creation
func BenchmarkComparisonOriginalVsOptimized(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	b.Run("Original-NewRequestMessage", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewRequestMessage("req-1", payload)
		}
	})

	b.Run("Optimized-OptimizedNewRequestMessage", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = OptimizedNewRequestMessage("req-1", payload)
		}
	})
}

// BenchmarkComparisonOriginalVsOptimizedMarshal benchmarks original vs optimized marshaling
func BenchmarkComparisonOriginalVsOptimizedMarshal(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	origMsg, _ := NewRequestMessage("req-1", payload)
	optMsg, _ := OptimizedNewRequestMessage("req-1", payload)

	b.Run("Original-Marshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = origMsg.Marshal()
		}
	})

	b.Run("Optimized-OptimizedMarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = optMsg.OptimizedMarshal()
		}
	})
	
	// Clean up
	PutMessage(origMsg)
	PutOptimizedMessage(optMsg)
}

// BenchmarkComparisonOriginalVsOptimizedUnmarshal benchmarks original vs optimized unmarshaling
func BenchmarkComparisonOriginalVsOptimizedUnmarshal(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	origMsg, _ := NewRequestMessage("req-1", payload)
	data, _ := origMsg.Marshal()
	PutMessage(origMsg)

	b.Run("Original-Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Unmarshal(data)
		}
	})

	b.Run("Optimized-OptimizedUnmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = OptimizedUnmarshal(data)
		}
	})
}