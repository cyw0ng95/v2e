package subprocess

import (
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler sets up signal handling for graceful shutdown
// Returns a channel that will receive signals
func SetupSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	return sigChan
}
