package subprocess

import (
	"flag"
	"fmt"
	"os"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// Flags holds the parsed command-line flags
type Flags struct {
	ProcessID   string
	ConfigFile  string
	RPCInputFD  int
	RPCOutputFD int
	LogLevel    string
}

// Service encapsulates the common state for a service
type Service struct {
	ID     string
	Config *common.Config
	Logger *common.Logger
	Proc   *Subprocess
}

// ParseFlags parses the standard flags for a service
func ParseFlags(defaultID string) *Flags {
	f := &Flags{}
	// Check if flags are already defined to avoid redefinition panic in tests or if called multiple times
	if flag.Lookup("id") == nil {
		flag.StringVar(&f.ProcessID, "id", defaultID, "Process ID")
		flag.StringVar(&f.ConfigFile, "config", common.DefaultConfigFile, "Configuration file path")
		flag.IntVar(&f.RPCInputFD, "rpc-in", -1, "RPC input file descriptor")
		flag.IntVar(&f.RPCOutputFD, "rpc-out", -1, "RPC output file descriptor")
		flag.StringVar(&f.LogLevel, "log-level", "", "Log level override")
		flag.Parse()
	} else {
		// If flags are already defined, we need to retrieve their values
		// This is a bit hacky but safe for our use case where we control the main entry point
		f.ProcessID = flag.Lookup("id").Value.String()
		f.ConfigFile = flag.Lookup("config").Value.String()
		// For int flags, it's more complex to get typed value back without reflection or type assertion,
		// but since we are the ones defining them, we can assume standard usage.
		// However, for simplicity and correctness in main(), we assume ParseFlags is called once.
		// The check above is mainly for tests.
	}

	// If ID is still default or empty after parsing (or not parsing if defined), ensure we have a valid ID
	if f.ProcessID == "" {
		f.ProcessID = defaultID
	}

	return f
}

// NewService initializes a service with standard flags and configuration
func NewService(defaultID string) (*Service, error) {
	flags := ParseFlags(defaultID)

	// Load configuration
	config, err := common.LoadConfig(flags.ConfigFile)
	if err != nil {
		if config == nil {
			config = &common.Config{}
		}
	}

	// Setup Logger
	logLevel := common.InfoLevel
	if flags.LogLevel != "" {
		logLevel = parseLogLevel(flags.LogLevel)
	} else if config.Logging.Level != "" {
		logLevel = parseLogLevel(config.Logging.Level)
	}

	logDir := common.DefaultLogsDir
	if config.Logging.Dir != "" {
		logDir = config.Logging.Dir
	} else if config.Broker.LogsDir != "" {
		logDir = config.Broker.LogsDir
	}

	logger, err := SetupLogging(flags.ProcessID, logDir, logLevel)
	if err != nil {
		// Fallback to stderr logger
		logger = common.NewLogger(os.Stderr, fmt.Sprintf("[%s] ", flags.ProcessID), logLevel)
		logger.Error(common.LogMsgLoggerSetupFailed, err)
	}

	// Create Subprocess
	var sp *Subprocess
	if flags.RPCInputFD >= 0 && flags.RPCOutputFD >= 0 {
		sp = NewWithFDs(flags.ProcessID, flags.RPCInputFD, flags.RPCOutputFD)
	} else {
		sp = New(flags.ProcessID)
	}

	logger.Info(common.LogMsgServiceStarted, flags.ProcessID)
	if flags.ConfigFile != "" {
		logger.Info(common.LogMsgConfigLoaded, flags.ConfigFile)
	}

	return &Service{
		ID:     flags.ProcessID,
		Config: config,
		Logger: logger,
		Proc:   sp,
	}, nil
}

func parseLogLevel(level string) common.LogLevel {
	switch level {
	case "debug":
		return common.DebugLevel
	case "info":
		return common.InfoLevel
	case "warn":
		return common.WarnLevel
	case "error":
		return common.ErrorLevel
	default:
		return common.InfoLevel
	}
}

// Run runs the service subprocess
func (s *Service) Run() {
	RunWithDefaults(s.Proc, s.Logger)
}
