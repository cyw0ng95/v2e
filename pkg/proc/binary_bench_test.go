package proc

import (
	"testing"
)

// BenchmarkBinaryHeader_Encode benchmarks encoding a binary header
func BenchmarkBinaryHeader_Encode(b *testing.B) {
	header := NewBinaryHeader()
	header.Encoding = EncodingJSON
	header.MsgType = BinaryMessageTypeRequest
	header.PayloadLen = 100
	header.SetMessageID("test-msg-id")
	header.SetSourceID("source")
	header.SetTargetID("target")
	header.SetCorrelationID("corr-123")

	buf := make([]byte, HeaderSize)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = header.EncodeHeader(buf)
	}
}

// BenchmarkBinaryHeader_Decode benchmarks decoding a binary header
func BenchmarkBinaryHeader_Decode(b *testing.B) {
	header := NewBinaryHeader()
	header.Encoding = EncodingJSON
	header.MsgType = BinaryMessageTypeRequest
	header.PayloadLen = 100
	header.SetMessageID("test-msg-id")
	header.SetSourceID("source")
	header.SetTargetID("target")
	header.SetCorrelationID("corr-123")

	buf := make([]byte, HeaderSize)
	_ = header.EncodeHeader(buf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeHeader(buf)
	}
}

// BenchmarkBinaryMessage_MarshalJSON benchmarks binary marshaling with JSON encoding
func BenchmarkBinaryMessage_MarshalJSON(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, _ := NewRequestMessage("req-1", payload)
	msg.Source = "source"
	msg.Target = "target"
	msg.CorrelationID = "corr-123"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = msg.MarshalBinary()
	}
}

// BenchmarkBinaryMessage_MarshalGOB benchmarks binary marshaling with GOB encoding
func BenchmarkBinaryMessage_MarshalGOB(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, _ := NewRequestMessage("req-1", payload)
	msg.Source = "source"
	msg.Target = "target"
	msg.CorrelationID = "corr-123"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalBinaryWithEncoding(msg, EncodingGOB)
	}
}

// BenchmarkBinaryMessage_UnmarshalJSON benchmarks binary unmarshaling with JSON encoding
func BenchmarkBinaryMessage_UnmarshalJSON(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, _ := NewRequestMessage("req-1", payload)
	msg.Source = "source"
	msg.Target = "target"
	msg.CorrelationID = "corr-123"

	data, _ := msg.MarshalBinary()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = UnmarshalBinary(data)
	}
}

// BenchmarkBinaryMessage_UnmarshalGOB benchmarks binary unmarshaling with GOB encoding
func BenchmarkBinaryMessage_UnmarshalGOB(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, _ := NewRequestMessage("req-1", payload)
	msg.Source = "source"
	msg.Target = "target"
	msg.CorrelationID = "corr-123"

	data, _ := MarshalBinaryWithEncoding(msg, EncodingGOB)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = UnmarshalBinary(data)
	}
}

// BenchmarkBinaryVsJSON_Marshal compares binary vs JSON marshaling
func BenchmarkBinaryVsJSON_Marshal(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, _ := NewRequestMessage("req-1", payload)
	msg.Source = "source-process"
	msg.Target = "target-process"
	msg.CorrelationID = "corr-123"

	b.Run("Binary", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = msg.MarshalBinary()
		}
	})

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = msg.Marshal()
		}
	})
}

// BenchmarkBinaryVsJSON_Unmarshal compares binary vs JSON unmarshaling
func BenchmarkBinaryVsJSON_Unmarshal(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	msg, _ := NewRequestMessage("req-1", payload)
	msg.Source = "source-process"
	msg.Target = "target-process"
	msg.CorrelationID = "corr-123"

	binaryData, _ := msg.MarshalBinary()
	jsonData, _ := msg.Marshal()

	b.Run("Binary", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = UnmarshalBinary(binaryData)
		}
	})

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Unmarshal(jsonData)
		}
	})
}

// BenchmarkBinaryVsJSON_RoundTrip compares full round-trip performance
func BenchmarkBinaryVsJSON_RoundTrip(b *testing.B) {
	type TestPayload struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	payload := TestPayload{
		Command: "echo",
		Args:    []string{"hello", "world"},
	}

	b.Run("Binary", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			msg, _ := NewRequestMessage("req-1", payload)
			msg.Source = "source"
			msg.Target = "target"
			msg.CorrelationID = "corr-123"

			data, _ := msg.MarshalBinary()
			decoded, _ := UnmarshalBinary(data)
			
			var result TestPayload
			_ = decoded.UnmarshalPayload(&result)
			
			PutMessage(msg)
			PutMessage(decoded)
		}
	})

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			msg, _ := NewRequestMessage("req-1", payload)
			msg.Source = "source"
			msg.Target = "target"
			msg.CorrelationID = "corr-123"

			data, _ := msg.Marshal()
			decoded, _ := Unmarshal(data)
			
			var result TestPayload
			_ = decoded.UnmarshalPayload(&result)
			
			PutMessage(msg)
		}
	})
}

// BenchmarkBinaryMessage_LargePayload benchmarks binary protocol with large payloads
func BenchmarkBinaryMessage_LargePayload(b *testing.B) {
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
	msg.Source = "source"
	msg.Target = "target"

	b.Run("BinaryMarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = msg.MarshalBinary()
		}
	})

	b.Run("JSONMarshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = msg.Marshal()
		}
	})
}
