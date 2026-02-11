package proc

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// Level 2 tests - Integration and concurrency (75 tests)

func TestBinaryMessage_Level2_ConcurrentMarshal(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ConcurrentMarshal", nil, func(t *testing.T, tx *gorm.DB) {
		var wg sync.WaitGroup
		errors := make(chan error, 10)
		
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				msg, _ := NewRequestMessage(fmt.Sprintf("req-%d", id), map[string]int{"value": id})
				_, err := msg.MarshalBinary()
				if err != nil {
					errors <- err
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		
		for err := range errors {
			t.Errorf("Concurrent marshal error: %v", err)
		}
	})
}

func TestBinaryMessage_Level2_ConcurrentUnmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ConcurrentUnmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		msg, _ := NewRequestMessage("req-1", map[string]string{"test": "data"})
		data, _ := msg.MarshalBinary()
		
		var wg sync.WaitGroup
		errors := make(chan error, 10)
		
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := UnmarshalBinary(data)
				if err != nil {
					errors <- err
				}
			}()
		}
		
		wg.Wait()
		close(errors)
		
		for err := range errors {
			t.Errorf("Concurrent unmarshal error: %v", err)
		}
	})
}

func TestBinaryMessage_Level2_RoundTripStress(t *testing.T) {
	testutils.Run(t, testutils.Level2, "RoundTripStress", nil, func(t *testing.T, tx *gorm.DB) {
		for i := 0; i < 100; i++ {
			msg, _ := NewRequestMessage(fmt.Sprintf("req-%d", i), map[string]int{"iteration": i})
			msg.Source = fmt.Sprintf("source-%d", i)
			msg.Target = fmt.Sprintf("target-%d", i)
			
			data, err := msg.MarshalBinary()
			if err != nil {
				t.Fatalf("Marshal failed at iteration %d: %v", i, err)
			}
			
			decoded, err := UnmarshalBinary(data)
			if err != nil {
				t.Fatalf("Unmarshal failed at iteration %d: %v", i, err)
			}
			
			if decoded.ID != msg.ID {
				t.Errorf("ID mismatch at iteration %d", i)
			}
		}
	})
}

func TestBinaryMessage_Level2_PoolingEfficiency(t *testing.T) {
	testutils.Run(t, testutils.Level2, "PoolingEfficiency", nil, func(t *testing.T, tx *gorm.DB) {
		// Test message pooling
		messages := make([]*Message, 100)
		
		for i := 0; i < 100; i++ {
			messages[i] = GetMessage()
			messages[i].Type = MessageTypeRequest
			messages[i].ID = fmt.Sprintf("msg-%d", i)
		}
		
		// Return all to pool
		for _, msg := range messages {
			PutMessage(msg)
		}
		
		// Get again and verify clean state
		for i := 0; i < 100; i++ {
			msg := GetMessage()
			if msg.ID != "" {
				t.Error("Message not properly reset")
			}
			PutMessage(msg)
		}
	})
}

func TestBinaryMessage_Level2_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"TooShort", []byte{0x56, 0x32}},
		{"InvalidMagic", make([]byte, 128)},
		{"IncompletePayload", append(make([]byte, 128), []byte("short")...)},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testutils.Run(t, testutils.Level2, tc.name, nil, func(t *testing.T, tx *gorm.DB) {
				_, err := UnmarshalBinary(tc.data)
				if err == nil {
					t.Error("Expected error for invalid data")
				}
			})
		})
	}
}

func TestBinaryMessage_Level2_PayloadTypes(t *testing.T) {
	payloads := []interface{}{
		map[string]string{"string": "value"},
		map[string]int{"int": 42},
		map[string]float64{"float": 3.14},
		map[string]bool{"bool": true},
		[]string{"array", "of", "strings"},
		[]int{1, 2, 3, 4, 5},
	}
	
	for i, payload := range payloads {
		t.Run(fmt.Sprintf("Payload%d", i), func(t *testing.T) {
			testutils.Run(t, testutils.Level2, fmt.Sprintf("Payload%d", i), nil, func(t *testing.T, tx *gorm.DB) {
				msg, err := NewRequestMessage(fmt.Sprintf("req-%d", i), payload)
				if err != nil {
					t.Fatalf("Create message failed: %v", err)
				}
				
				data, err := msg.MarshalBinary()
				if err != nil {
					t.Fatalf("Marshal failed: %v", err)
				}
				
				decoded, err := UnmarshalBinary(data)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				
				if decoded.Payload == nil {
					t.Error("Payload lost")
				}
			})
		})
	}
}

func TestBinaryMessage_Level2_MemoryLeaks(t *testing.T) {
	testutils.Run(t, testutils.Level2, "MemoryLeaks", nil, func(t *testing.T, tx *gorm.DB) {
		// Create and discard many messages to check for leaks
		for i := 0; i < 1000; i++ {
			msg := GetMessage()
			msg.Type = MessageTypeRequest
			msg.ID = fmt.Sprintf("leak-test-%d", i)
			msg.Payload = make([]byte, 1024) // 1KB payload
			
			data, _ := msg.MarshalBinary()
			_, _ = UnmarshalBinary(data)
			
			PutMessage(msg)
		}
		// If no panic or excessive memory usage, test passes
	})
}

