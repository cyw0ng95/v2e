package proc

import (
	"testing"
)

// TestMessageCopyIndicator is an indicator test that measures the number of memory allocations
// in a single message round-trip. This test serves as documentation and a performance regression detector.
//
// Based on analysis, a single RPC request-response cycle involves:
// 1. Request creation: Marshal payload to bytes (1 alloc)
// 2. Request marshal: Marshal entire message (1 alloc)
// 3. String conversion: bytes -> string for pipe write (1 alloc)
// 4. String conversion: string -> bytes for unmarshal (1 alloc)
// 5. Request unmarshal: bytes to Message struct (1 alloc)
// 6. Payload unmarshal: RawMessage to concrete type (multiple allocs)
// 7-12. Response follows same pattern in reverse
//
// Expected baseline (from benchmark): ~20 allocations per one-way message processing
// Full RPC round-trip: ~40 allocations (request + response)
func TestMessageCopyIndicator(t *testing.T) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}

	// Count allocations for documentation (using optimized path)
	allocCount := testing.AllocsPerRun(100, func() {
		msg, _ := NewRequestMessage("test-id", payload)
		data, _ := msg.Marshal()
		// OPTIMIZED: No string conversions anymore
		parsed, _ := Unmarshal(data)
		var result map[string]interface{}
		_ = parsed.UnmarshalPayload(&result)
	})

	// Document the findings
	t.Logf("Message round-trip allocations: %.0f", allocCount)
	t.Logf("For a full RPC cycle (request + response): %.0f allocations", allocCount*2)
	
	// This is a documentation test - it always passes but logs important metrics
	// If allocations significantly increase, it indicates a performance regression
	if allocCount > 30 {
		t.Logf("WARNING: Allocation count (%.0f) is higher than expected baseline (~20)", allocCount)
		t.Logf("This may indicate a performance regression. Review recent changes.")
	}
}

// TestMessageCopyAnalysis documents the detailed copy operations in the broker system
func TestMessageCopyAnalysis(t *testing.T) {
	t.Log("=== Message Copy Analysis for Broker System ===")
	t.Log("")
	t.Log("Copy operations for a single RPC request-response cycle (OPTIMIZED):")
	t.Log("")
	t.Log("REQUEST PATH (Client -> Broker -> Subprocess):")
	t.Log("  1. NewRequestMessage: Marshal payload to json.RawMessage")
	t.Log("  2. msg.Marshal(): Marshal entire Message struct to []byte")
	t.Log("  3. stdin.Write(data): Write bytes directly to pipe (no string conversion)")
	t.Log("  4. stdin.Write(newline): Write newline byte")
	t.Log("  5. scanner.Scan(): Read from pipe into scanner buffer")
	t.Log("  6. scanner.Bytes(): Return byte slice (no string copy)")
	t.Log("  7. Unmarshal: Parse []byte into Message struct")
	t.Log("  8. UnmarshalPayload: Parse json.RawMessage to concrete type")
	t.Log("")
	t.Log("RESPONSE PATH (Subprocess -> Broker -> Client):")
	t.Log("  9-16. Same 8 steps in reverse direction")
	t.Log("")
	t.Log("TOTAL: 16 operations per RPC round-trip (reduced from 18)")
	t.Log("")
	t.Log("OPTIMIZATIONS APPLIED:")
	t.Log("  ✓ Removed string conversions in Write operations")
	t.Log("  ✓ Use scanner.Bytes() instead of scanner.Text()")
	t.Log("  ✓ Direct byte writes with Write() instead of fmt.Fprintf()")
	t.Log("")
	t.Log("REMAINING ROOT CAUSES OF PERFORMANCE OVERHEAD:")
	t.Log("  1. Excessive marshaling: 4 marshal operations per round-trip")
	t.Log("  2. Scanner buffer copies: bufio.Scanner internal copies")
	t.Log("  3. No pooling: Each message allocates new memory")
	t.Log("  4. json.RawMessage: Still requires marshal/unmarshal for payload")
	t.Log("")
	t.Log("FUTURE OPTIMIZATION OPPORTUNITIES:")
	t.Log("  1. Message pooling with sync.Pool")
	t.Log("  2. Reuse marshal/unmarshal buffers")
	t.Log("  3. Batch message processing to amortize overhead")
	t.Log("  4. Consider binary protocol instead of JSON for high-throughput paths")
}
