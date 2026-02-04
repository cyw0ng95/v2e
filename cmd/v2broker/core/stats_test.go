package core

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestBroker_GetMessageCount(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_GetMessageCount", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Initially, message count should be 0
		count := broker.GetMessageCount()
		if count != 0 {
			t.Errorf("Expected initial message count to be 0, got %d", count)
		}

		// Send a message
		msg, _ := proc.NewRequestMessage("req-1", nil)
		err := broker.SendMessage(msg)
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		// Count should be 1 (1 sent)
		count = broker.GetMessageCount()
		if count != 1 {
			t.Errorf("Expected message count to be 1, got %d", count)
		}

		// Receive the message
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err = broker.ReceiveMessage(ctx)
		if err != nil {
			t.Fatalf("ReceiveMessage failed: %v", err)
		}

		// Count should be 2 (1 sent + 1 received)
		count = broker.GetMessageCount()
		if count != 2 {
			t.Errorf("Expected message count to be 2, got %d", count)
		}
	})

}

func TestBroker_GetMessageStats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_GetMessageStats", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Initially, all stats should be zero
		stats := broker.GetMessageStats()
		if stats.TotalSent != 0 {
			t.Errorf("Expected TotalSent to be 0, got %d", stats.TotalSent)
		}
		if stats.TotalReceived != 0 {
			t.Errorf("Expected TotalReceived to be 0, got %d", stats.TotalReceived)
		}
		if !stats.FirstMessageTime.IsZero() {
			t.Error("Expected FirstMessageTime to be zero")
		}
		if !stats.LastMessageTime.IsZero() {
			t.Error("Expected LastMessageTime to be zero")
		}

		// Send different types of messages
		reqMsg, _ := proc.NewRequestMessage("req-1", nil)
		respMsg, _ := proc.NewResponseMessage("resp-1", nil)
		eventMsg, _ := proc.NewEventMessage("event-1", nil)
		errorMsg := proc.NewErrorMessage("err-1", fmt.Errorf("test error"))

		reqMsg.Target = "test-target"
		respMsg.Target = "test-target"
		eventMsg.Target = "test-target"
		errorMsg.Target = "test-target"

		broker.SendMessage(reqMsg)
		broker.SendMessage(respMsg)
		broker.SendMessage(eventMsg)
		broker.SendMessage(errorMsg)

		// Check stats after sending
		stats = broker.GetMessageStats()
		if stats.TotalSent != 4 {
			t.Errorf("Expected TotalSent to be 4, got %d", stats.TotalSent)
		}
		if stats.RequestCount != 1 {
			t.Errorf("Expected RequestCount to be 1, got %d", stats.RequestCount)
		}
		if stats.ResponseCount != 1 {
			t.Errorf("Expected ResponseCount to be 1, got %d", stats.ResponseCount)
		}
		if stats.EventCount != 1 {
			t.Errorf("Expected EventCount to be 1, got %d", stats.EventCount)
		}
		if stats.ErrorCount != 1 {
			t.Errorf("Expected ErrorCount to be 1, got %d", stats.ErrorCount)
		}
		if stats.FirstMessageTime.IsZero() {
			t.Error("Expected FirstMessageTime to be set")
		}
		if stats.LastMessageTime.IsZero() {
			t.Error("Expected LastMessageTime to be set")
		}

		// Receive messages
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		for i := 0; i < 4; i++ {
			_, err := broker.ReceiveMessage(ctx)
			if err != nil {
				t.Fatalf("ReceiveMessage %d failed: %v", i, err)
			}
		}

		// Check stats after receiving
		stats = broker.GetMessageStats()
		if stats.TotalSent != 4 {
			t.Errorf("Expected TotalSent to remain 4, got %d", stats.TotalSent)
		}
		if stats.TotalReceived != 4 {
			t.Errorf("Expected TotalReceived to be 4, got %d", stats.TotalReceived)
		}
		// Type counts should have doubled (counted on both send and receive)
		if stats.RequestCount != 2 {
			t.Errorf("Expected RequestCount to be 2, got %d", stats.RequestCount)
		}
		if stats.ResponseCount != 2 {
			t.Errorf("Expected ResponseCount to be 2, got %d", stats.ResponseCount)
		}
		if stats.EventCount != 2 {
			t.Errorf("Expected EventCount to be 2, got %d", stats.EventCount)
		}
		if stats.ErrorCount != 2 {
			t.Errorf("Expected ErrorCount to be 2, got %d", stats.ErrorCount)
		}
	})

}

func TestBroker_MessageStats_Timestamps(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_MessageStats_Timestamps", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Send first message
		msg1, _ := proc.NewRequestMessage("req-1", nil)
		broker.SendMessage(msg1)

		stats := broker.GetMessageStats()
		firstTime := stats.FirstMessageTime
		lastTime := stats.LastMessageTime

		if firstTime.IsZero() {
			t.Error("Expected FirstMessageTime to be set")
		}
		if lastTime.IsZero() {
			t.Error("Expected LastMessageTime to be set")
		}

		// Wait a bit and send another message
		time.Sleep(10 * time.Millisecond)
		msg2, _ := proc.NewRequestMessage("req-2", nil)
		broker.SendMessage(msg2)

		stats = broker.GetMessageStats()

		// FirstMessageTime should not change
		if !stats.FirstMessageTime.Equal(firstTime) {
			t.Error("Expected FirstMessageTime to remain unchanged")
		}

		// LastMessageTime should be updated
		if !stats.LastMessageTime.After(lastTime) {
			t.Error("Expected LastMessageTime to be updated to a later time")
		}
	})

}

func TestBroker_MessageStats_ConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_MessageStats_ConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Test concurrent access to stats
		var wg sync.WaitGroup
		numGoroutines := 10
		messagesPerGoroutine := 10

		// Send messages concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					msg, _ := proc.NewRequestMessage(fmt.Sprintf("req-%d-%d", id, j), nil)
					msg.Target = "test-target"
					broker.SendMessage(msg)
				}
			}(i)
		}

		wg.Wait()

		// Check that all messages were counted
		stats := broker.GetMessageStats()
		expectedTotal := int64(numGoroutines * messagesPerGoroutine)
		if stats.TotalSent != expectedTotal {
			t.Errorf("Expected TotalSent to be %d, got %d", expectedTotal, stats.TotalSent)
		}
		if stats.RequestCount != expectedTotal {
			t.Errorf("Expected RequestCount to be %d, got %d", expectedTotal, stats.RequestCount)
		}
	})

}

func TestBroker_ProcessExitEvent_UpdatesStats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_ProcessExitEvent_UpdatesStats", nil, func(t *testing.T, tx *gorm.DB) {
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

		// Get initial stats
		initialStats := broker.GetMessageStats()

		_, err := broker.Spawn("test-1", cmd, args...)
		if err != nil {
			t.Fatalf("Spawn failed: %v", err)
		}

		// Wait for process to exit and event to be sent
		time.Sleep(500 * time.Millisecond)

		// Check that stats were updated (process exit event should be sent)
		stats := broker.GetMessageStats()
		if stats.TotalSent <= initialStats.TotalSent {
			t.Error("Expected TotalSent to increase after process exit event")
		}
		if stats.EventCount <= initialStats.EventCount {
			t.Error("Expected EventCount to increase after process exit event")
		}
	})

}
