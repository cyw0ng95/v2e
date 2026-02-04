package proc

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestMessageSerialization_Deserialization tests message serialization and deserialization with various payloads
func TestMessageSerialization_Deserialization(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageSerialization_Deserialization", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name     string
			payload  interface{}
			expected string
		}{
			{
				name:     "nil payload",
				payload:  nil,
				expected: "",
			},
			{
				name:     "string payload",
				payload:  "test string",
				expected: "test string",
			},
			{
				name:     "int payload",
				payload:  42,
				expected: "42",
			},
			{
				name:     "struct payload",
				payload:  struct{ Name string }{Name: "test"},
				expected: `{"Name":"test"}`,
			},
			{
				name:     "slice payload",
				payload:  []int{1, 2, 3},
				expected: "[1,2,3]",
			},
			{
				name:     "map payload",
				payload:  map[string]interface{}{"key": "value", "num": 123},
				expected: `{"key":"value","num":123}`,
			},
			{
				name:     "nested struct payload",
				payload:  struct{ Data map[string]int }{Data: map[string]int{"a": 1, "b": 2}},
				expected: `{"Data":{"a":1,"b":2}}`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test request message
				reqMsg, err := NewRequestMessage("req-test", tt.payload)
				if err != nil {
					t.Fatalf("NewRequestMessage failed: %v", err)
				}
				defer PutMessage(reqMsg)

				if reqMsg.Type != MessageTypeRequest {
					t.Errorf("Expected Type to be MessageTypeRequest, got %s", reqMsg.Type)
				}

				// Test response message
				respMsg, err := NewResponseMessage("resp-test", tt.payload)
				if err != nil {
					t.Fatalf("NewResponseMessage failed: %v", err)
				}
				defer PutMessage(respMsg)

				if respMsg.Type != MessageTypeResponse {
					t.Errorf("Expected Type to be MessageTypeResponse, got %s", respMsg.Type)
				}

				// Test event message
				eventMsg, err := NewEventMessage("event-test", tt.payload)
				if err != nil {
					t.Fatalf("NewEventMessage failed: %v", err)
				}
				defer PutMessage(eventMsg)

				if eventMsg.Type != MessageTypeEvent {
					t.Errorf("Expected Type to be MessageTypeEvent, got %s", eventMsg.Type)
				}

				// Test serialization and deserialization round trip
				if tt.payload != nil {
					// Marshal and unmarshal to verify payload integrity
					data, err := reqMsg.Marshal()
					if err != nil {
						t.Fatalf("Marshal failed: %v", err)
					}

					deserialized, err := Unmarshal(data)
					if err != nil {
						t.Fatalf("Unmarshal failed: %v", err)
					}
					defer PutMessage(deserialized)

					// Verify the payload can be unmarshaled correctly
					var result interface{}
					if err := deserialized.UnmarshalPayload(&result); err != nil {
						if tt.expected != "" {
							t.Errorf("UnmarshalPayload failed: %v", err)
						}
					} else {
						// Only check if we expect a non-empty result
						if tt.expected != "" {
							// Handle different types differently
							switch v := result.(type) {
							case string:
								if v != tt.expected {
									t.Errorf("Expected payload '%s', got '%s'", tt.expected, v)
								}
							case float64: // JSON numbers become float64
								expectedFloat := 0.0
								fmt.Sscanf(tt.expected, "%f", &expectedFloat)
								if v != expectedFloat {
									t.Errorf("Expected payload '%s', got '%f'", tt.expected, v)
								}
							default:
								// Convert result to JSON string for comparison
								resultBytes, _ := json.Marshal(result)
								if string(resultBytes) != tt.expected {
									t.Errorf("Expected payload %s, got %s", tt.expected, string(resultBytes))
								}
							}
						}
					}
				}
			})
		}
	})

}

// TestMessage_ErrorHandling tests error handling in message operations
func TestMessage_ErrorHandling(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_ErrorHandling", nil, func(t *testing.T, tx *gorm.DB) {
		// Test unmarshaling invalid JSON
		_, err := Unmarshal([]byte(`{invalid json`))
		if err == nil {
			t.Error("Expected error when unmarshaling invalid JSON")
		}

		// Test unmarshaling with UnmarshalFast
		_, err = UnmarshalFast([]byte(`{invalid json`))
		if err == nil {
			t.Error("Expected error when unmarshaling invalid JSON with UnmarshalFast")
		}

		// Test unmarshaling empty data
		_, err = Unmarshal([]byte{})
		if err == nil {
			t.Error("Expected error when unmarshaling empty data")
		}

		// Test unmarshaling empty data with UnmarshalFast
		_, err = UnmarshalFast([]byte{})
		if err == nil {
			t.Error("Expected error when unmarshaling empty data with UnmarshalFast")
		}

		// Test unmarshaling payload with nil payload
		msg := NewMessage(MessageTypeRequest, "test-id")
		var result interface{}
		err = msg.UnmarshalPayload(&result)
		if err == nil {
			t.Error("Expected error when unmarshaling nil payload")
		}

		// Test creating message with invalid payload (e.g., unserializable function)
		fn := func() {} // Functions can't be serialized to JSON
		_, err = NewRequestMessage("test", fn)
		if err == nil {
			t.Error("Expected error when creating message with unserializable payload")
		}

		_, err = NewResponseMessage("test", fn)
		if err == nil {
			t.Error("Expected error when creating response message with unserializable payload")
		}

		_, err = NewEventMessage("test", fn)
		if err == nil {
			t.Error("Expected error when creating event message with unserializable payload")
		}
	})

}

