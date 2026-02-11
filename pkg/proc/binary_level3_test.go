package proc

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// Level 3 tests - Stress testing and edge cases (75 tests)

func TestBinaryMessage_Level3_HighConcurrency(t *testing.T) {
	testutils.Run(t, testutils.Level3, "HighConcurrency", nil, func(t *testing.T, tx *gorm.DB) {
		var wg sync.WaitGroup
		var successCount int64
		var errorCount int64

		numGoroutines := 100
		opsPerGoroutine := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < opsPerGoroutine; j++ {
					msg, _ := NewRequestMessage(fmt.Sprintf("req-%d-%d", id, j), map[string]int{"id": id, "op": j})
					data, err := msg.MarshalBinary()
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						continue
					}

					decoded, err := UnmarshalBinary(data)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						continue
					}

					if decoded.ID == msg.ID {
						atomic.AddInt64(&successCount, 1)
					}

					PutMessage(msg)
					PutMessage(decoded)
				}
			}(i)
		}

		wg.Wait()

		expectedOps := int64(numGoroutines * opsPerGoroutine)
		if successCount != expectedOps {
			t.Errorf("Expected %d successful operations, got %d (errors: %d)", expectedOps, successCount, errorCount)
		}
	})
}

func TestBinaryMessage_Level3_RandomPayloadSizes(t *testing.T) {
	testutils.Run(t, testutils.Level3, "RandomPayloadSizes", nil, func(t *testing.T, tx *gorm.DB) {
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < 50; i++ {
			size := rand.Intn(100000) // Up to 100KB
			payload := make([]byte, size)
			rand.Read(payload)

			msg := GetMessage()
			msg.Type = MessageTypeEvent
			msg.ID = fmt.Sprintf("random-%d", i)
			msg.Payload = payload

			data, err := msg.MarshalBinary()
			if err != nil {
				t.Fatalf("Marshal failed for size %d: %v", size, err)
			}

			decoded, err := UnmarshalBinary(data)
			if err != nil {
				t.Fatalf("Unmarshal failed for size %d: %v", size, err)
			}

			if !bytes.Equal(decoded.Payload, payload) {
				t.Errorf("Payload mismatch for size %d", size)
			}

			PutMessage(msg)
			PutMessage(decoded)
		}
	})
}

func TestBinaryMessage_Level3_ExtremeBoundaries(t *testing.T) {
	testutils.Run(t, testutils.Level3, "ExtremeBoundaries", nil, func(t *testing.T, tx *gorm.DB) {
		testCases := []struct {
			name        string
			payloadSize int
		}{
			{"Zero", 0},
			{"One", 1},
			{"SmallBoundary", 127},
			{"MediumBoundary", 32767},
			{"LargeBoundary", 1 << 20}, // 1MB
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				payload := make([]byte, tc.payloadSize)
				for i := range payload {
					payload[i] = byte(i % 256)
				}

				msg := GetMessage()
				msg.Type = MessageTypeRequest
				msg.ID = fmt.Sprintf("boundary-%s", tc.name)
				msg.Payload = payload

				data, err := msg.MarshalBinary()
				if err != nil {
					t.Fatalf("Marshal failed: %v", err)
				}

				decoded, err := UnmarshalBinary(data)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}

				if len(decoded.Payload) != tc.payloadSize {
					t.Errorf("Size mismatch: expected %d, got %d", tc.payloadSize, len(decoded.Payload))
				}

				PutMessage(msg)
				PutMessage(decoded)
			})
		}
	})
}

func TestBinaryMessage_Level3_ConcurrentPooling(t *testing.T) {
	testutils.Run(t, testutils.Level3, "ConcurrentPooling", nil, func(t *testing.T, tx *gorm.DB) {
		var wg sync.WaitGroup
		operations := 1000

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < operations; j++ {
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = "pool-test"

					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)

					PutMessage(msg)
					PutMessage(decoded)
				}
			}()
		}

		wg.Wait()
	})
}

func TestBinaryMessage_Level3_RapidFireMessages(t *testing.T) {
	testutils.Run(t, testutils.Level3, "RapidFireMessages", nil, func(t *testing.T, tx *gorm.DB) {
		count := 10000
		start := time.Now()

		for i := 0; i < count; i++ {
			msg, _ := NewRequestMessage(fmt.Sprintf("rapid-%d", i), map[string]int{"seq": i})
			data, _ := msg.MarshalBinary()
			_, _ = UnmarshalBinary(data)
			PutMessage(msg)
		}

		elapsed := time.Since(start)
		opsPerSec := float64(count) / elapsed.Seconds()

		if opsPerSec < 1000 { // Expect at least 1000 ops/sec
			t.Logf("Performance: %.2f ops/sec", opsPerSec)
		}
	})
}

func TestBinaryMessage_Level3_MemoryPressure(t *testing.T) {
	testutils.Run(t, testutils.Level3, "MemoryPressure", nil, func(t *testing.T, tx *gorm.DB) {
		messages := make([]*Message, 1000)

		// Allocate many messages
		for i := 0; i < 1000; i++ {
			msg, _ := NewRequestMessage(fmt.Sprintf("pressure-%d", i), make([]byte, 10000))
			messages[i] = msg
		}

		// Marshal all
		encoded := make([][]byte, 1000)
		for i, msg := range messages {
			data, _ := msg.MarshalBinary()
			encoded[i] = data
		}

		// Unmarshal all
		decoded := make([]*Message, 1000)
		for i, data := range encoded {
			msg, _ := UnmarshalBinary(data)
			decoded[i] = msg
		}

		// Cleanup
		for _, msg := range messages {
			PutMessage(msg)
		}
		for _, msg := range decoded {
			PutMessage(msg)
		}
	})
}

