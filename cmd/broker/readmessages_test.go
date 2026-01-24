package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestReadProcessMessages_ParsesAndRoutes(t *testing.T) {
	b := NewBroker()

	// Create an os.Pipe to simulate process stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Insert fake process with stdout set to read end
	InsertFakeProcess(b, "fake-proc", nil, r, ProcessStatusRunning)

	// Start reader goroutine (mirror how SpawnRPC runs it)
	b.wg.Add(1)
	go b.readProcessMessages(b.processes["fake-proc"])

	// Craft a message and write it to the write end
	msg, _ := proc.NewEventMessage("evt-1", map[string]interface{}{"x": 42})
	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	if _, err := w.Write(append(data, '\n')); err != nil {
		t.Fatalf("failed to write to pipe: %v", err)
	}

	// Read from broker message channel
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got, err := b.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("failed to receive message from broker: %v", err)
	}

	if got.Source != "fake-proc" {
		t.Fatalf("expected source 'fake-proc', got %s", got.Source)
	}
	if got.ID != "evt-1" {
		t.Fatalf("expected id 'evt-1', got %s", got.ID)
	}

	// Ensure scanner sees EOF without hanging
	// write end closed by defer
}
