package main

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"
)

func TestMainFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMainFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		// Test that the default config can be generated and validated
		config := GetDefaultConfig()

		err := config.Validate()
		if err != nil {
			t.Fatalf("Default config should be valid: %v", err)
		}

		// Check that required features exist
		requiredFeatures := []string{"CONFIG_MIN_LOG_LEVEL"}
		for _, feature := range requiredFeatures {
			if _, exists := config.Features[feature]; !exists {
				t.Errorf("Expected feature %s to exist in default config", feature)
			}
		}

		// Check the first feature is CONFIG_MIN_LOG_LEVEL
		if _, exists := config.Features["CONFIG_MIN_LOG_LEVEL"]; !exists {
			t.Error("CONFIG_MIN_LOG_LEVEL should be the first (required) configuration option")
		}
	})

}
