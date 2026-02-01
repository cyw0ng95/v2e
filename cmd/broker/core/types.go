package core

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/mq"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// ProcessStatus represents the status of a subprocess.
type ProcessStatus string

const (
	// ProcessStatusRunning indicates the process is currently running.
	ProcessStatusRunning ProcessStatus = "running"
	// ProcessStatusExited indicates the process has exited.
	ProcessStatusExited ProcessStatus = "exited"
	// ProcessStatusFailed indicates the process failed to start.
	ProcessStatusFailed ProcessStatus = "failed"
)

// ProcessInfo contains information about a subprocess.
type ProcessInfo struct {
	ID        string
	PID       int
	Command   string
	Args      []string
	Status    ProcessStatus
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
}

// Process represents a managed subprocess.
type Process struct {
	info          *ProcessInfo
	cmd           *exec.Cmd
	cancel        context.CancelFunc
	done          chan struct{}
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	mu            sync.RWMutex
	restartConfig *RestartConfig
	readLoopWg    sync.WaitGroup // Tracks the readProcessMessages goroutine
}

// NewTestProcess constructs a Process instance for tests without spawning OS processes.
func NewTestProcess(id string, status ProcessStatus, stdin io.WriteCloser, stdout io.ReadCloser) *Process {
	return &Process{
		info:   &ProcessInfo{ID: id, Status: status},
		stdin:  stdin,
		stdout: stdout,
		done:   make(chan struct{}),
	}
}

// SetStatus updates the process status in a thread-safe manner.
func (p *Process) SetStatus(status ProcessStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.Status = status
}

// SetStdin sets the stdin writer for the process (used in tests).
func (p *Process) SetStdin(stdin io.WriteCloser) {
	p.stdin = stdin
}

// SetStdout sets the stdout reader for the process (used in tests).
func (p *Process) SetStdout(stdout io.ReadCloser) {
	p.stdout = stdout
}

// Info returns the underlying process info.
func (p *Process) Info() *ProcessInfo {
	return p.info
}

// Done returns the done channel for the process.
func (p *Process) Done() chan struct{} {
	return p.done
}

// RestartConfig holds restart configuration for a process.
type RestartConfig struct {
	Enabled      bool
	MaxRestarts  int
	RestartCount int
	Command      string
	Args         []string
	IsRPC        bool
}

// MessageStats aliases mq.MessageStats for compatibility.
type MessageStats = mq.MessageStats

// PerProcessStats aliases mq.PerProcessStats for compatibility.
type PerProcessStats = mq.PerProcessStats

// PendingRequest represents a pending request awaiting a response.
type PendingRequest struct {
	SourceProcess string
	ResponseChan  chan *proc.Message
	Timestamp     time.Time
}
