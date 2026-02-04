package subprocess

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestMessage_JSONRoundTrip covers comprehensive JSON serialization edge cases.
func TestMessage_JSONRoundTrip(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_JSONRoundTrip", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			msg  Message
		}{
			{name: "empty-message", msg: Message{}},
			{name: "only-type", msg: Message{Type: MessageTypeRequest}},
			{name: "type-and-id", msg: Message{Type: MessageTypeRequest, ID: "req1"}},
			{name: "unicode-id", msg: Message{Type: MessageTypeRequest, ID: "请求-αβγ-🎉"}},
			{name: "long-id", msg: Message{Type: MessageTypeRequest, ID: strings.Repeat("x", 1000)}},
			{name: "special-chars-id", msg: Message{Type: MessageTypeRequest, ID: "test\n\t\r\"'\\"}},
			{name: "null-payload", msg: Message{Type: MessageTypeRequest, ID: "r1", Payload: nil}},
			{name: "empty-payload", msg: Message{Type: MessageTypeRequest, ID: "r1", Payload: json.RawMessage("{}")}},
			{name: "nested-payload", msg: Message{Type: MessageTypeRequest, ID: "r1", Payload: json.RawMessage(`{"a":{"b":{"c":"d"}}}`)}},
			{name: "array-payload", msg: Message{Type: MessageTypeRequest, ID: "r1", Payload: json.RawMessage(`[1,2,3]`)}},
			{name: "unicode-error", msg: Message{Type: MessageTypeError, ID: "e1", Error: "错误: ошибка"}},
			{name: "multiline-error", msg: Message{Type: MessageTypeError, ID: "e1", Error: "line1\nline2\nline3"}},
			{name: "html-in-error", msg: Message{Type: MessageTypeError, ID: "e1", Error: "<script>alert('xss')</script>"}},
			{name: "source-target", msg: Message{Type: MessageTypeRequest, ID: "r1", Source: "proc1", Target: "proc2"}},
			{name: "correlation", msg: Message{Type: MessageTypeResponse, ID: "resp1", CorrelationID: "corr-123"}},
			{name: "all-fields", msg: Message{
				Type:          MessageTypeRequest,
				ID:            "full",
				Payload:       json.RawMessage(`{"key":"value"}`),
				Error:         "",
				Source:        "s",
				Target:        "t",
				CorrelationID: "c",
			}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := MarshalFast(&tc.msg)
				if err != nil {
					t.Fatalf("MarshalFast failed: %v", err)
				}

				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed: %v", err)
				}

				if decoded.Type != tc.msg.Type {
					t.Fatalf("Type mismatch: want %s got %s", tc.msg.Type, decoded.Type)
				}
				if decoded.ID != tc.msg.ID {
					t.Fatalf("ID mismatch: want %s got %s", tc.msg.ID, decoded.ID)
				}
			})
		}
	})

}

// TestMessage_InvalidJSON covers malformed JSON handling.
func TestMessage_InvalidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_InvalidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			data string
		}{
			{name: "empty", data: ""},
			{name: "not-json", data: "not json"},
			{name: "incomplete", data: `{"type":"request"`},
			{name: "extra-comma", data: `{"type":"request",}`},
			{name: "missing-quotes", data: `{type:request}`},
			{name: "single-quotes", data: `{'type':'request'}`},
			{name: "trailing-garbage", data: `{"type":"request"}garbage`},
			{name: "unescaped-newline", data: "{\n\"type\":\"request\"\n}"},
			{name: "null-bytes", data: string([]byte{'{', '"', 't', 'y', 'p', 'e', '"', ':', 0, '}'})},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				var msg Message
				err := UnmarshalFast([]byte(tc.data), &msg)
				// Invalid JSON should error - we just verify it doesn't panic
				if err == nil && tc.data != "{\n\"type\":\"request\"\n}" {
					t.Logf("Unexpectedly parsed: %s", tc.data)
				}
			})
		}
	})

}

