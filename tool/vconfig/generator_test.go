package main

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"strings"
	"testing"
)

func TestGenerateBuildFlags(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGenerateBuildFlags", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestGenerateBuildFlagsDefaultConfig(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGenerateBuildFlagsDefaultConfig", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a config with no enabled boolean features but with default string features
		config := GetDefaultConfig()

		// Disable all boolean features
		for key, option := range config.Features {
			if _, ok := option.Default.(bool); ok {
				option.Default = false
				config.Features[key] = option
			}
			// For string features like CONFIG_MIN_LOG_LEVEL, they are now handled via ldflags
			// so they shouldn't appear in build flags
		}

		flags, err := GenerateBuildFlags(&config)
		if err != nil {
			t.Fatalf("Failed to generate build flags: %v", err)
		}

		// With the updated implementation, string features like CONFIG_MIN_LOG_LEVEL
		// should not appear in build flags anymore (they're handled via ldflags)
		if flags != "" {
			t.Errorf("Expected empty build flags for config with only string features, got '%s'", flags)
		}
	})

}

func TestGenerateLdflags(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGenerateLdflags", nil, func(t *testing.T, tx *gorm.DB) {
		config := GetDefaultConfig()

		// Make sure CONFIG_MIN_LOG_LEVEL has the proper method and target
		for key, option := range config.Features {
			if key == "CONFIG_MIN_LOG_LEVEL" {
				option.Default = "DEBUG"
				option.Method = "ldflags"
				option.Target = "subprocess.buildLogLevel"
				config.Features[key] = option
				break
			}
		}

		ldflags, err := GenerateLdflags(&config)
		if err != nil {
			t.Fatalf("Failed to generate ldflags: %v", err)
		}

		// Expect the ldflags to contain the log level injection
		expectedLdflag := "-X 'subprocess.buildLogLevel=DEBUG'"
		if !strings.Contains(ldflags, expectedLdflag) {
			t.Errorf("Expected ldflags to contain '%s', got '%s'", expectedLdflag, ldflags)
		}
	})

}
