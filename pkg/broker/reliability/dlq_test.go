package routing

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// Test 1: DLQ - Create New DLQ
func TestDLQ_New(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewDLQ", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_new.db"
		defer os.Remove(dbPath)

		dlq, err := NewDeadLetterQueue(dbPath, 1000)
		if err != nil {
			t.Fatalf("Failed to create DLQ: %v", err)
		}
		defer dlq.Close()

		if dlq == nil {
			t.Fatal("DLQ is nil")
		}

		if dlq.maxSize != 1000 {
			t.Errorf("Max size = %d, want 1000", dlq.maxSize)
		}
	})
}

// Test 2: DLQ - Create with Default Max Size
func TestDLQ_New_DefaultMaxSize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewDLQDefaultSize", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_default.db"
		defer os.Remove(dbPath)

		dlq, err := NewDeadLetterQueue(dbPath, 0) // 0 triggers default
		if err != nil {
			t.Fatalf("Failed to create DLQ: %v", err)
		}
		defer dlq.Close()

		if dlq.maxSize != 10000 {
			t.Errorf("Max size = %d, want 10000 (default)", dlq.maxSize)
		}
	})
}

// Test 3: DLQ - Add Message
func TestDLQ_Add(t *testing.T) {
	testutils.Run(t, testutils.Level1, "AddMessage", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_add.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		msg := &proc.Message{
			ID:     "msg-001",
			Method: "RPCTest",
			Target: "test-service",
		}

		err := dlq.Add(msg, "test failure reason")
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}

		// Verify count
		count, _ := dlq.Count()
		if count != 1 {
			t.Errorf("DLQ count = %d, want 1", count)
		}
	})
}

// Test 4: DLQ - Get Message
func TestDLQ_Get(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetMessage", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_get.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		msg := &proc.Message{
			ID:     "msg-002",
			Method: "RPCTest",
			Target: "test-service",
		}

		dlq.Add(msg, "network timeout")

		// Retrieve message
		retrieved, err := dlq.Get("msg-002")
		if err != nil {
			t.Fatalf("Failed to get message: %v", err)
		}

		if retrieved.MessageID != "msg-002" {
			t.Errorf("Message ID = %s, want msg-002", retrieved.MessageID)
		}

		if retrieved.FailureReason != "network timeout" {
			t.Errorf("Failure reason = %s, want 'network timeout'", retrieved.FailureReason)
		}

		if retrieved.RetryCount != 0 {
			t.Errorf("Retry count = %d, want 0", retrieved.RetryCount)
		}
	})
}

// Test 5: DLQ - Get Non-Existent Message
func TestDLQ_Get_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetMessageNotFound", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_get_notfound.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		_, err := dlq.Get("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent message, got nil")
		}
	})
}

// Test 6: DLQ - List Messages
func TestDLQ_List(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ListMessages", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_list.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		// Add multiple messages
		for i := 1; i <= 5; i++ {
			msg := &proc.Message{
				ID:     fmt.Sprintf("msg-%03d", i),
				Method: "RPCTest",
			}
			dlq.Add(msg, "test failure")
		}

		// List all messages
		messages, err := dlq.List(0, 10)
		if err != nil {
			t.Fatalf("Failed to list messages: %v", err)
		}

		if len(messages) != 5 {
			t.Errorf("Message count = %d, want 5", len(messages))
		}
	})
}

// Test 7: DLQ - List with Pagination
func TestDLQ_List_Pagination(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ListMessagesPagination", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_list_page.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		// Add 10 messages
		for i := 1; i <= 10; i++ {
			msg := &proc.Message{
				ID:     fmt.Sprintf("msg-%03d", i),
				Method: "RPCTest",
			}
			dlq.Add(msg, "test failure")
		}

		// Get page 1 (offset 0, limit 3)
		page1, _ := dlq.List(0, 3)
		if len(page1) != 3 {
			t.Errorf("Page 1 count = %d, want 3", len(page1))
		}

		// Get page 2 (offset 3, limit 3)
		page2, _ := dlq.List(3, 3)
		if len(page2) != 3 {
			t.Errorf("Page 2 count = %d, want 3", len(page2))
		}

		// Get page 3 (offset 6, limit 3)
		page3, _ := dlq.List(6, 3)
		if len(page3) != 3 {
			t.Errorf("Page 3 count = %d, want 3", len(page3))
		}
	})
}

// Test 8: DLQ - Replay Message
func TestDLQ_Replay(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ReplayMessage", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_replay.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		originalMsg := &proc.Message{
			ID:     "msg-replay",
			Method: "RPCTest",
			Target: "test-service",
		}

		dlq.Add(originalMsg, "temporary failure")

		// Replay message
		replayedMsg, err := dlq.Replay("msg-replay")
		if err != nil {
			t.Fatalf("Failed to replay message: %v", err)
		}

		if replayedMsg.ID != "msg-replay" {
			t.Errorf("Replayed message ID = %s, want msg-replay", replayedMsg.ID)
		}

		if replayedMsg.Method != "RPCTest" {
			t.Errorf("Replayed message method = %s, want RPCTest", replayedMsg.Method)
		}

		// Check retry count incremented
		dlqMsg, _ := dlq.Get("msg-replay")
		if dlqMsg.RetryCount != 1 {
			t.Errorf("Retry count = %d, want 1", dlqMsg.RetryCount)
		}
	})
}

