package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "broker"
	}

	// Set up logger (send to stderr to keep stdout clean for RPC)
	common.SetLevel(common.InfoLevel)

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create subprocess instance for RPC communication
	sp := subprocess.New(processID)

	// Register RPC handlers
	sp.RegisterHandler("RPCSpawn", createSpawnHandler(broker))
	sp.RegisterHandler("RPCSpawnRPC", createSpawnRPCHandler(broker))
	sp.RegisterHandler("RPCGetProcess", createGetProcessHandler(broker))
	sp.RegisterHandler("RPCListProcesses", createListProcessesHandler(broker))
	sp.RegisterHandler("RPCKill", createKillHandler(broker))

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
func createSpawnHandler(broker *proc.Broker) subprocess.Handler {
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
		info, err := broker.Spawn(req.ID, req.Command, req.Args...)
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
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createSpawnRPCHandler creates a handler for RPCSpawnRPC
func createSpawnRPCHandler(broker *proc.Broker) subprocess.Handler {
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
		info, err := broker.SpawnRPC(req.ID, req.Command, req.Args...)
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
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetProcessHandler creates a handler for RPCGetProcess
func createGetProcessHandler(broker *proc.Broker) subprocess.Handler {
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
		info, err := broker.GetProcess(req.ID)
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
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createListProcessesHandler creates a handler for RPCListProcesses
func createListProcessesHandler(broker *proc.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get all processes
		processes := broker.ListProcesses()

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
		jsonData, err := json.Marshal(map[string]interface{}{
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
func createKillHandler(broker *proc.Broker) subprocess.Handler {
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
		err := broker.Kill(req.ID)
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
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}
