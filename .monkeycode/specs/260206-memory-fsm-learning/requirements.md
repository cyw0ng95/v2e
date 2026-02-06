# Requirements Document

## Introduction

This requirements document describes next iteration of v2e learning memory system. The system supports a single user learning process with persistent learning state management. The learning experience is designed to be passive and intuitive: users simply view and read objects fetched from UEE (Unified ETL Engine), make marks or notes, create memory cards, and review them.

The system automatically manages learning strategies (BFS/DFS) in background without user awareness or selection. Users only interact with learning objects through viewing, reading, marking, note-taking, memory card creation, and spaced repetition review.

**Final Learning Goal:** To master and memorize all security objects (CVE, CWE, CAPEC, ATT&CK) fetched from UEE data source.

The system reuses existing services (no new services) and manages all learning objects (notes, memory cards, bookmarks) through a unified MemoryFSM. Notes can be marked as learned, while memory cards support more complex state transitions for spaced repetition.

## Glossary

- **v2e System** (v2e): Vulnerabilities Viewer Engine, a broker-first microservices system for managing CVE, CWE, CAPEC, and ATT&CK security data
- **Note** (Note): A rich text object created by users for recording and reviewing security knowledge
- **Memory Card** (MemoryCard): A card object for spaced repetition learning containing front content (question), back content (answer), and rich text content
- **Bookmark** (Bookmark): A user-marked security item (CVE, CWE, CAPEC, ATT&CK) used for learning and review
- **URN** (URN): Uniform Resource Name, a unique identifier for objects in v2e system, formatted as `v2e::<provider>::<type>::<atomic_id>`
- **MemoryFSM** (MemoryFSM): Memory Finite State Machine, unified lifecycle state management for notes and memory cards
- **LearningFSM** (LearningFSM): Learning Finite State Machine, tracks user learning progress and state
- **UEE** (UEE): Unified ETL Engine, for fetching and importing security data from external sources
- **BFS** (BFS): Breadth-First Search, internal learning strategy for presenting items in list order (transparent to users)
- **DFS** (DFS): Depth-First Search, internal learning strategy for presenting items through analysis result links (transparent to users)
- **Passive Learning Experience** (Passive Learning): Users focus on viewing, reading, marking, note-taking, card creation, and review without managing learning strategies
- **Spaced Repetition** (Spaced Repetition): Learning technique that reviews material at increasing intervals to improve long-term retention

## Requirements

### Requirement 1: Passive Learning Experience Management

**User Story:** As a system user, I want to focus on viewing, reading, marking, note-taking, and memory card review without managing learning strategies.

#### Acceptance Criteria

1. WHEN the user opens the learning interface, the system SHALL display security objects (CVE, CWE, CAPEC, ATT&CK) from UEE data source
2. WHILE the user views a security object, the system SHALL allow the user to add marks, create notes, or create memory cards
3. WHEN the user adds a mark to an object, the system SHALL record the mark and automatically create a memory card for that object
4. WHILE the user navigates between objects, the system SHALL maintain the user's viewing context without requiring strategy selection
5. WHEN the user reviews memory cards, the system SHALL present cards due for review based on the spaced repetition algorithm

#### Design Considerations

- Learning strategies (BFS/DFS) are internal implementation details transparent to users
- System automatically determines object presentation order based on user's viewing history and linked relationships
- User interface emphasizes viewing and interaction over strategy management
- Learning progress is tracked implicitly through user actions (marks, notes, card reviews)

### Requirement 2: User-Created Memory Cards

**User Story:** As a learner, I want to create memory cards to record and organize knowledge points I need to learn.

#### Acceptance Criteria

1. WHEN the user submits a memory card creation request, the system SHALL accept and store the memory card containing front content, back content, and rich text content
2. WHEN the system creates a memory card, the system SHALL generate a unique URN as the identifier for that card
3. WHEN the system creates a memory card, the system SHALL initialize the card's MemoryFSM state to "new"
4. WHILE the user edits the memory card's rich text content, the system SHALL support TipTap JSON format serialization and deserialization
5. WHEN the user updates a memory card, the system SHALL record the update operation to the card history

### Requirement 3: Memory Card URN Link Management

**User Story:** As a learner, I want to link memory cards to security item URNs to establish connections between knowledge points.

#### Acceptance Criteria

1. WHEN the user adds a memory card to URN link, the system SHALL record the linked URN list in the memory card object
2. WHEN the user removes a memory card to URN link, the system SHALL delete the specified URN from the memory card's linked list
3. WHILE the user is viewing a memory card, the system SHALL display all URNs linked to that card and their corresponding security item information
4. WHEN the user searches for memory cards by URN, the system SHALL return all memory cards associated with that URN
5. WHILE the system manages memory card links, the system SHALL maintain bidirectional reference consistency of the relationships

