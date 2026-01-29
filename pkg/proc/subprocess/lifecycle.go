package subprocess

import (
	"fmt"
	"os"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// RunWithDefaults runs a subprocess with default signal handling and error handling
// This is a convenience function that wraps the common pattern of running a subprocess
func RunWithDefaults(sp *Subprocess, logger *common.Logger) {
	// Set up signal handling
	sigChan := SetupSignalHandler()

	// Run the subprocess in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- sp.Run()
	}()

	// Wait for either completion or signal
	select {
	case err := <-errChan:
		if err != nil {
			if logger != nil {
				logger.Error("Subprocess error: %v", err)
			}
			// In test environments, avoid os.Exit which can interfere with test execution
			if logger != nil {
				logger.Error("Fatal subprocess error: %v", err)
			}
			// Only send error if output is available (not during test coverage)
			if sp.output != os.Stdout && sp.output != os.Stderr {
				sp.SendError("fatal", fmt.Errorf("subprocess error: %w", err))
			}
			// In production, exit, but in tests avoid os.Exit to allow proper test cleanup
			// Check if we're in a test environment to avoid os.Exit during coverage
			// os.Exit can interfere with coverage report generation
			isTesting := len(os.Args) > 0 && strings.Contains(os.Args[0], ".test")
			if isTesting {
				// In test mode, just return the error to allow proper test cleanup
				return
			} else {
				os.Exit(1)
			}
		}
	case <-sigChan:
	case <-sigChan:
		if logger != nil {
			logger.Info("Signal received, shutting down...")
		}
		sp.SendEvent("subprocess_shutdown", map[string]string{
			"id":     sp.ID,
			"reason": "signal received",
		})
		sp.Stop()
	}
}
