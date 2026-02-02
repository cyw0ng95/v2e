# Meta Service State Machine & RPC Control Logic Audit

**Date:** 2026-02-03
**Scope:** Full audit of cmd/meta and pkg/cve/taskflow state machines and RPC control logic

## Executive Summary

The audit identified **9 critical issues** in the meta service state machine and RPC control logic:

1. Race condition in `StartTyped()` - activeRun set after lock release
2. Unused granular states (fetching, processing, saving, etc.)
3. Pause/Stop don't validate run exists in store
4. Resume allows double-resume
5. executeJob defer cleanup clears activeRun prematurely
6. CircuitBreaker defined but never used
7. No exponential backoff for API rate limits
8. Inconsistent RPC timeouts across services
9. StateMachineController exists but unused

## Detailed Findings

### Issue 1: Race Condition in StartTyped()

**Location:** `pkg/cve/taskflow/executor.go:50-86`

**Problem:**
```go
func (e *JobExecutor) StartTyped(...) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // Check for active run
    activeRun, err := e.runStore.GetActiveRun()
    if activeRun != nil {
        return fmt.Errorf("job already running")
    }

    // Create run and update state
    run, err := e.runStore.CreateRun(...)
    e.runStore.UpdateState(runID, StateRunning)  // Still holding lock

    // BUG: Lock released here by defer
    e.activeRun = run  // Set AFTER lock released
    e.cancelFunc = cancel

    go e.executeJob(jobCtx, runID)  // Goroutine starts
    // Between lock release and activeRun set, another caller could see nil
}
```

**Impact:** Two concurrent `StartTyped()` calls could both succeed, violating the single-active-run policy.

**Fix:**
```go
func (e *JobExecutor) StartTyped(...) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // Check in-memory activeRun first (faster)
    if e.activeRun != nil {
        return fmt.Errorf("job already running: %s", e.activeRun.ID)
    }

    // Double-check persisted store
    activeRun, err := e.runStore.GetActiveRun()
    if activeRun != nil {
        return fmt.Errorf("job already running: %s", activeRun.ID)
    }

    // Create run
    run, err := e.runStore.CreateRun(runID, startIndex, resultsPerBatch, dataType)
    if err != nil {
        return fmt.Errorf("failed to create run: %w", err)
    }

    // Set activeRun BEFORE any goroutine starts
    e.activeRun = run

    // Then update state
    if err := e.runStore.UpdateState(runID, StateRunning); err != nil {
        e.activeRun = nil  // Rollback on error
        return fmt.Errorf("failed to update state: %w", err)
    }

    // Create context and start goroutine (still holding lock, but that's OK)
    jobCtx, cancel := context.WithCancel(ctx)
    e.cancelFunc = cancel
    go e.executeJob(jobCtx, runID)

    return nil
}
```

### Issue 2: Pause/Stop Don't Validate Run Exists

**Location:** `pkg/cve/taskflow/executor.go:122-182`

**Problem:**
```go
func (e *JobExecutor) Pause(runID string) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // Only checks activeRun, doesn't verify run exists in store
    if e.activeRun == nil || e.activeRun.ID != runID {
        return fmt.Errorf("run not active: %s", runID)
    }

    // If run was deleted externally, this could operate on non-existent run
    run, err := e.runStore.GetRun(runID)  // Only fetched AFTER activeRun check
    ...
}
```

**Impact:** Could operate on non-existent runs or cause state inconsistency.

**Fix:**
```go
func (e *JobExecutor) Pause(runID string) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // First verify run exists in store
    run, err := e.runStore.GetRun(runID)
    if err != nil {
        return fmt.Errorf("run not found: %w", err)
    }

    // Validate state
    if run.State != StateRunning {
        return fmt.Errorf("run is not running (current state: %s)", run.State)
    }

    // Then verify we own it
    if e.activeRun == nil || e.activeRun.ID != runID {
        return fmt.Errorf("run not active: %s", runID)
    }

    // ... rest of pause logic
}
```

### Issue 3: Resume Allows Double-Resume

**Location:** `pkg/cve/taskflow/executor.go:88-120`

**Problem:**
```go
func (e *JobExecutor) Resume(ctx context.Context, runID string) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // BUG: Doesn't check if e.activeRun is already set
    run, err := e.runStore.GetRun(runID)
    if err != nil {
        return fmt.Errorf("failed to get run: %w", err)
    }

    if run.State != StatePaused {
        return fmt.Errorf("run is not paused")
    }

    // Starts goroutine without checking if one already exists
    e.activeRun = run
    go e.executeJob(jobCtx, runID)
}
```

**Impact:** Multiple concurrent Resume calls could start multiple goroutines.

**Fix:**
```go
func (e *JobExecutor) Resume(ctx context.Context, runID string) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // Validate no active run
    if e.activeRun != nil {
        return fmt.Errorf("cannot resume: another job is active: %s", e.activeRun.ID)
    }

    // ... rest of logic
}
```

