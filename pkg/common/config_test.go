package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Load config from non-existent file should return empty config
	config, err := LoadConfig("/tmp/non-existent-file.json")
	if err != nil {
		t.Errorf("LoadConfig should not return error for non-existent file, got: %v", err)
	}
	if config == nil {
		t.Error("LoadConfig should return empty config, got nil")
	}
}

func TestLoadConfig_ValidConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	configData := `{
		"server": {
			"address": ":9090"
		},
		"client": {
			"url": "http://example.com"
		}
	}`

	if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Verify the values
	if config.Server.Address != ":9090" {
		t.Errorf("Expected server address :9090, got %s", config.Server.Address)
	}
	if config.Client.URL != "http://example.com" {
		t.Errorf("Expected client URL http://example.com, got %s", config.Client.URL)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create a temporary config file with invalid JSON
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	invalidJSON := `{"server": "invalid`

	if err := os.WriteFile(configFile, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load the config - should return error
	_, err := LoadConfig(configFile)
	if err == nil {
		t.Error("LoadConfig should return error for invalid JSON")
	}
}

func TestLoadConfig_DefaultFile(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a temporary directory and change to it
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test with no config file - should return empty config
	config, err := LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig should not return error when default file doesn't exist, got: %v", err)
	}
	if config == nil {
		t.Error("LoadConfig should return empty config, got nil")
	}

	// Create default config file
	configData := `{"server": {"address": ":8080"}}`
	if err := os.WriteFile(DefaultConfigFile, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create default config file: %v", err)
	}

	// Load default config
	config, err = LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}
	if config.Server.Address != ":8080" {
		t.Errorf("Expected server address :8080, got %s", config.Server.Address)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	// Create a config
	config := &Config{
		Server: ServerConfig{
			Address: ":8080",
		},
		Client: ClientConfig{
			URL: "http://localhost:8080",
		},
	}

	// Save the config
	if err := SaveConfig(config, configFile); err != nil {
		t.Errorf("SaveConfig failed: %v", err)
	}

	// Load it back
	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Verify the values
	if loadedConfig.Server.Address != config.Server.Address {
		t.Errorf("Expected server address %s, got %s", config.Server.Address, loadedConfig.Server.Address)
	}
	if loadedConfig.Client.URL != config.Client.URL {
		t.Errorf("Expected client URL %s, got %s", config.Client.URL, loadedConfig.Client.URL)
	}
}

func TestSaveConfig_DefaultFile(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a temporary directory and change to it
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create a config
	config := &Config{
		Server: ServerConfig{
			Address: ":8080",
		},
	}

	// Save with empty filename (should use default)
	if err := SaveConfig(config, ""); err != nil {
		t.Errorf("SaveConfig failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(DefaultConfigFile); os.IsNotExist(err) {
		t.Error("Default config file was not created")
	}
}
