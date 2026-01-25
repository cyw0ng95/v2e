package mq

import (
	"context"
	"sync"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestBusConcurrentSendReceive stresses the Bus with multiple senders and receivers
// to ensure it handles concurrent operations and updates stats correctly.
func TestBusConcurrentSendReceive(t *testing.T) {
	const senders = 10
	const perSender = 100

	bus := NewBus(context.Background(), 32)

	var wg sync.WaitGroup
	// start receivers
	received := int64(0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < senders*perSender; i++ {
			if _, err := bus.Receive(context.Background()); err != nil {
				t.Fatalf("receive error: %v", err)
			}
			received++
		}
	}()

	// start senders
	var sw sync.WaitGroup
	for s := 0; s < senders; s++ {
		sw.Add(1)
		go func(si int) {
			defer sw.Done()
			for j := 0; j < perSender; j++ {
				msg := &proc.Message{Type: proc.MessageTypeRequest, ID: "id", Source: "s", Target: "t"}
				if err := bus.Send(context.Background(), msg); err != nil {
					t.Fatalf("send error: %v", err)
				}
			}
		}(s)
	}

	sw.Wait()
	wg.Wait()

	stats := bus.GetMessageStats()
	if stats.TotalSent != int64(senders*perSender) {
		t.Fatalf("unexpected total sent: %d", stats.TotalSent)
	}
	if stats.TotalReceived != int64(senders*perSender) {
		t.Fatalf("unexpected total received: %d", stats.TotalReceived)
	}
}
