package transport

import (
	"bytes"
	"sync"
	"testing"

	"github.com/bytedance/sonic"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestHybridTransportPendingMessagesNoLoss(t *testing.T) {
	// Create a hybrid transport with shared memory disabled (will use UDS)
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: false,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	// Verify initial state
	if ht.HasPendingMessages() {
		t.Error("New transport should not have pending messages")
	}

	if ht.PendingMessageCount() != 0 {
		t.Errorf("Pending message count should be 0, got %d", ht.PendingMessageCount())
	}
}

func TestHybridTransportSwitchToUDSWithPendingData(t *testing.T) {
	// Create a hybrid transport with shared memory enabled
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: true,
		SharedMemSize:   64 * 1024,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	// Write some data to shared memory to simulate pending messages
	if ht.sharedMem == nil {
		t.Skip("Shared memory not available on this platform")
	}

	testMessages := [][]byte{
		[]byte(`{"type":"request","id":"test-1"}`),
		[]byte(`{"type":"request","id":"test-2"}`),
		[]byte(`{"type":"request","id":"test-3"}`),
	}

	for _, msg := range testMessages {
		if err := ht.sharedMem.Write(msg); err != nil {
			t.Fatalf("Failed to write to shared memory: %v", err)
		}
	}

	// Verify data is available
	if ht.sharedMem.BytesAvailable() == 0 {
		t.Error("Expected data to be available in shared memory")
	}

	// Calculate total bytes written
	totalBytes := 0
	for _, msg := range testMessages {
		totalBytes += len(msg)
	}

	// Switch to UDS - this should transfer pending data
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("Failed to switch to UDS: %v", err)
	}

	// Verify transport switched
	if ht.GetActiveTransport() != "uds" {
		t.Errorf("Expected active transport to be 'uds', got %s", ht.GetActiveTransport())
	}

	// Verify pending messages were transferred
	if !ht.HasPendingMessages() {
		t.Error("Expected pending messages after switch")
	}

	// Verify that all data was preserved (may be in fewer buffer entries)
	pendingData := collectPendingData(ht)
	if len(pendingData) != totalBytes {
		t.Errorf("Expected %d bytes transferred, got %d", totalBytes, len(pendingData))
	}

	// Verify each message's data is present in the transferred data
	for _, msg := range testMessages {
		if !bytes.Contains(pendingData, msg) {
			t.Errorf("Expected message %s to be present in pending data", string(msg))
		}
	}
}

func TestHybridTransportReceivePendingMessagesFirst(t *testing.T) {
	// Create a hybrid transport with shared memory enabled
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: true,
		SharedMemSize:   64 * 1024,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	if ht.sharedMem == nil {
		t.Skip("Shared memory not available on this platform")
	}

	// Create test messages
	msg1 := &proc.Message{Type: proc.MessageTypeRequest, ID: "pending-msg-1"}
	msg2 := &proc.Message{Type: proc.MessageTypeRequest, ID: "pending-msg-2"}

	data1, err := sonic.Marshal(msg1)
	if err != nil {
		t.Fatalf("Failed to serialize message 1: %v", err)
	}
	data2, err := sonic.Marshal(msg2)
	if err != nil {
		t.Fatalf("Failed to serialize message 2: %v", err)
	}

	// Write messages to shared memory
	if err := ht.sharedMem.Write(data1); err != nil {
		t.Fatalf("Failed to write message 1: %v", err)
	}
	if err := ht.sharedMem.Write(data2); err != nil {
		t.Fatalf("Failed to write message 2: %v", err)
	}

	// Switch to UDS to trigger pending data transfer
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("Failed to switch to UDS: %v", err)
	}

	// Verify pending messages are available
	if ht.PendingMessageCount() == 0 {
		t.Error("Expected pending messages after switch")
	}

	// Verify data integrity
	pendingData := collectPendingData(ht)
	if !bytes.Contains(pendingData, data1) {
		t.Error("Expected msg1 data in pending buffer")
	}
	if !bytes.Contains(pendingData, data2) {
		t.Error("Expected msg2 data in pending buffer")
	}
}

