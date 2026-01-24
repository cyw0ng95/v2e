package proc

import (
	"testing"
)

func TestUnmarshalFast_EmptyOrTruncated(t *testing.T) {
	if _, err := UnmarshalFast([]byte{}); err == nil {
		t.Fatalf("expected error for empty input")
	}
	if _, err := UnmarshalFast([]byte("{\"type\":")); err == nil {
		t.Fatalf("expected error for truncated input")
	}
}

func TestNewResponseMessage_MarshalError(t *testing.T) {
	// channels are not JSON-marshalable; expect an error from NewResponseMessage
	_, err := NewResponseMessage("id", make(chan int))
	if err == nil {
		t.Fatalf("expected error when marshaling non-marshable payload")
	}
}
