package proc

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewBaseProcess(t *testing.T) {
	proc := NewBaseProcess("test-proc")
	if proc == nil {
		t.Fatal("NewBaseProcess returned nil")
	}
	if proc.ID() != "test-proc" {
		t.Errorf("Expected ID to be 'test-proc', got '%s'", proc.ID())
	}
}

func TestBaseProcess_Start(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if proc.broker == nil {
		t.Error("Expected broker to be set")
	}
	if proc.ctx == nil {
		t.Error("Expected context to be set")
	}
	if proc.cancel == nil {
		t.Error("Expected cancel function to be set")
	}
}

func TestBaseProcess_Stop(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = proc.stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Context should be cancelled after stop
	select {
	case <-proc.Context().Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected context to be cancelled after stop")
	}
}

func TestBaseProcess_OnMessage_NotImplemented(t *testing.T) {
	proc := NewBaseProcess("test-proc")
	msg, _ := NewRequestMessage("req-1", nil)

	err := proc.OnMessage(msg)
	if err == nil {
		t.Error("Expected error for unimplemented OnMessage")
	}
}

func TestBaseProcess_SendMessage(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	msg, _ := NewRequestMessage("req-1", map[string]string{"test": "data"})
	err = proc.SendMessage(msg)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Verify message was sent to broker
	recvCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	received, err := broker.ReceiveMessage(recvCtx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	if received.ID != msg.ID {
		t.Errorf("Expected message ID to be '%s', got '%s'", msg.ID, received.ID)
	}
}

func TestBaseProcess_SendMessage_NotInitialized(t *testing.T) {
	proc := NewBaseProcess("test-proc")
	msg, _ := NewRequestMessage("req-1", nil)

	err := proc.SendMessage(msg)
	if err == nil {
		t.Error("Expected error when sending message without initialization")
	}
}

func TestBaseProcess_SendRequest(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	payload := map[string]string{"key": "value"}
	err = proc.SendRequest("req-1", payload)
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}

	// Verify request was sent
	recvCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	received, err := broker.ReceiveMessage(recvCtx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	if received.Type != MessageTypeRequest {
		t.Errorf("Expected MessageTypeRequest, got %s", received.Type)
	}
}

func TestBaseProcess_SendResponse(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	payload := map[string]interface{}{"status": "ok"}
	err = proc.SendResponse("resp-1", payload)
	if err != nil {
		t.Fatalf("SendResponse failed: %v", err)
	}

	// Verify response was sent
	recvCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	received, err := broker.ReceiveMessage(recvCtx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	if received.Type != MessageTypeResponse {
		t.Errorf("Expected MessageTypeResponse, got %s", received.Type)
	}
}

func TestBaseProcess_SendEvent(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	payload := map[string]string{"event": "test_event"}
	err = proc.SendEvent("evt-1", payload)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// Verify event was sent
	recvCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	received, err := broker.ReceiveMessage(recvCtx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	if received.Type != MessageTypeEvent {
		t.Errorf("Expected MessageTypeEvent, got %s", received.Type)
	}
}

func TestBaseProcess_SendError(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	testErr := fmt.Errorf("test error")
	err = proc.SendError("err-1", testErr)
	if err != nil {
		t.Fatalf("SendError failed: %v", err)
	}

	// Verify error message was sent
	recvCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	received, err := broker.ReceiveMessage(recvCtx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	if received.Type != MessageTypeError {
		t.Errorf("Expected MessageTypeError, got %s", received.Type)
	}
	if received.Error != "test error" {
		t.Errorf("Expected error to be 'test error', got '%s'", received.Error)
	}
}

func TestBaseProcess_Context(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	procCtx := proc.Context()
	if procCtx == nil {
		t.Error("Expected context to be non-nil")
	}
}

func TestBaseProcess_Broker(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewBaseProcess("test-proc")
	ctx := context.Background()

	err := proc.start(ctx, broker)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	procBroker := proc.Broker()
	if procBroker == nil {
		t.Error("Expected broker to be non-nil")
	}
	if procBroker != broker {
		t.Error("Expected broker to match the one passed to Start")
	}
}

func TestTestProcess_Lifecycle(t *testing.T) {
	broker := NewBroker()
	defer broker.Shutdown()

	proc := NewTestProcess("test-proc")

	// Test initialization via broker registration
	err := broker.RegisterManagedProcess(proc)
	if err != nil {
		t.Fatalf("RegisterManagedProcess failed: %v", err)
	}
	
	if proc.Broker() == nil {
		t.Error("Expected Broker to be set")
	}
	if proc.Context() == nil {
		t.Error("Expected Context to be set")
	}

	// Test OnMessage
	msg, _ := NewRequestMessage("req-1", nil)
	err = proc.OnMessage(msg)
	if err != nil {
		t.Fatalf("OnMessage failed: %v", err)
	}
	if proc.messageCount != 1 {
		t.Errorf("Expected message count to be 1, got %d", proc.messageCount)
	}
	if proc.lastMessage != msg {
		t.Error("Expected lastMessage to match sent message")
	}

	// Test Stop via broker
	err = broker.StopManagedProcess("test-proc")
	if err != nil {
		t.Fatalf("StopManagedProcess failed: %v", err)
	}
	
	// Verify context was cancelled
	select {
	case <-proc.Context().Done():
		// Expected
	default:
		t.Error("Expected context to be cancelled after stop")
	}
}

func TestTestProcess_OnMessage_Error(t *testing.T) {
	proc := NewTestProcess("test-proc")
	proc.shouldFailMsg = true

	msg, _ := NewRequestMessage("req-1", nil)
	err := proc.OnMessage(msg)
	if err == nil {
		t.Error("Expected error from OnMessage")
	}
}

func TestManagedProcess_Interface(t *testing.T) {
	// Verify that TestProcess implements ManagedProcess interface
	var _ ManagedProcess = (*TestProcess)(nil)
	var _ ManagedProcess = (*BaseProcess)(nil)
}
