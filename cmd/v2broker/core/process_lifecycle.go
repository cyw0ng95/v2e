package core

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// reapProcess waits for a process to complete and updates its status.
// If auto-restart is enabled, it will attempt to restart the process according to the configured policy.
func (b *Broker) reapProcess(p *Process) {
	defer b.wg.Done()
	// Note: p.done is closed explicitly later in this function, not deferred,
	// because we need to close it before calling restartProcess for restart-enabled processes

	err := p.cmd.Wait()

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

	if p.info.ExitCode != 0 {
		if p.restartConfig != nil && p.restartConfig.Enabled {
			b.logger.Warn("Process exited abnormally (restart scheduled): id=%s pid=%d code=%d", p.info.ID, p.info.PID, p.info.ExitCode)
		} else {
			b.logger.Error("Process exited abnormally (no restart): id=%s pid=%d code=%d", p.info.ID, p.info.PID, p.info.ExitCode)
		}
	} else {
		b.logger.Info("Process exited successfully: id=%s pid=%d code=%d", p.info.ID, p.info.PID, p.info.ExitCode)
	}

	// Only close transport if process is NOT configured for restart.
	// For restart-enabled processes, we keep the UDS transport alive across
	// restarts to avoid socket close errors.
	shouldCloseTransport := true
	if p.restartConfig != nil && p.restartConfig.Enabled {
		shouldCloseTransport = false
		b.logger.Debug("Keeping transport alive for restart-enabled process %s", p.info.ID)
	}

	if shouldCloseTransport && b.transportManager != nil {
		b.transportManager.UnregisterTransport(p.info.ID)
		b.logger.Debug("Unregistered transport for process %s", p.info.ID)
	}

	event, _ := proc.NewEventMessage(p.info.ID, map[string]interface{}{
		"event":     "process_exited",
		"id":        p.info.ID,
		"pid":       p.info.PID,
		"exit_code": p.info.ExitCode,
	})
	event.Target = "test-target"

	// Release lock before sending message to avoid deadlock if channel is full
	p.mu.Unlock()
	b.sendMessageInternal(event)

	// Handle restart if configured
	if p.restartConfig != nil && p.restartConfig.Enabled {
		// Close the done channel before restart so that readUDSMessages can exit cleanly
		// We close it here instead of relying on defer because we're about to call restartProcess
		// which will spawn a new process with its own done channel
		close(p.done)
		b.restartProcess(p)
		return
	}

	// For non-restart processes, close done channel to signal goroutines
	close(p.done)
}

// restartProcess attempts to restart a process according to its restart configuration.
func (b *Broker) restartProcess(p *Process) {
	p.mu.Lock()

	// Check if max restarts exceeded
	if p.restartConfig.MaxRestarts >= 0 && p.restartConfig.RestartCount >= p.restartConfig.MaxRestarts {
		b.logger.Warn("Max restarts exceeded for process %s (%d), stopping restarts", p.info.ID, p.restartConfig.MaxRestarts)
		p.mu.Unlock()
		return
	}

	p.restartConfig.RestartCount++

	// Capture restart config before unlocking
	processID := p.info.ID
	command := p.restartConfig.Command
	args := p.restartConfig.Args
	isRPC := p.restartConfig.IsRPC
	maxRestarts := p.restartConfig.MaxRestarts
	restartCount := p.restartConfig.RestartCount

	p.mu.Unlock()

	b.logger.Info("Restarting process %s: attempt %d/%d", processID, restartCount, maxRestarts)

	// Delete old process from map
	b.processes.Delete(processID)

	// Delay before restart
	time.Sleep(1 * time.Second)

	// Spawn new process
	var restartErr error
	if isRPC {
		_, restartErr = b.SpawnRPCWithRestart(processID, command, maxRestarts, args...)
	} else {
		_, restartErr = b.SpawnWithRestart(processID, command, maxRestarts, args...)
	}

	if restartErr != nil {
		b.logger.Warn("Failed to restart process %s: %v", processID, restartErr)
		return
	}

	// Copy restart count to new process
	value, exists := b.processes.Load(processID)
	if exists {
		newProc := value.(*Process)
		newProc.mu.Lock()
		if newProc.restartConfig != nil {
			newProc.restartConfig.RestartCount = restartCount
		}
		newProc.mu.Unlock()
	}
}

// Kill terminates a process by ID.
func (b *Broker) Kill(id string) error {
	value, exists := b.processes.Load(id)
	if !exists {
		return fmt.Errorf("process with id '%s' not found", id)
	}
	proc := value.(*Process)

	proc.mu.RLock()
	status := proc.info.Status
	proc.mu.RUnlock()

	if status != ProcessStatusRunning {
		return fmt.Errorf("process '%s' is not running", id)
	}

	if proc.cmd == nil {
		// Test process or non-OS-backed process: mark as exited and return
		proc.mu.Lock()
		proc.info.Status = ProcessStatusExited
		proc.mu.Unlock()
		return nil
	}

	if proc.cmd != nil && proc.cmd.Process != nil {
		if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			if proc.cancel != nil {
				proc.cancel()
			}
		}
	}

	select {
	case <-proc.done:
		b.logger.Info("Process terminated gracefully: id=%s", id)
		return nil
	case <-time.After(5 * time.Second):
		b.logger.Warn("Force killing process %s (graceful termination failed)", id)
		if proc.cmd != nil && proc.cmd.Process != nil {
			if err := proc.cmd.Process.Kill(); err != nil {
				b.logger.Error("Failed to force kill process %s: %v", id, err)
				return fmt.Errorf("failed to force kill process: %w", err)
			}
		}
		<-proc.done
		b.logger.Warn("Process force killed: id=%s", id)
		return nil
	}
}

// GetProcess returns information about a process by ID.
func (b *Broker) GetProcess(id string) (*ProcessInfo, error) {
	value, exists := b.processes.Load(id)
	if !exists {
		return nil, fmt.Errorf("process with id '%s' not found", id)
	}
	proc := value.(*Process)

	proc.mu.RLock()
	defer proc.mu.RUnlock()

	info := *proc.info
	return &info, nil
}

// ListProcesses returns information about all managed processes.
func (b *Broker) ListProcesses() []*ProcessInfo {
	result := make([]*ProcessInfo, 0)
	b.processes.Range(func(key, value interface{}) bool {
		proc := value.(*Process)
		proc.mu.RLock()
		info := *proc.info
		proc.mu.RUnlock()
		result = append(result, &info)
		return true
	})
	return result
}

// Shutdown gracefully shuts down the broker and all managed processes.
func (b *Broker) Shutdown() error {
	b.logger.Info("Shutting down broker")

	b.cancel()

	processIDs := make([]string, 0)
	b.processes.Range(func(key, value interface{}) bool {
		processIDs = append(processIDs, key.(string))
		return true
	})

	for _, id := range processIDs {
		value, exists := b.processes.Load(id)
		if exists {
			proc := value.(*Process)
			status := proc.info.Status
			proc.mu.RUnlock()

			if status == ProcessStatusRunning {
				_ = b.Kill(id)
			}
		}
	}

	b.wg.Wait()

	b.bus.Close()

	// Clean up transport manager - close all transports including those
	// kept alive for process restart
	if b.transportManager != nil {
		b.transportManager.CloseAll()
	}

	b.logger.Info("Broker shutdown complete")
	return nil
}
