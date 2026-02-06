package taskflow

import (
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
	"testing"
)

func TestJobState_IsTerminal(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestJobState_IsTerminal", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

func TestJobState_CanTransitionTo_Matrix(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestJobState_CanTransitionTo_Matrix", nil, func(t *testing.T, tx *gorm.DB) {
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
	})
}

func TestJobState_IntermediateStates_IsTerminal(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestJobState_IntermediateStates_IsTerminal", nil, func(t *testing.T, tx *gorm.DB) {
		intermediate := []JobState{
			StateInitializing,
			StateFetching,
			StateProcessing,
			StateSaving,
			StateValidating,
			StateRecovering,
			StateRollingBack,
		}

		for _, s := range intermediate {
			if s.IsTerminal() {
				t.Errorf("expected %s to be non-terminal", s)
			}
		}
	})
}

func TestJobState_CanTransitionTo_IntermediateStates(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestJobState_CanTransitionTo_IntermediateStates", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			from     JobState
			to       JobState
			expected bool
		}{
			{StateQueued, StateInitializing, true},
			{StateInitializing, StateRunning, true},
			{StateInitializing, StateFailed, true},
			{StateRunning, StateFetching, true},
			{StateRunning, StateProcessing, true},
			{StateRunning, StateSaving, true},
			{StateRunning, StateValidating, true},
			{StateFetching, StateProcessing, true},
			{StateProcessing, StateSaving, true},
			{StatePaused, StateRecovering, true},
			{StateRecovering, StateRunning, true},
			{StateRollingBack, StateStopped, true},
			{StateRollingBack, StateFailed, true},
			{StateCompleted, StateRunning, false},
			{StateFailed, StateRunning, false},
			{StateStopped, StateRunning, false},
		}

		for _, tt := range tests {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CanTransitionTo(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.expected)
			}
		}
	})
}
