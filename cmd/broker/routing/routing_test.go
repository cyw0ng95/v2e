package routing

import (
	"io"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/core"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestGenerateCorrelationID(t *testing.T) {
	broker := core.NewBroker()
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
	broker := core.NewBroker()
	defer broker.Shutdown()

	// Create a test process without spawning an OS process
	ir, pw := io.Pipe()
	// Drain stdin so writes do not block
	go func() { _, _ = io.Copy(io.Discard, ir) }()
	r, w := io.Pipe()
	p := core.NewTestProcess("test-process", core.ProcessStatusRunning, pw, r)
	broker.InsertProcessForTest(p)
	broker.StartProcessReaderForTest(p)
	// Close the stdout writer so the reader goroutine can exit cleanly on shutdown
	_ = w.Close()

	// Create a message with a target
	msg, err := proc.NewRequestMessage("test", map[string]string{"data": "value"})
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	msg.Target = "test-process"

	// Route the message
	t.Log("Routing message to test-process")
	err = broker.RouteMessage(msg, "sender")
	t.Log("RouteMessage returned")
	if err != nil {
		t.Errorf("Failed to route message: %v", err)
	}

	// Verify source was set
	if msg.Source != "sender" {
		t.Errorf("Expected source to be 'sender', got: %s", msg.Source)
	}

	// Clean up
	_ = broker.Kill("test-process")
}

func TestRouteMessage_NoTarget(t *testing.T) {
	broker := core.NewBroker()
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
	broker := core.NewBroker()
	defer broker.Shutdown()

	// Create a simple echo-like process for testing
	// Note: This test requires spawning an actual RPC process
	// For now, we test the error cases and timeout behavior

	t.Run("Invalid target process", func(t *testing.T) {
		// Try to invoke RPC on non-existent process
		_, err := broker.InvokeRPC("source", "nonexistent", "RPCTest", map[string]string{}, 100*time.Millisecond)
		if err == nil {
			t.Error("Expected error when invoking RPC on non-existent process")
		}
	})

	t.Run("Timeout behavior", func(t *testing.T) {
		// Create a test process that won't respond to RPC
		ir, pw := io.Pipe()
		// Drain stdin so writes do not block
		go func() { _, _ = io.Copy(io.Discard, ir) }()
		r, w := io.Pipe()
		p := core.NewTestProcess("test-rpc", core.ProcessStatusRunning, pw, r)
		broker.InsertProcessForTest(p)
		broker.StartProcessReaderForTest(p)
		// Close writer to avoid blocking scanner on shutdown
		_ = w.Close()
		defer broker.Kill("test-rpc")

		// Try to invoke RPC with very short timeout
		_, err := broker.InvokeRPC("source", "test-rpc", "RPCTest", map[string]string{}, 50*time.Millisecond)
		if err == nil {
			t.Error("Expected timeout error when process doesn't respond")
		}
	})
}

func TestLoadProcessesFromConfig(t *testing.T) {
	broker := core.NewBroker()
	defer broker.Shutdown()

	t.Run("Config loading - should use build-time defaults", func(t *testing.T) {
		config := &common.Config{}
		err := broker.LoadProcessesFromConfig(config)
		if err != nil {
			t.Errorf("Expected no error with config, got: %v", err)
		}
	})

	t.Run("Nil config - should use build-time defaults", func(t *testing.T) {
		err := broker.LoadProcessesFromConfig(nil)
		if err != nil {
			t.Errorf("Expected no error for nil config, got: %v", err)
		}
	})
}
