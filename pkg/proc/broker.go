package proc

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// ProcessStatus represents the status of a subprocess
type ProcessStatus string

const (
	// ProcessStatusRunning indicates the process is currently running
	ProcessStatusRunning ProcessStatus = "running"
	// ProcessStatusExited indicates the process has exited
	ProcessStatusExited ProcessStatus = "exited"
	// ProcessStatusFailed indicates the process failed to start
	ProcessStatusFailed ProcessStatus = "failed"
)

// ProcessInfo contains information about a subprocess
type ProcessInfo struct {
	// ID is a unique identifier for the process
	ID string
	// PID is the process ID
	PID int
	// Command is the command that was executed
	Command string
	// Args are the arguments passed to the command
	Args []string
	// Status is the current status of the process
	Status ProcessStatus
	// ExitCode is the exit code of the process (if exited)
	ExitCode int
	// StartTime is when the process was started
	StartTime time.Time
	// EndTime is when the process ended (if exited)
	EndTime time.Time
}

// Process represents a managed subprocess
type Process struct {
	info   *ProcessInfo
	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.RWMutex
}

// Broker manages subprocesses and message passing
type Broker struct {
	processes        map[string]*Process
	managedProcesses map[string]ManagedProcess
	messages         chan *Message
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	logger           *common.Logger
}

// NewBroker creates a new Broker instance
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broker{
		processes:        make(map[string]*Process),
		managedProcesses: make(map[string]ManagedProcess),
		messages:         make(chan *Message, 100),
		ctx:              ctx,
		cancel:           cancel,
		logger:           common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
	}
}

// SetLogger sets the logger for the broker
func (b *Broker) SetLogger(logger *common.Logger) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger = logger
}

// Spawn starts a new subprocess with the given command and arguments
// It returns the process ID and an error if the process failed to start
func (b *Broker) Spawn(id, command string, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if process with this ID already exists
	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	// Create process context
	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	// Create process info
	info := &ProcessInfo{
		ID:        id,
		Command:   command,
		Args:      args,
		Status:    ProcessStatusRunning,
		StartTime: time.Now(),
	}

	proc := &Process{
		info:   info,
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned process: id=%s pid=%d command=%s", id, info.PID, command)

	// Start goroutine to wait for process completion
	b.wg.Add(1)
	go b.reapProcess(proc)

	return info, nil
}

// reapProcess waits for a process to complete and updates its status
func (b *Broker) reapProcess(proc *Process) {
	defer b.wg.Done()
	defer close(proc.done)

	// Wait for the process to complete
	err := proc.cmd.Wait()

	proc.mu.Lock()
	defer proc.mu.Unlock()

	proc.info.EndTime = time.Now()
	proc.info.Status = ProcessStatusExited

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				proc.info.ExitCode = status.ExitStatus()
			} else {
				proc.info.ExitCode = -1
			}
		} else {
			proc.info.ExitCode = -1
		}
	} else {
		proc.info.ExitCode = 0
	}

	b.logger.Info("Process exited: id=%s pid=%d exit_code=%d",
		proc.info.ID, proc.info.PID, proc.info.ExitCode)

	// Send event message about process exit
	event, _ := NewEventMessage(proc.info.ID, map[string]interface{}{
		"event":     "process_exited",
		"id":        proc.info.ID,
		"pid":       proc.info.PID,
		"exit_code": proc.info.ExitCode,
	})
	select {
	case b.messages <- event:
	case <-b.ctx.Done():
	}
}

// Kill terminates a process by ID
func (b *Broker) Kill(id string) error {
	b.mu.RLock()
	proc, exists := b.processes[id]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("process with id '%s' not found", id)
	}

	proc.mu.RLock()
	status := proc.info.Status
	proc.mu.RUnlock()

	if status != ProcessStatusRunning {
		return fmt.Errorf("process '%s' is not running", id)
	}

	// Cancel the context to stop the process
	proc.cancel()

	// Wait for process to exit with timeout
	select {
	case <-proc.done:
		b.logger.Info("Process killed: id=%s", id)
		return nil
	case <-time.After(5 * time.Second):
		// Force kill if graceful shutdown failed
		if proc.cmd.Process != nil {
			if err := proc.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to force kill process: %w", err)
			}
		}
		<-proc.done
		b.logger.Warn("Process force killed: id=%s", id)
		return nil
	}
}

// GetProcess returns information about a process by ID
func (b *Broker) GetProcess(id string) (*ProcessInfo, error) {
	b.mu.RLock()
	proc, exists := b.processes[id]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process with id '%s' not found", id)
	}

	proc.mu.RLock()
	defer proc.mu.RUnlock()

	// Return a copy of the process info
	info := *proc.info
	return &info, nil
}

