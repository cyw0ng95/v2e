package mq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// MessageStats contains aggregated message statistics.
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

// PerProcessStats contains per-process message statistics.
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

// Bus implements a buffered message bus with statistics tracking.
type Bus struct {
	ch              chan *proc.Message
	stats           MessageStats
	perProcessStats map[string]PerProcessStats
	mu              sync.RWMutex
	ctx             context.Context
}

// NewBus creates a new bus with the provided buffer size and lifecycle context.
func NewBus(ctx context.Context, buffer int) *Bus {
	return &Bus{
		ch:              make(chan *proc.Message, buffer),
		perProcessStats: make(map[string]PerProcessStats),
		ctx:             ctx,
	}
}

// Channel exposes the underlying channel for consumers that select directly.
func (b *Bus) Channel() chan *proc.Message { return b.ch }

// BufferCap returns the channel capacity.
func (b *Bus) BufferCap() int { return cap(b.ch) }

// Send enqueues a message and updates stats based on direction inferred from target.
func (b *Bus) Send(ctx context.Context, msg *proc.Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("message channel is closed")
		}
	}()
	select {
	case b.ch <- msg:
		b.updateStats(msg, isSent(msg))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-b.ctx.Done():
		return b.ctx.Err()
	}
}

// SendInternal enqueues without returning errors (best-effort), used from internal paths.
func (b *Bus) SendInternal(msg *proc.Message) {
	defer func() { recover() }()
	select {
	case b.ch <- msg:
		b.updateStats(msg, isSent(msg))
	case <-b.ctx.Done():
	}
}

// Receive dequeues a message, updating stats as a receive.
func (b *Bus) Receive(ctx context.Context) (*proc.Message, error) {
	select {
	case msg, ok := <-b.ch:
		if !ok {
			return nil, context.Canceled
		}
		b.updateStats(msg, false)
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.ctx.Done():
		return nil, b.ctx.Err()
	}
}

// GetMessageStats returns a snapshot of broker-wide message statistics.
func (b *Bus) GetMessageStats() MessageStats {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stats
}

// GetPerProcessStats returns a copy of per-process stats.
func (b *Bus) GetPerProcessStats() map[string]PerProcessStats {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make(map[string]PerProcessStats, len(b.perProcessStats))
	for k, v := range b.perProcessStats {
		out[k] = v
	}
	return out
}

// GetMessageCount returns total sent + received.
func (b *Bus) GetMessageCount() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stats.TotalSent + b.stats.TotalReceived
}

// Record allows external callers to update stats without enqueuing.
func (b *Bus) Record(msg *proc.Message, isSent bool) {
	b.updateStats(msg, isSent)
}

// Close closes the underlying channel.
func (b *Bus) Close() { close(b.ch) }

// isSent infers direction: true if broker is sending to a target, false if received.
func isSent(msg *proc.Message) bool {
	return !(msg.Target == "" || msg.Target == "broker")
}

// updateStats updates message statistics.
func (b *Bus) updateStats(msg *proc.Message, isSent bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	if b.stats.FirstMessageTime.IsZero() {
		b.stats.FirstMessageTime = now
	}
	b.stats.LastMessageTime = now

	if isSent {
		b.stats.TotalSent++
	} else {
		b.stats.TotalReceived++
	}

	switch msg.Type {
	case proc.MessageTypeRequest:
		b.stats.RequestCount++
	case proc.MessageTypeResponse:
		b.stats.ResponseCount++
	case proc.MessageTypeEvent:
		b.stats.EventCount++
	case proc.MessageTypeError:
		b.stats.ErrorCount++
	}

	var procID string
	if isSent {
		procID = msg.Target
	} else {
		procID = msg.Source
	}

	if procID != "" {
		ps := b.perProcessStats[procID]
		if ps.FirstMessageTime.IsZero() {
			ps.FirstMessageTime = now
		}
		ps.LastMessageTime = now

		if isSent {
			ps.TotalSent++
		} else {
			ps.TotalReceived++
		}

		switch msg.Type {
		case proc.MessageTypeRequest:
			ps.RequestCount++
		case proc.MessageTypeResponse:
			ps.ResponseCount++
		case proc.MessageTypeEvent:
			ps.EventCount++
		case proc.MessageTypeError:
			ps.ErrorCount++
		}

		b.perProcessStats[procID] = ps
	}
}
