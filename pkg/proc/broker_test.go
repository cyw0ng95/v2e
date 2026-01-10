package proc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

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

	var buf bytes.Buffer
	logger := common.NewLogger(&buf, "", common.DebugLevel)

	broker.SetLogger(logger)

	// Spawn a process to generate logs
	_, _ = broker.Spawn("test", "echo", "hello")
	time.Sleep(100 * time.Millisecond)

	if buf.Len() == 0 {
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

	msg, err := NewRequestMessage("req-1", map[string]string{"test": "data"})
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

	if msg.Type != MessageTypeEvent {
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
	msg, _ := NewRequestMessage("req-1", nil)
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
	var buf bytes.Buffer
	logger := common.NewLogger(&buf, "", common.DebugLevel)
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
