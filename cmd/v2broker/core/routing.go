package core

import (
	"context"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// MessageChannel exposes the broker message channel for read-only consumers (tests).
func (b *Broker) MessageChannel() <-chan *proc.Message {
	return b.messages
}

// SendMessage sends a message to the broker's message channel.
func (b *Broker) SendMessage(msg *proc.Message) error {
	return b.bus.Send(b.ctx, msg)
}

// sendMessageInternal sends a message internally (from broker processes) without error handling.
func (b *Broker) sendMessageInternal(msg *proc.Message) {
	b.bus.SendInternal(msg)
}

// ReceiveMessage receives a message from the broker's message channel.
// It blocks until a message is available or the context is cancelled.
func (b *Broker) ReceiveMessage(ctx context.Context) (*proc.Message, error) {
	return b.bus.Receive(ctx)
}

// RouteMessage routes a message to its target process or handles it locally.
// Satisfies routing.Router interface via Route() alias below.
func (b *Broker) RouteMessage(msg *proc.Message, sourceProcess string) error {
	if msg.Source == "" {
		msg.Source = sourceProcess
	}

	if msg.Type == proc.MessageTypeResponse && msg.CorrelationID != "" {
		b.logger.Debug("Received response message: id=%s correlation_id=%s from=%s", msg.ID, msg.CorrelationID, msg.Source)
		// Use atomic load-and-delete operation to reduce lock contention
		b.pendingMu.Lock()
		pending, exists := b.pendingRequests[msg.CorrelationID]
		if exists {
			delete(b.pendingRequests, msg.CorrelationID)
		}
		b.pendingMu.Unlock()

		if exists {
			b.logger.Debug("Routing response to pending request: correlation_id=%s", msg.CorrelationID)
			select {
			case pending.ResponseChan <- msg:
				b.logger.Debug("Response delivered to waiting channel: correlation_id=%s", msg.CorrelationID)
				return nil
			case <-time.After(5 * time.Second):
				b.logger.Warn("Timeout sending response to pending request: correlation_id=%s", msg.CorrelationID)
				return fmt.Errorf("timeout sending response to pending request")
			}
		}
		b.logger.Debug("No pending request found for correlation_id=%s (may be tracked by subprocess), trying target-based routing", msg.CorrelationID)
		// For responses, ensure the source is set to the responding process
		if msg.Type == proc.MessageTypeResponse && msg.Source == "" {
			msg.Source = sourceProcess
		}
	}

	if msg.Target != "" {
		if msg.Target == "broker" {
			b.logger.Debug("Routing message to broker for local processing: type=%s id=%s from=%s", msg.Type, msg.ID, msg.Source)
			return b.ProcessMessage(msg)
		}

		b.logger.Debug("Routing message from %s to %s: type=%s id=%s", msg.Source, msg.Target, msg.Type, msg.ID)
		return b.SendToProcess(msg.Target, msg)
	}

	return b.SendMessage(msg)
}

// Route satisfies the routing.Router interface.
// It routes a message to its target process or handles it locally.
// Delegates to RouteMessage to maintain backward compatibility.
func (b *Broker) Route(msg *proc.Message, sourceProcess string) error {
	return b.RouteMessage(msg, sourceProcess)
}

// ProcessMessage processes a message directed at the broker.
func (b *Broker) ProcessMessage(msg *proc.Message) error {
	if msg.Type != proc.MessageTypeRequest {
		return nil
	}

	var respMsg *proc.Message
	var err error

	switch msg.ID {
	case "RPCGetMessageStats":
		respMsg, err = b.HandleRPCGetMessageStats(msg)
	case "RPCGetMessageCount":
		respMsg, err = b.HandleRPCGetMessageCount(msg)
	case "RPCRequestPermits":
		respMsg, err = b.HandleRPCRequestPermits(msg)
	case "RPCReleasePermits":
		respMsg, err = b.HandleRPCReleasePermits(msg)
	case "RPCGetKernelMetrics":
		respMsg, err = b.HandleRPCGetKernelMetrics(msg)
	default:
		errMsg := proc.NewErrorMessage(msg.ID, fmt.Errorf("unknown RPC method: %s", msg.ID))
		errMsg.Source = "broker"
		errMsg.Target = msg.Source
		if msg.CorrelationID != "" {
			errMsg.CorrelationID = msg.CorrelationID
		}
		return b.RouteMessage(errMsg, "broker")
	}

	if err != nil {
		errMsg := proc.NewErrorMessage(msg.ID, err)
		errMsg.Source = "broker"
		errMsg.Target = msg.Source
		if msg.CorrelationID != "" {
			errMsg.CorrelationID = msg.CorrelationID
		}
		return b.RouteMessage(errMsg, "broker")
	}

	return b.RouteMessage(respMsg, "broker")
}

// ProcessBrokerMessage processes a message directed at the broker.
// Satisfies routing.Router interface. Delegates to ProcessMessage.
func (b *Broker) ProcessBrokerMessage(msg *proc.Message) error {
	return b.ProcessMessage(msg)
}
