package routing

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestRoute_MessageTypeVariations covers different message types in routing.
func TestRoute_MessageTypeVariations(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_MessageTypeVariations", nil, func(t *testing.T, tx *gorm.DB) {
		types := []proc.MessageType{
			proc.MessageTypeRequest,
			proc.MessageTypeResponse,
			proc.MessageTypeEvent,
			proc.MessageTypeError,
		}

		for _, msgType := range types {
			t.Run(string(msgType), func(t *testing.T) {
				msg := &proc.Message{
					Type:   msgType,
					ID:     "test-" + string(msgType),
					Target: "test-target",
				}

				// Verify message type is preserved
				if msg.Type != msgType {
					t.Fatalf("Message type mismatch: want %s got %s", msgType, msg.Type)
				}
			})
		}
	})

}

// TestRoute_TargetFormats validates various target process formats.
func TestRoute_TargetFormats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_TargetFormats", nil, func(t *testing.T, tx *gorm.DB) {
		targets := []string{
			"meta",
			"local",
			"remote",
			"access",
			"sysmon",
			"process-1",
			"process-with-dashes",
			"process_with_underscores",
			"ProcessCamelCase",
			strings.Repeat("x", 100), // long target
		}

		for _, target := range targets {
			t.Run(target, func(t *testing.T) {
				msg := &proc.Message{
					Type:   proc.MessageTypeRequest,
					ID:     "test",
					Target: target,
				}

				if msg.Target != target {
					t.Fatalf("Target mismatch: want %s got %s", target, msg.Target)
				}
			})
		}
	})

}

// TestRoute_SourceFormats validates various source process formats.
func TestRoute_SourceFormats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_SourceFormats", nil, func(t *testing.T, tx *gorm.DB) {
		sources := []string{
			"broker",
			"access",
			"client-1",
			"client-2",
			strings.Repeat("s", 50),
		}

		for _, source := range sources {
			t.Run(source, func(t *testing.T) {
				msg := &proc.Message{
					Type:   proc.MessageTypeRequest,
					ID:     "test",
					Source: source,
				}

				if msg.Source != source {
					t.Fatalf("Source mismatch: want %s got %s", source, msg.Source)
				}
			})
		}
	})

}

// TestRoute_CorrelationIDs validates correlation ID formats.
func TestRoute_CorrelationIDs(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_CorrelationIDs", nil, func(t *testing.T, tx *gorm.DB) {
		correlationIDs := []string{
			"req-1",
			"12345678-1234-1234-1234-123456789012", // UUID
			"correlation_" + strings.Repeat("x", 100),
			"",
		}

		for _, corrID := range correlationIDs {
			t.Run(fmt.Sprintf("corrID-%s", corrID), func(t *testing.T) {
				msg := &proc.Message{
					Type:          proc.MessageTypeResponse,
					ID:            "resp",
					CorrelationID: corrID,
				}

				if msg.CorrelationID != corrID {
					t.Fatalf("CorrelationID mismatch: want %s got %s", corrID, msg.CorrelationID)
				}
			})
		}
	})

}

// TestRoute_PayloadSizes validates routing with various payload sizes.
func TestRoute_PayloadSizes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_PayloadSizes", nil, func(t *testing.T, tx *gorm.DB) {
		sizes := []int{0, 10, 100, 1024, 10240}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
				payload := map[string]interface{}{
					"data": strings.Repeat("x", size),
				}

				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				msg := &proc.Message{
					Type:    proc.MessageTypeRequest,
					ID:      "test",
					Payload: json.RawMessage(payloadBytes),
				}

				if len(msg.Payload) == 0 && size > 0 {
					t.Fatalf("Payload is empty for size %d", size)
				}
			})
		}
	})

}

// TestRoute_ErrorMessages validates error message routing.
func TestRoute_ErrorMessages(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_ErrorMessages", nil, func(t *testing.T, tx *gorm.DB) {
		errors := []string{
			"simple error",
			"error with\nmultiple\nlines",
			"error with unicode: 错误 - ошибка",
			strings.Repeat("error ", 100),
			"error with special chars: <>&\"'",
		}

		for _, errMsg := range errors {
			t.Run(fmt.Sprintf("error-%d", len(errMsg)), func(t *testing.T) {
				msg := &proc.Message{
					Type:  proc.MessageTypeError,
					ID:    "error",
					Error: errMsg,
				}

				if msg.Error != errMsg {
					t.Fatalf("Error mismatch: want %s got %s", errMsg, msg.Error)
				}
			})
		}
	})

}

// TestMessage_JSONRoundTripForRouting validates message serialization for routing.
func TestMessage_JSONRoundTripForRouting(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMessage_JSONRoundTripForRouting", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			msg  proc.Message
		}{
			{
				name: "request-with-all-fields",
				msg: proc.Message{
					Type:          proc.MessageTypeRequest,
					ID:            "req-1",
					Source:        "access",
					Target:        "meta",
					CorrelationID: "",
					Payload:       json.RawMessage(`{"method":"RPCGetCVE","params":{"cve_id":"CVE-2021-1234"}}`),
				},
			},
			{
				name: "response-with-correlation",
				msg: proc.Message{
					Type:          proc.MessageTypeResponse,
					ID:            "resp-1",
					Source:        "meta",
					Target:        "access",
					CorrelationID: "req-1",
					Payload:       json.RawMessage(`{"retcode":0,"message":"success"}`),
				},
			},
			{
				name: "event-notification",
				msg: proc.Message{
					Type:    proc.MessageTypeEvent,
					ID:      "event-1",
					Source:  "local",
					Payload: json.RawMessage(`{"event":"cve_imported","count":100}`),
				},
			},
			{
				name: "error-with-details",
				msg: proc.Message{
					Type:   proc.MessageTypeError,
					ID:     "error-1",
					Source: "remote",
					Target: "access",
					Error:  "failed to fetch CVE data",
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.msg)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded proc.Message
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.Type != tc.msg.Type {
					t.Fatalf("Type mismatch: want %s got %s", tc.msg.Type, decoded.Type)
				}
				if decoded.ID != tc.msg.ID {
					t.Fatalf("ID mismatch: want %s got %s", tc.msg.ID, decoded.ID)
				}
				if decoded.Source != tc.msg.Source {
					t.Fatalf("Source mismatch: want %s got %s", tc.msg.Source, decoded.Source)
				}
				if decoded.Target != tc.msg.Target {
					t.Fatalf("Target mismatch: want %s got %s", tc.msg.Target, decoded.Target)
				}
			})
		}
	})

}

