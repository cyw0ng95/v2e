package core

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
	p := NewTestProcess("p-exited", ProcessStatusExited)
	b.InsertProcessForTest(p)
	pid := p.Info().ID

	m2, _ := proc.NewEventMessage("evt-exited", nil)
	if err := b.SendToProcess(pid, m2); err == nil {
		t.Fatalf("expected error when sending to non-running process")
	}
}
