package proc

import (
	"context"
	"fmt"
)

// ManagedProcess represents a process that can be controlled by the broker
// and interact with it through message passing.
// Users only need to implement OnMessage to handle business logic.
type ManagedProcess interface {
	// OnMessage handles messages received from the broker
	// This is where business logic should be implemented
	OnMessage(msg *Message) error

	// ID returns the unique identifier of this process
	ID() string
}

// managedProcessInternal is an internal interface for lifecycle management
type managedProcessInternal interface {
	start(ctx context.Context, broker *Broker) error
	stop() error
}

// BaseProcess provides a complete implementation of ManagedProcess
// with convenient helper methods for sending messages.
// Users can embed this to get full functionality with minimal boilerplate.
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

// OnMessage is a default no-op message handler
// Users should override this to implement their business logic
func (p *BaseProcess) OnMessage(msg *Message) error {
	return fmt.Errorf("message handling not implemented")
}

// ID returns the process identifier
func (p *BaseProcess) ID() string {
	return p.id
}

// start initializes the base process (internal method)
func (p *BaseProcess) start(ctx context.Context, broker *Broker) error {
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.broker = broker
	return nil
}

// stop gracefully stops the base process (internal method)
func (p *BaseProcess) stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
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
