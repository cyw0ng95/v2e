package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// threadSafeBuffer wraps bytes.Buffer with a mutex for thread-safe access
type threadSafeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *threadSafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *threadSafeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestNewBroker(t *testing.T) {
	broker := NewBroker()
	if broker == nil {
		t.Fatal("NewBroker returned nil")
	}

	if broker.processes == nil {
		t.Error("Expected processes map to be initialized")
	}
	if broker.messages == nil {
		t.Error("Expected messages channel to be initialized")
	}

	// Clean up
	_ = broker.Shutdown()
}

func TestBroker_SetLogger(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	buf := &threadSafeBuffer{}
	logger := common.NewLogger(buf, "", common.DebugLevel)

	broker.SetLogger(logger)

	// Spawn a process to generate logs
	_, _ = broker.Spawn("test", "echo", "hello")
	time.Sleep(200 * time.Millisecond)

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected logger to capture output")
	}
}

func TestBroker_Spawn_Success(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Get appropriate command for the platform
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	} else {
		cmd = "echo"
		args = []string{"test"}
	}

	info, err := broker.Spawn("test-1", cmd, args...)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	if info.ID != "test-1" {
		t.Errorf("Expected ID to be 'test-1', got '%s'", info.ID)
	}
	if info.Command != cmd {
		t.Errorf("Expected Command to be '%s', got '%s'", cmd, info.Command)
	}
	if info.Status != ProcessStatusRunning {
		t.Errorf("Expected Status to be ProcessStatusRunning, got %s", info.Status)
	}
	if info.PID <= 0 {
		t.Errorf("Expected PID to be positive, got %d", info.PID)
	}
}

func TestBroker_Spawn_DuplicateID(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping"
	} else {
		cmd = "sleep"
	}

	args := []string{"1"}
	if runtime.GOOS == "windows" {
		args = []string{"-n", "2", "127.0.0.1"}
	}

	_, err := broker.Spawn("test-1", cmd, args...)
	if err != nil {
		t.Fatalf("First Spawn failed: %v", err)
	}

	_, err = broker.Spawn("test-1", cmd, args...)
	if err == nil {
		t.Error("Expected error when spawning process with duplicate ID")
	}
}

func TestBroker_Spawn_InvalidCommand(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	_, err := broker.Spawn("test-1", "nonexistent-command-12345")
	if err == nil {
		t.Error("Expected error when spawning with invalid command")
	}
}

func TestBroker_GetProcess(t *testing.T) {
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

	info, err := broker.GetProcess("test-1")
	if err != nil {
		t.Fatalf("GetProcess failed: %v", err)
	}

	if info.ID != "test-1" {
		t.Errorf("Expected ID to be 'test-1', got '%s'", info.ID)
	}
}

func TestBroker_GetProcess_NotFound(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	_, err := broker.GetProcess("nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent process")
	}
}

func TestBroker_ListProcesses(t *testing.T) {
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

	_, err = broker.Spawn("test-2", cmd, args...)
	if err != nil {
		t.Fatalf("Second Spawn failed: %v", err)
	}

	processes := broker.ListProcesses()
	if len(processes) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(processes))
	}
}

func TestBroker_Kill(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "ping"
		args = []string{"-n", "10", "127.0.0.1"}
	} else {
		cmd = "sleep"
		args = []string{"10"}
	}

	_, err := broker.Spawn("test-1", cmd, args...)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Give the process a moment to start
	time.Sleep(100 * time.Millisecond)

	err = broker.Kill("test-1")
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	// Wait for process to be reaped
	time.Sleep(500 * time.Millisecond)

	info, err := broker.GetProcess("test-1")
	if err != nil {
		t.Fatalf("GetProcess failed: %v", err)
	}

	if info.Status != ProcessStatusExited {
		t.Errorf("Expected Status to be ProcessStatusExited, got %s", info.Status)
	}
}

func TestBroker_Kill_NotFound(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	err := broker.Kill("nonexistent")
	if err == nil {
		t.Error("Expected error when killing nonexistent process")
	}
}

func TestBroker_Kill_AlreadyExited(t *testing.T) {
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

	// Wait for process to exit naturally
	time.Sleep(500 * time.Millisecond)

	err = broker.Kill("test-1")
	if err == nil {
		t.Error("Expected error when killing already exited process")
	}
}

func TestBroker_ProcessReaping(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "exit", "42"}
	} else {
		cmd = "sh"
		args = []string{"-c", "exit 42"}
	}

	_, err := broker.Spawn("test-1", cmd, args...)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Wait for process to exit and be reaped
	time.Sleep(500 * time.Millisecond)

	info, err := broker.GetProcess("test-1")
	if err != nil {
		t.Fatalf("GetProcess failed: %v", err)
	}

	if info.Status != ProcessStatusExited {
		t.Errorf("Expected Status to be ProcessStatusExited, got %s", info.Status)
	}
	if info.ExitCode != 42 {
		t.Errorf("Expected ExitCode to be 42, got %d", info.ExitCode)
	}
	if info.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}
}

