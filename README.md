# v2e - Vulnerability Viewer Engine

A broker-first microservices system for managing CVE, CWE, CAPEC, ATT&CK, OWASP ASVS, SSG, and CCE security data.

## Design Philosophy

v2e implements a **broker-first architecture** where the broker serves as the central orchestrator that spawns, monitors, and manages all subprocess services.

### Core Principles

| Principle | Description |
|-----------|-------------|
| **Centralized Control** | Broker is the sole process manager and message router |
| **No Direct Communication** | Subprocesses cannot communicate directly with each other |
| **All Traffic Through Broker** | Every inter-service message flows through the broker |
| **Clean Separation** | Each service has a single responsibility |

### Architecture Benefits

- **Observability**: Complete visibility into all inter-service communication
- **Resilience**: Failures are contained and can be recovered gracefully
- **Security**: No unauthorized cross-service communication paths
- **Simplicity**: Clear data flow and easy debugging

---

## Unified Frameworks

v2e is built on four unified frameworks that provide separation of concerns while enabling comprehensive security data management.

### UEE - Unified ETL Engine

**Purpose**: Resource-aware ETL orchestration for security data ingestion.

**Why UEE?**
- Separates resource management (broker) from ETL logic (meta service)
- Provides observable, resumable workflows instead of hardcoded sync loops
- Prevents resource exhaustion through worker permits and quotas

**State Machine**:
```
IDLE → ACQUIRING → RUNNING → WAITING_QUOTA/WAITING_BACKOFF → PAUSED → TERMINATED
```

**Providers**:
- CVEProvider - NVD API with incremental updates
- CWEProvider - MITRE CWE import
- CAPECProvider - MITRE CAPEC with XSD validation
- ATTACKProvider - MITRE ATT&CK techniques
- SSGProvider - SCAP Security Guide import
- ASVSProvider - OWASP ASVS import
- CCEProvider - Common Configuration Enumeration import
- NoteProvider - Note import
- MemoryCardProvider - Memory card import

### UDA - Unified Data Analysis

**Purpose**: URN-based relationship graph analysis for security entity correlation.

**Why UDA?**
- Creates unified view across CVE, CWE, CAPEC, ATT&CK data
- Enables attack path discovery (CVE → CWE → CAPEC → ATT&CK)
- In-memory graph provides sub-microsecond query performance

**Graph Operations**:
- Add/retrieve nodes by URN
- Find neighbors (incoming/outgoing connections)
- BFS shortest path between entities
- Filter by type or provider

**Performance**:
- AddNode: ~150-200 ns/op
- GetNode: ~100-150 ns/op
- FindPath: ~10-15 us/op (4-node path)

### UME - Unified Message Exchanging

**Purpose**: High-performance message routing and transport management.

**Why UME?**
- Single communication pattern across all services
- Binary protocol with configurable encoding (JSON/GOB/PLAIN)
- Adaptive optimization based on workload
- Zero-copy operations on Linux

**Transport Features**:
- Unix Domain Sockets (0600 permissions)
- 128-byte fixed header binary protocol
- Message pooling with sync.Pool
- Adaptive worker pools and batching

### ULP - Unified Learning Portal

**Purpose**: Intelligent learning session management with adaptive navigation strategies for security knowledge acquisition.

**Why ULP?**
- Provides structured learning workflows for security practitioners
- Implements FSM-based session state management
- Supports multiple navigation strategies (BFS/DFS) for different learning styles
- Enables progress tracking and review scheduling

**Core Components**:

**FSM (Finite State Machine)** - `pkg/notes/fsm/`
- Manages learning session states: idle, browsing, deep_dive, reviewing, paused
- Tracks user progress with viewed/completed items
- Maintains item relationship graph for DFS navigation
- Supports session persistence with SQLite storage

**Strategy Pattern** - `pkg/notes/strategy/`
- **BFS Strategy**: Sequential browsing through available items
- **DFS Strategy**: Deep dive into related concepts via links
- Dynamic strategy switching based on user behavior
- Path stack management for navigation history

**Learning States**:
```
IDLE → BROWSING → DEEP_DIVE → REVIEWING → PAUSED
         ↓            ↓            ↓
       (BFS)       (DFS)      (Review)
```

**Key Features**:
- ItemGraph: Bidirectional link graph for DFS traversal
- Path stack: Navigate back through learning history
- Context tracking: Viewed items, completed items, available items
- Session metrics: Start time, last activity, strategy usage

**Architecture**:
```
LearningFSM (State Machine)
    ├── Storage (SQLite persistence via v2local)
    ├── ItemGraph (Link relationships)
    ├── Strategy (Navigation logic)
    │   ├── BFS Strategy (Sequential browsing)
    │   └── DFS Strategy (Deep dive exploration)
    └── Context (Learning progress tracking)
        ├── ViewedItems
        ├── CompletedItems
        ├── AvailableItems
        └── PathStack
```

