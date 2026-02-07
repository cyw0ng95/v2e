# v2e - Vulnerability Viewer Engine

A broker-first microservices system for managing CVE, CWE, CAPEC, ATT&CK, and OWASP ASVS security data.

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

v2e is built on three unified frameworks that provide separation of concerns while enabling comprehensive security data management.

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

### Notes - Learning Strategy System

**Purpose**: Intelligent learning session management with adaptive navigation strategies for security knowledge acquisition.

**Why Notes?**
- Provides structured learning workflows for security practitioners
- Implements FSM-based session state management
- Supports multiple navigation strategies (BFS/DFS) for different learning styles
- Enables progress tracking and review scheduling

**Core Components**:

**FSM (Finite State Machine)** - `pkg/notes/fsm/`
- Manages learning session states: idle, browsing, deep_dive, reviewing, paused
- Tracks user progress with viewed/completed items
- Maintains item relationship graph for DFS navigation
- Supports session persistence with BoltDB storage

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
    ├── Storage (BoltDB persistence)
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
- `RPCLearningStart`: Initialize learning session with available items
- `RPCLearningNextItem`: Get next item based on current strategy
- `RPCLearningViewItem`: Mark item as viewed and update context
- `RPCLearningFollowLink`: Follow related link (DFS mode)
- `RPCLearningGoBack`: Navigate back in path history
- `RPCLearningSwitchStrategy`: Switch between BFS and DFS modes
- `RPCLearningCompleteItem`: Mark item as learned
- `RPCLearningPauseResume`: Pause or resume learning session

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

---

## Service-Framework Matrix

| Service | UEE | UDA | UME | Notes | Key Responsibilities |
|---------|:---:|:---:|:---:|:-----:|----------------------|
| **v2broker** | - | - | X | - | Central orchestrator, process management, message routing, permit management |
| **v2access** | - | - | X | - | REST gateway, frontend communication, HTTP to RPC translation |
| **v2local** | X | X | - | - | Data persistence (CVE/CWE/CAPEC/ATT&CK/ASVS), CRUD operations, caching |
| **v2remote** | X | - | - | - | External API integration (NVD, MITRE), rate limiting, retry mechanisms |
| **v2meta** | X | - | X | - | ETL orchestration, provider management, URN checkpointing, state machines |
| **v2sysmon** | - | - | X | - | System metrics collection, health monitoring, performance reporting |
| **v2analysis** | - | X | X | - | Graph database management, relationship analysis, attack path discovery |
| **v2notes** | - | - | X | X | Learning session management, bookmark/note/memory card storage, strategy-based navigation |

### Framework Distribution

| Framework | Primary Service | Secondary Services | Building Blocks |
|-----------|----------------|-------------------|----------------|
| **UEE** (Unified ETL Engine) | v2meta | v2local, v2remote | URN (checkpointing), RPC (coordination) |
| **UDA** (Unified Data Analysis) | v2analysis | v2local (data source) | URN (node IDs), RPC (queries) |
| **UME** (Unified Message Exchanging) | v2broker | All services | RPC (message protocol), Binary Protocol |
| **Notes** (Learning Strategy System) | v2notes | v2local (bookmark storage), v2access (gateway) | FSM (state management), Strategy (navigation patterns), Storage (BoltDB) |

---

## System Overview

```
+----------------+      +-------------+      +---------+
| Next.js Frontend|----->| Access Svc  |----->| Broker  |
+----------------+      +-------------+      +---------+
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
                |
                v
          +------------+
          | v2notes    |
          | (Learning  |
          |  Sessions) |
          +------------+
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
**Types**: cve, cwe, capec, attack, ssg

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
# Prerequisites: Go 1.21+, Node.js 20+, npm 10+

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
  v2broker/           # Broker service (UME framework)
  v2access/           # REST gateway
  v2local/            # Data persistence (CVE/CWE/CAPEC/ATT&CK/ASVS)
  v2remote/           # External API integration
  v2meta/             # ETL orchestration (UEE framework)
  v2sysmon/           # System monitoring
  v2analysis/          # Graph analysis (UDA framework)
pkg/
  proc/               # Subprocess framework
  message/            # Message handling with pooling
  urn/                # URN atomic identifiers
  rpc/                # RPC client helpers
  graph/              # In-memory graph database
  analysis/           # FSM and storage for UDA
  notes/              # Learning strategy system (FSM, strategy pattern, BoltDB storage)
  cve/taskflow/       # ETL executor framework
website/              # Next.js frontend
assets/               # Data assets (CWE, CAPEC, ATT&CK)
```

---

## Further Reading

- [cmd/v2meta](cmd/v2meta) - UEE framework documentation
- [cmd/v2analysis](cmd/v2analysis) - UDA framework documentation
- [cmd/v2broker](cmd/v2broker) - Broker implementation
- [pkg/urn](pkg/urn) - URN identifier implementation
- [cmd/v2meta/providers](cmd/v2meta/providers) - ETL provider implementations
- [pkg/notes](pkg/notes) - Learning strategy system documentation

---

## License

MIT