### Issue 4: executeJob Defer Cleanup Premature

**Location:** `pkg/cve/taskflow/executor.go:245-251`

**Problem:**
```go
func (e *JobExecutor) executeJob(ctx context.Context, runID string) {
    defer func() {
        e.mu.Lock()
        e.activeRun = nil      // BUG: Always clears, even on pause
        e.cancelFunc = nil
        e.mu.Unlock()
    }()
    // When job pauses, activeRun should stay set for Resume to work
}
```

**Impact:** Paused jobs cannot be resumed because activeRun is cleared.

**Fix:**
```go
func (e *JobExecutor) executeJob(ctx context.Context, runID string) {
    // Use done channel for cleanup coordination
    doneChan := make(chan struct{})
    defer close(doneChan)

    for {
        select {
        case <-ctx.Done():
            // Only clear activeRun on cancellation
            e.mu.Lock()
            e.activeRun = nil
            e.mu.Unlock()
            return
        // ... rest of loop
        }
    }
}
```

### Issue 5: No Rate Limit Backoff

**Location:** `pkg/cve/taskflow/executor.go:386-397`

**Problem:**
```go
if fetchErr != nil {
    e.logger.Warn("Fetch failed: %v", fetchErr)
    e.runStore.UpdateProgress(runID, 0, 0, 1)

    // Always waits 5 seconds, even for rate limits
    select {
    case <-ctx.Done():
        return
    case <-time.After(5 * time.Second):
        continue
    }
}
```

**Impact:** Rate limit errors (429) retry too quickly, potentially getting blocked.

**Fix:**
```go
func isRateLimitError(err error) bool {
    if err == nil {
        return false
    }
    errStr := strings.ToLower(err.Error())
    return strings.Contains(errStr, "rate limit") ||
           strings.Contains(errStr, "429") ||
           strings.Contains(errStr, "too many requests")
}

if fetchErr != nil {
    e.logger.Warn("Fetch failed: %v", fetchErr)

    // Check if we should give up
    if shouldGiveUp(fetchErr) {
        e.runStore.UpdateState(runID, StateFailed)
        e.runStore.SetError(runID, fetchErr.Error())
        return
    }

    // Calculate appropriate backoff
    var backoff time.Duration
    if isRateLimitError(fetchErr) {
        backoff = 30 * time.Second  // Longer for rate limits
    } else {
        backoff = time.Duration(1<<uint(retryCount)) * time.Second  // Exponential
    }

    select {
    case <-ctx.Done():
        return
    case <-time.After(backoff):
        continue
    }
}
```

### Issue 6: CircuitBreaker Unused

**Location:** `pkg/cve/taskflow/error_handling.go:35-107`

**Problem:**
- CircuitBreaker type is fully implemented
- Never instantiated or used in JobExecutor
- No integration with fetch/store operations

**Impact:** No protection against cascading failures when remote/local services fail.

**Fix:** Integrate CircuitBreaker into RPC calls:

```go
// In JobExecutor
type JobExecutor struct {
    // ... existing fields
    remoteCircuitBreaker *CircuitBreaker
    localCircuitBreaker  *CircuitBreaker
}

func NewJobExecutor(...) *JobExecutor {
    return &JobExecutor{
        // ... existing init
        remoteCircuitBreaker: NewCircuitBreaker(5, 60*time.Second),
        localCircuitBreaker:  NewCircuitBreaker(10, 30*time.Second),
    }
}

func (e *JobExecutor) fetchWithBreaker(ctx context.Context, params interface{}) error {
    return e.remoteCircuitBreaker.Call(func() error {
        result, err := e.rpcInvoker.InvokeRPC(ctx, "remote", "RPCFetchCVEs", params)
        if err != nil {
            return err
        }
        // Check response...
        return nil
    })
}
```

### Issue 7: Unused Granular States

**Location:** `pkg/cve/taskflow/state.go:15-23`

**Problem:**
- States defined: `initializing`, `fetching`, `processing`, `saving`, `validating`, `recovering`, `rolling_back`
- Executor never uses these - only uses `running`, `paused`, `completed`, `failed`, `stopped`

**Impact:** State machine is more complex than actual implementation.

**Fix Options:**
1. **Remove unused states** - Simplify to only what's used
2. **Actually use them** - Update executor to transition through granular states
3. **Document them as reserved** - For future implementation

**Recommendation:** Option 1 - Remove unused states to reduce confusion.

### Issue 8: Inconsistent RPC Timeouts

**Locations:**
- `cmd/meta/main.go:237` - 60 seconds
- `cmd/sysmon/main.go:22` - 30 seconds (rpc.DefaultRPCTimeout)
- `pkg/cve/remote/fetcher.go:30` - 30 seconds (http client)

**Problem:** No centralized timeout configuration.

**Impact:**
- Long-running operations may timeout unexpectedly
- Quick operations wait too long on failure

