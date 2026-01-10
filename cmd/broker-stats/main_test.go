package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRPCGetMessageCount(t *testing.T) {
	// Create broker and send some messages
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Send a few messages
	msg1, _ := proc.NewRequestMessage("req-1", nil)
	broker.SendMessage(msg1)
	msg2, _ := proc.NewResponseMessage("resp-1", nil)
	broker.SendMessage(msg2)

	// Create handler
	handler := createGetMessageCountHandler(broker)

	// Create request message
	reqMsg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCGetMessageCount",
	}

	// Call handler
	ctx := context.Background()
	respMsg, err := handler(ctx, reqMsg)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if respMsg.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response message type, got %s", respMsg.Type)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respMsg.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	count, ok := result["total_count"].(float64)
	if !ok {
		t.Fatalf("Expected total_count to be a number, got %T", result["total_count"])
	}

	if count != 2 {
		t.Errorf("Expected total_count to be 2, got %v", count)
	}
}

func TestRPCGetMessageStats(t *testing.T) {
	// Create broker and send different types of messages
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Send different types of messages
	reqMsg, _ := proc.NewRequestMessage("req-1", nil)
	broker.SendMessage(reqMsg)
	respMsg, _ := proc.NewResponseMessage("resp-1", nil)
	broker.SendMessage(respMsg)
	eventMsg, _ := proc.NewEventMessage("event-1", nil)
	broker.SendMessage(eventMsg)

	// Create handler
	handler := createGetMessageStatsHandler(broker)

	// Create request message
	req := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCGetMessageStats",
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response message type, got %s", resp.Type)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response fields
	expectedFields := []string{
		"total_sent", "total_received", "request_count",
		"response_count", "event_count", "error_count",
		"first_message_time", "last_message_time",
	}

	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Expected field %s in response", field)
		}
	}

	// Check counts
	if result["total_sent"].(float64) != 3 {
		t.Errorf("Expected total_sent to be 3, got %v", result["total_sent"])
	}
	if result["request_count"].(float64) != 1 {
		t.Errorf("Expected request_count to be 1, got %v", result["request_count"])
	}
	if result["response_count"].(float64) != 1 {
		t.Errorf("Expected response_count to be 1, got %v", result["response_count"])
	}
	if result["event_count"].(float64) != 1 {
		t.Errorf("Expected event_count to be 1, got %v", result["event_count"])
	}
}

func TestBrokerStatsIntegration(t *testing.T) {
	// Create a buffer to capture subprocess output
	var output bytes.Buffer

	// Create a broker
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create subprocess
	sp := subprocess.New("broker-stats-test")
	sp.SetOutput(&output)

	// Register handlers
	sp.RegisterHandler("RPCGetMessageCount", createGetMessageCountHandler(broker))
	sp.RegisterHandler("RPCGetMessageStats", createGetMessageStatsHandler(broker))

	// Send some messages to the broker first
	msg1, _ := proc.NewRequestMessage("req-1", nil)
	broker.SendMessage(msg1)

	// Create input with RPC request
	requestMsg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCGetMessageCount",
	}
	requestData, _ := json.Marshal(requestMsg)
	input := bytes.NewBufferString(string(requestData) + "\n")
	sp.SetInput(input)

	// Run subprocess for a short time
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- sp.Run()
	}()

	// Wait for context or completion
	select {
	case <-ctx.Done():
		sp.Stop()
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			t.Logf("Subprocess completed with: %v", err)
		}
	}

	// Check output
	outputStr := output.String()
	if outputStr == "" {
		t.Fatal("Expected output from subprocess")
	}

	// Parse output lines
	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	foundResponse := false

	for scanner.Scan() {
		line := scanner.Text()
		var msg subprocess.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		if msg.ID == "RPCGetMessageCount" && msg.Type == subprocess.MessageTypeResponse {
			foundResponse = true
			var result map[string]interface{}
			if err := json.Unmarshal(msg.Payload, &result); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if _, ok := result["total_count"]; !ok {
				t.Error("Expected total_count in response")
			}
		}
	}

	if !foundResponse {
		t.Error("Expected to find RPCGetMessageCount response in output")
	}
}
