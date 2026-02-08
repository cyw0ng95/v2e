# MAINTENANCE TODO

This document tracks maintenance tasks for the v2e project. Tasks are organized by priority and type.

## TODO

| ID  | Package           | Type    | Description                                                                        | Est LoC | Priority | WONTFIX |
|-----|-------------------|---------|------------------------------------------------------------------------------------|---------|----------|---------|
| 350 | website/          | Bug     | Dark mode switch button not functional - investigate and fix theme toggle mechanism | 50      | 1        |         |
| 291 | website/          | Feature | Add data validation for all forms (currently minimal validation)                   | 200      | 2        |         |
| 345 | website/          | Perf    | Add React.memo and performance optimizations for large components                  | 400      | 2        |         |
| 346 | website/          | Feature | Implement responsive design improvements for mobile/tablet layouts                 | 300      | 2        |         |
| 347 | website/components| Feature | Add virtualization for large tables (>100 rows)                                    | 200      | 2        |         |
| 348 | website/components| Feature | Create onboarding tour and help system components                                  | 350      | 3        |         |
| 349 | website/          | Docs    | Add JSDoc comments and component documentation                                     | 500      | 3        |         |
| 017 | cmd/v2broker     | Config  | Create configuration hooks for batch size tuning                                   | 200      | 3        |         |
| 048 | cmd/v2broker/perf | Code    | Implement worker affinity to reduce cache misses                                   | 400      | 3        |         |
| 077 | pkg/cve/remote   | Code    | Maintain backward compatibility with HTTP/1.1                                      | 200      | 3        |         |
| 112 | pkg/cve/local    | Code    | Add metrics for connection pool efficiency                                         | 200      | 3        |         |
| 116 | pkg/cve/local    | Code    | Add batch merging strategy to combine small batches                                | 400      | 3        |         |
| 118 | pkg/cve/local    | Code    | Add batch size metrics and tuning recommendations                                  | 200      | 3        |         |
| 136 | pkg/proc         | Code    | Enable response buffer reuse for hot RPC methods                                   | 400      | 3        |         |
| 138 | pkg/proc         | Code    | Add prediction accuracy metrics                                                    | 200      | 3        |         |
| 169 | pkg/notes        | Code    | Extract common RPC validation and error handling patterns into helper functions    | 400      | 3        |         |
| 170 | pkg/notes/rpc_client | Code    | Refactor rpc_client.go (1101 lines) into smaller, focused modules                  | 600      | 3        |         |
| 171 | pkg/notes/service | Code    | Refactor service.go (1006 lines) into smaller, focused modules                     | 500      | 3        |         |
| 172 | pkg/ssg/local/store | Code    | Refactor store.go (828 lines) into smaller modules (models, queries, migrations)   | 400      | 3        |         |
| 173 | pkg/cwe/local    | Code    | Refactor local.go (665 lines) into smaller, focused modules                        | 400      | 3        |         |
| 174 | pkg/cve/taskflow/executor | Code    | Refactor executor.go (612 lines) - extract job state management                    | 350      | 3        |         |
| 175 | pkg/ssg/job/importer | Code    | Refactor importer.go (597 lines) - extract parsing logic                           | 350      | 3        |         |
| 176 | pkg/cve/local/learning | Code    | Refactor learning.go (518 lines) - extract pattern analysis                        | 300      | 3        |         |
| 177 | pkg/ssg/parser/guide | Code    | Refactor guide.go (512 lines) - extract parsing helpers                            | 300      | 3        |         |
| 181 | pkg/proc/subprocess | Code    | Extract subprocess reconnection logic into dedicated module                        | 200      | 3        |         |
| 182 | pkg/proc/subprocess | Code    | Add lock contention metrics for handlers map access                                | 150      | 3        |         |
| 184 | pkg/proc/subprocess | Code    | Add goroutine leak detection and monitoring                                        | 200      | 3        |         |
| 185 | cmd/v2broker/transport | Code    | Add lock contention metrics for transport operations                               | 150      | 3        |         |
| 189 | pkg/cve/local    | Code    | Add sync.Pool for frequently allocated database query structures                   | 200      | 3        |         |
| 192 | pkg/cve/local    | Code    | Implement connection leak detection with automatic cleanup                         | 250      | 3        |         |
| 199 | pkg/notes        | Code    | Add rate limiting for RPC endpoints to prevent abuse                               | 300      | 3        |         |
| 200 | pkg/cve/taskflow/executor | Test    | Add comprehensive concurrent execution tests for job executor                      | 400      | 3        |         |
| 201 | pkg/proc/subprocess | Test    | Add concurrent stress tests for message handling                                   | 300      | 3        |         |
| 202 | cmd/v2broker/transport | Test    | Add concurrent connection handling stress tests                                    | 300      | 3        |         |
| 209 | pkg/notes        | Code    | Add structured logging for RPC request/response tracing                            | 300      | 3        |         |
| 210 | pkg/proc/subprocess | Code    | Add structured logging for message lifecycle events                                | 200      | 3        |         |
| 211 | pkg/cve/taskflow/executor | Code    | Add structured logging for job state transitions                                   | 200      | 3        |         |
| 220 | pkg/cve/taskflow | Test    | Add integration tests for job workflow with subprocesses                           | 400      | 3        |         |
| 252 | website/lib      | Refactor | Split lib/hooks.ts (2439 lines) into separate modules by domain (attack, capec, session, etc.) | 300      | 3        |         |
| 253 | website/lib      | Refactor | Split lib/rpc-client.ts (1975 lines) - separate mock data, client logic, and type imports | 250      | 3        |         |
| 254 | website/lib      | Refactor | Split lib/types.ts (1954 lines) into separate type definition files (cve.ts, attack.ts, notes.ts, etc.) | 200      | 3        |         |
| 255 | website/components | Refactor | Split components/ssg-views.tsx (939 lines) into separate TreeViewNode, DetailPanel, and RuleDetail components | 150      | 3        |         |
| 256 | website/components | Refactor | Split components/notes-framework.tsx (724 lines) - extract bookmark, note, memory card logic into hooks | 200      | 3        |         |
| 257 | website/components | Refactor | Split components/ui/sidebar.tsx (726 lines) - extract sidebar subcomponents        | 150      | 3        |         |
| 258 | website/app      | Refactor | Split app/page.tsx (468 lines) - extract RightColumn component and lazy-loaded imports to separate file | 150      | 3        |         |
| 259 | website/components | Refactor | Split components/etl-topology-viewer.tsx (409 lines) into smaller focused components | 150      | 3        |         |
| 272 | website/         | A11y    | Add aria-label to all buttons without text content (only 19 aria-label attributes found in 55 components) | 100      | 3        |         |
| 273 | website/         | A11y    | Add keyboard navigation support for interactive components (graphs, modals, tables) | 200      | 3        |         |
| 274 | website/         | A11y    | Add proper role attributes to interactive elements (currently only 12 instances with role=) | 80       | 3        |         |
| 275 | website/components | Code    | Fix array index keys in components (notes-framework.tsx:570, notes-dashboard.tsx:259) | 20       | 3        |         |
| 279 | website/         | Perf    | Implement React.memo for 64 components (currently only 24 use useMemo/useCallback) | 300      | 3        |         |
| 329 | pkg/cve/provider  | Code    | Implement store logic using RPC calls to local service (TODO in cve_provider.go)      | 200      | 2        |         |
| 330 | pkg/ssg/provider  | Code    | Implement store logic using RPC calls to ssg service (TODO in git_provider.go)       | 200      | 2        |         |
| 331 | pkg/ssg/parser    | Code    | Parse SSG references (TODO in guide.go)                                      | 150      | 2        |         |
| 332 | pkg/attack/provider | Code    | Use RPCStoreAttack to store each item (TODO in attack_provider.go)                    | 100      | 2        |         |
| 333 | website/components | A11y    | Add accessibility attributes to 12 components lacking aria-label/role           | 200      | 2        |         |
| 334 | website/components | Code    | Fix 18 components using array index as key (React warning)                    | 100      | 2        |         |
| 335 | website/lib       | Types   | Replace 42 `any[]` type usages with proper TypeScript array types              | 150      | 2        |         |
| 336 | website/         | Code    | Fix 2 wildcard React imports (import * as React) - use named imports         | 50       | 2        |         |
| 281 | website/         | Perf    | Add virtualization for large tables (CVE, CWE, CAPEC lists can be thousands of rows) | 200      | 3        |         |
| 282 | website/lib      | Perf    | Optimize data fetching - implement request deduplication and caching for repeated calls | 150      | 3        |         |
| 284 | website/lib      | Docs    | Document RPC API methods in lib/rpc-client.ts (60+ methods need documentation)     | 200      | 3        |         |
| 285 | website/         | Docs    | Create architecture documentation for frontend structure and data flow             | 150      | 3        |         |
| 287 | website/         | Debt    | Remove duplicate code patterns in hooks.ts (16 hooks with nearly identical useEffect patterns) | 150      | 3        |         |
| 288 | website/         | Code    | Remove unused imports - currently no unused import detection                       | 50       | 3        |         |
| 292 | website/components | UX      | Add skeleton loaders for all data fetching operations (currently only in page.tsx dynamic imports) | 100      | 3        |         |
| 294 | website/         | Code    | Review and optimize 173 useEffect, useMemo, useCallback usages for dependency correctness | 200      | 3        |         |
| 245 | pkg/notes/strategy | Perf    | Benchmark strategy switching overhead                                              | 100      | 4        |         |
| 018 | cmd/v2broker     | Test    | Add tests for various data volume scenarios                                        | 400      | 4        |         |
| 019 | cmd/v2broker     | Docs    | Document optimal batch size ranges for different message types                     | 200      | 4        |         |
| 023 | cmd/v2broker     | Perf    | Profile lock contention in current codebase                                        | 400      | 4        |         |
| 024 | cmd/v2broker     | Perf    | Benchmark throughput improvements in high-concurrency (target: 40-50% improvement) | 400      | 4        |         |
| 042 | cmd/v2broker     | Test    | Add comprehensive tests for correctness under concurrent access                    | 600      | 4        |         |
| 043 | cmd/v2broker     | Perf    | Benchmark message routing latency (target: 20-25% reduction)                       | 400      | 4        |         |
| 044 | cmd/v2broker     | Docs    | Document migration strategy from current implementation                            | 200      | 4        |         |
| 052 | cmd/v2broker/perf | Test    | Test under varying load patterns (spiky, steady, gradual)                          | 400      | 4        |         |
| 053 | cmd/v2broker/perf | Perf    | Benchmark CPU utilization improvement (target: 15-20%) and P99 latency reduction   | 400      | 4        |         |
| 079 | pkg/cve/remote   | Test    | Test with external APIs that support HTTP/2                                        | 400      | 4        |         |
| 080 | pkg/cve/remote   | Perf    | Benchmark concurrent request latency (target: 30% reduction) and connection count  | 400      | 4        |         |
| 113 | pkg/cve/local    | Test    | Test with various query patterns                                                   | 400      | 4        |         |
| 119 | pkg/cve/local    | Test    | Test with various data volumes and types                                           | 400      | 4        |         |
| 127 | pkg/cve/local    | Perf    | Benchmark batch insert throughput (target: 2-3x improvement)                       | 400      | 4        |         |
| 134 | pkg/cve/remote   | Perf    | Benchmark external API call success rate (target: 20% improvement)                 | 400      | 4        |         |
| 139 | pkg/proc         | Test    | Test with various message patterns                                                 | 400      | 4        |         |
| 140 | pkg/proc         | Perf    | Benchmark performance improvement (target: 15%)                                    | 400      | 4        |         |
| 203 | pkg/proc/subprocess | Test    | Add benchmark tests for message batching performance                               | 200      | 4        |         |
| 204 | pkg/cve/local    | Test    | Add benchmark tests for database query patterns                                    | 200      | 4        |         |
| 205 | pkg/cve/remote   | Test    | Add benchmark tests for HTTP client connection pooling                             | 200      | 4        |         |
| 206 | cmd/v2broker/perf | Test    | Add benchmark tests for optimizer worker pool                                      | 200      | 4        |         |
| 215 | pkg/notes        | Docs    | Add inline comments for complex business logic in service.go                       | 150      | 4        |         |
| 216 | pkg/notes        | Docs    | Document RPC handler patterns and error handling strategies                        | 200      | 4        |         |
| 217 | pkg/proc/subprocess | Docs    | Document message batching and flush mechanisms                                     | 150      | 4        |         |
| 218 | pkg/cve/local    | Docs    | Document query pattern analysis and caching strategies                             | 200      | 4        |         |
| 219 | pkg/cve/remote   | Docs    | Document external API retry and fallback strategies                                | 200      | 4        |         |
| 276 | website/         | Code    | Ensure all components use consistent import statements (89 files import React, only 23 use named exports) | 50       | 4        |         |
| 280 | website/app      | Perf    | Optimize lazy-loading in page.tsx - combine duplicate loading skeletons (10 identical Skeleton components) | 50       | 4        |         |
| 283 | website/         | Docs    | Add component documentation (JSDoc comments) for all 55 components                 | 300      | 4        |         |
| 286 | website/         | Debt    | Standardize export patterns - choose between default and named exports (32 use default, 23 use named) | 100      | 4        |         |
| 293 | website/         | Types   | Create proper interface for memoized component props (RightColumn in page.tsx uses complex props object) | 50       | 4        |         |
| 030 | cmd/v2analysis   | Test    | Add test level definitions (currently 0 occurrences)                               | 400      | 5        |         |
| 031 | cmd/v2analysis   | Coverage | Improve coverage from 6.8% to at least 50%                                         | 600      | 5        |         |
| 032 | cmd/v2access     | Coverage | Improve coverage from 9.4% to at least 50%                                         | 400      | 5        |         |
| 033 | cmd/v2local      | Test    | Fix test suite (no coverage data available despite 24 test level occurrences)      | 600      | 5        |         |
| 034 | cmd/v2remote     | Test    | Fix test suite (currently 0.0% coverage despite 15 test level occurrences)         | 600      | 5        |         |
| 035 | cmd/v2broker     | Test    | Fix test suite (currently 0.0% coverage despite 4 test level occurrences)          | 600      | 5        |         |
| 036 | cmd/v2sysmon     | Test    | Fix test suite (currently 0.0% coverage despite 3 test level occurrences)          | 600      | 5        |         |
| 037 | cmd/v2meta       | Test    | Fix test suite (no coverage data available despite 35 test level occurrences)      | 600      | 5        |         |
| 054 | pkg/cve          | Test    | Add test level definitions to all tests (currently only 10 occurrences)            | 400      | 5        |         |
| 055 | pkg/cwe          | Coverage | Improve coverage for job package (currently 70.8%)                                 | 400      | 5        |         |
| 056 | pkg/cwe          | Test    | Add test level definitions to all tests (currently 20 occurrences)                 | 400      | 5        |         |
| 057 | pkg/common       | Coverage | Improve coverage for procfs package (currently 0.0%)                               | 400      | 5        |         |
| 058 | pkg/attack       | Coverage | Improve ImportFromXLSX coverage (currently 2.2%)                                   | 400      | 5        |         |
| 060 | pkg/proc         | Coverage | Improve coverage for subprocess package (currently 52.9%)                          | 600      | 5        |         |
| 061 | pkg/asvs         | Test    | Add test for ImportFromCSV method (currently 0% coverage)                          | 400      | 5        |         |
| 062 | pkg/asvs         | Coverage | Increase overall test coverage from 38.2% to at least 60%                          | 600      | 5        |         |
| 063 | pkg/analysis     | Test    | Review and verify test levels are correct                                          | 200      | 5        |         |
| 064 | pkg/notes        | Test    | Review test levels - ensure Level 1 for basic logic, Level 2 for database          | 200      | 5        |         |
| 065 | pkg/ssg          | Test    | Add test level definitions for all tests                                           | 400      | 5        |         |
| 109 | pkg/cve/local    | Code    | Implement connection health checks with automatic reconnection                     | 400      | 5        |         |
| 128 | pkg/cve/remote   | Code    | Design AdaptiveRetry struct with configurable strategies                           | 400      | 5        |         |
| 129 | pkg/cve/remote   | Code    | Implement exponential backoff with jitter generation                               | 400      | 5        |         |
| 130 | pkg/cve/remote   | Code    | Add circuit breaker pattern for fast failure                                       | 400      | 5        |         |
| 131 | pkg/cve/remote   | Code    | Implement priority-based retry for critical requests                               | 400      | 5        |         |
| 132 | pkg/cve/remote   | Code    | Add retry metrics (success rate, backoff time, circuit state)                      | 400      | 5        |         |
| 133 | pkg/cve/remote   | Test    | Test retry logic under various failure scenarios                                   | 400      | 5        |         |
| 144 | website          | Test    | Fix ESLint binary - lint command fails                                             | 15       | 5        |         |
| 145 | website          | Code    | Refactor large components (notes-framework.tsx 731 lines, page.tsx 468 lines)      | 150      | 5        |         |
| 146 | website          | Code    | Consolidate duplicate mock data generation in rpc-client.ts                        | 50       | 5        |         |
| 147 | website          | Code    | Add error boundaries to wrap async components appropriately                        | 50       | 5        |         |
| 148 | website          | Types   | Replace any types with proper TypeScript interfaces                                | 100      | 5        |         |
| 149 | website          | Perf    | Analyze bundle with @next/bundle-analyzer and implement code splitting             | 100      | 5        |         |
| 150 | website          | Perf    | Review and optimize component memoization usage                                    | 75       | 5        |         |
| 151 | website          | Deps    | Update packages to latest stable versions                                          | 25       | 5        |         |
| 152 | website          | Security | Add runtime validation for environment variables using zod                         | 25       | 5        |         |
| 160 | website          | DX      | Add Prettier configuration and integrate with linting                              | 25       | 5        |         |
| 161 | website          | Perf    | Benchmark build times and consider Turbopack or incremental builds                 | 100      | 5        |         |
| 163 | website          | UX      | Improve loading states with sophisticated animations                               | 75       | 5        |         |
| 164 | website          | UX      | Enhance error recovery with retry buttons, context, and suggested actions          | 100      | 5        |         |
| 165 | website          | Debt    | Make RPC timeout configurable per request type                                     | 50       | 5        |         |
| 166 | website          | Debt    | Remove or conditionalize console.log statements in production                      | 25       | 5        |         |
| 167 | website          | Debt    | Extract magic numbers to configuration constants                                   | 25       | 5        |         |

