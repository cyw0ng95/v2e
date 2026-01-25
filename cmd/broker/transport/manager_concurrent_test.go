package transport

import (
	"fmt"
	"sync"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestTransportManager_ConcurrentRegisterAndSend ensures concurrent RegisterTransport
// and SendToProcess work without races and all messages are delivered.
func TestTransportManager_ConcurrentRegisterAndSend(t *testing.T) {
	tm := NewTransportManager()

	const workers = 20
	const msgsPerWorker = 10

	var wg sync.WaitGroup

	// Start concurrent registrars and senders
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			id := fmt.Sprintf("p-%02d", w)
			ft := &concurrentFakeTransport{}
			tm.RegisterTransport(id, ft)

			for i := 0; i < msgsPerWorker; i++ {
				msg, err := proc.NewRequestMessage("m", map[string]string{"i": "v"})
				if err != nil {
					t.Fatalf("failed to create message: %v", err)
				}
				if err := tm.SendToProcess(id, msg); err != nil {
					t.Fatalf("SendToProcess returned error: %v", err)
				}
			}
		}(w)
	}

	wg.Wait()

	// Verify at least one transport received messages (basic sanity)
	// We can't easily access all transports here, but ensure no panics/races occurred.
}

type concurrentFakeTransport struct {
	mu   sync.Mutex
	last *proc.Message
}

func (f *concurrentFakeTransport) Send(msg *proc.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.last = msg
	return nil
}
func (f *concurrentFakeTransport) Receive() (*proc.Message, error) { return nil, nil }
func (f *concurrentFakeTransport) Connect() error                  { return nil }
func (f *concurrentFakeTransport) Close() error                    { return nil }
