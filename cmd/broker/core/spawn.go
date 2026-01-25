package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"github.com/cyw0ng95/v2e/pkg/common"
)

// Spawn starts a new subprocess with the given command and arguments.
func (b *Broker) Spawn(id, command string, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	setProcessEnv(cmd, id, b.config)

	info := &ProcessInfo{ID: id, Command: command, Args: args, Status: ProcessStatusRunning, StartTime: time.Now()}

	proc := &Process{info: info, cmd: cmd, cancel: cancel, done: make(chan struct{})}

	if err := cmd.Start(); err != nil {
		cancel()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned process: id=%s pid=%d command=%s", id, info.PID, command)

	infoCopy := *info

	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// SpawnRPC starts a new subprocess with RPC support using custom file descriptors.
func (b *Broker) SpawnRPC(id, command string, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

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

	cmd.ExtraFiles = []*os.File{readFromParent, writeToParent}

	inputFD, outputFD := b.getRPCFileDescriptors()

	setProcessEnv(cmd, id, b.config)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PROCESS_ID=%s", id))
	if b.config != nil && b.config.Proc.MaxMessageSizeBytes != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SUBPROCESS_MAX_MESSAGE_SIZE=%d", b.config.Proc.MaxMessageSizeBytes))
	}
	cmd.Env = append(cmd.Env, "BROKER_PASSING_RPC_FDS=1")

	info := &ProcessInfo{ID: id, Command: command, Args: args, Status: ProcessStatusRunning, StartTime: time.Now()}

	proc := &Process{
		info:   info,
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
		stdin:  writeToSubprocess,
		stdout: readFromSubprocess,
	}

	if err := cmd.Start(); err != nil {
		cancel()
		readFromSubprocess.Close()
		writeToSubprocess.Close()
		readFromParent.Close()
		writeToParent.Close()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	readFromParent.Close()
	writeToParent.Close()

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned RPC process: id=%s pid=%d command=%s (advertised fds=%d,%d)", id, info.PID, command, inputFD, outputFD)

	infoCopy := *info

	b.registerProcessTransport(id, inputFD, outputFD)

	b.wg.Add(1)
	go b.readProcessMessages(proc)

	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// SpawnWithRestart starts a new subprocess with auto-restart capability.
func (b *Broker) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	setProcessEnv(cmd, id, b.config)

	info := &ProcessInfo{ID: id, Command: command, Args: args, Status: ProcessStatusRunning, StartTime: time.Now()}

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

	if err := cmd.Start(); err != nil {
		cancel()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned process with restart: id=%s pid=%d command=%s max_restarts=%d", id, info.PID, command, maxRestarts)

	infoCopy := *info

	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// SpawnRPCWithRestart starts a new RPC subprocess with auto-restart capability using custom file descriptors.
func (b *Broker) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

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

	cmd.ExtraFiles = []*os.File{readFromParent, writeToParent}

	inputFD, outputFD := b.getRPCFileDescriptors()

	setProcessEnv(cmd, id, b.config)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PROCESS_ID=%s", id))
	if b.config != nil && b.config.Proc.MaxMessageSizeBytes != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SUBPROCESS_MAX_MESSAGE_SIZE=%d", b.config.Proc.MaxMessageSizeBytes))
	}
	cmd.Env = append(cmd.Env, "BROKER_PASSING_RPC_FDS=1")

	info := &ProcessInfo{ID: id, Command: command, Args: args, Status: ProcessStatusRunning, StartTime: time.Now()}

	proc := &Process{
		info:   info,
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
		stdin:  writeToSubprocess,
		stdout: readFromSubprocess,
		restartConfig: &RestartConfig{
			Enabled:      true,
			MaxRestarts:  maxRestarts,
			RestartCount: 0,
			Command:      command,
			Args:         args,
			IsRPC:        true,
		},
	}

	if err := cmd.Start(); err != nil {
		cancel()
		readFromSubprocess.Close()
		writeToSubprocess.Close()
		readFromParent.Close()
		writeToParent.Close()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	readFromParent.Close()
	writeToParent.Close()

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	b.logger.Info("Spawned RPC process with restart: id=%s pid=%d command=%s max_restarts=%d (advertised fds=%d,%d)", id, info.PID, command, maxRestarts, inputFD, outputFD)

	infoCopy := *info

	b.registerProcessTransport(id, inputFD, outputFD)

	b.wg.Add(1)
	go b.readProcessMessages(proc)

	b.wg.Add(1)
	go b.reapProcess(proc)

	return &infoCopy, nil
}

// LoadProcessesFromConfig loads and starts processes from a configuration.
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
				maxRestarts = -1
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

// setProcessEnv configures environment variables for a process based on its ID and the broker config.
// This consolidates the repeated env setup logic from Spawn, SpawnRPC, SpawnWithRestart, and SpawnRPCWithRestart.
func setProcessEnv(cmd *exec.Cmd, processID string, config *common.Config) {
	if config == nil {
		return
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	switch processID {
	case "local":
		if config.Local.CVEDBPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("CVE_DB_PATH=%s", config.Local.CVEDBPath))
		}
		if config.Local.CWEDBPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_DB_PATH=%s", config.Local.CWEDBPath))
		}
		if config.Local.CAPECDBPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_DB_PATH=%s", config.Local.CAPECDBPath))
		}
		if config.Capec.StrictXSDValidation {
			cmd.Env = append(cmd.Env, "CAPEC_STRICT_XSD=1")
		}
	case "meta":
		if config.Meta.SessionDBPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("SESSION_DB_PATH=%s", config.Meta.SessionDBPath))
		}
	case "remote":
		if config.Remote.NVDAPIKey != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("NVD_API_KEY=%s", config.Remote.NVDAPIKey))
		}
		if config.Remote.ViewFetchURL != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("VIEW_FETCH_URL=%s", config.Remote.ViewFetchURL))
		}
	case "access":
		if config.Access.StaticDir != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("ACCESS_STATIC_DIR=%s", config.Access.StaticDir))
		}
	}
}

