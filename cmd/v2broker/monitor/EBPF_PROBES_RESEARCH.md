# eBPF Probes Research for v2e Architecture

## Overview

This document researches and identifies key eBPF probes relevant to the v2e broker-first microservices architecture.

## Relevant System Calls and Events

### 1. IPC and Communication

#### Unix Domain Sockets (UDS)
- **syscalls**: `bind`, `listen`, `accept`, `connect`, `sendto`, `recvfrom`, `sendmsg`, `recvmsg`
- **tracepoints**: `sys_enter_unix_stream_connect`, `sys_exit_unix_stream_connect`
- **kprobes**: `unix_stream_connect`, `unix_stream_recvmsg`

#### Shared Memory
- **syscalls**: `mmap`, `munmap`, `memfd_create`, `ftruncate`
- **kprobes**: `shmem_file_setup`, `memfd_create`

### 2. Synchronization

#### Locks and Atomic Operations
- **kprobes**: `mutex_lock`, `mutex_unlock`, `spin_lock`, `spin_unlock`
- **tracepoints**: `lock_acquire`, `lock_release`

#### Semaphores
- **syscalls**: `futex`
- **tracepoints**: `sys_enter_futex`, `sys_exit_futex`

### 3. Memory Management

#### Allocation
- **syscalls**: `mmap`, `brk`
- **kprobes**: `__kmalloc`, `kfree`
- **tracepoints**: `kmalloc`, `kfree`

#### Page Faults
- **tracepoints**: `page_fault_user`, `page_fault_kernel`

### 4. CPU Scheduling

#### Context Switches
- **tracepoints**: `sched_switch`, `sched_wakeup`
- **kprobes**: `schedule`

#### CPU Usage
- **tracepoints**: `cpu_frequency`, `cpu_idle`

### 5. File I/O

#### File Operations
- **syscalls**: `open`, `close`, `read`, `write`, `fsync`
- **tracepoints**: `sys_enter_openat`, `sys_exit_openat`

#### Database Operations
- **syscalls**: `fdatasync`, `sync_file_range`
- **kprobes**: `submit_bio`

### 6. Network

#### TCP/IP Stack
- **tracepoints**: `tcp_probe`, `tcp_rcv_space_adjust`
- **kprobes**: `tcp_v4_connect`, `tcp_v4_rcv`

### 7. Go Runtime Events

#### Goroutine Scheduling
- **tracepoints**: `go:begin`, `go:end`, `go:block`, `go:unblock`
- **uprobe**: `runtime.goexit`, `runtime.gopark`

#### Garbage Collection
- **uprobe**: `runtime.gcStart`, `runtime.gcEnd`
- **tracepoints**: `go:gc`, `go:gc:begin`

#### Memory Allocation
- **uprobe**: `runtime.mallocgc`, `runtime.newobject`

## Priority Probes for v2e

### High Priority (Critical)

1. **UDS Communication Latency**
   - Probe: `sys_enter_sendmsg`, `sys_exit_sendmsg`
   - Metric: Round-trip time for broker-subprocess RPC
   - Target: < 10μs for shared memory, < 25μs for UDS

2. **Shared Memory Usage**
   - Probe: Custom user-space probe
   - Metric: Buffer utilization, overflow rate
   - Target: < 5% overflow rate

3. **Lock Contention**
   - Probe: `lock_acquire`, `lock_release`
   - Metric: Lock wait time, contention rate
   - Target: < 1% lock wait time

4. **Context Switches**
   - Probe: `sched_switch`
   - Metric: Context switches per second
   - Target: < 5000/second per core

### Medium Priority (Important)

5. **Memory Allocation Rate**
   - Probe: `kmalloc`, `kfree`, Go `runtime.mallocgc`
   - Metric: Allocation/deallocation rate
   - Target: Stable allocation pattern

6. **Database I/O**
   - Probe: `submit_bio`, `fdatasync`
   - Metric: I/O latency, throughput
   - Target: < 1ms for small writes

7. **Goroutine Scheduling**
   - Probe: `go:begin`, `go:end`, `go:block`
   - Metric: Goroutine count, scheduling latency
   - Target: < 1000 active goroutines

8. **GC Performance**
   - Probe: `go:gc`, `go:gc:begin`
   - Metric: GC pause time, frequency
   - Target: < 1ms pause time

### Low Priority (Nice to Have)

9. **CPU Frequency Scaling**
   - Probe: `cpu_frequency`
   - Metric: Frequency transitions
   - Target: Stable frequency

10. **Page Faults**
    - Probe: `page_fault_user`, `page_fault_kernel`
    - Metric: Page fault rate
    - Target: < 100 faults/second

