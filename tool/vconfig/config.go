package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the configuration structure
type Config struct {
	Build    BuildSection            `json:"build"`
	Features map[string]ConfigOption `json:"features"`
	Profiles map[string]interface{}  `json:"profiles,omitempty"`
}

// BuildSection represents build-related configuration
type BuildSection struct {
	ConfigFile string `json:"config_file"`
}

// ConfigOption represents a single configuration option
type ConfigOption struct {
	Description string      `json:"description"`
	Type        string      `json:"type"` // "bool", "string", "int", etc.
	Default     interface{} `json:"default"`
	Values      []string    `json:"values,omitempty"` // Available values for selection
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func GetDefaultConfigFromFile() (Config, error) {
	data, err := os.ReadFile("../config_spec.json")
	if err != nil {
		// If the file doesn't exist, return the default config
		return GetDefaultConfig(), nil
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func GetDefaultConfig() Config {
	return Config{
		Build: BuildSection{
			ConfigFile: ".config",
		},
		Features: map[string]ConfigOption{
			"CONFIG_MIN_LOG_LEVEL": {
				Description: "Minimum log level for the application",
				Type:        "string",
				Default:     "INFO",
				Values:      []string{"DEBUG", "INFO", "WARN", "ERROR"},
			},
			"CONFIG_DEBUG_MODE": {
				Description: "Enable debug mode with additional logging",
				Type:        "bool",
				Default:     false,
			},
			"CONFIG_ENABLE_METRICS": {
				Description: "Enable metrics collection and reporting",
				Type:        "bool",
				Default:     true,
			},
			"CONFIG_ENABLE_TRACING": {
				Description: "Enable distributed tracing",
				Type:        "bool",
				Default:     false,
			},
			"CONFIG_ENABLE_PROFILING": {
				Description: "Enable profiling for performance analysis",
				Type:        "bool",
				Default:     false,
			},
			"CONFIG_USE_CACHE": {
				Description: "Enable caching mechanisms",
				Type:        "bool",
				Default:     true,
			},
			"CONFIG_ENABLE_SSL": {
				Description: "Enable SSL/TLS encryption",
				Type:        "bool",
				Default:     true,
			},
			"CONFIG_ASYNC_PROCESSING": {
				Description: "Enable asynchronous processing",
				Type:        "bool",
				Default:     true,
			},
		},
		Profiles: map[string]interface{}{
			"development": map[string]interface{}{
				"CONFIG_MIN_LOG_LEVEL":    "DEBUG",
				"CONFIG_DEBUG_MODE":       true,
				"CONFIG_ENABLE_METRICS":   true,
				"CONFIG_ENABLE_TRACING":   true,
				"CONFIG_ENABLE_PROFILING": true,
				"CONFIG_USE_CACHE":        false,
				"CONFIG_ENABLE_SSL":       true,
				"CONFIG_ASYNC_PROCESSING": true,
			},
			"production": map[string]interface{}{
				"CONFIG_MIN_LOG_LEVEL":    "INFO",
				"CONFIG_DEBUG_MODE":       false,
				"CONFIG_ENABLE_METRICS":   true,
				"CONFIG_ENABLE_TRACING":   false,
				"CONFIG_ENABLE_PROFILING": false,
				"CONFIG_USE_CACHE":        true,
				"CONFIG_ENABLE_SSL":       true,
				"CONFIG_ASYNC_PROCESSING": true,
			},
		},
	}
}

func (c *Config) Validate() error {
	for key, option := range c.Features {
		if option.Description == "" {
			return fmt.Errorf("option %s missing description", key)
		}
		if option.Type == "" {
			return fmt.Errorf("option %s missing type", key)
		}
	}
	return nil
}

// LoadSimpleConfig loads a simple config file with CONFIG_XXX=YYY format
func LoadSimpleConfig(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	result := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result, nil
}

// ConvertSimpleToFullConfig converts a simple config map to a full Config struct
// Uses the default config as a template and applies the simple config values
func ConvertSimpleToFullConfig(simpleConfig map[string]string) (*Config, error) {
	// Start with the default config as a template
	fullConfig := GetDefaultConfig()

	// Apply values from simple config to the full config
	for key, value := range simpleConfig {
		if option, exists := fullConfig.Features[key]; exists {
			// Convert the string value to the appropriate type based on the option type
			switch option.Type {
			case "bool":
				if value == "y" || value == "true" || value == "1" {
					option.Default = true
				} else {
					option.Default = false
				}
			case "string":
				// Remove quotes if present
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
					value = value[1 : len(value)-1]
				}
				option.Default = value
			case "int":
				if intValue, err := strconv.Atoi(value); err == nil {
					option.Default = intValue
				}
			default:
				// For unknown types, treat as string
				option.Default = value
			}
			fullConfig.Features[key] = option
		}
	}

	return &fullConfig, nil
}
