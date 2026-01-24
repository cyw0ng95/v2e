package subprocess

import (
    "testing"
)

func TestUnmarshalFast_InvalidJSON(t *testing.T) {
    var dst struct{ A int `json:"a"` }
    // invalid JSON should return an error
    if err := UnmarshalFast([]byte("invalid json"), &dst); err == nil {
        t.Fatalf("expected error for invalid json, got nil")
    }
}