### Requirement 4: Bookmark Creation and Automatic Memory Card Generation

**User Story:** As a learner, I want to bookmark security items and automatically generate corresponding memory cards to quickly start learning.

#### Acceptance Criteria

1. WHEN the user performs a bookmark operation on any security item (identified by URN), the system SHALL create a bookmark record
2. WHEN the system creates a bookmark, the system SHALL automatically generate an initial memory card for that bookmark
3. WHEN the system generates an initial memory card, the system SHALL use the bookmark title as the card front content and the bookmark description as the card back content
4. WHEN the system generates an initial memory card, the system SHALL set the card's MemoryFSM state to "new"
5. WHEN the system generates an initial memory card, the system SHALL automatically link the bookmark's URN to the memory card

### Requirement 5: Note Creation and URN Link Management

**User Story:** As a learner, I want to create notes and link them to URNs to record my understanding of security items.

#### Acceptance Criteria

1. WHEN the user creates a note, the system SHALL assign a unique URN as the identifier for that note
2. WHEN the user creates a note, the system SHALL initialize the note's MemoryFSM state to "draft"
3. WHEN the user adds a note to URN link, the system SHALL record the linked URN list in the note object
4. WHEN the user removes a note to URN link, the system SHALL delete the specified URN from the note's linked list
5. WHILE the user edits note content, the system SHALL continuously auto-save the note to persistent storage

### Requirement 6: MemoryFSM Unified State Management

**User Story:** As the system, I need unified lifecycle state management for notes and memory cards to provide a consistent learning experience.

#### Acceptance Criteria

1. WHEN the system creates a note, the MemoryFSM SHALL initialize the note state to "draft"
2. WHEN the system creates a memory card, the MemoryFSM SHALL initialize the card state to "new"
3. WHEN the user marks a note as learned, the MemoryFSM SHALL transition the note state from "draft" to "learned"
4. WHEN the user marks a memory card as learned, the MemoryFSM SHALL transition the card state from "new" to "mastered"
5. WHILE the MemoryFSM manages object states, the system SHALL record a history of each state transition with timestamps

### Requirement 7: Note Learning Completion State

**User Story:** As a learner, I want to mark notes as learned to track my learning progress.

#### Acceptance Criteria

1. WHEN the user completes note learning and submits a completion operation, the system SHALL validate the note content completeness
2. WHILE the system validates note content completeness, the system SHALL confirm the note contains valid content and at least one URN link
3. WHEN the note validation passes, the system SHALL transition the note state from "draft" to "learned" via MemoryFSM
4. WHEN the note state transitions to "learned", the system SHALL record the learning completion time and user identifier
5. WHEN the note state is "learned", the system SHALL allow the user to continue editing the note while maintaining the learned status

### Requirement 8: Memory Card Complex State Management

**User Story:** As a learner, I want memory cards to support more complex state management to reflect learning progress at different stages.

#### Acceptance Criteria

1. WHEN the user starts learning a memory card, the MemoryFSM SHALL transition the card state from "new" to "learning"
2. WHEN the user completes the first review of a memory card, the MemoryFSM SHALL transition the card state from "learning" to "reviewed"
3. WHEN the user successfully reviews a memory card multiple times, the MemoryFSM SHALL transition the card state from "reviewed" to "mastered"
4. WHILE the memory card state is "learning", the system SHALL track review count and success rate
5. WHEN the memory card state is "mastered", the system SHALL calculate the next review time based on the spaced repetition algorithm

### Requirement 9: Use Existing Services for Implementation

**User Story:** As a system architect, I want to use existing services to implement new features to maintain system architecture consistency.

#### Acceptance Criteria

