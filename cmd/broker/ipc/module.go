// Package ipc implements inter-process communication for the broker
package ipc

import (
	"context"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// MessageRouter defines the interface for routing messages between processes
type MessageRouter interface {
	RouteMessage(msg *proc.Message, sourceProcess string) error
	ProcessMessage(msg *proc.Message) error
	GetMessageStats() MessageStats
	GetPerProcessStats() map[string]PerProcessStats
}

// MessageQueue defines the interface for queuing messages
type MessageQueue interface {
	Send(ctx context.Context, msg *proc.Message) error
	Receive(ctx context.Context) (*proc.Message, error)
	GetQueueDepth() int
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalSent        int64
	TotalReceived    int64
	RequestCount     int64
	ResponseCount    int64
	EventCount       int64
	ErrorCount       int64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}

// PerProcessStats represents per-process message statistics
type PerProcessStats struct {
	TotalSent        int64
	TotalReceived    int64
	RequestCount     int64
	ResponseCount    int64
	EventCount       int64
	ErrorCount       int64
	FirstMessageTime time.Time
	LastMessageTime  time.Time
}