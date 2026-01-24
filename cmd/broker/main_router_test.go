package main

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// stubCoreBroker captures calls for brokerRouter delegation tests.
type stubCoreBroker struct {
	routeCalled   bool
	processCalled bool
	lastMsg       *proc.Message
	lastSource    string
}

func (s *stubCoreBroker) RouteMessage(msg *proc.Message, sourceProcess string) error {
	s.routeCalled = true
	s.lastMsg = msg
	s.lastSource = sourceProcess
	return nil
}

func (s *stubCoreBroker) ProcessMessage(msg *proc.Message) error {
	s.processCalled = true
	s.lastMsg = msg
	return nil
}

func TestBrokerRouterDelegates(t *testing.T) {
	stub := &stubCoreBroker{}
	r := &brokerRouter{b: stub}
	msg := &proc.Message{ID: "123", Target: "broker"}

	if err := r.Route(msg, "source-proc"); err != nil {
		t.Fatalf("Route returned error: %v", err)
	}
	if !stub.routeCalled || stub.lastSource != "source-proc" || stub.lastMsg != msg {
		t.Fatalf("RouteMessage not invoked as expected: %+v", stub)
	}

	if err := r.ProcessBrokerMessage(msg); err != nil {
		t.Fatalf("ProcessBrokerMessage returned error: %v", err)
	}
	if !stub.processCalled || stub.lastMsg != msg {
		t.Fatalf("ProcessMessage not invoked as expected: %+v", stub)
	}
}
