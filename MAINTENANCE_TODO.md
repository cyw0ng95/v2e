# MAINTENANCE TODO

## TODO

 | ID  | Package           | Type    | Description                                                                        | Est LoC | Priority |
 |-----|------------------|---------|------------------------------------------------------------------------------------|----------|----------|
 | 020 | cmd/v2broker     | Code    | Implement segmented locks (sharded locks) for large maps                           | 600      | 1        |
 | 022 | cmd/v2broker     | Code    | Introduce Optimistic Concurrency Control (OCC) for read-heavy scenarios                | 600      | 1        |
| 038 | cmd/v2broker     | Code    | Research and select lock-free hash map implementation (Cuckoo Filter or SwissMap)      | 600      | 1        |
| 039 | cmd/v2broker     | Code    | Implement lock-free routing table with type-based fast indexing                          | 600      | 1        |
| 041 | cmd/v2broker     | Code    | Migrate existing map-based router to lock-free implementation                         | 600      | 1        |
| 045 | cmd/v2broker/perf| Code    | Design AdaptiveWorkerPool struct with load prediction capabilities                        | 600      | 1        |
| 046 | cmd/v2broker/perf| Code    | Implement load predictor using historical metrics and current system state                  | 600      | 1        |
| 047 | cmd/v2broker/perf| Code    | Add elastic scaling logic to add/remove workers dynamically                            | 600      | 1        |
| 049 | cmd/v2broker/perf| Code    | Add work-stealing algorithm for load balancing across workers                           | 600      | 1        |
| 016 | cmd/v2broker     | Code    | Monitor batch efficiency metrics (throughput vs latency trade-off)                   | 400      | 2        |
| 021 | cmd/v2broker     | Code    | Add per-method locks for high-frequency RPC handlers                                  | 400      | 2        |
| 025 | cmd/v2broker     | Code    | Add lock contention metrics to monitoring                                             | 400      | 2        |
| 040 | cmd/v2broker     | Code    | Add routing cache layer to avoid repeated lookups                                     | 400      | 2        |
| 050 | cmd/v2broker/perf| Code    | Integrate with existing Optimizer architecture                                         | 400      | 2        |
| 051 | cmd/v2broker/perf| Code    | Add metrics for worker utilization and scaling events                                    | 400      | 2        |
| 073 | pkg/cve/remote   | Code    | Implement request prioritization logic                                                  | 400      | 2        |
| 074 | pkg/cve/remote   | Code    | Add server push support for relevant endpoints                                          | 400      | 2        |
| 075 | pkg/cve/remote   | Code    | Create connection prewarming mechanism                                                  | 400      | 2        |
| 078 | pkg/cve/remote   | Code    | Add connection multiplexing metrics                                                    | 400      | 2        |
| 135 | pkg/proc         | Code    | Implement intelligent capacity prediction based on historical data                       | 600      | 2        |
| 137 | pkg/proc         | Code    | Add batch pre-allocation strategies                                                   | 400      | 2        |
| 106 | pkg/cve/local    | Code    | Design SmartConnectionPool with query pattern awareness                                | 600      | 2        |
| 107 | pkg/cve/local    | Code    | Implement QueryPatternAnalyzer for detecting frequently used queries                    | 600      | 2        |
| 108 | pkg/cve/local    | Code    | Add prepared statement caching based on query patterns                                | 400      | 2        |
| 110 | pkg/cve/local    | Code    | Add read/write separation connection pools                                               | 400      | 2        |
| 115 | pkg/cve/local    | Code    | Implement intelligent batch sizing based on data characteristics                        | 600      | 2        |
| 117 | pkg/cve/local    | Code    | Implement parallel batch operations using worker pools                                 | 600      | 2        |
| 017 | cmd/v2broker     | Config  | Create configuration hooks for batch size tuning                                     | 200      | 3        |
| 048 | cmd/v2broker/perf| Code    | Implement worker affinity to reduce cache misses                                       | 400      | 3        |
| 077 | pkg/cve/remote   | Code    | Maintain backward compatibility with HTTP/1.1                                        | 200      | 3        |
| 111 | pkg/cve/local    | Code    | Integrate with existing GORM configuration                                             | 200      | 3        |
| 112 | pkg/cve/local    | Code    | Add metrics for connection pool efficiency                                              | 200      | 3        |
| 116 | pkg/cve/local    | Code    | Add batch merging strategy to combine small batches                                     | 400      | 3        |
| 118 | pkg/cve/local    | Code    | Add batch size metrics and tuning recommendations                                      | 200      | 3        |
| 136 | pkg/proc         | Code    | Enable response buffer reuse for hot RPC methods                                      | 400      | 3        |
| 138 | pkg/proc         | Code    | Add prediction accuracy metrics                                                       | 200      | 3        |
| 018 | cmd/v2broker     | Test    | Add tests for various data volume scenarios                                           | 400      | 4        |
| 042 | cmd/v2broker     | Test    | Add comprehensive tests for correctness under concurrent access                          | 600      | 4        |
| 079 | pkg/cve/remote   | Test    | Test with external APIs that support HTTP/2                                           | 400      | 4        |
| 113 | pkg/cve/local    | Test    | Test with various query patterns                                                       | 400      | 4        |
| 119 | pkg/cve/local    | Test    | Test with various data volumes and types                                                | 400      | 4        |
| 139 | pkg/proc         | Test    | Test with various message patterns                                                    | 400      | 4        |
| 019 | cmd/v2broker     | Docs    | Document optimal batch size ranges for different message types                         | 200      | 4        |
| 044 | cmd/v2broker     | Docs    | Document migration strategy from current implementation                                   | 200      | 4        |
| 023 | cmd/v2broker     | Perf    | Profile lock contention in current codebase                                             | 400      | 4        |
| 024 | cmd/v2broker     | Perf    | Benchmark throughput improvements in high-concurrency (target: 40-50% improvement)     | 400      | 4        |
| 043 | cmd/v2broker     | Perf    | Benchmark message routing latency (target: 20-25% reduction)                      | 400      | 4        |
| 052 | cmd/v2broker/perf| Test    | Test under varying load patterns (spiky, steady, gradual)                           | 400      | 4        |
| 053 | cmd/v2broker/perf| Perf    | Benchmark CPU utilization improvement (target: 15-20%) and P99 latency reduction     | 400      | 4        |
| 080 | pkg/cve/remote   | Perf    | Benchmark concurrent request latency (target: 30% reduction) and connection count    | 400      | 4        |
| 127 | pkg/cve/local    | Perf    | Benchmark batch insert throughput (target: 2-3x improvement)                       | 400      | 4        |
| 134 | pkg/cve/remote   | Perf    | Benchmark external API call success rate (target: 20% improvement)                   | 400      | 4        |
| 140 | pkg/proc         | Perf    | Benchmark performance improvement (target: 15%)                                       | 400      | 4        |
| 026 | pkg/notes        | Coverage| Dramatically improve coverage from 14.0% to at least 60%                              | 800      | 5        |
| 027 | pkg/ssg          | Test    | Add comprehensive test suite for local package (currently 33.4%)                       | 600      | 5        |
| 028 | pkg/ssg          | Test    | Add comprehensive test suite for parser package (currently 42.9%)                      | 600      | 5        |
| 029 | pkg/ssg          | Coverage| Improve coverage for remote package (currently 26.0%)                                  | 400      | 5        |
| 030 | cmd/v2analysis   | Test    | Add test level definitions (currently 0 occurrences)                                     | 400      | 5        |
| 031 | cmd/v2analysis   | Coverage| Improve coverage from 6.8% to at least 50%                                           | 600      | 5        |
| 032 | cmd/v2access     | Coverage| Improve coverage from 9.4% to at least 50%                                           | 400      | 5        |
| 033 | cmd/v2local      | Test    | Fix test suite (no coverage data available despite 24 test level occurrences)           | 600      | 5        |
| 034 | cmd/v2remote     | Test    | Fix test suite (currently 0.0% coverage despite 15 test level occurrences)         | 600      | 5        |
| 035 | cmd/v2broker     | Test    | Fix test suite (currently 0.0% coverage despite 4 test level occurrences)          | 600      | 5        |
| 036 | cmd/v2sysmon     | Test    | Fix test suite (currently 0.0% coverage despite 3 test level occurrences)          | 600      | 5        |
| 037 | cmd/v2meta       | Test    | Fix test suite (no coverage data available despite 35 test level occurrences)         | 600      | 5        |
| 054 | pkg/cve          | Test    | Add test level definitions to all tests (currently only 10 occurrences)                | 400      | 5        |
| 055 | pkg/cwe          | Coverage| Improve coverage for job package (currently 70.8%)                                   | 400      | 5        |
| 056 | pkg/cwe          | Test    | Add test level definitions to all tests (currently 20 occurrences)                     | 400      | 5        |
| 057 | pkg/common        | Coverage| Improve coverage for procfs package (currently 0.0%)                                    | 400      | 5        |
| 058 | pkg/attack        | Coverage| Improve ImportFromXLSX coverage (currently 2.2%)                                    | 400      | 5        |
| 059 | pkg/attack        | Coverage| Improve ImportFromXLSX coverage (currently 2.2%)                                    | 400      | 5        |
| 060 | pkg/proc         | Coverage| Improve coverage for subprocess package (currently 52.9%)                                | 600      | 5        |
| 061 | pkg/asvs         | Test    | Add test for ImportFromCSV method (currently 0% coverage)                           | 400      | 5        |
| 062 | pkg/asvs         | Coverage| Increase overall test coverage from 38.2% to at least 60%                             | 600      | 5        |
| 063 | pkg/analysis      | Test    | Review and verify test levels are correct                                            | 200      | 5        |
| 064 | pkg/notes        | Test    | Review test levels - ensure Level 1 for basic logic, Level 2 for database           | 200      | 5        |
| 065 | pkg/ssg          | Test    | Add test level definitions for all tests                                             | 400      | 5        |
| 128 | pkg/cve/remote   | Code    | Design AdaptiveRetry struct with configurable strategies                                  | 400      | 5        |
| 129 | pkg/cve/remote   | Code    | Implement exponential backoff with jitter generation                                     | 400      | 5        |
| 130 | pkg/cve/remote   | Code    | Add circuit breaker pattern for fast failure                                            | 400      | 5        |
| 131 | pkg/cve/remote   | Code    | Implement priority-based retry for critical requests                                       | 400      | 5        |
| 132 | pkg/cve/remote   | Code    | Add retry metrics (success rate, backoff time, circuit state)                           | 400      | 5        |
| 133 | pkg/cve/remote   | Test    | Test retry logic under various failure scenarios                                         | 400      | 5        |
| 109 | pkg/cve/local    | Code    | Implement connection health checks with automatic reconnection                          | 400      | 5        |
| 144 | website           | Test    | Fix ESLint binary - lint command fails                                               | 15       | 5        |
| 145 | website           | Code    | Refactor large components (notes-framework.tsx 731 lines, page.tsx 468 lines)        | 150      | 5        |
| 146 | website           | Code    | Consolidate duplicate mock data generation in rpc-client.ts                                | 50       | 5        |
| 147 | website           | Code    | Add error boundaries to wrap async components appropriately                                 | 50       | 5        |
| 148 | website           | Types   | Replace any types with proper TypeScript interfaces                                      | 100      | 5        |
| 149 | website           | Perf    | Analyze bundle with @next/bundle-analyzer and implement code splitting                   | 100      | 5        |
| 150 | website           | Perf    | Review and optimize component memoization usage                                           | 75       | 5        |
| 151 | website           | Deps    | Update packages to latest stable versions                                              | 25       | 5        |
| 152 | website           | Security| Add runtime validation for environment variables using zod                               | 25       | 5        |
| 153 | website           | Test    | Set up Jest + React Testing Library and write unit tests                                | 600      | 5        |
| 154 | website           | Test    | Set up Playwright or Cypress for E2E testing                                        | 400      | 5        |
| 155 | website           | Docs    | Add JSDoc comments to all components                                                  | 150      | 5        |
| 156 | website           | Docs    | Add advanced usage examples to README                                                 | 75       | 5        |
| 157 | website           | A11y    | Audit with axe DevTools and fix ARIA label issues                                      | 75       | 5        |
| 158 | website           | A11y    | Ensure full keyboard navigation for custom components                                     | 50       | 5        |
| 159 | website           | DX      | Set up Husky for pre-commit hooks                                                   | 50       | 5        |
| 160 | website           | DX      | Add Prettier configuration and integrate with linting                                 | 25       | 5        |
| 161 | website           | Perf    | Benchmark build times and consider Turbopack or incremental builds                     | 100      | 5        |
| 162 | website           | Feature | Implement server-side search for CVE database                                            | 150      | 5        |
| 163 | website           | UX      | Improve loading states with sophisticated animations                                         | 75       | 5        |
| 164 | website           | UX      | Enhance error recovery with retry buttons, context, and suggested actions                 | 100      | 5        |
| 165 | website           | Debt    | Make RPC timeout configurable per request type                                          | 50       | 5        |
| 166 | website           | Debt    | Remove or conditionalize console.log statements in production                             | 25       | 5        |
 | 167 | website           | Debt    | Extract magic numbers to configuration constants                                          | 25       | 5        |
 | 168 | pkg/notes        | Code    | Refactor rpc_handlers.go (1527 lines) into smaller, focused modules                         | 800      | 2        |
 | 169 | pkg/notes        | Code    | Extract common RPC validation and error handling patterns into helper functions              | 400      | 3        |
 | 170 | pkg/notes/rpc_client | Code   | Refactor rpc_client.go (1101 lines) into smaller, focused modules                         | 600      | 3        |
 | 171 | pkg/notes/service | Code    | Refactor service.go (1006 lines) into smaller, focused modules                             | 500      | 3        |
 | 172 | pkg/ssg/local/store | Code   | Refactor store.go (828 lines) into smaller modules (models, queries, migrations)           | 400      | 3        |
 | 173 | pkg/cwe/local     | Code    | Refactor local.go (665 lines) into smaller, focused modules                                 | 400      | 3        |
 | 174 | pkg/cve/taskflow/executor | Code | Refactor executor.go (612 lines) - extract job state management                        | 350      | 3        |
 | 175 | pkg/ssg/job/importer | Code  | Refactor importer.go (597 lines) - extract parsing logic                                   | 350      | 3        |
 | 176 | pkg/cve/local/learning | Code | Refactor learning.go (518 lines) - extract pattern analysis                               | 300      | 3        |
 | 177 | pkg/ssg/parser/guide | Code  | Refactor guide.go (512 lines) - extract parsing helpers                                     | 300      | 3        |
 | 178 | pkg/cve/remote   | Code    | Replace panic() with error return in fetcher.go:50                                     | 20       | 1        |
 | 179 | pkg/proc/subprocess | Code   | Replace os.Exit(253) with graceful error handling in subprocess.go:273                   | 30       | 2        |
 | 180 | pkg/urn           | Code    | Replace panic(err) with error return in urn.go                                         | 40       | 2        |
 | 181 | pkg/proc/subprocess | Code   | Extract subprocess reconnection logic into dedicated module                                  | 200      | 3        |
 | 182 | pkg/proc/subprocess | Code   | Add lock contention metrics for handlers map access                                        | 150      | 3        |
 | 183 | pkg/proc/subprocess | Code   | Evaluate using atomic operations for hot path statistics counters                       | 100      | 3        |
 | 184 | pkg/proc/subprocess | Code   | Add goroutine leak detection and monitoring                                             | 200      | 3        |
 | 185 | cmd/v2broker/transport | Code | Add lock contention metrics for transport operations                                      | 150      | 3        |
 | 186 | pkg/cve/local     | Code    | Replace map[string]interface{} usage with typed structs (542 occurrences)              | 1200     | 2        |
 | 187 | pkg/notes        | Code    | Replace map[string]interface{} in RPC handlers with typed request/response structs       | 800      | 2        |
 | 188 | pkg/cve/local     | Code    | Implement batch query patterns for N+1 query prevention                                | 400      | 2        |
 | 189 | pkg/cve/local     | Code    | Add sync.Pool for frequently allocated database query structures                         | 200      | 3        |
 | 190 | pkg/proc/subprocess | Code   | Add sync.Pool for frequently allocated message structures                                 | 150      | 3        |
 | 191 | pkg/cve/local     | Code    | Add database connection pool health checks and metrics                                  | 200      | 2        |
 | 192 | pkg/cve/local     | Code    | Implement connection leak detection with automatic cleanup                                 | 250      | 3        |
 | 193 | pkg/cve/local     | Code    | Ensure all database connections are properly closed in defer statements                      | 100      | 2        |
 | 194 | pkg/cve/local     | Code    | Ensure all file handles are properly closed with defer                                     | 100      | 2        |
 | 195 | pkg/proc/subprocess | Code   | Ensure all network connections are properly closed in error paths                          | 100      | 2        |
 | 196 | pkg/notes        | Code    | Add comprehensive input validation for all RPC parameters                               | 400      | 2        |
 | 197 | pkg/cve/local     | Code    | Audit database queries for SQL injection vulnerabilities (GORM should prevent but verify)  | 200      | 2        |
 | 198 | pkg/cve/remote   | Code    | Add input sanitization for external API parameters                                       | 150      | 2        |
 | 199 | pkg/notes        | Code    | Add rate limiting for RPC endpoints to prevent abuse                                      | 300      | 3        |
 | 200 | pkg/cve/taskflow/executor | Test | Add comprehensive concurrent execution tests for job executor                                | 400      | 3        |
 | 201 | pkg/proc/subprocess | Test   | Add concurrent stress tests for message handling                                          | 300      | 3        |
 | 202 | cmd/v2broker/transport | Test | Add concurrent connection handling stress tests                                         | 300      | 3        |
 | 203 | pkg/proc/subprocess | Test   | Add benchmark tests for message batching performance                                       | 200      | 4        |
 | 204 | pkg/cve/local     | Test    | Add benchmark tests for database query patterns                                          | 200      | 4        |
 | 205 | pkg/cve/remote   | Test    | Add benchmark tests for HTTP client connection pooling                                   | 200      | 4        |
 | 206 | cmd/v2broker/perf | Test    | Add benchmark tests for optimizer worker pool                                           | 200      | 4        |
 | 207 | pkg/cwe/local     | Code    | Implement TODO at line 425: parse comma-separated string to []string for Phase field      | 50       | 4        |
 | 208 | pkg/cwe/local     | Code    | Implement TODO at line 574: parse comma-separated string to []string for Phase field      | 50       | 4        |
 | 209 | pkg/notes        | Code    | Add structured logging for RPC request/response tracing                                  | 300      | 3        |
 | 210 | pkg/proc/subprocess | Code   | Add structured logging for message lifecycle events                                       | 200      | 3        |
 | 211 | pkg/cve/taskflow/executor | Code | Add structured logging for job state transitions                                          | 200      | 3        |
 | 212 | pkg/cve/local     | Code    | Add context-based timeout for long-running database queries                                | 150      | 2        |
 | 213 | pkg/cve/remote   | Code    | Add context-based timeout for external API calls                                         | 100      | 2        |
 | 214 | pkg/proc/subprocess | Code   | Add context-based timeout for message handling                                          | 100      | 2        |
 | 215 | pkg/notes        | Docs    | Add inline comments for complex business logic in service.go                                | 150      | 4        |
 | 216 | pkg/notes        | Docs    | Document RPC handler patterns and error handling strategies                               | 200      | 4        |
 | 217 | pkg/proc/subprocess | Docs   | Document message batching and flush mechanisms                                          | 150      | 4        |
 | 218 | pkg/cve/local     | Docs    | Document query pattern analysis and caching strategies                                    | 200      | 4        |
 | 219 | pkg/cve/remote   | Docs    | Document external API retry and fallback strategies                                       | 200      | 4        |
 | 220 | pkg/cve/taskflow  | Test    | Add integration tests for job workflow with subprocesses                                   | 400      | 3        |
 | 221 | pkg/cve/remote   | Security| Replace math/rand with crypto/rand for secure random number generation in adaptive_retry.go | 50       | 1        |
 | 222 | pkg/cve/remote   | Security| Validate and sanitize all external API parameters before use                               | 150      | 1        |
 | 223 | pkg/notes        | Security| Add input validation and sanitization for user-provided data in RPC handlers              | 300      | 1        |
 | 224 | pkg/cve/local     | Security| Review and implement proper SQL injection prevention for all dynamic queries               | 200      | 1        |
 | 225 | pkg/proc/subprocess | Security| Validate message size limits to prevent memory exhaustion attacks                          | 100      | 1        |
 | 226 | cmd/v2broker     | Code    | Add goroutine leak detection and monitoring for broker processes                          | 300      | 2        |
 | 227 | pkg/cve/taskflow  | Code    | Add goroutine leak detection for job executor and worker pools                            | 250      | 2        |
 | 228 | pkg/meta/fsm     | Code    | Add goroutine leak detection for provider state machines                                  | 200      | 2        |
 | 229 | pkg/proc/subprocess | Code    | Implement graceful shutdown for all goroutines on context cancellation                      | 300      | 2        |
 | 230 | pkg/cve/local     | Code    | Replace time.Sleep with context-based timeout in retry logic                            | 100      | 2        |
 | 231 | pkg/meta/fsm     | Code    | Replace time.Sleep with context-based timeout in provider retry logic                    | 100      | 2        |
 | 232 | pkg/cve/local     | Code    | Add context-based timeouts to all database queries                                        | 200      | 2        |
 | 233 | pkg/cve/remote   | Code    | Ensure all HTTP response bodies are properly closed                                      | 100      | 2        |
 | 234 | pkg/asvs         | Code    | Ensure all HTTP response bodies are properly closed (resp.Body.Close())                   | 50       | 2        |
 | 235 | pkg/cve/local     | Code    | Add connection pool monitoring and leak detection                                         | 250      | 2        |
 | 236 | pkg/cwe/local     | Code    | Add connection pool monitoring and leak detection                                         | 250      | 2        |
 | 237 | pkg/attack/local  | Code    | Add connection pool monitoring and leak detection                                         | 250      | 2        |
 | 238 | pkg/capec/local   | Code    | Add connection pool monitoring and leak detection                                         | 250      | 2        |
 | 239 | pkg/asvs/local    | Code    | Add connection pool monitoring and leak detection                                         | 250      | 2        |
 | 240 | pkg/ssg/local/store | Code   | Add connection pool monitoring and leak detection                                         | 250      | 2        |
 | 241 | pkg/proc/subprocess | Code    | Add structured error handling for all ignored errors (_ = ...)                            | 300      | 2        |
 | 242 | pkg/notes        | Code    | Add structured error handling for all ignored errors (_ = ...)                            | 300      | 2        |
 | 243 | pkg/cve/local     | Code    | Add structured error handling for all ignored errors (_ = ...)                            | 300      | 2        |
 | 244 | pkg/proc/subprocess | Code    | Replace 130 instances of context.Background with context from parent                        | 400      | 2        |
 | 245 | pkg/notes        | Code    | Add proper error messages with context for all RPC handlers                              | 200      | 2        |
 | 246 | pkg/cve/local     | Code    | Add proper error messages with context for database operations                            | 200      | 2        |
 | 247 | pkg/cve/remote   | Code    | Add proper error messages with context for external API calls                            | 200      | 2        |
 | 248 | cmd/v2access     | Code    | Add comprehensive error logging for HTTP requests                                        | 150      | 2        |
 | 249 | cmd/v2local      | Code    | Add comprehensive error logging for database operations                                    | 150      | 2        |
 | 250 | pkg/proc/subprocess | Perf    | Optimize message batching to reduce memory allocations                                     | 200      | 2        |
 | 251 | pkg/proc/subprocess | Perf    | Pre-allocate slices in hot paths to reduce allocations                                    | 150      | 2        |
 | 252 | pkg/cve/local     | Perf    | Pre-allocate slices for batch operations to reduce allocations                             | 150      | 2        |
 | 253 | pkg/notes        | Perf    | Pre-allocate slices for list operations to reduce allocations                              | 150      | 2        |
 | 254 | pkg/cve/remote   | Perf    | Implement connection pooling for resty client beyond HTTP/2 transport                      | 200      | 2        |
 | 255 | pkg/cve/remote   | Perf    | Optimize JSON unmarshaling with faster library or zero-copy techniques                    | 200      | 2        |
 | 256 | pkg/proc/subprocess | Perf    | Use sync.Pool for temporary buffers in message processing                                 | 150      | 3        |
 | 257 | pkg/proc/subprocess | Perf    | Use atomic operations for message counters in hot paths                                   | 100      | 3        |
 | 258 | pkg/proc/subprocess | Perf    | Implement lock-free statistics collection for message metrics                             | 300      | 3        |
 | 259 | pkg/cve/local     | Perf    | Use sync.Pool for frequently allocated database query structures                           | 150      | 3        |
 | 260 | pkg/cve/local     | Perf    | Optimize batch insert with prepared statement caching                                    | 200      | 3        |
 | 261 | pkg/notes        | Perf    | Use sync.Pool for frequently allocated RPC response structures                            | 150      | 3        |
 | 262 | pkg/meta/fsm     | Perf    | Use sync.Pool for frequently allocated FSM state structures                              | 150      | 3        |
 | 263 | pkg/cve/remote   | Perf    | Implement request deduplication for repeated CVE fetches                                 | 200      | 3        |
 | 264 | pkg/cve/remote   | Perf    | Add response caching for frequently accessed CVEs                                       | 200      | 3        |
 | 265 | pkg/ssg/parser    | Code    | Implement TODO at guide.go: parse references from SSG guides                           | 150      | 4        |
 | 266 | pkg/notes/service | Code    | Implement TODO: Add CardType, Author, IsPrivate, Metadata fields to model                 | 100      | 4        |
 | 267 | pkg/notes/service | Code    | Implement TODO: Add author, is_private if added to model                                 | 50       | 4        |
 | 268 | pkg/ssg/local/store | Code    | Add context-based timeouts to all database operations                                     | 150      | 3        |
 | 269 | pkg/cwe/local     | Code    | Add context-based timeouts to all database operations                                     | 150      | 3        |
 | 270 | pkg/attack/local  | Code    | Add context-based timeouts to all database operations                                     | 150      | 3        |
 | 271 | pkg/capec/local   | Code    | Add context-based timeouts to all database operations                                     | 150      | 3        |
 | 272 | pkg/asvs/local    | Code    | Add context-based timeouts to all database operations                                     | 150      | 3        |
 | 273 | pkg/cve/local     | Code    | Add database transaction health checks and monitoring                                      | 200      | 3        |
 | 274 | pkg/notes        | Code    | Add database transaction health checks and monitoring                                      | 200      | 3        |
 | 275 | pkg/ssg/local/store | Code    | Add database transaction health checks and monitoring                                      | 200      | 3        |
 | 276 | pkg/proc/subprocess | Test    | Add race condition tests for concurrent message handling                                 | 300      | 3        |
 | 277 | pkg/cve/taskflow  | Test    | Add race condition tests for concurrent job execution                                   | 300      | 3        |
 | 278 | pkg/meta/fsm     | Test    | Add race condition tests for concurrent state transitions                                | 300      | 3        |
 | 279 | pkg/proc/subprocess | Test    | Add stress tests for high-frequency message processing                                   | 300      | 3        |
 | 280 | pkg/cve/remote   | Test    | Add stress tests for concurrent HTTP requests                                            | 300      | 3        |
 | 281 | pkg/cve/local     | Test    | Add stress tests for concurrent database operations                                      | 300      | 3        |
 | 282 | pkg/notes        | Test    | Add stress tests for concurrent RPC operations                                          | 300      | 3        |
 | 283 | pkg/proc/subprocess | Test    | Add t.Parallel() to all independent tests                                              | 400      | 4        |
 | 284 | pkg/cve/taskflow  | Test    | Add t.Parallel() to all independent tests                                              | 300      | 4        |
 | 285 | pkg/cve/local     | Test    | Add t.Parallel() to all independent tests                                              | 300      | 4        |
 | 286 | pkg/notes        | Test    | Add t.Parallel() to all independent tests                                              | 300      | 4        |
 | 287 | pkg/proc/subprocess | Docs    | Document goroutine lifecycle and cleanup requirements                                     | 150      | 4        |
 | 288 | pkg/cve/taskflow  | Docs    | Document job executor lifecycle and resource management                                    | 150      | 4        |
 | 289 | pkg/meta/fsm     | Docs    | Document provider FSM state transition rules and validation                              | 150      | 4        |
 | 290 | pkg/proc/subprocess | Docs    | Document context propagation requirements for RPC handlers                                | 100      | 4        |
 | 291 | pkg/cve/local     | Docs    | Document database connection pool configuration and tuning                               | 150      | 4        |
 | 292 | pkg/notes        | Docs    | Document RPC handler error handling patterns                                            | 150      | 4        |
 | 293 | pkg/proc/subprocess | Ops     | Add metrics for goroutine count and memory usage                                         | 150      | 3        |
 | 294 | pkg/cve/taskflow  | Ops     | Add metrics for worker pool utilization and queue depth                                   | 150      | 3        |
 | 295 | pkg/cve/local     | Ops     | Add metrics for database connection pool metrics                                        | 150      | 3        |
 | 296 | pkg/cve/remote   | Ops     | Add metrics for HTTP client connection pool and request latency                          | 150      | 3        |
 | 297 | pkg/notes        | Ops     | Add metrics for RPC request/response latency and error rates                              | 150      | 3        |
 | 298 | pkg/proc/subprocess | Ops     | Add health check endpoint for subprocess status                                          | 100      | 3        |
 | 299 | pkg/cve/taskflow  | Ops     | Add health check endpoint for job executor status                                        | 100      | 3        |
 | 300 | pkg/meta/fsm     | Ops     | Add health check endpoint for provider status                                             | 100      | 3        |
 | 301 | pkg/cve/local     | Ops     | Add health check endpoint for database connectivity                                       | 100      | 3        |
 | 302 | pkg/notes        | Ops     | Add health check endpoint for service status                                             | 100      | 3        |
 | 303 | pkg/proc/subprocess | Debug   | Add debug logging for message routing and handler dispatch                                | 150      | 4        |
 | 304 | pkg/cve/taskflow  | Debug   | Add debug logging for job state transitions and worker assignments                       | 150      | 4        |
 | 305 | pkg/meta/fsm     | Debug   | Add debug logging for provider state transitions                                      | 150      | 4        |
 | 306 | pkg/cve/local     | Debug   | Add debug logging for database query execution                                         | 150      | 4        |
 | 307 | pkg/notes        | Debug   | Add debug logging for RPC request/response lifecycle                                    | 150      | 4        |
 | 308 | pkg/cve/remote   | Debug   | Add debug logging for external API requests and retries                                | 150      | 4        |
 | 309 | pkg/proc/subprocess | Config  | Make message batch size configurable via environment variables                             | 100      | 3        |
 | 310 | pkg/cve/taskflow  | Config  | Make worker pool size configurable via environment variables                             | 100      | 3        |
 | 311 | pkg/cve/local     | Config  | Make connection pool sizes configurable via environment variables                        | 100      | 3        |
 | 312 | pkg/cve/remote   | Config  | Make HTTP client timeout and retry limits configurable                                 | 100      | 3        |
 | 313 | pkg/meta/fsm     | Config  | Make FSM state transition timeouts configurable                                        | 100      | 3        |
 | 314 | pkg/proc/subprocess | Feature| Add message replay capability for debugging                                            | 200      | 4        |
 | 315 | pkg/cve/taskflow  | Feature| Add job execution visualization and timeline view                                        | 200      | 4        |
 | 316 | pkg/meta/fsm     | Feature| Add provider state visualization dashboard                                             | 200      | 4        |
 | 317 | pkg/notes        | Feature| Add RPC call tracing and profiling                                                    | 200      | 4        |
 | 318 | pkg/cve/local     | Feature| Add database query execution plan visualization                                         | 200      | 4        |

