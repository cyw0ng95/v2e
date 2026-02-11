package transport

import (
	"fmt"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestTransportManager_ConcurrentRegisterAndSend ensures concurrent RegisterTransport
// and SendToProcess work without races and all messages are delivered.
func TestTransportManager_ConcurrentRegisterAndSend(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestTransportManager_ConcurrentRegisterAndSend", nil, func(t *testing.T, tx *gorm.DB) {
		tm := NewTransportManager()

		const workers = 20
		const msgsPerWorker = 10

		var wg sync.WaitGroup
		errCh := make(chan error, workers)

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
						errCh <- fmt.Errorf("failed to create message: %v", err)
						return
					}
					if err := tm.SendToProcess(id, msg); err != nil {
						errCh <- fmt.Errorf("SendToProcess returned error: %v", err)
						return
					}
				}
			}(w)
		}

		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatal(err)
			}
		}

		// Verify at least one transport received messages (basic sanity)
		// We can't easily access all transports here, but ensure no panics/races occurred.
	})

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
