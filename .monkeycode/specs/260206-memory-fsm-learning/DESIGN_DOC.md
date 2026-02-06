# MemoryFSM and LearningFSM Design Documentation

## Overview

The v2e Memory FSM Learning System implements two key state machines:

1. **MemoryFSM**: Unified state management for learning objects (notes and memory cards)
2. **LearningFSM**: User learning progress tracking and viewing context management

## MemoryFSM Design

### Purpose

MemoryFSM provides a unified state machine for managing the lifecycle of all learning objects (notes and memory cards). It ensures consistent state transitions, tracks state history, and persists state changes.

### State Definitions

The MemoryFSM defines 7 valid states:

```
draft      - Initial editing state for notes
new        - Initial state for memory cards
learning   - Active learning in progress
reviewed   - Post-first-review state for cards
learned    - Note marked as complete
mastered   - Card is fully learned
archived   - Item is archived
```

### State Transitions

#### Valid Transitions

| From State   | To State(s)                              | Description                                     |
|-------------|-------------------------------------------|------------------------------------------------|
| draft       | learned, archived                         | Note completion or archival                      |
| new         | learning, archived                         | Start learning or archival                         |
| learning    | reviewed, mastered, archived                | First review completion, mastery, or archival    |
| reviewed    | learning, mastered, archived                | More practice, mastery, or archival               |
| learned     | learned, archived                         | Editable while maintaining learned status          |
| mastered    | archived                                  | Final archival                                 |
| archived    | -                                        | Terminal state                                 |

#### State Transition Validation

All transitions are validated against `validMemoryTransitions` map:

```go
var validMemoryTransitions = map[MemoryState]map[MemoryState]bool{
    MemoryStateDraft: {
        MemoryStateLearned:  true,
        MemoryStateArchived: true,
    },
    MemoryStateNew: {
        MemoryStateLearning: true,
        MemoryStateArchived: true,
    },
    MemoryStateLearning: {
        MemoryStateReviewed: true,
        MemoryStateMastered: true,
        MemoryStateArchived: true,
    },
    MemoryStateReviewed: {
        MemoryStateLearning: true,
        MemoryStateMastered: true,
        MemoryStateArchived: true,
    },
    MemoryStateLearned: {
        MemoryStateLearned:  true,
        MemoryStateArchived: true,
    },
    MemoryStateMastered: {
        MemoryStateArchived: true,
    },
    MemoryStateArchived: {},
}
```

### State History

Each state transition is recorded with:

```go
type StateHistory struct {
    FromState MemoryState `json:"from_state"`
    ToState   MemoryState `json:"to_state"`
    Timestamp time.Time   `json:"timestamp"`
    Reason    string      `json:"reason"`
    UserID    string      `json:"user_id,omitempty"`
}
```

**Storage**: State history is stored as JSON in the `fsm_state_history` column.

### Persistence

MemoryFSM state is persisted to BoltDB:

**Storage Location**: `pkg/notes/fsm/storage.go`

**Key Methods**:
- `SaveMemoryFSMState(urn string, state *MemoryFSMState) error`
- `LoadMemoryFSMState(urn string) (*MemoryFSMState, error)`
- `ValidateMemoryFSMState(urn string) error`