**RPC Integration**:
Learning session management is integrated into v2local service:
- Bookmark-based learning sessions with state tracking
- Memory card creation with TipTap content support
- Learning state management (new, learning, reviewing, mastered)
- Strategy-based navigation for related concepts

**TipTap JSON Schema**:

TipTap is a headless editor framework that uses JSON to represent rich text content.

**Document Structure**:
```json
{
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        {
          "type": "text",
          "text": "Hello, world!"
        }
      ]
    }
  ]
}
```

**Supported Node Types**:
- **Document Structure**: doc, paragraph, heading, codeBlock, blockquote, listItem, bulletList, orderedList, text, hardBreak
- **Formatting**: bold, italic, strike, code, link
- **Task Lists**: taskList, taskItem
- **Common Extensions**: image, horizontalRule

**Supported Mark Types** (text formatting):
- bold, italic, strike, code, link (with required href attribute)

**Validation Rules**:
1. Root node must be type "doc"
2. All node types must be in ValidTipTapNodeTypes whitelist
3. All mark types must be in ValidTipTapMarkTypes whitelist
4. Text nodes must have non-empty text content or child nodes
5. Link marks must have href attribute
6. Recursive validation of all nested content nodes

**Helper Functions**:
- `ValidateTipTapJSON(content string)`: Validate JSON string is valid TipTap
- `ValidateTipTapDocument(doc *TipTapDocument)`: Validate document structure
- `IsTipTapJSONEmpty(content string)`: Check if content is effectively empty
- `GetTipTapText(content string)`: Extract plain text from TipTap JSON
- `CreateEmptyTipTapDocument()`: Create empty document structure
- `CreateTipTapDocumentFromText(text string)`: Convert plain text to TipTap JSON

### GLC - Graphized Learning Canvas

**Purpose**: Interactive visual canvas for creating and managing knowledge graphs with customizable node types and relationships.

**Why GLC?**
- Visual knowledge organization with drag-and-drop canvas
- Customizable node types with user-defined presets
- Version history for undo/restore operations
- Share links for collaboration and embedding

**Core Components**:

**Graph Model** - `pkg/glc/`
- GraphModel: Main graph entity with nodes, edges, viewport state
- GraphVersionModel: Version snapshots for undo functionality
- UserPresetModel: Custom presets with themes and node type definitions
- ShareLinkModel: Shareable links with optional password protection

**Node and Edge System**:
- CADNode: Canvas nodes with position, size, style, and data
- CADEdge: Connections between nodes with styling options
- Viewport: Camera state (zoom, pan) for canvas navigation

**Preset System**:
- Built-in presets for common use cases
- User-defined presets with custom themes and behaviors
- NodeTypeDefinition: Custom node types with styling
- RelationshipDefinition: Typed connections between nodes

**Key Features**:
- Real-time collaboration ready
- Version control with automatic snapshots
- Export/import in multiple formats
- Responsive canvas with smooth pan/zoom

**RPC Operations**:
- Graph CRUD: Create, Get, Update, Delete, List, ListRecent
- Version management: Get, List, Restore
- Preset management: Create, Get, Update, Delete, List
- Share links: Create, Get, Delete

**Storage**:
- SQLite via GORM with soft delete support
- Automatic version creation on graph updates
- Cascade delete for related entities

---

## UDE - Unified Desktop Experience

**Purpose**: A modern, responsive web interface for interacting with the v2e security data platform.

**Why UDE?**
- Single unified interface for all security data (CVE, CWE, CAPEC, ATT&CK, ASVS, SSG, CCE)
- Real-time data visualization with interactive tables and modals
- Integrated learning portal (ULP) for spaced repetition
- Knowledge graph visualization (GLC) for relationship exploration

**Tech Stack**:
- Next.js 15+ with App Router and Static Site Generation
- Tailwind CSS v4 + shadcn/ui (Radix UI components)
- TanStack Query v5 for data fetching
- Lucide React icons

**Key Features**:

**Data Browsing**:
- CVE, CWE, CAPEC, ATT&CK, ASVS, SSG, CCE data tables
- Paginated lists with filtering and sorting
- Detail modals with comprehensive entity information

**Learning Portal (ULP Integration)**:
- Bookmark security entities for learning
- Memory cards with rich text (TipTap)
- Spaced repetition scheduling
- BFS/DFS navigation strategies

**Graph Visualization (GLC Integration)**:
- Interactive canvas for knowledge graphs
- Node/edge visualization
- Pan and zoom controls
- Version history

**RPC Client Pattern**:
```typescript
POST /restful/rpc with {method, target, params}
Automatic case conversion: camelCase ↔ snake_case
Mock mode: NEXT_PUBLIC_USE_MOCK_DATA=true
```

---

## Advanced Features

