package main

import (
	"github.com/cyw0ng95/v2e/cmd/v2broker/core"
)

type spawnBroker interface {
	Spawn(id, command string, args ...string) (*core.ProcessInfo, error)
	SpawnRPC(id, command string, args ...string) (*core.ProcessInfo, error)
	SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*core.ProcessInfo, error)
	SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*core.ProcessInfo, error)
}

// SpawnAdapter delegates to existing in-process Broker spawn methods.
// This adapter is intentionally simple and non-invasive.
type SpawnAdapter struct {
	b spawnBroker
}

func NewSpawnAdapter(b spawnBroker) *SpawnAdapter { return &SpawnAdapter{b: b} }

func toResult(info *core.ProcessInfo) *core.SpawnResult {
	if info == nil {
		return nil
	}
	return &core.SpawnResult{
		ID:       info.ID,
		PID:      info.PID,
		Command:  info.Command,
		Args:     info.Args,
		Status:   string(info.Status),
		ExitCode: info.ExitCode,
	}
}

// Ensure SpawnAdapter implements core.Spawner
var _ core.Spawner = (*SpawnAdapter)(nil)

func (s *SpawnAdapter) Spawn(id, command string, args ...string) (*core.SpawnResult, error) {
	info, err := s.b.Spawn(id, command, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnRPC(id, command string, args ...string) (*core.SpawnResult, error) {
	info, err := s.b.SpawnRPC(id, command, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*core.SpawnResult, error) {
	info, err := s.b.SpawnWithRestart(id, command, maxRestarts, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*core.SpawnResult, error) {
	info, err := s.b.SpawnRPCWithRestart(id, command, maxRestarts, args...)
	return toResult(info), err
}
