# Transport Layer Documentation

## Shared Memory Transport

### Overview

The shared memory transport provides a high-performance IPC mechanism for v2e broker communication, using Linux's `memfd_create` syscall for zero-copy operations.

### Architecture

The shared memory implementation consists of several key components:

1. **SharedMemory**: Core shared memory ring buffer using `memfd_create`
2. **BatchAck**: Batch acknowledgment mechanism for reduced syscall overhead
3. **HybridTransport**: Fallback-capable transport supporting both shared memory and UDS
4. **Synchronization Primitives**: Lock-free and low-latency synchronization

### Shared Memory Size Requirements

#### Minimum Size
- **Absolute minimum**: 4KB (SharedMemMinSize)
- **Typical small**: 64KB (SharedMemDefaultSize)
- **Maximum**: 16MB (SharedMemMaxSize)

#### Size Selection Guidelines

| Use Case | Recommended Size | Rationale |
|----------|----------------|-----------|
| Low-frequency RPC (< 100 msgs/s) | 64KB | Minimal overhead |
| Medium-frequency (100-1000 msgs/s) | 256KB | Buffer for burst traffic |
| High-frequency (1000-10000 msgs/s) | 1MB | Accommodates bursts |
| Very high-frequency (> 10000 msgs/s) | 4-8MB | Maximum throughput |

#### Memory Calculation

The actual memory allocated is page-aligned:

```go
actualSize = ((requestedSize + pageSize - 1) / pageSize) * pageSize
```

On typical x86-64 systems, pageSize = 4KB.

### Tuning Parameters

#### SharedMemory Configuration

```go
type SharedMemConfig struct {
    Size     uint32  // Size in bytes (4KB to 16MB)
    IsServer bool    // True for server (broker), false for client (subprocess)
}
```

**Size Tuning:**
- Too small: Frequent buffer overflows, fallback to UDS
- Too large: Memory waste, potential cache miss overhead
- Optimal: 2-4x average burst size

#### BatchAck Configuration

```go
type BatchAckConfig struct {
    MaxBatchSize  int           // Number of acks to batch (default: 32)
    FlushInterval time.Duration // Maximum time between flushes (default: 5ms)
    AckType       AckType       // Immediate, Batch, or Deferred
}
```

**MaxBatchSize Tuning:**
- Small (8-16): Low latency, higher syscalls
- Medium (32-64): Balanced (recommended)
- Large (128-256): Higher throughput, increased latency

**FlushInterval Tuning:**
- Aggressive (1-2ms): Low latency
- Balanced (5-10ms): Default
- Conservative (20-50ms): Max throughput

#### HybridTransport Configuration

```go
type HybridTransportConfig struct {
    SocketPath      string  // UDS socket path for fallback
    UseSharedMemory bool    // Enable shared memory transport
    SharedMemSize   uint32  // Size of shared memory buffer
    IsServer        bool    // True for broker, false for subprocess
}
```

**Recommendations:**
- Always enable shared memory when kernel >= 3.17
- Size based on expected message rate
- UDS fallback is automatic on failure

### Performance Characteristics

#### Expected Latency Improvements

| Operation | UDS Baseline | Shared Memory | Improvement |
|-----------|--------------|---------------|-------------|
| Small message (< 256B) | ~15μs | ~5μs | 66% |
| Medium message (256-4KB) | ~25μs | ~8μs | 68% |
| Large message (> 4KB) | ~50μs | ~15μs | 70% |

#### Throughput Improvements

| Message Size | UDS (msgs/s) | Shared Memory (msgs/s) | Improvement |
|--------------|--------------|----------------------|-------------|
| 256B | 50,000 | 150,000 | 200% |
| 1KB | 30,000 | 100,000 | 233% |
| 4KB | 10,000 | 50,000 | 400% |

### Synchronization Primitives

#### SpinLock
- **Use case**: Very short critical sections (< 100ns)
- **Overhead**: CPU cycles on contention
- **Avoid**: Long-held locks, high contention

#### SeqLock
- **Use case**: Read-heavy workloads with infrequent writes
- **Latency**: Read-side wait-free
- **Capacity**: Limited by atomic counter rollover

#### ShardedMutex
- **Use case**: High contention, sharding-friendly workloads
- **Configuration**: Number of shards (power of 2)
- **Overhead**: Per-shard mutex contention only

