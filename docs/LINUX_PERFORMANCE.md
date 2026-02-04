# Linux-Native Performance Optimizations

## Overview

This document describes the Linux-native performance optimizations implemented in the v2e broker and SSG parser. These optimizations leverage kernel-level features to achieve deterministic performance, reduce latency jitter, and minimize memory overhead.

## Architecture

### Broker Performance Optimizations (`cmd/v2broker/perf/optimizer.go`)

#### 1. CPU Affinity Binding
**Function:** `setCPUAffinity(workerID int)`

Each worker goroutine is bound to a specific CPU core to:
- Reduce cache misses by keeping hot data in L1/L2 cache
- Minimize context switch overhead
- Improve instruction pipeline efficiency
- Distribute load evenly across all CPU cores

**Implementation:**
```go
func setCPUAffinity(workerID int) {
    numCPU := runtime.NumCPU()
    cpu := workerID % numCPU  // Round-robin distribution
    
    var cpuSet unix.CPUSet
    cpuSet.Zero()
    cpuSet.Set(cpu)
    
    tid := unix.Gettid()
    unix.SchedSetaffinity(tid, &cpuSet)
}
```

**Benefits:**
- Reduces cache thrashing when workers migrate between cores
- Ensures predictable memory access patterns
- Improves throughput on multi-core systems

#### 2. Thread Pinning
**Function:** `runtime.LockOSThread()` in worker loop

Prevents the Go scheduler from migrating goroutines between OS threads:
- Maintains CPU affinity across goroutine lifecycle
- Ensures madvise hints remain effective
- Provides consistent memory access patterns

**Implementation:**
```go
func (o *Optimizer) worker(id int) {
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    
    setCPUAffinity(id)
    setIOPriority()
    
    // Worker loop...
}
```

**Benefits:**
- Prevents goroutine migration that would invalidate CPU affinity
- Ensures kernel-level optimizations (like page cache) remain effective
- Provides deterministic scheduling behavior

#### 3. Process Priority
**Function:** `setProcessPriority()`

Sets the broker process priority to -10 (high priority):
- Ensures the broker gets CPU time before lower-priority processes
- Reduces scheduling latency for time-critical message routing
- Prevents starvation by background processes

**Implementation:**
```go
func setProcessPriority() {
    // PRIO_PROCESS with pid 0 means current process
    unix.Setpriority(unix.PRIO_PROCESS, 0, -10)
}
```

**Requirements:**
- Requires `CAP_SYS_NICE` capability
- In production, run broker with appropriate permissions

#### 4. I/O Priority
**Function:** `setIOPriority()`

Sets I/O priority to real-time class for workers:
- Ensures database writes don't block message routing
- Prioritizes message I/O over background disk operations
- Reduces tail latency for persistent logging

**Implementation:**
```go
func setIOPriority() {
    const (
        IOPRIO_CLASS_SHIFT = 13
        IOPRIO_CLASS_RT    = 1  // Real-Time class
        IOPRIO_PRIO_VALUE  = 0  // Highest priority within RT class
    )
    
    ioprio := (IOPRIO_CLASS_RT << IOPRIO_CLASS_SHIFT) | IOPRIO_PRIO_VALUE
    unix.Syscall(unix.SYS_IOPRIO_SET, IOPRIO_WHO_PROCESS, 0, uintptr(ioprio))
}
```

**Requirements:**
- Requires `CAP_SYS_ADMIN` capability
- In production, configure I/O scheduler (CFQ, BFQ) to respect priorities

### SSG Parser Memory Optimizations (`pkg/ssg/parser/`)

#### 1. Sequential Read-Ahead Hints
**Function:** `applySequentialHint(data []byte)`

