package taskflow

// JobState represents the state of a job run
type JobState string

const (
	// StateQueued means job is queued but not started
	StateQueued JobState = "queued"
	// StateRunning means job is actively running
	StateRunning JobState = "running"
	// StatePaused means job is paused by user
	StatePaused JobState = "paused"
	// StateCompleted means job finished successfully
	StateCompleted JobState = "completed"
	// StateFailed means job failed with error
	StateFailed JobState = "failed"
	// StateStopped means job was stopped by user
	StateStopped JobState = "stopped"
)

// IsTerminal returns whether this state is a terminal state
func (s JobState) IsTerminal() bool {
	return s == StateCompleted || s == StateFailed || s == StateStopped
}

// CanTransitionTo checks if transition to target state is allowed
func (s JobState) CanTransitionTo(target JobState) bool {
	switch s {
	case StateQueued:
		return target == StateRunning || target == StateStopped
	case StateRunning:
		return target == StatePaused || target == StateCompleted || target == StateFailed || target == StateStopped
	case StatePaused:
		return target == StateRunning || target == StateStopped
	case StateCompleted, StateFailed, StateStopped:
		return false // Terminal states cannot transition
	default:
		return false
	}
}
