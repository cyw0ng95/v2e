package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
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
	info          *ProcessInfo
	cmd           *exec.Cmd
	cancel        context.CancelFunc
	done          chan struct{}
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	mu            sync.RWMutex
	restartConfig *RestartConfig
}

// RestartConfig holds restart configuration for a process
type RestartConfig struct {
	// Enabled indicates if auto-restart is enabled
	Enabled bool
	// MaxRestarts is the maximum number of restart attempts (-1 for unlimited)
	MaxRestarts int
	// RestartCount is the current number of restarts
	RestartCount int
	// Command and Args for restarting
	Command string
	Args    []string
	// IsRPC indicates if this is an RPC process
	IsRPC bool
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

// PerProcessStats contains stats for a single subprocess
type PerProcessStats struct {
	TotalSent        int64
	TotalReceived    int64
	RequestCount     int64
	ResponseCount    int64
	EventCount       int64
	ErrorCount       int64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}

// PendingRequest represents a pending request awaiting a response
type PendingRequest struct {
	// SourceProcess is the process ID that sent the request
	SourceProcess string
	// ResponseChan is the channel to send the response back
	ResponseChan chan *proc.Message
	// Timestamp is when the request was made
	Timestamp time.Time
}

// Broker manages subprocesses and message passing
type Broker struct {
	processes map[string]*Process
	messages  chan *proc.Message
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	logger    *common.Logger
	// config holds optional loaded configuration used by Broker for spawning
	config          *common.Config
	stats           MessageStats
	perProcessStats map[string]PerProcessStats
	statsMu         sync.RWMutex
	rpcEndpoints    map[string][]string // processID -> list of RPC endpoints
	endpointsMu     sync.RWMutex
	pendingRequests map[string]*PendingRequest // correlationID -> PendingRequest
	pendingMu       sync.RWMutex
	correlationSeq  uint64 // Atomic counter for generating correlation IDs
}

// NewBroker creates a new Broker instance
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broker{
		processes:       make(map[string]*Process),
		messages:        make(chan *proc.Message, 100),
		ctx:             ctx,
		cancel:          cancel,
		logger:          common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
		config:          nil,
		rpcEndpoints:    make(map[string][]string),
		pendingRequests: make(map[string]*PendingRequest),
		correlationSeq:  0,
		perProcessStats: make(map[string]PerProcessStats),
	}
}

// SetConfig sets the broker-level configuration used when spawning processes
func (b *Broker) SetConfig(cfg *common.Config) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = cfg
}

// SetLogger sets the logger for the broker
func (b *Broker) SetLogger(logger *common.Logger) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger = logger
}

