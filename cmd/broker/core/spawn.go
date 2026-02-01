package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/meta"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

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

	var inputFD, outputFD int
	var readFromSubprocess, writeToSubprocess *os.File
	var readFromParent, writeToParent *os.File
	isRPC := restartConfig != nil && restartConfig.IsRPC

	if isRPC {
		var err error
		// Create output pipe: subprocess writes to writeToParent, parent reads from readFromSubprocess
		readFromSubprocess, writeToParent, err = os.Pipe()
		if err != nil {
			cancel()
			b.logger.Error("Failed to create output pipe for process %s: %v", id, err)
			return nil, fmt.Errorf("failed to create output pipe: %w", err)
		}

		// Create input pipe: parent writes to writeToSubprocess, subprocess reads from readFromParent
		readFromParent, writeToSubprocess, err = os.Pipe()
		if err != nil {
			cancel()
			readFromSubprocess.Close()
			writeToParent.Close()
			b.logger.Error("Failed to create input pipe for process %s: %v", id, err)
			return nil, fmt.Errorf("failed to create input pipe: %w", err)
		}

		cmd.ExtraFiles = []*os.File{readFromParent, writeToParent}
		inputFD, outputFD = b.getRPCFileDescriptors()
		cmd.Env = append(cmd.Env, "BROKER_PASSING_RPC_FDS=1")
	}

	setProcessEnv(cmd, id, nil)
	if isRPC {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PROCESS_ID=%s", id))
		// MaxMessageSize is now configured at build-time, no runtime override
	}

	info := &ProcessInfo{ID: id, Command: command, Args: args, Status: ProcessStatusRunning, StartTime: time.Now()}

	proc := &Process{
		info:   info,
		cmd:    cmd,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	if isRPC {
		proc.stdin = writeToSubprocess
		proc.stdout = readFromSubprocess
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

	if err := cmd.Start(); err != nil {
		cancel()
		if isRPC {
			readFromSubprocess.Close()
			writeToSubprocess.Close()
			readFromParent.Close()
			writeToParent.Close()
		}
		info.Status = ProcessStatusFailed
		return info, fmt.Errorf("failed to start process: %w", err)
	}

	if isRPC {
		// Close parent's copy of the pipe ends used by the child
		readFromParent.Close()
		writeToParent.Close()
	}

	info.PID = cmd.Process.Pid
	b.processes[id] = proc

	// Create a copy of the process info to return, before starting goroutines that might modify it
	infoCopy := *info

	if isRPC {
		b.logger.Info("Spawned RPC process: id=%s pid=%d command=%s (advertised fds=%d,%d)", id, info.PID, command, inputFD, outputFD)
		b.registerProcessTransport(id, inputFD, outputFD)
		b.wg.Add(1)
		go b.readProcessMessages(proc)
	} else {
		b.logger.Info("Spawned process: id=%s pid=%d command=%s", id, info.PID, command)
	}

	b.wg.Add(1)
	go b.reapProcess(proc)

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

	// Predefined list of expected service names
	expectedServices := map[string]bool{
		"access": true,
		"remote": true,
		"local":  true,
		"meta":   true,
		"sysmon": true,
	}

	// Track which services we've started
	startedServices := make(map[string]bool)

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
				if err := b.startService(fileName, true); err != nil {
					b.logger.Warn("Failed to start service %s: %v", fileName, err)
				} else {
					startedServices[fileName] = true
				}
			}
		}
	}

	// Report what we found
	b.logger.Info("Binary detection complete. Started services: %v", startedServices)
	return nil
}

// loadSpecifiedBinaries loads the specified binaries from the list.
func (b *Broker) loadSpecifiedBinaries(binNames []string) error {
	b.logger.Info("Loading specified binaries: %v", binNames)

	for _, binName := range binNames {
		if binName == "" {
			continue // Skip empty names
		}

		if err := b.startService(binName, true); err != nil {
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
func (b *Broker) startService(serviceName string, withRPC bool) error {
	// Start the service with RPC and auto-restart
	// Default to unlimited restarts (-1)
	info, err := b.SpawnRPCWithRestart(serviceName, "./"+serviceName, -1)
	if err != nil {
		return err
	}

	b.logger.Info("Started service %s (PID: %d) with RPC and auto-restart", info.ID, info.PID)
	return nil
}

// setProcessEnv configures environment variables for a process based on its ID and build-time config.
// This consolidates the repeated env setup logic from Spawn, SpawnRPC, SpawnWithRestart, and SpawnRPCWithRestart.
func setProcessEnv(cmd *exec.Cmd, processID string, config interface{}) {
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	switch processID {
	case "local":
		cmd.Env = append(cmd.Env, fmt.Sprintf("CVE_DB_PATH=%s", cve.DefaultBuildCVEDBPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_DB_PATH=%s", cwe.DefaultBuildCWEDBPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_DB_PATH=%s", capec.DefaultBuildCAPECDBPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CWE_RAW_PATH=%s", cwe.DefaultBuildCWERawPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_XML_PATH=%s", capec.DefaultBuildCAPECXMLPath()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CAPEC_XSD_PATH=%s", capec.DefaultBuildCAPECXSDPath()))
		if capec.DefaultBuildXSDValidation() {
			cmd.Env = append(cmd.Env, "CAPEC_STRICT_XSD=1")
		}
	case "meta":
		cmd.Env = append(cmd.Env, fmt.Sprintf("SESSION_DB_PATH=%s", meta.DefaultBuildSessionDBPath()))
	case "remote":
		// NVD_API_KEY is no longer supported
		cmd.Env = append(cmd.Env, fmt.Sprintf("VIEW_FETCH_URL=%s", cwe.DefaultBuildViewURL()))
	case "access":
		// Note: Static dir is now build-time config, so broker doesn't override it with runtime config
		// The access service will use its build-time static dir
		// No runtime config override is applied
	}
}

// getRPCFileDescriptors returns the configured input and output file descriptor numbers for RPC communication.
// Uses build-time configuration as default since runtime config is disabled.
func (b *Broker) getRPCFileDescriptors() (inputFD, outputFD int) {
	// Use build-time defaults since runtime config is disabled
	inputFD = subprocess.DefaultBuildRPCInputFD()
	outputFD = subprocess.DefaultBuildRPCOutputFD()
	return
}

// registerProcessTransport creates and registers the appropriate transport for a spawned RPC process.
// Returns an error only if transport registration fails critically.
func (b *Broker) registerProcessTransport(processID string, inputFD, outputFD int) {
	if b.transportManager == nil {
		return
	}
	if shouldUseUDSTransport(struct{UseUDS bool}{UseUDS: false}) {
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
		b.logger.Error("Failed to connect FD transport for process %s: %v", processID, err)
	}
}

// shouldUseUDSTransport determines whether UDS transport should be used based on the transport configuration
func shouldUseUDSTransport(config struct{UseUDS bool}) bool {
	// Always use FD transport instead of UDS transport
	return false
}
