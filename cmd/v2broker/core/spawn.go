package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/meta"
)

// serviceToSpawn represents a service to be spawned with its name and path.
type serviceToSpawn struct {
	name string
	path string
}

// normalizeServiceName converts a binary name (e.g., v2access) to the process ID (e.g., access).
// The subprocess binaries have v2* prefix but their internal ProcessID is without the prefix.
func normalizeServiceName(binaryName string) string {
	// Strip "v2" prefix if present (e.g., v2access -> access)
	if strings.HasPrefix(binaryName, "v2") && len(binaryName) > 2 {
		return binaryName[2:]
	}
	return binaryName
}

// Spawn starts a new subprocess with the given command and arguments.
func (b *Broker) Spawn(id, command string, args ...string) (*ProcessInfo, error) {
	return b.spawnInternal(id, command, args, nil)
}

// SpawnRPC starts a new subprocess with RPC support using custom file descriptors.
func (b *Broker) SpawnRPC(id, command string, args ...string) (*ProcessInfo, error) {
	restartConfig := &RestartConfig{
		Enabled: false,
		Command: command,
		Args:    args,
		IsRPC:   true,
	}
	return b.spawnInternal(id, command, args, restartConfig)
}

// SpawnWithRestart starts a new subprocess with auto-restart capability.
func (b *Broker) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error) {
	restartConfig := &RestartConfig{
		Enabled:     true,
		MaxRestarts: maxRestarts,
		Command:     command,
		Args:        args,
		IsRPC:       false,
	}
	return b.spawnInternal(id, command, args, restartConfig)
}

// SpawnRPCWithRestart starts a new RPC subprocess with auto-restart capability using custom file descriptors.
func (b *Broker) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error) {
	restartConfig := &RestartConfig{
		Enabled:     true,
		MaxRestarts: maxRestarts,
		Command:     command,
		Args:        args,
		IsRPC:       true,
	}
	return b.spawnInternal(id, command, args, restartConfig)
}