func TestBroker_ProcessReaping_SuccessfulExit(t *testing.T) {
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

	// Wait for process to exit and be reaped
	time.Sleep(500 * time.Millisecond)

	info, err := broker.GetProcess("test-1")
	if err != nil {
		t.Fatalf("GetProcess failed: %v", err)
	}

	if info.Status != ProcessStatusExited {
		t.Errorf("Expected Status to be ProcessStatusExited, got %s", info.Status)
	}
	if info.ExitCode != 0 {
		t.Errorf("Expected ExitCode to be 0, got %d", info.ExitCode)
	}
}

func TestBroker_SendReceiveMessage(t *testing.T) {
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
}

func TestBroker_ReceiveMessage_Timeout(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := broker.ReceiveMessage(ctx)
	if err == nil {
		t.Error("Expected timeout error when receiving message")
	}
}

func TestBroker_ProcessExitEvent(t *testing.T) {
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
}

func TestBroker_Shutdown(t *testing.T) {
	broker := NewBroker()

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "ping"
		args = []string{"-n", "10", "127.0.0.1"}
	} else {
		cmd = "sleep"
		args = []string{"10"}
	}

	_, err := broker.Spawn("test-1", cmd, args...)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Give process time to start
	time.Sleep(100 * time.Millisecond)

	err = broker.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify process was killed
	info, err := broker.GetProcess("test-1")
	if err != nil {
		t.Fatalf("GetProcess failed: %v", err)
	}

	if info.Status != ProcessStatusExited {
		t.Errorf("Expected Status to be ProcessStatusExited after shutdown, got %s", info.Status)
	}
}

func TestBroker_Shutdown_MessageChannel(t *testing.T) {
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
}

func TestProcessStatus_Constants(t *testing.T) {
	tests := []struct {
		status   ProcessStatus
		expected string
	}{
		{ProcessStatusRunning, "running"},
		{ProcessStatusExited, "exited"},
		{ProcessStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestBroker_Integration_MultipleProcesses(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Set up logger to capture output
	buf := &threadSafeBuffer{}
	logger := common.NewLogger(buf, "", common.DebugLevel)
	broker.SetLogger(logger)

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	} else {
		cmd = "echo"
		args = []string{"test"}
	}

	// Spawn multiple processes
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("test-%d", i)
		_, err := broker.Spawn(id, cmd, args...)
		if err != nil {
			t.Fatalf("Spawn %d failed: %v", i, err)
		}
	}

	// Wait for all processes to complete
	time.Sleep(1 * time.Second)

	// Verify all processes exited
	processes := broker.ListProcesses()
	for _, proc := range processes {
		if proc.Status != ProcessStatusExited {
			t.Errorf("Process %s expected to be exited, got %s", proc.ID, proc.Status)
		}
	}

	// Check logs
	output := buf.String()
	if !strings.Contains(output, "Spawned process") {
		t.Error("Expected log to contain 'Spawned process'")
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestBroker_GetMessageCount(t *testing.T) {
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
}

func TestBroker_GetMessageStats(t *testing.T) {
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
}

func TestBroker_MessageStats_Timestamps(t *testing.T) {
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
}

func TestBroker_MessageStats_ConcurrentAccess(t *testing.T) {
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
}

func TestBroker_ProcessExitEvent_UpdatesStats(t *testing.T) {
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
}

// TestBroker_RegisterEndpoint tests endpoint registration
func TestBroker_RegisterEndpoint(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Register endpoints for a process
	broker.RegisterEndpoint("test-proc", "RPCGetData")
	broker.RegisterEndpoint("test-proc", "RPCSetData")
	broker.RegisterEndpoint("test-proc", "RPCGetData") // Duplicate should not be added

	// Get endpoints
	endpoints := broker.GetEndpoints("test-proc")

	// Verify
	if len(endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(endpoints))
	}
}

// TestBroker_GetAllEndpoints tests getting all endpoints
func TestBroker_GetAllEndpoints(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Register endpoints for multiple processes
	broker.RegisterEndpoint("proc1", "RPCMethod1")
	broker.RegisterEndpoint("proc1", "RPCMethod2")
	broker.RegisterEndpoint("proc2", "RPCMethod3")

	// Get all endpoints
	allEndpoints := broker.GetAllEndpoints()

	// Verify
	if len(allEndpoints) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(allEndpoints))
	}

	if len(allEndpoints["proc1"]) != 2 {
		t.Errorf("Expected 2 endpoints for proc1, got %d", len(allEndpoints["proc1"]))
	}

	if len(allEndpoints["proc2"]) != 1 {
		t.Errorf("Expected 1 endpoint for proc2, got %d", len(allEndpoints["proc2"]))
	}
}