### Auto-Scaling (`cmd/v2broker/scaling/`)

The broker includes an intelligent auto-scaling system for dynamic resource management:

**Components**:
- **AutoScaler**: Makes scaling decisions based on predicted load
- **LoadPredictor**: Time-series forecasting for proactive scaling
- **AnomalyDetector**: Identifies unusual patterns in metrics
- **SelfHealing**: Automatic recovery from failures

**Scaling Decisions**:
- `scale_up`: Increase worker count based on load prediction
- `scale_down`: Decrease workers during low load periods
- `none`: No action needed

**Configuration**:
- Min/Max workers bounds
- Scale thresholds (CPU, memory, latency)
- Prediction horizon for proactive scaling
- Cooldown periods to prevent thrashing

### eBPF Monitoring (`cmd/v2broker/monitor/`)

Low-overhead kernel-level monitoring for deep observability:

**Probe Types**:
- `uds`: Unix Domain Socket performance
- `sharedmem`: Shared memory operations
- `locks`: Lock contention analysis
- `scheduling`: CPU scheduling events
- `memory`: Memory allocation patterns
- `io`: I/O operations
- `goroutines`: Go runtime scheduler events
- `gc`: Garbage collection pauses

**Features**:
- Sub-microsecond overhead
- Stack trace capture for profiling
- Configurable sample rates
- Alert thresholds for anomalies
- Flame graph generation support

### Message Queue (`cmd/v2broker/mq/`)

High-performance message bus for internal communication:

**Features**:
- In-memory message queue with configurable buffering
- Pub/sub pattern for event broadcasting
- Backpressure handling with flow control
- Dead letter queue for failed messages

---

## Service-Framework Matrix

| Service | UEE | UDA | UME | ULP | UDE | Key Responsibilities |
|---------|:---:|:---:|:---:|:-----:|:---:|----------------------|
| **v2broker** | - | - | X | - | - | Central orchestrator, process management, message routing, permit management, auto-scaling, eBPF monitoring |
| **v2access** | - | - | X | - | - | REST gateway, frontend communication, HTTP to RPC translation |
| **v2local** | X | X | - | X | - | Data persistence (CVE/CWE/CAPEC/ATT&CK/ASVS/SSG/CCE), CRUD operations, caching, bookmarks, learning sessions, memory cards, GLC graphs |
| **v2remote** | X | - | - | - | - | External API integration (NVD, MITRE), rate limiting, retry mechanisms |
| **v2meta** | X | - | X | - | - | ETL orchestration, provider management, URN checkpointing, state machines |
| **v2sysmon** | - | - | X | - | - | System metrics collection, health monitoring, performance reporting |
| **v2analysis** | - | X | X | - | - | Graph database management, relationship analysis, attack path discovery |
| **website (UDE)** | - | - | - | X | X | Unified Desktop Experience - frontend interface for all services |

### Framework Distribution

| Framework | Primary Service | Secondary Services | Building Blocks |
|-----------|----------------|-------------------|----------------|
| **UEE** (Unified ETL Engine) | v2meta | v2local, v2remote | URN (checkpointing), RPC (coordination) |
| **UDA** (Unified Data Analysis) | v2analysis | v2local (data source) | URN (node IDs), RPC (queries) |
| **UME** (Unified Message Exchanging) | v2broker | All services | RPC (message protocol), Binary Protocol |
| **ULP** (Unified Learning Portal) | v2local (bookmarks, memory cards, learning sessions, GLC graphs) | v2access (gateway) | FSM (state management), Strategy (navigation patterns), Storage (SQLite) |
| **UDE** (Unified Desktop Experience) | website | v2access (gateway) | Next.js, TanStack Query, React |

---

## System Overview

```
+----------------+      +-------------+      +---------+
| UDE Frontend  |----->| Access Svc  |----->| Broker  |
| (Next.js)     |      +-------------+      +---------+
+----------------+                              |
                                                |
                  +----------------------------------------+
                  |                                        |
                  v                                        v
       +----------+----------+          +----------+------------------+
       |   v2local          |          |   v2meta                   |
       |   (Data Storage)   |          |   (UEE Framework)          |
       +--------------------+          +----------------------------+
                  +----------+------------------+
                  |          |                |
                  v          v                v
       +----------+   +-----+-----+   +-------+------+
       | v2remote |   |v2sysmon |   |v2analysis    |
       | (APIs)   |   | (Monitor)|   | (UDA Graph)  |
       +----------+   +---------+   +--------------+
```

---

## Binary Message Protocol

The broker implements a 128-byte fixed header protocol for high performance:

