package main

// SpawnAdapter delegates to the existing in-process Broker spawn methods.
// This adapter is intentionally simple and non-invasive.
type SpawnAdapter struct {
	b *Broker
}

func NewSpawnAdapter(b *Broker) *SpawnAdapter { return &SpawnAdapter{b: b} }

func toResult(info *ProcessInfo) *SpawnResult {
	if info == nil {
		return nil
	}
	return &SpawnResult{
		ID:       info.ID,
		PID:      info.PID,
		Command:  info.Command,
		Args:     info.Args,
		Status:   string(info.Status),
		ExitCode: info.ExitCode,
	}
}

func (s *SpawnAdapter) Spawn(id, command string, args ...string) (*SpawnResult, error) {
	info, err := s.b.Spawn(id, command, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnRPC(id, command string, args ...string) (*SpawnResult, error) {
	info, err := s.b.SpawnRPC(id, command, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*SpawnResult, error) {
	info, err := s.b.SpawnWithRestart(id, command, maxRestarts, args...)
	return toResult(info), err
}

func (s *SpawnAdapter) SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*SpawnResult, error) {
	info, err := s.b.SpawnRPCWithRestart(id, command, maxRestarts, args...)
	return toResult(info), err
}
