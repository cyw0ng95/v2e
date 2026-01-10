package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "broker-stats"
	}

	// Create broker instance
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create subprocess instance
	sp := subprocess.New(processID)

	// Register RPC handlers
	sp.RegisterHandler("RPCGetMessageCount", createGetMessageCountHandler(broker))
	sp.RegisterHandler("RPCGetMessageStats", createGetMessageStatsHandler(broker))

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

// createGetMessageCountHandler creates a handler for RPCGetMessageCount
func createGetMessageCountHandler(broker *proc.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get message count from broker
		count := broker.GetMessageCount()

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
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}

// createGetMessageStatsHandler creates a handler for RPCGetMessageStats
func createGetMessageStatsHandler(broker *proc.Broker) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Get message statistics from broker
		stats := broker.GetMessageStats()

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
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		respMsg.Payload = jsonData

		return respMsg, nil
	}
}
