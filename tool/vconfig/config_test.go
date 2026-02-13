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

func TestConfigValidateMissingType(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateMissingType", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option",
					Type:        "",
					Default:     true,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for missing type, got nil")
		}
	})
}

func TestConfigValidateMissingDescription(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateMissingDescription", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "",
					Type:        "bool",
					Default:     true,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for missing description, got nil")
		}
	})
}

func TestConfigValidateShortDescription(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateShortDescription", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "Short",
					Type:        "bool",
					Default:     true,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for short description, got nil")
		}
	})
}

func TestConfigValidateInvalidType(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidType", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option with sufficient length",
					Type:        "invalid_type",
					Default:     true,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid type, got nil")
		}
	})
}

func TestConfigValidateTypeMismatchBool(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateTypeMismatchBool", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option with sufficient length",
					Type:        "bool",
					Default:     "not_a_bool",
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for bool type mismatch, got nil")
		}
	})
}

func TestConfigValidateTypeMismatchInt(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateTypeMismatchInt", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option with sufficient length",
					Type:        "int",
					Default:     3.14,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for int type mismatch, got nil")
		}
	})
}

func TestConfigValidateDuplicateValues(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateDuplicateValues", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option with sufficient length",
					Type:        "string",
					Default:     "value1",
					Values:      []string{"value1", "value2", "value1"},
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for duplicate values, got nil")
		}
	})
}

func TestConfigValidateInvalidMethod(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidMethod", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option with sufficient length",
					Type:        "bool",
					Default:     true,
					Method:      "invalid_method",
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid method, got nil")
		}
	})
}

func TestConfigValidateLdflagsMissingTarget(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateLdflagsMissingTarget", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST": {
					Description: "This is a test option with sufficient length",
					Type:        "bool",
					Default:     true,
					Method:      "ldflags",
					Target:      "",
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for ldflags without target, got nil")
		}
	})
}

func TestConfigValidateInvalidServerAddr(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidServerAddr", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_ACCESS_SERVERADDR": {
					Description: "Server address for access service",
					Type:        "string",
					Default:     "invalid:99999",
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid server address, got nil")
		}
	})
}

func TestConfigValidateValidServerAddr(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateValidServerAddr", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_ACCESS_SERVERADDR": {
					Description: "Server address for access service",
					Type:        "string",
					Default:     "0.0.0.0:8080",
				},
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no validation error for valid server address, got: %v", err)
		}
	})
}

func TestConfigValidateInvalidLogLevel(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidLogLevel", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_MIN_LOG_LEVEL": {
					Description: "Minimum log level for the application",
					Type:        "string",
					Default:     "INVALID",
					Values:      []string{"DEBUG", "INFO", "WARN", "ERROR"},
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid log level, got nil")
		}
	})
}

func TestConfigValidateInvalidLogLevelValue(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidLogLevelValue", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_MIN_LOG_LEVEL": {
					Description: "Minimum log level for the application",
					Type:        "string",
					Default:     "INFO",
					Values:      []string{"DEBUG", "INFO", "INVALID", "ERROR"},
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid log level in values array, got nil")
		}
	})
}

func TestConfigValidateNegativeInt(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateNegativeInt", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_OPTIMIZER_BATCH": {
					Description: "Batch size for message batching",
					Type:        "int",
					Default:     -1,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for negative integer, got nil")
		}
	})
}

func TestConfigValidateInvalidOptimizerPolicy(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidOptimizerPolicy", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_OPTIMIZER_POLICY": {
					Description: "Offer policy for message handling when buffer is full",
					Type:        "string",
					Default:     "invalid_policy",
					Values:      []string{"drop", "wait", "reject"},
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid optimizer policy, got nil")
		}
	})
}

func TestConfigValidateInvalidWorkerCount(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateInvalidWorkerCount", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_OPTIMIZER_WORKERS": {
					Description: "Number of worker goroutines for message processing",
					Type:        "int",
					Default:     0.0,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for zero worker count, got nil")
		}
	})
}

func TestConfigValidateEmptyBuildConfigFile(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateEmptyBuildConfigFile", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: "",
			},
			Features: map[string]ConfigOption{},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for empty build config file, got nil")
		}
	})
}

func TestConfigValidateCrossFieldWorkerPool(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateCrossFieldWorkerPool", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_WORKERPOOL_MIN_SIZE": {
					Description: "Minimum size of worker pool",
					Type:        "int",
					Default:     100.0,
				},
				"CONFIG_WORKERPOOL_MAX_SIZE": {
					Description: "Maximum size of worker pool",
					Type:        "int",
					Default:     50.0,
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for min > max worker pool size, got nil")
		}
	})
}

func TestConfigValidateValidConfig(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateValidConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_DEBUG": {
					Description: "Enable debug features for development",
					Type:        "bool",
					Default:     false,
					Method:      "build-tag",
				},
				"CONFIG_PORT": {
					Description: "Server port for the application",
					Type:        "int",
					Default:     8080.0,
					Method:      "ldflags",
					Target:      "github.com/cyw0ng95/v2e/cmd/server.port",
				},
				"CONFIG_MODE": {
					Description: "Operating mode for the application",
					Type:        "string",
					Default:     "production",
					Values:      []string{"development", "production", "testing"},
				},
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no validation error for valid config, got: %v", err)
		}
	})
}

func TestConfigValidatePathWithParentRef(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidatePathWithParentRef", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST_DBPATH": {
					Description: "Database path for test data",
					Type:        "string",
					Default:     "/etc/safe/../unsafe/db.db",
				},
			},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for path with parent reference, got nil")
		}
	})
}

func TestConfigValidateValidPath(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConfigValidateValidPath", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			Build: BuildSection{
				ConfigFile: ".config",
			},
			Features: map[string]ConfigOption{
				"CONFIG_TEST_DBPATH": {
					Description: "Database path for test data",
					Type:        "string",
					Default:     "/var/data/test.db",
				},
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("Expected no validation error for valid path, got: %v", err)
		}
	})
}
