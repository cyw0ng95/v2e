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
| TODO-017 | cwe | Optimization | Add database indexes for common query patterns | 50 | High | |
| TODO-018 | glc | Feature | Implement automatic schema migration with version tracking | 150 | Medium | |
| TODO-019 | graph | Feature | Add graph metrics (centrality, clustering coefficient) | 250 | Low | |
| TODO-020 | graph | Optimization | Implement graph compression for large-scale deployments | 300 | Low | |
| TODO-021 | graph | Test | Add property-based tests using testing/quick | 150 | Medium | |
| TODO-022 | jsonutil | Refactor | Unify error handling across jsonutil functions | 80 | Low | |
| TODO-023 | jsonutil | Feature | Add JSON schema validation support | 200 | Low | |
| TODO-024 | meta | Feature | Implement provider dependency management | 180 | High | |
| TODO-025 | meta | Refactor | Extract FSM transition logic into strategy pattern | 150 | Medium | |
| TODO-026 | meta | Test | Add chaos testing for provider coordination | 250 | Medium | |
| TODO-027 | notes | Feature | Add memory card export/import functionality | 200 | Low | |
| TODO-029 | notes | Refactor | Simplify bookmark service with repository pattern | 120 | Medium | |
| TODO-030 | notes | Test | Add performance benchmarks for FSM operations | 100 | Medium | |
| TODO-031 | proc | Feature | Implement subprocess health monitoring and auto-restart | 200 | High | |
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
| TODO-078 | cmd/v2broker | Feature | Add graceful shutdown with drain period to allow in-flight requests to complete before terminating | 180 | High | |
| TODO-080 | cmd/v2broker | Optimization | Implement request batching for frequently called RPC methods to reduce context switching overhead | 150 | Medium | |
| TODO-081 | cmd/v2local | Refactor | Simplify database connection pooling by using generic pool wrapper instead of repeated SetMaxIdleConns/SetMaxOpenConns calls | 80 | Medium | |
| TODO-082 | cmd/v2local | Feature | Add database query logging with execution time tracking to identify slow queries | 100 | Low | |
| TODO-083 | cmd/v2local | Bug Fix | Fix potential SQLite database lock contention when multiple services access same database file simultaneously | 120 | High | |
| TODO-084 | cmd/v2meta | Refactor | Extract provider FSM state transition logic into shared package to reduce code duplication across CVEProvider, CWEProvider, CAPECProvider, ATTACKProvider | 250 | Medium | |
| TODO-085 | cmd/v2meta | Feature | Add provider dependency graph to automatically determine provider startup order based on data dependencies (CWE before CAPEC, CAPEC before ATT&CK) | 200 | High | |
| TODO-086 | cmd/v2meta | Bug Fix | Fix potential deadlock in macro FSM when provider fails during bootstrap phase and cleanup routines block on same lock | 150 | High | |
| TODO-087 | cmd/v2meta | Optimization | Add provider health checks with automatic restart for providers in TERMINATED state that should be running | 120 | Medium | |
| TODO-088 | cmd/v2remote | Refactor | Simplify HTTP client configuration by extracting into shared package with retry logic and timeout handling | 100 | Medium | |
| TODO-089 | cmd/v2remote | Feature | Add rate limit detection with adaptive backoff using Retry-After header from HTTP responses | 150 | High | |
| TODO-090 | cmd/v2remote | Bug Fix | Fix potential memory leak when streaming large responses from NVD API without properly closing response body | 80 | High | |
| TODO-091 | cmd/v2remote | Optimization | Implement response caching for frequently accessed API endpoints (e.g., CVE by ID lookup) to reduce API calls | 120 | Medium | |
| TODO-092 | cmd/v2sysmon | Refactor | Extract metric collection logic into reusable functions to reduce code duplication for CPU, memory, and disk monitoring | 80 | Medium | |
| TODO-093 | cmd/v2sysmon | Feature | Add alert threshold configuration with webhook or email notifications when metrics exceed defined limits | 150 | Medium | |
| TODO-095 | cmd/v2sysmon | Optimization | Reduce CPU overhead in metric collection by sampling metrics at configurable intervals instead of every poll cycle | 100 | Medium | |
| TODO-096 | notes | Refactor | Compact LearningFSM code by extracting common save state patterns into helper function | 50 | Medium | |
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

## TODO Management Guidelines

For detailed guidelines on managing TODO items, see the **TODO Management** section in `CLAUDE.md`.

### Quick Reference

**Adding Tasks**: Use next available ID, include test code in LoC estimates, write detailed actionable descriptions

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