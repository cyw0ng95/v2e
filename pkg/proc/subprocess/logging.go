package subprocess

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// SetupLogging sets up logging for a subprocess with the given process ID, logs directory, and log level.
// If buildLogRefresh is true and the process is the broker, it will remove the log directory first to ensure fresh logs.
func SetupLogging(processID string, logsDir string, logLevel common.LogLevel) (*common.Logger, error) {
	// Check if we need to refresh the log directory (remove and recreate)
	// Only apply this to the broker process
	if DefaultBuildLogRefresh() && processID == "broker" {
		// Remove the entire logs directory first to ensure fresh logs
		if err := os.RemoveAll(logsDir); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to remove log directory: %w", err)
		}
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

	logger.Debug("Logger initialized with Debug level")

	return logger, nil
}