## Maintenance Strategy

### Priority Levels

- **Priority 1 (Critical)**: Core functionality issues, critical bugs, blocking features
- **Priority 2 (Important)**: Performance improvements, security fixes, important refactoring
- **Priority 3 (Nice to Have)**: Code quality improvements, documentation, non-critical optimizations
- **Priority 4 (Low)**: Minor improvements, enhancements, low-impact optimizations
- **Priority 5 (Deferred)**: Tasks that are out of scope or blocked by dependencies

### Task Types

- **Test**: Unit tests, integration tests, test coverage improvements
- **Code**: Feature implementation, bug fixes, refactoring
- **Docs**: Documentation updates, inline comments, architecture docs
- **Perf**: Performance optimization, benchmarking, profiling
- **Security**: Security fixes, input validation, vulnerability remediation
- **Config**: Configuration management, build flags
- **Refactor**: Code restructuring, reducing technical debt
- **Types**: TypeScript type safety improvements
- **Feature**: New feature implementation
- **UX**: User experience improvements
- **A11y**: Accessibility improvements
- **Debt**: Technical debt reduction
- **Coverage**: Test coverage improvements
- **Deps**: Dependency updates

### Task Status Management

**COMPLETED Status**
- Use to mark tasks that have been completed
- Task remains in the table with COMPLETED status for tracking
- Allows maintainers to see what has been done

