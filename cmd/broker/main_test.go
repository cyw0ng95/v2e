package main

import (
	"context"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRPCSpawn(t *testing.T) {
	// Create broker
	broker := NewBroker()
	defer broker.Shutdown()

	// Create handler
	handler := createSpawnHandler(broker)

	// Create request message
	payload, _ := sonic.Marshal(map[string]interface{}{
		"id":      "test-echo",
		"command": "echo",
		"args":    []string{"hello", "world"},
	})

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCSpawn",
		Payload: payload,
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// Check results
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	var result map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["id"] != "test-echo" {
		t.Errorf("Expected id=test-echo, got %s", result["id"])
	}

	if result["command"] != "echo" {
		t.Errorf("Expected command='echo', got %s", result["command"])
	}

	// Verify process exists
	info, err := broker.GetProcess("test-echo")
	if err != nil {
		t.Errorf("Process not found: %v", err)
	}
	if info.ID != "test-echo" {
		t.Errorf("Process ID mismatch: %s", info.ID)
	}
}

func TestRPCGetProcess(t *testing.T) {
	// Create broker and spawn a process
	broker := NewBroker()
	defer broker.Shutdown()

	_, err := broker.Spawn("test-process", "sleep", "1")
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Create handler
	handler := createGetProcessHandler(broker)

	// Create request message
	payload, _ := sonic.Marshal(map[string]string{
		"id": "test-process",
	})

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetProcess",
		Payload: payload,
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// Check results
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	var result map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["id"] != "test-process" {
		t.Errorf("Expected id=test-process, got %s", result["id"])
	}

	if result["command"] != "sleep" {
		t.Errorf("Expected command='sleep', got %s", result["command"])
	}
}

func TestRPCListProcesses(t *testing.T) {
	// Create broker and spawn processes
	broker := NewBroker()
	defer broker.Shutdown()

	broker.Spawn("proc-1", "echo", "test1")
	broker.Spawn("proc-2", "echo", "test2")

	// Create handler
	handler := createListProcessesHandler(broker)

	// Create request message
	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCListProcesses",
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// Check results
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	var result map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	count := int(result["count"].(float64))
	if count < 2 {
		t.Errorf("Expected at least 2 processes, got %d", count)
	}

	processes := result["processes"].([]interface{})
	if len(processes) < 2 {
		t.Errorf("Expected at least 2 processes in list, got %d", len(processes))
	}
}

func TestRPCKill(t *testing.T) {
	// Create broker and spawn a long-running process
	broker := NewBroker()
	defer broker.Shutdown()

	_, err := broker.Spawn("test-kill", "sleep", "10")
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Create handler
	handler := createKillHandler(broker)

	// Create request message
	payload, _ := sonic.Marshal(map[string]string{
		"id": "test-kill",
	})

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCKill",
		Payload: payload,
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// Check results
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	var result map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}

	if result["id"] != "test-kill" {
		t.Errorf("Expected id=test-kill, got %s", result["id"])
	}
}

func TestRPCGetMessageCount(t *testing.T) {
	// Create broker and send some messages
	broker := NewBroker()
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
	if err := sonic.Unmarshal(respMsg.Payload, &result); err != nil {
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
	broker := NewBroker()
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
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
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
	if totalSent, ok := result["total_sent"].(float64); !ok || totalSent != 3 {
		t.Errorf("Expected total_sent to be 3, got %v", result["total_sent"])
	}
	if requestCount, ok := result["request_count"].(float64); !ok || requestCount != 1 {
		t.Errorf("Expected request_count to be 1, got %v", result["request_count"])
	}
	if responseCount, ok := result["response_count"].(float64); !ok || responseCount != 1 {
		t.Errorf("Expected response_count to be 1, got %v", result["response_count"])
	}
	if eventCount, ok := result["event_count"].(float64); !ok || eventCount != 1 {
		t.Errorf("Expected event_count to be 1, got %v", result["event_count"])
	}
}
