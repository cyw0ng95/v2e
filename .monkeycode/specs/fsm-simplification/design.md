# FSM Simplification Design

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│  ┌─────────────────┐      ┌─────────────────────────────┐  │
│  │ Simplified View │◄────►│     Advanced View          │  │
│  │ (3 states)     │      │     (7 provider states)    │  │
│  └────────┬────────┘      └─────────────┬─────────────┘  │
│           │                              │                  │
│           └──────────┬───────────────────┘                  │
│                      ▼                                      │
│           ┌──────────────────────┐                          │
│           │   State Translator   │                          │
│           │   (Frontend Helper)  │                          │
│           └──────────────────────┘                          │
└──────────────────────┬─────────────────────────────────────┘
                       │ RPC
┌──────────────────────┼─────────────────────────────────────┐
│                      ▼           Backend (Go)              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           RPC Handler (fsm_integration.go)         │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │  RPCFSMControl                              │   │   │
│  │  │  - action: start/stop/pause/resume        │   │   │
│  │  │  - provider_id: optional (all if omitted) │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │  RPCFSMGetTopology                         │   │   │
│  │  │  - simplified: bool                         │   │   │
│  │  │  - include_checkpoints: bool               │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                      │                                      │
│           ┌──────────┴──────────┐                          │
│           ▼                     ▼                          │
│  ┌────────────────┐   ┌─────────────────┐                 │
│  │  MacroFSM      │   │  ProviderFSM   │                 │
│  │  (4 states)    │   │  (7 states)    │                 │
│  └────────────────┘   └─────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

## RPC Consolidation

### Before (11 handlers)

```
RPCFSMStartProvider ──────┐
RPCFSMStopProvider ───────┤
RPCFSMPauseProvider ──────┤
RPCFSMResumeProvider ─────┼──► 8 handlers for control
RPCFSMStartAllProviders ──┤
RPCFSMStopAllProviders ───┤
RPCFSMPauseAllProviders ──┤
RPCFSMResumeAllProviders ─┘

RPCFSMGetProviderList ────► Merge into topology
RPCFSMGetEtlTree ─────────► Rename to topology
RPCFSMGetProviderCheckpoints ──► Keep separate
```

### After (3 handlers)

```
RPCFSMControl ───────────────► Unified control (replaces 8)
RPCFSMGetTopology ──────────► Unified topology (replaces 2)
RPCFSMGetProviderCheckpoints ─► Keep separate
```

## Components

### 1. Backend Type Additions (`pkg/meta/fsm/types.go`)

```go
// Simplified states for frontend display
type SimplifiedProviderState string

const (
    SimplifiedActive   SimplifiedProviderState = "ACTIVE"
    SimplifiedReady    SimplifiedProviderState = "READY"
    SimplifiedInactive SimplifiedProviderState = "INACTIVE"
)

type SimplifiedMacroState string

const (
    SimplifiedRunning      SimplifiedMacroState = "RUNNING"
    SimplifiedTransitioning SimplifiedMacroState = "TRANSITIONING"
)

// Conversion functions
func ToSimplifiedProviderState(state ProviderState) SimplifiedProviderState
func ToSimplifiedMacroState(state MacroState) SimplifiedMacroState
```

### 2. Unified RPC Control Handler (`cmd/v2meta/fsm_integration.go`)

**New unified endpoint:**

```go
// RPCFSMControl - Unified FSM control endpoint
// Replaces: RPCFSMStartProvider, RPCFSMStopProvider, RPCFSMPauseProvider,
//           RPCFSMResumeProvider, RPCFSMStartAllProviders, RPCFSMStopAllProviders,
//           RPCFSMPauseAllProviders, RPCFSMResumeAllProviders

type FSMControlRequest struct {
    Action     string `json:"action"`     // "start", "stop", "pause", "resume"
    ProviderID string `json:"provider_id,omitempty"` // empty = all providers
}

type FSMControlResponse struct {
    Success   bool              `json:"success"`
    Action    string            `json:"action"`
    Affected  []string          `json:"affected"`   // provider IDs
    Failed    map[string]string `json:"failed,omitempty"` // provider -> error
}
```

### 3. Unified Topology Handler

```go
// RPCFSMGetTopology - Get FSM topology
// Replaces: RPCFSMGetProviderList, RPCFSMGetEtlTree

type FSMTopologyRequest struct {
    Simplified         bool `json:"simplified"`          // simplified states
    IncludeCheckpoints bool `json:"include_checkpoints"` // include checkpoint data
}

type FSMTopologyResponse struct {
    Macro     map[string]interface{} `json:"macro"`
    Providers []map[string]interface{} `json:"providers"`
}
```

### 4. Frontend Types (`website/lib/types.ts`)

```typescript
type SimplifiedProviderState = 'ACTIVE' | 'READY' | 'INACTIVE';
type SimplifiedMacroState = 'RUNNING' | 'TRANSITIONING';

interface SimplifiedProviderNode {
  id: string;
  providerType: string;
  state: SimplifiedProviderState;
  processedCount: number;
  errorCount: number;
}

interface SimplifiedTopologyData {
  macro: {
    state: SimplifiedMacroState;
  };
  providers: SimplifiedProviderNode[];
}
```

