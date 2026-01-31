package subprocess

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func TestDefaultBuildLogLevel(t *testing.T) {
	// This test verifies that the default log level is INFO
	// In a real build with ldflags, this could be different
	level := DefaultBuildLogLevel()

	// By default, without ldflags injection, it should be INFO
	if level != common.InfoLevel {
		t.Errorf("Expected default log level to be INFO, got %v", level)
	}
}