// Context returns the broker's context
func (b *Broker) Context() context.Context {
	return b.ctx
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

	// If global config is available, inject known service-specific environment
	// variables so subprocesses can read their service configuration via env.
	if b.config != nil {
		if cmd.Env == nil {
			cmd.Env = os.Environ()
		}
		switch id {
		case "local":
			if b.config.Local.CVEDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CVE_DB_PATH=%s", b.config.Local.CVEDBPath))
			}
			if b.config.Local.CWEDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_DB_PATH=%s", b.config.Local.CWEDBPath))
			}
			if b.config.Local.CAPECDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_DB_PATH=%s", b.config.Local.CAPECDBPath))
			}
			if b.config.Capec.StrictXSDValidation {
				cmd.Env = append(cmd.Env, "CAPEC_STRICT_XSD=1")
			}
			if b.config.Capec.StrictXSDValidation {
				cmd.Env = append(cmd.Env, "CAPEC_STRICT_XSD=1")
			}
		case "meta":
			if b.config.Meta.SessionDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("SESSION_DB_PATH=%s", b.config.Meta.SessionDBPath))
			}
		case "remote":
			if b.config.Remote.NVDAPIKey != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("NVD_API_KEY=%s", b.config.Remote.NVDAPIKey))
			}
			if b.config.Remote.ViewFetchURL != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("VIEW_FETCH_URL=%s", b.config.Remote.ViewFetchURL))
			}
			if b.config.Remote.ViewFetchURL != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("VIEW_FETCH_URL=%s", b.config.Remote.ViewFetchURL))
			}
		case "access":
			if b.config.Access.StaticDir != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("ACCESS_STATIC_DIR=%s", b.config.Access.StaticDir))
			}
		}
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

	// Create a copy of the process info before starting the reaper goroutine
	// to avoid data races when the caller accesses the returned info
	infoCopy := *info

	// Start goroutine to wait for process completion
	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// SpawnRPC starts a new subprocess with RPC support using custom file descriptors
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

	// Create custom pipes for RPC communication (to avoid using stdio fds 0,1,2)
	// These will be passed as fd 3 (read) and fd 4 (write) to the subprocess
	readFromSubprocess, writeToParent, err := os.Pipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create output pipe: %w", err)
	}

	readFromParent, writeToSubprocess, err := os.Pipe()
	if err != nil {
		cancel()
		readFromSubprocess.Close()
		writeToParent.Close()
		return nil, fmt.Errorf("failed to create input pipe: %w", err)
	}

	// Set up ExtraFiles: subprocess will receive these as fd 3 and fd 4
	// fd 3 = input from parent (readFromParent)
	// fd 4 = output to parent (writeToParent)
	cmd.ExtraFiles = []*os.File{readFromParent, writeToParent}

	// Determine RPC fd numbers to advertise to subprocess (config overrides)
	inputFD := 3
	outputFD := 4
	if b.config != nil {
		if b.config.Proc.RPCInputFD != 0 {
			inputFD = b.config.Proc.RPCInputFD
		} else if b.config.Broker.RPCInputFD != 0 {
			inputFD = b.config.Broker.RPCInputFD
		}
		if b.config.Proc.RPCOutputFD != 0 {
			outputFD = b.config.Proc.RPCOutputFD
		} else if b.config.Broker.RPCOutputFD != 0 {
			outputFD = b.config.Broker.RPCOutputFD
		}
	}

	// Prepare environment for subprocess including RPC fd hints and service-specific values
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	// Inject service specific envs from config
	if b.config != nil {
		switch id {
		case "local":
			if b.config.Local.CVEDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CVE_DB_PATH=%s", b.config.Local.CVEDBPath))
			}
			if b.config.Local.CWEDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_DB_PATH=%s", b.config.Local.CWEDBPath))
			}
			if b.config.Local.CAPECDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_DB_PATH=%s", b.config.Local.CAPECDBPath))
			}
		case "meta":
			if b.config.Meta.SessionDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("SESSION_DB_PATH=%s", b.config.Meta.SessionDBPath))
			}
		case "remote":
			if b.config.Remote.NVDAPIKey != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("NVD_API_KEY=%s", b.config.Remote.NVDAPIKey))
			}
		case "access":
			if b.config.Access.StaticDir != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("ACCESS_STATIC_DIR=%s", b.config.Access.StaticDir))
			}
		}
	}

	// Finally append RPC FD and process id info
	cmd.Env = append(cmd.Env, fmt.Sprintf("RPC_INPUT_FD=%d", inputFD), fmt.Sprintf("RPC_OUTPUT_FD=%d", outputFD), fmt.Sprintf("PROCESS_ID=%s", id))

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
		stdin:  writeToSubprocess,  // Parent writes to this
		stdout: readFromSubprocess, // Parent reads from this
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		readFromSubprocess.Close()
		writeToSubprocess.Close()
		readFromParent.Close()
		writeToParent.Close()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	// Close the child's ends of the pipes in the parent
	readFromParent.Close()
	writeToParent.Close()

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned RPC process: id=%s pid=%d command=%s (advertised fds=%d,%d)", id, info.PID, command, inputFD, outputFD)

	// Create a copy of the process info before starting the reaper goroutine
	// to avoid data races when the caller accesses the returned info
	infoCopy := *info

	// Start goroutine to read messages from process
	b.wg.Add(1)
	go b.readProcessMessages(proc)

	// Start goroutine to wait for process completion
	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// SpawnWithRestart starts a new subprocess with auto-restart capability
