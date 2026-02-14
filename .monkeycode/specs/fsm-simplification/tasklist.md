# FSM Simplification Implementation Plan

## Overview

Simplify the MacroFSM and ProviderFSM architecture by:
1. Adding a translation layer between backend states and frontend display states
2. Consolidating excessive RPC calls into unified endpoints

## Change Estimation Summary

| Component | Files Affected | Estimated LoC | Complexity |
|-----------|----------------|---------------|------------|
| Backend Types | 1 | +50 | Low |
| RPC Consolidation | 2 | +120 | High |
| Backend Tests | 2 | +100 | Medium |
| Frontend Types | 1 | +30 | Low |
| Frontend Components | 1 | +100 | Medium |
| Frontend Tests | 2 | +80 | Low |
| **Total** | **9** | **~480** | **High** |

---

- [ ] 1. Phase 1: Backend Type Definitions
  - [ ] 1.1 Add simplified state types in `pkg/meta/fsm/types.go`
    - Create `SimplifiedProviderState` type (ACTIVE, READY, INACTIVE)
    - Create `SimplifiedMacroState` type (RUNNING, TRANSITIONING)
    - Add conversion functions: `ToSimplifiedProviderState()`, `ToSimplifiedMacroState()`
    - Reference: Current ProviderState has 7 variants, MacroState has 4 variants

  - [ ] 1.2 Add optional advanced mode flag to provider config
    - Add `ShowAdvancedStates bool` field to existing config structs

- [ ] 2. Phase 2: RPC Consolidation (Remove Excessive Calls)
  - [ ] 2.1 Analyze current RPC handlers
    - Current FSM handlers: 11 total (individual + bulk + tree + checkpoints)
    - Remove duplicate handlers between main.go and fsm_integration.go
    - Identify handlers that can be merged

  - [ ] 2.2 Create unified FSM control RPC
    - New `RPCFSMControl` endpoint handling all actions (start/stop/pause/resume)
    - Request: `{ action: "start"|"stop"|"pause"|"resume", provider_id?: string }`
    - Single endpoint replaces 8 individual/bulk handlers
    - Action "all" operates on all providers

  - [ ] 2.3 Add simplified topology RPC
    - `RPCFSMGetTopology` - returns full topology (replaces GetEtlTree)
    - Add `simplified=true` parameter for simplified states
    - Add `include_checkpoints=true` for optional checkpoint data

  - [ ] 2.4 Remove deprecated handlers
    - Remove duplicates: RPCStartProvider, RPCPauseProvider, RPCStopProvider from main.go
    - Remove: RPCFSMGetProviderList (merged into topology)
    - Keep: RPCFSMGetProviderCheckpoints (separate - used for detailed view)

  - [ ] 2.5 Update access gateway routes
    - Update `cmd/v2access/rpc_routes.go` for new unified endpoints

- [ ] 3. Phase 3: Frontend Type Updates
  - [ ] 3.1 Update `website/lib/types.ts`
    - Add `SimplifiedProviderState` type
    - Add `SimplifiedMacroState` type
    - Add `FSMControlRequest` type

  - [ ] 3.2 Update RPC client in `website/lib/rpc-client.ts`
    - Replace individual provider methods with unified `controlFSM()`
    - Update `getTopology()` to support simplified mode

- [ ] 4. Phase 4: Frontend Component Updates
  - [ ] 4.1 Enhance `ETLTopologyViewer` component
    - Add simplified mode toggle (basic/advanced switch)
    - In simplified mode: show 3 provider states (ACTIVE, READY, INACTIVE)
    - In simplified mode: show 2 macro states (RUNNING, TRANSITIONING)
    - Keep current detailed view as "Advanced" mode

  - [ ] 4.2 Add simplified state mapping utilities
    - `mapToSimplifiedState(fullState)` helper function
    - Consistent color/icon mapping for simplified states

  - [ ] 4.3 Update controls for simplified mode
    - Hide Pause/Resume in basic mode
    - Show only Start/Stop controls
    - Keep full controls in advanced mode

