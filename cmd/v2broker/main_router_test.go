package main

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// stubCoreBroker captures calls for broker routing tests.
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

func (s *stubCoreBroker) ProcessBrokerMessage(msg *proc.Message) error {
	s.processCalled = true
	s.lastMsg = msg
	return nil
}

func TestBrokerRoutingSatisfiesRouterInterface(t *testing.T) {
	stub := &stubCoreBroker{}
	msg := &proc.Message{ID: "123", Target: "broker"}

	// Test Route method
	if err := stub.RouteMessage(msg, "source-proc"); err != nil {
		t.Fatalf("RouteMessage returned error: %v", err)
	}
	if !stub.routeCalled || stub.lastSource != "source-proc" || stub.lastMsg != msg {
		t.Fatalf("RouteMessage not invoked as expected: %+v", stub)
	}

	// Reset and test ProcessBrokerMessage
	stub.routeCalled = false
	stub.processCalled = false

	if err := stub.ProcessBrokerMessage(msg); err != nil {
		t.Fatalf("ProcessBrokerMessage returned error: %v", err)
	}
	if !stub.processCalled || stub.lastMsg != msg {
		t.Fatalf("ProcessBrokerMessage not invoked as expected: %+v", stub)
	}
}