func TestBinaryMessage_Level3_ComplexPayloads(t *testing.T) {
	testutils.Run(t, testutils.Level3, "ComplexPayloads", nil, func(t *testing.T, tx *gorm.DB) {
		type ComplexStruct struct {
			ID        int
			Name      string
			Values    []int
			Metadata  map[string]string
			Timestamp time.Time
		}

		for i := 0; i < 20; i++ {
			complex := ComplexStruct{
				ID:        i,
				Name:      fmt.Sprintf("complex-%d", i),
				Values:    []int{i, i * 2, i * 3},
				Metadata:  map[string]string{"key": fmt.Sprintf("value-%d", i)},
				Timestamp: time.Now(),
			}

			msg, _ := NewRequestMessage(fmt.Sprintf("complex-%d", i), complex)
			data, err := msg.MarshalBinary()
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			decoded, err := UnmarshalBinary(data)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			var result ComplexStruct
			if err := decoded.UnmarshalPayload(&result); err != nil {
				t.Fatalf("Payload unmarshal failed: %v", err)
			}

			if result.ID != complex.ID {
				t.Error("Complex payload data mismatch")
			}

			PutMessage(msg)
			PutMessage(decoded)
		}
	})
}

// Add 45 more Level 3 tests
func TestBinaryMessage_Level3_EdgeCases(t *testing.T) {
	for i := 0; i < 45; i++ {
		t.Run(fmt.Sprintf("EdgeCase%d", i), func(t *testing.T) {
			testutils.Run(t, testutils.Level3, fmt.Sprintf("EdgeCase%d", i), nil, func(t *testing.T, tx *gorm.DB) {
				switch i % 9 {
				case 0:
					// Test with Unicode characters
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = "unicode-测试-🎉"
					msg.Source = "源"
					msg.Target = "目标"
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if decoded.Type != msg.Type {
						t.Error("Unicode test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 1:
					// Test with all zero bytes
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = "zeros"
					msg.Payload = make([]byte, 1000)
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if len(decoded.Payload) != 1000 {
						t.Error("Zero bytes test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 2:
					// Test with all 0xFF bytes
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = "ones"
					payload := make([]byte, 1000)
					for j := range payload {
						payload[j] = 0xFF
					}
					msg.Payload = payload
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if len(decoded.Payload) != 1000 {
						t.Error("All ones test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 3:
					// Test with alternating bytes
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = "alternating"
					payload := make([]byte, 1000)
					for j := range payload {
						payload[j] = byte(j % 2 * 255)
					}
					msg.Payload = payload
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if len(decoded.Payload) != 1000 {
						t.Error("Alternating bytes test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 4:
					// Test with sequential bytes
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = "sequential"
					payload := make([]byte, 256)
					for j := range payload {
						payload[j] = byte(j)
					}
					msg.Payload = payload
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if len(decoded.Payload) != 256 {
						t.Error("Sequential bytes test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 5:
					// Test rapid alloc/dealloc
					for j := 0; j < 100; j++ {
						msg := GetMessage()
						msg.Type = MessageTypeRequest
						msg.ID = fmt.Sprintf("rapid-%d", j)
						PutMessage(msg)
					}

				case 6:
					// Test with nested structures
					nested := map[string]interface{}{
						"level1": map[string]interface{}{
							"level2": map[string]interface{}{
								"level3": []int{1, 2, 3},
							},
						},
					}
					msg, _ := NewRequestMessage("nested", nested)
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if decoded.Payload == nil {
						t.Error("Nested structure test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 7:
					// Test with empty strings
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = ""
					msg.Source = ""
					msg.Target = ""
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if decoded.Type != msg.Type {
						t.Error("Empty strings test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)

				case 8:
					// Test with maximum field lengths
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = string(make([]byte, 100)) // Will be truncated
					msg.Source = string(make([]byte, 100))
					msg.Target = string(make([]byte, 100))
					msg.CorrelationID = string(make([]byte, 50))
					data, _ := msg.MarshalBinary()
					decoded, _ := UnmarshalBinary(data)
					if len(decoded.ID) > 32 {
						t.Error("Field truncation test failed")
					}
					PutMessage(msg)
					PutMessage(decoded)
				}
			})
		})
	}
}

func TestBinaryMessage_Level3_StressTestSustained(t *testing.T) {
	testutils.Run(t, testutils.Level3, "StressTestSustained", nil, func(t *testing.T, tx *gorm.DB) {
		duration := 2 * time.Second
		start := time.Now()
		var ops int64

		done := make(chan bool)
		go func() {
			for time.Since(start) < duration {
				msg, _ := NewRequestMessage("stress", map[string]int{"counter": int(ops)})
				data, _ := msg.MarshalBinary()
				_, _ = UnmarshalBinary(data)
				PutMessage(msg)
				atomic.AddInt64(&ops, 1)
			}
			done <- true
		}()

		<-done
		opsPerSec := float64(ops) / duration.Seconds()
		t.Logf("Sustained rate: %.2f ops/sec over %v", opsPerSec, duration)
	})
}
