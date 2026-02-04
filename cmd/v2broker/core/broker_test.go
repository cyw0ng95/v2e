package core

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
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
	testutils.Run(t, testutils.Level1, "TestNewBroker", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		if broker == nil {
			t.Fatal("NewBroker returned nil")
		}

		if broker.ProcessCount() != 0 {
			t.Error("Expected no processes on new broker")
		}
		if broker.MessageChannel() == nil {
			t.Error("Expected messages channel to be initialized")
		}

		// Clean up
		_ = broker.Shutdown()
	})

}

func TestBroker_SetLogger(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_SetLogger", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		buf := &threadSafeBuffer{}
		logger := common.NewLogger(buf, "", common.DebugLevel)

		broker.SetLogger(logger)

		// Spawn a process to generate logs
		_, _ = broker.Spawn("test", "echo", "hello")
	
		// Poll for process to start and generate logs
		deadline := time.Now().Add(500 * time.Millisecond)
		for time.Now().Before(deadline) {
			if len(buf.String()) > 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		output := buf.String()
		if len(output) == 0 {
			t.Error("Expected logger to capture output")
		}
	})

}

func TestBroker_Shutdown(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_Shutdown", nil, func(t *testing.T, tx *gorm.DB) {
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

		// Poll for process to start instead of fixed sleep
		deadline := time.Now().Add(500 * time.Millisecond)
		var processStarted bool
		for time.Now().Before(deadline) {
			info, err := broker.GetProcess("test-1")
			if err == nil && info.Status == ProcessStatusRunning {
				processStarted = true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	
		if !processStarted {
			t.Fatal("Process did not start within timeout")
		}

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
	})

}

func TestProcessStatus_Constants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestProcessStatus_Constants", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}
