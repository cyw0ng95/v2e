package common

import (
	"fmt"
	"os"

	"github.com/bytedance/sonic"
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
	// Broker configuration
	Broker BrokerConfig `json:"broker,omitempty"`
	// Logging configuration
	Logging LoggingConfig `json:"logging,omitempty"`
	// Access service configuration
	Access AccessConfig `json:"access,omitempty"`
}

// AccessConfig holds configuration for the access (HTTP) service
type AccessConfig struct {
	// RPC timeout in seconds for forwarding RPC requests (default: 30)
	RPCTimeoutSeconds int `json:"rpc_timeout_seconds,omitempty"`
	// Shutdown timeout in seconds for graceful shutdown (default: 10)
	ShutdownTimeoutSeconds int `json:"shutdown_timeout_seconds,omitempty"`
	// StaticDir is the directory to serve static assets from (default: "website/out")
	StaticDir string `json:"static_dir,omitempty"`
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

// BrokerConfig holds broker-specific configuration
type BrokerConfig struct {
	// Processes is a list of processes to manage
	Processes []ProcessConfig `json:"processes,omitempty"`
	// LogFile is the path to the log file
	LogFile string `json:"log_file,omitempty"`
	// LogsDir is the directory where logs are stored
	LogsDir string `json:"logs_dir,omitempty"`
	// Authentication settings for RPC endpoints
	Authentication AuthenticationConfig `json:"authentication,omitempty"`
}

// ProcessConfig represents a process to be managed by the broker
type ProcessConfig struct {
	// ID is a unique identifier for the process
	ID string `json:"id"`
	// Command is the executable to run
	Command string `json:"command"`
	// Args are the command-line arguments
	Args []string `json:"args,omitempty"`
	// RPC indicates if this is an RPC-enabled process
	RPC bool `json:"rpc,omitempty"`
	// Restart indicates if the process should be restarted on exit
	Restart bool `json:"restart,omitempty"`
	// MaxRestarts is the maximum number of restart attempts (-1 for unlimited)
	MaxRestarts int `json:"max_restarts,omitempty"`
}

// AuthenticationConfig holds authentication settings for RPC endpoints
type AuthenticationConfig struct {
	// Enabled indicates if authentication is enabled
	Enabled bool `json:"enabled,omitempty"`
	// Tokens is a map of allowed tokens and their permissions
	Tokens map[string]TokenPermissions `json:"tokens,omitempty"`
}

// TokenPermissions represents permissions for a token
type TokenPermissions struct {
	// Endpoints is a list of allowed RPC endpoint patterns
	Endpoints []string `json:"endpoints,omitempty"`
	// Processes is a list of allowed process IDs
	Processes []string `json:"processes,omitempty"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error)
	Level string `json:"level,omitempty"`
	// Dir is the directory where logs are stored
	Dir string `json:"dir,omitempty"`
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
	if err := sonic.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return &config, nil
}

// SaveConfig saves configuration to the specified file
func SaveConfig(config *Config, filename string) error {
	if filename == "" {
		filename = DefaultConfigFile
	}

	data, err := sonic.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filename, err)
	}

	return nil
}
