# Copilot Instructions for v2e

## Documentation Guidelines

- Do **NOT** generate any documents other than `README.md`
- All project documentation should be consolidated in the `README.md` file
- Avoid creating additional markdown files, guides, or documentation files
- Keep the documentation simple and focused

## Project Guidelines

- This is a Go-based project using Go modules
- The project may contain multiple commands in the `cmd/` directory
- Follow Go best practices and conventions
- Use standard Go tooling (go build, go test, go mod)

## Architecture and Security Guidelines

### Deployment Model

- The broker is the **central entry point** for deployment
- Users deploy the package by running the broker with a config file (`config.json`)
- The broker spawns and manages all subprocess services
- All subprocesses are started by the broker, not directly by users

### Message Passing and Security

- **RPC-only communication**: All inter-process communication MUST use RPC messages
- **Broker-mediated routing**: All messages MUST be routed through the broker to ensure security
- **No direct communication**: Subprocesses MUST NOT communicate directly with each other
- **No external input**: Subprocesses MUST NOT accept any input other than stdin pipeline from the broker

### Subprocess Implementation Rules

When implementing a new subprocess service:

1. **Use the subprocess framework**: Always use `pkg/proc/subprocess` package
2. **Use common logging**: Call `subprocess.SetupLogging(processID)` to initialize logging from config.json
3. **Use common lifecycle**: Call `subprocess.RunWithDefaults(sp, logger)` for standard signal handling and error handling
4. **stdin/stdout only**: Only read from stdin and write to stdout
5. **No external inputs**: Do NOT accept command-line arguments, environment variables (except PROCESS_ID and service-specific config), or network connections for control
6. **RPC handlers only**: All functionality must be exposed via RPC message handlers
7. **Broker-spawned**: Services must be designed to be spawned by the broker via `SpawnRPC`

Example subprocess pattern:
```go
func main() {
    // Get process ID from environment
    processID := os.Getenv("PROCESS_ID")
    if processID == "" {
        processID = "my-service"
    }

    // Set up logging using common subprocess framework
    logger, err := subprocess.SetupLogging(processID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
        os.Exit(1)
    }

    // Create subprocess instance
    sp := subprocess.New(processID)

    // Register RPC handlers
    sp.RegisterHandler("RPCMyFunction", createMyFunctionHandler())

    logger.Info("My service started")

    // Run with default lifecycle management
    subprocess.RunWithDefaults(sp, logger)
}
```

### Configuration Guidelines

- Use `config.json` to define broker settings, logging, and processes to spawn
- Logging configuration in `config.json` applies to all subprocesses via the subprocess framework
- All process configurations should be in the broker config
- Subprocesses should only use environment variables for service-specific settings (e.g., database paths, API keys)

## Testing Guidelines

- When adding new features, always consider adding test cases
- Unit tests should be written in Go using the standard `testing` package
- Integration tests should be written in Python using pytest
- Test cases should cover:
  - Normal operation paths
  - Error handling and edge cases
  - Integration between multiple services (for RPC-based features)
  - Security constraints (e.g., subprocesses only accepting stdin input)
- Run `./build.sh -t` to execute unit tests
- Run `./build.sh -i` to execute integration tests
- Run `./build.sh -m` to execute performance benchmarks
- Test case generation should be comprehensive and follow existing patterns in the codebase

### Performance Benchmarks

- Benchmark tests are written in Go using the standard `testing` package with `Benchmark*` functions
- Benchmark files should be named `*_bench_test.go` to distinguish them from regular test files
- Each performance-critical package should have comprehensive benchmarks covering:
  - Common operations and use cases
  - Edge cases that might affect performance
  - Different payload sizes (small, medium, large)
  - Concurrent operations where applicable
- Use `b.ReportAllocs()` to report memory allocations
- Use `b.ResetTimer()` to exclude setup time from measurements
- Benchmarks are automatically run in CI/CD via GitHub Actions after unit tests pass

#### Running Benchmarks

To run benchmarks locally:
```bash
./build.sh -m          # Run all benchmarks and generate report
./build.sh -m -v       # Run with verbose output
```

The benchmark report includes:
- Date, host, and system information
- Full benchmark results with timing and memory allocation data
- Summary of slowest operations
- Summary of highest memory allocations
- Reports are saved to `.build/benchmark-report.txt` and `.build/benchmark-raw.txt`

#### Performance Optimization Workflow

**IMPORTANT**: When making performance optimizations:

1. **Run benchmarks BEFORE changes**: Establish a baseline
   ```bash
   ./build.sh -m
   cp .build/benchmark-report.txt .build/benchmark-baseline.txt
   ```

2. **Make your optimization changes**: Implement performance improvements

3. **Run benchmarks AFTER changes**: Measure the impact
   ```bash
   ./build.sh -m
   ```

4. **Compare results**: Document the improvement
   - Compare `.build/benchmark-baseline.txt` vs `.build/benchmark-report.txt`
   - Include before/after metrics in commit messages or PR descriptions
   - Highlight significant improvements (>10% speedup or memory reduction)