**Fix:**
```go
// In pkg/rpc/client.go
type TimeoutOperation int

const (
    TimeoutQuickCheck TimeoutOperation = iota
    TimeoutDefault
    TimeoutBulkOperation
    TimeoutImport
)

var timeoutDefaults = map[TimeoutOperation]time.Duration{
    TimeoutQuickCheck:     10 * time.Second,
    TimeoutDefault:        30 * time.Second,
    TimeoutBulkOperation:  60 * time.Second,
    TimeoutImport:         300 * time.Second,
}

func GetTimeout(op TimeoutOperation) time.Duration {
    if t, ok := timeoutDefaults[op]; ok {
        return t
    }
    return DefaultRPCTimeout
}
```

### Issue 9: StateMachineController Unused

**Location:** `pkg/cve/taskflow/error_handling.go:109-142`

**Problem:**
- StateMachineController.TransitTo() wraps state transitions with validation
- Never used - code directly calls RunStore.UpdateState()

**Impact:** Inconsistent state transition enforcement.

**Fix:**
```go
// Option 1: Use StateMachineController in JobExecutor
func (e *JobExecutor) StartTyped(...) error {
    // ...
    if err := e.stateMachine.TransitTo(runID, StateRunning); err != nil {
        return err
    }
}

// Option 2: Remove StateMachineController and inline validation
func (s *RunStore) UpdateState(runID string, state JobState) error {
    run, err := s.GetRun(runID)
    if err != nil {
        return err
    }

    if !run.State.CanTransitionTo(state) {
        return fmt.Errorf("invalid state transition: %s -> %s", run.State, state)
    }

    run.State = state
    run.UpdatedAt = time.Now()
    return s.saveRun(run)
}
```

## Recommended Implementation Plan

### Phase 1: Critical Race Condition Fixes (High Priority)

1. Fix `StartTyped()` race condition
2. Fix `Resume()` double-resume
3. Fix `Pause()` to validate run exists
4. Fix `Stop()` to validate run exists
5. Add `doneChan` for proper goroutine cleanup

**Files to modify:**
- `pkg/cve/taskflow/executor.go`

### Phase 2: Error Handling Improvements (High Priority)

1. Add rate limit detection and exponential backoff
2. Add `shouldGiveUp()` for unrecoverable errors
3. Integrate CircuitBreaker for RPC calls

**Files to modify:**
- `pkg/cve/taskflow/executor.go`
- `pkg/cve/taskflow/error_handling.go`

### Phase 3: State Machine Cleanup (Medium Priority)

1. Remove unused granular states OR implement them
2. Remove unused StateMachineController OR use it consistently
3. Add state transition unit tests

**Files to modify:**
- `pkg/cve/taskflow/state.go`
- `pkg/cve/taskflow/error_handling.go`
- Add `pkg/cve/taskflow/executor_state_test.go`

### Phase 4: RPC Timeout Standardization (Low Priority)

1. Define timeout operations
2. Update all RPC clients to use consistent timeouts
3. Document timeout behavior

**Files to modify:**
- `pkg/rpc/client.go`
- `cmd/meta/main.go`
- `cmd/sysmon/main.go`
- `cmd/access/` (if applicable)

## Test Coverage Needed

```go
// Critical test cases
func TestJobExecutor_ConcurrentStartPrevention(t *testing.T)
func TestJobExecutor_PauseResumeStateTransitions(t *testing.T)
func TestJobExecutor_DoubleResumePrevention(t *testing.T)
func TestJobExecutor_StopFromRunning(t *testing.T)
func TestJobExecutor_StopFromPaused(t *testing.T)
func TestJobExecutor_RateLimitBackoff(t *testing.T)
func TestJobExecutor_RecoveryAfterCrash(t *testing.T)
func TestJobExecutor_StateTransitionValidation(t *testing.T)
func TestJobExecutor_CircuitBreakerTrips(t *testing.T)
```

## Risk Assessment

| Issue | Severity | Likelihood | Impact | Priority |
|-------|----------|------------|--------|----------|
| Race condition in StartTyped | High | Medium | Data corruption, state inconsistency | P0 |
| Double-resume allowed | High | Low | Multiple goroutines, resource leak | P0 |
| Pause/Stop don't validate | Medium | Low | Operating on non-existent runs | P1 |
| No rate limit backoff | Medium | High | API blocking, wasted requests | P1 |
| Unused states | Low | N/A | Code confusion | P2 |
| CircuitBreaker unused | Medium | Medium | Cascading failures | P1 |
| Inconsistent timeouts | Low | Medium | Unexpected failures | P2 |

## Success Criteria

After implementing fixes:
1. All race conditions eliminated (run `go test -race`)
2. State transitions validated before execution
3. Rate limits handled with appropriate backoff
4. CircuitBreaker prevents cascading failures
5. Test coverage > 80% for executor code

## Notes

- All changes should be backward compatible
- Existing job runs should be recoverable after upgrade
- Consider adding metrics for state transitions and failures
