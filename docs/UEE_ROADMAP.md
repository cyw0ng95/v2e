# Unified ETL Engine (UEE) Implementation Roadmap

## Executive Summary

This roadmap details the complete implementation plan for migrating v2e from hardcoded synchronization loops to a resource-aware, hierarchical Finite State Machine (FSM) orchestration model. The architecture implements a Master-Slave pattern where the broker (Master) manages technical resources through permits, and the meta service (Slave) orchestrates domain-specific ETL logic.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Master (v2broker)                       │
│  - Resource Authority (Worker Permits)                       │
│  - Kernel Metrics (P99 Latency, Buffer Saturation)          │
│  - Adaptive Optimization & Revocation                        │
└──────────────────────┬──────────────────────────────────────┘
                       │ RPC: RequestPermits, ReleasePermits,
                       │      OnQuotaUpdate, GetKernelMetrics
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                      Slave (v2meta)                          │
│  - Macro FSM (Orchestration Manager)                        │
│  - Provider FSMs (CVE, CWE, CAPEC, ATT&CK Workers)         │
│  - URN-based Checkpointing                                  │
│  - Permit-Aware Execution                                   │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Status

### ✅ Phase 1: Foundation & URN System (COMPLETE)

**Status**: 100% Complete (4 commits, 20+ tests passing)

#### 1.1 URN Package (`pkg/urn`)
- **File**: `pkg/urn/urn.go`
- **Tests**: `pkg/urn/urn_test.go` (8 test cases)
- **Benchmarks**: `pkg/urn/urn_bench_test.go` (7 benchmarks)
- **Features**:
  - Hierarchical identifiers: `v2e::<provider>::<type>::<atomic_id>`
  - Providers: NVD, MITRE, SSG
  - Types: CVE, CWE, CAPEC, ATT&CK, SSG
  - Parse/New/String/Key/Equal operations
  - Validation and error handling
- **Performance**: 150-570 ns/op
- **Commit**: `d0f5af5`

#### 1.2 BoltDB Schema Extension (`pkg/meta/storage`)
- **File**: `pkg/meta/storage/storage.go`
- **Tests**: `pkg/meta/storage/storage_test.go` (6 test suites)
- **Buckets**:
  - `fsm_states` - Macro FSM states
  - `provider_states` - Provider FSM states
  - `checkpoints` - URN-validated checkpoints
  - `permits` - Permit allocation tracking
  - `sessions` - Legacy session data (preserved)
- **State Types**:
  - Macro: BOOTSTRAPPING, ORCHESTRATING, STABILIZING, DRAINING
  - Provider: IDLE, ACQUIRING, RUNNING, WAITING_QUOTA, WAITING_BACKOFF, PAUSED, TERMINATED
- **Commit**: `3c41641`

#### 1.3 RPC Contract Definitions
- **Files**:
  - `cmd/v2broker/service.md` - Broker RPC specs
  - `cmd/v2meta/service.md` - Meta RPC specs
- **Broker RPCs**:
  - `RPCRequestPermits` - Acquire worker slots
  - `RPCReleasePermits` - Return worker slots
  - `RPCOnQuotaUpdate` - Quota revocation event
  - `RPCGetKernelMetrics` - Performance metrics
- **Meta RPCs**:
  - `RPCGetEtlTree` - Hierarchical FSM tree
  - `RPCGetProviderCheckpoints` - Checkpoint retrieval
- **Commit**: `1429966`

#### 1.4 Permit Manager (`cmd/v2broker/core`)
- **File**: `cmd/v2broker/core/permits.go`
- **Tests**: `cmd/v2broker/core/permits_test.go` (6 test suites)
- **Features**:
  - Global permit pool (configurable, default 10)
  - Thread-safe request/release
  - Proportional revocation
  - Dynamic pool sizing
  - Comprehensive statistics
- **Commit**: `14d2a34`

---

### 🔄 Phase 2: Broker Master - Resource Control (IN PROGRESS)

**Status**: 25% Complete (PermitManager done)

**Remaining Work**:

#### 2.1 Kernel Metrics Collection (`cmd/v2broker/perf`)
- **File**: `cmd/v2broker/perf/metrics.go` (NEW)
- **Features**:
  - P99 latency tracking (rolling window)
  - Buffer saturation monitoring
  - Message rate/error rate calculation
  - Thread-safe metric collection
- **Integration**: Enhance existing `AdaptiveOptimizer`
- **Tests**: Unit tests for metric calculations
- **Estimate**: 4-6 hours

#### 2.2 Adaptive Optimizer Enhancements (`cmd/v2broker/perf`)
- **File**: `cmd/v2broker/perf/optimizer.go` (MODIFY)
- **Features**:
  - Integrate PermitManager
  - Add metric-driven revocation triggers
  - Threshold: P99 > 30-50ms → revoke permits
  - Event broadcasting to meta service
- **Tests**: Integration tests with PermitManager
- **Estimate**: 6-8 hours

#### 2.3 Broker RPC Handlers (`cmd/v2broker/main.go`)
- **File**: `cmd/v2broker/main.go` (MODIFY)
- **Handlers**:
  - `RPCRequestPermits` - Wire to PermitManager
  - `RPCReleasePermits` - Wire to PermitManager
  - `RPCGetKernelMetrics` - Return optimizer metrics
- **Broadcast**: Implement `RPCOnQuotaUpdate` event sending
- **Tests**: RPC handler tests
- **Estimate**: 4-6 hours

#### 2.4 Benchmarks & Performance Validation
- **File**: `cmd/v2broker/core/permits_bench_test.go` (NEW)
- **Benchmarks**:
  - Permit request/release throughput
  - Concurrent allocation scenarios
  - Revocation performance
- **Validation**: Ensure P99 < 20ms under load
- **Estimate**: 2-4 hours

**Total Phase 2 Estimate**: 16-24 hours

---

### 📋 Phase 3: Meta Slave - Hierarchical FSM (NOT STARTED)

**Status**: 0% Complete

**Dependencies**: Phase 2 complete

#### 3.1 FSM Framework (`pkg/meta/fsm`)
- **Files**:
  - `pkg/meta/fsm/macro.go` - Macro FSM interface
  - `pkg/meta/fsm/provider.go` - Provider FSM interface
  - `pkg/meta/fsm/events.go` - Event bubbling
  - `pkg/meta/fsm/transitions.go` - State validators
- **Features**:
  - MacroFSM interface with state transitions
  - ProviderFSM interface with permit awareness
  - Event bubbling from Provider → Macro
  - State transition validation
- **Tests**: FSM transition tests, event propagation tests
- **Estimate**: 8-12 hours

#### 3.2 Macro FSM Manager (`cmd/v2meta`)
- **File**: `cmd/v2meta/macro_fsm.go` (NEW)
- **Features**:
  - High-level orchestration state machine
  - Coordinate multiple provider FSMs
  - Aggregate provider events
  - Handle macro state transitions
- **Tests**: Macro FSM state transition tests
- **Estimate**: 6-8 hours

#### 3.3 Permit-Aware Executor (`cmd/v2meta`)
- **File**: `cmd/v2meta/executor.go` (NEW)
- **Features**:
  - Acquire permits before starting providers
  - Release permits on completion/pause
  - Handle RPCOnQuotaUpdate events
  - Transition to WAITING_QUOTA on revocation
- **Tests**: Permit coordination tests
- **Estimate**: 8-10 hours

#### 3.4 State Persistence Integration
- **File**: `cmd/v2meta/persistence.go` (NEW)
- **Features**:
  - Save/restore macro FSM state
  - Save/restore provider FSM states
  - Checkpoint management (every 100 items)
  - Use `pkg/meta/storage` APIs
- **Tests**: Persistence round-trip tests
- **Estimate**: 4-6 hours

#### 3.5 Auto-Recovery on Restart
- **File**: `cmd/v2meta/recovery.go` (NEW)
- **Features**:
  - Resume RUNNING jobs on startup
  - Maintain WAITING_BACKOFF jobs
  - Skip TERMINATED jobs
  - Restore checkpoint state