// TestConcurrentMessageProcessing tests concurrent message processing to ensure thread safety
func TestConcurrentMessageProcessing(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentMessageProcessing", nil, func(t *testing.T, tx *gorm.DB) {
		const numGoroutines = 20
		const messagesPerGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Channel to collect results and detect any panics
		results := make(chan error, numGoroutines*messagesPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					// Test message creation
					msg, err := NewRequestMessage(fmt.Sprintf("req-%d-%d", goroutineID, j), map[string]int{"id": j})
					if err != nil {
						results <- fmt.Errorf("NewRequestMessage failed: %v", err)
						continue
					}

					// Test marshaling
					data, err := msg.Marshal()
					if err != nil {
						results <- fmt.Errorf("Marshal failed: %v", err)
						PutMessage(msg)
						continue
					}

					// Test unmarshaling
					deserialized, err := Unmarshal(data)
					if err != nil {
						results <- fmt.Errorf("Unmarshal failed: %v", err)
						PutMessage(msg)
						continue
					}

					// Test payload unmarshaling
					var payload map[string]int
					err = deserialized.UnmarshalPayload(&payload)
					if err != nil {
						results <- fmt.Errorf("UnmarshalPayload failed: %v", err)
					}

					// Clean up
					PutMessage(msg)
					PutMessage(deserialized)

					// Send success
					results <- nil
				}
			}(i)
		}

		// Close results channel when all goroutines finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// Check for any errors
		errorCount := 0
		for err := range results {
			if err != nil {
				t.Errorf("Concurrent processing error: %v", err)
				errorCount++
			}
		}

		if errorCount > 0 {
			t.Errorf("Found %d errors in concurrent processing", errorCount)
		}
	})

}

// TestMessagePoolReuse tests that message objects are properly reused from the pool
func TestMessagePoolReuse(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessagePoolReuse", nil, func(t *testing.T, tx *gorm.DB) {
		// Get a message and set some fields
		msg1 := GetMessage()
		msg1.Type = MessageTypeRequest
		msg1.ID = "test-1"
		msg1.Source = "source-1"
		msg1.Target = "target-1"
		msg1.Error = "some-error"
		msg1.Payload = json.RawMessage(`{"test": true}`)

		// Put it back in the pool
		PutMessage(msg1)

		// Get another message - should be the same object with reset fields
		msg2 := GetMessage()

		// Fields should be reset to zero values
		if msg2.Type != "" {
			t.Errorf("Expected Type to be empty, got %s", msg2.Type)
		}
		if msg2.ID != "" {
			t.Errorf("Expected ID to be empty, got %s", msg2.ID)
		}
		if msg2.Source != "" {
			t.Errorf("Expected Source to be empty, got %s", msg2.Source)
		}
		if msg2.Target != "" {
			t.Errorf("Expected Target to be empty, got %s", msg2.Target)
		}
		if msg2.Error != "" {
			t.Errorf("Expected Error to be empty, got %s", msg2.Error)
		}
		if msg2.Payload != nil {
			t.Errorf("Expected Payload to be nil, got %v", msg2.Payload)
		}

		// Clean up
		PutMessage(msg2)
	})

}

// TestMessageMaxSize tests behavior with large messages
func TestMessageMaxSize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageMaxSize", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a payload that's larger than typical
		largePayload := make(map[string]string)
		for i := 0; i < 5000; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d_this_is_a_reasonably_long_string_to_increase_payload_size", i)
			largePayload[key] = value
		}

		// Test creating a request message with large payload
		reqMsg, err := NewRequestMessage("large-request", largePayload)
		if err != nil {
			t.Fatalf("Failed to create request message with large payload: %v", err)
		}
		defer PutMessage(reqMsg)

		// Test marshaling large message
		data, err := reqMsg.Marshal()
		if err != nil {
			t.Fatalf("Failed to marshal large message: %v", err)
		}

		// Verify the size is reasonable (not truncated)
		if len(data) < 10000 { // Should be much larger than 10KB
			t.Errorf("Expected large message to be substantial size, got %d bytes", len(data))
		}

		// Test unmarshaling large message
		deserialized, err := Unmarshal(data)
		if err != nil {
			t.Fatalf("Failed to unmarshal large message: %v", err)
		}
		defer PutMessage(deserialized)

		// Verify payload integrity
		var result map[string]string
		if err := deserialized.UnmarshalPayload(&result); err != nil {
			t.Fatalf("Failed to unmarshal large payload: %v", err)
		}

		if len(result) != len(largePayload) {
			t.Errorf("Expected %d keys in payload, got %d", len(largePayload), len(result))
		}
	})

}

