package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// Create a subprocess instance
	// Use process ID from environment or default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "worker"
	}

	sp := subprocess.New(processID)

	// Register handlers for different message types
	sp.RegisterHandler("request", handleRequest)
	sp.RegisterHandler("ping", handlePing)

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
		// Send shutdown event
		sp.SendEvent("subprocess_shutdown", map[string]string{
			"id":     sp.ID,
			"reason": "signal received",
		})
		sp.Stop()
	}
}

// handleRequest handles request messages
func handleRequest(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Parse the request payload
	var req map[string]interface{}
	if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Process the request
	action, ok := req["action"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'action' field")
	}

	// Perform action
	var result interface{}
	switch action {
	case "echo":
		data := req["data"]
		result = map[string]interface{}{
			"action": "echo",
			"data":   data,
		}
	case "uppercase":
		data, ok := req["data"].(string)
		if !ok {
			return nil, fmt.Errorf("data must be a string for uppercase action")
		}
		result = map[string]interface{}{
			"action": "uppercase",
			"result": strings.ToUpper(data),
		}
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	// Create response
	response := &subprocess.Message{
		Type: subprocess.MessageTypeResponse,
		ID:   msg.ID,
	}

	// Marshal the result
	if result != nil {
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		response.Payload = jsonData
	}

	return response, nil
}

// handlePing handles ping messages
func handlePing(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Respond with pong
	response := &subprocess.Message{
		Type: subprocess.MessageTypeResponse,
		ID:   msg.ID,
	}

	pongData := map[string]string{"status": "pong"}
	jsonData, err := json.Marshal(pongData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pong response: %w", err)
	}
	response.Payload = jsonData

	return response, nil
}
