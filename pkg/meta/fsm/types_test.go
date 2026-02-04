package fsm

import (
	"testing"
	"time"
)

func TestMacroStateConstants(t *testing.T) {
	states := []MacroState{
		MacroBootstrapping,
		MacroOrchestrating,
		MacroStabilizing,
		MacroDraining,
	}

	for _, state := range states {
		if state == "" {
			t.Errorf("Macro state should not be empty")
		}
	}
}

func TestProviderStateConstants(t *testing.T) {
	states := []ProviderState{
		ProviderIdle,
		ProviderAcquiring,
		ProviderRunning,
		ProviderWaitingQuota,
		ProviderWaitingBackoff,
		ProviderPaused,
		ProviderTerminated,
	}

	for _, state := range states {
		if state == "" {
			t.Errorf("Provider state should not be empty")
		}
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent(EventProviderStarted, "provider-1")

	if event.Type != EventProviderStarted {
		t.Errorf("Event type = %v, want %v", event.Type, EventProviderStarted)
	}
	if event.ProviderID != "provider-1" {
		t.Errorf("ProviderID = %v, want provider-1", event.ProviderID)
	}
	if event.Timestamp.IsZero() {
		t.Error("Event timestamp should be set")
	}
	if event.Data == nil {
		t.Error("Event data map should be initialized")
	}
}

func TestValidateMacroTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    MacroState
		to      MacroState
		wantErr bool
	}{
		{
			name:    "BOOTSTRAPPING to ORCHESTRATING",
			from:    MacroBootstrapping,
			to:      MacroOrchestrating,
			wantErr: false,
		},
		{
			name:    "ORCHESTRATING to STABILIZING",
			from:    MacroOrchestrating,
			to:      MacroStabilizing,
			wantErr: false,
		},
		{
			name:    "ORCHESTRATING to DRAINING (emergency)",
			from:    MacroOrchestrating,
			to:      MacroDraining,
			wantErr: false,
		},
		{
			name:    "STABILIZING to DRAINING",
			from:    MacroStabilizing,
			to:      MacroDraining,
			wantErr: false,
		},
		{
			name:    "STABILIZING to ORCHESTRATING (restart)",
			from:    MacroStabilizing,
			to:      MacroOrchestrating,
			wantErr: false,
		},
		{
			name:    "Same state transition",
			from:    MacroOrchestrating,
			to:      MacroOrchestrating,
			wantErr: false,
		},
		{
			name:    "Invalid: BOOTSTRAPPING to DRAINING",
			from:    MacroBootstrapping,
			to:      MacroDraining,
			wantErr: true,
		},
		{
			name:    "Invalid: DRAINING to BOOTSTRAPPING",
			from:    MacroDraining,
			to:      MacroBootstrapping,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMacroTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMacroTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProviderTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    ProviderState
		to      ProviderState
		wantErr bool
	}{
		{
			name:    "IDLE to ACQUIRING",
			from:    ProviderIdle,
			to:      ProviderAcquiring,
			wantErr: false,
		},
		{
			name:    "ACQUIRING to RUNNING",
			from:    ProviderAcquiring,
			to:      ProviderRunning,
			wantErr: false,
		},
		{
			name:    "RUNNING to WAITING_QUOTA",
			from:    ProviderRunning,
			to:      ProviderWaitingQuota,
			wantErr: false,
		},
		{
			name:    "RUNNING to WAITING_BACKOFF",
			from:    ProviderRunning,
			to:      ProviderWaitingBackoff,
			wantErr: false,
		},
		{
			name:    "RUNNING to PAUSED",
			from:    ProviderRunning,
			to:      ProviderPaused,
			wantErr: false,
		},
		{
			name:    "RUNNING to TERMINATED",
			from:    ProviderRunning,
			to:      ProviderTerminated,
			wantErr: false,
		},
		{
			name:    "WAITING_QUOTA to ACQUIRING",
			from:    ProviderWaitingQuota,
			to:      ProviderAcquiring,
			wantErr: false,
		},
		{
			name:    "WAITING_BACKOFF to ACQUIRING",
			from:    ProviderWaitingBackoff,
			to:      ProviderAcquiring,
			wantErr: false,
		},
		{
			name:    "PAUSED to ACQUIRING",
			from:    ProviderPaused,
			to:      ProviderAcquiring,
			wantErr: false,
		},
		{
			name:    "Same state transition",
			from:    ProviderRunning,
			to:      ProviderRunning,
			wantErr: false,
		},
		{
			name:    "Invalid: IDLE to RUNNING",
			from:    ProviderIdle,
			to:      ProviderRunning,
			wantErr: true,
		},
		{
			name:    "Invalid: TERMINATED to RUNNING",
			from:    ProviderTerminated,
			to:      ProviderRunning,
			wantErr: true,
		},
		{
			name:    "Invalid: IDLE to WAITING_QUOTA",
			from:    ProviderIdle,
			to:      ProviderWaitingQuota,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProviderTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventWithData(t *testing.T) {
	event := NewEvent(EventCheckpoint, "provider-cve")
	event.Data["urn"] = "v2e::nvd::cve::CVE-2024-12233"
	event.Data["count"] = 100

	if event.Data["urn"] != "v2e::nvd::cve::CVE-2024-12233" {
		t.Errorf("Event data URN not set correctly")
	}
	if event.Data["count"] != 100 {
		t.Errorf("Event data count not set correctly")
	}
}

func TestEventTimestamp(t *testing.T) {
	before := time.Now()
	event := NewEvent(EventProviderStarted, "test")
	after := time.Now()

	if event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Errorf("Event timestamp %v not between %v and %v", event.Timestamp, before, after)
	}
}
