package main

import (
	"errors"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// stubSpawnBroker returns errors without creating processes, used to verify error propagation.
type stubSpawnBroker struct{}

func (stubSpawnBroker) Spawn(id, command string, args ...string) (*ProcessInfo, error) {
	return nil, errors.New("spawn failed")
}

func (stubSpawnBroker) SpawnRPC(id, command string, args ...string) (*ProcessInfo, error) {
	return nil, errors.New("spawn rpc failed")
}

func (stubSpawnBroker) SpawnWithRestart(id, command string, maxRestarts int, restartDelay time.Duration, args ...string) (*ProcessInfo, error) {
	return nil, errors.New("spawn restart failed")
}

func (stubSpawnBroker) SpawnRPCWithRestart(id, command string, maxRestarts int, restartDelay time.Duration, args ...string) (*ProcessInfo, error) {
	return nil, errors.New("spawn rpc restart failed")
}

func TestSpawnAdapter_ErrorPropagation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSpawnAdapter_ErrorPropagation", nil, func(t *testing.T, tx *gorm.DB) {
		adapter := NewSpawnAdapter(stubSpawnBroker{})

		if res, err := adapter.Spawn("id", "cmd"); err == nil || res != nil {
			t.Fatalf("expected spawn error, got res=%v err=%v", res, err)
		}
		if res, err := adapter.SpawnRPC("id", "cmd"); err == nil || res != nil {
			t.Fatalf("expected spawn rpc error, got res=%v err=%v", res, err)
		}
		if res, err := adapter.SpawnWithRestart("id", "cmd", 1, 0); err == nil || res != nil {
			t.Fatalf("expected spawn restart error, got res=%v err=%v", res, err)
		}
		if res, err := adapter.SpawnRPCWithRestart("id", "cmd", 1, 0); err == nil || res != nil {
			t.Fatalf("expected spawn rpc restart error, got res=%v err=%v", res, err)
		}
	})

}
