package subprocess

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestUnmarshalPayload_NilPayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshalPayload_NilPayload", nil, func(t *testing.T, tx *gorm.DB) {
		msg := &Message{Payload: nil}
		var v interface{}
		if err := UnmarshalPayload(msg, &v); err == nil {
			t.Fatalf("expected error for nil payload, got nil")
		}
	})

}

func TestUnmarshalPayload_InvalidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshalPayload_InvalidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		msg := &Message{Payload: []byte("not-json")}
		var v interface{}
		if err := UnmarshalPayload(msg, &v); err == nil {
			t.Fatalf("expected json unmarshal error, got nil")
		}
	})

}

func TestUnmarshalPayload_ValidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshalPayload_ValidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		msg := &Message{Payload: []byte(`{"x":123}`)}
		var v map[string]interface{}
		if err := UnmarshalPayload(msg, &v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vv, ok := v["x"].(float64); !ok || vv != 123 {
			t.Fatalf("unexpected value parsed: %#v", v)
		}
	})

}
