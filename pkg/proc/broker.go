package proc

import (
	"bufio"
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
	stdin  io.WriteCloser
	stdout io.ReadCloser
	mu     sync.RWMutex
}

// MessageStats contains statistics about messages processed by the broker
type MessageStats struct {
	// TotalSent is the total number of messages sent through the broker
	TotalSent int64
	// TotalReceived is the total number of messages received by the broker
	TotalReceived int64
	// RequestCount is the number of request messages
	RequestCount int64
	// ResponseCount is the number of response messages
	ResponseCount int64
	// EventCount is the number of event messages
	EventCount int64
	// ErrorCount is the number of error messages
	ErrorCount int64
	// FirstMessageTime is when the first message was processed
	FirstMessageTime time.Time
	// LastMessageTime is when the last message was processed
	LastMessageTime time.Time
}

// Broker manages subprocesses and message passing
type Broker struct {
	processes map[string]*Process
	messages  chan *Message
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	logger    *common.Logger
	stats     MessageStats
	statsMu   sync.RWMutex
}

// NewBroker creates a new Broker instance
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broker{
		processes: make(map[string]*Process),
		messages:  make(chan *Message, 100),
		ctx:       ctx,
		cancel:    cancel,
		logger:    common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
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

// SpawnRPC starts a new subprocess with RPC support (stdin/stdout pipes)
// It returns the process ID and an error if the process failed to start
func (b *Broker) SpawnRPC(id, command string, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if process with this ID already exists
	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	// Create process context
	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	// Set up stdin and stdout pipes for RPC communication
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

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
		stdin:  stdin,
		stdout: stdout,
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		stdin.Close()
		stdout.Close()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned RPC process: id=%s pid=%d command=%s", id, info.PID, command)

	// Start goroutine to read messages from process stdout
	b.wg.Add(1)
	go b.readProcessMessages(proc)

	// Start goroutine to wait for process completion
	b.wg.Add(1)
	go b.reapProcess(proc)

	return info, nil
}

// readProcessMessages reads messages from a process's stdout and forwards them to the broker
func (b *Broker) readProcessMessages(proc *Process) {
	defer b.wg.Done()

	scanner := bufio.NewScanner(proc.stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse the message
		msg, err := Unmarshal([]byte(line))
		if err != nil {
			b.logger.Warn("Failed to parse message from process %s: %v", proc.info.ID, err)
			continue
		}

		// Forward the message to the broker's message channel
		b.sendMessageInternal(msg)
	}

	if err := scanner.Err(); err != nil {
		b.logger.Warn("Error reading from process %s: %v", proc.info.ID, err)
	}
}

// SendToProcess sends a message to a specific process via stdin
func (b *Broker) SendToProcess(processID string, msg *Message) error {
	b.mu.RLock()
	proc, exists := b.processes[processID]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("process with id '%s' not found", processID)
	}

	proc.mu.RLock()
	stdin := proc.stdin
	status := proc.info.Status
	proc.mu.RUnlock()

	if status != ProcessStatusRunning {
		return fmt.Errorf("process '%s' is not running", processID)
	}

	if stdin == nil {
		return fmt.Errorf("process '%s' does not support RPC (no stdin pipe)", processID)
	}

	// Marshal the message to JSON
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write the message to the process's stdin
	proc.mu.Lock()
	defer proc.mu.Unlock()

	if _, err := fmt.Fprintf(stdin, "%s\n", string(data)); err != nil {
		return fmt.Errorf("failed to write message to process: %w", err)
	}

	b.logger.Debug("Sent message to process %s: type=%s id=%s", processID, msg.Type, msg.ID)
	return nil
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
	b.sendMessageInternal(event)
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
		b.updateStats(msg, true)
		return nil
	case <-b.ctx.Done():
		return fmt.Errorf("broker is shutting down")
	}
}

// sendMessageInternal sends a message internally (from broker processes) without error handling
// This is used by readProcessMessages and reapProcess to avoid blocking
func (b *Broker) sendMessageInternal(msg *Message) {
	select {
	case b.messages <- msg:
		b.updateStats(msg, true)
	case <-b.ctx.Done():
	}
}

// ReceiveMessage receives a message from the broker's message channel
// It blocks until a message is available or the context is cancelled
func (b *Broker) ReceiveMessage(ctx context.Context) (*Message, error) {
	select {
	case msg := <-b.messages:
		b.updateStats(msg, false)
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.ctx.Done():
		return nil, fmt.Errorf("broker is shutting down")
	}
}

// updateStats updates message statistics
func (b *Broker) updateStats(msg *Message, isSent bool) {
	b.statsMu.Lock()
	defer b.statsMu.Unlock()

	now := time.Now()
	
	// Update first message time if not set
	if b.stats.FirstMessageTime.IsZero() {
		b.stats.FirstMessageTime = now
	}
	b.stats.LastMessageTime = now

	// Update counters
	if isSent {
		b.stats.TotalSent++
	} else {
		b.stats.TotalReceived++
	}

	// Update type-specific counters
	switch msg.Type {
	case MessageTypeRequest:
		b.stats.RequestCount++
	case MessageTypeResponse:
		b.stats.ResponseCount++
	case MessageTypeEvent:
		b.stats.EventCount++
	case MessageTypeError:
		b.stats.ErrorCount++
	}
}

// GetMessageStats returns a copy of the current message statistics
func (b *Broker) GetMessageStats() MessageStats {
	b.statsMu.RLock()
	defer b.statsMu.RUnlock()
	return b.stats
}

// GetMessageCount returns the total number of messages processed (sent + received)
func (b *Broker) GetMessageCount() int64 {
	b.statsMu.RLock()
	defer b.statsMu.RUnlock()
	return b.stats.TotalSent + b.stats.TotalReceived
}

// Shutdown gracefully shuts down the broker and all managed processes
func (b *Broker) Shutdown() error {
	b.logger.Info("Shutting down broker")

	// Cancel the broker context
	b.cancel()

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