| Offset | Size | Field | Purpose |
|--------|------|-------|---------|
| 0-1 | 2B | Magic | Protocol identification (0x56 0x32 = 'V2') |
| 2 | 1B | Version | Protocol version |
| 3 | 1B | Encoding | Payload encoding (0=JSON, 1=GOB, 2=PLAIN) |
| 4 | 1B | MsgType | Message type (request/response/event/error) |
| 8-11 | 4B | PayloadLen | Payload length in bytes |
| 12-43 | 32B | MessageID | Unique message identifier |
| 44-75 | 32B | SourceID | Sending process identifier |
| 76-107 | 32B | TargetID | Receiving process identifier |
| 108-127 | 20B | CorrelationID | Request-response matching |

### Encoding Performance

| Operation | JSON | GOB | PLAIN |
|-----------|------|-----|-------|
| Small Message Marshal | 418 ns/op | 1286 ns/op | 669 ns/op |
| Small Message Unmarshal | 236 ns/op | 1592 ns/op | 2060 ns/op |
| Round-trip Latency | ~2.1 us | ~4.6 us | ~4.5 us |

**Recommendation**: Use JSON encoding (default) for optimal performance.

---

## URN - Atomic Identifiers

URNs provide hierarchical, immutable identification for all security data entities.

**Format**: `v2e::<provider>::<type>::<atomic_id>`

**Examples**:
```
v2e::nvd::cve::CVE-2024-12233
v2e::mitre::cwe::CWE-79
v2e::mitre::capec::CAPEC-66
v2e::mitre::attack::T1566
v2e::ssg::ssg::rhel9-guide-ospp
```

**Providers**: nvd, mitre, ssg
**Types**: cve, cwe, capec, attack, ssg, asvs, cce

**Why URN?**
- Immutable identity across services and databases
- Type safety through structured validation
- Enables relationship tracking between entities
- Supports checkpoint/resume in ETL pipelines

---

## Communication Flow

1. **Frontend Request** → Access Service REST API (`/restful/rpc`)
2. **Access → Broker** → Request routed via UDS
3. **Broker → Target Service** → Message delivered to subprocess
4. **Response Path** → Broker → Access Service → Frontend

**Rules**:
- No direct subprocess-to-subprocess communication
- All traffic through broker
- UDS-only transport (0600 permissions)

---

## Quick Start

```bash
# Prerequisites: Go 1.25.6, Node.js 20+, npm 10+

# Run full development environment (recommended)
./build.sh -r

# Run unit tests
./build.sh -t

# Build and package
./build.sh -p
```

### Build Script Options

| Option | Description |
|--------|-------------|
| `-c` | Run vconfig TUI for configuration |
| `-t` | Run unit tests |
| `-f` | Run fuzz tests |
| `-m` | Run benchmarks |
| `-p` | Build and package |
| `-r` | Run full system |

---

## Project Structure

```
  cmd/
    v2broker/           # Broker service (UME framework), auto-scaling, eBPF monitoring
    v2access/           # REST gateway
    v2local/            # Data persistence (CVE/CWE/CAPEC/ATT&CK/ASVS/SSG/CCE), bookmarks, memory cards, learning sessions, GLC graphs
    v2remote/           # External API integration
    v2meta/             # ETL orchestration (UEE framework)
    v2sysmon/           # System monitoring
    v2analysis/         # Graph analysis (UDA framework)
  pkg/
    proc/               # Subprocess framework
    message/            # Message handling with pooling
    urn/                # URN atomic identifiers
    rpc/                # RPC client helpers
    graph/              # In-memory graph database
    analysis/           # FSM and storage for UDA
    notes/              # ULP framework: FSM, strategy pattern, SQLite storage
    cve/taskflow/       # ETL executor framework
    ssg/                # SSG (SCAP Security Guide) parsing and models
    cce/                # CCE (Common Configuration Enumeration) models
    asvs/               # ASVS (Application Security Verification Standard) models
    glc/                # GLC (Graphized Learning Canvas) models and storage
    jsonutil/           # JSON utilities with Sonic optimization
  website/              # UDE (Unified Desktop Experience) - Next.js frontend
  assets/               # Data assets (CWE, CAPEC, ATT&CK)
```

---

## Further Reading

- [cmd/v2meta](cmd/v2meta) - UEE framework documentation
- [cmd/v2analysis](cmd/v2analysis) - UDA framework documentation
- [cmd/v2broker](cmd/v2broker) - Broker implementation, auto-scaling, eBPF monitoring
- [cmd/v2local](cmd/v2local) - Local data storage, bookmarks, memory cards, learning sessions, GLC graphs
- [pkg/urn](pkg/urn) - URN identifier implementation
- [cmd/v2meta/providers](cmd/v2meta/providers) - ETL provider implementations
- [pkg/notes](pkg/notes) - ULP (Unified Learning Portal) framework documentation
- [pkg/glc](pkg/glc) - GLC (Graphized Learning Canvas) models and storage
- [website](website) - UDE (Unified Desktop Experience) frontend documentation

---

## License

MIT
