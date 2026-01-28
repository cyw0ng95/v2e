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
	// Try multiple possible locations for the config spec file
	possiblePaths := []string{
		"config_spec.json",          // Current directory
		"../config_spec.json",       // One level up
		"../../config_spec.json",    // Two levels up (project root from tool/vconfig)
		"../../../config_spec.json", // Three levels up
	}

	var data []byte
	var err error

	// Try each path until we find the file
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// If the file doesn't exist anywhere, return an empty config
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func GetDefaultConfig() Config {
	// Try to load from spec file first
	specConfig, err := GetDefaultConfigFromFile()
	if err != nil {
		panic("Failed to load default config: " + err.Error())
	}

	return specConfig
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
