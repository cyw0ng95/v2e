# Performance Optimizer - Linux-Native Optimizations

## Overview

The performance optimizer provides high-throughput, low-latency message routing for the v2e broker using Linux-native kernel features.

## Key Features

### 1. CPU Affinity Binding
Workers are bound to specific CPU cores to:
- Reduce cache misses
- Minimize context switches
- Improve instruction pipeline efficiency

```go
setCPUAffinity(workerID)  // Binds worker to CPU core
```

### 2. Thread Pinning
Goroutines are locked to OS threads to:
- Prevent migration between threads
- Maintain CPU affinity across lifecycle
- Enable deterministic scheduling

```go
runtime.LockOSThread()  // Called at start of worker loop
```

### 3. Process Priority
Broker process runs at high priority (-10):
- Gets CPU time before lower-priority processes
- Reduces scheduling latency
- Prevents starvation by background tasks

```go
setProcessPriority()  // Called during initialization
```

### 4. I/O Priority
Workers use real-time I/O class:
- Disk operations don't block message routing
- Prioritizes message I/O over background writes
- Reduces tail latency for logging

```go
setIOPriority()  // Called in each worker
```

## Benchmarks

### Latency Consistency
Measures message routing latency variance under CPU load:
```bash
go test -bench=BenchmarkOptimizer_LatencyConsistency -benchmem
```

**Baseline:**
- Mean: 228.1 ns/op
- Stddev: 836,288,740 ns/op (high jitter)

### Large Payload Memory
Measures memory efficiency with 100MB+ payloads:
```bash
go test -bench=BenchmarkOptimizer_LargePayloadMemory -benchmem
```

**Baseline:**
- Latency: 704.4 ns/op
- Memory: 243 B/op, 3 allocs/op

## Testing

Verify Linux optimizations are working:
```bash
go test -v -run=TestCPUAffinitySetup
go test -v -run=TestProcessPrioritySetup
go test -v -run=TestIOPrioritySetup
go test -v -run=TestWorkerThreadPinning
```

## Requirements

- **Linux Kernel:** 2.6.23+
- **Capabilities:** CAP_SYS_NICE, CAP_SYS_ADMIN
- **Architecture:** Any Linux-supported (x86_64, ARM64, etc.)

## Production Deployment

Grant capabilities to the broker binary:
```bash
sudo setcap 'cap_sys_nice,cap_sys_admin=+ep' ./v2broker
```

Or run in Docker/Kubernetes with appropriate security context:
```yaml
securityContext:
  capabilities:
    add:
    - SYS_NICE
    - SYS_ADMIN
```

## Configuration

The optimizer uses sensible defaults:
- **Workers:** Number of CPU cores (min 4)
- **Buffer:** 1000 messages
- **Batch size:** 1 (immediate processing)
- **Flush interval:** 10ms

Custom configuration:
```go
opt := perf.NewWithConfig(router, perf.Config{
    BufferCap:     1000,
    NumWorkers:    runtime.NumCPU(),
    StatsInterval: 100 * time.Millisecond,
    OfferPolicy:   "drop",
    BatchSize:     1,
    FlushInterval: 10 * time.Millisecond,
})
```

## Monitoring

Check if optimizations are active:
```bash
# CPU affinity
taskset -cp $(pgrep v2broker)

# Process priority
ps -o pid,ni,cmd -p $(pgrep v2broker)

# I/O priority
ionice -p $(pgrep v2broker)
```

## See Also

- [../../../docs/LINUX_PERFORMANCE.md](../../../docs/LINUX_PERFORMANCE.md) - Comprehensive optimization guide
- [optimizer.go](optimizer.go) - Implementation
- [optimizer_test.go](optimizer_test.go) - Benchmarks
- [linux_optimizations_test.go](linux_optimizations_test.go) - Verification tests
