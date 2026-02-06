/*
Package main implements the broker service.

Refer to service.md for details about the RPC API Specification.
*/
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/cmd/v2broker/perf"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// No config needed since runtime config is disabled

	// Use subprocess package for logging to ensure build-time log level and directory from .config is used
	logLevel := subprocess.DefaultBuildLogLevel()
	logDir := subprocess.DefaultBuildLogDir()
	logger, err := subprocess.SetupLogging("broker", logDir, logLevel)
	if err != nil {
		fallbackLogger := common.NewLogger(os.Stderr, "[BROKER] ", logLevel)
		fallbackLogger.Error("Failed to setup logging: %v", err)
		os.Exit(1)
	}
	// Use the logger's output writer

	// Create broker instance
	broker := NewBroker()
	// Placeholder for config if needed in future
	// Currently no config is passed to broker
	// Install a default SpawnAdapter that delegates to existing spawn methods.
	spawnAdapter := NewSpawnAdapter(broker)
	broker.SetSpawner(spawnAdapter)
	defer broker.Shutdown()

	// Use the subprocess logger as the broker logger
	broker.SetLogger(logger)

	// Load processes from configuration
	if err := broker.LoadProcessesFromConfig(nil); err != nil {
		logger.Error(LogMsgErrorLoadingProcesses, err)
	}

	// Create and attach an optimizer using broker config (optional tuning)
	// Broker directly satisfies routing.Router interface
	// Configuration values come from build-time ldflags set by vconfig

	optConfig := perf.Config{
		BufferCap:      buildOptimizerBufferValue(),  // Build-time configurable
		NumWorkers:     buildOptimizerWorkersValue(), // Build-time configurable
		StatsInterval:  100 * time.Millisecond,       // Default stats interval
		OfferPolicy:    buildOptimizerPolicyValue(),  // Build-time configurable
		OfferTimeout:   0,                            // Default offer timeout
		BatchSize:      buildOptimizerBatchValue(),   // Build-time configurable
		FlushInterval:  buildOptimizerFlushValue(),   // Build-time configurable
		AdaptationFreq: 10 * time.Second,             // Default adaptation frequency
	}

	opt := perf.NewWithConfig(broker, optConfig)
	broker.SetOptimizer(opt)

	metrics := opt.Metrics()
	logger.Info(LogMsgOptimizerStarted,
		metrics["message_channel_buffer"],
		metrics["active_workers"],
		optConfig.OfferPolicy,
		optConfig.BatchSize,
		optConfig.FlushInterval)

	logger.Info("Broker started, processes will be loaded based on configuration")

	// Start message processing goroutine
	// This processes RPC requests directed at the broker
	go func() {
		for {
			msg, err := broker.ReceiveMessage(broker.Context())
			if err != nil {
				return
			}
			// Process messages directed at the broker
			if err := broker.ProcessMessage(msg); err != nil {
				logger.Warn(LogMsgErrorProcessingMessage, msg.ID, msg.Source, msg.Target, err)
			} else {
				logger.Debug(LogMsgSuccessProcessingMessage, msg.ID, msg.Source, msg.Target)
			}
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	<-sigChan
	logger.Info(LogMsgShutdownSignal)
}
