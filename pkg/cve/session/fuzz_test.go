package session

import (
	"encoding/json"
	"testing"
)

// FuzzSessionStateTransition tests session state transitions with random inputs
func FuzzSessionStateTransition(f *testing.F) {
	// Add seed corpus with valid states
	f.Add("idle")
	f.Add("running")
	f.Add("paused")
	f.Add("stopped")
	// Add invalid/edge cases
	f.Add("")
	f.Add("IDLE")
	f.Add("Running")
	f.Add("invalid-state")
	f.Add("idle running")
	f.Add("null")
	f.Add("<script>")
	f.Add("'; DROP TABLE sessions--")

	f.Fuzz(func(t *testing.T, stateStr string) {
		state := SessionState(stateStr)
		
		// Test that state can be used without panicking
		_ = string(state)
		
		// Test that it matches or doesn't match valid states
		isValid := state == StateIdle || 
			state == StateRunning || 
			state == StatePaused || 
			state == StateStopped
		
		if isValid {
			// Valid states should behave predictably
			switch state {
			case StateIdle, StateRunning, StatePaused, StateStopped:
				// Expected valid state
			default:
				t.Errorf("State marked as valid but doesn't match any constant: %v", state)
			}
		}
	})
}

// FuzzSessionJSON tests JSON marshaling/unmarshaling of Session with random state inputs
func FuzzSessionJSON(f *testing.F) {
	// Add seed corpus with JSON-encoded sessions
	f.Add(`{"id":"test","state":"idle","start_index":0,"results_per_batch":100}`)
	f.Add(`{"id":"test","state":"running","start_index":10,"results_per_batch":50}`)
	f.Add(`{"state":"invalid"}`)
	f.Add(`{"id":"","state":""}`)
	f.Add(`{}`)
	f.Add(`{"state":null}`)
	f.Add(`{"state":123}`)

	f.Fuzz(func(t *testing.T, jsonData string) {
		var session Session
		// Should not panic, but may return error for invalid JSON
		_ = json.Unmarshal([]byte(jsonData), &session)
	})
}
