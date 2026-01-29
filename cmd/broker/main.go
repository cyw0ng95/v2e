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

	"github.com/cyw0ng95/v2e/cmd/broker/perf"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
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
	// Get config file from argv[1] or use default
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	// Load configuration
	config, err := common.LoadConfig(configFile)
	if err != nil {
		common.Error(LogMsgErrorLoadingConfig, err)
		os.Exit(1)
	}

	// Set up logger with dual output (stdout + file) if log file is configured
	var logOutput io.Writer
	if config.Broker.LogFile != "" {
		// Ensure parent directory exists so opening the log file won't fail
		logDir := filepath.Dir(config.Broker.LogFile)
		if logDir != "." && logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				common.Error(LogMsgErrorCreatingLogDir, logDir, err)
				os.Exit(1)
			}
		}

		logFile, err := os.OpenFile(config.Broker.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			common.Error(LogMsgErrorOpeningLogFile, config.Broker.LogFile, err)
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
	// Configure transport based on configuration
	broker.ConfigureTransportFromConfig()
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
	routerAdapter := &brokerRouter{b: broker}

	optConfig := perf.Config{
		BufferCap:      config.Broker.OptimizerBufferCap,
		NumWorkers:     config.Broker.OptimizerNumWorkers,
		StatsInterval:  time.Duration(config.Broker.OptimizerStatsIntervalMs) * time.Millisecond,
		OfferPolicy:    config.Broker.OptimizerOfferPolicy,
		OfferTimeout:   time.Duration(config.Broker.OptimizerOfferTimeoutMs) * time.Millisecond,
		BatchSize:      config.Broker.OptimizerBatchSize,
		FlushInterval:  time.Duration(config.Broker.OptimizerFlushIntervalMs) * time.Millisecond,
		AdaptationFreq: time.Duration(config.Broker.OptimizerAdaptationFreqMs) * time.Millisecond,
	}

	opt := perf.NewWithConfig(routerAdapter, optConfig)
	opt.SetLogger(brokerLogger)

	if config.Broker.OptimizerEnableAdaptive {
		opt.EnableAdaptiveOptimization()
		common.Info("Adaptive optimization enabled (freq=%v)", optConfig.AdaptationFreq)
	}

	broker.SetOptimizer(opt)
	
	// Get actual values from optimizer metrics or use config (config might be 0/empty, handled by defaults)
	// We can trust NewWithConfig set defaults, but we don't have easy access to the final config struct inside opt
	// except via side channels. For logging, we'll just log what we have or query metrics.
	metrics := opt.Metrics()
	common.Info("Optimizer started: buffer=%v workers=%v policy=%s batch=%d flush=%v", 
		metrics["message_channel_buffer"], 
		metrics["active_workers"], 
		optConfig.OfferPolicy, 
		optConfig.BatchSize, 
		optConfig.FlushInterval)

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