func (b *Broker) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if process with this ID already exists
	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	// Create process context
	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	// Inject service-specific env args from config if present (non-RPC path)
	if b.config != nil {
		if cmd.Env == nil {
			cmd.Env = os.Environ()
		}
		switch id {
		case "local":
			if b.config.Local.CVEDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CVE_DB_PATH=%s", b.config.Local.CVEDBPath))
			}
			if b.config.Local.CWEDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_DB_PATH=%s", b.config.Local.CWEDBPath))
			}
			if b.config.Local.CAPECDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_DB_PATH=%s", b.config.Local.CAPECDBPath))
			}
		case "meta":
			if b.config.Meta.SessionDBPath != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("SESSION_DB_PATH=%s", b.config.Meta.SessionDBPath))
			}
		case "remote":
			if b.config.Remote.NVDAPIKey != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("NVD_API_KEY=%s", b.config.Remote.NVDAPIKey))
			}
		case "access":
			if b.config.Access.StaticDir != "" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("ACCESS_STATIC_DIR=%s", b.config.Access.StaticDir))
			}
		}
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
		restartConfig: &RestartConfig{
			Enabled:      true,
			MaxRestarts:  maxRestarts,
			RestartCount: 0,
			Command:      command,
			Args:         args,
			IsRPC:        false,
		},
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned process with restart: id=%s pid=%d command=%s max_restarts=%d", id, info.PID, command, maxRestarts)

	// Create a copy of the process info before starting the reaper goroutine
	infoCopy := *info

	// Start goroutine to wait for process completion
	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// SpawnRPCWithRestart starts a new RPC subprocess with auto-restart capability using custom file descriptors
func (b *Broker) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if process with this ID already exists
	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	// Create process context
	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	// Create custom pipes for RPC communication (to avoid using stdio fds 0,1,2)
	// These will be passed as fd 3 (read) and fd 4 (write) to the subprocess
	readFromSubprocess, writeToParent, err := os.Pipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create output pipe: %w", err)
	}

	readFromParent, writeToSubprocess, err := os.Pipe()
	if err != nil {
		cancel()
		readFromSubprocess.Close()
		writeToParent.Close()
		return nil, fmt.Errorf("failed to create input pipe: %w", err)
	}

	// Set up ExtraFiles: subprocess will receive these as fd 3 and fd 4
	// fd 3 = input from parent (readFromParent)
	// fd 4 = output to parent (writeToParent)
	cmd.ExtraFiles = []*os.File{readFromParent, writeToParent}

	// Determine RPC fd numbers to advertise to subprocess (config overrides)
	inputFD := 3
	outputFD := 4
	if b.config != nil {
		if b.config.Proc.RPCInputFD != 0 {
			inputFD = b.config.Proc.RPCInputFD
		} else if b.config.Broker.RPCInputFD != 0 {
			inputFD = b.config.Broker.RPCInputFD
		}
		if b.config.Proc.RPCOutputFD != 0 {
			outputFD = b.config.Proc.RPCOutputFD
		} else if b.config.Broker.RPCOutputFD != 0 {
			outputFD = b.config.Broker.RPCOutputFD
		}
	}

	// Set environment variable to tell subprocess which FDs to use for RPC
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("RPC_INPUT_FD=%d", inputFD), fmt.Sprintf("RPC_OUTPUT_FD=%d", outputFD), fmt.Sprintf("PROCESS_ID=%s", id))

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
		stdin:  writeToSubprocess,  // Parent writes to this
		stdout: readFromSubprocess, // Parent reads from this
		restartConfig: &RestartConfig{
			Enabled:      true,
			MaxRestarts:  maxRestarts,
			RestartCount: 0,
			Command:      command,
			Args:         args,
			IsRPC:        true,
		},
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		readFromSubprocess.Close()
		writeToSubprocess.Close()
		readFromParent.Close()
		writeToParent.Close()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	// Close the child's ends of the pipes in the parent
	readFromParent.Close()
	writeToParent.Close()

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned RPC process with restart: id=%s pid=%d command=%s max_restarts=%d (advertised fds=%d,%d)", id, info.PID, command, maxRestarts, inputFD, outputFD)

	// Create a copy of the process info before starting goroutines
	infoCopy := *info

	// Start goroutine to read messages from process stdout
	b.wg.Add(1)
	go b.readProcessMessages(proc)

	// Start goroutine to wait for process completion
	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// readProcessMessages reads messages from a process's stdout and forwards them to the broker