**State Structure**:
```go
type MemoryFSMState struct {
    URN          string         `json:"urn"`
    State        MemoryState    `json:"state"`
    StateHistory []StateHistory `json:"state_history"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
}
```

### MemoryObject Interface

All learning objects implement the MemoryObject interface:

```go
type MemoryObject interface {
    GetURN() string
    GetMemoryFSMState() MemoryState
    SetMemoryFSMState(state MemoryState) error
}
```

**Implementations**:
- `NoteModel` - Notes with draft/learned states
- `MemoryCardModel` - Cards with new/learning/reviewed/mastered states

### BaseMemoryFSM Implementation

The `BaseMemoryFSM` provides:

1. **Thread-Safe State Management**
   - Uses `sync.RWMutex` for concurrent access
   - Prevents race conditions

2. **State Transition Validation**
   - Validates all transitions before execution
   - Rolls back on error

3. **State Persistence**
   - Automatically persists state changes
   - Loads state on initialization

4. **State History Tracking**
   - Records all transitions with timestamps
   - Supports audit trail

### Concurrency Model

- **Read Operations**: Use RLock (multiple readers allowed)
- **Write Operations**: Use Lock (exclusive access)
- **State Consistency**: Ensure no concurrent state mutations

**Important**: The `Transition` method captures state for persistence without re-acquiring locks to avoid deadlock.

## LearningFSM Design

### Purpose

LearningFSM tracks user's overall learning session state, manages learning strategies (BFS/DFS), and presents next items for learning.

### State Definitions

The LearningFSM defines 5 valid states:

```
idle       - Initial state
browsing   - BFS mode (list order)
deep_dive  - DFS mode (following links)
reviewing  - Card review mode
paused     - Session paused
```

### Learning Strategies

#### BFS (Breadth-First Search)

**Purpose**: Present items in list order

**Behavior**:
- Returns next unviewed item from available list
- Maintains `viewedItems` list
- Sets state to `browsing`

**Use Case**: Initial exploration of all available items

#### DFS (Depth-First Search)

**Purpose**: Present items through link relationships

**Behavior**:
- Uses `pathStack` for navigation
- Follows links between related items
- Switches back to BFS when path exhausted
- Sets state to `deep_dive`

**Use Case**: Deep dive into related items

### Item Graph

The `ItemGraph` maintains bidirectional relationships:

```go
type ItemGraph struct {
    links map[string][]string // URN -> linked URNs
    mu    sync.RWMutex
}
```

**Construction**:
- Inter-type links: CVE → CWE → CAPEC → ATT&CK
- Intra-type links: Items of same type linked in sequence

### Learning Context

`LearningContext` provides current session information:

```go
type LearningContext struct {
    ViewedItems    []string       `json:"viewed_items"`
    CompletedItems []string       `json:"completed_items"`
    AvailableItems []SecurityItem `json:"available_items"`
    PathStack      []string       `json:"path_stack"`
}
```

### Persistence

LearningFSM state is persisted to BoltDB:

**Storage Location**: Same BoltDB as MemoryFSM

**Key Methods**:
- `SaveLearningFSMState(state *LearningFSMState) error`
- `LoadLearningFSMState() (*LearningFSMState, error)`
- `ClearLearningFSMState() error`

**State Structure**:
```go
type LearningFSMState struct {
    State           LearningState `json:"state"`
    CurrentStrategy string        `json:"current_strategy"`
    CurrentItemURN  string        `json:"current_item_urn"`
    ViewedItems     []string      `json:"viewed_items"`
    CompletedItems  []string      `json:"completed_items"`
    PathStack       []string      `json:"path_stack"`
    SessionStart    time.Time     `json:"session_start"`
    LastActivity    time.Time     `json:"last_activity"`
    UpdatedAt       time.Time     `json:"updated_at"`
}
```

### Concurrency Model

- **Thread-Safe**: Uses `sync.RWMutex` for all state changes
- **Activity Tracking**: Updates `lastActivity` on every operation
- **State Persistence**: Automatic persistence on state changes

## Integration Between FSMs

### Interaction Flow

```
User Action → LearningFSM → Selects Learning Item
                ↓
            User Interacts with Item
                ↓
            User Creates Note/Memory Card
                ↓
            MemoryFSM Manages Card State
                ↓
            User Marks as Learned/Reviews
                ↓
            MemoryFSM Transitions State
                ↓
            LearningFSM Updates Progress
```

### Shared URN Namespace

Both FSMs use the same URN format:

- **Notes**: `v2e::note::<id>`
- **Memory Cards**: `v2e::card::<id>`
- **Bookmarks**: `v2e::<provider>::<type>::<id>`

This enables:
- Unified lookups across all object types
- Consistent linking between objects
- Efficient reverse queries via URNIndex

## Performance Optimizations

### MemoryFSM Optimizations

1. **Lock-Free State Reads**: Use RLock for concurrent reads
2. **Batch State Updates**: Group related state changes
3. **Efficient History Storage**: JSON serialization for compact storage

### LearningFSM Optimizations

1. **Pre-built Item Graph**: Construct once at initialization
2. **Efficient BFS Lookup**: Use map for viewed items
3. **Path Stack Reuse**: Reuse slice for DFS navigation

### BoltDB Optimizations

1. **Bucket Separation**: Separate buckets for each FSM type
2. **Read Transactions**: Use View for read-only operations
3. **Write Batching**: Batch writes when possible

## Error Handling

### MemoryFSM Errors

- **Invalid Transition**: Returned when transition is not allowed
- **Storage Error**: Returned when BoltDB operation fails
- **Object Update Error**: Returned when object state update fails

### LearningFSM Errors

- **No More Items**: Returned when no items available for learning
- **Storage Error**: Returned when BoltDB operation fails
- **Invalid Strategy**: Returned when strategy is not recognized

## Testing Strategy

### Unit Tests

1. **MemoryFSM State Transitions**
   - Test all valid transitions
   - Test invalid transitions
   - Test state history tracking

2. **LearningFSM Behavior**
   - Test BFS item loading
   - Test DFS navigation
   - Test state persistence

3. **Storage Validation**
   - Test valid state loading
   - Test invalid state detection
   - Test state integrity checks

### Integration Tests

1. **FSM Interaction**
   - Test MemoryFSM and LearningFSM working together
   - Test URN-based lookups
   - Test state synchronization

2. **End-to-End Workflows**
   - Test complete learning session
   - Test bookmark → card → review flow
   - Test note creation → learning completion flow

## Security Considerations

1. **State Consistency**: Prevent unauthorized state transitions
2. **Data Integrity**: Validate state history integrity
3. **Access Control**: Ensure only authorized state changes
4. **Audit Trail**: Maintain complete state transition history

## Future Enhancements

1. **Multi-User Support**: Extend FSMs to track multiple users
2. **Machine Learning**: Use ML to optimize learning strategies
3. **Adaptive Scheduling**: Adjust review intervals based on performance
4. **Advanced Analytics**: Provide learning insights and recommendations

## Conclusion

The MemoryFSM and LearningFSM provide a robust, unified framework for managing learning objects and tracking user progress. Their combined state management ensures consistency across the system while enabling a passive, intuitive learning experience.
