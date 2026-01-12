package proc

import (
	"testing"
)

// BenchmarkMessageCreation measures the cost of creating a new message
func BenchmarkMessageCreation(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewRequestMessage("test-id", payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageMarshal measures the cost of marshaling a message
func BenchmarkMessageMarshal(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	msg, err := NewRequestMessage("test-id", payload)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageUnmarshal measures the cost of unmarshaling a message
func BenchmarkMessageUnmarshal(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	msg, err := NewRequestMessage("test-id", payload)
	if err != nil {
		b.Fatal(err)
	}
	data, err := msg.Marshal()
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Unmarshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageUnmarshalPayload measures the cost of unmarshaling message payload
func BenchmarkMessageUnmarshalPayload(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	msg, err := NewRequestMessage("test-id", payload)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		err := msg.UnmarshalPayload(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageRoundTrip measures the full cost of a message round-trip
// This simulates: create -> marshal -> unmarshal -> unmarshal payload
func BenchmarkMessageRoundTrip(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create request
		msg, err := NewRequestMessage("test-id", payload)
		if err != nil {
			b.Fatal(err)
		}
		
		// Marshal
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		
		// Unmarshal
		parsed, err := Unmarshal(data)
		if err != nil {
			b.Fatal(err)
		}
		
		// Unmarshal payload
		var result map[string]interface{}
		err = parsed.UnmarshalPayload(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeMessageRoundTrip measures round-trip for large messages (simulating CVE data)
func BenchmarkLargeMessageRoundTrip(b *testing.B) {
	// Create a large payload similar to CVE data
	descriptions := make([]map[string]string, 10)
	for i := 0; i < 10; i++ {
		descriptions[i] = map[string]string{
			"lang":  "en",
			"value": "This is a long description of a vulnerability that contains detailed information about the security issue, its impact, and potential mitigations. " + "Extra padding to make it larger. " + "More text here. " + "Even more text to simulate realistic CVE descriptions.",
		}
	}
	
	payload := map[string]interface{}{
		"id":           "CVE-2021-44228",
		"descriptions": descriptions,
		"published":    "2021-12-10T10:00:00.000",
		"modified":     "2021-12-15T15:30:00.000",
		"vulnStatus":   "Analyzed",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create request
		msg, err := NewRequestMessage("RPCGetCVE", payload)
		if err != nil {
			b.Fatal(err)
		}
		
		// Marshal
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		
		// Unmarshal
		parsed, err := Unmarshal(data)
		if err != nil {
			b.Fatal(err)
		}
		
		// Unmarshal payload
		var result map[string]interface{}
		err = parsed.UnmarshalPayload(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentMessageProcessing measures concurrent message processing
func BenchmarkConcurrentMessageProcessing(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create request
			msg, err := NewRequestMessage("test-id", payload)
			if err != nil {
				b.Fatal(err)
			}
			
			// Marshal
			data, err := msg.Marshal()
			if err != nil {
				b.Fatal(err)
			}
			
			// Unmarshal
			parsed, err := Unmarshal(data)
			if err != nil {
				b.Fatal(err)
			}
			
			// Unmarshal payload
			var result map[string]interface{}
			err = parsed.UnmarshalPayload(&result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMessageCopyCount counts the number of allocations in a message round-trip
// This helps identify how many copies are being made
// This benchmark simulates the OPTIMIZED path where bytes are kept throughout
func BenchmarkMessageCopyCount(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create request (1st marshal - payload to bytes)
		msg, err := NewRequestMessage("test-id", payload)
		if err != nil {
			b.Fatal(err)
		}
		
		// Marshal message (2nd marshal - entire message to bytes)
		data, err := msg.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		
		// OPTIMIZED: Write bytes directly (no string conversion)
		// This simulates: stdin.Write(data); stdin.Write([]byte{'\n'})
		// The scanner.Bytes() also avoids string conversion
		
		// Unmarshal message (copy - bytes to message struct)
		parsed, err := Unmarshal(data)
		if err != nil {
			b.Fatal(err)
		}
		
		// Unmarshal payload (copy - payload bytes to concrete type)
		var result map[string]interface{}
		err = parsed.UnmarshalPayload(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
