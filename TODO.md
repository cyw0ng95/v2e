# v2e Go Packages Maintain TODO List

| ID | Package | Type | Description | Estimate LoC | Priority | Mark WONTFIX? |
|----|---------|------|-------------|--------------|----------|---------------|
| TODO-001 | analysis | Refactor | Add comprehensive error handling and recovery mechanisms for FSM state transitions | 150 | High | |
| TODO-002 | analysis | Feature | Implement graph persistence with incremental checkpointing | 200 | High | |
| TODO-003 | analysis | Test | Add integration tests for graph analysis workflows | 300 | Medium | |
| TODO-004 | asvs | Optimization | Optimize CSV import with streaming and parallel processing | 100 | Medium | |
| TODO-005 | asvs | Feature | Add incremental update support for CSV imports | 80 | Medium | |
| TODO-006 | attack | Refactor | Simplify Excel parsing logic by extracting helper functions | 120 | Medium | |
| TODO-008 | capec | Feature | Add streaming XML parser for large CAPEC files | 250 | High | |
| TODO-009 | capec | Refactor | Deduplicate duplicate code in ImportFromXML transaction handling | 60 | Medium | |
| TODO-010 | cce | Feature | Implement CCE update API with field-level diffing | 120 | Low | |
| TODO-011 | cce | Test | Add table-driven tests for all CRUD operations | 200 | Medium | |
| TODO-012 | common | Refactor | Consolidate duplicate logging patterns across services | 100 | Low | |
| TODO-013 | common | Feature | Add structured logging with context propagation | 150 | Medium | |
| TODO-014 | cve | Optimization | Add caching layer for frequently accessed CVE items | 180 | Medium | |
| TODO-015 | cve | Feature | Implement incremental sync from NVD API | 300 | High | |
| TODO-016 | cwe | Refactor | Normalize database access patterns with generic repository pattern | 200 | Medium | |
| TODO-018 | glc | Feature | Implement automatic schema migration with version tracking | 150 | Medium | |
| TODO-019 | graph | Feature | Add graph metrics (centrality, clustering coefficient) | 250 | Low | |
| TODO-020 | graph | Optimization | Implement graph compression for large-scale deployments | 300 | Low | |
| TODO-021 | graph | Test | Add property-based tests using testing/quick | 150 | Medium | |
| TODO-022 | jsonutil | Refactor | Unify error handling across jsonutil functions | 80 | Low | |
| TODO-023 | jsonutil | Feature | Add JSON schema validation support | 200 | Low | |
| TODO-025 | meta | Refactor | Extract FSM transition logic into strategy pattern | 150 | Medium | |
| TODO-026 | meta | Test | Add chaos testing for provider coordination | 250 | Medium | |
| TODO-027 | notes | Feature | Add memory card export/import functionality | 200 | Low | |
| TODO-029 | notes | Refactor | Simplify bookmark service with repository pattern | 120 | Medium | |
| TODO-030 | notes | Test | Add performance benchmarks for FSM operations | 100 | Medium | |
| TODO-032 | proc | Optimization | Add message prioritization and backpressure handling | 150 | Medium | |
| TODO-033 | proc | Test | Add fuzz testing for message serialization/deserialization | 150 | Medium | |
| TODO-035 | rpc | Feature | Add request retry with exponential backoff | 120 | Medium | |
| TODO-036 | rpc | Refactor | Implement connection pooling for RPC clients | 100 | Medium | |
| TODO-037 | ssg | Feature | Add incremental SSG data update support with field-level diffing to avoid re-importing entire datasets when only subset changes | 250 | High | |
| TODO-038 | ssg | Optimization | Implement parallel parsing with worker pools for large SSG XML files to reduce import time by 40% | 180 | Medium | |
| TODO-039 | testutils | Feature | Add mock HTTP server for testing remote providers with configurable responses and delay simulation | 150 | Medium | |
| TODO-040 | testutils | Refactor | Extract common test patterns into helpers (database fixtures, assertion helpers, context factories) | 100 | Low | |
| TODO-042 | urn | Test | Add property-based tests using testing/quick for URN parsing edge cases | 120 | Medium | |
| TODO-043 | vconfig | Feature | Add configuration validation with detailed error messages indicating which option failed and why | 120 | High | |
| TODO-044 | vconfig | Bug Fix | Fix TUI rendering issues on terminals with non-standard dimensions (handle resize gracefully) | 80 | High | |
| TODO-045 | vconfig | Feature | Add search/filter functionality in TUI to quickly locate configuration options by name or description | 150 | Medium | |
| TODO-046 | vconfig | Refactor | Extract TUI event handling logic into separate package to improve testability and reduce coupling | 180 | Medium | |
| TODO-047 | vconfig | Feature | Implement configuration profiles to allow switching between different build configurations (dev, test, prod) | 200 | Medium | |
| TODO-048 | vconfig | Test | Add integration tests for TUI interactions using simulated terminal input | 200 | Medium | |
| TODO-049 | vconfig | Documentation | Add detailed inline comments explaining config spec loading and conversion logic | 80 | Low | |
| TODO-050 | analysis | Documentation | Add examples for graph analysis API usage showing path finding, neighbor queries, and metrics | 80 | Low | |
| TODO-051 | meta | Documentation | Document provider FSM lifecycle and transitions with state diagrams and example flows | 100 | Low | |
| TODO-052 | notes | Documentation | Add learning strategy comparison guide explaining when to use BFS vs DFS navigation | 60 | Low | |
| TODO-053 | proc | Documentation | Document message flow and serialization format with examples of each message type | 100 | Low | |
| TODO-054 | all | Refactor | Standardize error codes across all packages with custom error types and error code constants | 400 | Medium | |
| TODO-055 | all | Feature | Add distributed tracing support with OpenTelemetry for request tracking across services | 500 | Low | |
| TODO-056 | all | Test | Add integration test suite for broker-subprocess communication covering all RPC methods | 300 | High | |
| TODO-057 | all | Optimization | Profile and optimize hot paths across all packages using pprof and benchmarking | 200 | Medium | |
| TODO-058 | build.sh | Refactor | Extract common logging functions into a separate script library (runenv.sh and build.sh both use identical logging) | 80 | Medium | |
| TODO-060 | build.sh | Feature | Add caching layer for Go module downloads to speed up CI builds by using GOMODCACHE mount in containerized environments | 150 | Medium | |
| TODO-061 | build.sh | Feature | Implement parallel frontend build with npm build flags to reduce build time on multi-core systems | 120 | Medium | |
| TODO-062 | build.sh | Optimization | Optimize go build arguments by reusing build tags and ldflags from get_config_* functions to avoid duplicate config parsing | 50 | Medium | |
| TODO-063 | build.sh | Refactor | Simplify build_and_package() parallel build logic to extract into reusable function with better error handling | 100 | Medium | |
| TODO-064 | build.sh | Feature | Add build artifact signing or checksum generation for reproducible builds | 100 | Low | |
| TODO-065 | build.sh | Test | Add unit tests for build script helper functions (version_ge, check_go_version, check_node_version) using shunit2 or bats | 200 | Low | |
| TODO-066 | build.sh | Documentation | Add inline comments explaining incremental build detection logic and file timestamp comparison in build_project() | 60 | Low | |
| TODO-067 | runenv.sh | Refactor | Extract Podman container image building logic into separate function to avoid code duplication between macOS and Linux paths | 100 | Medium | |
| TODO-068 | runenv.sh | Feature | Add Docker container support as alternative to Podman for environments without Podman | 150 | Medium | |
| TODO-070 | runenv.sh | Feature | Add container health check with retry logic to ensure container is fully initialized before executing commands | 120 | Medium | |
| TODO-071 | runenv.sh | Refactor | Simplify environment variable passing by using environment file instead of long -e flag chain in run_container_env() | 80 | Medium | |
| TODO-072 | runenv.sh | Documentation | Add inline comments explaining container volume mount paths and Go module cache optimization in run_container_env() | 60 | Low | |
| TODO-073 | cmd/v2access | Refactor | Extract RPC handler registration logic into a shared package to reduce code duplication across cmd/* services | 150 | Medium | |
| TODO-074 | cmd/v2access | Feature | Add request timeout configuration per RPC method to override DefaultRPCTimeout for long-running operations | 120 | Medium | |
| TODO-076 | cmd/v2access | Optimization | Add connection pooling for HTTP clients to reduce connection overhead for repeated requests | 100 | Medium | |
| TODO-077 | cmd/v2broker | Refactor | Extract subprocess lifecycle management into core package to improve testability and separation of concerns | 200 | Medium | |
| TODO-080 | cmd/v2broker | Optimization | Implement request batching for frequently called RPC methods to reduce context switching overhead | 150 | Medium | |
| TODO-081 | cmd/v2local | Refactor | Simplify database connection pooling by using generic pool wrapper instead of repeated SetMaxIdleConns/SetMaxOpenConns calls | 80 | Medium | |
| TODO-082 | cmd/v2local | Feature | Add database query logging with execution time tracking to identify slow queries | 100 | Low | |
| TODO-084 | cmd/v2meta | Refactor | Extract provider FSM state transition logic into shared package to reduce code duplication across CVEProvider, CWEProvider, CAPECProvider, ATTACKProvider | 250 | Medium | |
| TODO-087 | cmd/v2meta | Optimization | Add provider health checks with automatic restart for providers in TERMINATED state that should be running | 120 | Medium | |
| TODO-088 | cmd/v2remote | Refactor | Simplify HTTP client configuration by extracting into shared package with retry logic and timeout handling | 100 | Medium | |
| TODO-091 | cmd/v2remote | Optimization | Implement response caching for frequently accessed API endpoints (e.g., CVE by ID lookup) to reduce API calls | 120 | Medium | |
| TODO-093 | cmd/v2sysmon | Feature | Add alert threshold configuration with webhook or email notifications when metrics exceed defined limits | 150 | Medium | |
| TODO-097 | notes | Test | Add concurrent stress test for LearningFSM state transitions with multiple goroutines | 100 | Medium | |
| TODO-098 | all | Refactor | Remove duplicate error handling patterns across cmd/* services with centralized error wrapper | 150 | Medium | |
| TODO-099 | all | Test | Add race condition tests using -race flag for all critical path functions | 200 | High | |
| TODO-100 | all | Refactor | Consolidate similar test setup/teardown code into test helpers package | 100 | Low | |
| TODO-101 | website | Refactor | Review and optimize website/app/layout.tsx for code quality and performance | 80 | Medium | |
| TODO-102 | website | Refactor | Review and optimize website/components/navbar.tsx for accessibility and responsiveness | 80 | Medium | |
| TODO-103 | website | Refactor | Review and optimize website/components/session-control.tsx for state management | 80 | Medium | |
| TODO-104 | website | Refactor | Review and optimize website/lib/rpc-client.ts for error handling and retry logic | 100 | Medium | |
| TODO-105 | website | Refactor | Review and optimize website/app/glc/page.tsx and related components for performance | 120 | Medium | |
| TODO-106 | website | Refactor | Review and optimize website/components/providers/ for code consistency and reusability | 100 | Medium | |
| TODO-107 | website | Documentation | Add JSDoc comments to website/lib/types.ts for better type documentation | 60 | Low | |
| TODO-108 | website | Refactor | Review and optimize website/app/page.tsx for bundle size and performance | 80 | Medium | |
| TODO-109 | website | Test | Add unit tests for website/lib/utils.ts utility functions | 100 | Medium | |
| TODO-110 | website | Refactor | Review and optimize website/components/cve-detail-modal.tsx for UX and performance | 80 | Medium | |
| TODO-111 | test | Review | Review and add missing test cases for pkg/common package (66.5% coverage) | 150 | Medium | |
| TODO-112 | test | Review | Review and add missing test cases for pkg/analysis package (85.1% coverage) | 100 | Low | |
| TODO-114 | test | Review | Review and add missing test cases for pkg/attack provider package (0% coverage) | 200 | High | |
| TODO-115 | test | Review | Review and add missing test cases for pkg/asvs package (5.7% coverage) | 200 | High | |
| TODO-116 | test | Review | Review and add missing test cases for pkg/meta package (0% coverage) | 200 | High | |
| TODO-117 | test | Review | Review and add missing test cases for pkg/cwe package (0% coverage) | 200 | High | |
| TODO-118 | test | Review | Review and add missing test cases for cmd/v2broker/core package (11.3% coverage) | 150 | High | |
| TODO-119 | test | Review | Review and add missing test cases for cmd/v2access package (0% coverage) | 200 | High | |
| TODO-120 | test | Review | Review and add missing test cases for cmd/v2local package (3.4% coverage) | 150 | High | |
| TODO-126 | cve | Test | Add tests for CVEProvider.execute() checkpoint saving logic when RPC calls fail | 80 | Medium | |
| TODO-127 | cve | Documentation | Document exported types in pkg/cve/types.go - CVSSDataV40, CVSSMetricV40 lack comments explaining v4.0 differences | 60 | Low | |
| TODO-129 | cwe | Refactor | Reduce massive code duplication in local.go - GetByID and ListCWEsPaginated share 90% identical nested field loading logic | 150 | Medium | |
| TODO-130 | cwe | Optimization | Use Preload or eager loading for nested relations instead of N+1 queries in GetByID/ListCWEsPaginated | 100 | Medium | |
| TODO-133 | cwe | Documentation | Document LocalCWEStore.SaveView nested array deletion order and why it's safe | 40 | Low | |
| TODO-136 | capec | Refactor | Eliminate code duplication between LocalCAPECStore and CachedLocalCAPECStore - share ImportFromXML logic via base struct | 180 | Medium | |
| TODO-144 | cmd/v2broker | Documentation | Document SendQuotaUpdateEvent method in cmd/v2broker/service.md | 20 | Low | |
| TODO-145 | cmd/v2broker | Test | Add integration tests for main.go covering signal handling and graceful shutdown flow | 150 | Medium | |
| TODO-148 | cmd/v2access | Optimization | Use sync.Pool for context creation in handlers.go:90 instead of creating new context per RPC call | 60 | Low | |
| TODO-151 | cmd/v2access | Test | Add tests for graceful shutdown logic in run.go | 100 | Medium | |
| TODO-157 | cmd/v2remote | Test | Add tests for FSM recovery scenarios in cmd/v2meta | 150 | Medium | |
| TODO-158 | cmd/v2remote | Test | Add tests for CAPEC handlers in cmd/v2remote | 100 | Medium | |
| TODO-161 | cmd/v2meta | Test | Add tests for FSM recovery scenarios | 150 | Medium | |
| TODO-166 | cmd/v2local | Documentation | Document database schema and access patterns | 60 | Low | |
| TODO-167 | cmd/v2local | Documentation | Document RPC handlers in cmd/v2local/service.md | 80 | Low | |
| TODO-168 | website | Bug Fix | Add global error boundary for RPC failures - implement in app/layout.tsx to catch errors and show user-friendly messages with retry options | 100 | High | |
| TODO-169 | website | Bug Fix | Fix race conditions in hooks during rapid unmount - add AbortController usage and cleanup functions to data-fetching hooks in lib/hooks.ts | 150 | High | |
| TODO-170 | website | Bug Fix | Add timeout handling to all long-running operations with configurable timeout options and user-facing progress indicators | 120 | Medium | |
| TODO-171 | website | Optimization | Implement React.memo optimization for table row components to improve scroll performance for large datasets | 80 | Medium | |
| TODO-172 | website | Optimization | Implement virtualization for horizontal tab lists when there are 10+ tabs | 100 | Low | |
| TODO-173 | website | Optimization | Add route-based code splitting for major pages using Next.js dynamic imports | 60 | Medium | |
| TODO-174 | website | Refactor | Extract duplicate severity utility functions from cve-table.tsx and cve-detail-modal.tsx to lib/utils.ts | 40 | Low | |
| TODO-175 | website | Refactor | Consolidate duplicate ErrorBoundary components in lib/error-handler.tsx and components/error-boundary.tsx | 50 | Low | |
| TODO-176 | website | Refactor | Replace any types with proper TypeScript interfaces in lib/hooks.ts, components/notes-framework.tsx, app/page.tsx | 100 | Medium | |
| TODO-177 | website | Test | Add unit tests for RPC client (lib/rpc-client.ts) covering case conversion, mock responses, error handling, timeout behavior | 250 | High | |
| TODO-178 | website | Test | Add unit tests for custom hooks (lib/hooks.ts) using React Testing Library for critical hooks: useCVEList, useSessionStatus, useStartSession, useMemoryCards | 300 | High | |
| TODO-179 | website | Test | Add integration tests for main page tabs verifying navigation, view/learn mode switching, and data display | 200 | Medium | |
| TODO-180 | website | Test | Add accessibility tests for UI components using jest-axe for Button, Table, Dialog, Form components | 150 | Medium | |
| TODO-182 | website | Security | Sanitize user-generated content in notes-framework.tsx using DOMPurify to prevent XSS attacks | 60 | Medium | |
| TODO-183 | website | Accessibility | Add keyboard navigation support to main page tabs with Arrow keys, Home/End, and proper focus management | 80 | Medium | |
| TODO-184 | website | Accessibility | Add live region announcements for async operations to communicate loading states, errors, and success messages | 100 | Medium | |
| TODO-185 | website | Accessibility | Add proper ARIA labels to graph visualization in graph-viewer.tsx and graph-analysis-page.tsx | 120 | Medium | |
| TODO-186 | website | Documentation | Add JSDoc to all public API functions in lib/rpc-client.ts, lib/hooks.ts, and lib/utils.ts | 150 | Low | |
| TODO-187 | website | Documentation | Document component prop interfaces in components/ with JSDoc comments describing required and optional props | 200 | Low | |
| TODO-188 | website | Feature | Connect navbar search to actual search functionality querying across CVEs, CWEs, CAPECs, and ATT&CK data | 150 | Medium | |
| TODO-189 | website | Feature | Add settings functionality for settings button - create dialog/page for user preferences (theme, API endpoints, data refresh) | 180 | Low | |
| TODO-191 | analysis | Optimization | Add sync.Pool buffer pool for JSON serialization in graph_store.go to reduce GC pressure | 80 | Medium | |
| TODO-193 | analysis | Feature | Implement incremental graph save to avoid clearing entire bucket on each SaveGraph call | 200 | Medium | |
| TODO-194 | analysis | Feature | Add batch operations API for saving multiple nodes/edges in single transaction | 150 | Medium | |
| TODO-195 | rpc-client | Refactor | Consolidate duplicate mock response logic between getMockResponseForCache and getMockResponse (~600 lines of duplication) | 100 | Medium | |
| TODO-197 | rpc-client | Bug Fix | Add proper cleanup strategy for pendingRequests Map to prevent memory leaks from failed requests | 60 | Medium | |
| TODO-198 | rpc-client | Optimization | Implement response caching with TTL for frequently accessed read-only endpoints | 150 | Medium | |
| TODO-199 | rpc-client | Test | Add unit tests for RPC client covering case conversion, mock responses, error handling, timeout behavior | 250 | High | |
| TODO-202 | glc | Bug Fix | Add missing foreign key constraint to ShareLinkModel.GraphID for CASCADE delete on graph deletion - prevents orphaned share links | 40 | Medium | |
| TODO-203 | glc | Bug Fix | Fix JSON field name inconsistency - struct has Relations but JSON tag is relationships, causes frontend/backend communication failure | 30 | Medium | |
| TODO-204 | glc | Bug Fix | Implement password hashing with bcrypt or argon2 for CreateShareLink - currently stores plaintext which is security vulnerability | 60 | High | |
| TODO-205 | glc | Bug Fix | Fix race condition in GetGraphByShareLink - view count update happens after graph return, should update in transaction before return | 50 | Medium | |
| TODO-206 | glc | Bug Fix | Fix generateLinkID length calculation - requesting 8 bytes but only taking first 4 hex characters due to [:8] slice, should use [:length*2] or uuid.New().String()[:8] | 20 | Low | |
| TODO-208 | ume | Refactor | Add backpressure mechanism to message routing - when route channels are full, messages are dropped with "channel full" error instead of blocking send, implement proper queue or buffer | 200 | High | |
| TODO-209 | ume | Feature | Add message batching support to router - group multiple messages and route them in batches to improve throughput for high-volume scenarios | 250 | Medium | |
| TODO-210 | ume | Feature | Add message delivery tracking - track whether messages were successfully delivered or dropped for observability and debugging | 150 | Medium | |
| TODO-212 | ume | Bug Fix | Fix shared memory transport non-functional - SharedMemoryTransport uses memfd_create which only creates memory accessible within same process, there's no actual fd sharing with subprocesses, SendFd returns "not implemented" error | 200 | High | |
| TODO-213 | ume | Feature | Remove or complete shared memory transport - current implementation is incomplete and cannot be used for cross-process communication, should either implement actual fd passing or remove entirely | 150 | High | |
| TODO-214 | ume | Bug Fix | Fix ring buffer calculation in shared_memory.go - remaining := shm.header.Capacity - (shm.header.WritePos % shm.header.Capacity) doesn't correctly calculate remaining space in ring buffer, needs to account for ReadPos and handle wrap-around properly | 60 | Medium | |
| TODO-217 | ume | Bug Fix | Fix fixedBytesToString edge case inconsistency - behavior inconsistent for all-zero byte arrays between different functions, needs unified handling | 40 | Low | |
| TODO-218 | ume | Bug Fix | Fix SetSocketOptions ignoring TCP_QUICKACK error - TCP_QUICKACK socket option error is explicitly ignored, masking legitimate socket configuration failures | 60 | Medium | |
| TODO-219 | ume | Feature | Add connection pooling to UDS transport - UDS transport uses single connection per pair, not a pool for efficiency, should implement connection reuse | 180 | High | |
| TODO-220 | ume | Feature | Add message delivery guarantees - router currently provides no delivery confirmation, messages can be silently dropped, should implement ack/nack mechanism | 200 | High | |
| TODO-222 | ume | Bug Fix | Fix transport manager CloseAll ignoring errors - errors from individual transport Close() calls are silently discarded, making debugging difficult, should be logged or aggregated | 80 | Medium | |
| TODO-224 | ume | Feature | Implement actual fd passing to subprocesses - shared memory transport needs mechanism to pass file descriptors to subprocesses for true IPC, not memfd_create which is same-process only | 250 | High | |
| TODO-225 | ume | Feature | Implement SelectTarget method in Router interface - task description mentioned SelectTarget method but router interface doesn't define it, routing cannot select targets dynamically | 150 | High | |
| TODO-226 | ume | Refactor | Add pool statistics for Message Pool - ResponseBufferPool is missing hit/miss tracking which is needed for optimization and debugging | 80 | Medium | |
| TODO-227 | ume | Bug Fix | Fix UnmarshalBatch not using pooled messages - performance impact for bulk message operations, defeats pool optimization purpose | 100 | High | |
| TODO-228 | sysmon | Bug Fix | Fix collectMetric helper unused - collectMetric() helper doesn't handle errors or return anything, it's unused dead code that should be removed or integrated properly | 40 | Low | |
| TODO-229 | sysmon | Feature | Add configurable sampling intervals - setSamplingInterval() and getSamplingInterval() functions exist but are not exposed via RPC, users cannot configure sampling rates at runtime | 100 | Medium | |
| TODO-230 | sysmon | Feature | Add active/passive health monitoring - service only responds to RPCGetSysMetrics requests, there's no push-based or scheduled monitoring that could alert on threshold crossings or process death | 200 | High | |
| TODO-231 | sysmon | Feature | Add process-level metrics collection - missing per-process resource usage monitoring (goroutine count, memory per subprocess), sysmon could query broker's process stats and expose new RPCGetProcessMetrics that returns per-subprocess resource usage | 180 | High | |
| TODO-232 | sysmon | Refactor | Improve metric error handling - when optional metrics fail they're silently skipped with continue statement, should at least log metric-specific failures for debugging | 100 | Medium | |
| TODO-233 | sysmon | Feature | Add historical data storage - metrics are collected with sampling intervals but no persistent storage for trend analysis or historical querying | 150 | Medium | |
| TODO-234 | sysmon | Feature | Add shutdown timeout - Stop() method waits indefinitely via wg.Wait(), in production a timeout would prevent hanging if a handler is stuck | 120 | Medium | |
| TODO-235 | sysmon | Feature | Add graceful shutdown hook for handlers - handlers cannot register cleanup functions to run when shutdown is signaled, no way to do cleanup on shutdown | 150 | High | |
| TODO-237 | sysmon | Cleanup | Remove unused log constants - many log constants in constants.go are defined but never used (LogMsgCPUUsageCollected, LogMsgMemoryUsageCollected, LogMsgPerformanceMetricsCollected), should be removed or implemented | 20 | Low | |
| TODO-238 | sysmon | Feature | Move goroutine and connection metrics to sysmon - broker's scaling module tracks goroutine count and connections but doesn't actually collect these values, it receives them via AddMetric() calls, sysmon could add RPCGetProcessMetrics that returns per-subprocess resource usage | 180 | High | |

## TODO Management Guidelines

For detailed guidelines on managing TODO items, see the **TODO Management** section in `CLAUDE.md`.

### Quick Reference

**Adding Tasks**: Use next available ID (continue from last TODO-NNN), include test code in LoC estimates, write detailed actionable descriptions. **NEVER reuse existing task IDs** - always find the highest current ID and increment by 1.

**Completing Tasks**:
1. Verify all acceptance criteria are met
2. Run relevant tests (`./build.sh -t`)
3. Remove entire row from table (not just mark as done)
4. Maintain markdown table formatting consistency
5. Commit deletion of task row from TODO.md

**Repository Cleanup**:
- Ensure `.build/*` is ignored via `.gitignore` (pattern: `.build/*`)
- Clean up committed build artifacts if present

**Marking WONTFIX**:
- Only project maintainers can mark tasks as obsolete
- AI agents MUST NOT add "WONTFIX" to any task

**Priority**: High (critical), Medium (important), Low (nice-to-have)

**LoC Estimates**: Small (<100), Medium (100-300), Large (300+)