- **Tests**: Recovery scenario tests
- **Estimate**: 6-8 hours

**Total Phase 3 Estimate**: 32-44 hours

---

### 🔄 Phase 4: Provider Migration (NOT STARTED)

**Status**: 0% Complete

**Dependencies**: Phase 3 complete

#### 4.1 CVE Provider FSM (`cmd/v2meta/providers`)
- **File**: `cmd/v2meta/providers/cve_provider.go` (NEW)
- **Features**:
  - Implement ProviderFSM interface
  - Migrate fetch/parse/store logic from existing code
  - URN-based checkpoint generation
  - Error handling and retry logic
- **Tests**: CVE provider FSM tests
- **Estimate**: 8-10 hours

#### 4.2 CWE Provider FSM (`cmd/v2meta/providers`)
- **File**: `cmd/v2meta/providers/cwe_provider.go` (NEW)
- **Features**:
  - Implement ProviderFSM interface
  - Migrate import logic
  - URN checkpointing
- **Tests**: CWE provider FSM tests
- **Estimate**: 6-8 hours

#### 4.3 CAPEC Provider FSM (`cmd/v2meta/providers`)
- **File**: `cmd/v2meta/providers/capec_provider.go` (NEW)
- **Features**:
  - Implement ProviderFSM interface
  - Migrate XML parsing logic
  - URN checkpointing
- **Tests**: CAPEC provider FSM tests
- **Estimate**: 6-8 hours

#### 4.4 ATT&CK Provider FSM (`cmd/v2meta/providers`)
- **File**: `cmd/v2meta/providers/attack_provider.go` (NEW)
- **Features**:
  - Implement ProviderFSM interface
  - Migrate ATT&CK data import
  - URN checkpointing
- **Tests**: ATT&CK provider FSM tests
- **Estimate**: 6-8 hours

#### 4.5 RPC Handler Updates (`cmd/v2meta/main.go`)
- **File**: `cmd/v2meta/main.go` (MODIFY)
- **Features**:
  - Update session control RPCs to use FSM
  - Add backward compatibility layer
  - Deprecation warnings for old APIs
  - Wire new RPCGetEtlTree handler
  - Wire new RPCGetProviderCheckpoints handler
- **Tests**: RPC integration tests
- **Estimate**: 4-6 hours

**Total Phase 4 Estimate**: 30-40 hours

---

### 🎨 Phase 5: Frontend - ETL Engine Tab (NOT STARTED)

**Status**: 0% Complete

**Dependencies**: Phase 4 complete for full functionality

#### 5.1 TypeScript Types (`website/lib/types.ts`)
- **Features**:
  - Mirror Go structs in camelCase
  - MacroFSM state types
  - ProviderFSM state types
  - Kernel metrics types
  - ETL tree structure types
- **Estimate**: 2-3 hours

#### 5.2 RPC Client Methods (`website/lib/rpc-client.ts`)
- **Features**:
  - `getEtlTree()` method
  - `getKernelMetrics()` method
  - `getProviderCheckpoints(providerId)` method
  - Mock data for development mode
- **Estimate**: 3-4 hours

#### 5.3 React Query Hooks (`website/lib/hooks.ts`)
- **Features**:
  - `useEtlTree()` with polling
  - `useKernelMetrics()` with polling
  - `useProviderCheckpoints(providerId)`
  - 5-second refresh intervals
- **Estimate**: 2-3 hours

#### 5.4 ETL Tree Visualization (`website/components/etl`)
- **Files**:
  - `website/components/etl/etl-tree.tsx` - Tree component
  - `website/components/etl/macro-node.tsx` - Macro FSM node
  - `website/components/etl/provider-node.tsx` - Provider FSM node
- **Features**:
  - Hierarchical tree display (Macro → Providers)
  - State badges with color coding
  - Expand/collapse providers
  - Real-time state updates
- **Estimate**: 8-10 hours

#### 5.5 Kernel Metrics Gauges (`website/components/etl`)
- **File**: `website/components/etl/kernel-metrics.tsx`
- **Features**:
  - P99 latency gauge (threshold indicator at 30ms)
  - Buffer saturation gauge
  - Active workers count
  - Permit allocation pie chart
  - Message/error rate trends
