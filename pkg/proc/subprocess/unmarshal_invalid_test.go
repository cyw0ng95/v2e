package subprocess

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestUnmarshalFast_InvalidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestUnmarshalFast_InvalidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		var dst struct {
			A int `json:"a"`
		}
		// invalid JSON should return an error
		if err := UnmarshalFast([]byte("invalid json"), &dst); err == nil {
			t.Fatalf("expected error for invalid json, got nil")
		}
	})

}
