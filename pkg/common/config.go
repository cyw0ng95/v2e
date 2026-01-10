package common

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "config.json"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server,omitempty"`
	// Client configuration
	Client ClientConfig `json:"client,omitempty"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	// Address to listen on (e.g., ":8080")
	Address string `json:"address,omitempty"`
}

// ClientConfig holds client-specific configuration
type ClientConfig struct {
	// URL to connect to (e.g., "http://localhost:8080")
	URL string `json:"url,omitempty"`
}

// LoadConfig loads configuration from the specified file
// If filename is empty, it uses the default config file
// If the file doesn't exist, it returns an empty config (not an error)
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = DefaultConfigFile
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, return empty config
		return &Config{}, nil
	}

	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return &config, nil
}

// SaveConfig saves configuration to the specified file
func SaveConfig(config *Config, filename string) error {
	if filename == "" {
		filename = DefaultConfigFile
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filename, err)
	}

	return nil
}
