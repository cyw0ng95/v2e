package main

import (
	"bufio"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestInvokeRPC_SuccessRoundTrip(t *testing.T) {
	b := NewBroker()

	// Create a pipe: broker will write to p.stdin (writer), test reads from reader
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer pr.Close()
	defer pw.Close()

	// Insert fake target process with stdin = pw
	InsertFakeProcess(b, "target-proc", pw, nil, ProcessStatusRunning)

	// Start InvokeRPC in background
	resultCh := make(chan *proc.Message, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, err := b.InvokeRPC("src-proc", "target-proc", "DoThing", map[string]interface{}{"n": 1}, 3*time.Second)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- resp
	}()

	// Read the outgoing request from the pipe
	reader := bufio.NewReader(pr)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("failed to read request from pipe: %v", err)
	}

	reqMsg, err := proc.Unmarshal(line[:len(line)-1])
	if err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	// Craft response with same correlation id
	resp, _ := proc.NewResponseMessage(reqMsg.ID, map[string]interface{}{"ok": true})
	resp.CorrelationID = reqMsg.CorrelationID
	resp.Source = "target-proc"
	resp.Target = "src-proc"

	// Route response into broker (as if from target process)
	if err := b.RouteMessage(resp, "target-proc"); err != nil {
		t.Fatalf("failed to route response: %v", err)
	}

	// Await result
	select {
	case r := <-resultCh:
		if r.Type != proc.MessageTypeResponse {
			t.Fatalf("expected response message, got type %s", r.Type)
		}
		if r.CorrelationID != reqMsg.CorrelationID {
			t.Fatalf("mismatched correlation ids: want %s got %s", reqMsg.CorrelationID, r.CorrelationID)
		}
	case e := <-errCh:
		t.Fatalf("InvokeRPC returned error: %v", e)
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for InvokeRPC result")
	}
}
