package core

import (
	"bufio"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// readProcessMessages reads messages from a process's stdout and forwards them to the broker.
func (b *Broker) readProcessMessages(p *Process) {
	defer p.readLoopWg.Done()
	defer b.wg.Done()

	scanner := bufio.NewScanner(p.stdout)
	buf := make([]byte, proc.MaxMessageSize)
	scanner.Buffer(buf, proc.MaxMessageSize)

	for {
		// Check if process is done before each scan
		select {
		case <-p.done:
			b.logger.Debug("Process %s done, stopping message reading", p.info.ID)
			return
		default:
		}

		if !scanner.Scan() {
			// Scan failed - check if process exited
			select {
			case <-p.done:
				b.logger.Debug("Process %s done, scanner.Scan returned false", p.info.ID)
			default:
				// Scanner error without process being done
				if err := scanner.Err(); err != nil {
					b.logger.Warn("Error reading from process %s: %v", p.info.ID, err)
				}
			}
			return
		}

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
						b.logger.Debug("Failed to send error response to %s: %v", msg.Source, sendErr)
					}
				}()
			}
		}
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

// readUDSMessages reads messages from a UDS transport and forwards them to the broker's router.
// This is used instead of readProcessMessages when UDS transport is enabled for a subprocess.
func (b *Broker) readUDSMessages(processID string, transport transport.Transport) {
	defer b.wg.Done()

	b.logger.Debug("Starting UDS message reading for process %s", processID)

	for {
		// Check if process still exists and is running
		b.mu.RLock()
		p, exists := b.processes[processID]
		b.mu.RUnlock()

		if !exists || p.info.Status != ProcessStatusRunning {
			b.logger.Debug("Process %s no longer running, stopping UDS message reading", processID)
			return
		}

		// Receive message from UDS transport
		msg, err := transport.Receive()
		if err != nil {
			// Check if process exited
			b.mu.RLock()
			p, exists = b.processes[processID]
			b.mu.RUnlock()

			if !exists || p.info.Status != ProcessStatusRunning {
				b.logger.Debug("Process %s exited, stopping UDS message reading", processID)
				return
			}

			// Log error and continue
			b.logger.Warn("Error receiving message from UDS transport for process %s: %v", processID, err)
			// Don't return immediately - might be a temporary error
			// But sleep a bit to avoid tight error loop
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Empty message - might be a heartbeat or keepalive
		if msg == nil {
			continue
		}

		// Route the message through the broker's router
		if err := b.RouteMessage(msg, processID); err != nil {
			b.logger.Warn("Failed to route message from UDS transport for process %s: %v", processID, err)

			// Send error response if this was a request
			if msg.Type == proc.MessageTypeRequest && msg.CorrelationID != "" {
				errorMsg := &proc.Message{
					Type:          proc.MessageTypeError,
					ID:            msg.ID,
					Error:         err.Error(),
					Target:        msg.Source,
					CorrelationID: msg.CorrelationID,
				}
				_ = b.SendToProcess(msg.Source, errorMsg)
			}
		}
	}
}
