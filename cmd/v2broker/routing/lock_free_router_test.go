package routing

import (
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestLockFreeRouter_ImplementsRouter verifies that LockFreeRouter implements Router interface.
func TestLockFreeRouter_ImplementsRouter(t *testing.T) {
	var _ Router = (*LockFreeRouter)(nil)
}

// TestLockFreeRouter_RouteSignatureMatchesInterface verifies the Route method signature.
func TestLockFreeRouter_RouteSignatureMatchesInterface(t *testing.T) {
	router := NewLockFreeRouter()

	// Create a test channel
	ch := make(chan *proc.Message, 10)
	router.RegisterRoute("test-target", ch)

	// Create a test message
	msg := &proc.Message{
		Type:   proc.MessageTypeRequest,
		ID:     "test-1",
		Source: "test-source",
		Target: "test-target",
	}

	// Call Route with sourceProcess parameter
	err := router.Route(msg, "test-source")
	if err != nil {
		t.Fatalf("Route failed: %v", err)
	}

	// Verify message was routed
	select {
	case routedMsg := <-ch:
		if routedMsg.ID != msg.ID {
			t.Fatalf("Message ID mismatch: want %s got %s", msg.ID, routedMsg.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Message was not routed within timeout")
	}
}

// TestLockFreeRouter_RouteNoTarget verifies error when target not registered.
func TestLockFreeRouter_RouteNoTarget(t *testing.T) {
	router := NewLockFreeRouter()

	msg := &proc.Message{
		Type:   proc.MessageTypeRequest,
		ID:     "test-1",
		Target: "non-existent-target",
	}

	// Call Route with sourceProcess parameter
	err := router.Route(msg, "test-source")
	if err == nil {
		t.Fatal("Expected error for non-existent target, got nil")
	}

	expectedErrMsg := "no route for target: non-existent-target"
	if err.Error() != expectedErrMsg {
		t.Fatalf("Error message mismatch: want %s got %s", expectedErrMsg, err.Error())
	}
}

// TestLockFreeRouter_ChannelFull verifies error when channel is full.
func TestLockFreeRouter_ChannelFull(t *testing.T) {
	router := NewLockFreeRouter()

	// Create a channel with zero capacity (will be full immediately)
	ch := make(chan *proc.Message)
	router.RegisterRoute("full-target", ch)

	msg := &proc.Message{
		Type:   proc.MessageTypeRequest,
		ID:     "test-1",
		Target: "full-target",
	}

	// First send in a goroutine to block
	go func() {
		<-ch
	}()

	// Give the goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Call Route - should succeed since channel reader is waiting
	err := router.Route(msg, "test-source")
	if err != nil {
		t.Fatalf("Route failed unexpectedly: %v", err)
	}
}

// TestLockFreeRouter_ProcessBrokerMessage verifies ProcessBrokerMessage routing.
func TestLockFreeRouter_ProcessBrokerMessage(t *testing.T) {
	router := NewLockFreeRouter()

	ch := make(chan *proc.Message, 10)
	router.RegisterRoute("broker-target", ch)

	msg := &proc.Message{
		Type:   proc.MessageTypeEvent,
		ID:     "broker-1",
		Target: "broker-target",
	}

	err := router.ProcessBrokerMessage(msg)
	if err != nil {
		t.Fatalf("ProcessBrokerMessage failed: %v", err)
	}

	select {
	case routedMsg := <-ch:
		if routedMsg.ID != msg.ID {
			t.Fatalf("Message ID mismatch: want %s got %s", msg.ID, routedMsg.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Message was not routed within timeout")
	}
}

// TestLockFreeRouter_RegisterUnregister verifies route registration and removal.
func TestLockFreeRouter_RegisterUnregister(t *testing.T) {
	router := NewLockFreeRouter()

	ch := make(chan *proc.Message, 10)
	router.RegisterRoute("target-1", ch)

	if router.GetRouteCount() != 1 {
		t.Fatalf("Expected 1 route, got %d", router.GetRouteCount())
	}

	router.UnregisterRoute("target-1")

	if router.GetRouteCount() != 0 {
		t.Fatalf("Expected 0 routes after unregister, got %d", router.GetRouteCount())
	}
}

// TestLockFreeRouter_ListRoutes verifies route listing.
func TestLockFreeRouter_ListRoutes(t *testing.T) {
	router := NewLockFreeRouter()

	ch1 := make(chan *proc.Message, 10)
	ch2 := make(chan *proc.Message, 10)
	router.RegisterRoute("target-1", ch1)
	router.RegisterRoute("target-2", ch2)

	routes := router.ListRoutes()
	if len(routes) != 2 {
		t.Fatalf("Expected 2 routes, got %d", len(routes))
	}
}
