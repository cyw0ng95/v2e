# Performance Optimization Requirements

## Overview

This document describes the performance monitoring and optimization architecture for the v2e broker system.

**Important**: This is a small-scale system with only 5 subprocesses. Complex adaptive optimization models are overkill.

### System Scale

| Component | Value |
|-----------|-------|
| Subprocesses | 5 (access, local, meta, remote, sysmon) |
| Default buffer | 1000 |
| Default workers | 4 |
| Default batch size | 1 |

---

## Architecture

### Current Issues

| Issue | Location | Description |
|-------|----------|-------------|
| **Responsibility overlap** | sysmon vs sched/monitor | Both collect system metrics |
| **Subprocess data missing** | sched/monitor.go:31-36 | ProcessResourceMetrics is empty |
| **Over-complex optimization** | AdaptiveOptimizer | Moving averages, trends, percentiles are overkill |
| **Duplicate metrics** | perf/metrics.go + sched/monitor.go | Two places collect metrics |

### Target Architecture

**Philosophy**: 
- NO auto-optimization
- Data collection only (for frontend display)
- Manual tuning via vconfig

```
                    +-------------------------------------+
                    |             sysmon                  |
                    |  (collects /proc data)             |
                    +----------------+--------------------+
                                     | RPCSubmitProcessMetrics
                                     v
+---------------------------------------------------------------+
|                         broker                                  |
|  +-----------------------------------------------------------+|
|  |           ProcessMetricsStore                               ||
|  |  - Stores subprocess performance data                      ||
|  |  - Exposed via RPCGetMessageStats                         ||
|  +-----------------------------------------------------------+|
+---------------------------------------------------------------+
                           |
                           | (existing RPC)
                           v
+---------------------------------------------------------------+
|                      Frontend                                   |
|  - Displays metrics (existing page - NOT affected)            |
+---------------------------------------------------------------+

OPTIMIZATION: NONE - manual tuning via vconfig only
```

---

## Key Points

1. **Optimizer does NOTHING** - no auto-tuning, no decisions
2. **Data flows to frontend** - via existing RPC endpoints
3. **Existing frontend page works** - no changes needed to display

---

## Requirements

### 1. sysmon: Subprocess Performance Collection

**Purpose**: Collect data for frontend display

**Metrics from /proc/{pid}/status**:

| Metric | Description |
|--------|-------------|
| VmRSS | Physical memory (KB) |
| VmSize | Virtual memory (KB) |
| Threads | Thread count |

**Collection**: Every 30 seconds, submit to broker

| File | Change |
|------|--------|
| `cmd/v2sysmon/process_metrics.go` | NEW - Read /proc |
| `cmd/v2sysmon/main.go` | Add periodic collection |

### 2. broker: Store Metrics

**Purpose**: Store data for frontend display via existing RPC

**RPC**: `RPCSubmitProcessMetrics` (submit data)

**Storage**:
```go
type ProcessMetricsStore struct {
    mu     sync.RWMutex
    latest map[string]ProcessMetrics  // key: process ID
}
```

**Expose via**: Existing `RPCGetMessageStats` - add subprocess metrics to response

| File | Change |
|------|--------|
| `cmd/v2broker/core/process_metrics.go` | NEW - Store only |
| `cmd/v2broker/core/broker.go` | Register RPC |

### 3. Keep ONE Minimal Optimizer (Message Routing Only)

**Keep**: `perf.Optimizer` (in `cmd/v2broker/perf/optimizer.go`)

This optimizer ONLY does:
- Message routing to workers
- Basic counters (sent, received, errors)
- Fixed configuration from build-time

**Does NOT do**:
- No auto-adjustment
- No data analysis
- No optimization decisions

### 4. Remove Duplicate Metrics

**DELETE from sched/monitor.go**:
- CPU utilization (already in sysmon)
- Memory utilization (already in sysmon)
- System load average (already in sysmon)

**KEEP in sched/monitor.go**:
- Message queue depth (broker-internal)
- Latency metrics (broker-internal)

---

## Data Flow

```
/proc/{pid}/status (sysmon reads)
        |
        v (every 30s)
broker ProcessMetricsStore
        |
        v (via existing RPCGetMessageStats)
Frontend display page (EXISTING - NOT affected)
        |
        v
Operators view metrics manually
```

**Note**: Existing frontend page continues to work. New subprocess metrics are added to existing response.

---

## Acceptance Criteria

1. **sysmon collects subprocess data**: VmRSS, VmSize, Threads from /proc
2. **broker stores metrics**: ProcessMetricsStore accepts and stores data
3. **Data exposed to frontend**: Via existing RPCGetMessageStats
4. **Frontend page NOT affected**: Existing display continues to work
5. **No auto-optimization**: Optimizer does nothing actively
6. **Existing tests pass**: No regression

---

## Implementation Priority

| Priority | Task | Description |
|----------|------|-------------|
| 1 | Add subprocess metrics | sysmon → broker storage |
| 2 | Remove complex optimization | Delete AdaptiveOptimizer complexity |
| 3 | Remove duplicate metrics | Clean sched/monitor |

---

## Files

| File | Action |
|------|--------|
| `cmd/v2sysmon/process_metrics.go` | Create |
| `cmd/v2sysmon/main.go` | Modify |
| `cmd/v2broker/core/process_metrics.go` | Create |
| `cmd/v2broker/core/broker.go` | Modify |
| `cmd/v2broker/sched/module.go` | Simplify (keep minimal OptimizerInterface) |
| `cmd/v2broker/sched/monitor.go` | Remove duplicates |
| `cmd/v2broker/perf/analysis_optimizer.go` | **DELETE** |
| `cmd/v2broker/perf/batch_predictor.go` | **DELETE** |
| `cmd/v2broker/scaling/load_predictor.go` | **DELETE** (if unused) |
| `cmd/v2broker/scaling/auto_scaler.go` | **DELETE** (if unused) |
| `docs/perf.reqs.md` | This document |

---

## Minimal Optimizer Specification

Only ONE optimizer: `perf.Optimizer`

### Interface (unchanged)

```go
type OptimizerInterface interface {
    Offer(msg *proc.Message) bool
    Stop()
    Metrics() map[string]interface{}
    SetLogger(l *common.Logger)
    GetKernelMetrics() *perf.KernelMetrics
}
```

### Behavior

| Action | Implementation |
|--------|----------------|
| Message routing | Round-robin to workers |
| Worker count | Fixed (from ldflags) |
| Batch size | Fixed (from ldflags) |
| Auto-tuning | **NONE** - manual via vconfig |
| Metrics | Basic counters only |

---

## Why Minimal Optimizer?

1. **Small scale**: 5 subprocesses, low message volume
2. **Static workload**: CVE/CWE data is relatively static
3. **Build-time config**: vconfig already provides tunable parameters
4. **Single optimizer**: Just `perf.Optimizer` - message routing only
5. **No auto-tuning**: Manual tuning via vconfig + rebuild

**What to DELETE**:
- `AdaptiveOptimizer` (trends, moving averages)
- `AnalysisOptimizer` (service priorities)
- `BatchSizePredictor` (pattern recognition)
- `LoadPredictor` (unused)
- Duplicate metrics collection