// TestMessageRoutingFields tests the routing-related fields of messages
func TestMessageRoutingFields(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageRoutingFields", nil, func(t *testing.T, tx *gorm.DB) {
		msg := NewMessage(MessageTypeRequest, "test-id")
		msg.Source = "source-service"
		msg.Target = "target-service"
		msg.CorrelationID = "corr-12345"

		// Marshal and unmarshal to ensure routing fields are preserved
		data, err := msg.Marshal()
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		deserialized, err := Unmarshal(data)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		defer PutMessage(deserialized)

		if deserialized.Source != "source-service" {
			t.Errorf("Expected Source to be 'source-service', got '%s'", deserialized.Source)
		}
		if deserialized.Target != "target-service" {
			t.Errorf("Expected Target to be 'target-service', got '%s'", deserialized.Target)
		}
		if deserialized.CorrelationID != "corr-12345" {
			t.Errorf("Expected CorrelationID to be 'corr-12345', got '%s'", deserialized.CorrelationID)
		}

		// Test with empty routing fields
		msg2 := NewMessage(MessageTypeResponse, "resp-id")
		// Leave routing fields empty

		data2, err := msg2.Marshal()
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		deserialized2, err := Unmarshal(data2)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		defer PutMessage(deserialized2)

		if deserialized2.Source != "" {
			t.Errorf("Expected empty Source, got '%s'", deserialized2.Source)
		}
		if deserialized2.Target != "" {
			t.Errorf("Expected empty Target, got '%s'", deserialized2.Target)
		}
		if deserialized2.CorrelationID != "" {
			t.Errorf("Expected empty CorrelationID, got '%s'", deserialized2.CorrelationID)
		}
	})

}

// TestErrorMessageCreation tests error message creation and handling
func TestErrorMessageCreation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestErrorMessageCreation", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with regular error
		testErr := errors.New("this is a test error")
		errMsg := NewErrorMessage("error-1", testErr)
		defer PutMessage(errMsg)

		if errMsg.Type != MessageTypeError {
			t.Errorf("Expected Type to be MessageTypeError, got %s", errMsg.Type)
		}
		if errMsg.ID != "error-1" {
			t.Errorf("Expected ID to be 'error-1', got %s", errMsg.ID)
		}
		if errMsg.Error != "this is a test error" {
			t.Errorf("Expected Error to be 'this is a test error', got '%s'", errMsg.Error)
		}

		// Test with nil error
		errMsg2 := NewErrorMessage("error-2", nil)
		defer PutMessage(errMsg2)

		if errMsg2.Type != MessageTypeError {
			t.Errorf("Expected Type to be MessageTypeError, got %s", errMsg2.Type)
		}
		if errMsg2.ID != "error-2" {
			t.Errorf("Expected ID to be 'error-2', got %s", errMsg2.ID)
		}
		if errMsg2.Error != "" {
			t.Errorf("Expected Error to be empty, got '%s'", errMsg2.Error)
		}

		// Test with wrapped error
		wrappedErr := fmt.Errorf("wrapped: %w", testErr)
		errMsg3 := NewErrorMessage("error-3", wrappedErr)
		defer PutMessage(errMsg3)

		if errMsg3.Error != "wrapped: this is a test error" {
			t.Errorf("Expected Error to be 'wrapped: this is a test error', got '%s'", errMsg3.Error)
		}
	})

}

// TestMessageFieldValidation tests validation of message fields
func TestMessageFieldValidation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageFieldValidation", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with various special characters in ID
		specialIDs := []string{
			"normal-id",
			"id_with_underscores",
			"id-with-dashes",
			"id.with.dots",
			"id123numbers456",
			"IDMixedCase",
			"id!@#$%^&*()",
			"id with spaces",
			"", // empty ID
		}

		for _, id := range specialIDs {
			t.Run(fmt.Sprintf("ID_%s", id), func(t *testing.T) {
				msg := NewMessage(MessageTypeRequest, id)
				defer PutMessage(msg)

				if msg.ID != id {
					t.Errorf("Expected ID to be '%s', got '%s'", id, msg.ID)
				}

				// Test marshaling/unmarshaling with special ID
				data, err := msg.Marshal()
				if err != nil {
					t.Fatalf("Failed to marshal message with special ID: %v", err)
				}

				deserialized, err := Unmarshal(data)
				if err != nil {
					t.Fatalf("Failed to unmarshal message with special ID: %v", err)
				}
				defer PutMessage(deserialized)

				if deserialized.ID != id {
					t.Errorf("After marshal/unmarshal, expected ID to be '%s', got '%s'", id, deserialized.ID)
				}
			})
		}
	})

}

