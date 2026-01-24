package core

import (
	"bufio"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// readProcessMessages reads messages from a process's stdout and forwards them to the broker.
func (b *Broker) readProcessMessages(p *Process) {
	defer b.wg.Done()

	scanner := bufio.NewScanner(p.stdout)
	buf := make([]byte, proc.MaxMessageSize)
	scanner.Buffer(buf, proc.MaxMessageSize)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		msg, err := proc.Unmarshal([]byte(line))
		if err != nil {
			b.logger.Warn("Failed to parse message from process %s: %v", p.info.ID, err)
			continue
		}

		// Prefer asynchronous routing via optimizer if configured.
		routedViaOptimizer := false
		b.mu.RLock()
		opt := b.optimizer
		b.mu.RUnlock()
		if opt != nil {
			if accepted := opt.Offer(msg); accepted {
				// optimizer accepted the message for async routing
				routedViaOptimizer = true
			}
		}

		if routedViaOptimizer {
			continue
		}

		if err := b.RouteMessage(msg, p.info.ID); err != nil {
			b.logger.Warn("Failed to route message from process %s: %v", p.info.ID, err)

			if msg.Type == proc.MessageTypeRequest && msg.CorrelationID != "" {
				errorMsg := &proc.Message{
					Type:          proc.MessageTypeError,
					ID:            msg.ID,
					Error:         err.Error(),
					Target:        msg.Source,
					CorrelationID: msg.CorrelationID,
				}
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

// SendToProcess sends a message to a specific process via stdin.
func (b *Broker) SendToProcess(processID string, msg *proc.Message) error {
	// First try to use transport if available
	if b.transportManager != nil {
		if err := b.transportManager.SendToProcess(processID, msg); err == nil {
			b.bus.Record(msg, true)
			b.logger.Debug("Sent message to process %s via transport: type=%s id=%s", processID, msg.Type, msg.ID)
			return nil
		}
		// If transport fails, fall back to the original stdin method
	}

	// Fallback to original stdin-based implementation for backward compatibility
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

	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := fmt.Fprintf(stdin, "%s\n", string(data)); err != nil {
		return fmt.Errorf("failed to write message to process: %w", err)
	}

	b.bus.Record(msg, true)

	b.logger.Debug("Sent message to process %s: type=%s id=%s", processID, msg.Type, msg.ID)
	return nil
}

// reapProcess waits for a process to complete and updates its status.
// If auto-restart is enabled, it will attempt to restart the process according to the configured policy.
func (b *Broker) reapProcess(p *Process) {
	defer b.wg.Done()
	defer close(p.done)

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

	b.logger.Info("Process exited: id=%s pid=%d exit_code=%d", p.info.ID, p.info.PID, p.info.ExitCode)

	// Unregister transport for the process if it exists
	if b.transportManager != nil {
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
	b.sendMessageInternal(event)

	if p.restartConfig != nil && p.restartConfig.Enabled {
		if p.restartConfig.MaxRestarts >= 0 && p.restartConfig.RestartCount >= p.restartConfig.MaxRestarts {
			b.logger.Warn("Process %s exceeded max restarts (%d), not restarting", p.info.ID, p.restartConfig.MaxRestarts)
			p.mu.Unlock()
			return
		}

		p.restartConfig.RestartCount++

		processID := p.info.ID
		command := p.restartConfig.Command
		args := p.restartConfig.Args
		isRPC := p.restartConfig.IsRPC
		maxRestarts := p.restartConfig.MaxRestarts
		restartCount := p.restartConfig.RestartCount

		p.mu.Unlock()

		b.logger.Info("Restarting process %s (attempt %d/%d)", processID, restartCount, maxRestarts)

		b.mu.Lock()
		delete(b.processes, processID)
		b.mu.Unlock()

		time.Sleep(1 * time.Second)

		var restartErr error
		if isRPC {
			_, restartErr = b.SpawnRPCWithRestart(processID, command, maxRestarts, args...)
		} else {
			_, restartErr = b.SpawnWithRestart(processID, command, maxRestarts, args...)
		}

		if restartErr != nil {
			b.logger.Warn("Failed to restart process %s: %v", processID, restartErr)
		} else {
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

// Kill terminates a process by ID.
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

	if proc.cmd.Process != nil {
		if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			proc.cancel()
		}
	}

	select {
	case <-proc.done:
		b.logger.Info("Process terminated gracefully: id=%s", id)
		return nil
	case <-time.After(5 * time.Second):
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

// GetProcess returns information about a process by ID.
func (b *Broker) GetProcess(id string) (*ProcessInfo, error) {
	b.mu.RLock()
	proc, exists := b.processes[id]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process with id '%s' not found", id)
	}

	proc.mu.RLock()
	defer proc.mu.RUnlock()

	info := *proc.info
	return &info, nil
}

// ListProcesses returns information about all managed processes.
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

// Shutdown gracefully shuts down the broker and all managed processes.
func (b *Broker) Shutdown() error {
	b.logger.Info("Shutting down broker")

	b.cancel()

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

	b.wg.Wait()

	b.bus.Close()

	// Clean up transport manager
	if b.transportManager != nil {
		// Close all transports
		// Note: In a real implementation, we'd iterate through all transports and close them
		// For now, we just let the process cleanup handle it
	}

	b.logger.Info("Broker shutdown complete")
	return nil
}
