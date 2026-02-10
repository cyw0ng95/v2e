# v2e Go Packages Maintain TODO List

| ID | Package | Type | Description | Estimate LoC | Priority | Mark WONTFIX? |
|----|---------|------|-------------|--------------|----------|---------------|
| TODO-001 | analysis | Refactor | Add comprehensive error handling and recovery mechanisms for FSM state transitions | 150 | High | |
| TODO-002 | analysis | Feature | Implement graph persistence with incremental checkpointing | 200 | High | |
| TODO-003 | analysis | Test | Add integration tests for graph analysis workflows | 300 | Medium | |
EOF'| TODO-119 | all | Feature | Add OpenTelemetry instrumentation to pkg/graph for distributed tracing of graph operations | 250 | Low | |
| TODO-120 | all | Feature | Add OpenTelemetry instrumentation to pkg/capec for distributed tracing of import operations | 250 | Low | |
| TODO-121 | all | Feature | Add OpenTelemetry instrumentation to pkg/ssg for distributed tracing of import operations | 250 | Low | |
| TODO-122 | all | Feature | Add OpenTelemetry instrumentation to pkg/asvs for distributed tracing of import operations | 250 | Low | |
| TODO-123 | all | Feature | Add OpenTelemetry instrumentation to pkg/attack for distributed tracing of import operations | 250 | Low | |
| TODO-124 | all | Feature | Add OpenTelemetry instrumentation to pkg/cce for distributed tracing of import operations | 250 | Low | |
| TODO-125 | all | Feature | Add OpenTelemetry instrumentation to pkg/cwe for distributed tracing of import operations | 250 | Low | |
| TODO-126 | all | Feature | Add OpenTelemetry instrumentation to pkg/glc for distributed tracing of database operations | 250 | Low | |
| TODO-127 | all | Feature | Add OpenTelemetry instrumentation to pkg/urn for distributed tracing of URN parsing and validation | 250 | Low | |
| TODO-128 | all | Feature | Add OpenTelemetry instrumentation to pkg/jsonutil for distributed tracing of JSON marshaling/unmarshaling | 250 | Low | |

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
| TODO-119 | all | Feature | Add OpenTelemetry instrumentation to pkg/graph for distributed tracing of graph operations | 250 | Low | |
| TODO-120 | all | Feature | Add OpenTelemetry instrumentation to pkg/capec for distributed tracing of import operations | 250 | Low | |
| TODO-121 | all | Feature | Add OpenTelemetry instrumentation to pkg/ssg for distributed tracing of import operations | 250 | Low | |
| TODO-122 | all | Feature | Add OpenTelemetry instrumentation to pkg/asvs for distributed tracing of import operations | 250 | Low | |
| TODO-123 | all | Feature | Add OpenTelemetry instrumentation to pkg/attack for distributed tracing of import operations | 250 | Low | |
| TODO-124 | all | Feature | Add OpenTelemetry instrumentation to pkg/cce for distributed tracing of import operations | 250 | Low | |
| TODO-125 | all | Feature | Add OpenTelemetry instrumentation to pkg/cwe for distributed tracing of import operations | 250 | Low | |
| TODO-126 | all | Feature | Add OpenTelemetry instrumentation to pkg/glc for distributed tracing of database operations | 250 | Low | |
| TODO-127 | all | Feature | Add OpenTelemetry instrumentation to pkg/urn for distributed tracing of URN parsing and validation | 250 | Low | |
| TODO-128 | all | Feature | Add OpenTelemetry instrumentation to pkg/jsonutil for distributed tracing of JSON marshaling/unmarshaling | 250 | Low | |
