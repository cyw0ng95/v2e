package mq

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestBusConcurrentSendReceive stresses the Bus with multiple senders and receivers
// to ensure it handles concurrent operations and updates stats correctly.
func TestBusConcurrentSendReceive(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBusConcurrentSendReceive", nil, func(t *testing.T, tx *gorm.DB) {
		const senders = 10
		const perSender = 100

		bus := NewBus(context.Background(), 32)

		var wg sync.WaitGroup
		// start receivers
		received := int64(0)
		recvErrCh := make(chan error, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < senders*perSender; i++ {
				if _, err := bus.Receive(context.Background()); err != nil {
					recvErrCh <- fmt.Errorf("receive error: %v", err)
					return
				}
				received++
			}
		}()

		// start senders
		var sw sync.WaitGroup
		sendErrCh := make(chan error, senders)
		for s := 0; s < senders; s++ {
			sw.Add(1)
			go func(si int) {
				defer sw.Done()
				for j := 0; j < perSender; j++ {
					msg := &proc.Message{Type: proc.MessageTypeRequest, ID: "id", Source: "s", Target: "t"}
					if err := bus.Send(context.Background(), msg); err != nil {
						sendErrCh <- fmt.Errorf("send error: %v", err)
						return
					}
				}
			}(s)
		}

		// Check for errors from goroutines
		// Wait for senders first
		sw.Wait()
		close(sendErrCh)
		if err := <-sendErrCh; err != nil {
			t.Fatal(err)
		}

		// Then wait for receiver
		close(recvErrCh)
		wg.Wait()
		if err := <-recvErrCh; err != nil {
			t.Fatal(err)
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
	})

}