1. WHEN implementing the learning memory system, the system SHALL use existing services: cmd/access, cmd/local, cmd/meta, cmd/remote, cmd/sysmon
2. WHEN implementing memory card and note management, the system SHALL extend the existing pkg/notes service package
3. WHEN adding new RPC handlers, the system SHALL register the handlers in existing service cmd/*/service.md documentation
4. WHEN storing learning state, the system SHALL use existing database storage mechanisms (BoltDB or SQLite)
5. WHEN handling inter-process communication, the system SHALL use existing stdin/stdout RPC message mechanisms routed through the Broker

### Requirement 10: Internal Learning Strategy for Object Presentation

**User Story:** As a system, I need internal learning strategies to determine object presentation order without user intervention.

#### Acceptance Criteria

1. WHEN user views security objects, system SHALL automatically determine presentation order using internal learning strategy
2. WHILE user navigates through objects, system SHALL use BFS strategy for presenting items in list order by default
3. WHILE user follows links between related objects, system SHALL use DFS strategy to present linked items in depth-first order
4. WHEN user marks an object as learned, system SHALL update internal learning path tracking
5. WHILE system presents objects, system SHALL seamlessly switch between BFS and DFS strategies based on user navigation patterns

#### Design Considerations

- Learning strategies (BFS/DFS) are completely transparent to users
- System automatically selects strategy based on user's navigation context
- BFS strategy presents objects in UEE list order
- DFS strategy follows link relationships to present related objects

### Requirement 12: Learning State Persistence and Recovery

**User Story:** As the system, I need to correctly handle rich text content for memory cards and notes to support the TipTap editor format.

#### Acceptance Criteria

1. WHEN the system receives TipTap JSON formatted memory card or note content, the system SHALL validate the JSON structure validity
2. WHEN the system validates TipTap JSON structure, the system SHALL confirm the presence of required document node types and content nodes
3. WHEN the system stores TipTap JSON content, the system SHALL serialize the content to a string and store it in the database content field
4. WHEN the system retrieves TipTap JSON content from the database, the system SHALL deserialize the content back to the original TipTap JSON object
5. WHEN the system performs TipTap JSON round-trip conversion, the system SHALL ensure the converted content matches the original content exactly

### Requirement 13: Learning State Persistence and Recovery

**User Story:** As a learner, I want learning state to be persistently saved and restored on the next startup to avoid losing learning progress.

#### Acceptance Criteria

1. WHEN the LearningFSM state changes, the system SHALL immediately persist the new state to storage
2. WHEN the system persists learning state, the system SHALL include current learning item index, learning strategy type, and completed item list
3. WHEN the user restarts the system, the system SHALL load the last learning state from persistent storage
4. WHEN the system restores learning state, the system SHALL validate the integrity of the stored learning state data
5. WHILE the system is running, the system SHALL periodically back up learning state to prevent data loss

### Requirement 14: Bidirectional URN Link Relationship Maintenance

**User Story:** As the system, I need to maintain bidirectional relationships between notes, memory cards, and URNs to support multi-angle queries and navigation.

#### Acceptance Criteria

1. WHEN a memory card adds a URN link, the system SHALL record that URN in the memory card object
2. WHEN a URN is added to a memory card, the system SHALL update the URN's associated memory card list in the global URN index
3. WHEN a note adds a URN link, the system SHALL record that URN in the note object
4. WHEN a URN is linked to a note, the system SHALL update the URN's associated note list in the global URN index
5. WHILE the system maintains URN relationships, the system SHALL ensure relationship updates from either side synchronize to the other side

### Requirement 15: Refactor Existing Notes and Bookmarks Implementation

**User Story:** As a system architect, I want to refactor the existing notes, bookmarks, and memory cards implementation to achieve a consistent user experience.

#### Acceptance Criteria

1. WHEN the system refactors the existing notes implementation, the system SHALL ensure all notes use the MemoryFSM for state management
2. WHEN the system refactors the existing bookmarks implementation, the system SHALL ensure bookmark creation automatically generates memory cards with proper URN linking
3. WHEN the system refactors the existing memory cards implementation, the system SHALL ensure all memory cards follow the unified MemoryFSM state transitions
4. WHILE the system performs refactoring, the system SHALL maintain backward compatibility with existing data in the database
5. WHEN the system completes refactoring, the system SHALL provide a consistent user experience across notes, bookmarks, and memory cards

### Requirement 16: Unified User Experience for Learning Objects

**User Story:** As a user, I want a consistent interface and interaction pattern when working with notes, memory cards, and bookmarks.

#### Acceptance Criteria

1. WHEN the user views a note, memory card, or bookmark, the system SHALL display the object using a consistent UI layout
2. WHEN the user edits a note or memory card, the system SHALL provide the same TipTap editor interface for rich text editing
3. WHEN the user links a note or memory card to a URN, the system SHALL use the same URN selection and linking interface
4. WHILE the user navigates between learning objects, the system SHALL maintain consistent navigation patterns and breadcrumbs
5. WHEN the user performs learning operations, the system SHALL provide consistent feedback and status indicators across all object types
#### Design Considerations

- Use same UI component library (shadcn/ui) across all learning object views
- Implement unified state display badges (draft, new, learning, reviewed, learned, mastered)
- Maintain consistent navigation patterns (back button, breadcrumbs, related objects panel)
- Use identical TipTap editor component for notes and memory cards
- Provide consistent action buttons (mark as learned, create card, edit, delete)

---

## Implementation Task List

### Phase 1: Core MemoryFSM and LearningFSM Implementation
- [ ] Design and implement MemoryFSM state machine with transitions for notes and memory cards
- [ ] Design and implement LearningFSM for tracking user learning progress and viewing context
- [ ] Implement persistent storage for LearningFSM state
- [ ] Implement state history tracking with timestamps

### Phase 2: Notes and Memory Cards with URN Links
- [ ] Update NoteModel to include URN links list and unique URN identifier
- [ ] Update MemoryCardModel to include URN links list and unique URN identifier
- [ ] Implement URN link management services (add, remove, query)
- [ ] Implement bidirectional URN index for efficient queries
- [ ] Update TipTap JSON serialization/deserialization for both notes and memory cards

### Phase 3: Bookmark Auto-Generation with Memory Cards
- [ ] Update BookmarkService to automatically generate memory cards on bookmark creation
- [ ] Configure auto-generated card with bookmark title/description as front/back content
- [ ] Link auto-generated card to bookmark URN
- [ ] Set initial MemoryFSM state to "new" for auto-generated cards

### Phase 4: Passive Learning Experience
- [ ] Implement object viewing interface for CVE, CWE, CAPEC, ATT&CK from UEE
- [ ] Add marking functionality (mark as learned/in-progress)
- [ ] Implement note-taking in viewing context
- [ ] Implement memory card creation in viewing context
- [ ] Hide learning strategy selection from user interface

### Phase 5: Internal Learning Strategies (Transparent to Users)
- [ ] Implement BFS strategy for presenting objects in list order
- [ ] Implement DFS strategy for presenting objects through link relationships
- [ ] Implement automatic strategy switching based on user navigation
- [ ] Maintain learning path tracking for both strategies
- [ ] Present next object based on active strategy

### Phase 6: Spaced Repetition Review System
- [ ] Implement memory card review queue based on due date
- [ ] Implement spaced repetition algorithm (SM-2 variant)
- [ ] Implement review rating input (again, hard, good, easy)
- [ ] Update MemoryFSM states based on review results
- [ ] Calculate next review dates based on ease factor and interval

### Phase 7: TipTap JSON Serialization
- [ ] Implement TipTap JSON validation for notes and memory cards
- [ ] Implement TipTap JSON serializer for storage
- [ ] Implement TipTap JSON deserializer for retrieval
- [ ] Add round-trip verification tests
- [ ] Update content field to store TipTap JSON string

### Phase 8: Learning State Persistence
- [ ] Implement LearningFSM state persistence to BoltDB/SQLite
- [ ] Implement automatic state backup
- [ ] Implement state recovery on system startup
- [ ] Validate persisted state integrity
- [ ] Implement periodic state backup (every 5 minutes)

### Phase 9: Refactor Existing Services
- [ ] Refactor BookmarkService to use MemoryFSM for generated cards
- [ ] Refactor NoteService to use MemoryFSM for state management
- [ ] Refactor MemoryCardService to use unified MemoryFSM
- [ ] Update database migrations for new URN fields
- [ ] Maintain backward compatibility with existing data

### Phase 10: RPC Handler Extensions
- [ ] Add RPC handler for marking objects as learned
- [ ] Add RPC handler for creating notes in viewing context
- [ ] Add RPC handler for creating memory cards in viewing context
- [ ] Add RPC handler for managing URN links
- [ ] Update service.md documentation for all new RPC handlers

### Phase 11: Frontend Integration
- [ ] Implement unified viewing interface for all learning objects
- [ ] Implement consistent TipTap editor component
- [ ] Implement URN selection and linking interface
- [ ] Add memory card review interface with rating buttons
- [ ] Implement consistent navigation patterns and breadcrumbs

### Phase 12: Testing and Validation
- [ ] Write unit tests for MemoryFSM state transitions
- [ ] Write unit tests for LearningFSM state persistence
- [ ] Write integration tests for URN link management
- [ ] Write tests for TipTap JSON serialization round-trip
- [ ] Write end-to-end tests for passive learning workflow
- [ ] Validate spaced repetition algorithm calculations
- [ ] Test backward compatibility with existing data

### Phase 13: Documentation
- [ ] Update pkg/notes/service.md with new RPC handlers
- [ ] Update design documentation with MemoryFSM and LearningFSM details
- [ ] Document internal learning strategies (BFS/DFS)
- [ ] Document data migration strategy
- [ ] Create user guide for passive learning experience