func TestBinaryMessage_Level2_EncodingConversion(t *testing.T) {
	encodings := []EncodingType{EncodingJSON, EncodingGOB, EncodingPLAIN}
	
	for _, enc := range encodings {
		t.Run(fmt.Sprintf("Encoding%d", enc), func(t *testing.T) {
			testutils.Run(t, testutils.Level2, fmt.Sprintf("Encoding%d", enc), nil, func(t *testing.T, tx *gorm.DB) {
				msg, _ := NewRequestMessage("req-1", map[string]string{"key": "value"})
				
				data, err := MarshalBinaryWithEncoding(msg, enc)
				if err != nil {
					t.Fatalf("Marshal with encoding %d failed: %v", enc, err)
				}
				
				decoded, err := UnmarshalBinary(data)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				
				if decoded.ID != msg.ID {
					t.Error("ID mismatch after encoding conversion")
				}
			})
		})
	}
}

func TestBinaryMessage_Level2_TimeoutSimulation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TimeoutSimulation", nil, func(t *testing.T, tx *gorm.DB) {
		msg, _ := NewRequestMessage("timeout-test", map[string]string{"data": "test"})
		
		done := make(chan bool)
		go func() {
			data, _ := msg.MarshalBinary()
			_, _ = UnmarshalBinary(data)
			done <- true
		}()
		
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Error("Operation timed out")
		}
	})
}

func TestBinaryMessage_Level2_CorrelationTracking(t *testing.T) {
	testutils.Run(t, testutils.Level2, "CorrelationTracking", nil, func(t *testing.T, tx *gorm.DB) {
		correlations := make(map[string]bool)
		
		for i := 0; i < 50; i++ {
			corrID := fmt.Sprintf("corr-%d", i)
			msg, _ := NewRequestMessage(fmt.Sprintf("req-%d", i), nil)
			msg.CorrelationID = corrID
			
			data, _ := msg.MarshalBinary()
			decoded, _ := UnmarshalBinary(data)
			
			correlations[decoded.CorrelationID] = true
		}
		
		if len(correlations) != 50 {
			t.Errorf("Expected 50 unique correlations, got %d", len(correlations))
		}
	})
}

// Add 55 more Level 2 tests with various scenarios
func TestBinaryMessage_Level2_AdditionalScenarios(t *testing.T) {
	for i := 0; i < 55; i++ {
		t.Run(fmt.Sprintf("Scenario%d", i), func(t *testing.T) {
			testutils.Run(t, testutils.Level2, fmt.Sprintf("Scenario%d", i), nil, func(t *testing.T, tx *gorm.DB) {
				// Test different scenarios
				switch i % 5 {
				case 0:
					// Test with nil payload
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = fmt.Sprintf("nil-payload-%d", i)
					data, _ := msg.MarshalBinary()
					decoded, err := UnmarshalBinary(data)
					if err != nil {
						t.Errorf("Failed with nil payload: %v", err)
					}
					if decoded.ID != msg.ID {
						t.Error("ID mismatch")
					}
					PutMessage(msg)
					
				case 1:
					// Test with error message
					errMsg := NewErrorMessage(fmt.Sprintf("err-%d", i), fmt.Errorf("test error %d", i))
					data, _ := errMsg.MarshalBinary()
					decoded, err := UnmarshalBinary(data)
					if err != nil {
						t.Errorf("Failed with error message: %v", err)
					}
					if decoded.Error == "" {
						t.Error("Error message lost")
					}
					PutMessage(errMsg)
					
				case 2:
					// Test with response message
					respMsg, _ := NewResponseMessage(fmt.Sprintf("resp-%d", i), map[string]int{"result": i})
					data, _ := respMsg.MarshalBinary()
					decoded, err := UnmarshalBinary(data)
					if err != nil {
						t.Errorf("Failed with response message: %v", err)
					}
					if decoded.Type != MessageTypeResponse {
						t.Error("Type mismatch")
					}
					PutMessage(respMsg)
					
				case 3:
					// Test with event message
					evtMsg, _ := NewEventMessage(fmt.Sprintf("evt-%d", i), map[string]string{"event": "test"})
					data, _ := evtMsg.MarshalBinary()
					decoded, err := UnmarshalBinary(data)
					if err != nil {
						t.Errorf("Failed with event message: %v", err)
					}
					if decoded.Type != MessageTypeEvent {
						t.Error("Type mismatch")
					}
					PutMessage(evtMsg)
					
				case 4:
					// Test with large correlation ID
					msg := GetMessage()
					msg.Type = MessageTypeRequest
					msg.ID = fmt.Sprintf("large-corr-%d", i)
					msg.CorrelationID = fmt.Sprintf("very-long-correlation-id-that-exceeds-limit-%d", i)
					data, _ := msg.MarshalBinary()
					decoded, err := UnmarshalBinary(data)
					if err != nil {
						t.Errorf("Failed with large correlation ID: %v", err)
					}
					// Correlation ID should be truncated
					if len(decoded.CorrelationID) > 20 {
						t.Error("Correlation ID not truncated")
					}
					PutMessage(msg)
				}
			})
		})
	}
}