**WONTFIX Column Purpose**
- **Only for AI guidance**: Mark tasks that AI should NOT implement
- Use cases:
  - Tasks that are out of scope for this project
  - Tasks that require different approach/tools
  - Tasks blocked by architectural decisions
- **NOT for completed tasks**: When a task is done, remove it from the list entirely
- **NOT for deprecated tasks**: When a task becomes obsolete, remove it from the list

**Removing Tasks from TODO**
- **Completed tasks**: Delete the entire row when task is fully completed
- **Deprecated tasks**: Delete the entire row when task is no longer relevant
- **Duplicate tasks**: Consolidate or remove duplicate entries
- Reason: Keep TODO list focused on actionable work

### Markdown Table Syntax Requirements

**CRITICAL: Always maintain proper Markdown table format**

- **Column alignment**: Keep `|-----|` separator line with proper dashes for each column
- **Pipe separators**: Every row must start and end with `|`, including header and separator rows
- **Empty cells**: Use spaces between pipes for empty cells, never omit pipes
- **No trailing spaces**: Remove unnecessary whitespace after row endings
- **Table structure example**:

```markdown
| ID  | Package | Type | Description | Est LoC | Priority | WONTFIX |
|-----|---------|------|-------------|---------|----------|---------|
| 001 | pkg/foo | Code | Fix bug     | 100     | 1        |         |
```

