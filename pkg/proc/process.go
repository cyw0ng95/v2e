package proc

import (
	"context"
	"fmt"
)

// ManagedProcess represents a process that can be controlled by the broker
// and interact with it through message passing.
type ManagedProcess interface {
	// Start initializes the process and begins execution
	// The broker context is passed for lifecycle management
	Start(ctx context.Context, broker *Broker) error

	// Stop gracefully stops the process
	Stop() error

	// OnMessage handles messages received from the broker
	OnMessage(msg *Message) error

	// ID returns the unique identifier of this process
	ID() string
}

// BaseProcess provides a default implementation of common ManagedProcess functionality
type BaseProcess struct {
	id     string
	broker *Broker
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBaseProcess creates a new BaseProcess with the given ID
func NewBaseProcess(id string) *BaseProcess {
	return &BaseProcess{
		id: id,
	}
}

// Start initializes the base process
func (p *BaseProcess) Start(ctx context.Context, broker *Broker) error {
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.broker = broker
	return nil
}

// Stop gracefully stops the base process
func (p *BaseProcess) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

// OnMessage is a default no-op message handler
// Subprocesses should override this to handle messages
func (p *BaseProcess) OnMessage(msg *Message) error {
	return fmt.Errorf("message handling not implemented")
}

// ID returns the process identifier
func (p *BaseProcess) ID() string {
	return p.id
}

// SendMessage sends a message to the broker
func (p *BaseProcess) SendMessage(msg *Message) error {
	if p.broker == nil {
		return fmt.Errorf("process not initialized with broker")
	}
	return p.broker.SendMessage(msg)
}

// SendRequest sends a request message to the broker
func (p *BaseProcess) SendRequest(id string, payload interface{}) error {
	msg, err := NewRequestMessage(id, payload)
	if err != nil {
		return err
	}
	return p.SendMessage(msg)
}

// SendResponse sends a response message to the broker
func (p *BaseProcess) SendResponse(id string, payload interface{}) error {
	msg, err := NewResponseMessage(id, payload)
	if err != nil {
		return err
	}
	return p.SendMessage(msg)
}

// SendEvent sends an event message to the broker
func (p *BaseProcess) SendEvent(id string, payload interface{}) error {
	msg, err := NewEventMessage(id, payload)
	if err != nil {
		return err
	}
	return p.SendMessage(msg)
}

// SendError sends an error message to the broker
func (p *BaseProcess) SendError(id string, err error) error {
	msg := NewErrorMessage(id, err)
	return p.SendMessage(msg)
}

// Context returns the process context
func (p *BaseProcess) Context() context.Context {
	return p.ctx
}

// Broker returns the broker instance
func (p *BaseProcess) Broker() *Broker {
	return p.broker
}
