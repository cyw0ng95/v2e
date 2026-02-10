# v2e Go Packages Maintain TODO List

| ID | Package | Type | Description | Estimate LoC | Priority | Mark WONTFIX? |
|----|---------|------|-------------|--------------|----------|---------------|
| TODO-001 | analysis | Refactor | Add comprehensive error handling and recovery mechanisms for FSM state transitions | 150 | High | |
| TODO-002 | analysis | Feature | Implement graph persistence with incremental checkpointing | 200 | High | |
| TODO-003 | analysis | Test | Add integration tests for graph analysis workflows | 300 | Medium | |
| TODO-004 | asvs | Optimization | Optimize CSV import with streaming and parallel processing | 100 | Medium | |
| TODO-005 | asvs | Feature | Add incremental update support for CSV imports | 80 | Medium | |
| TODO-006 | attack | Refactor | Simplify Excel parsing logic by extracting helper functions | 120 | Medium | |
| TODO-007 | attack | Bug Fix | Fix potential panic when XLSX sheet contains unexpected data types | 50 | High | |
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
| TODO-028 | notes | Bug Fix | Fix potential deadlock in LearningFSM state transitions | 60 | High | |
| TODO-029 | notes | Refactor | Simplify bookmark service with repository pattern | 120 | Medium | |
| TODO-030 | notes | Test | Add performance benchmarks for FSM operations | 100 | Medium | |
| TODO-031 | proc | Feature | Implement subprocess health monitoring and auto-restart | 200 | High | |
| TODO-032 | proc | Optimization | Add message prioritization and backpressure handling | 150 | Medium | |
| TODO-033 | proc | Test | Add fuzz testing for message serialization/deserialization | 150 | Medium | |
| TODO-034 | rpc | Bug Fix | Fix potential memory leak in pending request map | 40 | High | |
| TODO-035 | rpc | Feature | Add request retry with exponential backoff | 120 | Medium | |
| TODO-036 | rpc | Refactor | Implement connection pooling for RPC clients | 100 | Medium | |
| TODO-037 | ssg | Feature | Add incremental SSG data update support with field-level diffing to avoid re-importing entire datasets when only subset changes | 250 | High | |
| TODO-038 | ssg | Optimization | Implement parallel parsing with worker pools for large SSG XML files to reduce import time by 40% | 180 | Medium | |
| TODO-039 | testutils | Feature | Add mock HTTP server for testing remote providers with configurable responses and delay simulation | 150 | Medium | |
| TODO-040 | testutils | Refactor | Extract common test patterns into helpers (database fixtures, assertion helpers, context factories) | 100 | Low | |
| TODO-041 | urn | Feature | Add URN validation with comprehensive rules to catch malformed URNs before database operations | 80 | Medium | |
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

**Marking WONTFIX**:
- Only project maintainers can mark tasks as obsolete
- AI agents MUST NOT add "WONTFIX" to any task

**Priority**: High (critical), Medium (important), Low (nice-to-have)

**LoC Estimates**: Small (<100), Medium (100-300), Large (300+)