#### Semaphore
- **Use case**: Resource pooling, rate limiting
- **Capacity**: Maximum concurrent access count
- **Fairness**: FIFO ordering (approximate)

#### AtomicFlag
- **Use case**: Simple boolean flags, state transitions
- **Operations**: Set, Clear, Get, Swap, CAS
- **Latency**: Single atomic operation

### Kernel Requirements

#### Minimum Kernel Version
- **memfd_create**: Linux 3.17+
- **MFD_CLOEXEC flag**: Linux 3.17+

#### Verify Support
```bash
uname -r  # Check kernel version
```

#### Alternative on Older Kernels
The hybrid transport automatically falls back to UDS if `memfd_create` is unavailable.

### Monitoring and Metrics

#### Key Metrics
1. **SharedMemory.Active**: Currently active transport (uds/sharedmem)
2. **SharedMemory.BytesAvailable**: Available buffer space
3. **SharedMemory.FallbackCount**: Number of fallbacks to UDS
4. **BatchAck.PendingCount**: Number of pending acknowledgments
5. **BatchAck.FlushCount**: Number of batch flushes performed

#### Performance Monitoring
- Monitor transport switching frequency
- Track buffer utilization
- Measure actual vs. theoretical throughput
- Profile syscall count and latency

### Common Issues and Solutions

#### Issue: Frequent Fallback to UDS
**Symptoms**: High fallback count, degraded performance
**Solutions**:
1. Increase shared memory size
2. Reduce message frequency or size
3. Check for memory leaks
4. Verify kernel version support

#### Issue: High Latency Despite Shared Memory
**Symptoms**: Latency close to UDS baseline
**Solutions**:
1. Reduce batch ack size
2. Decrease flush interval
3. Check lock contention
4. Profile hot paths

#### Issue: Memory Leaks
**Symptoms**: Increasing memory usage over time
**Solutions**:
1. Ensure proper Close() calls
2. Verify shared memory cleanup
3. Check for goroutine leaks
4. Monitor file descriptor count

### Best Practices

1. **Size selection**: Start with 64KB, increase only if monitoring shows buffer pressure
2. **Batch ack**: Use batch mode with medium settings for balanced performance
3. **Error handling**: Monitor fallback events and investigate root cause
4. **Testing**: Run benchmarks in production-like conditions
5. **Monitoring**: Set up alerts for transport switches and buffer exhaustion

### Migration Path

#### From UDS to Shared Memory
1. Enable `UseSharedMemory` in config
2. Start with conservative size (64KB)
3. Monitor for fallbacks and performance improvements
4. Gradually increase size if needed
5. Keep UDS fallback enabled for safety

#### Rollback
To disable shared memory:
```go
config := HybridTransportConfig{
    UseSharedMemory: false,
    // ... other config
}
```

### Security Considerations

1. **File permissions**: UDS sockets use 0600 by default
2. **Shared memory**: `memfd_create` uses MFD_CLOEXEC to prevent inheritance
3. **No disk backing**: Memory-only, no persistent storage
4. **Process isolation**: Only processes with fd access can read/write

## Unix Domain Socket (UDS) Transport

### Overview

The UDS transport provides reliable IPC with automatic reconnection and error handling.

### Configuration

```go
type UDSTransport struct {
    socketPath           string
    reconnectAttempts    int
    reconnectDelay       time.Duration
    maxReconnectAttempts int
}
```

### Default Values
- **Max reconnect attempts**: 5
- **Reconnect delay**: 1 second
- **Socket permissions**: 0600 (owner read/write only)

### Reconnection Logic

The transport automatically reconnects on:
- Broken pipe errors
- Connection closed
- Connection reset
- EOF

### Error Handling

Errors can be handled via callbacks:
- `SetReconnectCallback()`: Called when max reconnection attempts exceeded
- `SetErrorHandler()`: Called for asynchronous errors

## API Reference

See individual Go files for complete API documentation:
- `shared_memory.go`: SharedMemory implementation
- `batch_ack.go`: Batch acknowledgment
- `sync_primitives.go`: Synchronization primitives
- `hybrid_transport.go`: Hybrid transport with fallback
- `uds_transport.go`: UDS transport implementation
