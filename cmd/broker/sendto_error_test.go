package main

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func TestSendToProcess_ErrorCases(t *testing.T) {
	b := NewBroker()

	// Case: missing process
	m, _ := proc.NewEventMessage("evt-missing", map[string]interface{}{"v": 1})
	if err := b.SendToProcess("no-such", m); err == nil {
		t.Fatalf("expected error when sending to missing process")
	}

	// Case: process exists but is not running
	pInfo := &ProcessInfo{ID: "p-exited", Status: ProcessStatusExited}
	p := &Process{info: pInfo, done: make(chan struct{})}
	b.mu.Lock()
	b.processes[pInfo.ID] = p
	b.mu.Unlock()

	m2, _ := proc.NewEventMessage("evt-exited", nil)
	if err := b.SendToProcess(pInfo.ID, m2); err == nil {
		t.Fatalf("expected error when sending to non-running process")
	}

	// Case: running but no stdin (does not support RPC)
	p.info.Status = ProcessStatusRunning
	p.stdin = nil
	m3, _ := proc.NewEventMessage("evt-nostdin", nil)
	if err := b.SendToProcess(pInfo.ID, m3); err == nil {
		t.Fatalf("expected error when sending to process without stdin")
	}
}