### 4. Frontend Component (`website/components/etl-topology-viewer.tsx`)

**State Management:**
- Add `simplifiedMode` boolean state (default: true)
- Add `toggleViewMode()` callback

**Render Logic:**
```typescript
// Simplified mode (default)
const displayState = simplifiedMode 
  ? mapToSimplifiedState(provider.state)
  : provider.state;

// Controls
const showBasicControls = simplifiedMode;
const showAdvancedControls = !simplifiedMode || showAllControls;
```

## State Mapping Tables

### Provider State Mapping

| Full State | Simplified | Display Color |
|------------|------------|---------------|
| RUNNING | ACTIVE | green |
| IDLE | READY | blue |
| ACQUIRING | READY | blue |
| WAITING_QUOTA | INACTIVE | yellow |
| WAITING_BACKOFF | INACTIVE | yellow |
| PAUSED | INACTIVE | yellow |
| TERMINATED | INACTIVE | red |

### Macro State Mapping

| Full State | Simplified | Display Color |
|------------|------------|---------------|
| ORCHESTRATING | RUNNING | green |
| BOOTSTRAPPING | TRANSITIONING | blue |
| STABILIZING | TRANSITIONING | yellow |
| DRAINING | TRANSITIONING | red |

## API Compatibility

### Old RPC (To be removed/deprecated)

| Method | Status | Replacement |
|--------|--------|-------------|
| `RPCFSMStartProvider` | DEPRECATED | RPCFSMControl |
| `RPCFSMStopProvider` | DEPRECATED | RPCFSMControl |
| `RPCFSMPauseProvider` | DEPRECATED | RPCFSMControl |
| `RPCFSMResumeProvider` | DEPRECATED | RPCFSMControl |
| `RPCFSMStartAllProviders` | DEPRECATED | RPCFSMControl |
| `RPCFSMStopAllProviders` | DEPRECATED | RPCFSMControl |
| `RPCFSMPauseAllProviders` | DEPRECATED | RPCFSMControl |
| `RPCFSMResumeAllProviders` | DEPRECATED | RPCFSMControl |
| `RPCFSMGetProviderList` | DEPRECATED | RPCFSMGetTopology |
| `RPCFSMGetEtlTree` | DEPRECATED | RPCFSMGetTopology |

### New RPC

| Method | Description |
|--------|-------------|
| `RPCFSMControl` | Unified control (start/stop/pause/resume) |
| `RPCFSMGetTopology` | Unified topology with optional simplified states |
| `RPCFSMGetProviderCheckpoints` | Keep separate for detailed view |

## Migration Path

1. **Phase 1**: Add backend types and conversion functions
2. **Phase 2**: Add new unified RPC endpoints
3. **Phase 3**: Update frontend with unified client
4. **Phase 4**: Set simplified as default
5. **Phase 5**: Deprecate old endpoints (maintain for 1 release)
6. **Phase 6**: Remove deprecated endpoints (future)

## Testing Strategy

### 1. Unit Tests: State Conversion Functions

**File:** `pkg/meta/fsm/types_test.go`

```go
func TestToSimplifiedProviderState(t *testing.T) {
    tests := []struct {
        input    ProviderState
        expected SimplifiedProviderState
    }{
        {ProviderRunning, SimplifiedActive},
        {ProviderIdle, SimplifiedReady},
        {ProviderAcquiring, SimplifiedReady},
        {ProviderWaitingQuota, SimplifiedInactive},
        {ProviderWaitingBackoff, SimplifiedInactive},
        {ProviderPaused, SimplifiedInactive},
        {ProviderTerminated, SimplifiedInactive},
    }
    // ... test implementation
}

func TestToSimplifiedMacroState(t *testing.T) {
    // Similar table-driven tests
}
```

### 2. Integration Tests: RPC Handlers

**File:** `pkg/meta/fsm/rpc_test.go` (NEW)

```go
func TestRPCFSMControl(t *testing.T) {
    // Test each action type
    // Test single provider vs all providers
    // Test error cases
}

func TestRPCFSMGetTopology(t *testing.T) {
    // Test simplified vs full
    // Test checkpoint inclusion
}
```

### 3. Frontend Tests

**File:** `website/components/etl-topology-viewer.test.tsx`

```typescript
describe('ETLTopologyViewer', () => {
  it('displays simplified states by default');
  it('toggles to advanced mode');
  it('hides pause/resume in simplified mode');
  it('shows all controls in advanced mode');
});
```

### Test Coverage Targets

| Package | Target Coverage |
|---------|---------------|
| pkg/meta/fsm | > 80% |
| website/components | > 70% |

## Performance Considerations

- State translation is O(1) map lookup
- No additional database queries
- Optional simplified response reduces payload size by ~40%
