package core

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestBroker_SendReceiveMessage(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_SendReceiveMessage", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		msg, err := proc.NewRequestMessage("req-1", map[string]string{"test": "data"})
		if err != nil {
			t.Fatalf("NewRequestMessage failed: %v", err)
		}

		err = broker.SendMessage(msg)
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		received, err := broker.ReceiveMessage(ctx)
		if err != nil {
			t.Fatalf("ReceiveMessage failed: %v", err)
		}

		if received.ID != msg.ID {
			t.Errorf("Expected message ID to be '%s', got '%s'", msg.ID, received.ID)
		}
	})

}

func TestBroker_ReceiveMessage_Timeout(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_ReceiveMessage_Timeout", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := broker.ReceiveMessage(ctx)
		if err == nil {
			t.Error("Expected timeout error when receiving message")
		}
	})

}

func TestBroker_ProcessExitEvent(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_ProcessExitEvent", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		var cmd string
		var args []string
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "echo", "test"}
		} else {
			cmd = "echo"
			args = []string{"test"}
		}

		_, err := broker.Spawn("test-1", cmd, args...)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}

		// Wait for process to exit
		time.Sleep(500 * time.Millisecond)

		// Check for exit event message
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		msg, err := broker.ReceiveMessage(ctx)
		if err != nil {
			t.Fatalf("ReceiveMessage failed: %v", err)
		}

		if msg.Type != proc.MessageTypeEvent {
			t.Errorf("Expected MessageTypeEvent, got %s", msg.Type)
		}

		var payload map[string]interface{}
		if err := msg.UnmarshalPayload(&payload); err != nil {
			t.Fatalf("UnmarshalPayload failed: %v", err)
		}

		if payload["event"] != "process_exited" {
			t.Errorf("Expected event to be 'process_exited', got %v", payload["event"])
		}
	})

}

func TestBroker_Shutdown_MessageChannel(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_Shutdown_MessageChannel", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()

		err := broker.Shutdown()
		if err != nil {
			t.Errorf("Shutdown failed: %v", err)
		}

		// Try to send a message after shutdown
		msg, _ := proc.NewRequestMessage("req-1", nil)
		err = broker.SendMessage(msg)
		if err == nil {
			t.Error("Expected error when sending message after shutdown")
		}
	})

}
