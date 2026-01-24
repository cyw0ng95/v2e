package jsonutil

import (
	"bytes"
	"testing"
)

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	original := sample{Name: "alpha", Count: 42}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	var decoded sample
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if decoded != original {
		t.Fatalf("Decoded mismatch: %+v", decoded)
	}
}

func TestMarshalIndent(t *testing.T) {
	payload := map[string]int{"k": 1}

	data, err := MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent returned error: %v", err)
	}

	if len(data) == 0 || data[0] != '{' || !bytes.Contains(data, []byte("\n")) {
		t.Fatalf("MarshalIndent did not indent output: %q", string(data))
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	var decoded sample
	if err := Unmarshal([]byte("{invalid"), &decoded); err == nil {
		t.Fatalf("Expected error for invalid JSON")
	}
}
