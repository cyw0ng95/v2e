package transport

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

type fakeTransport struct {
	last *proc.Message
}

func (f *fakeTransport) Send(msg *proc.Message) error {
	f.last = msg
	return nil
}

func (f *fakeTransport) Receive() (*proc.Message, error) { return nil, nil }
func (f *fakeTransport) Connect() error                  { return nil }
func (f *fakeTransport) Close() error                    { return nil }

func TestTransportManager_RegisterAndSend(t *testing.T) {
	tm := NewTransportManager()
	ft := &fakeTransport{}
	tm.RegisterTransport("p1", ft)

	msg, err := proc.NewRequestMessage("test", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if err := tm.SendToProcess("p1", msg); err != nil {
		t.Fatalf("SendToProcess returned error: %v", err)
	}

	if ft.last == nil || ft.last.ID != msg.ID {
		t.Fatalf("transport did not receive message; got=%v want id=%s", ft.last, msg.ID)
	}
}
