package main

import (
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestGenerateCorrelationID(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Generate two correlation IDs
	id1 := broker.GenerateCorrelationID()
	id2 := broker.GenerateCorrelationID()

	// They should be different
	if id1 == id2 {
		t.Errorf("Generated correlation IDs should be unique, got: %s and %s", id1, id2)
	}

	// They should start with "corr-"
	if len(id1) < 5 || id1[:5] != "corr-" {
		t.Errorf("Correlation ID should start with 'corr-', got: %s", id1)
	}
}

func TestRouteMessage_WithTarget(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Spawn a simple RPC process (just use the echo command for testing)
	info, err := broker.SpawnRPC("test-process", "cat")
	if err != nil {
		t.Fatalf("Failed to spawn test process: %v", err)
	}

	// Wait a bit for process to start
	time.Sleep(100 * time.Millisecond)

	// Create a message with a target
	msg, err := proc.NewRequestMessage("test", map[string]string{"data": "value"})
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	msg.Target = "test-process"

	// Route the message
	err = broker.RouteMessage(msg, "sender")
	if err != nil {
		t.Errorf("Failed to route message: %v", err)
	}

	// Verify source was set
	if msg.Source != "sender" {
		t.Errorf("Expected source to be 'sender', got: %s", msg.Source)
	}

	// Clean up
	broker.Kill(info.ID)
}

func TestRouteMessage_NoTarget(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a message without a target
	msg, err := proc.NewRequestMessage("test", map[string]string{"data": "value"})
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Route the message (should go to broker's message channel)
	err = broker.RouteMessage(msg, "sender")
	if err != nil {
		t.Errorf("Failed to route message: %v", err)
	}

	// Verify source was set
	if msg.Source != "sender" {
		t.Errorf("Expected source to be 'sender', got: %s", msg.Source)
	}
}
