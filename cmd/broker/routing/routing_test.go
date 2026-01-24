package routing

import (
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

	// Spawn a simple RPC process (use sleep to keep it alive for testing)
	info, err := broker.SpawnRPC("test-process", "sleep", "60")
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
		// Spawn a simple process that won't respond to RPC
		info, err := broker.SpawnRPC("test-rpc", "sleep", "60")
		if err != nil {
			t.Fatalf("Failed to spawn test process: %v", err)
		}
		defer broker.Kill(info.ID)

		time.Sleep(100 * time.Millisecond)

		// Try to invoke RPC with very short timeout
		_, err = broker.InvokeRPC("source", "test-rpc", "RPCTest", map[string]string{}, 50*time.Millisecond)
		if err == nil {
			t.Error("Expected timeout error when process doesn't respond")
		}
	})
}

func TestLoadProcessesFromConfig(t *testing.T) {
	broker := core.NewBroker()
	defer broker.Shutdown()

	t.Run("Nil config", func(t *testing.T) {
		err := broker.LoadProcessesFromConfig(nil)
		if err != nil {
			t.Errorf("Expected no error for nil config, got: %v", err)
		}
	})

	t.Run("Empty processes", func(t *testing.T) {
		config := &common.Config{}
		err := broker.LoadProcessesFromConfig(config)
		if err != nil {
			t.Errorf("Expected no error for empty processes, got: %v", err)
		}
	})

	t.Run("Invalid process config - missing ID", func(t *testing.T) {
		config := &common.Config{
			Broker: common.BrokerConfig{
				Processes: []common.ProcessConfig{
					{
						// Missing ID
						Command: "echo",
						Args:    []string{"test"},
					},
				},
			},
		}
		err := broker.LoadProcessesFromConfig(config)
		// Should not error, just skip the invalid config
		if err != nil {
			t.Errorf("Expected no error when skipping invalid config, got: %v", err)
		}
	})

	t.Run("Invalid process config - missing command", func(t *testing.T) {
		config := &common.Config{
			Broker: common.BrokerConfig{
				Processes: []common.ProcessConfig{
					{
						ID: "test",
						// Missing Command
						Args: []string{"test"},
					},
				},
			},
		}
		err := broker.LoadProcessesFromConfig(config)
		// Should not error, just skip the invalid config
		if err != nil {
			t.Errorf("Expected no error when skipping invalid config, got: %v", err)
		}
	})

	t.Run("Valid process config - non-RPC without restart", func(t *testing.T) {
		config := &common.Config{
			Broker: common.BrokerConfig{
				Processes: []common.ProcessConfig{
					{
						ID:      "test-echo",
						Command: "echo",
						Args:    []string{"hello"},
						RPC:     false,
						Restart: false,
					},
				},
			},
		}
		err := broker.LoadProcessesFromConfig(config)
		if err != nil {
			t.Errorf("Expected no error for valid config, got: %v", err)
		}

		// Verify process was spawned
		time.Sleep(100 * time.Millisecond)
		info, err := broker.GetProcess("test-echo")
		if err == nil && info != nil {
			// Process was spawned, clean up
			broker.Kill("test-echo")
		}
	})

	t.Run("Valid process config - with restart", func(t *testing.T) {
		config := &common.Config{
			Broker: common.BrokerConfig{
				Processes: []common.ProcessConfig{
					{
						ID:          "test-sleep",
						Command:     "sleep",
						Args:        []string{"10"},
						RPC:         false,
						Restart:     true,
						MaxRestarts: 1,
					},
				},
			},
		}
		err := broker.LoadProcessesFromConfig(config)
		if err != nil {
			t.Errorf("Expected no error for valid config with restart, got: %v", err)
		}

		// Verify process was spawned
		time.Sleep(100 * time.Millisecond)
		info, err := broker.GetProcess("test-sleep")
		if err == nil && info != nil {
			// Process was spawned, clean up
			broker.Kill("test-sleep")
		}
	})

	t.Run("Valid RPC process config", func(t *testing.T) {
		config := &common.Config{
			Broker: common.BrokerConfig{
				Processes: []common.ProcessConfig{
					{
						ID:      "test-rpc-sleep",
						Command: "sleep",
						Args:    []string{"60"},
						RPC:     true,
						Restart: false,
					},
				},
			},
		}
		err := broker.LoadProcessesFromConfig(config)
		if err != nil {
			t.Errorf("Expected no error for valid RPC config, got: %v", err)
		}

		// Verify process was spawned
		time.Sleep(100 * time.Millisecond)
		info, err := broker.GetProcess("test-rpc-sleep")
		if err == nil && info != nil {
			// Process was spawned, clean up
			broker.Kill("test-rpc-sleep")
		}
	})
}