- **Libraries**: Recharts or similar for gauges
- **Estimate**: 6-8 hours

#### 5.6 Job Control UI (`website/components/etl`)
- **File**: `website/components/etl/job-controls.tsx`
- **Features**:
  - Start/Pause/Stop buttons
  - Provider selection dropdown
  - Permit count adjuster
  - Resource allocation slider
- **Estimate**: 4-6 hours

#### 5.7 ETL Engine Page (`website/app/etl/page.tsx`)
- **Features**:
  - Integrate all ETL components
  - Layout with tree, metrics, and controls
  - Add to navigation in `layout.tsx`
  - Responsive design
- **Estimate**: 4-6 hours

#### 5.8 Testing & Polish
- **Tasks**:
  - Test with mock data
  - Test with live backend
  - UI polish and error states
  - Loading skeletons
- **Estimate**: 4-6 hours

**Total Phase 5 Estimate**: 33-46 hours

---

### 📚 Phase 6: Documentation & Testing (NOT STARTED)

**Status**: 0% Complete

**Dependencies**: Phases 2-5 complete

#### 6.1 README Updates (`README.md`)
- **Sections**:
  - Master-Slave architecture overview
  - URN system explanation
  - FSM hierarchy diagram
  - Permit management guide
  - Migration guide from old APIs
- **Estimate**: 4-6 hours

#### 6.2 Service Documentation
- **Files**:
  - Update `cmd/v2broker/service.md` with implementation notes
  - Update `cmd/v2meta/service.md` with implementation notes
  - Add examples and usage patterns
- **Estimate**: 2-3 hours

#### 6.3 Integration Tests (`tests/`)
- **Tests**:
  - Full sync workflow with permits
  - Permit revocation scenario
  - Auto-recovery after restart
  - Concurrent provider execution
  - Checkpoint resumption
- **Estimate**: 10-12 hours

#### 6.4 Performance Validation
- **Tasks**:
  - Load testing with permit contention
  - Verify P99 latency < 20ms under load
  - Measure checkpoint overhead
  - Profile permit allocation
  - Compare before/after benchmarks
- **Estimate**: 6-8 hours

#### 6.5 Monitoring & Observability
- **Tasks**:
  - Add metrics export for kernel stats
  - Structured logging for FSM transitions
  - Distributed tracing for permit lifecycle
  - Grafana dashboard examples (optional)
- **Estimate**: 6-8 hours

**Total Phase 6 Estimate**: 28-37 hours

---

## Total Implementation Estimate

| Phase | Status | Estimate (hours) | Dependencies |
|-------|--------|------------------|--------------|
| Phase 1 | ✅ Complete | N/A | None |
| Phase 2 | 🔄 25% Complete | 16-24 | Phase 1 |
| Phase 3 | ⏳ Not Started | 32-44 | Phase 2 |
| Phase 4 | ⏳ Not Started | 30-40 | Phase 3 |
| Phase 5 | ⏳ Not Started | 33-46 | Phase 4 |
| Phase 6 | ⏳ Not Started | 28-37 | Phases 2-5 |
| **Total Remaining** | | **139-191 hours** | |

**Note**: Estimates assume focused work without interruptions. Actual time may vary based on:
- Unforeseen technical challenges
- Integration issues with existing code
- Additional testing requirements
- Scope changes or feature additions

## Success Criteria

### Functional Requirements
- ✅ URN system operational for all resource types
- ✅ BoltDB schema supports FSM state persistence
- ✅ PermitManager correctly allocates/releases/revokes permits
- ⏳ Broker tracks kernel metrics (P99, buffer saturation)
- ⏳ Meta service uses hierarchical FSM for orchestration
- ⏳ All providers (CVE, CWE, CAPEC, ATT&CK) migrated to ProviderFSM
- ⏳ Frontend displays ETL tree and kernel metrics
- ⏳ Auto-recovery resumes jobs after restart