func TestHybridTransportConcurrentPendingMessages(t *testing.T) {
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: true,
		SharedMemSize:   128 * 1024,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	if ht.sharedMem == nil {
		t.Skip("Shared memory not available on this platform")
	}

	const numMessages = 10
	var allData [][]byte
	var dataMu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numMessages)

	// Concurrently write messages
	for i := 0; i < numMessages; i++ {
		go func(id int) {
			defer wg.Done()
			msg := &proc.Message{Type: proc.MessageTypeRequest, ID: string(rune('A' + id))}
			data, _ := sonic.Marshal(msg)
			ht.sharedMem.Write(data)
			dataMu.Lock()
			allData = append(allData, data)
			dataMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Switch to UDS
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("Failed to switch to UDS: %v", err)
	}

	// Verify all messages were transferred (data preservation)
	pendingData := collectPendingData(ht)
	dataMu.Lock()
	defer dataMu.Unlock()
	for _, data := range allData {
		if !bytes.Contains(pendingData, data) {
			t.Errorf("Expected message data %s to be present in pending buffer", string(data))
		}
	}
}

func TestHybridTransportMultipleSwitches(t *testing.T) {
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: true,
		SharedMemSize:   64 * 1024,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	if ht.sharedMem == nil {
		t.Skip("Shared memory not available on this platform")
	}

	// First switch: UDS (no pending data)
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("First switch to UDS failed: %v", err)
	}

	// Switch back to shared memory
	if err := ht.SwitchToSharedMemory(); err != nil {
		t.Fatalf("Switch to shared memory failed: %v", err)
	}

	// Write some data
	msg := &proc.Message{Type: proc.MessageTypeRequest, ID: "test-msg"}
	data, _ := sonic.Marshal(msg)
	ht.sharedMem.Write(data)

	// Second switch: UDS (with pending data)
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("Second switch to UDS failed: %v", err)
	}

	// Verify pending data was transferred
	if !ht.HasPendingMessages() {
		t.Error("Expected pending messages after second switch")
	}

	// Verify data integrity
	pendingData := collectPendingData(ht)
	if !bytes.Contains(pendingData, data) {
		t.Error("Expected message data in pending buffer after second switch")
	}
}

func TestHybridTransportEmptyPendingBuffer(t *testing.T) {
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: true,
		SharedMemSize:   64 * 1024,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	if ht.sharedMem == nil {
		t.Skip("Shared memory not available on this platform")
	}

	// Switch to UDS without any pending data
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("Switch to UDS failed: %v", err)
	}

	// Should not have pending messages
	if ht.HasPendingMessages() {
		t.Error("Should not have pending messages when shared memory was empty")
	}

	if ht.PendingMessageCount() != 0 {
		t.Errorf("Expected 0 pending messages, got %d", ht.PendingMessageCount())
	}
}

func TestHybridTransportFallbackOnReadError(t *testing.T) {
	// Create a hybrid transport with shared memory enabled
	config := HybridTransportConfig{
		SocketPath:      t.TempDir() + "/test.sock",
		UseSharedMemory: true,
		SharedMemSize:   64 * 1024,
		IsServer:        true,
	}

	ht, err := NewHybridTransport(config)
	if err != nil {
		t.Fatalf("Failed to create hybrid transport: %v", err)
	}
	defer ht.Close()

	if ht.sharedMem == nil {
		t.Skip("Shared memory not available on this platform")
	}

	// Write test messages
	msg1 := &proc.Message{Type: proc.MessageTypeRequest, ID: "fallback-test-1"}
	msg2 := &proc.Message{Type: proc.MessageTypeRequest, ID: "fallback-test-2"}

	data1, _ := sonic.Marshal(msg1)
	data2, _ := sonic.Marshal(msg2)

	ht.sharedMem.Write(data1)
	ht.sharedMem.Write(data2)

	// Close shared memory to simulate read error
	ht.sharedMem.Close()

	// The active transport should still be sharedmem initially
	if ht.GetActiveTransport() != "sharedmem" {
		t.Error("Expected initial transport to be sharedmem")
	}

	// SwitchToUDS should handle the closed shared memory gracefully
	// (transferPendingDataLocked will check if shared memory is closed)
	if err := ht.SwitchToUDS(); err != nil {
		t.Fatalf("Switch to UDS should succeed even with closed shared memory: %v", err)
	}

	// Verify transport switched
	if ht.GetActiveTransport() != "uds" {
		t.Errorf("Expected transport to be 'uds', got %s", ht.GetActiveTransport())
	}
}

// collectPendingData collects all pending message data into a single byte slice
func collectPendingData(ht *HybridTransport) []byte {
	ht.pendingMu.Lock()
	defer ht.pendingMu.Unlock()

	var result []byte
	for _, data := range ht.pendingMessages {
		result = append(result, data...)
	}
	return result
}
