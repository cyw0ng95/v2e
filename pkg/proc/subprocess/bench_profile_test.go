package subprocess

import (
	"io"
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

// TestProfileCPUAndHeap runs a workload and writes CPU and heap profiles to .build
// This is intended to be executed manually when profiling improvements.
func TestProfileCPUAndHeap(t *testing.T) {
	// Skip in short mode to avoid running in normal CI
	if testing.Short() {
		t.Skip("skipping profiling test in short mode")
	}

	if err := os.MkdirAll(".build", 0755); err != nil {
		t.Fatalf("failed to create .build dir: %v", err)
	}

	cpuFile, err := os.Create(".build/cpu.prof")
	if err != nil {
		t.Fatalf("failed to create cpu profile: %v", err)
	}
	defer cpuFile.Close()

	// Set up subprocess and writer
	sp := New("profile-test")
	sp.SetOutput(io.Discard)
	// Start writer goroutine
	sp.wg.Add(1)
	go sp.messageWriter()

	// Start CPU profiling
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		t.Fatalf("failed to start cpu profile: %v", err)
	}

	// Run workload for a fixed duration
	dur := 3 * time.Second
	end := time.Now().Add(dur)
	msg := &Message{Type: MessageTypeEvent, ID: "bench", Payload: nil}
	for time.Now().Before(end) {
		// send many messages
		for i := 0; i < 1000; i++ {
			_ = sp.sendMessage(msg)
		}
		// small sleep to let writer flush
		time.Sleep(10 * time.Millisecond)
	}

	pprof.StopCPUProfile()

	// Write heap profile
	heapFile, err := os.Create(".build/heap.prof")
	if err != nil {
		t.Fatalf("failed to create heap profile: %v", err)
	}
	defer heapFile.Close()
	if err := pprof.WriteHeapProfile(heapFile); err != nil {
		t.Fatalf("failed to write heap profile: %v", err)
	}

	// Stop subprocess
	_ = sp.Stop()
}
