# Copilot Instructions for v2e

Purpose: give an AI coding agent the exact, actionable constraints and shortcuts
to be productive in this repo. Keep edits small, preserve broker-first rules.

- **Big picture (core rule):** This is a broker-first system. `cmd/broker` is the
   only process that spawns and manages subprocesses. Do not add process management
   into services under `cmd/*` (they are subprocesses).

 - **Where to look first:** [cmd/broker](../cmd/broker), [cmd/access](../cmd/access),
    [pkg/proc/subprocess](../pkg/proc/subprocess), `config.json`, and the per-service
   `service.md` files under `cmd/*`.

- **SSG (SCAP Security Guide) development:** When working on SSG-related code (`pkg/ssg/*`, `cmd/*/ssg_*`):
   - **ALWAYS** initialize the git submodule first: `git submodule update --init --recursive`
   - The submodule at `assets/ssg-static` contains reference data:
     - `guides/` - HTML guide files for testing parsers
     - `tables/` - HTML table files for validation
     - `manifests/` - JSON manifest files
     - `ssg-*-ds.xml` - SCAP data stream XML files
   - Use real files from the submodule for testing and validation
   - Do NOT create mock/fake SSG data when real data is available

- **Subprocess rules (must follow):** use `pkg/proc/subprocess`, call
   `subprocess.SetupLogging(PROCESS_ID)` and `subprocess.RunWithDefaults(sp, logger)`,
   expose functionality via RPC handlers (e.g., `sp.RegisterHandler("RPC...", ...)`),
   and only read stdin / write stdout for broker-controlled IO.

- **Communication model:** RPC-only between broker and services; external tests
   and tools call `cmd/access` REST endpoint `/restful/rpc` with `target=<service>`.

- **Build / test commands:** use the repository `build.sh` wrapper:
   - `./build.sh -t` → run Go unit tests (default: Level 1)
   - `V2E_TEST_LEVEL=2 ./build.sh -t` → run Level 1 and 2 tests (includes database tests)
   - `V2E_TEST_LEVEL=3 ./build.sh -t` → run all test levels
   - `./build.sh -i` → run integration tests (pytest; broker must be started)
   - `./build.sh -m` → run benchmarks
   Use `runenv.sh` or the broker config (`config.json`) to start local runs.

- **Hierarchical testing system:**
   - Always use `testutils.Run()` for Level 1 tests (pure logic, no DB)
   - Always use `testutils.RunWithDB()` for Level 2+ tests (database operations)
   - RunWithDB automatically creates transactions and rolls them back (no persistent side effects)
   - All tests run in parallel automatically via the wrapper
   - Example:
     ```go
     testutils.Run(t, testutils.Level1, "TestName", func(t *testing.T) {
         // Test implementation
     })
     testutils.RunWithDB(t, testutils.Level2, "DBTest", db, func(t *testing.T, tx *gorm.DB) {
         // Use tx for all database operations
     })
     ```

- **Conventions & expectations:**
   - Keep changes minimal and focused; follow Go module style.
   - Update `cmd/*/service.md` (RPC spec) when adding RPC handlers.
   - Integration tests must start the broker/access gateway — do not spawn subprocesses directly.

- **Quick examples:**
   - New subprocess skeleton: create `cmd/myservice`, use `pkg/proc/subprocess.New`,
      register RPC handlers, and call `RunWithDefaults`.
   - Integration testing: call `/restful/rpc?target=meta` on `cmd/access`.

- **Where not to add files:** avoid creating extra documentation files — keep
   public documentation in `README.md` unless asked otherwise.

If anything here is unclear or you'd like additional examples (small service
template, integration test snippet, or a local run recipe), tell me which one
and I'll add it.
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

## Updated Documentation Locations

The service descriptions have been moved to standalone `service.md` files for each service. Refer to the following files for detailed descriptions:

- Access Service: `cmd/access/service.md`
- Broker Service: `cmd/broker/service.md`
- Local Service: `cmd/local/service.md`
- Meta Service: `cmd/meta/service.md`
- Remote Service: `cmd/remote/service.md`

### Additional Performance Principles (added 2026-01-22)

Principle: Pool short-lived RPC parameter objects
- When building small parameter objects repeatedly in hot loops (e.g., RPC invocations), use `sync.Pool` to reuse small structs instead of allocating maps or new structs each iteration.
- Implementation: use a typed struct for RPC params and a `sync.Pool{New: func() interface{} { return &T{} }}`; get from pool, set fields, call RPC, then reset fields and put back.
- Impact: reduces allocations/op and GC pressure for tight loops that prepare RPC calls.

Principle: Prefer typed structs over maps for RPC parameters
- Maps allocate more and require runtime type information; use small typed structs with JSON tags when the RPC/serialization layer supports them.
- Implementation: replace `map[string]interface{}` with `struct { Field int `json:"field"` }` for commonly used RPC calls.
- Impact: reduces allocations, improves marshal/unmarshal performance, and is more type-safe.

Principle: Clear pooled objects before returning to pool
- Always reset or zero fields of pooled objects before `Put` to avoid leaking large data and to keep memory profiles predictable.
- Implementation: set strings to "", numeric fields to 0, and nested structs to their zero value (e.g., `v = T{}`).
- Impact: prevents inadvertent retention of large buffers and helps the allocator reuse smaller backing arrays.