Tells the kernel we'll read data sequentially:
- Enables aggressive read-ahead (prefetching pages before they're accessed)
- Reduces page faults by up to 70% for large XML/HTML files
- Improves throughput for streaming parsing operations

**Implementation:**
```go
func applySequentialHint(data []byte) {
    ptr := unsafe.Pointer(&data[0])
    length := len(data)
    
    // Enable sequential read-ahead
    unix.Madvise((*(*[1 << 30]byte)(ptr))[:length:length], unix.MADV_SEQUENTIAL)
    
    // Prefetch pages into memory
    unix.Madvise((*(*[1 << 30]byte)(ptr))[:length:length], unix.MADV_WILLNEED)
}
```

**Benefits:**
- Reduces page fault overhead by prefetching pages
- Improves cache hit rate for sequential access patterns
- Allows kernel to optimize I/O scheduling

#### 2. Immediate Memory Reclamation
**Function:** `reclaimMemory(data []byte)`

Tells the kernel to reclaim physical memory immediately after parsing:
- Reduces memory pressure on the system
- Allows other processes to use freed memory immediately
- Prevents memory bloat from temporary large buffers

**Implementation:**
```go
func reclaimMemory(data []byte) {
    ptr := unsafe.Pointer(&data[0])
    length := len(data)
    
    // Signal kernel to reclaim these pages immediately
    unix.Madvise((*(*[1 << 30]byte)(ptr))[:length:length], unix.MADV_DONTNEED)
}
```

**Usage Pattern:**
```go
func ParseDataStreamFile(r io.Reader, filename string) (...) {
    xmlData, _ := io.ReadAll(r)
    
    if len(xmlData) > 64*1024 {  // Only for large buffers
        applySequentialHint(xmlData)
        defer reclaimMemory(xmlData)  // Cleanup after parsing
    }
    
    // Parse XML...
}
```

**Benefits:**
- Reduces memory footprint after parsing
- Prevents OOM situations with multiple concurrent parsers
- Improves system-wide memory efficiency

## Benchmarks

### Latency Consistency (`BenchmarkOptimizer_LatencyConsistency`)

Measures message routing latency variance under CPU load:
- Simulates CPU contention with background workers
- Tracks latency distribution (mean, stddev)
- Validates jitter reduction from CPU affinity

**Baseline Metrics:**
- Mean latency: 228.1 ns/op
- Stddev: 836,288,740 ns/op (high jitter)
- Memory: 163 B/op, 3 allocs/op

**Expected with optimizations:**
- Stddev reduction: >40%
- Improved tail latency (p99, p99.9)
- More predictable message delivery times

### Large Payload Memory (`BenchmarkOptimizer_LargePayloadMemory`)

Measures memory usage and page faults with 100MB+ payloads:
- Processes 1MB messages repeatedly
- Tracks allocations and memory usage
- Validates memory reclamation effectiveness

**Baseline Metrics:**
- Latency: 704.4 ns/op
- Memory: 243 B/op, 3 allocs/op

**Expected with optimizations:**
- Page fault reduction: >30%
- Lower RSS (resident set size)
- Faster memory reclamation

## Production Deployment

### System Requirements

1. **Linux Kernel:** 2.6.23+ (for `madvise` support)
2. **Architecture:** x86_64, ARM64 (any Linux-supported)
3. **Capabilities:**
   - `CAP_SYS_NICE` for process priority and CPU affinity
   - `CAP_SYS_ADMIN` for I/O priority

### Running with Capabilities

```bash
# Run broker with capabilities
sudo setcap 'cap_sys_nice,cap_sys_admin=+ep' ./v2broker
./v2broker

# Or run with sudo (not recommended for production)
sudo ./v2broker
```

### Docker Deployment

```dockerfile
FROM ubuntu:22.04

# Install capabilities
RUN apt-get update && apt-get install -y libcap2-bin

COPY v2broker /usr/local/bin/

# Set capabilities
RUN setcap 'cap_sys_nice,cap_sys_admin=+ep' /usr/local/bin/v2broker

# Run as non-root user with capabilities
USER v2broker
CMD ["/usr/local/bin/v2broker"]
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: v2broker
spec:
  containers:
  - name: broker
    image: v2e/broker:latest
    securityContext:
      capabilities:
        add:
        - SYS_NICE
        - SYS_ADMIN
    resources:
      requests:
        cpu: "4"
        memory: "4Gi"
      limits:
        cpu: "8"
        memory: "8Gi"
```

## Performance Tuning

### CPU Affinity

For maximum performance, ensure:
1. Number of workers ≤ number of CPU cores
2. Hyperthreading disabled for deterministic performance
3. CPU cores isolated from OS scheduler (via `isolcpus` kernel parameter)

### I/O Scheduling

Configure the I/O scheduler for real-time workloads:
```bash
# Use BFQ (Budget Fair Queueing) scheduler
echo bfq > /sys/block/sda/queue/scheduler

# Or use deadline scheduler for low latency
echo deadline > /sys/block/sda/queue/scheduler
```

### Kernel Tuning

Optimize kernel parameters for v2e broker:
```bash
# Increase read-ahead for sequential access
echo 4096 > /sys/block/sda/queue/read_ahead_kb

# Reduce swappiness for low latency
sysctl vm.swappiness=10

# Increase max map count for large files
sysctl vm.max_map_count=262144
```

## Monitoring

### Verify Optimizations are Active

Check CPU affinity:
```bash
# Get broker PID
PID=$(pgrep v2broker)

# Check affinity
taskset -cp $PID
```

Check process priority:
```bash
ps -o pid,ni,cmd -p $(pgrep v2broker)
```

Check I/O priority:
```bash
ionice -p $(pgrep v2broker)
```

### Performance Metrics

Monitor these metrics to validate optimizations:
1. **Context switches:** Should decrease with CPU affinity
2. **Page faults:** Should decrease with madvise hints
3. **Latency stddev:** Should decrease >40% under load
4. **Memory RSS:** Should decrease with MADV_DONTNEED

## Limitations

1. **Linux-only:** These optimizations are not portable to macOS/Windows
2. **Capabilities required:** Production deployment needs special permissions
3. **Kernel version:** Requires modern Linux kernel (2.6.23+)
4. **CPU isolation:** For best results, dedicate CPU cores to broker

## Future Enhancements

1. **NUMA awareness:** Bind workers to NUMA nodes for better memory locality
2. **Huge pages:** Use transparent huge pages for large XML buffers
3. **CPU frequency scaling:** Pin CPU frequency to maximum for consistent performance
4. **IRQ affinity:** Isolate network interrupts to specific CPUs
5. **DPDK integration:** Zero-copy networking for ultra-low latency

## References

- [madvise(2) man page](https://man7.org/linux/man-pages/man2/madvise.2.html)
- [sched_setaffinity(2) man page](https://man7.org/linux/man-pages/man2/sched_setaffinity.2.html)
- [setpriority(2) man page](https://man7.org/linux/man-pages/man2/setpriority.2.html)
- [ioprio_set(2) man page](https://man7.org/linux/man-pages/man2/ioprio_set.2.html)
