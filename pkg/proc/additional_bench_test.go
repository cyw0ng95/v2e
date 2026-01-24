package proc

import (
	"encoding/json"
	"testing"
)

// BenchmarkMessagePoolOperations benchmarks the various operations of message pools
func BenchmarkMessagePoolOperations(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	b.Run("MessagePool-CreateAndReturn", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := GetMessage()
			msg.Type = MessageTypeRequest
			msg.ID = "test-id"
			PutMessage(msg)
		}
	})

	b.Run("OptimizedMessagePool-CreateAndReturn", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := GetOptimizedMessage()
			msg.Type = MessageTypeRequest
			msg.ID = "test-id"
			PutOptimizedMessage(msg)
		}
	})

	b.Run("MessagePool-PayloadHandling", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := GetMessage()
			_, _ = NewRequestMessage("req-1", payload)
			PutMessage(msg)
		}
	})

	b.Run("OptimizedMessagePool-PayloadHandling", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := GetOptimizedMessage()
			_, _ = OptimizedNewRequestMessage("req-1", payload)
			PutOptimizedMessage(msg)
		}
	})
}

// BenchmarkPayloadHandling benchmarks different payload sizes and types
func BenchmarkPayloadHandling(b *testing.B) {
	// Small payload
	smallPayload := map[string]string{
		"key": "value",
	}

	// Medium payload
	mediumPayload := make(map[string]interface{})
	for i := 0; i < 10; i++ {
		mediumPayload[formatString("key_%d", i)] = formatString("value_%d", i)
	}

	// Large payload
	largePayload := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		key := formatString("key_%d", i)
		largePayload[key] = map[string]interface{}{
			"id":          i,
			"name":        formatString("item_%d", i),
			"description": formatString("This is item number %d with detailed information", i),
			"values":      []int{i, i + 1, i + 2},
		}
	}

	b.Run("SmallPayload-Original", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewRequestMessage("req-small", smallPayload)
		}
	})

	b.Run("SmallPayload-Optimized", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = OptimizedNewRequestMessage("req-small", smallPayload)
		}
	})

	b.Run("MediumPayload-Original", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewRequestMessage("req-medium", mediumPayload)
		}
	})

	b.Run("MediumPayload-Optimized", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = OptimizedNewRequestMessage("req-medium", mediumPayload)
		}
	})

	b.Run("LargePayload-Original", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewRequestMessage("req-large", largePayload)
		}
	})

	b.Run("LargePayload-Optimized", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = OptimizedNewRequestMessage("req-large", largePayload)
		}
	})
}

// BenchmarkMarshalUnmarshalCycles benchmarks different cycles of marshal/unmarshal
func BenchmarkMarshalUnmarshalCycles(b *testing.B) {
	type TestData struct {
		Name     string            `json:"name"`
		Value    int               `json:"value"`
		Data     map[string]string `json:"data"`
		List     []string          `json:"list"`
		Nested   map[string]interface{} `json:"nested"`
	}

	testData := TestData{
		Name:  "test",
		Value: 42,
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		List: []string{"item1", "item2", "item3"},
		Nested: map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep_value",
			},
		},
	}

	b.Run("MessageRoundTrip-Original", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create message
			msg, _ := NewRequestMessage("roundtrip-test", testData)
			
			// Marshal
			data, _ := msg.Marshal()
			
			// Unmarshal
			received, _ := Unmarshal(data)
			
			// Extract payload
			var result TestData
			_ = received.UnmarshalPayload(&result)
			
			// Cleanup
			PutMessage(msg)
			PutMessage(received)
		}
	})

	b.Run("MessageRoundTrip-Optimized", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create message
			msg, _ := OptimizedNewRequestMessage("roundtrip-test", testData)
			
			// Marshal
			data, _ := msg.OptimizedMarshal()
			
			// Unmarshal
			received, _ := OptimizedUnmarshal(data)
			
			// Extract payload
			var result TestData
			_ = received.OptimizedUnmarshalPayload(&result)
			
			// Cleanup
			PutOptimizedMessage(msg)
			PutOptimizedMessage(received)
		}
	})

	b.Run("DirectJSONRoundTrip", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Direct marshal
			data, _ := json.Marshal(testData)
			
			// Direct unmarshal
			var result TestData
			_ = json.Unmarshal(data, &result)
		}
	})
}

// BenchmarkConcurrentAccess benchmarks concurrent access to message pools
func BenchmarkConcurrentAccess(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := GetMessage()
			msg.Type = MessageTypeRequest
			msg.ID = "concurrent-test"
			PutMessage(msg)
		}
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := GetOptimizedMessage()
			msg.Type = MessageTypeRequest
			msg.ID = "concurrent-test"
			PutOptimizedMessage(msg)
		}
	})
}

// Helper function for formatting strings (to avoid fmt.Sprintf overhead in benchmarks)
func formatString(format string, args ...interface{}) string {
	// Simple implementation for benchmark purposes
	if len(args) == 0 {
		return format
	}
	// Very simplified formatting for benchmark purposes
	if format == "key_%d" {
		if val, ok := args[0].(int); ok {
			return "key_" + itoa(val)
		}
	}
	if format == "value_%d" {
		if val, ok := args[0].(int); ok {
			return "value_" + itoa(val)
		}
	}
	if format == "item_%d" {
		if val, ok := args[0].(int); ok {
			return "item_" + itoa(val)
		}
	}
	if format == "This is item number %d with detailed information" {
		if val, ok := args[0].(int); ok {
			return "This is item number " + itoa(val) + " with detailed information"
		}
	}
	return format
}

// Simple integer to string conversion for benchmark purposes
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	
	s := make([]byte, 0, 10)
	for i > 0 {
		s = append(s, byte(i%10)+'0')
		i /= 10
	}
	
	// Reverse the string
	for j, k := 0, len(s)-1; j < k; j, k = j+1, k-1 {
		s[j], s[k] = s[k], s[j]
	}
	
	if neg {
		s = append(s, '-')
		// Reverse again to get correct order
		for j, k := 0, len(s)-1; j < k; j, k = j+1, k-1 {
			s[j], s[k] = s[k], s[j]
		}
	}
	
	return string(s)
}