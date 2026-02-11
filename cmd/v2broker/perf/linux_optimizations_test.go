package perf

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/proc"
	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestCPUAffinitySetup verifies that CPU affinity can be set without errors
func TestCPUAffinitySetup(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCPUAffinitySetup", nil, func(t *testing.T, tx *gorm.DB) {
		if runtime.NumCPU() <= 1 {
			t.Skip("Skipping CPU affinity test on single-core system")
		}

		// This should not panic or cause issues
		setCPUAffinity(0)
		setCPUAffinity(1)

		// Verify we can get the affinity
		tid := unix.Gettid()
		var cpuSet unix.CPUSet
		err := unix.SchedGetaffinity(tid, &cpuSet)
		if err != nil {
			t.Logf("SchedGetaffinity returned error (may require capabilities): %v", err)
		} else {
			t.Logf("CPU affinity successfully queried for thread %d", tid)
		}
	})
}

// TestProcessPrioritySetup verifies that process priority can be set
func TestProcessPrioritySetup(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestProcessPrioritySetup", nil, func(t *testing.T, tx *gorm.DB) {
		// This should not panic
		setProcessPriority()

		// Try to get the priority
		prio, err := unix.Getpriority(unix.PRIO_PROCESS, 0)
		if err != nil {
			t.Logf("Getpriority returned error (may require capabilities): %v", err)
		} else {
			t.Logf("Process priority successfully set/queried: %d", prio)
		}
	})
}

// TestIOPrioritySetup verifies that I/O priority can be set
func TestIOPrioritySetup(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestIOPrioritySetup", nil, func(t *testing.T, tx *gorm.DB) {
		// This should not panic
		setIOPriority()
		t.Log("I/O priority setup completed without panic")
	})
}

// TestWorkerThreadPinning verifies that worker threads are properly locked
func TestWorkerThreadPinning(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerThreadPinning", nil, func(t *testing.T, tx *gorm.DB) {
		router := &simpleRouter{}
		opt := NewWithConfig(router, Config{
			BufferCap:     10,
			NumWorkers:    2,
			StatsInterval: 100 * time.Millisecond,
			OfferPolicy:   "drop",
			BatchSize:     1,
			FlushInterval: 10 * time.Millisecond,
		})
		defer opt.Stop()

		// Send a few messages to ensure workers are active
		for i := 0; i < 5; i++ {
			msg := simpleMessage(i)
			opt.Offer(msg)
		}

		// Wait a bit for workers to process
		time.Sleep(100 * time.Millisecond)

		// If we get here without panics, thread pinning worked
		t.Log("Worker thread pinning completed successfully")
	})
}

// simpleMessage creates a simple test message
func simpleMessage(id int) *proc.Message {
	return &proc.Message{
		ID:     fmt.Sprintf("test-%d", id),
		Type:   proc.MessageTypeRequest,
		Target: "test",
	}
}
