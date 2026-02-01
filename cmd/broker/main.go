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

	"github.com/cyw0ng95/v2e/cmd/broker/perf"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// brokerRouteBridge captures the minimal broker surface needed by perf.Optimizer.
type brokerRouteBridge interface {
	RouteMessage(msg *proc.Message, sourceProcess string) error
	ProcessMessage(msg *proc.Message) error
}

// brokerRouter adapts the core Broker to the routing.Router interface used by perf.Optimizer.
type brokerRouter struct {
	b brokerRouteBridge
}

func (r *brokerRouter) Route(msg *proc.Message, sourceProcess string) error {
	return r.b.RouteMessage(msg, sourceProcess)
}

func (r *brokerRouter) ProcessBrokerMessage(msg *proc.Message) error {
	return r.b.ProcessMessage(msg)
}

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
	routerAdapter := &brokerRouter{b: broker}

	optConfig := perf.Config{
		BufferCap:      1000,                   // Default buffer capacity
		NumWorkers:     4,                      // Default number of workers
		StatsInterval:  100 * time.Millisecond, // Default stats interval
		OfferPolicy:    "drop",                 // Default offer policy
		OfferTimeout:   0,                      // Default offer timeout
		BatchSize:      1,                      // Default batch size
		FlushInterval:  10 * time.Millisecond,  // Default flush interval
		AdaptationFreq: 10 * time.Second,       // Default adaptation frequency
	}

	opt := perf.NewWithConfig(routerAdapter, optConfig)
	opt.SetLogger(logger)

	if false { // Adaptive optimization disabled by default
		opt.EnableAdaptiveOptimization()
		logger.Info(LogMsgAdaptiveOptEnabled, optConfig.AdaptationFreq)
	}

	broker.SetOptimizer(opt)

	// Get actual values from optimizer metrics or use config (config might be 0/empty, handled by defaults)
	// We can trust NewWithConfig set defaults, but we don't have easy access to the final config struct inside opt
	// except via side channels. For logging, we'll just log what we have or query metrics.
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
