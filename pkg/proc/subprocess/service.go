package subprocess

import (
	"flag"
	"fmt"
	"os"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// Flags holds the parsed command-line flags
type Flags struct {
	ProcessID string
}

// Service encapsulates the common state for a service
type Service struct {
	ID     string
	Config interface{}
	Logger *common.Logger
	Proc   *Subprocess
}

// ParseFlags parses the standard flags for a service
func ParseFlags(defaultID string) *Flags {
	f := &Flags{}
	// Check if flags are already defined to avoid redefinition panic in tests or if called multiple times
	if flag.Lookup("id") == nil {
		flag.StringVar(&f.ProcessID, "id", defaultID, "Process ID")
		flag.Parse()
	} else {
		// If flags are already defined, we need to retrieve their values
		// This is a bit hacky but safe for our use case where we control the main entry point
		f.ProcessID = flag.Lookup("id").Value.String()
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

	// Use empty config since runtime config is disabled
	config := struct{}{}

	// Setup Logger - use build-time configured log level and directory (no runtime config overrides for these)
	logLevel := DefaultBuildLogLevel()
	logDir := DefaultBuildLogDir()

	logger, err := SetupLogging(flags.ProcessID, logDir, logLevel)
	if err != nil {
		// Fallback to stderr logger
		logger = common.NewLogger(os.Stderr, fmt.Sprintf("[%s] ", flags.ProcessID), logLevel)
		logger.Error(common.LogMsgLoggerSetupFailed, err)
	}

	// Create Subprocess with UDS transport
	socketPath := fmt.Sprintf("%s_%s.sock", DefaultProcUDSBasePath(), flags.ProcessID)
	sp, err := NewWithUDS(flags.ProcessID, socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create subprocess with UDS transport: %w", err)
	}

	logger.Info(common.LogMsgServiceStarted, flags.ProcessID)
	logger.Info(common.LogMsgConfigLoaded, "build-time (runtime config disabled)")

	return &Service{
		ID:     flags.ProcessID,
		Config: config,
		Logger: logger,
		Proc:   sp,
	}, nil
}

// Run runs the service subprocess
func (s *Service) Run() {
	RunWithDefaults(s.Proc, s.Logger)
}
