# UEE Provider Integration Analysis

## Executive Summary

This document analyzes the Unified ETL Engine (UEE) provider implementation and identifies critical gaps between the FSM infrastructure and actual provider implementations.

## Critical Issues Found

### 1. **Provider Implementations Do Not Use FSM Infrastructure**

**Problem**: All data source providers (CVE, CWE, CAPEC, ATT&CK) implement the `DataSourceProvider` interface but return hardcoded `ProviderIdle` states and stub implementations for state transitions.

**Affected Files**:
- `pkg/cve/provider/cve_provider.go`
- `pkg/cwe/provider/cwe_provider.go`
- `pkg/capec/provider/capec_provider.go`
- `pkg/attack/provider/attack_provider.go`

**Evidence**:
```go
// All providers have this pattern:
func (p *CVEProvider) GetState() fsm.ProviderState {
    return fsm.ProviderIdle  // HARDCODED!
}

func (p *CVEProvider) Start() error {
    return nil  // STUB - no state transition!
}

func (p *CVEProvider) Transition(newState fsm.ProviderState) error {
    // TODO: Implement state transition logic with validation
    return nil  // STUB - does nothing!
}

func (p *CVEProvider) Pause() error {
    return nil  // STUB
}

func (p *CVEProvider) Resume() error {
    return nil  // STUB
}

func (p *CVEProvider) OnQuotaRevoked(revokedCount int) error {
    return nil  // STUB
}

func (p *CVEProvider) OnQuotaGranted(grantedCount int) error {
    return nil  // STUB
}

func (p *CVEProvider) OnRateLimited(retryAfter time.Duration) error {
    return nil  // STUB
}
```

**Impact**:
- FSM state transitions are completely bypassed
- No state persistence to BoltDB
- No checkpointing support
- No quota management
- No rate limiting backoff handling
- Auto-recovery on restart is impossible
- FSM observability is broken

### 2. **Missing Storage Integration**

**Problem**: Providers don't initialize `BaseProviderFSM` with storage or use it for state persistence.

**Expected Pattern**:
```go
type CVEProvider struct {
    *fsm.BaseProviderFSM  // SHOULD embed this!
    fetcher              *remote.Fetcher
}

func NewCVEProvider(apiKey string) (*CVEProvider, error) {
    // Create base FSM with storage
    base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
        ID:           "cve",
        ProviderType:  "cve",
        Storage:       store,  // MISSING!
        Executor:      p.execute,
    })

    return &CVEProvider{
        BaseProviderFSM: base,
        fetcher:        fetcher,
    }, nil
}
```

**Actual Implementation**:
```go
type CVEProvider struct {
    config       *provider.ProviderConfig  // Custom config, not FSM!
    rateLimiter  *provider.RateLimiter  // Custom rate limiter!
    progress     *provider.ProviderProgress
    // NO BaseProviderFSM!
}
```

### 3. **Missing FSM Integration in Meta Service**

**Problem**: The `cmd/v2meta` package has two provider initialization systems that are not integrated:

1. `provider_handlers.go` - Uses custom `DataSourceProvider` interface with stub implementations
2. `provider_registry.go` - Uses `ProviderRegistry` to manage providers

**Neither integrates with `pkg/meta/fsm/BaseProviderFSM`**.

**Evidence from `cmd/v2meta/provider_handlers.go`**:
```go
// Global registry of data source providers
var providers []provider.DataSourceProvider  // Custom type, not FSM!

func initProviders(logger *common.Logger) error {
    // Creates custom providers, not FSM providers
    cveProv, err := cveprovider.NewCVEProvider("")  // Returns custom type!
    providerList = append(providerList, cveProv)  // Not FSM!
}
```

### 4. **Missing MacroFSM Integration**

**Problem**: There's no integration between `MacroFSMManager` and the actual provider instances.

