/*
Package main implements the broker service.

Refer to service.md for details about the RPC API Specification.
*/
package main

import (
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/core"
	"github.com/cyw0ng95/v2e/cmd/broker/perf"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// brokerRouter adapts the core Broker to the routing.Router interface used by perf.Optimizer.
type brokerRouter struct {
	b *core.Broker
}

func (r *brokerRouter) Route(msg *proc.Message, sourceProcess string) error {
	return r.b.RouteMessage(msg, sourceProcess)
}

func (r *brokerRouter) ProcessBrokerMessage(msg *proc.Message) error {
	return r.b.ProcessMessage(msg)
}

func main() {
	// Get config file from argv[1] or use default
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	// Load configuration
	config, err := common.LoadConfig(configFile)
	if err != nil {
		common.Error("Error loading config: %v", err)
		os.Exit(1)
	}

	// Set up logger with dual output (stdout + file) if log file is configured
	var logOutput io.Writer
	if config.Broker.LogFile != "" {
		// Ensure parent directory exists so opening the log file won't fail
		logDir := filepath.Dir(config.Broker.LogFile)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				common.Error("Error creating log directory '%s': %v", logDir, err)
				os.Exit(1)
			}
		}

		logFile, err := os.OpenFile(config.Broker.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			common.Error("Error opening log file '%s': %v", config.Broker.LogFile, err)
			os.Exit(1)
		}
		defer logFile.Close()
		logOutput = io.MultiWriter(os.Stdout, logFile)
	} else {
		logOutput = os.Stdout
	}

	// Set default logger output
	common.SetOutput(logOutput)
	// Set log level from config if present, default to InfoLevel
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
	common.SetLevel(logLevel)

	// Create broker instance
	broker := NewBroker()
	// Provide loaded config to broker so it can use configured settings when spawning
	broker.SetConfig(config)
	// Install a default SpawnAdapter that delegates to existing spawn methods.
	spawnAdapter := NewSpawnAdapter(broker)
	broker.SetSpawner(spawnAdapter)
	defer broker.Shutdown()

	// Set up broker logger with dual output and correct level
	brokerLogger := common.NewLogger(logOutput, "[BROKER] ", logLevel)
	broker.SetLogger(brokerLogger)

	// Load processes from configuration
	if err := broker.LoadProcessesFromConfig(config); err != nil {
		common.Error("Error loading processes from config: %v", err)
	}

	// Create and attach an optimizer using broker config (optional tuning)
	bufferCap := config.Broker.OptimizerBufferCap
	numWorkers := config.Broker.OptimizerNumWorkers
	statsInterval := time.Duration(config.Broker.OptimizerStatsIntervalMs) * time.Millisecond
	policy := config.Broker.OptimizerOfferPolicy
	offerTimeout := time.Duration(config.Broker.OptimizerOfferTimeoutMs) * time.Millisecond

	routerAdapter := &brokerRouter{b: broker}
	batchSize := config.Broker.OptimizerBatchSize
	flushMs := config.Broker.OptimizerFlushIntervalMs
	flushInterval := time.Duration(flushMs) * time.Millisecond
	opt := perf.NewWithParams(routerAdapter, bufferCap, numWorkers, statsInterval, policy, offerTimeout, batchSize, flushInterval)
	opt.SetLogger(brokerLogger)
	broker.SetOptimizer(opt)
	common.Info("Optimizer started: buffer=%d workers=%d policy=%s batch=%d flush_ms=%d", bufferCap, numWorkers, policy, batchSize, int(flushInterval/time.Millisecond))

	common.Info("Broker started, managing %d processes", len(config.Broker.Processes))

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
				common.Warn("Error processing broker message - Message ID: %s, Source: %s, Target: %s, Error: %v", msg.ID, msg.Source, msg.Target, err)
			} else {
				common.Debug("Successfully processed broker message - Message ID: %s, Source: %s, Target: %s", msg.ID, msg.Source, msg.Target)
			}
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	<-sigChan
	common.Info("Shutdown signal received, stopping broker...")
}