---

## DONE


| ID  | Package          | Type    | Description                                                                      | Date       |
|-----|------------------|---------|------------------------------------------------------------------------------------|------------|
| 002 | pkg/meta          | Debug   | Investigate and fix pkg/meta/fsm test failures                                     | 2026-02-06 |
| 003 | pkg/rpc          | Test    | Add tests for InvokeRPC, HandleResponse, HandleError methods                     | 2026-02-06 |
| 004 | pkg/analysis      | Test    | Review and verify test levels are correct (Level 1 for basic, Level 2 for database) | 2026-02-06 |
| 034 | General          | Code    | Add tests for list methods with 0% coverage (pkg/attack)            | 2026-02-06 |
| 035 | General          | Code    | Add tests for client methods (pkg/rpc)                                  | 2026-02-06 |
| 036 | General          | Code    | Add constants.go files to multiple packages (graph, jsonutil, testutils, urn, cce, ssg, meta) | 2026-02-06 |
| 037 | General          | Code    | Add test level definitions to all tests (pkg/graph, pkg/asvs, pkg/cce, pkg/rpc) | 2026-02-06 |
| 038 | General          | Code    | Create constants.go file for pkg/analysis                                | 2026-02-06 |
| 039 | General          | Ops     | Split website/lib/types.ts and other description files into shorter files         | 2026-02-06 |
| 040 | General          | Ops     | Create basic test suite for pkg/cce                                    | 2026-02-06 |
| 080 | cmd/v2broker/transport | Code    | Design shared memory ring buffer architecture for UDS transport            | 2026-02-06 |
| 081 | cmd/v2broker/transport | Code    | Implement memfd_create support for zero-copy operations                   | 2026-02-06 |
| 082 | cmd/v2broker/transport | Code    | Add batch acknowledgment mechanism to reduce syscall overhead               | 2026-02-06 |
| 083 | cmd/v2broker/transport | Code    | Create synchronization primitives for shared memory access                 | 2026-02-06 |
| 084 | cmd/v2broker/transport | Code    | Implement fallback to regular UDS if shared memory unavailable          | 2026-02-06 |
| 085 | cmd/v2broker/transport | Test    | Add comprehensive tests for shared memory correctness                | 2026-02-06 |
| 086 | cmd/v2broker/transport | Perf    | Benchmark high-frequency message transmission latency (target: 30-40% reduction) | 2026-02-06 |
| 087 | cmd/v2broker/transport | Docs    | Document shared memory size requirements and tuning parameters         | 2026-02-06 |
| 014 | cmd/v2broker/perf    | Code    | Implement batch size predictor based on historical data patterns       | 2026-02-06 |
| 015 | cmd/v2broker/perf    | Code    | Add dynamic batch size adjustment in Optimizer                        | 2026-02-06 |
| 088 | cmd/v2broker/monitor | Code    | Research eBPF probes relevant to v2e architecture                 | 2026-02-06 |
| 089 | cmd/v2broker/monitor | Code    | Implement eBPF-based kernel-level monitoring for key system calls  | 2026-02-06 |
| 090 | cmd/v2broker/monitor | Code    | Create flame graph generation tool for hotspot identification           | 2026-02-06 |
| 091 | cmd/v2broker/monitor | Code    | Add eBPF metrics collection and aggregation                        | 2026-02-06 |
| 092 | cmd/v2broker/monitor | Code    | Integrate eBPF data into existing metrics pipeline                  | 2026-02-06 |
| 093 | cmd/v2broker/monitor | Code    | Add alerting for anomalous kernel-level behavior                   | 2026-02-06 |
| 094 | cmd/v2broker/monitor | Test    | Test eBPF probes in development environment                    | 2026-02-06 |
| 095 | cmd/v2broker/monitor | Docs    | Document required kernel versions and permissions                | 2026-02-06 |
| 096 | cmd/v2broker/monitor | Perf    | Validate overhead impact on system performance                    | 2026-02-06 |
| 097 | cmd/v2broker/scaling | Code    | Design load prediction model based on historical metrics              | 2026-02-06 |
| 098 | cmd/v2broker/scaling | Code    | Implement ML or statistical model for resource demand prediction         | 2026-02-06 |
| 099 | cmd/v2broker/scaling | Code    | Add proactive resource scaling logic based on predictions             | 2026-02-06 |
| 100 | cmd/v2broker/scaling | Code    | Implement anomaly detection with automatic alerting                | 2026-02-06 |
| 101 | cmd/v2broker/scaling | Code    | Create self-healing capabilities for automatic fault recovery         | 2026-02-06 |
| 102 | cmd/v2broker/scaling | Code    | Add prediction accuracy metrics and model retraining logic           | 2026-02-06 |
| 103 | cmd/v2broker/scaling | Test    | Test prediction model with historical data                       | 2026-02-06 |
 | 104 | cmd/v2broker/scaling | Docs    | Document model training process and feature selection              | 2026-02-06 |
 | 105 | cmd/v2broker/scaling | Perf    | Benchmark prediction accuracy and scaling effectiveness            | 2026-02-06 |
 | 141 | General          | Ops     | Optimize website build bundle size, achieve better loading speed      | 2026-02-06 |
 | 072 | pkg/cve/remote   | Code    | Upgrade HTTP clients from HTTP/1.1 to HTTP/2                          | 2026-02-06 |
 | 076 | pkg/cve/remote   | Code    | Implement persistent HTTP/2 connection pools                          | 2026-02-06 |
 | 066 | cmd/v2access     | Docs    | Ensure service.md is up to date with current implementation          | 2026-02-06 |
 | 067 | cmd/v2local      | Docs    | Ensure service.md is up to date with current implementation          | 2026-02-06 |
 | 068 | cmd/v2remote     | Docs    | Ensure service.md is up to date with current implementation          | 2026-02-06 |
 | 069 | cmd/v2broker     | Docs    | Ensure service.md is up to date with current implementation          | 2026-02-06 |
 | 070 | cmd/v2sysmon     | Docs    | Ensure service.md is up to date with current implementation          | 2026-02-06 |
 | 071 | cmd/v2meta       | Docs    | Ensure service.md is up to date with current implementation          | 2026-02-06 |
 | 142 | website           | Critical| Fix missing node_modules - all dependencies show as MISSING           | 2026-02-06 |
  | 143 | website           | Critical| Fix TypeScript compiler accessibility - tsc not found                  | 2026-02-06 |
