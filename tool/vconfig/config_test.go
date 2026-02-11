package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestLoadConfig(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLoadConfig", nil, func(t *testing.T, tx *gorm.DB) {
		// Use t.TempDir() for cleaner cleanup
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test-config.json")

		testConfig := Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_DEBUG": {
					Description: "Enable debug features",
					Type:        "bool",
					Default:     false,
					Values:      []string{"true", "false"},
				},
			},
		}

		jsonData, err := json.MarshalIndent(testConfig, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal test config: %v", err)
		}

		if err := os.WriteFile(configPath, jsonData, 0600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Test loading the config
		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Build.ConfigFile != ".config" {
			t.Errorf("Expected build config file '.config', got '%s'", config.Build.ConfigFile)
		}

		option, exists := config.Features["CONFIG_DEBUG"]
		if !exists {
			t.Error("Expected CONFIG_DEBUG option to exist")
		} else {
			if option.Default != false {
				t.Errorf("Expected CONFIG_DEBUG default to be false, got %v", option.Default)
			}
			if option.Description != "Enable debug features" {
				t.Errorf("Expected CONFIG_DEBUG description to be 'Enable debug features', got '%s'", option.Description)
			}
			if option.Type != "bool" {
				t.Errorf("Expected CONFIG_DEBUG type to be 'bool', got '%s'", option.Type)
			}
		}
	})

}

func TestSaveConfig(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSaveConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: "test.config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_DEBUG": {
					Description: "Enable debug features",
					Type:        "bool",
					Default:     true,
					Values:      []string{"true", "false"},
				},
			},
		}

		// Use t.TempDir() for cleaner cleanup
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test-save-config.json")

		err := SaveConfig(configPath, config)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Load the saved config and verify it
		loadedConfig, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.Build.ConfigFile != "test.config" {
			t.Errorf("Expected build config file 'test.config', got '%s'", loadedConfig.Build.ConfigFile)
		}

		option, exists := loadedConfig.Features["CONFIG_DEBUG"]
		if !exists {
			t.Error("CONFIG_DEBUG should exist in loaded config")
		} else {
			if option.Default != true {
				t.Errorf("Expected CONFIG_DEBUG default to be true, got %v", option.Default)
			}
		}
	})

}

func TestGetDefaultConfig(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestGetDefaultConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := GetDefaultConfig()

		if config.Build.ConfigFile != ".config" {
			t.Errorf("Expected default config file '.config', got '%s'", config.Build.ConfigFile)
		}

		// Check that required features exist
		requiredFeatures := []string{"CONFIG_MIN_LOG_LEVEL"}
		for _, feature := range requiredFeatures {
			if _, exists := config.Features[feature]; !exists {
				t.Errorf("Expected feature %s to exist in default config", feature)
			}
		}

		// Check CONFIG_MIN_LOG_LEVEL specifically
		minLogLevel, exists := config.Features["CONFIG_MIN_LOG_LEVEL"]
		if !exists {
			t.Error("CONFIG_MIN_LOG_LEVEL should exist in default config")
		} else {
			if minLogLevel.Description != "Minimum log level for the application" {
				t.Errorf("Expected CONFIG_MIN_LOG_LEVEL description to be 'Minimum log level for the application', got '%s'", minLogLevel.Description)
			}
			if minLogLevel.Default != "INFO" {
				t.Errorf("Expected CONFIG_MIN_LOG_LEVEL default to be 'INFO', got '%v'", minLogLevel.Default)
			}
		}
	})

}
