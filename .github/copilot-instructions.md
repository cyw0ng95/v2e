# Copilot Instructions for v2e

## Architecture Principles

**CRITICAL: Broker-First Architecture**

The broker (`cmd/broker`) is a standalone process that boots and manages all subprocesses in the system.

**Strict Rules:**
- The broker is the ONLY process that can spawn and manage subprocesses
- The `access` service is a subprocess - it provides REST API but does NOT manage processes
- Subprocesses must NEVER embed broker logic or create their own broker instances
- All inter-process communication must go through broker-mediated RPC messages
- Never add process management capabilities to subprocesses

This architecture is fundamental to the system design. Violating it will cause circular dependencies and architectural problems.

## Documentation Guidelines

- Do **NOT** generate any documents other than `README.md`
- All project documentation should be consolidated in the `README.md` file
- Avoid creating additional markdown files, guides, or documentation files
- **NEVER** create implementation summaries, change logs, or status documents (e.g., IMPLEMENTATION_SUMMARY.md, CHANGELOG.md, STATUS.md)
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

### Principle 10: Database WAL Mode and Pragma Optimization
- **When to use**: SQLite databases with concurrent read/write access
- **Implementation**:
  ```go
  // Enable WAL mode for better concurrent access
  db.Exec("PRAGMA journal_mode=WAL")
  
  // Optimize synchronous mode (NORMAL is faster than FULL)
  db.Exec("PRAGMA synchronous=NORMAL")
  
  // Increase cache size for better query performance
  // -40000 means 40MB cache (negative = KB units)
  db.Exec("PRAGMA cache_size=-40000")
  
  // Set connection lifetime
  sqlDB.SetConnMaxLifetime(time.Hour)
  ```
- **Impact**: Better concurrent access, faster queries, reduced I/O
- **Evidence**: WAL allows readers and writers simultaneously; larger cache reduces disk I/O

### Principle 11: Worker Pool Pattern for Parallel Processing
- **When to use**: Processing multiple independent items (API calls, file operations)
- **Implementation**:
  ```go
  func FetchConcurrent(items []string, workers int) ([]Result, []error) {
      jobs := make(chan string, len(items))
      results := make(chan Result, len(items))
      errors := make(chan error, len(items))
      
      // Start worker pool
      for w := 0; w < workers; w++ {
          go func() {
              for item := range jobs {
                  result, err := processItem(item)
                  if err != nil {
                      errors <- err
                  } else {
                      results <- result
                  }
              }
          }()
      }
      
      // Send jobs
      for _, item := range items {
          jobs <- item
      }
      close(jobs)
      
      // Collect results
      // ... collect from results and errors channels
  }
  ```
- **Impact**: Parallel processing reduces total time for batch operations
- **Evidence**: N items processed in ~N/workers time units (minus overhead)

### Principle 12: Adaptive Buffer and Batch Sizing
- **When to use**: Message batching, I/O buffering, channel sizing
- **Implementation**:
  ```go
  // Before: Fixed small sizes
  outChan := make(chan []byte, 100)
  batch := make([][]byte, 0, 10)
  ticker := time.NewTicker(10 * time.Millisecond)
  
  // After: Optimized sizes based on workload
  outChan := make(chan []byte, 256)  // 2.5x larger for better throughput
  batch := make([][]byte, 0, 20)     // 2x larger batch
  ticker := time.NewTicker(5 * time.Millisecond)  // 2x faster flush for lower latency
  ```
- **Impact**: Better balance between throughput and latency
- **Evidence**: Larger batches reduce syscalls; smaller intervals reduce latency
- **Tuning**: Monitor metrics and adjust based on actual workload patterns

### Principle 13: Efficient Batch Writing with bufio.Writer
- **When to use**: Writing multiple messages or data blocks in sequence
- **Implementation**:
  ```go
  // Use pooled bufio.Writer for efficient batch writes
  writer := writerPool.Get().(*bufio.Writer)
  writer.Reset(output)
  defer writerPool.Put(writer)
  
  for _, data := range batch {
      writer.Write(data)
      writer.WriteByte('\n')
  }
  writer.Flush()
  ```
- **Impact**: Reduces syscalls and allocation overhead compared to fmt.Fprintf
- **Evidence**: SendMessage: 29% faster (294.4 → 209.1 ns/op), 51% less memory (162 → 80 B/op)

### Principle 14: Persistent bufio.Writer Pool
- **When to use**: When creating bufio.Writer objects repeatedly
- **Implementation**:
  ```go
  var writerPool = sync.Pool{
      New: func() interface{} {
          return bufio.NewWriterSize(nil, 4096)
      },
  }
  
  // Usage:
  writer := writerPool.Get().(*bufio.Writer)
  writer.Reset(output)
  defer writerPool.Put(writer)
  ```
- **Impact**: Eliminates repeated allocations of Writer objects
- **Evidence**: Reduced allocations from 4 to 2 per operation (50% reduction)

