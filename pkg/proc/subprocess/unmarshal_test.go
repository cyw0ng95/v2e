package subprocess

import (
	"testing"
)

func TestUnmarshalPayload_NilPayload(t *testing.T) {
	msg := &Message{Payload: nil}
	var v interface{}
	if err := UnmarshalPayload(msg, &v); err == nil {
		t.Fatalf("expected error for nil payload, got nil")
	}
}

func TestUnmarshalPayload_InvalidJSON(t *testing.T) {
	msg := &Message{Payload: []byte("not-json")}
	var v interface{}
	if err := UnmarshalPayload(msg, &v); err == nil {
		t.Fatalf("expected json unmarshal error, got nil")
	}
}

func TestUnmarshalPayload_ValidJSON(t *testing.T) {
	msg := &Message{Payload: []byte(`{"x":123}`)}
	var v map[string]interface{}
	if err := UnmarshalPayload(msg, &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vv, ok := v["x"].(float64); !ok || vv != 123 {
		t.Fatalf("unexpected value parsed: %#v", v)
	}
}
