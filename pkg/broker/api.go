package broker

// SpawnResult is a lightweight DTO returned by Spawner implementations.
type SpawnResult struct {
	ID       string
	PID      int
	Command  string
	Args     []string
	Status   string
	ExitCode int
}

// Spawner is a minimal interface for creating subprocesses. Implementations
// should delegate to existing broker spawn logic for this low-risk initial
// refactor.
type Spawner interface {
	Spawn(id, command string, args ...string) (*SpawnResult, error)
	SpawnRPC(id, command string, args ...string) (*SpawnResult, error)
	SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*SpawnResult, error)
	SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*SpawnResult, error)
}

// Auth is a small extension hook for future authentication/authorization
// implementations. The initial implementation may be a no-op.
type Auth interface {
	Init() error
}