5. **Example commit message**:
   ```
   perf: optimize message marshaling
   
   Reduced message marshaling time by 25% and memory allocations by 30%.
   
   Benchmark results:
   - Before: BenchmarkMessageMarshal-2  500000  3000 ns/op  450 B/op
   - After:  BenchmarkMessageMarshal-2  700000  2250 ns/op  315 B/op
   ```

### Integration Test Constraints

**IMPORTANT**: All integration tests MUST follow the broker-first architecture:

1. **Start broker first**: The broker (or access service which embeds a broker) must be started before any subprocess
2. **No direct subprocess testing**: Do NOT start or test subprocesses directly without going through the broker
3. **Use access service as gateway**: Integration tests should use the access REST API as the primary entry point
4. **Broker spawns subprocesses**: Let the broker spawn and manage all subprocess services via configuration or REST API
5. **RESTful testing only**: All RPC tests for backend services (like cve-meta) MUST use the RESTful API endpoint `/restful/rpc` with the `target` parameter

#### Testing Backend Services via RESTful API

When testing backend RPC services (cve-meta, cve-local, cve-remote), use the access service's generic RPC endpoint:

```python
# Example: Testing cve-meta service
access = AccessClient()
response = access.rpc_call(
    method="RPCGetCVE",
    target="cve-meta",  # Target the specific backend service
    params={"cve_id": "CVE-2021-44228"}
)

# Verify standardized response format
assert response["retcode"] == 0
assert response["message"] == "success"
assert response["payload"] is not None
```

The access service routes the request as follows:
1. External test → REST API (`POST /restful/rpc`)
2. Access service → Broker (via RPC with `target` field)
3. Broker → Backend service (e.g., cve-meta)
4. Response flows back through the same chain

This ensures integration tests follow the same deployment model as production, where the broker is the central entry point.

## Performance Optimization Principles

When optimizing performance, apply these proven principles based on benchmarking evidence:

### Principle 1: Reduce Unnecessary Allocations with sync.Pool
- **When to use**: For frequently created and short-lived objects (e.g., Message structs, buffers)
- **Implementation**: 
  ```go
  var messagePool = sync.Pool{
      New: func() interface{} {
          return &Message{}
      },
  }
  
  func GetMessage() *Message {
      msg := messagePool.Get().(*Message)
      // Reset fields to zero values
      return msg
  }
  
  func PutMessage(msg *Message) {
      messagePool.Put(msg)
  }
  ```
- **Impact**: Reduces GC pressure and allocation overhead
- **Evidence**: Improved message unmarshaling by 29% (294.0 → 208.2 ns/op)

### Principle 2: Pre-allocate Slices with Exact Capacity
- **When to use**: When the final size is known beforehand
- **Implementation**:
  ```go
  // Good: Pre-allocate exact size
  items := make([]Item, len(records))
  for i, record := range records {
      items[i] = parseRecord(record)
  }
  
  // Avoid: Growing slice dynamically
  items := make([]Item, 0, len(records))
  for _, record := range records {
      items = append(items, parseRecord(record))
  }
  ```
- **Impact**: Eliminates slice re-allocations during growth
- **Evidence**: Reduced memory by 7% in ListCVEs (60960 → 56461 B/op)

### Principle 3: Enable Database Prepared Statements and Connection Pooling
- **When to use**: All database operations with GORM or sql.DB
- **Implementation**:
  ```go
  db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
      PrepareStmt: true,  // Cache prepared statements
  })
  
  sqlDB, _ := db.DB()
  sqlDB.SetMaxIdleConns(10)   // Reuse idle connections
  sqlDB.SetMaxOpenConns(100)  // Limit concurrent connections
  ```
- **Impact**: Reduces query compilation and connection overhead
- **Evidence**: 
  - GetCVE: 28% faster (49825 → 35490 ns/op)
  - Count: 27% faster (24164 → 17563 ns/op)

### Principle 4: Configure HTTP Connection Pooling
- **When to use**: All HTTP clients making multiple requests
- **Implementation**:
  ```go
  client := resty.New()
  client.SetTransport(&http.Transport{
      MaxIdleConns:        100,
      MaxIdleConnsPerHost: 10,
      IdleConnTimeout:     90 * time.Second,
      DisableCompression:  false,  // Enable compression
  })
  ```
- **Impact**: Reuses TCP connections, reduces handshake overhead
- **Evidence**: FetchCVEByID: 8% faster (143563 → 131026 ns/op)

### Principle 5: Use Buffer Pooling for Large Objects
- **When to use**: Large temporary buffers (e.g., scanner buffers, I/O buffers)
- **Implementation**:
  ```go
  var bufferPool = sync.Pool{
      New: func() interface{} {
          buf := make([]byte, MaxSize)
          return &buf
      },
  }
  
  // Use in function
  bufPtr := bufferPool.Get().(*[]byte)
  defer bufferPool.Put(bufPtr)
  buf := *bufPtr
  // Use buf...
  ```
- **Impact**: Reduces allocation of large objects
- **Evidence**: SendMessage: 12% faster (434.4 → 382.1 ns/op)