// spawnInternal handles the common logic for spawning processes.
func (b *Broker) spawnInternal(id, command string, args []string, restartConfig *RestartConfig) (*ProcessInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.processes[id]; exists {
		return nil, fmt.Errorf("process with id '%s' already exists", id)
	}

	ctx, cancel := context.WithCancel(b.ctx)
	cmd := exec.CommandContext(ctx, command, args...)

	isRPC := restartConfig != nil && restartConfig.IsRPC

	setProcessEnv(cmd, id)
	// Do not inject PROCESS_ID or RPC-related environment variables. Processes
	// compute their own IDs and transport paths deterministically from
	// build-time defaults (ldflags). This avoids runtime env coordination.

	info := &ProcessInfo{ID: id, Command: command, Args: args, Status: ProcessStatusRunning, StartTime: time.Now()}

	proc := &Process{
		info:   info,
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
		ready:  make(chan struct{}),
	}

	if restartConfig != nil {
		// Ensure restart config has correct command/args if not set
		if restartConfig.Command == "" {
			restartConfig.Command = command
		}
		if len(restartConfig.Args) == 0 {
			restartConfig.Args = args
		}
		proc.restartConfig = restartConfig
	}

	// If this is an RPC process and transport manager exists, register a UDS
	// transport before starting the process. The socket path is deterministic
	// and based on the build-time UDS base path so the subprocess can compute
	// it without runtime environment variables.
	// This is a hard failure - if UDS registration fails, we don't spawn the process.
	var udsRegistered bool
	if isRPC && b.transportManager != nil {
		if _, err := b.transportManager.RegisterUDSTransport(id, true); err != nil {
			cancel()
			info.Status = ProcessStatusFailed
			return info, fmt.Errorf("failed to register UDS transport for process %s: %w", id, err)
		}
		b.logger.Debug("Registered UDS transport for process %s before start", id)
		udsRegistered = true
	}

	// Ensure UDS transport is cleaned up if spawning fails
	if udsRegistered {
		defer func() {
			if info.Status == ProcessStatusFailed {
				b.logger.Warn("Cleaning up UDS transport for failed process %s", id)
				b.transportManager.UnregisterTransport(id)
			}
		}()
	}

	if err := cmd.Start(); err != nil {
		cancel()
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	// Create a copy of the process info to return, before starting goroutines that might modify it
	infoCopy := *info

	if isRPC {
		b.logger.Info("Spawned RPC process: id=%s pid=%d cmd=%s", id, info.PID, command)
		// For UDS transport, start a receive loop to read messages from the UDS transport
		udsTransport, err := b.transportManager.GetTransport(id)
		if err != nil {
			b.logger.Warn("Failed to get UDS transport for process %s: %v", id, err)
		} else {
			b.wg.Add(1)
			go b.readUDSMessages(id, udsTransport)
			b.logger.Debug("Started UDS message reading for process %s", id)
		}
	} else {
		b.logger.Info("Spawned process: id=%s pid=%d cmd=%s", id, info.PID, command)
	}

	b.wg.Add(1)
	go b.reapProcess(proc)

	// Wait for subprocess_ready event with timeout for RPC processes
	// This ensures the subprocess has initialized and registered its handlers
	// before we consider it fully running and start routing messages to it
	if isRPC {
		const readyTimeout = 5 * time.Second
		select {
		case <-proc.ready:
			b.logger.Debug("Process %s is ready and accepting messages", id)
		case <-time.After(readyTimeout):
			// Timeout - subprocess didn't send ready event in time
			// Don't kill the process, but log a warning
			b.logger.Warn("Process %s did not send ready event within %v, may not be fully initialized", id, readyTimeout)
		case <-b.ctx.Done():
			// Broker context cancelled during spawn
			info.Status = ProcessStatusFailed
			return &infoCopy, fmt.Errorf("broker context cancelled while waiting for process %s to be ready", id)
		}
	}

	return &infoCopy, nil
}

// LoadProcessesFromConfig loads and starts processes based on new binary detection logic.
func (b *Broker) LoadProcessesFromConfig(config interface{}) error {
	// Use build-time defaults since runtime config is disabled
	b.logger.Info("Using build-time defaults for process loading")
	// Default to detecting binaries
	return b.loadProcessesByDetection(true, []string{"access", "remote", "local", "meta", "sysmon"})
}

// loadProcessesByDetection implements the core logic for loading processes based on detection settings.
func (b *Broker) loadProcessesByDetection(detectBins bool, bootBins []string) error {
	if detectBins {
		// Detect binaries in the same directory as the broker executable
		return b.loadDetectedBinaries()
	} else {
		// Use the provided list of binaries
		return b.loadSpecifiedBinaries(bootBins)
	}
}

// loadDetectedBinaries detects executables in the same directory as the broker.
func (b *Broker) loadDetectedBinaries() error {
	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		b.logger.Error("Failed to get executable path: %v", err)
		return err
	}

	execDir := filepath.Dir(execPath)
	b.logger.Info("Detecting binaries in directory: %s", execDir)

	// Read directory
	entries, err := os.ReadDir(execDir)
	if err != nil {
		b.logger.Error("Failed to read directory %s: %v", execDir, err)
		return err
	}

	// Predefined list of expected service names (v2* prefixed)
	expectedServices := map[string]bool{
		"v2access":  true,
		"v2remote":  true,
		"v2local":   true,
		"v2meta":    true,
		"v2sysmon":  true,
		"v2analysis": true,
	}

	// Collect executables to spawn
	var services []serviceToSpawn

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		fileName := entry.Name()
		// Check if this file is one of our expected services
		if expectedServices[fileName] {
			// Check if it's executable
			filePath := filepath.Join(execDir, fileName)
			if b.isExecutable(filePath) {
				b.logger.Info("Detected executable: %s", fileName)
				services = append(services, serviceToSpawn{name: fileName, path: filePath})
			}
		}
	}

	// Spawn all services in parallel, then wait for ready events
	return b.spawnServicesParallel(services)
}

// loadSpecifiedBinaries loads the specified binaries from the list.
func (b *Broker) loadSpecifiedBinaries(binNames []string) error {
	b.logger.Info("Loading specified binaries: %v", binNames)

	for _, binName := range binNames {
		if binName == "" {
			continue // Skip empty names
		}

		if err := b.startService(binName); err != nil {
			b.logger.Warn("Failed to start service %s: %v", binName, err)
			// Continue with other services even if one fails
		}
	}

	return nil
}

// isExecutable checks if a file is executable.
func (b *Broker) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if the file mode indicates it's executable
	mode := info.Mode()
	return mode&0111 != 0 // Check if any execute bit is set (owner, group, or other)
}

// startService starts a service by name with RPC capability.
func (b *Broker) startService(serviceName string) error {
	// Start the service with RPC and auto-restart
	// Default to unlimited restarts (-1)
	info, err := b.SpawnRPCWithRestart(serviceName, "./"+serviceName, -1)
	if err != nil {
		return err
	}

	b.logger.Info("Started service %s (PID: %d) with RPC and auto-restart", info.ID, info.PID)
	return nil
}