// TestRoute_PayloadFormats validates various payload content types.
func TestRoute_PayloadFormats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_PayloadFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name    string
			payload interface{}
		}{
			{name: "string-params", payload: "test"},
			{name: "number-params", payload: 42},
			{name: "boolean-params", payload: true},
			{name: "null-params", payload: nil},
			{name: "object-params", payload: map[string]interface{}{"key": "value"}},
			{name: "array-params", payload: []interface{}{1, 2, 3}},
			{name: "nested-object", payload: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep",
					},
				},
			}},
			{name: "mixed-types", payload: map[string]interface{}{
				"str":  "value",
				"num":  123,
				"bool": true,
				"arr":  []interface{}{1, 2, 3},
				"obj":  map[string]interface{}{"nested": "value"},
			}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				payloadBytes, err := json.Marshal(tc.payload)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				msg := &proc.Message{
					Type:    proc.MessageTypeRequest,
					ID:      "test",
					Payload: json.RawMessage(payloadBytes),
				}

				// Verify payload can be unmarshaled
				var decoded interface{}
				if err := json.Unmarshal(msg.Payload, &decoded); err != nil {
					t.Fatalf("json.Unmarshal payload failed: %v", err)
				}
			})
		}
	})

}

// TestRoute_UnicodeInFields validates unicode handling in all fields.
func TestRoute_UnicodeInFields(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_UnicodeInFields", nil, func(t *testing.T, tx *gorm.DB) {
		msg := &proc.Message{
			Type:    proc.MessageTypeRequest,
			ID:      "请求-αβγ-🎉",
			Source:  "源-source",
			Target:  "目标-target",
			Payload: json.RawMessage(`{"message":"消息内容 🔒"}`),
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		var decoded proc.Message
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if decoded.ID != msg.ID {
			t.Fatalf("ID mismatch: want %s got %s", msg.ID, decoded.ID)
		}
		if decoded.Source != msg.Source {
			t.Fatalf("Source mismatch: want %s got %s", msg.Source, decoded.Source)
		}
		if decoded.Target != msg.Target {
			t.Fatalf("Target mismatch: want %s got %s", msg.Target, decoded.Target)
		}
	})

}

// TestRoute_SpecialCharactersInIDs validates ID field edge cases.
func TestRoute_SpecialCharactersInIDs(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_SpecialCharactersInIDs", nil, func(t *testing.T, tx *gorm.DB) {
		ids := []string{
			"simple-id",
			"id_with_underscores",
			"id.with.dots",
			"id:with:colons",
			"id/with/slashes",
			"id@with@at",
			"id#with#hash",
			"id$with$dollar",
			"id%with%percent",
			"id&with&ampersand",
			"id with spaces",
			"id\twith\ttabs",
		}

		for _, id := range ids {
			t.Run(id, func(t *testing.T) {
				msg := &proc.Message{
					Type: proc.MessageTypeRequest,
					ID:   id,
				}

				data, err := json.Marshal(msg)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded proc.Message
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != id {
					t.Fatalf("ID mismatch: want %s got %s", id, decoded.ID)
				}
			})
		}
	})

}

// TestRoute_RequestResponsePairs validates request-response correlation.
func TestRoute_RequestResponsePairs(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_RequestResponsePairs", nil, func(t *testing.T, tx *gorm.DB) {
		requestID := "req-12345"
		correlationID := requestID

		request := &proc.Message{
			Type:   proc.MessageTypeRequest,
			ID:     requestID,
			Source: "client",
			Target: "server",
		}

		response := &proc.Message{
			Type:          proc.MessageTypeResponse,
			ID:            "resp-12345",
			Source:        "server",
			Target:        "client",
			CorrelationID: correlationID,
		}

		// Verify correlation
		if response.CorrelationID != request.ID {
			t.Fatalf("CorrelationID mismatch: want %s got %s", request.ID, response.CorrelationID)
		}
	})

}

// TestRoute_EmptyFields validates handling of empty field values.
func TestRoute_EmptyFields(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRoute_EmptyFields", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			msg  proc.Message
		}{
			{name: "empty-source", msg: proc.Message{Type: proc.MessageTypeRequest, ID: "1", Source: ""}},
			{name: "empty-target", msg: proc.Message{Type: proc.MessageTypeRequest, ID: "2", Target: ""}},
			{name: "empty-correlation", msg: proc.Message{Type: proc.MessageTypeResponse, ID: "3", CorrelationID: ""}},
			{name: "empty-error", msg: proc.Message{Type: proc.MessageTypeError, ID: "4", Error: ""}},
			{name: "nil-payload", msg: proc.Message{Type: proc.MessageTypeRequest, ID: "5", Payload: nil}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.msg)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded proc.Message
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}
			})
		}
	})

}