**Expected**:
```go
// In cmd/v2meta/main.go or similar
func startUEE() error {
    store, _ := storage.NewStore("session.db", logger)

    // Create macro FSM
    macro, _ := fsm.NewMacroFSMManager("uee-macro", store)

    // Create providers with FSM support
    cveProvider, _ := cveprovider.NewCVEProviderWithFSM(apiKey, store)
    cweProvider, _ := cweprovider.NewCWEProviderWithFSM(store)
    capecProvider, _ := capecprovider.NewCAPECProviderWithFSM(store)
    attackProvider, _ := attackprovider.NewATTACKProviderWithFSM(store)

    // Add to macro FSM
    macro.AddProvider(cveProvider)
    macro.AddProvider(cweProvider)
    macro.AddProvider(capecProvider)
    macro.AddProvider(attackProvider)

    // Start macro orchestration
    macro.Transition(fsm.MacroOrchestrating)
}
```

**Actual**: No macro FSM instantiation or provider registration found in meta service.

### 5. **Missing Checkpointing Implementation**

**Problem**: Providers don't save checkpoints with URN identifiers.

**Expected in Provider**:
```go
func (p *CVEProvider) processCVE(cveID string) error {
    urn := urn.MustParse(fmt.Sprintf("v2e::nvd::cve::%s", cveID))

    // Process CVE...

    // Save checkpoint
    return p.SaveCheckpoint(urn, true, "")
}
```

**Actual**: No checkpointing logic in any provider implementation.

## Required Fixes

### Priority 1: Refactor Providers to Use BaseProviderFSM

**Action Items**:

1. **CVEProvider** (`pkg/cve/provider/cve_provider.go`):
   ```go
   type CVEProvider struct {
       *fsm.BaseProviderFSM
       fetcher *remote.Fetcher
   }

   func NewCVEProvider(apiKey string, storage *storage.Store) (*CVEProvider, error) {
       fetcher, err := remote.NewFetcher(apiKey)
       if err != nil {
           return nil, err
       }

       base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
           ID:           "cve",
           ProviderType:  "cve",
           Storage:       storage,
           Executor:      p.execute,
       })
       if err != nil {
           return nil, err
       }

       return &CVEProvider{
           BaseProviderFSM: base,
           fetcher:        fetcher,
       }, nil
   }

   func (p *CVEProvider) execute() error {
       // Fetch and process CVEs
       for _, cve := range cveList {
           urn := urn.MustParse(fmt.Sprintf("v2e::nvd::cve::%s", cve.ID))

           // Process CVE...

           // Save checkpoint
           if err := p.SaveCheckpoint(urn, true, ""); err != nil {
               return err
           }
       }
       return nil
   }
   ```

2. **Repeat for**: CWEProvider, CAPECProvider, ATTACKProvider, SSGProvider, ASVSProvider

### Priority 2: Update Provider Registry to Use FSM

**Action**: Modify `cmd/v2meta/provider_handlers.go`:

```go
var providers []fsm.ProviderFSM  // Change from []provider.DataSourceProvider

func initProviders(logger *common.Logger, storage *storage.Store) error {
    store, err := storage.NewStore("session.db", logger)
    if err != nil {
        return err
    }

    var providerList []fsm.ProviderFSM

    // Create FSM-based providers
    cveProv, err := cveprovider.NewCVEProviderWithFSM("", store)
    if err != nil {
        return fmt.Errorf("failed to create CVE provider: %w", err)
    }
    providerList = append(providerList, cveProv)

    // ... repeat for other providers

    providers = providerList

    // Initialize each provider (loads state from storage)
    for _, p := range providers {
        if err := p.Initialize(context.Background()); err != nil {
            return fmt.Errorf("failed to initialize provider %s: %w", p.GetID(), err)
        }
    }

    logger.Info("All providers initialized successfully")
    return nil
}
```

### Priority 3: Integrate MacroFSM Manager

**Action**: Add macro FSM initialization to `cmd/v2meta/main.go`:

