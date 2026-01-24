// Package process implements process lifecycle management for the broker
package process

import (
	"context"
	"time"
)

// ProcessManager defines the interface for managing processes
type ProcessManager interface {
	Spawn(id, command string, args ...string) (*ProcessInfo, error)
	SpawnRPC(id, command string, args ...string) (*ProcessInfo, error)
	SpawnWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error)
	SpawnRPCWithRestart(id, command string, maxRestarts int, args ...string) (*ProcessInfo, error)
	Kill(id string) error
	GetProcess(id string) (*ProcessInfo, error)
	ListProcesses() []*ProcessInfo
	LoadProcessesFromConfig(config interface{}) error
}

// ProcessInfo contains information about a managed process
type ProcessInfo struct {
	ID        string
	PID       int
	Command   string
	Args      []string
	Status    ProcessStatus
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
}

// ProcessStatus represents the status of a process
type ProcessStatus string

const (
	ProcessStatusRunning ProcessStatus = "running"
	ProcessStatusExited  ProcessStatus = "exited"
	ProcessStatusFailed  ProcessStatus = "failed"
)

// RestartConfig holds restart configuration for a process
type RestartConfig struct {
	Enabled      bool
	MaxRestarts  int
	RestartCount int
	Command      string
	Args         []string
	IsRPC        bool
}

// Process represents a managed process
type Process struct {
	info          *ProcessInfo
	cancel        context.CancelFunc
	done          chan struct{}
	restartConfig *RestartConfig
}