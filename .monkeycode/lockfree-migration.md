# Lock-Free Router Migration Strategy

## Overview
This document describes the strategy for migrating the broker's map-based routing implementation to a lock-free approach using Go's `sync.Map` and `sync/atomic` operations.

## Current State

### Data Structures
The `Broker` struct in `cmd/v2broker/core/broker.go` currently uses mutex-protected maps:

```go
type Broker struct {
    processes        map[string]*Process
    mu               sync.RWMutex

    rpcEndpoints    map[string][]string
    endpointsMu     sync.RWMutex

    pendingRequests map[string]*PendingRequest
    pendingMu       sync.RWMutex
    // ... other fields
}
```

### Lock Patterns
- Read operations use `mu.RLock()` / `mu.RUnlock()`
- Write operations use `mu.Lock()` / `mu.Unlock()`
- All map accesses are protected by these locks

## Proposed Lock-Free Implementation

### 1. Migrate to sync.Map

Replace `map[string]*Process` with `sync.Map`:

```go
type Broker struct {
    processes        sync.Map  // maps string -> *Process

    rpcEndpoints    sync.Map  // maps string -> []string
    pendingRequests sync.Map  // maps string -> *PendingRequest

    // Remove: mu, endpointsMu, pendingMu
    correlationSeq  uint64  // Keep atomic operations for this
    // ... other fields
}
```

### 2. sync.Map API Mapping

| Operation | Current (map+lock) | New (sync.Map) |
|-----------|---------------------|----------------|
| Read | `b.mu.RLock(); v, ok := b.processes[key]; b.mu.RUnlock()` | `v, ok := b.processes.Load(key)` |
| Write | `b.mu.Lock(); b.processes[key] = val; b.mu.Unlock()` | `b.processes.Store(key, val)` |
| Delete | `b.mu.Lock(); delete(b.processes, key); b.mu.Unlock()` | `b.processes.Delete(key)` |
| Range | `b.mu.RLock(); for k, v := range b.processes { ... }; b.mu.RUnlock()` | `b.processes.Range(func(k, v interface{}) bool { ... })` |
| LoadOrStore | `b.mu.Lock(); if v, ok := b.processes[key]; ok { ... } else { b.processes[key] = val }; b.mu.Unlock()` | `v, ok := b.processes.LoadOrStore(key, val)` |

### 3. Atomic Operations for Counters

For `correlationSeq`, use `sync/atomic`:

```go
import "sync/atomic"

// Increment
nextID := atomic.AddUint64(&b.correlationSeq, 1)

// Load
seq := atomic.LoadUint64(&b.correlationSeq)
```

## Migration Steps

### Phase 1: Process Map (Highest Impact)
1. Update `Broker` struct: `processes map[string]*Process` → `processes sync.Map`
2. Update all process map accesses:
   - `broker.go`: `ProcessCount()`, `InsertProcessForTest()`
   - `spawn.go`: All `b.processes` accesses in `spawnInternal()`
   - `process_io.go`: `SendToProcess()`, `SendToAllProcesses()`
   - `process_lifecycle.go`: All `b.processes` accesses
3. Remove `b.mu` where it was only protecting `processes`
4. Add tests for concurrent access patterns

### Phase 2: Pending Requests Map (Medium Impact)
1. Update `Broker` struct: `pendingRequests map[string]*PendingRequest` → `pendingRequests sync.Map`
2. Update `PendingRequestCount()` to use `sync.Map`
3. Update `routing.go`: Load-and-delete pattern using `LoadAndDelete()` (Go 1.15+)
4. Update `rpc.go`: Store pending requests using `LoadOrStore()`
5. Remove `b.pendingMu`
6. Add tests for concurrent request/response correlation

### Phase 3: RPC Endpoints Map (Lower Impact)
1. Update `Broker` struct: `rpcEndpoints map[string][]string` → `rpcEndpoints sync.Map`
2. Update all endpoint map accesses
3. Remove `b.endpointsMu`
4. Add tests for concurrent endpoint registration

### Phase 4: Correlation Sequence Counter
1. Replace `correlationSeq` accesses with `sync/atomic`:
   - `atomic.AddUint64(&b.correlationSeq, 1)` for increment
   - `atomic.LoadUint64(&b.correlationSeq)` for read

## Benefits

1. **Reduced Lock Contention**: Eliminates mutex contention in high-concurrency scenarios
2. **Better Performance**: `sync.Map` is optimized for read-heavy workloads (typical for routing)
3. **Simpler Code**: No need to remember lock/unlock pairs
4. **Goroutine Safety**: Built-in goroutine safety without manual lock management

## Trade-offs

1. **Memory Overhead**: `sync.Map` uses more memory than `map` + mutex (amortized cost)
2. **Type Safety**: `sync.Map` uses `interface{}` values, requiring type assertions
3. **Range Performance**: `sync.Map.Range()` is slightly slower than `for range` on regular maps
4. **No Built-in Len**: Need to count manually or maintain separate counter

## Testing Strategy

1. **Unit Tests**: Add tests for all public methods under concurrent access
2. **Race Detector**: Run `go test -race` to verify no data races
3. **Benchmarks**: Compare performance before and after migration
4. **Stress Tests**: Run high-concurrency scenarios to validate lock-free behavior

## Estimated Effort

- Phase 1 (Process Map): 1 day
- Phase 2 (Pending Requests): 0.5 day
- Phase 3 (RPC Endpoints): 0.5 day
- Phase 4 (Correlation Sequence): 0.5 day
- Testing & Validation: 1 day

**Total**: 3.5 days

## References

- [Go sync.Map documentation](https://pkg.go.dev/sync#Map)
- [sync/atomic package](https://pkg.go.dev/sync/atomic)
- [Go blog: Intro to sync.Map](https://go.dev/blog/syncmap)

## Risks

1. **Type Assertion Errors**: Need to ensure all type assertions from `sync.Map` are safe
2. **Memory Leaks**: Improper use of `sync.Map` can lead to memory leaks if values aren't properly deleted
3. **Performance Regression**: In some workloads, `sync.Map` may be slower than mutex-protected maps

## Mitigation

1. **Comprehensive Testing**: Extensive test coverage for all map operations
2. **Performance Profiling**: Benchmark before/after to measure actual performance impact
3. **Gradual Rollout**: Test in staging environment before production deployment
4. **Fallback Plan**: Keep old mutex-protected code in a separate branch for quick rollback
