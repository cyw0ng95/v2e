package main

const (
	// Log messages
	LogMsgErrorLoadingConfig       = "Error loading config: %v"
	LogMsgErrorCreatingLogDir      = "Error creating log directory '%s': %v"
	LogMsgErrorOpeningLogFile      = "Error opening log file '%s': %v"
	LogMsgErrorLoadingProcesses    = "Error loading processes from config: %v"
	LogMsgAdaptiveOptEnabled       = "Adaptive optimization enabled (freq=%v)"
	LogMsgOptimizerStarted         = "Optimizer started: buffer=%v workers=%v policy=%s batch=%d flush=%v"
	LogMsgBrokerStarted            = "Broker started, managing %d processes"
	LogMsgErrorProcessingMessage   = "Error processing broker message - Message ID: %s, Source: %s, Target: %s, Error: %v"
	LogMsgSuccessProcessingMessage = "Successfully processed broker message - Message ID: %s, Source: %s, Target: %s"
	LogMsgShutdownSignal           = "Shutdown signal received, stopping broker..."
)
