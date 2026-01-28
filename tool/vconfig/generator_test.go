package main

import (
	"strings"
	"testing"
)

func TestGenerateBuildFlags(t *testing.T) {
	config := GetDefaultConfig()

	flags, err := GenerateBuildFlags(&config)
	if err != nil {
		t.Fatalf("Failed to generate build flags: %v", err)
	}

	// Should not include CONFIG_DEBUG_MODE since it's false by default
	unexpectedFlag := "CONFIG_DEBUG_MODE"
	if strings.Contains(flags, unexpectedFlag) {
		t.Errorf("Build flags should not contain '%s', got '%s'", unexpectedFlag, flags)
	}
}

func TestGenerateBuildFlagsEmpty(t *testing.T) {
	// Create a config with no enabled boolean features
	config := GetDefaultConfig()

	// Disable all boolean features
	for key, option := range config.Features {
		if _, ok := option.Default.(bool); ok {
			option.Default = false
			config.Features[key] = option
		}
	}

	flags, err := GenerateBuildFlags(&config)
	if err != nil {
		t.Fatalf("Failed to generate build flags: %v", err)
	}

	if flags != "" {
		t.Errorf("Expected empty build flags for config with no enabled features, got '%s'", flags)
	}
}