### Performance Requirements
- ⏳ P99 latency < 20ms under normal load
- ⏳ P99 latency < 50ms under peak load (before revocation)
- ⏳ Checkpoint overhead < 1% of total processing time
- ⏳ Permit allocation latency < 1ms

### Quality Requirements
- ✅ All new code has unit tests (>80% coverage)
- ⏳ Integration tests cover critical workflows
- ⏳ No regression in existing functionality
- ⏳ All tests pass in CI/CD pipeline
- ⏳ Documentation updated and accurate

## Risk Assessment

### High Risk Areas
1. **State Migration**: Migrating from existing session model to FSM states
   - **Mitigation**: Maintain backward compatibility layer during transition
2. **Performance Regression**: New architecture may introduce latency
   - **Mitigation**: Continuous benchmarking, early optimization
3. **Complexity**: Hierarchical FSM adds conceptual complexity
   - **Mitigation**: Clear documentation, examples, and diagrams

### Medium Risk Areas
1. **Frontend Integration**: ETL tab requires live backend for testing
   - **Mitigation**: Mock data mode for development
2. **Permit Contention**: High contention may cause starvation
   - **Mitigation**: Fairness algorithms in PermitManager

### Low Risk Areas
1. **URN System**: Simple, well-tested, no dependencies
2. **BoltDB Schema**: Additive changes, no breaking changes

## Next Immediate Steps

Based on current progress (Phase 1 complete, Phase 2 25% complete):

1. **Kernel Metrics Collection** (2-3 hours)
   - Implement `metrics.go` with P99 tracker
   - Add buffer saturation monitoring

2. **Optimizer Enhancement** (4-6 hours)
   - Integrate PermitManager into optimizer
   - Add revocation triggers

3. **RPC Handler Wiring** (3-4 hours)
   - Wire permit RPCs in broker main.go
   - Test RPC functionality

4. **Phase 2 Validation** (2-3 hours)
   - Run benchmarks
   - Validate P99 < 20ms

**Total for completing Phase 2**: 11-16 hours

After Phase 2, proceed to Phase 3 FSM framework implementation.

---

## Appendix: Architecture Diagrams

### Master-Slave Interaction Flow

```
┌─────────────┐                    ┌─────────────┐
│   v2meta    │                    │  v2broker   │
│   (Slave)   │                    │  (Master)   │
└──────┬──────┘                    └──────┬──────┘
       │                                  │
       │  RPCRequestPermits(5)            │
       │─────────────────────────────────>│
       │                                  │ PermitManager
       │  Response: {granted: 5}          │ allocates
       │<─────────────────────────────────│
       │                                  │
       │  [Provider FSM: ACQUIRING -> RUNNING]
       │                                  │
       │  ... work in progress ...        │
       │                                  │ KernelMetrics
       │                                  │ P99 > 50ms!
       │                                  │
       │  RPCOnQuotaUpdate(revoked: 2)    │ Revoke permits
       │<─────────────────────────────────│
       │                                  │
       │  [Provider FSM: RUNNING -> WAITING_QUOTA]
       │                                  │
       │  RPCReleasePermits(3)            │
       │─────────────────────────────────>│
       │                                  │
       │  Response: {available: 5}        │
       │<─────────────────────────────────│
       │                                  │
       │  [Provider FSM: WAITING_QUOTA -> TERMINATED]
       │                                  │
```

### FSM State Transitions

```
Macro FSM:
  BOOTSTRAPPING ──┐
                  ├──> ORCHESTRATING ──┐
                  │                    ├──> STABILIZING ──> DRAINING
                  └────────────────────┘

Provider FSM:
  IDLE ──> ACQUIRING ──> RUNNING ──┬──> PAUSED ──> RUNNING
                                   │
                                   ├──> WAITING_QUOTA ──> RUNNING
                                   │
                                   ├──> WAITING_BACKOFF ──> RUNNING
                                   │
                                   └──> TERMINATED
```

---

**Document Version**: 1.0  
**Last Updated**: 2026-02-04  
**Author**: GitHub Copilot  
**Status**: Living Document - Updated as implementation progresses