### Principle 15: Avoid String Formatting in Hot Paths
- **When to use**: Error messages and logging in frequently-called code
- **Implementation**:
  ```go
  // Before: fmt.Sprintf allocates
  Error: fmt.Sprintf("failed to parse message: %v", err)
  
  // After: Direct concatenation (no allocation for simple cases)
  Error: "failed to parse message: " + err.Error()
  ```
- **Impact**: Eliminates unnecessary allocations in error paths
- **Evidence**: ConcurrentSend: 37% faster (356.8 → 223.8 ns/op), 44% less memory (287 → 160 B/op)

### Performance Optimization Checklist

## RPC API Specification Guidelines

All RPC services (cmd/* except broker) MUST include a structured API specification at the top of their main.go file.

### RPC API Spec Format

Each cmd/*/main.go file should have a comment block at the top following this format:

```go
/*
Package main implements the [service-name] RPC service.

RPC API Specification:

[Service Name] Service
====================

Service Type: [RPC/REST]
Description: [Brief description of the service purpose]

Available RPC Methods:
---------------------

1. RPCMethodName
   Description: [What this method does]
   Request Parameters:
     - param_name (type, required/optional): Description
     - param_name2 (type, required/optional): Description
   Response:
     - field_name (type): Description
     - field_name2 (type): Description
   Errors:
     - Error condition 1: Description
     - Error condition 2: Description
   Example:
     Request:  {"param_name": "value"}
     Response: {"field_name": "result"}

2. RPCMethodName2
   [Similar structure as above]

Notes:
------
- [Any additional notes about the service]
- [Usage constraints or requirements]
- [Dependencies on other services]

*/
package main
```

### Development Process Requirements

**CRITICAL**: When developing new features or modifying existing RPC APIs, MUST follow this process:

1. **Design Phase**
   - Design the RPC method interface (name, parameters, response)
   - Consider error cases and validation requirements
   - Plan integration with other services

2. **Update Specification**
   - Update the RPC API Spec comment block in cmd/*/main.go
   - Document all parameters, response fields, and error cases
   - Add usage examples

3. **Implementation**
   - Implement the RPC handler following the spec
   - Add parameter validation
   - Implement error handling as specified

4. **Update Integration Tests**
   - Add integration test cases in tests/ directory
   - Test all documented scenarios (success, errors, edge cases)
   - Verify the spec matches actual behavior

5. **Verification**
   - Run unit tests: `./build.sh -t`
   - Run integration tests: `./build.sh -i`
   - Verify API spec is accurate and complete

### RPC API Specification Principles

- **Single Source of Truth**: The spec in main.go is the authoritative documentation
- **Always Up-to-Date**: Spec MUST be updated before or with code changes
- **Complete**: Document all parameters, responses, and errors
- **Testable**: Every spec item should have corresponding test cases
- **Examples**: Include realistic request/response examples

### Unit Test and Fuzz Test Separation

The unit test stage (via `./build.sh -t`) excludes fuzz tests by using the `-run='^Test'` flag, which only runs functions matching the Test* pattern. Fuzz tests (Fuzz*) are run separately in the fuzz test stage (via `./build.sh -f`).

This separation ensures:
- Fast feedback from unit tests
- Fuzz tests run with appropriate timeouts
- No unintentional fuzz test execution during normal unit testing

## Frontend Website Guidelines (v0.3.0+)

### Frontend Core Architecture

The website is located in the `website/` directory and follows these principles:

1. **Framework**: Next.js 15+ (App Router)
   - Output Strategy: Static Site Generation (SSG)
   - Requirement: `next.config.ts` MUST have `output: 'export'`
   
2. **Styling**: Tailwind CSS + shadcn/ui (Radix UI based)
   - Use shadcn/ui components for consistency
   - Follow the neutral color scheme
   
3. **Icons**: Lucide React
   - Use Lucide icons throughout the UI
   
4. **Data Fetching**: TanStack Query (React Query) v5
   - All data fetching MUST use React Query hooks
   - Hooks are defined in `lib/hooks.ts`

### RPC Adapter & Data Logic

The frontend uses a "Service-Consumer" pattern to bridge UI and backend:

1. **Client Factory**: `lib/rpc-client.ts`
   - Handles HTTP requests via `POST /restful/rpc`
   - Implements automatic case conversion
   - Supports mock mode for development
   
2. **Type Mirroring**: `lib/types.ts`
   - TypeScript interfaces MUST mirror Go structs
   - Use camelCase for TypeScript (conversion is automatic)
   - Keep types in sync with backend RPC specs
   
3. **Case Conversion**:
   - Outgoing: camelCase → snake_case (for Go backend)
   - Incoming: PascalCase/snake_case → camelCase (for TypeScript)
   - Conversion is handled in `convertKeysToCamelCase()` and `convertKeysToSnakeCase()`
   
4. **Mock Mode**: `NEXT_PUBLIC_USE_MOCK_DATA=true`
   - When true, RPC client returns simulated responses
   - Realistic delays simulate network latency
   - Allows frontend development without Go backend

### UI/UX Specifications

Use the following shadcn components for consistency:

1. **Layout**:
   - Container with responsive padding
   - Cards for grouping related content
   - Use spacing utilities consistently
   
2. **Forms**:
   - Use react-hook-form + zod validation
   - Integrate with shadcn/ui Form components
   - Display validation errors inline
   
3. **Data Display**:
   - Tables with client-side pagination
   - Badge components for status indicators
   - Skeleton loaders during loading states
   
4. **Feedback**:
   - Use Sonner (toasts) for success/error notifications
   - Toast messages should be concise and actionable
   - Use appropriate toast variants (success, error, info)

### Integration Requirements for Go

1. **Path Compatibility**:
   - All assets MUST use relative paths
   - No absolute paths in href, src, or import statements
   - Ensures assets load correctly when served from Go sub-route
   
2. **SPA Routing**:
   - NO `next/headers` or `next/cache` features
   - NO server-side features requiring Node.js runtime
   - Static pages only - use client-side routing
   
3. **Build Output**:
   - `npm run build` produces `out/` directory
   - Contains only HTML/JS/CSS/assets
   - Can be copied to `.build/package/` for Go access service
   
4. **Dynamic Routes**:
   - Dynamic routes require `generateStaticParams()`
   - Or avoid them and use client-side navigation only
   - Keep it simple for static export

### Development Workflow

When working on the frontend:

1. **Initial Setup**:
   ```bash
   cd website
   npm install
   cp .env.local.example .env.local
   # Set NEXT_PUBLIC_USE_MOCK_DATA=true for mock mode
   ```

2. **Development**:
   ```bash
   npm run dev  # Start dev server with hot reload
   ```

3. **Type Generation**:
   - When Go RPC APIs change, update `lib/types.ts`
   - Ensure camelCase naming convention
   - Keep types in sync with backend

4. **Testing**:
   ```bash
   npm run build  # Test static export
   npm run lint   # Check code style
   ```

5. **Integration**:
   - Build produces `out/` directory
   - Copy to Go service for static hosting
   - Test with Go backend running

### Frontend File Structure

```
website/
├── app/                    # Next.js app directory
│   ├── layout.tsx         # Root layout with providers
│   └── page.tsx           # Main dashboard page
├── components/            # React components
│   ├── ui/               # shadcn/ui components (auto-generated)
│   ├── cve-table.tsx     # Custom: CVE data table
│   └── session-control.tsx # Custom: Job session controls
├── lib/                   # Library code
│   ├── hooks.ts          # React Query hooks
│   ├── providers.tsx     # React providers (QueryClient, etc.)
│   ├── rpc-client.ts     # RPC client implementation
│   ├── types.ts          # TypeScript types from Go structs
│   └── utils.ts          # Utility functions (shadcn)
├── public/               # Static assets
├── .env.local.example    # Environment variables template
├── next.config.ts        # Next.js config (output: 'export')
├── package.json          # Dependencies
└── README.md             # Frontend documentation
```

### Adding New Features

When adding new frontend features:

1. **New RPC Method**:
   - Add TypeScript types to `lib/types.ts` (camelCase)
   - Add method to `RPCClient` in `lib/rpc-client.ts`
   - Add React Query hook to `lib/hooks.ts`
   - Add mock data if needed

2. **New UI Component**:
   - Use shadcn/ui components when possible
   - Place custom components in `components/`
   - Use Lucide icons for consistency
   - Follow existing patterns

3. **New Page**:
   - Add page in `app/` directory
   - Use client components (`'use client'`)
   - Avoid dynamic routes if possible
   - Use React Query hooks for data

### Common Patterns

**Data Fetching:**
```typescript
// In a component
const { data, isLoading, error } = useCVEList(offset, limit);
```

**Mutations:**
```typescript
const createCVE = useCreateCVE();

const handleCreate = () => {
  createCVE.mutate(cveId, {
    onSuccess: () => toast.success("Created!"),
    onError: (err) => toast.error(err.message),
  });
};
```

**Mock Mode:**
```typescript
// In .env.local
NEXT_PUBLIC_USE_MOCK_DATA=true  // Use mock data
NEXT_PUBLIC_USE_MOCK_DATA=false // Use real backend
```

### Troubleshooting

**Build fails with dynamic route error:**
- Remove dynamic routes or add `generateStaticParams()`
- Use client-side navigation instead

**Types mismatch:**
- Check case conversion (Go uses snake_case, TS uses camelCase)
- Verify RPC client conversion functions

**Mock mode not working:**
- Check `.env.local` has `NEXT_PUBLIC_USE_MOCK_DATA=true`
- Restart dev server after changing env vars
- Verify mock data in `lib/rpc-client.ts`
