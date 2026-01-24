package main

import "testing"

func TestSetSpawnerAndToResult(t *testing.T) {
	b := NewBroker()
	adapter := NewSpawnAdapter(b)
	b.SetSpawner(adapter)

	if got := b.Spawner(); got == nil {
		t.Fatalf("expected spawner to be set, got nil")
	}

	// Validate toResult mapping without starting processes
	info := &ProcessInfo{
		ID:       "test-id",
		PID:      12345,
		Command:  "echo",
		Args:     []string{"hello"},
		Status:   ProcessStatusRunning,
		ExitCode: 0,
	}

	res := toResult(info)
	if res == nil {
		t.Fatalf("toResult returned nil")
	}
	if res.ID != "test-id" || res.PID != 12345 || res.Command != "echo" {
		t.Fatalf("toResult mapping incorrect: %+v", res)
	}
}