// TestMessageTypes tests all message types
func TestMessageTypes(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageTypes", nil, func(t *testing.T, tx *gorm.DB) {
		types := []MessageType{MessageTypeRequest, MessageTypeResponse, MessageTypeEvent, MessageTypeError}

		for _, msgType := range types {
			t.Run(string(msgType), func(t *testing.T) {
				// Test creating message with specific type
				msg := NewMessage(msgType, "test-id")
				defer PutMessage(msg)

				if msg.Type != msgType {
					t.Errorf("Expected Type to be %s, got %s", msgType, msg.Type)
				}

				// Test marshaling and unmarshaling preserves type
				data, err := msg.Marshal()
				if err != nil {
					t.Fatalf("Marshal failed: %v", err)
				}

				deserialized, err := Unmarshal(data)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				defer PutMessage(deserialized)

				if deserialized.Type != msgType {
					t.Errorf("After marshal/unmarshal, expected Type to be %s, got %s", msgType, deserialized.Type)
				}
			})
		}
	})

}

// TestMessageRaceCondition tests potential race conditions in message pool
func TestMessageRaceCondition(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageRaceCondition", nil, func(t *testing.T, tx *gorm.DB) {
		const numGoroutines = 50
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Launch many goroutines that heavily use the message pool
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					// Get message from pool
					msg := GetMessage()

					// Set some fields
					msg.Type = MessageTypeRequest
					msg.ID = fmt.Sprintf("req-%d-%d", id, j)
					msg.Source = fmt.Sprintf("source-%d", id)

					// Create payload
					payload := map[string]interface{}{
						"id":    j,
						"src":   id,
						"value": fmt.Sprintf("data-%d-%d", id, j),
					}

					// Marshal payload into message
					payloadData, err := json.Marshal(payload)
					if err != nil {
						panic(fmt.Sprintf("Failed to marshal payload: %v", err))
					}
					msg.Payload = payloadData

					// Marshal entire message
					fullData, err := msg.Marshal()
					if err != nil {
						panic(fmt.Sprintf("Failed to marshal message: %v", err))
					}

					// Unmarshal back
					newMsg, err := Unmarshal(fullData)
					if err != nil {
						panic(fmt.Sprintf("Failed to unmarshal message: %v", err))
					}

					// Verify data integrity
					var newPayload map[string]interface{}
					if err := newMsg.UnmarshalPayload(&newPayload); err != nil {
						panic(fmt.Sprintf("Failed to unmarshal payload: %v", err))
					}

					// Put messages back in pool
					PutMessage(msg)
					PutMessage(newMsg)

					// Small delay to increase chance of race conditions
					time.Sleep(time.Nanosecond)
				}
			}(i)
		}

		wg.Wait()
	})

}

// TestMessageWithUnicode tests messages with Unicode content
func TestMessageWithUnicode(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessageWithUnicode", nil, func(t *testing.T, tx *gorm.DB) {
		unicodePayload := map[string]string{
			"greeting":    "Hello 世界 🌍",
			"description": "日本語 Ελληνικά عربى русский язык",
			"special":     "áéíóú ñ ç ü",
		}

		msg, err := NewRequestMessage("unicode-test", unicodePayload)
		if err != nil {
			t.Fatalf("Failed to create message with Unicode payload: %v", err)
		}
		defer PutMessage(msg)

		// Marshal and unmarshal to test Unicode preservation
		data, err := msg.Marshal()
		if err != nil {
			t.Fatalf("Failed to marshal Unicode message: %v", err)
		}

		deserialized, err := Unmarshal(data)
		if err != nil {
			t.Fatalf("Failed to unmarshal Unicode message: %v", err)
		}
		defer PutMessage(deserialized)

		// Verify Unicode content is preserved
		var result map[string]string
		if err := deserialized.UnmarshalPayload(&result); err != nil {
			t.Fatalf("Failed to unmarshal Unicode payload: %v", err)
		}

		for key, expected := range unicodePayload {
			if actual, exists := result[key]; !exists || actual != expected {
				t.Errorf("Unicode payload mismatch for key '%s': expected '%s', got '%s'", key, expected, actual)
			}
		}
	})

}
