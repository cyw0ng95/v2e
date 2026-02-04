package subprocess

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentMessageSending tests sending messages from multiple goroutines
func TestConcurrentMessageSending(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentMessageSending", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		const numGoroutines = 10
		const messagesPerGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Send messages concurrently from multiple goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					msg := &Message{
						Type: MessageTypeEvent,
						ID:   fmt.Sprintf("goroutine-%d-msg-%d", id, j),
					}
					if err := sp.SendMessage(msg); err != nil {
						t.Errorf("Failed to send message from goroutine %d: %v", id, err)
					}
				}
			}(i)
		}

		wg.Wait()
		sp.Flush()

		// Verify all messages were sent
		_ = output.String() // Just ensure no panic when reading
		t.Logf("Successfully sent %d messages concurrently", numGoroutines*messagesPerGoroutine)
	})

}

// TestConcurrentHandlerRegistration tests registering handlers concurrently
func TestConcurrentHandlerRegistration(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentHandlerRegistration", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")

		const numGoroutines = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Register handlers concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				pattern := fmt.Sprintf("pattern-%d", id)
				handler := func(ctx context.Context, msg *Message) (*Message, error) {
					return &Message{Type: MessageTypeResponse, ID: msg.ID}, nil
				}
				sp.RegisterHandler(pattern, handler)
			}(i)
		}

		wg.Wait()

		// Verify all handlers were registered
		sp.mu.RLock()
		count := len(sp.handlers)
		sp.mu.RUnlock()

		if count != numGoroutines {
			t.Errorf("Expected %d handlers, got %d", numGoroutines, count)
		}
	})

}

// TestConcurrentHandlerExecution tests executing handlers concurrently
func TestConcurrentHandlerExecution(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentHandlerExecution", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		const numRequests = 50
		handlerCallCount := 0
		var mu sync.Mutex

		// Register handler that tracks calls
		sp.RegisterHandler("request", func(ctx context.Context, msg *Message) (*Message, error) {
			mu.Lock()
			handlerCallCount++
			mu.Unlock()

			// Simulate some work
			time.Sleep(1 * time.Millisecond)

			return &Message{
				Type: MessageTypeResponse,
				ID:   msg.ID,
			}, nil
		})

		var wg sync.WaitGroup
		wg.Add(numRequests)

		// Send requests concurrently
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				defer wg.Done()
				msg := &Message{
					Type: MessageTypeRequest,
					ID:   fmt.Sprintf("req-%d", id),
				}
				// Note: Direct access to wg and handleMessage is necessary for testing
				// concurrent handler execution without running the full subprocess
				sp.wg.Add(1)
				sp.handleMessage(msg)
			}(i)
		}

		wg.Wait()
		sp.wg.Wait()

		mu.Lock()
		finalCount := handlerCallCount
		mu.Unlock()

		if finalCount != numRequests {
			t.Errorf("Expected handler to be called %d times, got %d", numRequests, finalCount)
		}
	})

}

// TestConcurrentSendAndReceive tests sending and receiving messages concurrently
func TestConcurrentSendAndReceive(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentSendAndReceive", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")

		// Use a thread-safe buffer for output
		var outputMu sync.Mutex
		outputBuf := &bytes.Buffer{}
		safeOutput := &struct {
			*bytes.Buffer
			mu *sync.Mutex
		}{outputBuf, &outputMu}

		sp.output = safeOutput

		const numMessages = 100

		var sendWg sync.WaitGroup
		sendWg.Add(numMessages)

		// Start sender goroutines
		for i := 0; i < numMessages; i++ {
			go func(id int) {
				defer sendWg.Done()
				msg := &Message{
					Type: MessageTypeEvent,
					ID:   fmt.Sprintf("event-%d", id),
				}
				_ = sp.SendMessage(msg)
			}(i)
		}

		sendWg.Wait()
		sp.Flush()

		// Verify output was written (no panic is success)
		outputMu.Lock()
		_ = outputBuf.String()
		outputMu.Unlock()

		t.Logf("Successfully handled concurrent send operations")
	})

}

