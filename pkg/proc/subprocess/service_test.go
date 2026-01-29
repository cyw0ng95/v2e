package subprocess

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected common.LogLevel
	}{
		{"debug", common.DebugLevel},
		{"info", common.InfoLevel},
		{"warn", common.WarnLevel},
		{"error", common.ErrorLevel},
		{"invalid", common.InfoLevel}, // default
		{"", common.InfoLevel},        // default
	}

	for _, tt := range tests {
		result := parseLogLevel(tt.input)
		if result != tt.expected {
			t.Errorf("parseLogLevel(%q) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}
