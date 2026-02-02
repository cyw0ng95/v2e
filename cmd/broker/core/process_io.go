package core

import (
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// SendToProcess sends a message to a specific process via UDS transport.
func (b *Broker) SendToProcess(processID string, msg *proc.Message) error {
	// Use transport for message delivery
	if b.transportManager != nil {
		err := b.transportManager.SendToProcess(processID, msg)
		if err == nil {
			b.bus.Record(msg, true)
			b.logger.Debug("Sent message to process %s via transport: type=%s id=%s", processID, msg.Type, msg.ID)
			return nil
		}
		return fmt.Errorf("failed to send message to process %s via transport: %w", processID, err)
	}

	return fmt.Errorf("no transport manager available")
}

// readUDSMessages reads messages from a UDS transport and forwards them to the broker's router.
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

		// Handle subprocess_ready event - close the ready channel to signal
		// that the subprocess has initialized and registered its handlers
		if msg.Type == proc.MessageTypeEvent && msg.ID == "subprocess_ready" {
			b.mu.RLock()
			p, exists := b.processes[processID]
			b.mu.RUnlock()

			if exists {
				p.mu.Lock()
				if p.ready != nil {
					select {
					case <-p.ready:
						// Already closed
					default:
						close(p.ready)
						b.logger.Debug("Process %s signaled ready", processID)
					}
				}
				p.mu.Unlock()
			}
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
