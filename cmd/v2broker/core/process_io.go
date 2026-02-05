package core

import (
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/cmd/v2broker/metrics"
	"github.com/cyw0ng95/v2e/cmd/v2broker/transport"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// MetricsEncodingUnknown is used for JSON messages in the metrics registry
const MetricsEncodingUnknown = metrics.EncodingUnknown

// SendToProcess sends a message to a specific process via UDS transport.
func (b *Broker) SendToProcess(processID string, msg *proc.Message) error {
	// Use transport for message delivery
	if b.transportManager != nil {
		// Marshal message to get wire size
		data, err := msg.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		
		wireSize := len(data)
		
		err = b.transportManager.SendToProcess(processID, msg)
		if err == nil {
			b.bus.Record(msg, true)
			// Record in metrics registry with wire size
			// For JSON messages, encoding is JSON; binary messages would be detected elsewhere
			b.metricsRegistry.RecordMessage(msg, true, wireSize, MetricsEncodingUnknown)
			b.logger.Debug("Sent message to process %s via transport: type=%s id=%s size=%d", processID, msg.Type, msg.ID, wireSize)
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

			// Check if process has sent ready event yet
			p.mu.RLock()
			readyClosed := false
			select {
			case <-p.ready:
				readyClosed = true
			default:
			}
			p.mu.RUnlock()

			// If process hasn't signaled ready yet, transport not connected is expected
			// Log as debug instead of warn during initial startup
			if !readyClosed && err.Error() == "transport not connected" {
				b.logger.Debug("Process %s UDS transport not yet connected (waiting for subprocess_ready)", processID)
			} else {
				// "token too long" indicates message exceeded max buffer size - this is a critical error
				if err.Error() == "failed to scan message: bufio.Scanner: token too long" {
					b.logger.Error("Message size exceeded maximum buffer for process %s: %v", processID, err)
				} else {
					b.logger.Warn("Error receiving message from UDS transport for process %s: %v", processID, err)
				}
			}
			// Don't return immediately - might be a temporary error
			// But sleep a bit to avoid tight error loop
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Empty message - might be a heartbeat or keepalive
		if msg == nil {
			continue
		}

		// Record received message in metrics (approximate size for now)
		// In the future, this could be enhanced to use actual wire size from transport
		data, _ := msg.Marshal()
		wireSize := len(data)
		if wireSize > 0 {
			b.metricsRegistry.RecordMessage(msg, false, wireSize, MetricsEncodingUnknown)
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
