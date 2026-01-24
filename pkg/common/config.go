package common

import (
	"fmt"
	"os"

	"github.com/cyw0ng95/v2e/pkg/jsonutil"
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
	// Proc configuration for subprocess framework
	Proc ProcConfig `json:"proc,omitempty"`
	// Local service configuration
	Local LocalConfig `json:"local,omitempty"`
	// Meta service configuration
	Meta MetaConfig `json:"meta,omitempty"`
	// Remote service configuration
	Remote RemoteConfig `json:"remote,omitempty"`
	// Asset paths
	Assets AssetsConfig `json:"assets,omitempty"`
	// CAPEC related toggles
	Capec CapecConfig `json:"capec,omitempty"`
	// Website/frontend settings
	Website WebsiteConfig `json:"website,omitempty"`
	// Logging configuration
	Logging LoggingConfig `json:"logging,omitempty"`
	// Access service configuration
	Access AccessConfig `json:"access,omitempty"`
}

// ProcConfig holds process-level configuration (subprocess framework)
type ProcConfig struct {
	// MaxMessageSizeBytes is the maximum size for RPC messages
	MaxMessageSizeBytes int `json:"max_message_size_bytes,omitempty"`
	// RPCInputFD is the file descriptor number used for RPC input
	RPCInputFD int `json:"rpc_input_fd,omitempty"`
	// RPCOutputFD is the file descriptor number used for RPC output
	RPCOutputFD int `json:"rpc_output_fd,omitempty"`
}

// LocalConfig holds local service settings such as DB paths
type LocalConfig struct {
	CVEDBPath   string `json:"cve_db_path,omitempty"`
	CWEDBPath   string `json:"cwe_db_path,omitempty"`
	CAPECDBPath string `json:"capec_db_path,omitempty"`
}

// MetaConfig holds meta service settings
type MetaConfig struct {
	SessionDBPath string `json:"session_db_path,omitempty"`
}

// RemoteConfig holds remote service settings
type RemoteConfig struct {
	NVDAPIKey    string `json:"nvd_api_key,omitempty"`
	ViewFetchURL string `json:"view_fetch_url,omitempty"`
}

// AssetsConfig holds default asset paths used by importers
type AssetsConfig struct {
	CWERawPath   string `json:"cwe_raw_path,omitempty"`
	CAPECXMLPath string `json:"capec_xml_path,omitempty"`
	CAPECXSDPath string `json:"capec_xsd_path,omitempty"`
}

// CapecConfig holds CAPEC-specific toggles
type CapecConfig struct {
	StrictXSDValidation bool `json:"strict_xsd_validation,omitempty"`
}

// WebsiteConfig holds frontend-related configuration
type WebsiteConfig struct {
	APIBaseURL  string `json:"api_base_url,omitempty"`
	UseMockData bool   `json:"use_mock_data,omitempty"`
}

// Extend Config struct with new top-level sections
func init() {
	// noop: types declared for JSON unmarshalling
}

// AccessConfig holds configuration for the access (HTTP) service
type AccessConfig struct {
	// RPC timeout in seconds for forwarding RPC requests (default: 30)
	RPCTimeoutSeconds int `json:"rpc_timeout_seconds,omitempty"`
	// Shutdown timeout in seconds for graceful shutdown (default: 10)
	ShutdownTimeoutSeconds int `json:"shutdown_timeout_seconds,omitempty"`
	// StaticDir is the directory to serve static assets from (default: "website")
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
	// Optional RPC file descriptor overrides for broker-managed processes
	RPCInputFD  int `json:"rpc_input_fd,omitempty"`
	RPCOutputFD int `json:"rpc_output_fd,omitempty"`
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
	if err := jsonutil.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return &config, nil
}

// SaveConfig saves configuration to the specified file
func SaveConfig(config *Config, filename string) error {
	if filename == "" {
		filename = DefaultConfigFile
	}

	data, err := jsonutil.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filename, err)
	}

	return nil
}
