package core

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"runtime"
	"testing"
	"time"
)

func TestBroker_GetProcess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_GetProcess", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestBroker_GetProcess_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_GetProcess_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		_, err := broker.GetProcess("nonexistent")
		if err == nil {
			t.Error("Expected error when getting nonexistent process")
		}
	})

}

func TestBroker_ListProcesses(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_ListProcesses", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestBroker_Kill(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_Kill", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestBroker_Kill_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_Kill_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		err := broker.Kill("nonexistent")
		if err == nil {
			t.Error("Expected error when killing nonexistent process")
		}
	})

}

func TestBroker_Kill_AlreadyExited(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_Kill_AlreadyExited", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestBroker_ProcessReaping(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_ProcessReaping", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestBroker_ProcessReaping_SuccessfulExit(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_ProcessReaping_SuccessfulExit", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

// TestBroker_Integration_MultipleProcesses moved to lifecycle_integration_test.go
