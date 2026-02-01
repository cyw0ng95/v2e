package common

const (
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "config.json"
)

// Config represents the application configuration
type Config struct {
	// Broker configuration
	Broker BrokerConfig `json:"broker,omitempty"`
	// Proc configuration for subprocess framework
	Proc ProcConfig `json:"proc,omitempty"`
	// Access service configuration (minimal needed)
	Access AccessConfig `json:"access,omitempty"`
	// Logging configuration
	Logging LoggingConfig `json:"logging,omitempty"`
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
	// Local config is now build-time only, no runtime config fields
}

// MetaConfig holds meta service settings
type MetaConfig struct {
	// Session DB path is now build-time only, no runtime config fields
}

// RemoteConfig holds remote service settings
type RemoteConfig struct {
	// Remote config is now build-time only, no runtime config fields
}

// AssetsConfig holds default asset paths used by importers
type AssetsConfig struct {
	// Assets config is now build-time only, no runtime config fields
}

// CapecConfig holds CAPEC-specific toggles
type CapecConfig struct {
	// CAPEC config is now build-time only, no runtime config fields
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
	// Authentication settings for RPC endpoints
	Authentication AuthenticationConfig `json:"authentication,omitempty"`
	// Optional RPC file descriptor overrides for broker-managed processes
	RPCInputFD  int `json:"rpc_input_fd,omitempty"`
	RPCOutputFD int `json:"rpc_output_fd,omitempty"`
	// Transport configuration
	Transport TransportConfigOptions `json:"transport,omitempty"`
	// Optimizer runtime tuning (optional)
	OptimizerBufferCap       int    `json:"optimizer_buffer_cap,omitempty"`
	OptimizerNumWorkers      int    `json:"optimizer_num_workers,omitempty"`
	OptimizerStatsIntervalMs int    `json:"optimizer_stats_interval_ms,omitempty"`
	OptimizerOfferPolicy     string `json:"optimizer_offer_policy,omitempty"`
	OptimizerOfferTimeoutMs  int    `json:"optimizer_offer_timeout_ms,omitempty"`
	// Batching controls
	OptimizerBatchSize       int `json:"optimizer_batch_size,omitempty"`
	OptimizerFlushIntervalMs int `json:"optimizer_flush_interval_ms,omitempty"`
	// Adaptive optimization
	OptimizerEnableAdaptive   bool `json:"optimizer_enable_adaptive,omitempty"`
	OptimizerAdaptationFreqMs int  `json:"optimizer_adaptation_freq_ms,omitempty"`
	// Enable automatic binary detection (default: true)
	DetectBins bool `json:"detect_bins,omitempty"`
	// Comma-separated list of binaries to boot when detect_bins is false (default: "access,remote,local,meta,sysmon")
	BootBins string `json:"boot_bins,omitempty"`
}

// TransportConfigOptions holds configuration for transport mechanisms
type TransportConfigOptions struct {
	// Type specifies the default transport type ("fd", "uds", or "auto")
	Type string `json:"type,omitempty"`
	// UDSBasePath specifies the base path for UDS socket files
	UDSBasePath string `json:"uds_base_path,omitempty"`
	// UDSReconnectAttempts specifies the number of reconnection attempts for UDS
	UDSReconnectAttempts int `json:"uds_reconnect_attempts,omitempty"`
	// UDSReconnectDelayMs specifies the delay between reconnection attempts in milliseconds
	UDSReconnectDelayMs int `json:"uds_reconnect_delay_ms,omitempty"`
	// EnableUDS enables Unix Domain Socket transport
	EnableUDS bool `json:"enable_uds,omitempty"`
	// EnableFD enables File Descriptor transport
	EnableFD bool `json:"enable_fd,omitempty"`
	// DualMode enables both UDS and FD transports for migration
	DualMode bool `json:"dual_mode,omitempty"`
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

// LoadConfig returns an empty configuration since runtime config is disabled
func LoadConfig(filename string) (*Config, error) {
	// Runtime configuration loading is disabled. Use build-time configuration instead.
	return &Config{}, nil
}

// SaveConfig is disabled since runtime config is disabled
func SaveConfig(config *Config, filename string) error {
	// Runtime configuration saving is disabled. Use build-time configuration instead.
	return nil
}
