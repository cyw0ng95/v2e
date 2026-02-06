package subprocess

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestMalformedJSON tests handling of invalid JSON input
func TestMalformedJSON(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMalformedJSON", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Test various malformed JSON inputs
		testCases := []struct {
			name  string
			input string
		}{
			{"incomplete json", `{"type":"request","id":"1"`},
			{"invalid json", `{invalid json}`},
			{"trailing comma", `{"type":"request","id":"1",}`},
			{"missing quotes", `{type:request,id:1}`},
			{"single quotes", `{'type':'request','id':'1'}`},
			{"empty line", ``},
			{"only whitespace", `   `},
			{"null", `null`},
			{"array instead of object", `["type", "request"]`},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				sp.SetInput(strings.NewReader(tc.input + "\n"))

				// Run should handle malformed input gracefully
				done := make(chan error, 1)
				go func() {
					done <- sp.Run()
				}()

				// Should either return error or continue
				err := <-done
				// Error is acceptable for malformed JSON
				if err == nil {
					// If no error, that's also okay (skipped the bad line)
					t.Logf("Malformed JSON handled gracefully: %s", tc.input)
				}
			})
		}
	})

}

// TestInvalidMessageTypes tests handling of invalid message types
func TestInvalidMessageTypes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestInvalidMessageTypes", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		testCases := []struct {
			name        string
			messageType MessageType
		}{
			{"empty type", MessageType("")},
			{"unknown type", MessageType("unknown")},
			{"numeric type", MessageType("123")},
			{"special chars", MessageType("req@ест!")},
			{"very long type", MessageType(strings.Repeat("a", 1000))},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				msg := &Message{
					Type: tc.messageType,
					ID:   "test-1",
				}

				// Register a handler for this type
				sp.RegisterHandler(string(tc.messageType), func(ctx context.Context, msg *Message) (*Message, error) {
					return nil, nil
				})

				// Handle the message - accessing wg and handleMessage is necessary for testing
				// message handling behavior without running the full subprocess loop
				sp.wg.Add(1)
				sp.handleMessage(msg)
				sp.wg.Wait()

				// Message should be handled
				t.Logf("Message with type '%s' handled", tc.messageType)
			})
		}
	})

}

// TestMissingRequiredFields tests messages with missing required fields
func TestMissingRequiredFields(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMissingRequiredFields", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		testCases := []struct {
			name    string
			message string
		}{
			{"missing id", `{"type":"request"}`},
			{"missing type", `{"id":"test-1"}`},
			{"missing both", `{}`},
			{"only payload", `{"payload":{"key":"value"}}`},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				sp.SetInput(strings.NewReader(tc.message + "\n"))

				// Should handle gracefully
				done := make(chan error, 1)
				go func() {
					done <- sp.Run()
				}()

				<-done
				// Any result is acceptable - we just don't want panic
				t.Logf("Message with missing fields handled: %s", tc.message)
			})
		}
	})

}

// TestOversizedPayload tests handling of very large payloads
func TestOversizedPayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestOversizedPayload", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Create a large payload (1MB)
		largeData := strings.Repeat("x", 1*1024*1024)
		payload := map[string]string{
			"data": largeData,
		}

		payloadBytes, _ := json.Marshal(payload)
		msg := &Message{
			Type:    MessageTypeRequest,
			ID:      "large-1",
			Payload: payloadBytes,
		}

		// Should handle large messages
		if err := sp.SendMessage(msg); err != nil {
			t.Logf("Large message handled with result: %v", err)
		} else {
			t.Log("Large message sent successfully")
		}
	})

}

// TestEmptyAndNullValues tests handling of empty and null values
func TestEmptyAndNullValues(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestEmptyAndNullValues", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		testCases := []struct {
			name    string
			message Message
		}{
			{
				"empty id",
				Message{Type: MessageTypeRequest, ID: ""},
			},
			{
				"null payload",
				Message{Type: MessageTypeResponse, ID: "test-1", Payload: nil},
			},
			{
				"empty error",
				Message{Type: MessageTypeError, ID: "test-1", Error: ""},
			},
			{
				"all fields empty",
				Message{Type: "", ID: "", Payload: nil, Error: ""},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Should handle empty values gracefully
				if err := sp.SendMessage(&tc.message); err != nil {
					t.Logf("Empty value handled with result: %v", err)
				} else {
					t.Log("Empty value message sent successfully")
				}
			})
		}
	})

}

// TestRapidMessageSequence tests handling of rapid message sequences
func TestRapidMessageSequence(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRapidMessageSequence", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Send many messages rapidly
		for i := 0; i < 1000; i++ {
			msg := &Message{
				Type: MessageTypeEvent,
				ID:   "rapid",
			}
			if err := sp.SendMessage(msg); err != nil {
				t.Fatalf("Failed to send message %d: %v", i, err)
			}
		}

		// Flush to ensure all messages are written
		sp.Flush()

		// Count lines in output
		lines := strings.Split(strings.TrimSpace(output.String()), "\n")
		if len(lines) < 1000 {
			t.Errorf("Expected 1000 messages, got %d", len(lines))
		}
	})

}

// TestSpecialCharactersInFields tests handling of special characters
func TestSpecialCharactersInFields(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSpecialCharactersInFields", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		specialStrings := []string{
			"",
			" ",
			"\n",
			"\t",
			"\r\n",
			"null",
			"undefined",
			"<script>alert('xss')</script>",
			"'; DROP TABLE messages; --",
			"../../etc/passwd",
			"\x00\x01\x02",
			"你好世界",
			"🔥💥✨",
		}

		for i, special := range specialStrings {
			t.Run(special, func(t *testing.T) {
				msg := &Message{
					Type: MessageTypeEvent,
					ID:   special,
				}

				// Should handle special characters
				if err := sp.SendMessage(msg); err != nil {
					t.Logf("Special char %d handled with result: %v", i, err)
				}
			})
		}
	})

}

// TestConcurrentUnmarshalPayload tests concurrent payload unmarshaling
func TestConcurrentUnmarshalPayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConcurrentUnmarshalPayload", nil, func(t *testing.T, tx *gorm.DB) {
		payload := map[string]string{"key": "value"}
		data, _ := json.Marshal(payload)

		msg := &Message{
			Type:    MessageTypeRequest,
			ID:      "test",
			Payload: data,
		}

		// Unmarshal concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				var result map[string]string
				if err := UnmarshalPayload(msg, &result); err != nil {
					t.Errorf("Failed to unmarshal payload: %v", err)
				}
				if result["key"] != "value" {
					t.Errorf("Expected 'value', got '%s'", result["key"])
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})

}

// TestInvalidPayloadUnmarshal tests unmarshaling invalid payloads
func TestInvalidPayloadUnmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestInvalidPayloadUnmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		testCases := []struct {
			name    string
			payload json.RawMessage
		}{
			{"invalid json", json.RawMessage(`{invalid}`)},
			{"wrong type", json.RawMessage(`"string instead of object"`)},
			{"number", json.RawMessage(`123`)},
			{"null", json.RawMessage(`null`)},
			{"empty", json.RawMessage(``)},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				msg := &Message{
					Type:    MessageTypeRequest,
					ID:      "test",
					Payload: tc.payload,
				}

				var result map[string]string
				err := UnmarshalPayload(msg, &result)
				if err == nil && tc.name != "null" {
					t.Logf("Unexpected success for %s", tc.name)
				} else {
					t.Logf("Invalid payload handled correctly: %s", tc.name)
				}
			})
		}
	})

}