// Test 9: DLQ - Replay Updates Last Retry Time
func TestDLQ_Replay_UpdatesRetryTime(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ReplayUpdatesRetryTime", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_replay_time.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		msg := &proc.Message{
			ID:     "msg-retry-time",
			Method: "RPCTest",
		}

		dlq.Add(msg, "failure")
		time.Sleep(10 * time.Millisecond)

		// Replay
		dlq.Replay("msg-retry-time")

		// Check last retry time is set
		dlqMsg, _ := dlq.Get("msg-retry-time")
		if dlqMsg.LastRetryTime.IsZero() {
			t.Error("Last retry time should be set after replay")
		}
	})
}

// Test 10: DLQ - Remove Message
func TestDLQ_Remove(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RemoveMessage", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_remove.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		msg := &proc.Message{
			ID:     "msg-remove",
			Method: "RPCTest",
		}

		dlq.Add(msg, "failure")

		// Remove message
		err := dlq.Remove("msg-remove")
		if err != nil {
			t.Fatalf("Failed to remove message: %v", err)
		}

		// Verify removed
		_, err = dlq.Get("msg-remove")
		if err == nil {
			t.Error("Message should not exist after removal")
		}
	})
}

// Test 11: DLQ - Count Messages
func TestDLQ_Count(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CountMessages", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_count.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		// Add 7 messages
		for i := 1; i <= 7; i++ {
			msg := &proc.Message{
				ID: fmt.Sprintf("msg-%d", i),
			}
			dlq.Add(msg, "failure")
		}

		count, err := dlq.Count()
		if err != nil {
			t.Fatalf("Failed to count messages: %v", err)
		}

		if count != 7 {
			t.Errorf("Count = %d, want 7", count)
		}
	})
}

// Test 12: DLQ - Clear All Messages
func TestDLQ_Clear(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ClearMessages", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_clear.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		// Add messages
		for i := 1; i <= 5; i++ {
			msg := &proc.Message{
				ID: fmt.Sprintf("msg-%d", i),
			}
			dlq.Add(msg, "failure")
		}

		// Clear
		err := dlq.Clear()
		if err != nil {
			t.Fatalf("Failed to clear DLQ: %v", err)
		}

		// Verify empty
		count, _ := dlq.Count()
		if count != 0 {
			t.Errorf("Count after clear = %d, want 0", count)
		}
	})
}

// Test 13: DLQ - Get Stats
func TestDLQ_GetStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetStats", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_stats.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		// Add messages
		msg1 := &proc.Message{ID: "msg-1"}
		msg2 := &proc.Message{ID: "msg-2"}
		msg3 := &proc.Message{ID: "msg-3"}

		dlq.Add(msg1, "failure 1")
		time.Sleep(10 * time.Millisecond)
		dlq.Add(msg2, "failure 2")
		time.Sleep(10 * time.Millisecond)
		dlq.Add(msg3, "failure 3")

		// Replay some messages
		dlq.Replay("msg-1")
		dlq.Replay("msg-1") // Retry again
		dlq.Replay("msg-2")

		stats, err := dlq.GetStats()
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}

		if stats.TotalMessages != 3 {
			t.Errorf("Total messages = %d, want 3", stats.TotalMessages)
		}

		if stats.MaxRetries != 2 {
			t.Errorf("Max retries = %d, want 2", stats.MaxRetries)
		}

		if stats.AverageRetries == 0 {
			t.Error("Average retries should be > 0")
		}

		if stats.OldestMessage.IsZero() {
			t.Error("Oldest message time should be set")
		}

		if stats.NewestMessage.IsZero() {
			t.Error("Newest message time should be set")
		}
	})
}

// Test 14: DLQ - Max Size Enforcement
func TestDLQ_MaxSize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "MaxSizeEnforcement", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_maxsize.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 5) // Max 5 messages
		defer dlq.Close()

		// Add 5 messages (should succeed)
		for i := 1; i <= 5; i++ {
			msg := &proc.Message{
				ID: fmt.Sprintf("msg-%d", i),
			}
			err := dlq.Add(msg, "failure")
			if err != nil {
				t.Fatalf("Failed to add message %d: %v", i, err)
			}
		}

		// Try to add 6th message (should fail)
		msg6 := &proc.Message{ID: "msg-6"}
		err := dlq.Add(msg6, "failure")
		if err == nil {
			t.Error("Expected error when exceeding max size, got nil")
		}
	})
}

// Test 15: DLQ - Stats with Empty Queue
func TestDLQ_GetStats_Empty(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetStatsEmpty", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_dlq_stats_empty.db"
		defer os.Remove(dbPath)

		dlq, _ := NewDeadLetterQueue(dbPath, 1000)
		defer dlq.Close()

		stats, err := dlq.GetStats()
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}

		if stats.TotalMessages != 0 {
			t.Errorf("Total messages = %d, want 0", stats.TotalMessages)
		}

		if stats.MaxRetries != 0 {
			t.Errorf("Max retries = %d, want 0", stats.MaxRetries)
		}

		if stats.AverageRetries != 0 {
			t.Errorf("Average retries = %f, want 0", stats.AverageRetries)
		}
	})
}
