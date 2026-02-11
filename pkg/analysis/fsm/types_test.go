package fsm

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestValidateGraphTransition(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ValidateGraphTransition", nil, func(t *testing.T, _ *gorm.DB) {
		tests := []struct {
			name    string
			from    GraphState
			to      GraphState
			wantErr bool
		}{
			{
				name:    "idle to building",
				from:    GraphIdle,
				to:      GraphBuilding,
				wantErr: false,
			},
			{
				name:    "building to ready",
				from:    GraphBuilding,
				to:      GraphReady,
				wantErr: false,
			},
			{
				name:    "building to error",
				from:    GraphBuilding,
				to:      GraphError,
				wantErr: false,
			},
			{
				name:    "ready to analyzing",
				from:    GraphReady,
				to:      GraphAnalyzing,
				wantErr: false,
			},
			{
				name:    "ready to persisting",
				from:    GraphReady,
				to:      GraphPersisting,
				wantErr: false,
			},
			{
				name:    "error to idle",
				from:    GraphError,
				to:      GraphIdle,
				wantErr: false,
			},
			{
				name:    "invalid: idle to analyzing",
				from:    GraphIdle,
				to:      GraphAnalyzing,
				wantErr: true,
			},
			{
				name:    "invalid: building to persisting",
				from:    GraphBuilding,
				to:      GraphPersisting,
				wantErr: true,
			},
			{
				name:    "same state is valid",
				from:    GraphReady,
				to:      GraphReady,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateGraphTransition(tt.from, tt.to)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateGraphTransition() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestValidateAnalyzeTransition(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ValidateAnalyzeTransition", nil, func(t *testing.T, _ *gorm.DB) {
		tests := []struct {
			name    string
			from    AnalyzeState
			to      AnalyzeState
			wantErr bool
		}{
			{
				name:    "bootstrapping to idle",
				from:    AnalyzeBootstrapping,
				to:      AnalyzeIdle,
				wantErr: false,
			},
			{
				name:    "idle to processing",
				from:    AnalyzeIdle,
				to:      AnalyzeProcessing,
				wantErr: false,
			},
			{
				name:    "processing to paused",
				from:    AnalyzeProcessing,
				to:      AnalyzePaused,
				wantErr: false,
			},
			{
				name:    "paused to idle",
				from:    AnalyzePaused,
				to:      AnalyzeIdle,
				wantErr: false,
			},
			{
				name:    "draining to terminated",
				from:    AnalyzeDraining,
				to:      AnalyzeTerminated,
				wantErr: false,
			},
			{
				name:    "invalid: idle to terminated",
				from:    AnalyzeIdle,
				to:      AnalyzeTerminated,
				wantErr: true,
			},
			{
				name:    "invalid: processing to bootstrapping",
				from:    AnalyzeProcessing,
				to:      AnalyzeBootstrapping,
				wantErr: true,
			},
			{
				name:    "same state is valid",
				from:    AnalyzeIdle,
				to:      AnalyzeIdle,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateAnalyzeTransition(tt.from, tt.to)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateAnalyzeTransition() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestNewEvent(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewEvent", nil, func(t *testing.T, _ *gorm.DB) {
		event := NewEvent(EventGraphBuildStarted)

		if event.Type != EventGraphBuildStarted {
			t.Errorf("NewEvent() Type = %v, want %v", event.Type, EventGraphBuildStarted)
		}

		if event.Timestamp.IsZero() {
			t.Error("NewEvent() Timestamp is zero")
		}

		if event.Data == nil {
			t.Error("NewEvent() Data is nil")
		}
	})
}

func TestGraphStateTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStateTransitions", nil, func(t *testing.T, _ *gorm.DB) {
		validTransitions := []GraphStateTransition{
			{GraphIdle, GraphBuilding},
			{GraphBuilding, GraphReady},
			{GraphBuilding, GraphError},
			{GraphReady, GraphAnalyzing},
			{GraphReady, GraphPersisting},
			{GraphReady, GraphBuilding},
			{GraphReady, GraphIdle},
			{GraphAnalyzing, GraphReady},
			{GraphAnalyzing, GraphError},
			{GraphPersisting, GraphReady},
			{GraphPersisting, GraphError},
			{GraphError, GraphIdle},
			{GraphError, GraphBuilding},
		}

		for _, trans := range validTransitions {
			if !validGraphTransitions[trans] {
				t.Errorf("Transition %s -> %s should be valid but is not in map", trans.From, trans.To)
			}
		}
	})
}

func TestAnalyzeStateTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "AnalyzeStateTransitions", nil, func(t *testing.T, _ *gorm.DB) {
		validTransitions := []AnalyzeStateTransition{
			{AnalyzeBootstrapping, AnalyzeIdle},
			{AnalyzeBootstrapping, AnalyzeTerminated},
			{AnalyzeIdle, AnalyzeProcessing},
			{AnalyzeIdle, AnalyzePaused},
			{AnalyzeIdle, AnalyzeDraining},
			{AnalyzeProcessing, AnalyzeIdle},
			{AnalyzeProcessing, AnalyzePaused},
			{AnalyzeProcessing, AnalyzeDraining},
			{AnalyzePaused, AnalyzeIdle},
			{AnalyzePaused, AnalyzeProcessing},
			{AnalyzePaused, AnalyzeDraining},
			{AnalyzeDraining, AnalyzeTerminated},
		}

		for _, trans := range validTransitions {
			if !validAnalyzeTransitions[trans] {
				t.Errorf("Transition %s -> %s should be valid but is not in map", trans.From, trans.To)
			}
		}
	})
}
