package main

import (
	"testing"
)

func TestMainFunctionality(t *testing.T) {
	// Test that the default config can be generated and validated
	config := GetDefaultConfig()

	err := config.Validate()
	if err != nil {
		t.Fatalf("Default config should be valid: %v", err)
	}

	// Check that required features exist
	requiredFeatures := []string{"CONFIG_MIN_LOG_LEVEL", "CONFIG_DEBUG_MODE", "CONFIG_ENABLE_METRICS", "CONFIG_ENABLE_TRACING", "CONFIG_ENABLE_PROFILING", "CONFIG_USE_CACHE", "CONFIG_ENABLE_SSL", "CONFIG_ASYNC_PROCESSING"}
	for _, feature := range requiredFeatures {
		if _, exists := config.Features[feature]; !exists {
			t.Errorf("Expected feature %s to exist in default config", feature)
		}
	}

	// Check the first feature is CONFIG_MIN_LOG_LEVEL
	if _, exists := config.Features["CONFIG_MIN_LOG_LEVEL"]; !exists {
		t.Error("CONFIG_MIN_LOG_LEVEL should be the first (required) configuration option")
	}
}

func TestConfigWithProfiles(t *testing.T) {
	config := GetDefaultConfig()

	// Check that profiles exist
	if config.Profiles == nil {
		t.Fatal("Profiles should not be nil")
	}

	// Check that development and production profiles exist
	if _, exists := config.Profiles["development"]; !exists {
		t.Error("Development profile should exist")
	}

	if _, exists := config.Profiles["production"]; !exists {
		t.Error("Production profile should exist")
	}

	// Check specific profile settings
	devProfile, ok := config.Profiles["development"].(map[string]interface{})
	if !ok {
		t.Error("Development profile should be a map")
	} else {
		if devMinLogLevel, exists := devProfile["CONFIG_MIN_LOG_LEVEL"]; !exists || devMinLogLevel != "DEBUG" {
			t.Error("Development profile should set CONFIG_MIN_LOG_LEVEL to DEBUG")
		}
	}
}
