# v2e Go Packages TODO List

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
| TODO-037 | ssg | Feature | Add incremental SSG data update support | 250 | High | |
| TODO-038 | ssg | Optimization | Implement parallel parsing for large SSG files | 180 | Medium | |
| TODO-039 | testutils | Feature | Add mock HTTP server for testing remote providers | 150 | Medium | |
| TODO-040 | testutils | Refactor | Extract common test patterns into helpers | 100 | Low | |
| TODO-041 | urn | Feature | Add URN validation with comprehensive rules | 80 | Medium | |
| TODO-042 | urn | Test | Add property-based tests for URN parsing | 120 | Medium | |
| TODO-043 | analysis | Documentation | Add examples for graph analysis API usage | 80 | Low | |
| TODO-044 | meta | Documentation | Document provider FSM lifecycle and transitions | 100 | Low | |
| TODO-045 | notes | Documentation | Add learning strategy comparison guide | 60 | Low | |
| TODO-046 | proc | Documentation | Document message flow and serialization format | 100 | Low | |
| TODO-047 | all | Refactor | Standardize error codes across all packages | 400 | Medium | |
| TODO-048 | all | Feature | Add distributed tracing support | 500 | Low | |
| TODO-049 | all | Test | Add integration test suite for broker-subprocess communication | 300 | High | |
| TODO-050 | all | Optimization | Profile and optimize hot paths across all packages | 200 | Medium | |

## TODO Management Guidelines

### Priority Levels
- **High**: Critical bugs, security issues, or blocking features
- **Medium**: Important improvements, performance optimizations
- **Low**: Nice-to-have features, code quality improvements

### When to Mark as WONTFIX
- Feature is no longer relevant to project goals
- Issue is superseded by a better approach
- Cost of implementation exceeds benefit
- External dependency handles the requirement

### Removing Completed Tasks
When a TODO is complete:
1. Verify all acceptance criteria are met
2. Run relevant tests (`./build.sh -t`)
3. Remove the corresponding row from this table
4. Keep the markdown formatting consistent

### Estimating LoC
- **Small**: < 100 lines of code
- **Medium**: 100-300 lines of code
- **Large**: 300+ lines of code
- Include test code in estimates (aim for >80% test coverage)