// getRPCFileDescriptors returns the configured input and output file descriptor numbers for RPC communication.
// Defaults to 3 (input) and 4 (output) if not configured.
func (b *Broker) getRPCFileDescriptors() (inputFD, outputFD int) {
	inputFD, outputFD = 3, 4
	if b.config == nil {
		return
	}
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
	return
}

// registerProcessTransport creates and registers the appropriate transport for a spawned RPC process.
// Returns an error only if transport registration fails critically.
func (b *Broker) registerProcessTransport(processID string, inputFD, outputFD int) {
	if b.transportManager == nil {
		return
	}
	if b.config != nil && shouldUseUDSTransport(b.config.Broker.Transport) {
		if err := b.transportManager.RegisterUDSTransport(processID, true); err == nil {
			b.logger.Debug("Registered UDS transport for process %s", processID)
			return
		}
		b.logger.Warn("Failed to connect UDS transport for process %s, falling back to FD transport", processID)
	}
	// Use FD transport (default or fallback)
	fdTransport := transport.NewFDPipeTransport(inputFD, outputFD)
	if err := fdTransport.Connect(); err == nil {
		b.transportManager.RegisterTransport(processID, fdTransport)
		b.logger.Debug("Registered FD transport for process %s", processID)
	} else {
		b.logger.Warn("Failed to connect FD transport for process %s: %v", processID, err)
	}
}

// shouldUseUDSTransport determines whether UDS transport should be used based on the transport configuration
func shouldUseUDSTransport(config common.TransportConfigOptions) bool {
	// If Type is explicitly set to "uds", use UDS
	if config.Type == "uds" {
		return true
	}
	// If Type is explicitly set to "fd", don't use UDS
	if config.Type == "fd" {
		return false
	}
	// If Type is "auto" or not set, fall back to EnableUDS flag
	// If both EnableUDS and EnableFD are set, prioritize UDS unless DualMode is enabled
	if config.EnableUDS && config.EnableFD {
		// In dual mode, we might need special handling, but for now default to UDS
		// If DualMode is enabled, we may want to handle differently
		return !config.DualMode // If dual mode, prefer FD initially
	}
	// Otherwise, use EnableUDS flag
	return config.EnableUDS
}
