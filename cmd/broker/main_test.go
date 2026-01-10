package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRPCSpawn(t *testing.T) {
	// Create broker
	broker := proc.NewBroker()
	defer broker.Shutdown()

	// Create handler
	handler := createSpawnHandler(broker)

	// Create request message
	payload, _ := json.Marshal(map[string]interface{}{
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
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
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
	broker := proc.NewBroker()
	defer broker.Shutdown()

	_, err := broker.Spawn("test-process", "sleep", "1")
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Create handler
	handler := createGetProcessHandler(broker)

	// Create request message
	payload, _ := json.Marshal(map[string]string{
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
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
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
	broker := proc.NewBroker()
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
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
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
	broker := proc.NewBroker()
	defer broker.Shutdown()

	_, err := broker.Spawn("test-kill", "sleep", "10")
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Create handler
	handler := createKillHandler(broker)

	// Create request message
	payload, _ := json.Marshal(map[string]string{
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
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}

	if result["id"] != "test-kill" {
		t.Errorf("Expected id=test-kill, got %s", result["id"])
	}
}
