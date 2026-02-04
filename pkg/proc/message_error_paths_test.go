package proc

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"
)

func TestUnmarshalFast_EmptyOrTruncated(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUnmarshalFast_EmptyOrTruncated", nil, func(t *testing.T, tx *gorm.DB) {
		if _, err := UnmarshalFast([]byte{}); err == nil {
			t.Fatalf("expected error for empty input")
		}
		if _, err := UnmarshalFast([]byte("{\"type\":")); err == nil {
			t.Fatalf("expected error for truncated input")
		}
	})

}

func TestNewResponseMessage_MarshalError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewResponseMessage_MarshalError", nil, func(t *testing.T, tx *gorm.DB) {
		// channels are not JSON-marshalable; expect an error from NewResponseMessage
		_, err := NewResponseMessage("id", make(chan int))
		if err == nil {
			t.Fatalf("expected error when marshaling non-marshable payload")
		}
	})

}