```go
var (
    macroFSM *fsm.MacroFSMManager
    providers []fsm.ProviderFSM
)

func main() {
    // ...

    // Initialize storage
    store, err := storage.NewStore(config.SessionDBPath, logger)
    if err != nil {
        logger.Fatal("Failed to create storage: %v", err)
    }
    defer store.Close()

    // Initialize macro FSM
    macroFSM, err = fsm.NewMacroFSMManager("uee-orchestrator", store)
    if err != nil {
        logger.Fatal("Failed to create macro FSM: %v", err)
    }
    defer macroFSM.Stop()

    // Initialize providers
    if err := initProviders(logger, store); err != nil {
        logger.Fatal("Failed to initialize providers: %v", err)
    }

    // Register providers with macro FSM
    for _, provider := range providers {
        if err := macroFSM.AddProvider(provider); err != nil {
            logger.Error("Failed to add provider %s: %v", provider.GetID(), err)
        }
    }

    // Register RPC handlers
    registerFSMRPCHandlers(sp, logger)
}

func registerFSMRPCHandlers(sp *subprocess.Subprocess, logger *common.Logger) {
    // ETL monitoring
    sp.RegisterHandler("RPCGetEtlTree", handleGetEtlTree)
    sp.RegisterHandler("RPCGetProviderCheckpoints", handleGetProviderCheckpoints)
    sp.RegisterHandler("RPCStartProvider", handleStartProvider)
    sp.RegisterHandler("RPCStopProvider", handleStopProvider)
    sp.RegisterHandler("RPCPauseProvider", handlePauseProvider)
    sp.RegisterHandler("RPCResumeProvider", handleResumeProvider)
}

func handleGetEtlTree(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
    stats := macroFSM.GetStats()

    providers := make([]map[string]interface{}, 0)
    for _, provider := range macroFSM.GetProviders() {
        providers = append(providers, provider.GetStats())
    }

    return subprocess.NewSuccessResponse(msg, map[string]interface{}{
        "macro_fsm": stats,
        "providers": providers,
    })
}
```

### Priority 4: Implement Auto-Recovery

**Action**: Add recovery logic to meta service startup:

```go
func recoverRunningProviders() error {
    // Load all provider states from storage
    providerStates, err := store.ListProviderStates()
    if err != nil {
        return err
    }

    for _, state := range providerStates {
        provider := getProvider(state.ID)
        if provider == nil {
            continue
        }

        // Recovery strategy per state
        switch fsm.ProviderState(state.State) {
        case fsm.ProviderRunning:
            // Resume execution
            logger.Info("Resuming provider %s from RUNNING state", state.ID)
            if err := provider.Resume(); err != nil {
                logger.Error("Failed to resume provider %s: %v", state.ID, err)
            }

        case fsm.ProviderWaitingQuota:
            // Transition to ACQUIRING to retry
            logger.Info("Retrying quota acquisition for provider %s", state.ID)
            if err := provider.Transition(fsm.ProviderAcquiring); err != nil {
                logger.Error("Failed to transition provider %s: %v", state.ID, err)
            }

        case fsm.ProviderPaused:
            // Keep paused - manual resume needed
            logger.Info("Provider %s is paused, keeping paused", state.ID)

        case fsm.ProviderWaitingBackoff:
            // Maintain state, auto-retry will kick in
            logger.Info("Provider %s in backoff, maintaining state", state.ID)

        case fsm.ProviderTerminated, fsm.ProviderIdle:
            // Skip recovery
            logger.Info("Provider %s in %s state, skipping recovery", state.ID, state.State)
        }
    }

    return nil
}
```

### Priority 5: Add Integration Tests

**Status**: ✅ Created `pkg/meta/fsm/provider_integration_test.go` with comprehensive tests:

- `TestProviderFSM_FullLifecycle` - Complete lifecycle with state persistence
- `TestProviderFSM_QuotaRevocation` - Quota revocation handling
- `TestProviderFSM_RateLimiting` - Rate limit backoff handling
- `TestProviderFSM_CheckpointPersistence` - Checkpoint saving and recovery
- `TestProviderFSM_InvalidTransitions` - State transition validation
- `TestProviderFSM_StateRecovery` - State recovery after restart
- `TestProviderFSM_MultipleProviders` - Multiple providers coordination
- `TestProviderFSM_ConcurrentOperations` - Concurrent access safety
- `TestProviderFSM_ContextCancellation` - Graceful shutdown

**Next Steps**:
1. Run integration tests with `./build.sh -t`
2. Fix any failing tests
3. Add tests to verify macro-provider integration

## Testing Strategy

### Unit Tests
- ✅ `pkg/meta/fsm/provider_test.go` - BaseProviderFSM tests
- ✅ `pkg/meta/fsm/macro_test.go` - MacroFSMManager tests
- ✅ `pkg/meta/fsm/types_test.go` - State validation tests
- ✅ `pkg/meta/fsm/provider_integration_test.go` - Integration tests (NEW)

### Integration Tests
**TODO**: Add end-to-end integration tests:

```go
// tests/integration/uee_integration_test.py
def test_full_uee_workflow():
    """Test complete UEE workflow from start to completion"""
    # 1. Start meta service
    # 2. Start all providers via RPCStartProvider
    # 3. Verify FSM states via RPCGetEtlTree
    # 4. Simulate quota revocation
    # 5. Simulate rate limiting
    # 6. Verify checkpointing via RPCGetProviderCheckpoints
    # 7. Stop all providers
    # 8. Restart service and verify auto-recovery
    pass
```

### Manual Testing
Use `./build.sh -r` to start full system and verify:

1. **FSM Transitions**: Monitor logs for `[FSM_TRANSITION]` messages
2. **Macro FSM Transitions**: Monitor for `[MACRO_FSM_TRANSITION]` messages
3. **State Persistence**: Stop and restart, verify states are recovered
4. **Checkpointing**: Verify checkpoints are saved to BoltDB
5. **Auto-Recovery**: Stop while provider is RUNNING, restart and verify resume

## Stability and Maintainability Recommendations

### 1. **FSM State Validation**
All state transitions should validate the current state before transitioning. The `ValidateProviderTransition` and `ValidateMacroTransition` functions should be called on every transition.

### 2. **State Persistence**
Every state change must be persisted to BoltDB immediately with rollback on failure. This ensures the system can recover from crashes at any point.

### 3. **Logging Requirements**
Per service.md Requirement 6, all FSM transitions must be logged:
- Provider FSM: `[FSM_TRANSITION]` messages with URN, processed count, error count
- Macro FSM: `[MACRO_FSM_TRANSITION]` messages with provider count, provider states

### 4. **Graceful Degradation**
The system should handle partial failures gracefully:
- If one provider fails, others should continue
- If storage write fails, state transition should be rolled back
- If quota is revoked, provider should pause and retry

### 5. **Observability**
Add metrics and monitoring:
- FSM state transition latency
- Time spent in each state
- Checkpoint frequency and success rate
- Quota acquisition retry count
- Rate limit backoff frequency

### 6. **Testing Coverage**
- Add unit tests for all state transitions
- Add integration tests for multi-provider coordination
- Add chaos tests for crash recovery
- Add load tests for concurrent operations

## Conclusion

The UEE FSM infrastructure is well-designed (`pkg/meta/fsm/`) but **completely disconnected** from the actual provider implementations. To make the system production-ready:

1. **Immediate**: Refactor all providers to embed `BaseProviderFSM`
2. **Immediate**: Integrate `MacroFSMManager` in meta service
3. **Short-term**: Implement checkpointing and state persistence in providers
4. **Short-term**: Add auto-recovery logic
5. **Medium-term**: Add comprehensive integration tests
6. **Long-term**: Add observability and monitoring

Without these changes, the UEE infrastructure is non-functional. The FSM code exists but is not used, and providers operate as simple fetchers without state management, persistence, or recovery capabilities.
