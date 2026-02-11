package core

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// Re-declaring tests that need fmt and strings

func TestBroker_Integration_MultipleProcesses(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_Integration_MultipleProcesses", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}
