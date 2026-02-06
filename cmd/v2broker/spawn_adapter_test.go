package main

import (
	"testing"

	"github.com/cyw0ng95/v2e/cmd/v2broker/core"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestSetSpawnerAndToResult(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetSpawnerAndToResult", nil, func(t *testing.T, tx *gorm.DB) {
		b := core.NewBroker()
		adapter := NewSpawnAdapter(b)
		b.SetSpawner(adapter)

		if got := b.Spawner(); got == nil {
			t.Fatalf("expected spawner to be set, got nil")
		}

		// Validate toResult mapping without starting processes
		info := &core.ProcessInfo{
			ID:       "test-id",
			PID:      12345,
			Command:  "echo",
			Args:     []string{"hello"},
			Status:   core.ProcessStatusRunning,
			ExitCode: 0,
		}

		res := toResult(info)
		if res == nil {
			t.Fatalf("toResult returned nil")
		}
		if res.ID != "test-id" || res.PID != 12345 || res.Command != "echo" {
			t.Fatalf("toResult mapping incorrect: %+v", res)
		}
	})

}

func TestToResult_NilSafe(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestToResult_NilSafe", nil, func(t *testing.T, tx *gorm.DB) {
		if toResult(nil) != nil {
			t.Fatalf("expected nil when input info is nil")
		}
	})

}
