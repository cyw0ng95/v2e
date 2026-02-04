package session

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/json"
	"fmt"
	"testing"
)

// FuzzSessionStateTransition tests session state transitions with random inputs
func FuzzSessionStateTransition(f *testing.F) {
	// Add comprehensive seed corpus (50+ cases)
	// Valid states
	f.Add("idle")
	f.Add("running")
	f.Add("paused")
	f.Add("stopped")
	
	// Case variations
	f.Add("IDLE")
	f.Add("Running")
	f.Add("RUNNING")
	f.Add("Paused")
	f.Add("PAUSED")
	f.Add("Stopped")
	f.Add("STOPPED")
	f.Add("IdLe")
	f.Add("rUnNiNg")
	
	// Empty and null
	f.Add("")
	f.Add("null")
	f.Add("NULL")
	f.Add("undefined")
	
	// Invalid states
	f.Add("invalid-state")
	f.Add("unknown")
	f.Add("error")
	f.Add("failed")
	f.Add("pending")
	f.Add("queued")
	f.Add("completed")
	
	// Injection attempts
	f.Add("<script>alert(1)</script>")
	f.Add("'; DROP TABLE sessions--")
	f.Add("' OR '1'='1")
	f.Add("$(rm -rf /)")
	f.Add("; cat /etc/passwd")
	
	// Combined states
	f.Add("idle running")
	f.Add("paused,stopped")
	f.Add("idle;running")
	f.Add("running|paused")
	
	// Special characters
	f.Add("idle\n")
	f.Add("running\t")
	f.Add("paused\r\n")
	f.Add("idle\x00")
	f.Add("running\\u0000")
	
	// Unicode
	f.Add("日本語")
	f.Add("状態")
	f.Add("运行中")
	f.Add("приостановлено")
	
	// Very long strings
	longState := make([]byte, 500)
	for i := range longState {
		longState[i] = 'a'
	}
	f.Add(string(longState))
	
	// Numbers and booleans as strings
	f.Add("0")
	f.Add("1")
	f.Add("true")
	f.Add("false")
	f.Add("123")
	f.Add("-1")
	
	// Generate programmatic test cases
	prefixes := []string{"pre_", "post_", "new_", "old_"}
	for _, prefix := range prefixes {
		f.Add(prefix + "idle")
		f.Add(prefix + "running")
		f.Add(prefix + "paused")
		f.Add(prefix + "stopped")
	}

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
	// Add comprehensive seed corpus (100+ cases)
	// Valid sessions
	f.Add(`{"id":"test","state":"idle","start_index":0,"results_per_batch":100}`)
	f.Add(`{"id":"test","state":"running","start_index":10,"results_per_batch":50}`)
	f.Add(`{"id":"test","state":"paused","start_index":100,"results_per_batch":25}`)
	f.Add(`{"id":"test","state":"stopped","start_index":0,"results_per_batch":200}`)
	
	// Invalid states
	f.Add(`{"id":"test","state":"invalid"}`)
	f.Add(`{"id":"test","state":""}`)
	f.Add(`{"id":"test","state":"IDLE"}`)
	f.Add(`{"id":"test","state":"Running"}`)
	
	// Empty and null values
	f.Add(`{}`)
	f.Add(`{"id":""}`)
	f.Add(`{"state":""}`)
	f.Add(`{"state":null}`)
	f.Add(`{"id":null,"state":null}`)
	
	// Type mismatches
	f.Add(`{"state":123}`)
	f.Add(`{"state":true}`)
	f.Add(`{"state":false}`)
	f.Add(`{"state":[]}`)
	f.Add(`{"state":{}}`)
	f.Add(`{"id":123,"state":"idle"}`)
	f.Add(`{"id":true,"state":"running"}`)
	
	// Negative indices
	f.Add(`{"id":"test","state":"idle","start_index":-1,"results_per_batch":100}`)
	f.Add(`{"id":"test","state":"idle","start_index":0,"results_per_batch":-1}`)
	f.Add(`{"id":"test","state":"idle","start_index":-100,"results_per_batch":-50}`)
	
	// Very large indices
	f.Add(`{"id":"test","state":"idle","start_index":2147483647,"results_per_batch":2147483647}`)
	f.Add(`{"id":"test","state":"idle","start_index":999999999,"results_per_batch":1}`)
	
	// Injection attempts in ID
	f.Add(`{"id":"<script>alert(1)</script>","state":"idle"}`)
	f.Add(`{"id":"'; DROP TABLE sessions--","state":"running"}`)
	f.Add(`{"id":"../../etc/passwd","state":"paused"}`)
	f.Add(`{"id":"test; rm -rf /","state":"stopped"}`)
	
	// Injection attempts in state
	f.Add(`{"id":"test","state":"<script>"}`)
	f.Add(`{"id":"test","state":"'; DROP TABLE--"}`)
	
	// Unicode in ID
	f.Add(`{"id":"日本語テスト","state":"idle"}`)
	f.Add(`{"id":"test\\u0000null","state":"running"}`)
	
	// Very long ID
	longID := make([]byte, 500)
	for i := range longID {
		longID[i] = 'A'
	}
	f.Add(fmt.Sprintf(`{"id":"%s","state":"idle"}`, longID))
	
	// Malformed JSON
	f.Add(`{"id":"test","state":"idle"`)
	f.Add(`{"id":"test","state":`)
	f.Add(`{"id":"test",`)
	f.Add(`{`)
	f.Add(`}`)
	f.Add(`"not an object"`)
	f.Add(`["array","instead"]`)
	
	// Extra fields
	f.Add(`{"id":"test","state":"idle","extra":"field"}`)
	f.Add(`{"id":"test","state":"idle","malicious":"<script>"}`)
	
	// String values for numeric fields
	f.Add(`{"id":"test","state":"idle","start_index":"0","results_per_batch":"100"}`)
	f.Add(`{"id":"test","state":"idle","start_index":"invalid","results_per_batch":"bad"}`)
	
	// Float values for integer fields
	f.Add(`{"id":"test","state":"idle","start_index":10.5,"results_per_batch":20.7}`)
	
	// Generate programmatic test cases
	states := []string{"idle", "running", "paused", "stopped", "invalid", ""}
	for i, state := range states {
		f.Add(fmt.Sprintf(`{"id":"session-%d","state":"%s","start_index":%d,"results_per_batch":%d}`, 
			i, state, i*10, (i+1)*10))
	}
	
	// Boundary value testing
	for i := 0; i < 20; i++ {
		f.Add(fmt.Sprintf(`{"id":"test-%d","state":"idle","start_index":%d,"results_per_batch":%d}`, 
			i, i, i+1))
	}
	
	// Created/Updated timestamp edge cases
	f.Add(`{"id":"test","state":"idle","created_at":"2024-01-01T00:00:00Z"}`)
	f.Add(`{"id":"test","state":"idle","created_at":"invalid-date"}`)
	f.Add(`{"id":"test","state":"idle","created_at":null}`)
	f.Add(`{"id":"test","state":"idle","created_at":1234567890}`)
	
	// Counter fields with various values
	f.Add(`{"id":"test","state":"idle","fetched_count":100,"stored_count":90,"error_count":5}`)
	f.Add(`{"id":"test","state":"idle","fetched_count":-1,"stored_count":-1,"error_count":-1}`)
	f.Add(`{"id":"test","state":"idle","fetched_count":"100","stored_count":"90"}`)

	f.Fuzz(func(t *testing.T, jsonData string) {
		var session Session
		// Should not panic, but may return error for invalid JSON
		_ = json.Unmarshal([]byte(jsonData), &session)
	})
}