## Implementation Considerations

### eBPF Map Types

1. **Hash Map**: For tracking per-process metrics
2. **Perf Event Array**: For sending events to user space
3. **Ring Buffer**: For high-throughput event streaming
4. **LPM Trie**: For efficient routing table lookups

### Sampling Rates

- High-frequency events (syscall): 1:1000 sample rate
- Medium-frequency events (scheduling): 1:100 sample rate
- Low-frequency events (GC): Full capture

### Overhead Target

- Target: < 1% CPU overhead
- Max: < 5% CPU overhead (alert threshold)
- Impact: Negligible impact on application latency

## Integration Points

### 1. Metrics Collection

- Export metrics in Prometheus format
- Include labels: process_id, event_type, subsystem
- Update frequency: 1 second

### 2. Alerting

- Threshold-based alerts for each probe type
- Integration with existing alerting system
- Alert severity: Info, Warning, Critical

### 3. Flame Graphs

- Stack trace collection for hot functions
- Profile duration: 10-60 seconds
- Output format: SVG for visualization

## Tools and Libraries

### eBPF Toolchains

1. **bpftrace**: High-level tracing language
   - Pros: Easy to write probes
   - Cons: Limited runtime overhead control

2. **bcc**: BPF Compiler Collection
   - Pros: Full feature set, stable
   - Cons: Verbose, complex setup

3. **libbpf**: Native eBPF library
   - Pros: Modern, efficient
   - Cons: Lower-level API

4. **cilium/ebpf (Go)**: Go eBPF library
   - Pros: Native Go integration
   - Cons: Smaller ecosystem

### Recommended Stack

- **Primary**: cilium/ebpf for Go integration
- **Development**: bpftrace for rapid prototyping
- **Production**: libbpf-based tools for stability

## Example Probes

### UDS Latency Probe (bpftrace)

```bpf
#!/usr/bin/bpftrace

BEGIN {
  printf("Tracing UDS latency... Hit Ctrl-C to end.\n");
}

tracepoint:syscalls:sys_enter_sendmsg /comm == "v2broker"/ {
  @start[tid] = nsecs;
}

tracepoint:syscalls:sys_exit_sendmsg /@start[tid]/ {
  $latency = nsecs - @start[tid];
  @latency[comm] = hist($latency);
  delete(@start[tid]);
}
```

### Lock Contention Probe (bpftrace)

```bpf
#!/usr/bin/bpftrace

kprobe:mutex_lock_slowpath {
  @contentions[comm] = count();
}
```

### Go Goroutine Count (uprobe)

```bpf
#!/usr/bin/bpftrace

uprobe:/proc/$1/exe:runtime.newobject {
  @allocs++;
}

profile:hz:99 {
  printf("Allocs: %d\n", @allocs);
}
```

## Kernel Version Requirements

### Minimum Versions

- **General eBPF support**: Linux 4.4+
- **Full probe support**: Linux 4.9+
- **Ring buffer**: Linux 5.8+
- **BTF (type information)**: Linux 5.2+

### Recommended Version

- **Minimum**: Linux 5.10 LTS
- **Recommended**: Linux 5.15 LTS
- **Optimal**: Linux 6.1 LTS

### Verify BTF Support

```bash
# Check if BTF is available
ls /sys/kernel/btf/vmlinux

# Check kernel version
uname -r

# Check eBPF support
cat /boot/config-$(uname -r) | grep BPF
```

## Permissions

### Required Capabilities

- `CAP_BPF`: Load and attach eBPF programs
- `CAP_PERFMON`: Read performance events
- `CAP_SYS_ADMIN`: (Alternative) Full admin privileges

### Container Considerations

- Privileged containers: Full eBPF access
- Non-privileged: Limited to user-space uprobes
- BPF LSM: Additional security layer

## Performance Impact Analysis

### Expected Overhead

| Probe Type | CPU Overhead | Memory Impact |
|------------|--------------|---------------|
| UDS tracing | < 0.1% | Negligible |
| Lock contention | < 0.5% | < 1MB per map |
| Memory allocation | < 0.2% | Negligible |
| Scheduling | < 0.3% | Negligible |

### Mitigation Strategies

1. **Sampling**: Reduce event capture rate
2. **Filtering**: Only capture relevant processes
3. **Burst control**: Limit event rate to user space
4. **Dynamic enabling**: Enable/disable probes at runtime

## Next Steps

1. **Prototype high-priority probes** using bpftrace
2. **Measure baseline performance** without eBPF
3. **Implement production-ready probes** using cilium/ebpf
4. **Integrate with metrics pipeline**
5. **Set up alerting thresholds**
6. **Validate overhead targets**
