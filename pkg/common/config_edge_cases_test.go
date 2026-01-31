package common

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfig_EdgeCases tests edge cases for config loading
func TestLoadConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*testing.T) string
		expectError   bool
		expectNil     bool
		expectAddress string
	}{
		{
			name: "empty config file",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.json")
				if err := os.WriteFile(configFile, []byte("{}"), 0644); err != nil {
					t.Fatalf("Failed to create empty config file: %v", err)
				}
				return configFile
			},
			expectError:   false,
			expectNil:     false,
			expectAddress: "",
		},
		{
			name: "config with only server section",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.json")
				configData := `{"server": {"address": ":7777"}}`
				if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
					t.Fatalf("Failed to create server-only config file: %v", err)
				}
				return configFile
			},
			expectError:   false,
			expectNil:     false,
			expectAddress: ":7777",
		},
		{
			name: "config with nested sections",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.json")
				configData := `{
					"server": {"address": ":8081"},
					"client": {"url": "https://api.example.com"},
					"broker": {
						"processes": [
							{
								"id": "test-process",
								"command": "/bin/test",
								"args": ["--verbose"],
								"rpc": true
							}
						],
						"authentication": {
							"enabled": true,
							"tokens": {
								"secret-token": {
									"endpoints": ["/api/*"],
									"processes": ["test-process"]
								}
							}
						}
					}
				}`
				if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
					t.Fatalf("Failed to create nested config file: %v", err)
				}
				return configFile
			},
			expectError:   false,
			expectNil:     false,
			expectAddress: ":8081",
		},
		{
			name: "config with unicode characters",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.json")
				configData := `{
					"server": {"address": ":9000"},
					"local": {"cve_db_path": "/path/to/数据库.db", "cwe_db_path": "/path/to/数据.cwe"}
				}`
				if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
					t.Fatalf("Failed to create unicode config file: %v", err)
				}
				return configFile
			},
			expectError:   false,
			expectNil:     false,
			expectAddress: ":9000",
		},
		{
			name: "config with special characters in keys",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.json")
				configData := `{
					"server": {"address": ":9001"},
					"custom-section": {"special.key": "value", "special/key": "another-value"}
				}`
				if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
					t.Fatalf("Failed to create special char config file: %v", err)
				}
				return configFile
			},
			expectError:   false,
			expectNil:     false,
			expectAddress: ":9001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := tt.setupFunc(t)
			config, err := LoadConfig(configFile)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.expectNil && config != nil {
				t.Errorf("Expected nil config but got %v", config)
				return
			}
			if !tt.expectNil && config == nil {
				t.Errorf("Expected non-nil config but got nil")
				return
			}
			if config != nil && config.Server.Address != tt.expectAddress {
				t.Errorf("Expected server address %s, got %s", tt.expectAddress, config.Server.Address)
			}
		})
	}
}

// TestLoadConfig_Permissions tests loading config with different file permissions
func TestLoadConfig_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	configData := `{"server": {"address": ":6666"}}`
	if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Change file permissions to read-only
	if err := os.Chmod(configFile, 0444); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}

	// Should still be readable
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("LoadConfig failed with read-only file: %v", err)
	}
	if config.Server.Address != ":6666" {
		t.Errorf("Expected server address :6666, got %s", config.Server.Address)
	}
}

// TestLoadConfig_LongPaths tests loading config with long file paths
func TestLoadConfig_LongPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deeply nested directory structure
	longPath := tmpDir
	for i := 0; i < 10; i++ {
		longPath = filepath.Join(longPath, "very", "long", "path", "structure")
	}
	if err := os.MkdirAll(longPath, 0755); err != nil {
		t.Fatalf("Failed to create long path: %v", err)
	}

	configFile := filepath.Join(longPath, "config.json")
	configData := `{"server": {"address": ":5555"}}`
	if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("LoadConfig failed with long path: %v", err)
	}
	if config.Server.Address != ":5555" {
		t.Errorf("Expected server address :5555, got %s", config.Server.Address)
	}
}

// TestSaveConfig_EdgeCases tests edge cases for config saving
func TestSaveConfig_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Test saving to a directory that doesn't exist yet
	newDir := filepath.Join(tmpDir, "new", "directory", "path")
	configFile := filepath.Join(newDir, "config.json")

	config := &Config{
		Server: ServerConfig{
			Address: ":4444",
		},
		Client: ClientConfig{
			URL: "https://test.example.com",
		},
	}

	// This should fail because the directory doesn't exist
	err := SaveConfig(config, configFile)
	if err == nil {
		t.Errorf("SaveConfig should fail when directory doesn't exist")
	}

	// Create the directory
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Now it should work
	err = SaveConfig(config, configFile)
	if err != nil {
		t.Errorf("SaveConfig failed after creating directory: %v", err)
	}

	// Load it back to verify
	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("Failed to load saved config: %v", err)
	}
	if loadedConfig.Server.Address != ":4444" {
		t.Errorf("Expected server address :4444, got %s", loadedConfig.Server.Address)
	}
}

