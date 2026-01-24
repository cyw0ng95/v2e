package transport

import (
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// Transport defines the interface for communication protocols
type Transport interface {
	Send(msg *proc.Message) error
	Receive() (*proc.Message, error)
	Connect() error
	Close() error
}