// TestConcurrentStopAndSend tests stopping while messages are being sent
func TestConcurrentStopAndSend(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentStopAndSend", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		// Start message writer
		sp.wg.Add(1)
		go sp.messageWriter()

		const numSenders = 5
		var wg sync.WaitGroup
		wg.Add(numSenders)

		// Send messages continuously
		stopSending := make(chan bool)
		for i := 0; i < numSenders; i++ {
			go func(id int) {
				defer wg.Done()
				for {
					select {
					case <-stopSending:
						return
					default:
						msg := &Message{
							Type: MessageTypeEvent,
							ID:   fmt.Sprintf("msg-%d", id),
						}
						_ = sp.SendMessage(msg)
						time.Sleep(1 * time.Millisecond)
					}
				}
			}(i)
		}

		// Let messages be sent for a bit
		time.Sleep(50 * time.Millisecond)

		// Stop the subprocess while messages are being sent
		_ = sp.Stop()
		close(stopSending)

		wg.Wait()

		// Should not panic
		t.Log("Successfully stopped subprocess during concurrent message sending")
	})

}

// TestConcurrentFlush tests calling Flush from multiple goroutines
func TestConcurrentFlush(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentFlush", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		const numFlushers = 10

		var wg sync.WaitGroup
		wg.Add(numFlushers)

		// Flush concurrently
		for i := 0; i < numFlushers; i++ {
			go func() {
				defer wg.Done()
				sp.Flush()
			}()
		}

		wg.Wait()

		// Should not panic
		t.Log("Successfully flushed concurrently")
	})

}

// TestRaceOnHandlerMap tests for race conditions on handler map access
func TestRaceOnHandlerMap(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRaceOnHandlerMap", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")

		const numOps = 100
		var wg sync.WaitGroup
		wg.Add(numOps * 2)

		// Concurrent reads and writes to handler map
		for i := 0; i < numOps; i++ {
			// Writer
			go func(id int) {
				defer wg.Done()
				pattern := fmt.Sprintf("handler-%d", id%10)
				handler := func(ctx context.Context, msg *Message) (*Message, error) {
					return nil, nil
				}
				sp.RegisterHandler(pattern, handler)
			}(i)

			// Reader (via message handling)
			go func(id int) {
				defer wg.Done()
				msg := &Message{
					Type: MessageType(fmt.Sprintf("handler-%d", id%10)),
					ID:   fmt.Sprintf("msg-%d", id),
				}
				sp.wg.Add(1)
				sp.handleMessage(msg)
			}(i)
		}

		wg.Wait()
		sp.wg.Wait()

		// Should not panic or race
		t.Log("No race conditions detected on handler map")
	})

}

// TestConcurrentContextCancellation tests context cancellation during concurrent operations
func TestConcurrentContextCancellation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConcurrentContextCancellation", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		const numWorkers = 10
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		// Start workers that check context
		for i := 0; i < numWorkers; i++ {
			go func(id int) {
				defer wg.Done()
				for {
					select {
					case <-sp.ctx.Done():
						return
					default:
						msg := &Message{
							Type: MessageTypeEvent,
							ID:   fmt.Sprintf("worker-%d", id),
						}
						_ = sp.SendMessage(msg)
						time.Sleep(1 * time.Millisecond)
					}
				}
			}(i)
		}

		// Let workers run for a bit
		time.Sleep(20 * time.Millisecond)

		// Cancel context
		_ = sp.Stop()

		// Wait for workers to stop
		wg.Wait()

		// Should not deadlock
		t.Log("Successfully handled concurrent context cancellation")
	})

}

// TestHighConcurrencyMessageProcessing tests message processing under high concurrency
func TestHighConcurrencyMessageProcessing(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestHighConcurrencyMessageProcessing", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		output := &bytes.Buffer{}
		sp.SetOutput(output)

		const numWorkers = 50
		const messagesPerWorker = 20

		// Register a handler
		sp.RegisterHandler("work", func(ctx context.Context, msg *Message) (*Message, error) {
			// Simulate work
			time.Sleep(1 * time.Millisecond)
			return &Message{Type: MessageTypeResponse, ID: msg.ID}, nil
		})

		var wg sync.WaitGroup
		wg.Add(numWorkers)

		start := time.Now()

		// Process messages concurrently
		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < messagesPerWorker; j++ {
					msg := &Message{
						Type: "work",
						ID:   fmt.Sprintf("w%d-m%d", workerID, j),
					}
					sp.wg.Add(1)
					sp.handleMessage(msg)
				}
			}(i)
		}

		wg.Wait()
		sp.wg.Wait()

		elapsed := time.Since(start)
		totalMessages := numWorkers * messagesPerWorker

		t.Logf("Processed %d messages in %v (%.2f msg/s)",
			totalMessages, elapsed, float64(totalMessages)/elapsed.Seconds())
	})

}
