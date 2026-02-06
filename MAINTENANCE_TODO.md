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

---

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
| 216 | pkg/cve/taskflow/executor | Docs | Add inline comments for job state machine logic                                          | 150      | 4        |
| 217 | pkg/ssg/parser/guide | Docs  | Add inline comments for guide parsing algorithm                                           | 100      | 4        |
| 218 | pkg/meta/fsm/provider | Docs  | Add inline comments for provider state transitions                                       | 100      | 4        |
| 219 | pkg/notes        | Test    | Add integration tests for RPC handler workflows                                         | 400      | 3        |
| 220 | pkg/cve/local     | Test    | Add integration tests for database migration scenarios                                     | 300      | 3        |
| 221 | pkg/cve/taskflow  | Test    | Add integration tests for job failure and recovery scenarios                             | 400      | 3        |
| 222 | pkg/proc/subprocess | Code   | Implement graceful shutdown with context cancellation for all goroutines                    | 300      | 2        |
| 223 | cmd/v2broker/transport | Code | Implement graceful shutdown for all connection handlers                                   | 250      | 2        |
| 224 | pkg/cve/local     | Code    | Add database transaction rollback verification in error paths                              | 150      | 2        |
| 225 | pkg/proc/subprocess | Code   | Add circuit breaker pattern for RPC client calls                                         | 300      | 3        |
| 226 | pkg/cve/remote   | Code    | Add request deduplication to prevent duplicate API calls                                 | 200      | 3        |
| 227 | pkg/cve/local     | Code    | Add query result caching for frequently accessed data                                      | 300      | 3        |
| 228 | pkg/notes        | Code    | Add query result caching for frequently accessed bookmarks/notes                              | 300      | 3        |
| 229 | pkg/proc/subprocess | Code   | Optimize message serialization by reusing buffers for similar messages                   | 250      | 3        |
| 230 | pkg/cve/local     | Code    | Optimize JSON unmarshaling by using typed structs instead of map[string]interface{}          | 400      | 2        |
| 231 | pkg/cve/remote   | Code    | Optimize HTTP client by reusing transport connections efficiently                         | 200      | 3        |
| 232 | cmd/v2broker/perf | Code    | Add adaptive worker pool scaling based on message queue depth                              | 400      | 2        |
| 233 | pkg/notes        | Code    | Add database indexes for frequently queried fields (review and optimize)                    | 200      | 3        |
| 234 | pkg/cve/local     | Code    | Add database indexes for frequently queried fields (review and optimize)                    | 200      | 3        |
| 235 | pkg/ssg/local     | Code    | Add database indexes for frequently queried fields (review and optimize)                    | 200      | 3        |
| 236 | pkg/proc/subprocess | Security| Add rate limiting per RPC handler to prevent abuse                                       | 200      | 2        |
| 237 | pkg/notes        | Security| Add authorization checks for bookmark/note access                                      | 300      | 2        |
| 238 | pkg/cve/local     | Security| Add input validation for CVE IDs to prevent injection                                   | 150      | 2        |
| 239 | pkg/cve/remote   | Security| Add rate limiting for external API calls to prevent blocking                             | 200      | 2        |
| 240 | pkg/notes        | Code    | Reduce interface{} usage in RPC payloads by defining typed structs                         | 600      | 2        |
| 241 | pkg/ssg/local     | Code    | Reduce interface{} usage in database models by defining typed structs                       | 500      | 2        |
| 242 | pkg/cve/local     | Code    | Reduce interface{} usage in query results by using typed structs                          | 400      | 2        |
| 243 | pkg/meta/fsm     | Test    | Add comprehensive tests for FSM state transitions                                        | 400      | 2        |
| 244 | pkg/cve/taskflow  | Test    | Add comprehensive tests for leak detector                                              | 300      | 3        |
| 245 | pkg/cve/taskflow  | Test    | Add comprehensive tests for tiered pool                                               | 300      | 3        |
| 246 | pkg/cve/remote   | Test    | Add comprehensive tests for adaptive retry logic                                         | 300      | 3        |
| 247 | pkg/cve/remote   | Test    | Add comprehensive tests for circuit breaker                                              | 300      | 3        |
| 248 | pkg/proc/subprocess | Test   | Add comprehensive tests for message batching                                           | 300      | 3        |
| 249 | pkg/proc/subprocess | Test   | Add comprehensive tests for reconnection logic                                          | 300      | 3        |
| 250 | cmd/v2broker/transport | Test | Add comprehensive tests for UDS transport error handling                                   | 400      | 2        |
| 251 | cmd/v2broker/transport | Test | Add comprehensive tests for shared memory transport                                      | 400      | 2        |
| 252 | pkg/notes        | Perf    | Benchmark RPC handler performance with typed vs untyped payloads                         | 200      | 4        |
| 253 | pkg/cve/local     | Perf    | Benchmark database query performance with different indexing strategies                     | 300      | 4        |
| 254 | pkg/cve/remote   | Perf    | Benchmark HTTP client performance with different connection pool configurations               | 300      | 4        |
| 255 | pkg/proc/subprocess | Perf   | Benchmark message throughput with different batch sizes                                  | 300      | 4        |
| 256 | cmd/v2broker/perf | Perf    | Benchmark optimizer performance with different worker counts                                | 300      | 4        |
| 257 | pkg/notes        | Docs    | Document RPC API contract with request/response examples                                   | 400      | 4        |
| 258 | pkg/cve/local     | Docs    | Document database schema and indexing strategy                                           | 300      | 4        |
| 259 | pkg/cve/remote   | Docs    | Document external API integration and retry strategy                                       | 300      | 4        |
| 260 | pkg/proc/subprocess | Docs   | Document subprocess lifecycle and reconnection behavior                                   | 300      | 4        |
| 261 | cmd/v2broker/transport | Docs | Document transport layer architecture and configuration options                            | 400      | 4        |
| 262 | pkg/cve/taskflow  | Docs    | Document job execution flow and error handling                                          | 300      | 4        |
| 263 | pkg/meta/fsm     | Docs    | Document provider state machine and transitions                                          | 300      | 4        |
| 264 | pkg/notes        | Code    | Extract mock data generation into separate test utilities package                           | 200      | 4        |
| 265 | pkg/cve/local     | Code    | Extract database test fixtures into separate package                                    | 200      | 4        |
| 266 | pkg/cve/remote   | Code    | Extract API mock responses into separate test fixtures package                             | 200      | 4        |
| 267 | pkg/proc/subprocess | Code   | Add telemetry for message processing latency percentiles                                  | 200      | 3        |
| 268 | pkg/cve/local     | Code    | Add telemetry for database query latency percentiles                                     | 200      | 3        |
| 269 | pkg/cve/remote   | Code    | Add telemetry for external API call latency percentiles                                   | 200      | 3        |
| 270 | pkg/notes        | Code    | Add telemetry for RPC handler latency percentiles                                       | 200      | 3        |
| 271 | cmd/v2broker/perf | Code    | Add telemetry for optimizer queue depth and processing time                               | 200      | 3        |
| 272 | pkg/proc/subprocess | Code   | Implement backpressure handling when message queue is full                                 | 300      | 2        |
| 273 | pkg/cve/local     | Code    | Implement backpressure handling for database connection pool exhaustion                     | 300      | 2        |
| 274 | pkg/cve/remote   | Code    | Implement backpressure handling for external API rate limiting                                | 300      | 2        |
| 275 | pkg/notes        | Code    | Implement exponential backoff for transient failures                                      | 200      | 3        |
| 276 | pkg/cve/local     | Code    | Implement exponential backoff for database lock conflicts                                | 200      | 3        |
| 277 | pkg/cve/remote   | Code    | Implement exponential backoff for network failures                                       | 200      | 3        |
| 278 | pkg/notes        | Code    | Add request/response size validation to prevent memory exhaustion                        | 150      | 2        |
| 279 | pkg/proc/subprocess | Code   | Add message size validation to prevent memory exhaustion                                  | 150      | 2        |
| 280 | pkg/cve/remote   | Code    | Add response size limits to prevent memory exhaustion from large API responses             | 150      | 2        |
| 281 | pkg/cve/local     | Code    | Add query result size limits to prevent memory exhaustion                                  | 150      | 2        |
| 282 | pkg/notes        | Code    | Add pagination for all list endpoints to prevent large result sets                        | 300      | 2        |
| 283 | pkg/cve/local     | Code    | Add pagination for all list endpoints to prevent large result sets                        | 300      | 2        |
| 284 | pkg/ssg/local     | Code    | Add pagination for all list endpoints to prevent large result sets                        | 300      | 2        |
| 285 | pkg/proc/subprocess | Code   | Implement message deduplication to prevent processing duplicate messages                    | 250      | 3        |
| 286 | pkg/cve/taskflow  | Code    | Implement job deduplication to prevent duplicate job execution                             | 250      | 3        |
| 287 | pkg/notes        | Code    | Add audit logging for all state-changing operations (create, update, delete)                 | 300      | 3        |
| 288 | pkg/cve/local     | Code    | Add audit logging for database modifications                                           | 300      | 3        |
| 289 | pkg/proc/subprocess | Code   | Add audit logging for message routing and delivery                                        | 200      | 3        |
| 290 | pkg/cve/taskflow  | Code    | Add job progress tracking and reporting                                              | 300      | 3        |
| 291 | pkg/cve/local     | Code    | Add bulk update operations for efficient data modification                                 | 400      | 3        |
| 292 | pkg/notes        | Code    | Add bulk update operations for efficient bookmark/note modification                          | 400      | 3        |
| 293 | pkg/proc/subprocess | Code   | Implement message compression for large payloads                                          | 300      | 4        |
| 294 | pkg/cve/remote   | Code    | Implement response compression for large API responses                                    | 300      | 4        |
| 295 | pkg/notes        | Code    | Implement response compression for large data payloads                                     | 300      | 4        |
| 296 | pkg/cve/taskflow  | Code    | Add job priority queue for critical jobs                                             | 400      | 3        |
| 297 | pkg/proc/subprocess | Code   | Add message priority queue for critical messages                                         | 300      | 3        |
| 298 | pkg/cve/remote   | Code    | Add request priority queue for critical API calls                                         | 300      | 3        |
| 299 | pkg/notes        | Code    | Implement optimistic concurrency control for bookmark/note updates                     | 400      | 3        |


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
| 216 | pkg/cve/taskflow/executor | Docs | Add inline comments for job state machine logic                                          | 150      | 4        |
| 217 | pkg/ssg/parser/guide | Docs  | Add inline comments for guide parsing algorithm                                           | 100      | 4        |
| 218 | pkg/meta/fsm/provider | Docs  | Add inline comments for provider state transitions                                       | 100      | 4        |
| 219 | pkg/notes        | Test    | Add integration tests for RPC handler workflows                                         | 400      | 3        |
| 220 | pkg/cve/local     | Test    | Add integration tests for database migration scenarios                                     | 300      | 3        |
| 221 | pkg/cve/taskflow  | Test    | Add integration tests for job failure and recovery scenarios                             | 400      | 3        |
| 222 | pkg/proc/subprocess | Code   | Implement graceful shutdown with context cancellation for all goroutines                    | 300      | 2        |
| 223 | cmd/v2broker/transport | Code | Implement graceful shutdown for all connection handlers                                   | 250      | 2        |
| 224 | pkg/cve/local     | Code    | Add database transaction rollback verification in error paths                              | 150      | 2        |
| 225 | pkg/proc/subprocess | Code   | Add circuit breaker pattern for RPC client calls                                         | 300      | 3        |
| 226 | pkg/cve/remote   | Code    | Add request deduplication to prevent duplicate API calls                                 | 200      | 3        |
| 227 | pkg/cve/local     | Code    | Add query result caching for frequently accessed data                                      | 300      | 3        |
| 228 | pkg/notes        | Code    | Add query result caching for frequently accessed bookmarks/notes                              | 300      | 3        |
| 229 | pkg/proc/subprocess | Code   | Optimize message serialization by reusing buffers for similar messages                   | 250      | 3        |
| 230 | pkg/cve/local     | Code    | Optimize JSON unmarshaling by using typed structs instead of map[string]interface{}          | 400      | 2        |
| 231 | pkg/cve/remote   | Code    | Optimize HTTP client by reusing transport connections efficiently                         | 200      | 3        |
| 232 | cmd/v2broker/perf | Code    | Add adaptive worker pool scaling based on message queue depth                              | 400      | 2        |
| 233 | pkg/notes        | Code    | Add database indexes for frequently queried fields (review and optimize)                    | 200      | 3        |
| 234 | pkg/cve/local     | Code    | Add database indexes for frequently queried fields (review and optimize)                    | 200      | 3        |
| 235 | pkg/ssg/local     | Code    | Add database indexes for frequently queried fields (review and optimize)                    | 200      | 3        |
| 236 | pkg/proc/subprocess | Security| Add rate limiting per RPC handler to prevent abuse                                       | 200      | 2        |
| 237 | pkg/notes        | Security| Add authorization checks for bookmark/note access                                      | 300      | 2        |
| 238 | pkg/cve/local     | Security| Add input validation for CVE IDs to prevent injection                                   | 150      | 2        |
| 239 | pkg/cve/remote   | Security| Add rate limiting for external API calls to prevent blocking                             | 200      | 2        |
| 240 | pkg/notes        | Code    | Reduce interface{} usage in RPC payloads by defining typed structs                         | 600      | 2        |
| 241 | pkg/ssg/local     | Code    | Reduce interface{} usage in database models by defining typed structs                       | 500      | 2        |
| 242 | pkg/cve/local     | Code    | Reduce interface{} usage in query results by using typed structs                          | 400      | 2        |
| 243 | pkg/meta/fsm     | Test    | Add comprehensive tests for FSM state transitions                                        | 400      | 2        |
| 244 | pkg/cve/taskflow  | Test    | Add comprehensive tests for leak detector                                              | 300      | 3        |
| 245 | pkg/cve/taskflow  | Test    | Add comprehensive tests for tiered pool                                               | 300      | 3        |
| 246 | pkg/cve/remote   | Test    | Add comprehensive tests for adaptive retry logic                                         | 300      | 3        |
| 247 | pkg/cve/remote   | Test    | Add comprehensive tests for circuit breaker                                              | 300      | 3        |
| 248 | pkg/proc/subprocess | Test   | Add comprehensive tests for message batching                                           | 300      | 3        |
| 249 | pkg/proc/subprocess | Test   | Add comprehensive tests for reconnection logic                                          | 300      | 3        |
| 250 | cmd/v2broker/transport | Test | Add comprehensive tests for UDS transport error handling                                   | 400      | 2        |
| 251 | cmd/v2broker/transport | Test | Add comprehensive tests for shared memory transport                                      | 400      | 2        |
| 252 | pkg/notes        | Perf    | Benchmark RPC handler performance with typed vs untyped payloads                         | 200      | 4        |
| 253 | pkg/cve/local     | Perf    | Benchmark database query performance with different indexing strategies                     | 300      | 4        |
| 254 | pkg/cve/remote   | Perf    | Benchmark HTTP client performance with different connection pool configurations               | 300      | 4        |
| 255 | pkg/proc/subprocess | Perf   | Benchmark message throughput with different batch sizes                                  | 300      | 4        |
| 256 | cmd/v2broker/perf | Perf    | Benchmark optimizer performance with different worker counts                                | 300      | 4        |
| 257 | pkg/notes        | Docs    | Document RPC API contract with request/response examples                                   | 400      | 4        |
| 258 | pkg/cve/local     | Docs    | Document database schema and indexing strategy                                           | 300      | 4        |
| 259 | pkg/cve/remote   | Docs    | Document external API integration and retry strategy                                       | 300      | 4        |
| 260 | pkg/proc/subprocess | Docs   | Document subprocess lifecycle and reconnection behavior                                   | 300      | 4        |
| 261 | cmd/v2broker/transport | Docs | Document transport layer architecture and configuration options                            | 400      | 4        |
| 262 | pkg/cve/taskflow  | Docs    | Document job execution flow and error handling                                          | 300      | 4        |
| 263 | pkg/meta/fsm     | Docs    | Document provider state machine and transitions                                          | 300      | 4        |
| 264 | pkg/notes        | Code    | Extract mock data generation into separate test utilities package                           | 200      | 4        |
| 265 | pkg/cve/local     | Code    | Extract database test fixtures into separate package                                    | 200      | 4        |
| 266 | pkg/cve/remote   | Code    | Extract API mock responses into separate test fixtures package                             | 200      | 4        |
| 267 | pkg/proc/subprocess | Code   | Add telemetry for message processing latency percentiles                                  | 200      | 3        |
| 268 | pkg/cve/local     | Code    | Add telemetry for database query latency percentiles                                     | 200      | 3        |
| 269 | pkg/cve/remote   | Code    | Add telemetry for external API call latency percentiles                                   | 200      | 3        |
| 270 | pkg/notes        | Code    | Add telemetry for RPC handler latency percentiles                                       | 200      | 3        |
| 271 | cmd/v2broker/perf | Code    | Add telemetry for optimizer queue depth and processing time                               | 200      | 3        |
| 272 | pkg/proc/subprocess | Code   | Implement backpressure handling when message queue is full                                 | 300      | 2        |
| 273 | pkg/cve/local     | Code    | Implement backpressure handling for database connection pool exhaustion                     | 300      | 2        |
| 274 | pkg/cve/remote   | Code    | Implement backpressure handling for external API rate limiting                                | 300      | 2        |
| 275 | pkg/notes        | Code    | Implement exponential backoff for transient failures                                      | 200      | 3        |
| 276 | pkg/cve/local     | Code    | Implement exponential backoff for database lock conflicts                                | 200      | 3        |
| 277 | pkg/cve/remote   | Code    | Implement exponential backoff for network failures                                       | 200      | 3        |
| 278 | pkg/notes        | Code    | Add request/response size validation to prevent memory exhaustion                        | 150      | 2        |
| 279 | pkg/proc/subprocess | Code   | Add message size validation to prevent memory exhaustion                                  | 150      | 2        |
| 280 | pkg/cve/remote   | Code    | Add response size limits to prevent memory exhaustion from large API responses             | 150      | 2        |
| 281 | pkg/cve/local     | Code    | Add query result size limits to prevent memory exhaustion                                  | 150      | 2        |
| 282 | pkg/notes        | Code    | Add pagination for all list endpoints to prevent large result sets                        | 300      | 2        |
| 283 | pkg/cve/local     | Code    | Add pagination for all list endpoints to prevent large result sets                        | 300      | 2        |
| 284 | pkg/ssg/local     | Code    | Add pagination for all list endpoints to prevent large result sets                        | 300      | 2        |
| 285 | pkg/proc/subprocess | Code   | Implement message deduplication to prevent processing duplicate messages                    | 250      | 3        |
| 286 | pkg/cve/taskflow  | Code    | Implement job deduplication to prevent duplicate job execution                             | 250      | 3        |
| 287 | pkg/notes        | Code    | Add audit logging for all state-changing operations (create, update, delete)                 | 300      | 3        |
| 288 | pkg/cve/local     | Code    | Add audit logging for database modifications                                           | 300      | 3        |
| 289 | pkg/proc/subprocess | Code   | Add audit logging for message routing and delivery                                        | 200      | 3        |
| 290 | pkg/cve/taskflow  | Code    | Add job progress tracking and reporting                                              | 300      | 3        |
| 291 | pkg/cve/local     | Code    | Add bulk update operations for efficient data modification                                 | 400      | 3        |
| 292 | pkg/notes        | Code    | Add bulk update operations for efficient bookmark/note modification                          | 400      | 3        |
| 293 | pkg/proc/subprocess | Code   | Implement message compression for large payloads                                          | 300      | 4        |
| 294 | pkg/cve/remote   | Code    | Implement response compression for large API responses                                    | 300      | 4        |
| 295 | pkg/notes        | Code    | Implement response compression for large data payloads                                     | 300      | 4        |
| 296 | pkg/cve/taskflow  | Code    | Add job priority queue for critical jobs                                             | 400      | 3        |
| 297 | pkg/proc/subprocess | Code   | Add message priority queue for critical messages                                         | 300      | 3        |
| 298 | pkg/cve/remote   | Code    | Add request priority queue for critical API calls                                         | 300      | 3        |
| 299 | pkg/notes        | Code    | Implement optimistic concurrency control for bookmark/note updates                     | 400      | 3        |
| 300 | pkg/cve/local     | Code    | Implement optimistic concurrency control for CVE updates                                  | 400      | 3        |