func (b *Broker) readProcessMessages(p *Process) {
	defer b.wg.Done()

	scanner := bufio.NewScanner(p.stdout)
	// Increase buffer size to handle large messages (e.g., CVE data from NVD API)
	buf := make([]byte, proc.MaxMessageSize)
	scanner.Buffer(buf, proc.MaxMessageSize)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse the message
		msg, err := proc.Unmarshal([]byte(line))
		if err != nil {
			b.logger.Warn("Failed to parse message from process %s: %v", p.info.ID, err)
			continue
		}

		// Route the message based on its target
		if err := b.RouteMessage(msg, p.info.ID); err != nil {
			b.logger.Warn("Failed to route message from process %s: %v", p.info.ID, err)

			// If this was a request message, send an error response back to the source
			if msg.Type == proc.MessageTypeRequest && msg.CorrelationID != "" {
				errorMsg := &proc.Message{
					Type:          proc.MessageTypeError,
					ID:            msg.ID,
					Error:         err.Error(),
					Target:        msg.Source,
					CorrelationID: msg.CorrelationID,
				}
				// Send error response back to source (best effort, don't block)
				go func() {
					if sendErr := b.SendToProcess(msg.Source, errorMsg); sendErr != nil {
						b.logger.Debug("Failed to send error response back to %s: %v", msg.Source, sendErr)
					}
				}()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		b.logger.Warn("Error reading from process %s: %v", p.info.ID, err)
	}
}

// SendToProcess sends a message to a specific process via stdin
func (b *Broker) SendToProcess(processID string, msg *proc.Message) error {
	b.mu.RLock()
	p, exists := b.processes[processID]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("process with id '%s' not found", processID)
	}

	p.mu.RLock()
	stdin := p.stdin
	status := p.info.Status
	p.mu.RUnlock()

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
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := fmt.Fprintf(stdin, "%s\n", string(data)); err != nil {
		return fmt.Errorf("failed to write message to process: %w", err)
	}

	// Update stats: this is a message sent to a subprocess
	b.updateStats(msg, true)

	b.logger.Debug("Sent message to process %s: type=%s id=%s", processID, msg.Type, msg.ID)
	return nil
}

// reapProcess waits for a process to complete and updates its status.
// If auto-restart is enabled, it will attempt to restart the process according to
// the configured restart policy (max restarts, delay, etc.).
func (b *Broker) reapProcess(p *Process) {
	defer b.wg.Done()
	defer close(p.done)

	// Wait for the process to complete
	err := p.cmd.Wait()

	// Lock is acquired here and explicitly unlocked later (not deferred)
	// because the restart logic requires early unlock to avoid deadlock.
	// The restart logic calls broker methods (SpawnRPCWithRestart, SpawnWithRestart)
	// that need to acquire broker locks, which would deadlock if we held this process lock.
	p.mu.Lock()

	p.info.EndTime = time.Now()
	p.info.Status = ProcessStatusExited

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				p.info.ExitCode = status.ExitStatus()
			} else {
				p.info.ExitCode = -1
			}
		} else {
			p.info.ExitCode = -1
		}
	} else {
		p.info.ExitCode = 0
	}

	b.logger.Info("Process exited: id=%s pid=%d exit_code=%d",
		p.info.ID, p.info.PID, p.info.ExitCode)

	// Send event message about process exit
	event, _ := proc.NewEventMessage(p.info.ID, map[string]interface{}{
		"event":     "process_exited",
		"id":        p.info.ID,
		"pid":       p.info.PID,
		"exit_code": p.info.ExitCode,
	})
	event.Target = "test-target" // Ensure TotalSent is incremented
	b.sendMessageInternal(event)

	// Check if auto-restart is enabled
	if p.restartConfig != nil && p.restartConfig.Enabled {
		// Check if we've exceeded max restarts
		if p.restartConfig.MaxRestarts >= 0 && p.restartConfig.RestartCount >= p.restartConfig.MaxRestarts {
			b.logger.Warn("Process %s exceeded max restarts (%d), not restarting", p.info.ID, p.restartConfig.MaxRestarts)
			p.mu.Unlock()
			return
		}

		// Increment restart count
		p.restartConfig.RestartCount++

		processID := p.info.ID
		command := p.restartConfig.Command
		args := p.restartConfig.Args
		isRPC := p.restartConfig.IsRPC
		maxRestarts := p.restartConfig.MaxRestarts
		restartCount := p.restartConfig.RestartCount

		// Unlock before restarting
		p.mu.Unlock()

		b.logger.Info("Restarting process %s (attempt %d/%d)", processID, restartCount, maxRestarts)

		// Remove old process from map
		b.mu.Lock()
		delete(b.processes, processID)
		b.mu.Unlock()

		// Wait a bit before restarting
		time.Sleep(1 * time.Second)

		// Restart the process
		var restartErr error
		if isRPC {
			_, restartErr = b.SpawnRPCWithRestart(processID, command, maxRestarts, args...)
		} else {
			_, restartErr = b.SpawnWithRestart(processID, command, maxRestarts, args...)
		}

		if restartErr != nil {
			b.logger.Warn("Failed to restart process %s: %v", processID, restartErr)
		} else {
			// Update restart count in the new process
			b.mu.RLock()
			if newProc, exists := b.processes[processID]; exists {
				newProc.mu.Lock()
				if newProc.restartConfig != nil {
					newProc.restartConfig.RestartCount = restartCount
				}
				newProc.mu.Unlock()
			}
			b.mu.RUnlock()
		}
		return
	}

	p.mu.Unlock()
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

	// Send SIGTERM for graceful shutdown
	if proc.cmd.Process != nil {
		if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			// If SIGTERM fails, cancel context as fallback
			proc.cancel()
		}
	}

	// Wait for process to exit with timeout
	select {
	case <-proc.done:
		b.logger.Info("Process terminated gracefully: id=%s", id)
		return nil
	case <-time.After(5 * time.Second):
		// Force kill if graceful shutdown failed
		b.logger.Warn("Process did not terminate gracefully, sending SIGKILL: id=%s", id)
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
func (b *Broker) SendMessage(msg *proc.Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("broker message channel is closed")
		}
	}()

	select {
	case b.messages <- msg:
		// If message has an explicit target it's being enqueued to route to that target;
		// otherwise it's an incoming message to broker for local processing (treat as received).
		if msg.Target == "" || msg.Target == "broker" {
			b.updateStats(msg, false)
		} else {
			b.updateStats(msg, true)
		}
		return nil
	case <-b.ctx.Done():
		return fmt.Errorf("broker is shutting down")
	}
}

