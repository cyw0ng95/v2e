# FSM Simplification Requirements

## Introduction

This document specifies requirements for simplifying the MacroFSM and ProviderFSM architecture presentation from backend to frontend. The goal is to reduce cognitive load on users by exposing only essential states while preserving advanced functionality for power users.

## Glossary

- **MacroFSM**: High-level orchestration state machine managing all providers
- **ProviderFSM**: Individual data provider (CVE, CWE, CAPEC, ATT&CK, etc.) state machine
- **Simplified State**: User-friendly state representation (ACTIVE, READY, INACTIVE)
- **Full State**: Complete internal state representation (7 provider states, 4 macro states)
- **Advanced Mode**: Display showing all internal states and controls

## Requirements

### REQ-001: Simplified Provider States

**User Story:** AS A user, I want to see provider status in simple terms so I can quickly understand if data is being processed.

#### Acceptance Criteria

1. WHEN a provider is in RUNNING state, the UI SHALL display "ACTIVE"
2. WHEN a provider is in IDLE or ACQUIRING state, the UI SHALL display "READY"
3. WHEN a provider is in WAITING_QUOTA, WAITING_BACKOFF, PAUSED, or TERMINATED state, the UI SHALL display "INACTIVE"
4. WHEN a user hovers over simplified state, a tooltip SHALL show the full internal state

### REQ-002: Simplified Macro States

**User Story:** AS A user, I want to see the overall ETL status in simple terms.

#### Acceptance Criteria

1. WHEN MacroFSM is in ORCHESTRATING state, the UI SHALL display "RUNNING"
2. WHEN MacroFSM is in BOOTSTRAPPING, STABILIZING, or DRAINING state, the UI SHALL display "TRANSITIONING"

### REQ-003: Advanced Mode Toggle

**User Story:** AS A power user, I want to see all internal states and controls for debugging.

#### Acceptance Criteria

1. WHEN the user toggles "Advanced Mode", the UI SHALL display full provider states (7 variants)
2. WHEN the user toggles "Advanced Mode", the UI SHALL display full macro states (4 variants)
3. WHEN the user toggles "Advanced Mode", the UI SHALL show Pause/Resume controls
4. The default view SHALL be simplified mode

### REQ-004: Backward Compatibility

**User Story:** AS A system integrator, I want existing API calls to continue working.

#### Acceptance Criteria

1. WHEN an RPC call is made without the simplified flag, the response SHALL include full state information
2. WHEN frontend makes existing API calls, the behavior SHALL remain unchanged

### REQ-005: Simplified RPC Endpoint

**User Story:** AS A frontend developer, I want a dedicated simplified endpoint to reduce client-side processing.

#### Acceptance Criteria

1. WHEN `RPCFSMGetTopologySimplified` is called, the response SHALL contain pre-computed simplified states
2. The simplified endpoint SHALL NOT include internal states like ACQUIRING or WAITING_QUOTA

### REQ-006: Control Simplification

**User Story:** AS A basic user, I want only Start and Stop controls in the simplified view.

#### Acceptance Criteria

1. WHEN in simplified mode, the UI SHALL hide the Pause button
2. WHEN in simplified mode, the UI SHALL hide the Resume button
3. WHEN in advanced mode, the UI SHALL show all controls (Start, Pause, Resume, Stop)

### REQ-007: Unified FSM Control RPC

**User Story:** AS A system architect, I want to reduce the number of RPC endpoints for FSM management.

#### Acceptance Criteria

1. WHEN a client sends `{ action: "start", provider_id: "cve" }`, the provider SHALL start
2. WHEN a client sends `{ action: "stop" }` without provider_id, ALL providers SHALL stop
3. WHEN a client sends invalid action, the response SHALL contain an error message
4. The unified endpoint SHALL replace 8 individual RPC handlers

### REQ-008: Reduced RPC Surface

**User Story:** AS A frontend developer, I want fewer endpoints to manage.

#### Acceptance Criteria

1. WHEN managing FSM, there SHALL be only 3 RPC endpoints (control, topology, checkpoints)
2. WHEN comparing with previous version, the number of FSM-related handlers SHALL decrease from 11 to 3
3. The backward-compatible endpoints MAY be deprecated but not removed immediately

### REQ-009: State Conversion Tests

**User Story:** AS A developer, I want tests to verify state mapping correctness.

#### Acceptance Criteria

1. WHEN testing ToSimplifiedProviderState(), all 7 input states SHALL have correct mapped outputs
2. WHEN testing ToSimplifiedMacroState(), all 4 input states SHALL have correct mapped outputs
3. WHEN unknown state is passed, the conversion SHALL return a sensible default
4. The test coverage SHALL exceed 80% for the types package

### REQ-010: RPC Handler Tests

**User Story:** AS A QA engineer, I want comprehensive tests for the unified RPC endpoints.

#### Acceptance Criteria

1. WHEN calling RPCFSMControl with valid action and provider_id, it SHALL return success
2. WHEN calling RPCFSMControl with action="stop" without provider_id, it SHALL affect all providers
3. WHEN calling RPCFSMControl with invalid action, it SHALL return error response
4. WHEN calling RPCFSMControl with non-existent provider, it SHALL return error in Failed map
5. WHEN calling RPCFSMGetTopology with simplified=true, it SHALL return simplified states

---

## Non-Functional Requirements

- NFR-001: State translation MUST be stateless (no database changes)
- NFR-002: Translation layer MUST have <1ms overhead
- NFR-003: Default view MUST show simplified states (backward compatible)