**Common mistakes to avoid**:

❌ Missing pipe at end of row: `| 001 | pkg/foo | Code | Fix bug | 100 | 1`
✅ Correct: `| 001 | pkg/foo | Code | Fix bug | 100 | 1 |`

❌ Incorrect separator line: `|--|--|--|--|--|--|--|` or `|---|---|---|---|---|---|`
✅ Correct: `|-----|---------|------|-------------|---------|----------|---------|`

❌ Inconsistent spacing: `|001|pkg/foo|Code|Fix bug|100|1|`
✅ Correct: `| 001 | pkg/foo | Code | Fix bug | 100 | 1 |`

**When editing the table**:
1. Preserve the exact column structure
2. Count pipes to ensure correct number of columns (7 columns total)
3. Match column widths with header alignment
4. After edits, verify table renders correctly in Markdown preview

### Execution Process

1. **Priority 1 tasks first** - Focus on critical issues that block development
2. **Evaluate task value** - Assess ROI and impact before starting
3. **Downgrade if necessary** - Move to lower priority if not critical
4. **Mark COMPLETED** - Update status when task is done
5. **Remove completed/deprecated tasks** - Delete tasks from TODO list when done (don't use WONTFIX)
6. **Use WONTFIX only for AI guidance** - Mark tasks as WONTFIX to inform AI they should not be implemented
7. **Use build.sh** - All builds and tests must use `./build.sh` wrapper
8. **Commit frequently** - Make incremental commits at logical milestones
9. **Keep markdown format** - Maintain table structure and formatting
10. **Don't break table structure** - When updating TODO list, preserve markdown table formatting (column alignment, pipe separators)

### Key Principles

- **No force push** - Resolve conflicts manually, keep features functional
- **Remote API testing** - Tests must NOT access external APIs (use mocks/fixtures)
- **Fast tests** - Unit tests should run in milliseconds
- **Document RPC APIs** - Update `service.md` for each service when adding RPC handlers
- **Broker-first architecture** - Only broker spawns subprocesses, no direct subprocess-to-subprocess communication
- **Performance focus** - Use lock-free patterns, connection pooling, batch operations
