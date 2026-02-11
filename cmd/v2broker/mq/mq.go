package mq

import (
	"context"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// MessageBus abstracts broker message send/receive and stats snapshot.
type MessageBus interface {
	Send(ctx context.Context, msg *proc.Message) error
	Receive(ctx context.Context) (*proc.Message, error)
	BufferCap() int
}
