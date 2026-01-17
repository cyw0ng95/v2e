package subprocess

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// SetupLogging initializes logging for a subprocess
// It reads config from config.json and sets up logging to both stderr and a file
func SetupLogging(processID string) (*common.Logger, error) {
	// Load configuration
	config, err := common.LoadConfig("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Determine log level
	logLevel := common.InfoLevel
	if config.Logging.Level != "" {
		switch config.Logging.Level {
		case "debug":
			logLevel = common.DebugLevel
		case "info":
			logLevel = common.InfoLevel
		case "warn":
			logLevel = common.WarnLevel
		case "error":
			logLevel = common.ErrorLevel
		}
	}

	// Determine log directory
	logsDir := "./logs"
	if config.Logging.Dir != "" {
		logsDir = config.Logging.Dir
	} else if config.Broker.LogsDir != "" {
		logsDir = config.Broker.LogsDir
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file path
	logFile := filepath.Join(logsDir, fmt.Sprintf("%s.log", processID))

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// For RPC subprocesses, log to stderr and file (not stdout, since stdout is used for RPC messages)
	multiWriter := io.MultiWriter(os.Stderr, file)

	// Create logger with the multi-writer
	logger := common.NewLogger(multiWriter, fmt.Sprintf("[%s] ", processID), logLevel)

	return logger, nil
}
