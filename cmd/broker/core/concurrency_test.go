package core

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestBroker_ConcurrentSpawn stresses the processes map lock by spawning many short-lived processes concurrently.
func TestBroker_ConcurrentSpawn(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	const workers = 10
	const iterations = 10

	var wg sync.WaitGroup
	wg.Add(workers)

	cmd := "echo"
	args := []string{"test"}
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	}

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Use unique IDs to avoid "already exists" errors since we don't clean up exited processes here
				id := fmt.Sprintf("proc-%d-%d", workerID, i)
				_, err := broker.Spawn(id, cmd, args...)
				if err != nil {
					t.Errorf("Worker %d failed to spawn %s: %v", workerID, id, err)
				}
				// Small random sleep to vary timing
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}(w)
	}

	wg.Wait()

	// Verify count
	count := broker.ProcessCount()
	expected := workers * iterations
	if count != expected {
		t.Errorf("Expected %d processes, got %d", expected, count)
	}
}

// TestBroker_ConcurrentListAndSpawn stresses the interaction between ListProcesses (RLock) and Spawn (Lock).
func TestBroker_ConcurrentListAndSpawn(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// Starter goroutines
	const spawners = 5
	wg.Add(spawners)

	cmd := "echo"
	args := []string{"test"}
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	}

	for w := 0; w < spawners; w++ {
		go func(workerID int) {
			defer wg.Done()
			i := 0
			for {
				select {
				case <-stopCh:
					return
				default:
					id := fmt.Sprintf("proc-list-%d-%d", workerID, i)
					broker.Spawn(id, cmd, args...)
					i++
					time.Sleep(time.Millisecond * 5)
				}
			}
		}(w)
	}

	// Lister goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopCh:
				return
			default:
				procs := broker.ListProcesses()
				// access the info to ensure no data races on the pointers
				for _, p := range procs {
					_ = p.ID
					_ = p.Status
				}
				time.Sleep(time.Millisecond * 2)
			}
		}
	}()

	// Run for a bit
	time.Sleep(200 * time.Millisecond)
	close(stopCh)
	wg.Wait()
}

// TestBroker_ConcurrentRestart stresses the reapProcess logic which deletes from the map and respawns.
func TestBroker_ConcurrentRestart(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	const numProcs = 2
	const maxRestarts = 1

	var wg sync.WaitGroup
	wg.Add(numProcs)

	cmd := "echo"
	args := []string{"test"}
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	}

	for i := 0; i < numProcs; i++ {
		go func(idStr string) {
			defer wg.Done()
			// Spawn with restart. The process (echo) will exit immediately, triggering restart.
			// This causes rapid delete/insert in the map.
			_, err := broker.SpawnWithRestart(idStr, cmd, maxRestarts, args...)
			if err != nil {
				t.Errorf("Failed to spawn %s: %v", idStr, err)
			}
		}(fmt.Sprintf("restart-proc-%d", i))
	}

	// Wait enough time for restarts to happen
	// Each restart has 1s sleep in reapProcess (hardcoded in process_lifecycle.go)
	// So 3 restarts = ~3s.
	time.Sleep(4 * time.Second)
	wg.Wait()

	// Check final states
	procs := broker.ListProcesses()
	for _, p := range procs {
		if p.Status != ProcessStatusExited {
			t.Logf("Process %s status: %s (might still be running if timing was off)", p.ID, p.Status)
		}
	}
}

// TestBroker_ConcurrentKillAndSpawn stresses locking between Kill (RLock then Lock on process) and Spawn.
func TestBroker_ConcurrentKillAndSpawn(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	const workers = 5
	const iterations = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	// Use sleep command so we have something to kill
	cmd := "sleep"
	args := []string{"1"}
	if runtime.GOOS == "windows" {
		cmd = "powershell"
		args = []string{"-Command", "Start-Sleep -Seconds 1"}
	}

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				id := fmt.Sprintf("kill-proc-%d-%d", workerID, i)
				
				// Spawn
				_, err := broker.Spawn(id, cmd, args...)
				if err != nil {
					t.Errorf("Spawn failed: %v", err)
					continue
				}

				// Immediately try to kill it
				// This races with the process starting and potentially exiting (though sleep 1s shouldn't exit fast)
				go func(pid string) {
					err := broker.Kill(pid)
					if err != nil {
						// It's acceptable to fail if it already exited or wasn't found yet (though Spawn returned)
						// but checking for data races is the main goal.
					}
				}(id)

				time.Sleep(10 * time.Millisecond)
			}
		}(w)
	}

	wg.Wait()
}
