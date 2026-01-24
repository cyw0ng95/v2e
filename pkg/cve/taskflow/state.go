package taskflow

// JobState represents the state of a job run
type JobState string

const (
	// Base states (existing)
	StateQueued JobState = "queued"
	StateRunning JobState = "running"
	StatePaused JobState = "paused"
	StateCompleted JobState = "completed"
	StateFailed JobState = "failed"
	StateStopped JobState = "stopped"
	
	// New granular states
	StateInitializing JobState = "initializing"
	StateFetching JobState = "fetching"
	StateProcessing JobState = "processing"
	StateSaving JobState = "saving"
	StateValidating JobState = "validating"
	StateRecovering JobState = "recovering"
	StateRollingBack JobState = "rolling_back"
)

// IsTerminal returns whether this state is a terminal state
func (s JobState) IsTerminal() bool {
	return s == StateCompleted || s == StateFailed || s == StateStopped
}

// CanTransitionTo checks if transition to target state is allowed
func (s JobState) CanTransitionTo(target JobState) bool {
	switch s {
	case StateQueued:
		return target == StateInitializing || target == StateRunning || target == StateStopped
	case StateInitializing:
		return target == StateRunning || target == StateFailed || target == StateStopped
	case StateRunning:
		return target == StatePaused || target == StateCompleted || target == StateFailed || 
			target == StateStopped || target == StateFetching || target == StateProcessing || 
			target == StateSaving || target == StateValidating
	case StateFetching:
		return target == StateProcessing || target == StateRunning || target == StatePaused || 
			target == StateFailed || target == StateStopped
	case StateProcessing:
		return target == StateSaving || target == StateRunning || target == StatePaused || 
			target == StateFailed || target == StateStopped
	case StateSaving:
		return target == StateRunning || target == StatePaused || target == StateCompleted || 
			target == StateFailed || target == StateStopped
	case StateValidating:
		return target == StateRunning || target == StatePaused || target == StateCompleted || 
			target == StateFailed || target == StateStopped
	case StatePaused:
		return target == StateRunning || target == StateStopped || target == StateRecovering
	case StateRecovering:
		return target == StateRunning || target == StateFailed || target == StateStopped
	case StateCompleted, StateFailed, StateStopped:
		return false // Terminal states cannot transition
	case StateRollingBack:
		return target == StateStopped || target == StateFailed
	default:
		return false
	}
}
