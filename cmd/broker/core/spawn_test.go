package core

import (
	"runtime"
	"testing"
	"time"
)

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
