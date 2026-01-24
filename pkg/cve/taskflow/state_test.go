package taskflow

import "testing"

func TestJobState_IsTerminal(t *testing.T) {
	terminal := []JobState{StateCompleted, StateFailed, StateStopped}
	nonTerminal := []JobState{StateQueued, StateRunning, StatePaused}

	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Errorf("expected %s to be terminal", s)
		}
	}

	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("expected %s to be non-terminal", s)
		}
	}
}

func TestJobState_CanTransitionTo_Matrix(t *testing.T) {
	states := []JobState{StateQueued, StateRunning, StatePaused, StateCompleted, StateFailed, StateStopped}

	allowed := map[JobState]map[JobState]bool{
		StateQueued: {
			StateRunning: true,
			StateStopped: true,
		},
		StateRunning: {
			StatePaused:    true,
			StateCompleted: true,
			StateFailed:    true,
			StateStopped:   true,
		},
		StatePaused: {
			StateRunning: true,
			StateStopped: true,
		},
		StateCompleted: {},
		StateFailed:    {},
		StateStopped:   {},
	}

	for _, from := range states {
		for _, to := range states {
			exp := allowed[from][to]
			got := from.CanTransitionTo(to)
			if got != exp {
				t.Errorf("transition %s -> %s: expected %v, got %v", from, to, exp, got)
			}
		}
	}
}