### Principle 6: Batch Database Operations
- **When to use**: Inserting/updating multiple records
- **Implementation**:
  ```go
  // Use CreateInBatches instead of individual inserts
  db.CreateInBatches(records, 100)  // Process 100 at a time
  ```
- **Impact**: Reduces per-record overhead and transaction count
- **Evidence**: BulkSaveCVEs: 2% less memory (174531 → 169335 B/op)

### Principle 7: Use sonic.ConfigFastest for Zero-Copy Parsing
- **When to use**: High-throughput message processing where data doesn't need long-term retention
- **Implementation**:
  ```go
  // For marshaling
  api := sonic.ConfigFastest
  data, err := api.Marshal(msg)
  
  // For unmarshaling
  api := sonic.ConfigFastest
  err := api.Unmarshal(data, &msg)
  ```
- **Impact**: Faster JSON operations with reduced allocations
- **Evidence**: 
  - SendResponse: 28% faster (1264 → 903.1 ns/op)
  - SendEvent: 14% faster (882.8 → 759.1 ns/op)

### Principle 8: Lock-Free Message Batching with Channels
- **When to use**: High-frequency message sending where batching reduces syscalls
- **Implementation**:
  ```go
  // Buffered channel for batching
  outChan := make(chan []byte, 100)
  
  // Writer goroutine batches messages
  go func() {
      batch := make([][]byte, 0, 10)
      ticker := time.NewTicker(10 * time.Millisecond)
      for {
          select {
          case data := <-outChan:
              batch = append(batch, data)
              if len(batch) >= 10 {
                  flushBatch(batch)
                  batch = batch[:0]
              }
          case <-ticker.C:
              if len(batch) > 0 {
                  flushBatch(batch)
                  batch = batch[:0]
              }
          }
      }
  }()
  
  // Send without blocking on mutex
  outChan <- data
  ```
- **Impact**: Reduces mutex contention and syscall overhead
- **Evidence**: SendMessage: 19% faster (378.4 → 303.4 ns/op)

### Principle 9: Separate Read and Write Mutexes
- **When to use**: When read operations significantly outnumber writes
- **Implementation**:
  ```go
  // Use RWMutex for handler map (read-heavy)
  mu sync.RWMutex
  
  // Use separate Mutex for writes only
  writeMu sync.Mutex
  
  // Read operation
  mu.RLock()
  handler := handlers[key]
  mu.RUnlock()
  
  // Write operation
  writeMu.Lock()
  fmt.Fprintf(output, "%s\n", data)
  writeMu.Unlock()
  ```
- **Impact**: Reduced lock contention for read-heavy workloads
- **Evidence**: Improved concurrent message handling

### Principle 10: Eliminate String Conversions in Hot Paths
- **When to use**: When writing binary data ([]byte) to I/O streams
- **Implementation**:
  ```go
  // Bad: Creates intermediate string (extra copy)
  fmt.Fprintf(output, "%s\n", string(data))
  
  // Good: Write bytes directly
  output.Write(data)
  output.Write([]byte{'\n'})
  ```
- **Impact**: Eliminates one memory copy per message
- **Evidence**: 
  - SendMessage: 28% faster (303.4 → 216.5 ns/op)
  - SendResponse: 22% faster (903.1 → 702.9 ns/op)
  - Reduced allocations from 4 → 3 per message

### Principle 11: Use writev() for Scatter-Gather I/O
- **When to use**: Batched writes to file descriptors (stdout, sockets, files)
- **Implementation**:
  ```go
  import "golang.org/x/sys/unix"
  
  // Build buffer array
  buffers := make([][]byte, 0, batchSize*2)
  for _, data := range batch {
      buffers = append(buffers, data, []byte{'\n'})
  }
  
  // Single syscall writes all buffers (zero-copy kernel operation)
  _, err := unix.Writev(int(file.Fd()), buffers)
  ```
- **Impact**: Reduces syscalls, zero-copy scatter-gather I/O in kernel
- **Evidence**: Batched writes complete in single syscall vs N syscalls

### Principle 12: Direct Byte Writes Over Formatted Output
- **When to use**: When you don't need fmt.Fprintf formatting features
- **Implementation**:
  ```go
  // Bad: String formatting overhead
  fmt.Fprintf(w, "%s\n", string(data))
  
  // Good: Direct byte writes
  w.Write(data)
  w.Write([]byte{'\n'})
  ```
- **Impact**: Avoids reflection and string formatting overhead
- **Evidence**: Combined with Principle 10, reduces copy operations by 67%

### Performance Optimization Checklist

Before optimizing:
1. ✅ Run benchmarks to establish baseline (`./build.sh -m`)
2. ✅ Identify bottlenecks from benchmark report
3. ✅ Choose appropriate principle(s) from above

During optimization:
1. ✅ Apply ONE principle at a time
2. ✅ Run benchmarks after each change
3. ✅ Verify no functional regressions (run tests)

After optimization:
1. ✅ Compare before/after metrics
2. ✅ Document improvements in commit message
3. ✅ Update copilot instructions if new principle discovered