// ListProcesses returns information about all managed processes
func (b *Broker) ListProcesses() []*ProcessInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*ProcessInfo, 0, len(b.processes))
	for _, proc := range b.processes {
		proc.mu.RLock()
		info := *proc.info
		proc.mu.RUnlock()
		result = append(result, &info)
	}

	return result
}

// SendMessage sends a message to the broker's message channel
func (b *Broker) SendMessage(msg *Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("broker message channel is closed")
		}
	}()
	
	select {
	case b.messages <- msg:
		return nil
	case <-b.ctx.Done():
		return fmt.Errorf("broker is shutting down")
	}
}

// ReceiveMessage receives a message from the broker's message channel
// It blocks until a message is available or the context is cancelled
func (b *Broker) ReceiveMessage(ctx context.Context) (*Message, error) {
	select {
	case msg := <-b.messages:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.ctx.Done():
		return nil, fmt.Errorf("broker is shutting down")
	}
}

// Shutdown gracefully shuts down the broker and all managed processes
func (b *Broker) Shutdown() error {
	b.logger.Info("Shutting down broker")

	// Cancel the broker context
	b.cancel()

	// Stop all managed processes
	b.mu.RLock()
	managedIDs := make([]string, 0, len(b.managedProcesses))
	for id := range b.managedProcesses {
		managedIDs = append(managedIDs, id)
	}
	b.mu.RUnlock()

	for _, id := range managedIDs {
		_ = b.StopManagedProcess(id)
	}

	// Kill all running processes
	b.mu.RLock()
	processIDs := make([]string, 0, len(b.processes))
	for id := range b.processes {
		processIDs = append(processIDs, id)
	}
	b.mu.RUnlock()

	for _, id := range processIDs {
		proc, exists := b.processes[id]
		if exists {
			proc.mu.RLock()
			status := proc.info.Status
			proc.mu.RUnlock()

			if status == ProcessStatusRunning {
				_ = b.Kill(id)
			}
		}
	}

	// Wait for all processes to complete
	b.wg.Wait()

	// Close the message channel
	close(b.messages)

	b.logger.Info("Broker shutdown complete")
	return nil
}

// RegisterManagedProcess registers a ManagedProcess with the broker and starts it
func (b *Broker) RegisterManagedProcess(proc ManagedProcess) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := proc.ID()
	if _, exists := b.managedProcesses[id]; exists {
		return fmt.Errorf("managed process with id '%s' already exists", id)
	}

	// Start the process
	if err := proc.Start(b.ctx, b); err != nil {
		return fmt.Errorf("failed to start managed process: %w", err)
	}

	b.managedProcesses[id] = proc
	b.logger.Info("Registered managed process: id=%s", id)

	return nil
}

// UnregisterManagedProcess removes a ManagedProcess from the broker
func (b *Broker) UnregisterManagedProcess(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.managedProcesses[id]; !exists {
		return fmt.Errorf("managed process with id '%s' not found", id)
	}

	delete(b.managedProcesses, id)
	b.logger.Info("Unregistered managed process: id=%s", id)
	return nil
}

// StopManagedProcess stops a managed process by ID
func (b *Broker) StopManagedProcess(id string) error {
	b.mu.RLock()
	proc, exists := b.managedProcesses[id]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("managed process with id '%s' not found", id)
	}

	b.logger.Info("Stopping managed process: id=%s", id)
	if err := proc.Stop(); err != nil {
		return fmt.Errorf("failed to stop managed process: %w", err)
	}

	return b.UnregisterManagedProcess(id)
}

// GetManagedProcess returns a managed process by ID
func (b *Broker) GetManagedProcess(id string) (ManagedProcess, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	proc, exists := b.managedProcesses[id]
	if !exists {
		return nil, fmt.Errorf("managed process with id '%s' not found", id)
	}

	return proc, nil
}

// ListManagedProcesses returns all registered managed processes
func (b *Broker) ListManagedProcesses() []ManagedProcess {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]ManagedProcess, 0, len(b.managedProcesses))
	for _, proc := range b.managedProcesses {
		result = append(result, proc)
	}

	return result
}

// DispatchMessage sends a message to a specific managed process
func (b *Broker) DispatchMessage(processID string, msg *Message) error {
	b.mu.RLock()
	proc, exists := b.managedProcesses[processID]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("managed process with id '%s' not found", processID)
	}

	return proc.OnMessage(msg)
}
