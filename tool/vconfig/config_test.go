package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

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

	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test loading the config
	config, err := LoadConfig(tempFile.Name())
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
}

func TestSaveConfig(t *testing.T) {
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

	tempFile, err := os.CreateTemp("", "test-save-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	os.Remove(tempFile.Name()) // Remove the file so SaveConfig can create it

	err = SaveConfig(tempFile.Name(), config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Load the saved config and verify it
	loadedConfig, err := LoadConfig(tempFile.Name())
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
}

func TestGetDefaultConfig(t *testing.T) {
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
}
