package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/broker"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "broker"
	}

	// Load configuration
	config, err := common.LoadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Set up logger with dual output (stdout + file) if log file is configured
	var logOutput io.Writer
	if config.Broker.LogFile != "" {
		logFile, err := os.OpenFile(config.Broker.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
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
	brokerInstance := broker.NewBroker()
	defer brokerInstance.Shutdown()

	// Set up broker logger with dual output
	brokerLogger := common.NewLogger(logOutput, "[BROKER] ", common.InfoLevel)
	brokerInstance.SetLogger(brokerLogger)

	// Load processes from configuration
	if err := brokerInstance.LoadProcessesFromConfig(config); err != nil {
		common.Error("Error loading processes from config: %v", err)
	}

	// Create subprocess instance for RPC communication
	sp := subprocess.New(processID)

	// Register RPC handlers
	sp.RegisterHandler("RPCSpawn", createSpawnHandler(brokerInstance))
	sp.RegisterHandler("RPCSpawnRPC", createSpawnRPCHandler(brokerInstance))
	sp.RegisterHandler("RPCGetProcess", createGetProcessHandler(brokerInstance))
	sp.RegisterHandler("RPCListProcesses", createListProcessesHandler(brokerInstance))
	sp.RegisterHandler("RPCKill", createKillHandler(brokerInstance))
	sp.RegisterHandler("RPCGetMessageCount", createGetMessageCountHandler(brokerInstance))
	sp.RegisterHandler("RPCGetMessageStats", createGetMessageStatsHandler(brokerInstance))
	sp.RegisterHandler("RPCRegisterEndpoint", createRegisterEndpointHandler(brokerInstance))
	sp.RegisterHandler("RPCGetEndpoints", createGetEndpointsHandler(brokerInstance))
	sp.RegisterHandler("RPCGetAllEndpoints", createGetAllEndpointsHandler(brokerInstance))

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run the subprocess in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- sp.Run()
	}()

	// Wait for either completion or signal
	select {
	case err := <-errChan:
		if err != nil {
			sp.SendError("fatal", fmt.Errorf("subprocess error: %w", err))
			os.Exit(1)
		}
	case <-sigChan:
		sp.SendEvent("subprocess_shutdown", map[string]string{
			"id":     sp.ID,
			"reason": "signal received",
		})
		sp.Stop()
	}
}

// createSpawnHandler creates a handler for RPCSpawn
func createSpawnHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			ID      string   `json:"id"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		}

		if req.ID == "" {
			return nil, fmt.Errorf("id is required")
		}
		if req.Command == "" {
			return nil, fmt.Errorf("command is required")
		}

		// Spawn the process
		info, err := b.Spawn(req.ID, req.Command, req.Args...)
		if err != nil {
			return nil, fmt.Errorf("failed to spawn process: %w", err)
		}

		// Create response
		result := map[string]interface{}{
			"id":      info.ID,
			"pid":     info.PID,
			"command": info.Command,
			"status":  info.Status,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createSpawnRPCHandler creates a handler for RPCSpawnRPC
func createSpawnRPCHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			ID      string   `json:"id"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		}

		if req.ID == "" {
			return nil, fmt.Errorf("id is required")
		}
		if req.Command == "" {
			return nil, fmt.Errorf("command is required")
		}

		// Spawn the RPC process
		info, err := b.SpawnRPC(req.ID, req.Command, req.Args...)
		if err != nil {
			return nil, fmt.Errorf("failed to spawn RPC process: %w", err)
		}

		// Create response
		result := map[string]interface{}{
			"id":      info.ID,
			"pid":     info.PID,
			"command": info.Command,
			"status":  info.Status,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetProcessHandler creates a handler for RPCGetProcess
func createGetProcessHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		}

		if req.ID == "" {
			return nil, fmt.Errorf("id is required")
		}

		// Get process info
		info, err := b.GetProcess(req.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get process: %w", err)
		}

		// Create response
		result := map[string]interface{}{
			"id":        info.ID,
			"pid":       info.PID,
			"command":   info.Command,
			"status":    info.Status,
			"exit_code": info.ExitCode,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createListProcessesHandler creates a handler for RPCListProcesses
func createListProcessesHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get all processes
		processes := b.ListProcesses()

		// Convert to response format
		result := make([]map[string]interface{}, 0, len(processes))
		for _, p := range processes {
			result = append(result, map[string]interface{}{
				"id":        p.ID,
				"pid":       p.PID,
				"command":   p.Command,
				"status":    p.Status,
				"exit_code": p.ExitCode,
			})
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(map[string]interface{}{
			"processes": result,
			"count":     len(result),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createKillHandler creates a handler for RPCKill
func createKillHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		}

		if req.ID == "" {
			return nil, fmt.Errorf("id is required")
		}

		// Kill the process
		err := b.Kill(req.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to kill process: %w", err)
		}

		// Create response
		result := map[string]interface{}{
			"success": true,
			"id":      req.ID,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetMessageCountHandler creates a handler for RPCGetMessageCount
func createGetMessageCountHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get message count from broker
		count := b.GetMessageCount()

		// Create response
		result := map[string]interface{}{
			"total_count": count,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetMessageStatsHandler creates a handler for RPCGetMessageStats
func createGetMessageStatsHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get message statistics from broker
		stats := b.GetMessageStats()

		// Create response with all statistics
		result := map[string]interface{}{
			"total_sent":         stats.TotalSent,
			"total_received":     stats.TotalReceived,
			"request_count":      stats.RequestCount,
			"response_count":     stats.ResponseCount,
			"event_count":        stats.EventCount,
			"error_count":        stats.ErrorCount,
			"first_message_time": stats.FirstMessageTime,
			"last_message_time":  stats.LastMessageTime,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createRegisterEndpointHandler creates a handler for RPCRegisterEndpoint
func createRegisterEndpointHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			ProcessID string `json:"process_id"`
			Endpoint  string `json:"endpoint"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		}

		if req.ProcessID == "" {
			return nil, fmt.Errorf("process_id is required")
		}
		if req.Endpoint == "" {
			return nil, fmt.Errorf("endpoint is required")
		}

		// Register the endpoint
		b.RegisterEndpoint(req.ProcessID, req.Endpoint)

		// Create response
		result := map[string]interface{}{
			"success":    true,
			"process_id": req.ProcessID,
			"endpoint":   req.Endpoint,
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetEndpointsHandler creates a handler for RPCGetEndpoints
func createGetEndpointsHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			ProcessID string `json:"process_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			return nil, fmt.Errorf("failed to parse request: %w", err)
		}

		if req.ProcessID == "" {
			return nil, fmt.Errorf("process_id is required")
		}

		// Get endpoints for the process
		endpoints := b.GetEndpoints(req.ProcessID)

		// Create response
		result := map[string]interface{}{
			"process_id": req.ProcessID,
			"endpoints":  endpoints,
			"count":      len(endpoints),
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetAllEndpointsHandler creates a handler for RPCGetAllEndpoints
func createGetAllEndpointsHandler(b *broker.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get all endpoints
		allEndpoints := b.GetAllEndpoints()

		// Create response
		result := map[string]interface{}{
			"endpoints": allEndpoints,
			"count":     len(allEndpoints),
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type: subprocess.MessageTypeResponse,
			ID:   msg.ID,
		}

		// Marshal the result
		jsonData, err := sonic.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}
