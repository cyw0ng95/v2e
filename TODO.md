# v2e Go Packages Maintain TODO List

| ID | Package | Type | Description | Estimate LoC | Priority | Mark WONTFIX? |
|----|---------|------|-------------|--------------|----------|---------------|
| TODO-003 | analysis | Test | Add integration tests for graph analysis workflows | 300 | Medium | |
| TODO-004 | asvs | Optimization | Optimize CSV import with streaming and parallel processing | 100 | Medium | |
| TODO-005 | asvs | Feature | Add incremental update support for CSV imports | 80 | Medium | |
| TODO-006 | attack | Refactor | Simplify Excel parsing logic by extracting helper functions | 120 | Medium | |
| TODO-009 | capec | Refactor | Deduplicate duplicate code in ImportFromXML transaction handling | 60 | Medium | |
| TODO-010 | cce | Feature | Implement CCE update API with field-level diffing | 120 | Low | |
| TODO-011 | cce | Test | Add table-driven tests for all CRUD operations | 200 | Medium | |
| TODO-012 | common | Refactor | Consolidate duplicate logging patterns - cmd/v2broker, cmd/v2meta, cmd/v2access each have custom log wrappers instead of using pkg/common logger consistently | 100 | Low | |
| TODO-054 | all | Refactor | Expand pkg/common/error_registry.go usage - error codes exist in pkg/common (ErrorCode, StandardizedError) but cmd/v2local, cmd/v2remote, cmd/v2meta still use ad-hoc errors instead of standardized ErrorCode system | 400 | Medium | |
| TODO-084 | cmd/v2meta | Refactor | Extract provider FSM state transition logic - pkg/meta/fsm/provider.go has BaseProviderFSM but CVEProvider, CWEProvider, CAPECProvider, ATTACKProvider still have duplicate field-level diff logic in their execute() methods | 250 | Medium | |
| TODO-098 | all | Refactor | Remove duplicate error handling - create centralized error wrapper in pkg/common/error.go that wraps errors with ErrorCode from error_registry.go instead of each service creating custom error types | 150 | Medium | |
| TODO-099 | all | Test | Add race condition tests using -race flag for all critical path functions | 200 | High | |
| TODO-100 | all | Refactor | Consolidate similar test setup/teardown code into test helpers package | 100 | Low | |
| TODO-101 | website | Refactor | Review and optimize website/app/layout.tsx for code quality and performance | 80 | Medium | |
| TODO-102 | website | Refactor | Review and optimize website/components/navbar.tsx for accessibility and responsiveness | 80 | Medium | |
| TODO-103 | website | Refactor | Review and optimize website/components/session-control.tsx for state management | 80 | Medium | |
| TODO-104 | website | Refactor | Review and optimize website/lib/rpc-client.ts for error handling and retry logic | 100 | Medium | |
| TODO-105 | website | Refactor | Review and optimize website/app/glc/page.tsx and related components for performance | 120 | Medium | |
| TODO-106 | website | Refactor | Review and optimize website/components/providers/ for code consistency and reusability | 100 | Medium | |
| TODO-108 | website | Refactor | Review and optimize website/app/page.tsx for bundle size and performance | 80 | Medium | |
| TODO-109 | website | Test | Add unit tests for website/lib/utils.ts utility functions | 100 | Medium | |
| TODO-110 | website | Refactor | Review and optimize website/components/cve-detail-modal.tsx for UX and performance | 80 | Medium | |
| TODO-111 | test | Review | Review and add missing test cases for pkg/common package (66.5% coverage) - focus on error_registry.go, workerpool/, procfs/ | 150 | Medium | |
| TODO-114 | test | Review | Review and add missing test cases for pkg/attack/provider package (0% coverage) - add tests for provider state transitions, data fetching | 200 | High | |
| TODO-115 | test | Review | Review and add missing test cases for pkg/asvs package (5.7% coverage) - focus on local.go CRUD operations | 200 | High | |
| TODO-116 | test | Review | Review and add missing test cases for pkg/meta package (0% coverage) - add tests for fsm/, storage/ | 200 | High | |
| TODO-117 | test | Review | Review and add missing test cases for pkg/cwe package (0% coverage) - focus on local.go GetByID, ListCWEsPaginated | 200 | High | |
| TODO-118 | test | Review | Review and add missing test cases for cmd/v2broker/core package (11.3% coverage) - focus on rpc.go, routing.go, broker.go | 150 | High | |
| TODO-119 | test | Review | Review and add missing test cases for cmd/v2access package (0% coverage) - add tests for handlers and middleware | 200 | High | |
| TODO-120 | test | Review | Review and add missing test cases for cmd/v2local package (3.4% coverage) - focus on *_handlers.go files | 150 | High | |
| TODO-126 | cve | Test | Add tests for CVEProvider.execute() checkpoint saving logic when RPC calls fail | 80 | Medium | |
| TODO-127 | cve | Documentation | Document exported types in pkg/cve/types.go - CVSSDataV40, CVSSMetricV40 lack comments explaining v4.0 differences | 60 | Low | |
| TODO-129 | cwe | Refactor | Reduce code duplication in pkg/cwe/local.go - GetByID (line ~600) and ListCWEsPaginated (line ~450) share ~90% identical nested field loading logic for RelatedAttackPatterns, ExternalReferences, Notes, WeaknessRhetoric | 150 | Medium | |
| TODO-130 | cwe | Optimization | Use GORM Preload for nested relations in pkg/cve/local.go and pkg/cwe/local.go - GetByID/ListCWEsPaginated currently use N+1 queries instead of eager loading | 100 | Medium | |
| TODO-136 | capec | Refactor | Eliminate code duplication between LocalCAPECStore and CachedLocalCAPECStore in pkg/capec/ - both implement ImportFromXML separately, should share via base struct or interface | 180 | Medium | |
| TODO-144 | cmd/v2broker | Documentation | Document SendQuotaUpdateEvent method in cmd/v2broker/service.md | 20 | Low | |
| TODO-145 | cmd/v2broker | Test | Add integration tests for main.go covering signal handling and graceful shutdown flow | 150 | Medium | |
| TODO-151 | cmd/v2access | Test | Add tests for graceful shutdown logic in run.go | 100 | Medium | |
| TODO-157 | cmd/v2remote | Test | Add tests for FSM recovery scenarios in cmd/v2meta | 150 | Medium | |
| TODO-158 | cmd/v2remote | Test | Add tests for CAPEC handlers in cmd/v2remote | 100 | Medium | |
| TODO-161 | cmd/v2meta | Test | Add tests for FSM recovery scenarios | 150 | Medium | |
| TODO-166 | cmd/v2local | Documentation | Document database schema and access patterns | 60 | Low | |
| TODO-167 | cmd/v2local | Documentation | Document RPC handlers in cmd/v2local/service.md | 80 | Low | |
| TODO-169 | website | Bug Fix | Fix race conditions in hooks during rapid unmount - add AbortController usage and cleanup functions to data-fetching hooks in lib/hooks.ts | 150 | High | |
| TODO-171 | website | Optimization | Implement React.memo optimization for table row components to improve scroll performance for large datasets | 80 | Medium | |
| TODO-172 | website | Optimization | Implement virtualization for horizontal tab lists when there are 10+ tabs | 100 | Low | |
| TODO-173 | website | Optimization | Add route-based code splitting for major pages using Next.js dynamic imports | 60 | Medium | |
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
| TODO-191 | analysis | Optimization | Add sync.Pool buffer pool for JSON serialization in pkg/analysis/storage/graph_store.go to reduce GC pressure during graph persistence | 80 | Medium | |
| TODO-193 | analysis | Feature | Implement incremental graph save to avoid clearing entire bucket on each SaveGraph call | 200 | Medium | |
| TODO-194 | analysis | Feature | Add batch operations API for saving multiple nodes/edges in single transaction | 150 | Medium | |
| TODO-195 | website | Refactor | Consolidate duplicate mock response logic in website/lib/rpc-client.ts - getMockResponseForCache (line 544) and getMockResponse (line 712) have ~600 lines of duplication handling the same mock data | 100 | Medium | |
| TODO-198 | rpc-client | Optimization | Implement response caching with TTL for frequently accessed read-only endpoints | 150 | Medium | |
| TODO-199 | rpc-client | Test | Add unit tests for RPC client covering case conversion, mock responses, error handling, timeout behavior | 250 | High | |
| TODO-208 | ume | Refactor | Add backpressure mechanism to message routing - when route channels are full, messages are dropped with "channel full" error instead of blocking send, implement proper queue or buffer | 200 | High | |
| TODO-209 | ume | Feature | Add message batching support to router - group multiple messages and route them in batches to improve throughput for high-volume scenarios | 250 | Medium | |
| TODO-210 | ume | Feature | Add message delivery tracking - track whether messages were successfully delivered or dropped for observability and debugging | 150 | Medium | |
| TODO-213 | ume | Feature | Remove or complete shared memory transport - current implementation is incomplete and cannot be used for cross-process communication, should either implement actual fd passing or remove entirely | 150 | High | |
| TODO-219 | ume | Feature | Add connection pooling to UDS transport - UDS transport uses single connection per pair, not a pool for efficiency, should implement connection reuse | 180 | High | |
| TODO-220 | ume | Feature | Add message delivery guarantees - router currently provides no delivery confirmation, messages can be silently dropped, should implement ack/nack mechanism | 200 | High | |
| TODO-224 | ume | Feature | Implement actual fd passing to subprocesses - shared memory transport needs mechanism to pass file descriptors to subprocesses for true IPC, not memfd_create which is same-process only | 250 | High | |
| TODO-225 | ume | Feature | Implement SelectTarget method in Router interface - task description mentioned SelectTarget method but router interface doesn't define it, routing cannot select targets dynamically | 150 | High | |
| TODO-226 | ume | Refactor | Add pool statistics for Message Pool - ResponseBufferPool is missing hit/miss tracking which is needed for optimization and debugging | 80 | Medium | |
| TODO-229 | sysmon | Feature | Add configurable sampling intervals - setSamplingInterval() and getSamplingInterval() functions exist but are not exposed via RPC, users cannot configure sampling rates at runtime | 100 | Medium | |
| TODO-230 | sysmon | Feature | Add active/passive health monitoring - service only responds to RPCGetSysMetrics requests, there's no push-based or scheduled monitoring that could alert on threshold crossings or process death | 200 | High | |
| TODO-231 | sysmon | Feature | Add process-level metrics collection - missing per-process resource usage monitoring (goroutine count, memory per subprocess), sysmon could query broker's process stats and expose new RPCGetProcessMetrics that returns per-subprocess resource usage | 180 | High | |
| TODO-232 | sysmon | Refactor | Improve metric error handling - when optional metrics fail they're silently skipped with continue statement, should at least log metric-specific failures for debugging | 100 | Medium | |
| TODO-233 | sysmon | Feature | Add historical data storage - metrics are collected with sampling intervals but no persistent storage for trend analysis or historical querying | 150 | Medium | |
| TODO-234 | sysmon | Feature | Add shutdown timeout - Stop() method waits indefinitely via wg.Wait(), in production a timeout would prevent hanging if a handler is stuck | 120 | Medium | |
| TODO-235 | sysmon | Feature | Add graceful shutdown hook for handlers - handlers cannot register cleanup functions to run when shutdown is signaled, no way to do cleanup on shutdown | 150 | High | |
| TODO-238 | sysmon | Feature | Move goroutine and connection metrics to sysmon - broker's scaling module tracks goroutine count and connections but doesn't actually collect these values, it receives them via AddMetric() calls, sysmon could add RPCGetProcessMetrics that returns per-subprocess resource usage | 180 | High | |
| TODO-240 | cwe | Documentation | Document delete order in SaveView for foreign key safety and why cascading deletes work correctly | 40 | Medium | |
| TODO-243 | attack | Refactor | Use UUID for relationship ID generation instead of sheetIndex-based format to avoid duplicates in concurrent imports | 40 | Medium | |
| TODO-244 | attack | Feature | Track and log skipped Excel rows for observability and debugging | 50 | Medium | |
| TODO-245 | asvs | Feature | Make HTTP timeout configurable instead of hardcoded 30 seconds | 25 | Low | |
| TODO-246 | asvs | Security | Add URL validation before HTTP requests to prevent SSRF attacks | 30 | Low | |
| TODO-247 | ssg | Refactor | Implement savepoints for large transactions to improve rollback granularity | 80 | Medium | |
| TODO-248 | ssg | Optimization | Add query result caching for frequently accessed tree structures | 60 | Medium | |
| TODO-250 | cce | Feature | Add max pagination limit validation (cap at 1000) to prevent excessive queries | 30 | Low | |
| TODO-251 | cce | Refactor | Create generic toModel function to eliminate manual CCE->CCEModel mapping code duplication | 80 | High | |
| TODO-253 | common | Optimization | Optimize ErrorRegistry pattern matching with case-insensitive map for O(1) lookup | 80 | Medium | |
| TODO-254 | notes | Optimization | Optimize Manager.GetContext to avoid copying viewedItems slice on every call | 50 | Medium | |
| TODO-256 | rpc | Optimization | Optimize InvokeRPC defer execution logic to avoid unnecessary cleanup calls | 80 | Medium | |
| TODO-264 | uptime | Feature | Add uptime monitoring service with configurable polling intervals | 200 | Medium | |
| TODO-265 | asvs | Optimization | Create global HTTP client with connection pooling in pkg/asvs/local.go - line~67 creates new http.Client on each ImportFromCSV call, should use shared client | 40 | High | |
| TODO-266 | asvs | Optimization | Enable GORM PrepareStmt:true in pkg/asvs/local.go:35 and pkg/cwe/local.go:372 - currently PrepareStmt:false, enables prepared statements for better query performance | 20 | High | |
| TODO-267 | cwe | Optimization | Enable GORM PrepareStmt:true in pkg/cwe/local.go:372 - currently PrepareStmt:false, enables prepared statements for better query performance | 20 | High | |
| TODO-271 | rpc | Optimization | Use strings.Builder for correlationID generation in pkg/rpc/client.go:132 to reduce allocation overhead from fmt.Sprintf | 40 | Low | |
| TODO-254 | notes | Optimization | Optimize Manager.GetContext in pkg/notes/strategy/manager.go to avoid copying viewedItems slice - current implementation returns copy on every call | 50 | Medium | |
| TODO-275 | notes | Refactor | BFSStrategy and DFSStrategy should embed *BaseStrategy - pkg/notes/strategy/base_strategy.go has reusable code but bfs.go and dfs.go don't use it, causing duplicate GetViewedCount (bfs:104, dfs:178), Reset, viewed map management | 80 | Medium | |
| TODO-276 | rpc | Refactor | Consolidate RPC client implementations - pkg/rpc/client.go (RequestEntry line~23, InvokeRPC line~118) and pkg/notes/rpc_client.go (requestEntry line~23, InvokeRPC line~79) are nearly identical, should share a common implementation | 100 | High | |
| TODO-277 | common | Optimization | Pre-allocate map capacity in hot paths - pkg/graph/graph.go nodes/edges/reverseEdges (line~51), pkg/notes/strategy viewed maps (bfs:19, dfs:58), pkg/notes/fsm/learning_fsm links (line~78) | 60 | Medium | |
| TODO-278 | meta | Refactor | Create shared Event type - pkg/meta/fsm/types.go (Event line~81) and pkg/analysis/fsm/types.go (Event line~78) both define identical structs with Type, Timestamp, Data fields | 40 | Medium | |
| TODO-279 | meta | Refactor | Extract common field diff logic from provider packages - cmd/v2meta/providers/cve_provider.go:250, cwe_provider.go:251, capec_provider.go:251, attack_provider.go:250 all use identical `changed := make(map[string]interface{})` pattern | 60 | Medium | |
| TODO-280 | graph | Optimization | Pre-allocate capacity in pkg/graph/graph.go - neighborsMap in GetNeighbors (line~143), edges slice in AddEdge (line~100) should use make with capacity hint | 40 | Medium | |
## TODO Management Guidelines

For detailed guidelines on managing TODO items, see the **TODO Management** section in `CLAUDE.md`.

### Quick Reference

**Adding Tasks**: Use next available ID (continue from last TODO-NNN), include test code in LoC estimates, write detailed actionable descriptions. **NEVER reuse existing task IDs** - always find the highest current ID and increment by 1.

**Description Requirements**: When adding new tasks, be specific and actionable:
- Include exact file paths and function names (e.g., "Refactor GetByID in pkg/capec/local.go to reduce duplication")
- Specify the current problem or code smell to fix
- If refactoring, mention which files have the duplicate code
- If optimizing, indicate the hot path or benchmark data if available
- Avoid vague descriptions like "Improve performance" or "Fix duplication"

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