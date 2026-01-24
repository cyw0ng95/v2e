package core

import (
	"testing"

	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

type successTransport struct {
	last *proc.Message
}

func (s *successTransport) Send(msg *proc.Message) error {
	s.last = msg
	return nil
}

func (s *successTransport) Receive() (*proc.Message, error) { return nil, nil }
func (s *successTransport) Connect() error                  { return nil }
func (s *successTransport) Close() error                    { return nil }

func TestBroker_SendToProcess_UsesTransportWhenAvailable(t *testing.T) {
	b := NewBroker()
	// use a transport manager with a fake transport
	tm := transport.NewTransportManager()
	st := &successTransport{}
	tm.RegisterTransport("p-1", st)
	b.transportManager = tm

	msg, err := proc.NewRequestMessage("hello", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if err := b.SendToProcess("p-1", msg); err != nil {
		t.Fatalf("SendToProcess returned error: %v", err)
	}

	if st.last == nil || st.last.ID != msg.ID {
		t.Fatalf("transport did not receive message; got=%v want id=%s", st.last, msg.ID)
	}
}
