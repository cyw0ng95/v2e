package core

import (
	"bufio"
	"fmt"

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