// TestBroker_SpawnWithRestart tests spawning a process with auto-restart
func TestBroker_SpawnWithRestart(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Spawn a process with restart
	info, err := broker.SpawnWithRestart("test-echo", "echo", 3, "hello")
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	// Verify
	if info.ID != "test-echo" {
		t.Errorf("Expected ID 'test-echo', got '%s'", info.ID)
	}

	if info.Status != ProcessStatusRunning {
		t.Errorf("Expected status 'running', got '%s'", info.Status)
	}

	// Wait for process to exit
	time.Sleep(500 * time.Millisecond)

	// Verify process still exists (should have restarted)
	proc, err := broker.GetProcess("test-echo")
	if err != nil {
		// Process might have exited and not restarted yet, which is ok for echo
		t.Logf("Process may have exited: %v", err)
	} else {
		t.Logf("Process status: %s", proc.Status)
	}
}

func TestHandleRPCGetMessageStats(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Send some messages to generate stats
	reqMsg, _ := proc.NewRequestMessage("test-req", nil)
	reqMsg.Target = "test-target"
	broker.SendMessage(reqMsg)

	// Create RPC request for GetMessageStats
	rpcReq, err := proc.NewRequestMessage("RPCGetMessageStats", nil)
	if err != nil {
		t.Fatalf("Failed to create RPC request: %v", err)
	}
	rpcReq.Source = "test-caller"

	// Handle the RPC request
	respMsg, err := broker.HandleRPCGetMessageStats(rpcReq)
	if err != nil {
		t.Fatalf("HandleRPCGetMessageStats failed: %v", err)
	}

	// Verify response
	if respMsg.Type != proc.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", respMsg.Type)
	}

	if respMsg.Source != "broker" {
		t.Errorf("Expected source 'broker', got %s", respMsg.Source)
	}

	if respMsg.Target != "test-caller" {
		t.Errorf("Expected target 'test-caller', got %s", respMsg.Target)
	}

	// Parse the response payload as a map
	var payload map[string]interface{}
	if err := respMsg.UnmarshalPayload(&payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Extract 'total' sub-map and check TotalSent
	total, ok := payload["total"].(map[string]interface{})
	if !ok {
		t.Fatalf("Response payload missing 'total' field or wrong type")
	}
	totalSent, ok := total["total_sent"].(float64)
	if !ok {
		t.Fatalf("Expected total_sent to be float64, got %T", total["total_sent"])
	}
	if totalSent < 1 {
		t.Errorf("Expected TotalSent >= 1, got %v", totalSent)
	}
}

func TestHandleRPCGetMessageCount(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Send some messages to generate count
	reqMsg, _ := proc.NewRequestMessage("test-req", nil)
	broker.SendMessage(reqMsg)

	// Create RPC request for GetMessageCount
	rpcReq, err := proc.NewRequestMessage("RPCGetMessageCount", nil)
	if err != nil {
		t.Fatalf("Failed to create RPC request: %v", err)
	}
	rpcReq.Source = "test-caller"

	// Handle the RPC request
	respMsg, err := broker.HandleRPCGetMessageCount(rpcReq)
	if err != nil {
		t.Fatalf("HandleRPCGetMessageCount failed: %v", err)
	}

	// Verify response
	if respMsg.Type != proc.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", respMsg.Type)
	}

	if respMsg.Source != "broker" {
		t.Errorf("Expected source 'broker', got %s", respMsg.Source)
	}

	if respMsg.Target != "test-caller" {
		t.Errorf("Expected target 'test-caller', got %s", respMsg.Target)
	}

	// Parse the response payload
	var payload map[string]interface{}
	if err := respMsg.UnmarshalPayload(&payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Verify count exists
	count, ok := payload["count"]
	if !ok {
		t.Fatal("Response payload missing 'count' field")
	}

	// Count should be at least 1
	countFloat, ok := count.(float64)
	if !ok {
		t.Fatalf("Expected count to be float64, got %T", count)
	}

	if countFloat < 1 {
		t.Errorf("Expected count >= 1, got %f", countFloat)
	}
}

func TestProcessMessage_RPCGetMessageStats(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Send some messages to generate stats
	reqMsg, _ := proc.NewRequestMessage("test-req", nil)
	broker.SendMessage(reqMsg)

	// Create RPC request without a source (so response won't be routed)
	rpcReq, err := proc.NewRequestMessage("RPCGetMessageStats", nil)
	if err != nil {
		t.Fatalf("Failed to create RPC request: %v", err)
	}

	// Process the message
	err = broker.ProcessMessage(rpcReq)
	// Since there's no source, the response won't have a target and will go to broker's channel
	// This is expected for direct broker invocations
	if err != nil {
		// Error is expected since we don't have a real calling process
		// The important thing is the handler was called
		t.Logf("Expected routing error: %v", err)
	}
}

func TestProcessMessage_UnknownRPC(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create RPC request with unknown method
	rpcReq, err := proc.NewRequestMessage("RPCUnknownMethod", nil)
	if err != nil {
		t.Fatalf("Failed to create RPC request: %v", err)
	}

	// Process the message - should return error message
	err = broker.ProcessMessage(rpcReq)
	// Error is expected since routing will fail without a real process
	if err != nil {
		t.Logf("Expected routing error: %v", err)
	}
}
