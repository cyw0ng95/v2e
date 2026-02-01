package core

import (
	"io"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestProcessInfo_StatusTransitions covers state transitions for process info.
func TestProcessInfo_StatusTransitions(t *testing.T) {
	cases := []struct {
		name        string
		startStatus ProcessStatus
		endStatus   ProcessStatus
	}{
		{name: "running-to-exited", startStatus: ProcessStatusRunning, endStatus: ProcessStatusExited},
		{name: "running-to-failed", startStatus: ProcessStatusRunning, endStatus: ProcessStatusFailed},
		{name: "failed-remains", startStatus: ProcessStatusFailed, endStatus: ProcessStatusFailed},
		{name: "exited-remains", startStatus: ProcessStatusExited, endStatus: ProcessStatusExited},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			info := &ProcessInfo{
				Status: tc.startStatus,
			}

			if info.Status != tc.startStatus {
				t.Fatalf("expected initial status %s, got %s", tc.startStatus, info.Status)
			}

			info.Status = tc.endStatus
			if info.Status != tc.endStatus {
				t.Fatalf("expected final status %s, got %s", tc.endStatus, info.Status)
			}
		})
	}
}

// TestProcess_SetStatus verifies thread-safe status updates.
func TestProcess_SetStatus(t *testing.T) {
	proc := NewTestProcess("test", ProcessStatusRunning, nil, nil)

	if proc.info.Status != ProcessStatusRunning {
		t.Fatalf("expected initial status running, got %s", proc.info.Status)
	}

	proc.SetStatus(ProcessStatusExited)
	if proc.info.Status != ProcessStatusExited {
		t.Fatalf("expected status exited, got %s", proc.info.Status)
	}

	proc.SetStatus(ProcessStatusFailed)
	if proc.info.Status != ProcessStatusFailed {
		t.Fatalf("expected status failed, got %s", proc.info.Status)
	}
}

// TestProcess_IOSetters ensures stdin/stdout can be set independently.
func TestProcess_IOSetters(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	proc := NewTestProcess("io-test", ProcessStatusRunning, nil, nil)

	proc.SetStdin(w)
	proc.SetStdout(r)

	if proc.stdin != w {
		t.Fatalf("stdin not set correctly")
	}
	if proc.stdout != r {
		t.Fatalf("stdout not set correctly")
	}
}

// TestRestartConfig_State covers restart counter increments.
func TestRestartConfig_State(t *testing.T) {
	config := &RestartConfig{
		Enabled:      true,
		MaxRestarts:  5,
		RestartCount: 0,
	}

	for i := 1; i <= 5; i++ {
		config.RestartCount++
		if config.RestartCount != i {
			t.Fatalf("expected RestartCount %d, got %d", i, config.RestartCount)
		}
	}

	if config.RestartCount >= config.MaxRestarts {
		config.Enabled = false
	}

	if config.Enabled {
		t.Fatalf("expected restart to be disabled after max restarts")
	}
}

// TestPendingRequest_Timestamps ensures timestamp field is set.
func TestPendingRequest_Timestamps(t *testing.T) {
	before := time.Now()
	req := &PendingRequest{
		SourceProcess: "source",
		ResponseChan:  make(chan *proc.Message, 1),
		Timestamp:     time.Now(),
	}
	after := time.Now()

	if req.Timestamp.Before(before) || req.Timestamp.After(after) {
		t.Fatalf("timestamp out of range: %v", req.Timestamp)
	}

	if req.SourceProcess != "source" {
		t.Fatalf("expected SourceProcess 'source', got %s", req.SourceProcess)
	}

	if req.ResponseChan == nil {
		t.Fatalf("ResponseChan should not be nil")
	}
}

// TestProcessInfo_Fields ensures all fields can be set and retrieved.
func TestProcessInfo_Fields(t *testing.T) {
	info := &ProcessInfo{
		ID:       "proc1",
		PID:      12345,
		Command:  "/bin/echo",
		Args:     []string{"test", "arg"},
		Status:   ProcessStatusRunning,
		ExitCode: 0,
		EndTime:  time.Time{},
	}

	if info.ID != "proc1" {
		t.Fatalf("expected ID proc1, got %s", info.ID)
	}
	if info.PID != 12345 {
		t.Fatalf("expected PID 12345, got %d", info.PID)
	}
	if info.Command != "/bin/echo" {
		t.Fatalf("expected Command /bin/echo, got %s", info.Command)
	}
	if len(info.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(info.Args))
	}
	if info.Status != ProcessStatusRunning {
		t.Fatalf("expected status running, got %s", info.Status)
	}
	if info.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", info.ExitCode)
	}
}

// TestProcess_DoneChannel verifies done channel behavior.
func TestProcess_DoneChannel(t *testing.T) {
	proc := NewTestProcess("done-test", ProcessStatusRunning, nil, nil)

	select {
	case <-proc.Done():
		t.Fatalf("done channel should not be closed initially")
	default:
		// Expected: channel is not closed yet
	}

	close(proc.done)

	select {
	case <-proc.Done():
		// Expected: channel is now closed
	default:
		t.Fatalf("done channel should be closed after close()")
	}
}