- [ ] 5. Phase 5: Testing
  - [ ] 5.1 Add unit tests for state conversion functions in `pkg/meta/fsm/types_test.go`
    - [ ] 5.1.1 Test ToSimplifiedProviderState() for all 7 provider states
      - RUNNING → ACTIVE
      - IDLE → READY
      - ACQUIRING → READY
      - WAITING_QUOTA → INACTIVE
      - WAITING_BACKOFF → INACTIVE
      - PAUSED → INACTIVE
      - TERMINATED → INACTIVE

    - [ ] 5.1.2 Test ToSimplifiedMacroState() for all 4 macro states
      - ORCHESTRATING → RUNNING
      - BOOTSTRAPPING → TRANSITIONING
      - STABILIZING → TRANSITIONING
      - DRAINING → TRANSITIONING

    - [ ] 5.1.3 Test edge cases (unknown states return default)
    - [ ] 5.1.4 Test validation functions: ValidateProviderTransition, ValidateMacroTransition

  - [ ] 5.2 Add RPC handler tests in `pkg/meta/fsm/rpc_test.go` (NEW FILE)
    - [ ] 5.2.1 Test RPCFSMControl with action="start" and valid provider_id
    - [ ] 5.2.2 Test RPCFSMControl with action="stop" (all providers)
    - [ ] 5.2.3 Test RPCFSMControl with action="pause" and invalid provider_id
    - [ ] 5.2.4 Test RPCFSMControl with action="resume" 
    - [ ] 5.2.5 Test RPCFSMControl with invalid action (error response)
    - [ ] 5.2.6 Test RPCFSMControl with non-existent provider (error in Failed map)

  - [ ] 5.3 Add topology response tests
    - [ ] 5.3.1 Test RPCFSMGetTopology with simplified=false (full states)
    - [ ] 5.3.2 Test RPCFSMGetTopology with simplified=true (simplified states)
    - [ ] 5.3.3 Test RPCFSMGetTopology with include_checkpoints=true
    - [ ] 5.3.4 Verify response structure matches expected schema

  - [ ] 5.4 Add frontend component tests in `website/components/etl-topology-viewer.test.tsx` (NEW FILE)
    - [ ] 5.4.1 Test simplified state mapping display
    - [ ] 5.4.2 Test advanced mode toggle functionality
    - [ ] 5.4.3 Test control buttons visibility in simplified mode
    - [ ] 5.4.4 Test control buttons visibility in advanced mode
    - [ ] 5.4.5 Test provider card click opens detail dialog
    - [ ] 5.4.6 Test loading state displays correctly

  - [ ] 5.5 Add RPC client tests in `website/lib/__tests__/rpc-client.test.ts`
    - [ ] 5.5.1 Test controlFSM() constructs correct request
    - [ ] 5.5.2 Test getTopology() with simplified parameter
    - [ ] 5.5.3 Test backward compatibility (existing methods still work)

  - [ ] 5.6 Run all tests and verify
    - [ ] 5.6.1 Run Go tests: `go test ./pkg/meta/fsm/...`
    - [ ] 5.6.2 Run frontend tests: `npm test -- --passWithNoTests`
    - [ ] 5.6.3 Verify test coverage > 80% for modified packages

- [ ] 6. Checkpoint - Ensure all tests pass

---

## RPC Consolidation Details

### Current State (11 handlers)

| Handler | Purpose | Suggested Action |
|---------|---------|------------------|
| `RPCFSMStartProvider` | Start single provider | Merge into RPCFSMControl |
| `RPCFSMStopProvider` | Stop single provider | Merge into RPCFSMControl |
| `RPCFSMPauseProvider` | Pause single provider | Merge into RPCFSMControl |
| `RPCFSMResumeProvider` | Resume single provider | Merge into RPCFSMControl |
| `RPCFSMStartAllProviders` | Start all | Merge into RPCFSMControl |
| `RPCFSMStopAllProviders` | Stop all | Merge into RPCFSMControl |
| `RPCFSMPauseAllProviders` | Pause all | Merge into RPCFSMControl |
| `RPCFSMResumeAllProviders` | Resume all | Merge into RPCFSMControl |
| `RPCFSMGetProviderList` | List providers | Merge into topology |
| `RPCFSMGetEtlTree` | Get full tree | Rename to RPCFSMGetTopology |
| `RPCFSMGetProviderCheckpoints` | Get checkpoints | Keep separate |

### Duplicates to Remove (main.go)

| Handler | Duplicate Of | Action |
|---------|--------------|--------|
| `RPCStartProvider` | RPCFSMStartProvider | Remove |
| `RPCPauseProvider` | RPCFSMPauseProvider | Remove |
| `RPCStopProvider` | RPCFSMStopProvider | Remove |

### New State (3 handlers)

| Handler | Purpose |
|---------|---------|
| `RPCFSMControl` | Unified control (start/stop/pause/resume single or all) |
| `RPCFSMGetTopology` | Get topology with optional simplified states |
| `RPCFSMGetProviderCheckpoints` | Keep separate for detailed view |

### Request/Response Format

```go
// RPCFSMControl Request
type FSMControlRequest struct {
    Action     string `json:"action"` // "start", "stop", "pause", "resume"
    ProviderID string `json:"provider_id,omitempty"` // omit for "all"
}

// RPCFSMControl Response
type FSMControlResponse struct {
    Success bool     `json:"success"`
    Action  string   `json:"action"`
    Affected []string `json:"affected"` // provider IDs affected
    Failed   map[string]string `json:"failed,omitempty"` // provider -> error
}

// RPCFSMGetTopology Request
type FSMTopologyRequest struct {
    Simplified bool `json:"simplified"`
    IncludeCheckpoints bool `json:"include_checkpoints"`
}
```

## State Mapping Reference

### Provider States

| Current State | Simplified State | Condition |
|---------------|-----------------|-----------|
| RUNNING | ACTIVE | - |
| IDLE | READY | - |
| ACQUIRING | READY | Internal state |
| WAITING_QUOTA | INACTIVE | Paused externally |
| WAITING_BACKOFF | INACTIVE | Rate limited |
| PAUSED | INACTIVE | User paused |
| TERMINATED | INACTIVE | Completed/stopped |

### Macro States

| Current State | Simplified State | Condition |
|---------------|-----------------|-----------|
| ORCHESTRATING | RUNNING | Active coordination |
| BOOTSTRAPPING | TRANSITIONING | Initializing |
| STABILIZING | TRANSITIONING | Winding down |
| DRAINING | TRANSITIONING | Cleanup |

## Implementation Priority

1. **P0 (Critical)**: Phase 1 + Phase 2 - Core simplification logic
2. **P1 (High)**: Phase 3 + Phase 4 - Frontend display changes
3. **P2 (Medium)**: Phase 5 - Testing and validation

## Notes

- Backend keeps full state machine unchanged for internal operations
- Frontend receives both simplified and full states (when advanced=true)
- Default behavior preserved for backward compatibility
- No database migrations needed (stateless translation layer)