// TestConfig_StructureIntegrity tests that config structure maintains integrity through save/load cycle
func TestConfig_StructureIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	originalConfig := &Config{
		Server: ServerConfig{
			Address: ":3000",
		},
		Client: ClientConfig{
			URL: "https://api.service.com",
		},
		Broker: BrokerConfig{
			DetectBins: true,
			Authentication: AuthenticationConfig{
				Enabled: true,
				Tokens: map[string]TokenPermissions{
					"admin-token": {
						Endpoints: []string{"/admin/*", "/api/*"},
						Processes: []string{"service-1"},
					},
				},
			},
			Transport: TransportConfigOptions{
				Type:                 "auto",
				UDSBasePath:          "/tmp/sockets",
				UDSReconnectAttempts: 3,
				UDSReconnectDelayMs:  1000,
			},
		},
		Proc: ProcConfig{
			MaxMessageSizeBytes: 1048576, // 1MB
			RPCInputFD:          3,
			RPCOutputFD:         4,
		},
		Local: LocalConfig{
			CVEDBPath:   "/data/cve.db",
			CWEDBPath:   "/data/cwe.db",
			CAPECDBPath: "/data/capec.db",
		},
		Meta: MetaConfig{
			SessionDBPath: "/data/session.db",
		},
		Remote: RemoteConfig{
			NVDAPIKey:    "test-api-key",
			ViewFetchURL: "https://nvd.nist.gov/feeds/json/cve/1.1",
		},
		Assets: AssetsConfig{
			CWERawPath:   "/assets/cwe/raw",
			CAPECXMLPath: "/assets/capec/xml",
			CAPECXSDPath: "/assets/capec/xsd",
		},
		Capec: CapecConfig{
			StrictXSDValidation: true,
		},

		Access: AccessConfig{
			RPCTimeoutSeconds:      45,
			ShutdownTimeoutSeconds: 15,
			StaticDir:              "public",
		},
		Logging: LoggingConfig{
			Level: "debug",
			Dir:   "/var/log/v2e",
		},
	}

	// Save the config
	err := SaveConfig(originalConfig, configFile)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load it back
	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Compare key values to ensure integrity
	if loadedConfig.Server.Address != originalConfig.Server.Address {
		t.Errorf("Server address mismatch: expected %s, got %s", originalConfig.Server.Address, loadedConfig.Server.Address)
	}
	if loadedConfig.Client.URL != originalConfig.Client.URL {
		t.Errorf("Client URL mismatch: expected %s, got %s", originalConfig.Client.URL, loadedConfig.Client.URL)
	}

	if loadedConfig.Proc.MaxMessageSizeBytes != originalConfig.Proc.MaxMessageSizeBytes {
		t.Errorf("Proc MaxMessageSizeBytes mismatch: expected %d, got %d", originalConfig.Proc.MaxMessageSizeBytes, loadedConfig.Proc.MaxMessageSizeBytes)
	}
	if loadedConfig.Local.CVEDBPath != originalConfig.Local.CVEDBPath {
		t.Errorf("Local CVEDBPath mismatch: expected %s, got %s", originalConfig.Local.CVEDBPath, loadedConfig.Local.CVEDBPath)
	}
	if loadedConfig.Logging.Level != originalConfig.Logging.Level {
		t.Errorf("Logging Level mismatch: expected %s, got %s", originalConfig.Logging.Level, loadedConfig.Logging.Level)
	}
}

// TestConfig_EmptyValues tests handling of empty/zero values in config
func TestConfig_EmptyValues(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	// Config with empty/zero values
	configWithEmpties := &Config{
		Server: ServerConfig{
			Address: "", // empty address
		},
		Client: ClientConfig{
			URL: "", // empty URL
		},
		Proc: ProcConfig{
			MaxMessageSizeBytes: 0, // zero value
			RPCInputFD:          0, // zero value
			RPCOutputFD:         0, // zero value
		},
	}

	err := SaveConfig(configWithEmpties, configFile)
	if err != nil {
		t.Fatalf("SaveConfig failed with empty values: %v", err)
	}

	loadedConfig, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.Server.Address != "" {
		t.Errorf("Expected empty server address, got %s", loadedConfig.Server.Address)
	}
	if loadedConfig.Client.URL != "" {
		t.Errorf("Expected empty client URL, got %s", loadedConfig.Client.URL)
	}
	if loadedConfig.Proc.MaxMessageSizeBytes != 0 {
		t.Errorf("Expected zero MaxMessageSizeBytes, got %d", loadedConfig.Proc.MaxMessageSizeBytes)
	}
	if loadedConfig.Proc.RPCInputFD != 0 {
		t.Errorf("Expected zero RPCInputFD, got %d", loadedConfig.Proc.RPCInputFD)
	}
	if loadedConfig.Proc.RPCOutputFD != 0 {
		t.Errorf("Expected zero RPCOutputFD, got %d", loadedConfig.Proc.RPCOutputFD)
	}
}