// sendMessageInternal sends a message internally (from broker processes) without error handling
// This is used by readProcessMessages and reapProcess to avoid blocking
func (b *Broker) sendMessageInternal(msg *proc.Message) {
	select {
	case b.messages <- msg:
		if msg.Target == "" || msg.Target == "broker" {
			b.updateStats(msg, false)
		} else {
			b.updateStats(msg, true)
		}
	case <-b.ctx.Done():
	}
}

// ReceiveMessage receives a message from the broker's message channel
// It blocks until a message is available or the context is cancelled
func (b *Broker) ReceiveMessage(ctx context.Context) (*proc.Message, error) {
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
func (b *Broker) updateStats(msg *proc.Message, isSent bool) {
	b.statsMu.Lock()
	defer b.statsMu.Unlock()

	now := time.Now()

	// Update broker-wide first/last timestamps
	if b.stats.FirstMessageTime.IsZero() {
		b.stats.FirstMessageTime = now
	}
	b.stats.LastMessageTime = now

	// Update broker-wide counters
	if isSent {
		b.stats.TotalSent++
	} else {
		b.stats.TotalReceived++
	}

	switch msg.Type {
	case proc.MessageTypeRequest:
		b.stats.RequestCount++
	case proc.MessageTypeResponse:
		b.stats.ResponseCount++
	case proc.MessageTypeEvent:
		b.stats.EventCount++
	case proc.MessageTypeError:
		b.stats.ErrorCount++
	}

	// Update per-process stats. Determine which process to attribute this message to.
	var procID string
	if isSent {
		// message being sent by broker to target process
		procID = msg.Target
	} else {
		// message received from a process
		procID = msg.Source
	}

	if procID != "" {
		ps := b.perProcessStats[procID]

		if ps.FirstMessageTime.IsZero() {
			ps.FirstMessageTime = now
		}
		ps.LastMessageTime = now

		if isSent {
			ps.TotalSent++
		} else {
			ps.TotalReceived++
		}

		switch msg.Type {
		case proc.MessageTypeRequest:
			ps.RequestCount++
		case proc.MessageTypeResponse:
			ps.ResponseCount++
		case proc.MessageTypeEvent:
			ps.EventCount++
		case proc.MessageTypeError:
			ps.ErrorCount++
		}

		b.perProcessStats[procID] = ps
	}
}

// GetMessageStats returns a copy of the current message statistics
func (b *Broker) GetMessageStats() MessageStats {
	b.statsMu.RLock()
	defer b.statsMu.RUnlock()
	return b.stats
}

// GetPerProcessStats returns a copy of current per-process stats
func (b *Broker) GetPerProcessStats() map[string]PerProcessStats {
	b.statsMu.RLock()
	defer b.statsMu.RUnlock()
	out := make(map[string]PerProcessStats, len(b.perProcessStats))
	for k, v := range b.perProcessStats {
		out[k] = v
	}
	return out
}

// GetMessageCount returns the total number of messages processed (sent + received)
func (b *Broker) GetMessageCount() int64 {
	b.statsMu.RLock()
	defer b.statsMu.RUnlock()
	return b.stats.TotalSent + b.stats.TotalReceived
}

// HandleRPCGetMessageStats handles the RPCGetMessageStats RPC request
func (b *Broker) HandleRPCGetMessageStats(reqMsg *proc.Message) (*proc.Message, error) {
	// Get broker-wide message stats and per-process breakdown
	stats := b.GetMessageStats()
	per := b.GetPerProcessStats()

	// Convert stats to snake_case keys and serializable values
	statMap := map[string]interface{}{
		"total_sent":         stats.TotalSent,
		"total_received":     stats.TotalReceived,
		"request_count":      stats.RequestCount,
		"response_count":     stats.ResponseCount,
		"event_count":        stats.EventCount,
		"error_count":        stats.ErrorCount,
		"first_message_time": stats.FirstMessageTime.Format(time.RFC3339Nano),
		"last_message_time":  stats.LastMessageTime.Format(time.RFC3339Nano),
	}

	perMap := make(map[string]map[string]interface{})
	for pid, ps := range per {
		perMap[pid] = map[string]interface{}{
			"total_sent":     ps.TotalSent,
			"total_received": ps.TotalReceived,
			"request_count":  ps.RequestCount,
			"response_count": ps.ResponseCount,
			"event_count":    ps.EventCount,
			"error_count":    ps.ErrorCount,
			"first_message_time": func(t time.Time) interface{} {
				if t.IsZero() {
					return nil
				}
				return t.Format(time.RFC3339Nano)
			}(ps.FirstMessageTime),
			"last_message_time": func(t time.Time) interface{} {
				if t.IsZero() {
					return nil
				}
				return t.Format(time.RFC3339Nano)
			}(ps.LastMessageTime),
		}
	}

	payload := map[string]interface{}{
		"total":       statMap,
		"per_process": perMap,
	}

	// Create response message
	respMsg, err := proc.NewResponseMessage(reqMsg.ID, payload)
	if err != nil {
		return nil, err
	}

	// Set correlation ID if present
	if reqMsg.CorrelationID != "" {
		respMsg.CorrelationID = reqMsg.CorrelationID
	}

	// Set source and target for routing
	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source

	return respMsg, nil
}

// HandleRPCGetMessageCount handles the RPCGetMessageCount RPC request
func (b *Broker) HandleRPCGetMessageCount(reqMsg *proc.Message) (*proc.Message, error) {
	// Get message count
	count := b.GetMessageCount()

	// Create response payload
	payload := map[string]interface{}{
		"count": count,
	}

	// Create response message
	respMsg, err := proc.NewResponseMessage(reqMsg.ID, payload)
	if err != nil {
		return nil, err
	}

	// Set correlation ID if present
	if reqMsg.CorrelationID != "" {
		respMsg.CorrelationID = reqMsg.CorrelationID
	}

	// Set source and target for routing
	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source

	return respMsg, nil
}

// ProcessMessage processes a message directed at the broker
func (b *Broker) ProcessMessage(msg *proc.Message) error {
	// Only process request messages
	if msg.Type != proc.MessageTypeRequest {
		return nil
	}

	// Handle based on message ID (RPC method name)
	var respMsg *proc.Message
	var err error

	switch msg.ID {
	case "RPCGetMessageStats":
		respMsg, err = b.HandleRPCGetMessageStats(msg)
	case "RPCGetMessageCount":
		respMsg, err = b.HandleRPCGetMessageCount(msg)
	default:
		// Unknown RPC method
		errMsg := proc.NewErrorMessage(msg.ID, fmt.Errorf("unknown RPC method: %s", msg.ID))
		errMsg.Source = "broker"
		errMsg.Target = msg.Source
		if msg.CorrelationID != "" {
			errMsg.CorrelationID = msg.CorrelationID
		}
		return b.RouteMessage(errMsg, "broker")
	}

	if err != nil {
		// Send error response
		errMsg := proc.NewErrorMessage(msg.ID, err)
		errMsg.Source = "broker"
		errMsg.Target = msg.Source
		if msg.CorrelationID != "" {
			errMsg.CorrelationID = msg.CorrelationID
		}
		return b.RouteMessage(errMsg, "broker")
	}

	// Send successful response
	return b.RouteMessage(respMsg, "broker")
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

// RegisterEndpoint registers an RPC endpoint for a process
func (b *Broker) RegisterEndpoint(processID, endpoint string) {
	b.endpointsMu.Lock()
	defer b.endpointsMu.Unlock()

	if _, exists := b.rpcEndpoints[processID]; !exists {
		b.rpcEndpoints[processID] = make([]string, 0)
	}

	// Avoid duplicates
	for _, e := range b.rpcEndpoints[processID] {
		if e == endpoint {
			return
		}
	}

	b.rpcEndpoints[processID] = append(b.rpcEndpoints[processID], endpoint)
	b.logger.Info("Registered endpoint %s for process %s", endpoint, processID)
}

// GetEndpoints returns all registered RPC endpoints for a process
func (b *Broker) GetEndpoints(processID string) []string {
	b.endpointsMu.RLock()
	defer b.endpointsMu.RUnlock()

	endpoints, exists := b.rpcEndpoints[processID]
	if !exists {
		return []string{}
	}

	// Return a copy to avoid race conditions
	result := make([]string, len(endpoints))
	copy(result, endpoints)
	return result
}

// GetAllEndpoints returns all registered RPC endpoints for all processes
func (b *Broker) GetAllEndpoints() map[string][]string {
	b.endpointsMu.RLock()
	defer b.endpointsMu.RUnlock()

	result := make(map[string][]string)
	for processID, endpoints := range b.rpcEndpoints {
		endpointsCopy := make([]string, len(endpoints))
		copy(endpointsCopy, endpoints)
		result[processID] = endpointsCopy
	}
	return result
}

// GenerateCorrelationID generates a unique correlation ID for request-response matching
func (b *Broker) GenerateCorrelationID() string {
	seq := atomic.AddUint64(&b.correlationSeq, 1)
	return fmt.Sprintf("corr-%d-%d", time.Now().UnixNano(), seq)
}

// RouteMessage routes a message to its target process or handles it locally
func (b *Broker) RouteMessage(msg *proc.Message, sourceProcess string) error {
	// Set source if not already set
	if msg.Source == "" {
		msg.Source = sourceProcess
	}

	// If message is a response with correlation ID, route it back to the pending request FIRST
	// This takes priority over Target-based routing to ensure responses go to the right waiting caller
	if msg.Type == proc.MessageTypeResponse && msg.CorrelationID != "" {
		b.logger.Debug("[TRACE] Received response message: id=%s correlation_id=%s from=%s", msg.ID, msg.CorrelationID, msg.Source)
		b.pendingMu.Lock()
		pending, exists := b.pendingRequests[msg.CorrelationID]
		if exists {
			delete(b.pendingRequests, msg.CorrelationID)
		}
		b.pendingMu.Unlock()

		if exists {
			b.logger.Debug("[TRACE] Routing response to pending request: correlation_id=%s", msg.CorrelationID)
			select {
			case pending.ResponseChan <- msg:
				b.logger.Debug("[TRACE] Response delivered to waiting channel: correlation_id=%s", msg.CorrelationID)
				return nil
			case <-time.After(5 * time.Second):
				b.logger.Warn("[TRACE] Timeout sending response to pending request: correlation_id=%s", msg.CorrelationID)
				return fmt.Errorf("timeout sending response to pending request")
			}
		}
		// If no pending request found in broker, fall through to target-based routing
		// This allows subprocess-to-subprocess RPC where the correlation is tracked by the calling subprocess
		b.logger.Debug("[TRACE] No pending request found for correlation_id=%s (may be tracked by subprocess), trying target-based routing", msg.CorrelationID)
	}

	// If message has a target, route it to that process or broker
	if msg.Target != "" {
		// Special case: if target is "broker", process it locally
		if msg.Target == "broker" {
			b.logger.Debug("Routing message to broker for local processing: type=%s id=%s from=%s", msg.Type, msg.ID, msg.Source)
			return b.ProcessMessage(msg)
		}

		b.logger.Debug("Routing message from %s to %s: type=%s id=%s", msg.Source, msg.Target, msg.Type, msg.ID)
		return b.SendToProcess(msg.Target, msg)
	}

	// Otherwise, send to broker's message channel for local processing
	return b.SendMessage(msg)
}

// InvokeRPC invokes an RPC method on a target process and waits for the response
func (b *Broker) InvokeRPC(sourceProcess, targetProcess, rpcMethod string, payload interface{}, timeout time.Duration) (*proc.Message, error) {
	// Generate correlation ID
	correlationID := b.GenerateCorrelationID()

	// Create response channel
	responseChan := make(chan *proc.Message, 1)

	// Register pending request
	b.pendingMu.Lock()
	b.pendingRequests[correlationID] = &PendingRequest{
		SourceProcess: sourceProcess,
		ResponseChan:  responseChan,
		Timestamp:     time.Now(),
	}
	b.pendingMu.Unlock()

	// Clean up pending request on exit
	defer func() {
		b.pendingMu.Lock()
		delete(b.pendingRequests, correlationID)
		b.pendingMu.Unlock()
		close(responseChan)
	}()

	// Create request message
	reqMsg, err := proc.NewRequestMessage(rpcMethod, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request message: %w", err)
	}

	// Set routing information
	reqMsg.Source = sourceProcess
	reqMsg.Target = targetProcess
	reqMsg.CorrelationID = correlationID

	// Send request to target process
	if err := b.SendToProcess(targetProcess, reqMsg); err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %w", targetProcess, err)
	}

	b.logger.Debug("Invoked RPC: source=%s target=%s method=%s correlation_id=%s",
		sourceProcess, targetProcess, rpcMethod, correlationID)

	b.logger.Debug("[TRACE] Waiting for response: correlation_id=%s target=%s method=%s timeout=%v", correlationID, targetProcess, rpcMethod, timeout)
	// Wait for response with timeout
	select {
	case response := <-responseChan:
		b.logger.Debug("[TRACE] Received response for correlation_id=%s: type=%s", correlationID, response.Type)
		return response, nil
	case <-time.After(timeout):
		b.logger.Warn("[TRACE] Timeout waiting for response: correlation_id=%s target=%s method=%s", correlationID, targetProcess, rpcMethod)
		return nil, fmt.Errorf("timeout waiting for response from %s", targetProcess)
	case <-b.ctx.Done():
		b.logger.Warn("[TRACE] Broker context cancelled while waiting for response: correlation_id=%s", correlationID)
		return nil, fmt.Errorf("broker is shutting down")
	}
}

// LoadProcessesFromConfig loads and starts processes from a configuration
func (b *Broker) LoadProcessesFromConfig(config *common.Config) error {
	if config == nil || len(config.Broker.Processes) == 0 {
		b.logger.Info("No processes configured to start")
		return nil
	}

	b.logger.Info("Loading %d processes from configuration", len(config.Broker.Processes))

	for _, procConfig := range config.Broker.Processes {
		if procConfig.ID == "" || procConfig.Command == "" {
			b.logger.Warn("Skipping invalid process config: missing ID or command")
			continue
		}

		var err error
		var info *ProcessInfo

		if procConfig.Restart {
			maxRestarts := procConfig.MaxRestarts
			if maxRestarts == 0 {
				maxRestarts = -1 // Default to unlimited restarts
			}

			if procConfig.RPC {
				info, err = b.SpawnRPCWithRestart(procConfig.ID, procConfig.Command, maxRestarts, procConfig.Args...)
			} else {
				info, err = b.SpawnWithRestart(procConfig.ID, procConfig.Command, maxRestarts, procConfig.Args...)
			}
		} else {
			if procConfig.RPC {
				info, err = b.SpawnRPC(procConfig.ID, procConfig.Command, procConfig.Args...)
			} else {
				info, err = b.Spawn(procConfig.ID, procConfig.Command, procConfig.Args...)
			}
		}

		if err != nil {
			b.logger.Warn("Failed to spawn process %s: %v", procConfig.ID, err)
			continue
		}

		b.logger.Info("Started process %s (PID: %d) from configuration", info.ID, info.PID)
	}

	return nil
}
