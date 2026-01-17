package subprocess

import (
	"fmt"
	"os"

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
			sp.SendError("fatal", fmt.Errorf("subprocess error: %w", err))
			os.Exit(1)
		}
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