// spawnServicesParallel spawns multiple services in parallel and waits for all ready events.
// This significantly speeds up startup compared to sequential spawning.
func (b *Broker) spawnServicesParallel(services []serviceToSpawn) error {
	type spawnResult struct {
		name string
		info *ProcessInfo
		proc *Process
		err  error
	}

	const readyTimeout = 5 * time.Second
	deadline := time.Now().Add(readyTimeout * time.Duration(len(services)))

	// Phase 1: Spawn all processes in parallel
	resultChan := make(chan spawnResult, len(services))

	for _, svc := range services {
		go func(s serviceToSpawn) {
			// Normalize binary name to process ID (v2access -> access)
			// The subprocess uses this normalized ID for its internal ProcessID
			processID := normalizeServiceName(s.name)

			b.mu.Lock()
			if _, exists := b.processes[processID]; exists {
				b.mu.Unlock()
				resultChan <- spawnResult{name: processID, err: fmt.Errorf("process with id '%s' already exists", processID)}
				return
			}
			b.mu.Unlock()

			ctx, cancel := context.WithCancel(b.ctx)
			cmd := exec.CommandContext(ctx, "./"+s.name)
			setProcessEnv(cmd, processID)

			info := &ProcessInfo{ID: processID, Command: "./" + s.name, Status: ProcessStatusRunning, StartTime: time.Now()}
			proc := &Process{
				info:          info,
				cmd:           cmd,
				cancel:        cancel,
				done:          make(chan struct{}),
				ready:         make(chan struct{}),
				restartConfig: &RestartConfig{Enabled: true, MaxRestarts: -1, Command: "./" + s.name, IsRPC: true},
			}

			// Register UDS transport before starting using the process ID (not binary name)
			if b.transportManager != nil {
				if _, err := b.transportManager.RegisterUDSTransport(processID, true); err != nil {
					cancel()
					info.Status = ProcessStatusFailed
					resultChan <- spawnResult{name: processID, err: fmt.Errorf("failed to register UDS transport: %w", err)}
					return
				}
			}

			if err := cmd.Start(); err != nil {
				cancel()
				if b.transportManager != nil {
					b.transportManager.UnregisterTransport(processID)
				}
				info.Status = ProcessStatusFailed
				resultChan <- spawnResult{name: processID, err: fmt.Errorf("failed to start process: %w", err)}
				return
			}

			info.PID = cmd.Process.Pid
			b.mu.Lock()
			b.processes[processID] = proc
			b.mu.Unlock()

			// Start UDS message reading goroutine
			if b.transportManager != nil {
				udsTransport, err := b.transportManager.GetTransport(processID)
				if err == nil {
					b.wg.Add(1)
					go b.readUDSMessages(processID, udsTransport)
				}
			}

			b.wg.Add(1)
			go b.reapProcess(proc)

			resultChan <- spawnResult{name: processID, info: info, proc: proc, err: nil}
		}(svc)
	}

	// Collect spawn results
	var results []spawnResult
	startedServices := make(map[string]bool)
	var procs []*Process

	for range services {
		result := <-resultChan
		results = append(results, result)
		if result.err != nil {
			b.logger.Warn("Failed to start service %s: %v", result.name, result.err)
		} else {
			b.logger.Info("Spawned RPC process: id=%s pid=%d cmd=%s", result.name, result.info.PID, result.name)
			startedServices[result.name] = true
			procs = append(procs, result.proc)
		}
	}

	// Phase 2: Wait for all ready events in parallel
	var readyWg sync.WaitGroup
	for _, result := range results {
		if result.err != nil || result.proc == nil {
			continue
		}

		readyWg.Add(1)
		go func(p *Process, name string) {
			defer readyWg.Done()
			select {
			case <-p.ready:
				b.logger.Debug("Process %s is ready and accepting messages", name)
			case <-time.After(time.Until(deadline)):
				b.logger.Warn("Process %s did not send ready event before deadline", name)
			case <-b.ctx.Done():
				return
			}
		}(result.proc, result.name)
	}

	// Wait for all processes to be ready before proceeding
	readyWg.Wait()

	// Report what we found
	b.logger.Info("Started services: %v", startedServices)
	return nil
}

// setProcessEnv configures environment variables for a process based on its ID and build-time config.
// This consolidates the repeated env setup logic from Spawn, SpawnRPC, SpawnWithRestart, and SpawnRPCWithRestart.
func setProcessEnv(cmd *exec.Cmd, processID string) {
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	switch processID {
	case "local":
		cmd.Env = append(cmd.Env, fmt.Sprintf("CVE_DB_PATH=%s", cve.DefaultBuildCVEDBPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_DB_PATH=%s", cwe.DefaultBuildCWEDBPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_DB_PATH=%s", capec.DefaultBuildCAPECDBPath()))
	case "meta":
		cmd.Env = append(cmd.Env, fmt.Sprintf("SESSION_DB_PATH=%s", meta.DefaultBuildSessionDBPath()))
	case "remote":
		// NVD_API_KEY is no longer supported
	case "access":
		// Note: Static dir is now build-time config, so broker doesn't override it with runtime config
		// The access service will use its build-time static dir
		// No runtime config override is applied
	}
}
