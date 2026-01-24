package main

import (
	"github.com/cyw0ng95/v2e/pkg/broker"
)

// spawnBroker captures the subset of broker behavior used by SpawnAdapter.
type spawnBroker interface {
	Spawn(id, command string, args ...string) (*ProcessInfo, error)
	SpawnRPC(id, command string, args ...string) (*ProcessInfo, error)
	SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error)
	SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error)
}

// SpawnAdapter delegates to the existing in-process Broker spawn methods.
// This adapter is intentionally simple and non-invasive.
type SpawnAdapter struct {
	b spawnBroker
}

func NewSpawnAdapter(b spawnBroker) *SpawnAdapter { return &SpawnAdapter{b: b} }

func toResult(info *ProcessInfo) *broker.SpawnResult {
	if info == nil {
		return nil
	}
	return &broker.SpawnResult{
		ID:       info.ID,
		PID:      info.PID,
		Command:  info.Command,
		Args:     info.Args,
		Status:   string(info.Status),
		ExitCode: info.ExitCode,
	}
}

// Ensure SpawnAdapter implements broker.Spawner
var _ broker.Spawner = (*SpawnAdapter)(nil)

func (s *SpawnAdapter) Spawn(id, command string, args ...string) (*broker.SpawnResult, error) {
	info, err := s.b.Spawn(id, command, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnRPC(id, command string, args ...string) (*broker.SpawnResult, error) {
	info, err := s.b.SpawnRPC(id, command, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*broker.SpawnResult, error) {
	info, err := s.b.SpawnWithRestart(id, command, maxRestarts, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*broker.SpawnResult, error) {
	info, err := s.b.SpawnRPCWithRestart(id, command, maxRestarts, args...)
	return toResult(info), err
}
