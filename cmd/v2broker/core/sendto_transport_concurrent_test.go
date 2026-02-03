package core

import (
	"sync"
	"testing"

	"github.com/cyw0ng95/v2e/cmd/v2broker/transport"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

type recordingTransport struct {
	mu   sync.Mutex
	msgs []*proc.Message
}

func (r *recordingTransport) Send(msg *proc.Message) error {
	r.mu.Lock()
	r.msgs = append(r.msgs, msg)
	r.mu.Unlock()
	return nil
}
func (r *recordingTransport) Receive() (*proc.Message, error) { return nil, nil }
func (r *recordingTransport) Connect() error                  { return nil }
func (r *recordingTransport) Close() error                    { return nil }

func TestBroker_SendToProcess_Concurrent(t *testing.T) {
	b := NewBroker()
	tm := transport.NewTransportManager()
	rt := &recordingTransport{}
	tm.RegisterTransport("p-conc", rt)
	b.transportManager = tm

	const N = 50
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			msg, err := proc.NewRequestMessage("m", map[string]string{"n": "v"})
			if err != nil {
				t.Errorf("failed to create message: %v", err)
				return
			}
			if err := b.SendToProcess("p-conc", msg); err != nil {
				t.Errorf("SendToProcess returned error: %v", err)
				return
			}
		}(i)
	}
	wg.Wait()

	// verify all messages delivered
	rt.mu.Lock()
	got := len(rt.msgs)
	rt.mu.Unlock()
	if got != N {
		t.Fatalf("expected %d messages, got %d", N, got)
	}
}