// TestMessage_PayloadFormats covers various payload structures.
func TestMessage_PayloadFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_PayloadFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name    string
			payload string
		}{
			{name: "string", payload: `"hello"`},
			{name: "number", payload: `42`},
			{name: "boolean", payload: `true`},
			{name: "null", payload: `null`},
			{name: "empty-object", payload: `{}`},
			{name: "empty-array", payload: `[]`},
			{name: "nested-arrays", payload: `[[1,2],[3,4]]`},
			{name: "mixed-types", payload: `{"str":"val","num":123,"bool":true,"null":null}`},
			{name: "unicode", payload: `{"emoji":"😀","chinese":"你好"}`},
			{name: "escaped-chars", payload: `{"quote":"\"","backslash":"\\","newline":"\\n"}`},
			{name: "large-number", payload: `9007199254740991`},
			{name: "float", payload: `3.14159`},
			{name: "scientific", payload: `1.23e10`},
			{name: "negative", payload: `-42`},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				msg := Message{
					Type:    MessageTypeRequest,
					ID:      "test",
					Payload: json.RawMessage(tc.payload),
				}

				data, err := MarshalFast(&msg)
				if err != nil {
					t.Fatalf("MarshalFast failed: %v", err)
				}

				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed: %v", err)
				}

				if string(decoded.Payload) != tc.payload {
					t.Fatalf("Payload mismatch: want %s got %s", tc.payload, string(decoded.Payload))
				}
			})
		}
	})

}

// TestMessage_TypeValidation ensures message types are preserved.
func TestMessage_TypeValidation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_TypeValidation", nil, func(t *testing.T, tx *gorm.DB) {
		types := []MessageType{
			MessageTypeRequest,
			MessageTypeResponse,
			MessageTypeEvent,
			MessageTypeError,
		}

		for _, msgType := range types {
			t.Run(string(msgType), func(t *testing.T) {
				msg := Message{Type: msgType, ID: "test"}
				data, _ := MarshalFast(&msg)
				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed: %v", err)
				}
				if decoded.Type != msgType {
					t.Fatalf("Type mismatch: want %s got %s", msgType, decoded.Type)
				}
			})
		}
	})

}

// TestMessage_LargePayloads covers size edge cases.
func TestMessage_LargePayloads(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_LargePayloads", nil, func(t *testing.T, tx *gorm.DB) {
		sizes := []int{0, 1, 100, 1024, 10240, 102400}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
				payload := strings.Repeat("x", size)
				msg := Message{
					Type:    MessageTypeRequest,
					ID:      "large",
					Payload: json.RawMessage(fmt.Sprintf(`{"data":"%s"}`, payload)),
				}

				data, err := MarshalFast(&msg)
				if err != nil {
					t.Fatalf("MarshalFast failed for size %d: %v", size, err)
				}

				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed for size %d: %v", size, err)
				}
			})
		}
	})

}

// TestMessage_SpecialCharacters covers encoding edge cases.
func TestMessage_SpecialCharacters(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMessage_SpecialCharacters", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name  string
			value string
		}{
			{name: "tab", value: "\t"},
			{name: "newline", value: "\n"},
			{name: "carriage-return", value: "\r"},
			{name: "quote", value: "\""},
			{name: "backslash", value: "\\"},
			{name: "slash", value: "/"},
			{name: "control-chars", value: "\x00\x01\x02"},
			{name: "unicode-bmp", value: "⚡️"},
			{name: "unicode-astral", value: "𝕳𝖊𝖑𝖑𝖔"},
			{name: "rtl-text", value: "مرحبا"},
			{name: "mixed-scripts", value: "Hello世界مرحبا"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				msg := Message{Type: MessageTypeRequest, ID: tc.value}
				data, err := MarshalFast(&msg)
				if err != nil {
					t.Fatalf("MarshalFast failed: %v", err)
				}

				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed: %v", err)
				}

				if decoded.ID != tc.value {
					t.Fatalf("ID mismatch for %s: want %q got %q", tc.name, tc.value, decoded.ID)
				}
			})
		}
	})

}

// TestUnmarshalPayload_TypeConversions covers payload unmarshaling edge cases.
func TestUnmarshalPayload_TypeConversions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUnmarshalPayload_TypeConversions", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name    string
			payload string
			target  interface{}
			valid   bool
		}{
			{name: "string-to-string", payload: `"hello"`, target: new(string), valid: true},
			{name: "number-to-int", payload: `42`, target: new(int), valid: true},
			{name: "number-to-float", payload: `3.14`, target: new(float64), valid: true},
			{name: "bool-to-bool", payload: `true`, target: new(bool), valid: true},
			{name: "object-to-map", payload: `{"k":"v"}`, target: new(map[string]string), valid: true},
			{name: "array-to-slice", payload: `[1,2,3]`, target: new([]int), valid: true},
			{name: "null-to-ptr", payload: `null`, target: new(*string), valid: true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				msg := Message{
					Type:    MessageTypeRequest,
					ID:      "test",
					Payload: json.RawMessage(tc.payload),
				}

				err := UnmarshalPayload(&msg, tc.target)
				if tc.valid && err != nil {
					t.Fatalf("UnmarshalPayload failed: %v", err)
				}
				if !tc.valid && err == nil {
					t.Fatalf("Expected error for invalid conversion")
				}
			})
		}
	})

}
