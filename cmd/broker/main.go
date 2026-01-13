/*
Package main implements the broker service.

RPC API Specification:

Broker Service
====================

Service Type: RPC (stdin/stdout message passing)
Description: Central process manager and message router for the v2e system.
             Spawns and manages all subprocess services, routes RPC messages between services.

Available RPC Methods:
---------------------

1. RPCSpawn
   Description: Spawns a new subprocess with specified command and arguments
   Request Parameters:
     - id (string, required): Unique identifier for the process
     - command (string, required): Command to execute
     - args ([]string, optional): Command arguments
   Response:
     - id (string): Process identifier
     - pid (int): Process ID
     - status (string): Process status ("running", "exited", "failed")
   Errors:
     - Missing ID: Process ID is required
     - Duplicate ID: Process with this ID already exists
     - Spawn failure: Failed to start the process
   Example:
     Request:  {"id": "worker-1", "command": "./worker", "args": ["--config", "config.json"]}
     Response: {"id": "worker-1", "pid": 12345, "status": "running"}

2. RPCSpawnRPC
   Description: Spawns a subprocess with RPC support (stdin/stdout pipes for message passing)
   Request Parameters:
     - id (string, required): Unique identifier for the process
     - command (string, required): Command to execute
     - args ([]string, optional): Command arguments
   Response:
     - id (string): Process identifier
     - pid (int): Process ID
     - status (string): Process status
   Errors:
     - Missing ID: Process ID is required
     - Duplicate ID: Process with this ID already exists
     - Spawn failure: Failed to start the process
   Example:
     Request:  {"id": "cve-remote", "command": "./cve-remote", "args": []}
     Response: {"id": "cve-remote", "pid": 12346, "status": "running"}

3. RPCKill
   Description: Terminates a running subprocess
   Request Parameters:
     - id (string, required): Process identifier to kill
   Response:
     - success (bool): true if process was terminated
   Errors:
     - Missing ID: Process ID is required
     - Not found: Process with this ID does not exist
   Example:
     Request:  {"id": "worker-1"}
     Response: {"success": true}

4. RPCGetProcess
   Description: Gets information about a specific subprocess
   Request Parameters:
     - id (string, required): Process identifier
   Response:
     - id (string): Process identifier
     - pid (int): Process ID
     - command (string): Command being executed
     - args ([]string): Command arguments
     - status (string): Process status
     - start_time (string): When the process started
   Errors:
     - Missing ID: Process ID is required
     - Not found: Process with this ID does not exist
   Example:
     Request:  {"id": "worker-1"}
     Response: {"id": "worker-1", "pid": 12345, "command": "./worker", "status": "running", ...}

5. RPCListProcesses
   Description: Lists all managed subprocesses
   Request Parameters: None
   Response:
     - processes ([]object): Array of process information objects
   Errors: None
   Example:
     Request:  {}
     Response: {"processes": [{"id": "worker-1", "pid": 12345, "status": "running"}, ...]}

6. RPCGetMessageStats
   Description: Gets statistics about messages processed by the broker
   Request Parameters: None
   Response:
     - total_sent (int): Total messages sent
     - total_received (int): Total messages received
     - request_count (int): Number of request messages
     - response_count (int): Number of response messages
     - event_count (int): Number of event messages
     - error_count (int): Number of error messages
     - first_message_time (string): Timestamp of first message
     - last_message_time (string): Timestamp of last message
   Errors: None
   Example:
     Request:  {}
     Response: {"total_sent": 150, "total_received": 120, "request_count": 50, ...}

7. RPCGetMessageCount
   Description: Gets the total number of messages processed (sent + received)
   Request Parameters: None
   Response:
     - count (int): Total message count
   Errors: None
   Example:
     Request:  {}
     Response: {"count": 270}

Notes:
------
- Broker is the only entry point for process management in the v2e system
- All inter-process communication is routed through the broker
- Processes can be configured for auto-restart on exit
- Message routing uses correlation IDs to match requests with responses
- Broker loads process configuration from config.json at startup

*/
package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyw0ng95/v2e/pkg/common"
)

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
		logFile, err := os.OpenFile(config.Broker.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			common.Error("Error opening log file: %v", err)
			os.Exit(1)
		}
		defer logFile.Close()
		logOutput = io.MultiWriter(os.Stdout, logFile)
	} else {
		logOutput = os.Stdout
	}

	// Set default logger output
	common.SetOutput(logOutput)
	common.SetLevel(common.InfoLevel)

	// Create broker instance
	broker := NewBroker()
	defer broker.Shutdown()

	// Set up broker logger with dual output
	brokerLogger := common.NewLogger(logOutput, "[BROKER] ", common.InfoLevel)
	broker.SetLogger(brokerLogger)

	// Load processes from configuration
	if err := broker.LoadProcessesFromConfig(config); err != nil {
		common.Error("Error loading processes from config: %v", err)
	}

	common.Info("Broker started, managing %d processes", len(config.Broker.Processes))

	// Start message processing goroutine
	// This processes RPC requests directed at the broker
	go func() {
		for {
			select {
			case msg := <-broker.messages:
				// Process messages directed at the broker
				if err := broker.ProcessMessage(msg); err != nil {
					common.Warn("Error processing broker message: %v", err)
				}
			case <-broker.Context().Done():
				return
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
