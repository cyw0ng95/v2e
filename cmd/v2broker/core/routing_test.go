package core

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

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
	t.Skip("Skipping stdin/stdout test - UDS-only transport")
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

func TestInvokeRPC(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a simple echo-like process for testing
	// Note: This test requires spawning an actual RPC process
	// For now, we test the error cases and timeout behavior

	testutils.Run(t, testutils.Level2, "Invalid target process", nil, func(t *testing.T, tx *gorm.DB) {
		// Try to invoke RPC on non-existent process
		_, err := broker.InvokeRPC("source", "nonexistent", "RPCTest", map[string]string{}, 100*time.Millisecond)
		if err == nil {
			t.Error("Expected error when invoking RPC on non-existent process")
		}
	})

	testutils.Run(t, testutils.Level2, "Timeout behavior", nil, func(t *testing.T, tx *gorm.DB) {
		t.Skip("Skipping stdin/stdout test - UDS-only transport")
	})
}

func TestLoadProcessesFromConfig(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	testutils.Run(t, testutils.Level2, "Config loading - should use build-time defaults", nil, func(t *testing.T, tx *gorm.DB) {
		err := broker.LoadProcessesFromConfig(nil)
		if err != nil {
			t.Errorf("Expected no error with config, got: %v", err)
		}
	})

	testutils.Run(t, testutils.Level2, "Nil config - should use build-time defaults", nil, func(t *testing.T, tx *gorm.DB) {
		err := broker.LoadProcessesFromConfig(nil)
		if err != nil {
			t.Errorf("Expected no error for nil config, got: %v", err)
		}
	})
}
