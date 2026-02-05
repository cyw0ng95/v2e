# v2e

A sophisticated Go-based system that demonstrates a broker-first architecture for orchestrating multiple subprocess services that communicate via RPC messages over stdin/stdout. The system provides a comprehensive CVE (Common Vulnerabilities and Exposures) management platform with integrated CWE (Common Weakness Enumeration), CAPEC (Common Attack Pattern Enumeration and Classification), and ATT&CK (Adversarial Tactics, Techniques, and Common Knowledge) framework data handling.

## Executive Summary

The v2e project implements a broker-first architecture where `cmd/v2broker` serves as the central process manager that spawns, monitors, and manages all subprocess services. This design enforces a strict communication pattern where all inter-service communication flows through the broker, preventing direct subprocess-to-subprocess interaction. The architecture ensures clean separation of concerns while maintaining robust message routing and process lifecycle management.

Key architectural principles:
- **Centralized Process Management**: The broker is the sole orchestrator of all subprocess services
- **Enforced Communication Pattern**: All inter-service communication occurs through broker routing
- **RPC-Based Messaging**: Services communicate via structured JSON RPC messages over stdin/stdout
- **Comprehensive Data Handling**: Integrated CVE, CWE, CAPEC, and ATT&CK data management
- **Frontend Integration**: A Next.js-based web application provides user interface access
- **Performance Monitoring**: Built-in metrics collection and system monitoring capabilities
- **Message Optimization**: Asynchronous message routing with configurable buffering and batching
- **Adaptive Optimization**: Dynamic performance tuning based on workload with adaptive algorithms
- **Enhanced Configuration**: Advanced configuration management via vconfig tool with TUI interface
- **Cross-Platform Support**: Containerized development environment for macOS with Linux support
- **Linux-Native Performance**: CPU affinity binding, thread pinning, and kernel memory hints for deterministic low-latency operation
- **Binary Message Protocol**: High-performance 128-byte fixed header with multiple encoding options (JSON/GOB/PLAIN)
- **Comprehensive Telemetry**: Wire-size tracking, encoding distribution, and per-process metrics

## Binary Message Protocol

The v2e broker implements an optimized binary message protocol with a 128-byte fixed header, providing significant performance improvements over pure JSON encoding.

### Header Layout

The binary header consists of exactly 128 bytes with the following structure:

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0-1 | 2 bytes | Magic | Protocol identifier (0x56 0x32 = 'V2') |
| 2 | 1 byte | Version | Protocol version (0x01) |
| 3 | 1 byte | Encoding | Payload encoding (0=JSON, 1=GOB, 2=PLAIN) |
| 4 | 1 byte | MsgType | Message type (0=Request, 1=Response, 2=Event, 3=Error) |
| 5-7 | 3 bytes | Reserved | Reserved for future use |
| 8-11 | 4 bytes | PayloadLen | Payload length (uint32, big-endian) |
| 12-43 | 32 bytes | MessageID | Message ID (null-terminated string) |
| 44-75 | 32 bytes | SourceID | Source process ID (null-terminated string) |
| 76-107 | 32 bytes | TargetID | Target process ID (null-terminated string) |
| 108-127 | 20 bytes | CorrelationID | Correlation ID for request-response matching |

### Encoding Options

Three encoding types are supported for payload serialization:

1. **JSON (Type 0)** - Default encoding, fastest for small messages
   - Best for messages < 1KB
   - Unmarshal: 236 ns/op, 304 B/op
   - Marshal: 418 ns/op, 424 B/op

2. **GOB (Type 1)** - Go-native binary encoding
   - Better for large structured payloads
   - Unmarshal: 1592 ns/op, 1432 B/op
   - Marshal: 1286 ns/op, 1360 B/op

3. **PLAIN (Type 2)** - Raw bytes without encoding
   - Most efficient for binary data
   - No serialization overhead

### Benchmark Results

Performance comparison (Intel Xeon 8370C @ 2.80GHz):

| Operation | JSON | GOB | PlainJSON | Winner |
|-----------|------|-----|-----------|--------|
| Small Message Marshal | 418 ns/op | 1286 ns/op | 669 ns/op | **JSON** (3.1x faster than GOB) |
| Small Message Unmarshal | 236 ns/op | 1592 ns/op | 2060 ns/op | **JSON** (6.7x faster than GOB) |
| Round-trip | 2139 ns/op | 4595 ns/op | 4508 ns/op | **JSON** (2.1x faster than GOB) |
| Large Payload Marshal | 5430 ns/op | 17359 ns/op | 106225 ns/op | **JSON** (3.2x faster than GOB) |

**Recommendation**: Use JSON encoding (default) for optimal performance on typical message sizes.

### Linux-Specific Optimizations

On Linux platforms, the following optimizations are automatically enabled:

- **Zero-copy operations**: `splice()` and `sendfile()` syscalls for efficient data transfer
- **Socket tuning**: TCP_NODELAY, TCP_QUICKACK, optimized buffer sizes
- **Memory hints**: `madvise()` for sequential access patterns
- **CPU affinity**: Thread pinning for deterministic latency
- **Optimized memcpy**: Direct memory operations bypassing bounds checking

### Usage Example

```go
import "github.com/cyw0ng95/v2e/pkg/proc"

// Create and marshal a message (default JSON encoding)
msg, _ := proc.NewRequestMessage("RPCGetStatus", map[string]string{
    "component": "broker",
})
msg.Source = "client"
msg.Target = "broker"

data, _ := msg.MarshalBinary() // Uses JSON encoding by default

// Check message type
if proc.IsBinaryMessage(data) {
    decoded, _ := proc.UnmarshalBinary(data)
    // Process decoded message
}

// Use GOB encoding for large structured data
largeData, _ := proc.MarshalBinaryWithEncoding(msg, proc.EncodingGOB)
```

### Metrics & Telemetry

The broker tracks comprehensive message statistics:

- **Global metrics**: Total messages/bytes sent/received, encoding distribution
- **Per-process metrics**: Message counts and byte totals per process
- **Encoding distribution**: Breakdown of JSON/GOB/PLAIN usage
- **Wire-size tracking**: Accurate byte-level bandwidth monitoring

Access metrics via RPC:
```go
// Get detailed statistics
stats, _ := broker.HandleRPCGetMessageStats(reqMsg)
// Returns: total_bytes_sent, total_bytes_received, encoding_distribution, per_process stats

// Get message count
count, _ := broker.HandleRPCGetMessageCount(reqMsg)
```

## System Architecture

```mermaid
graph TB
    subgraph "Frontend Layer"
        A[Next.js Web App]
    end

    subgraph "Broker Layer"
        B[Broker Service]
        Opt[Adaptive Optimizer]
    end

    subgraph "Backend Services"
        C[Access Service]
        D[Meta Service]
        E[Local Service]
        F[Remote Service]
        G[SysMon Service]
    end

    subgraph "Transport Layer"
        UDS[Unix Domain Sockets]
    end

    A <--> C
    C <--> B
    B <--> D
    B <--> E
    B <--> F
    B <--> G
    B <--> Opt
    B <--> UDS
    UDS <--> C
    UDS <--> D
    UDS <--> E
    UDS <--> F
    UDS <--> G
```

The system utilizes a Unix Domain Sockets (UDS) transport layer with 0600 permissions for secure inter-process communication. The broker incorporates an advanced adaptive optimizer that dynamically adjusts performance parameters based on system load and message throughput.

### Unified ETL Engine (UEE) Architecture

The v2e system implements a **Master-Slave hierarchical FSM (Finite State Machine) model** for resource-aware ETL orchestration, replacing hardcoded sync loops with an observable, resumable workflow engine.

#### Master-Slave Roles

- **Master (Broker)**: The technical resource authority
  - Manages a global pool of "Worker Permits" for concurrency control
  - Monitors kernel metrics: P99 latency (< 20ms target), buffer saturation, message rates
  - Revokes permits when thresholds breach (P99 > 30ms OR buffer > 80%)
  - Broadcasts `RPCOnQuotaUpdate` events to providers
  - **Pure technical layer**: No business logic, only resource management

- **Slave (Meta Service)**: The ETL orchestrator
  - Manages domain logic (what to fetch, how to parse, where to store)
  - Requests permits before starting providers
  - Handles quota revocations gracefully (transitions providers to `WAITING_QUOTA`)
  - Coordinates hierarchical state machines (Macro FSM + Provider FSMs)

#### Hierarchical State Machines

**Macro FSM (High-Level Orchestration)**:
```
BOOTSTRAPPING → ORCHESTRATING → STABILIZING → DRAINING
                      ↓
                (emergency drain)
```

**Provider FSM (Worker-Level Execution)**:
```
IDLE → ACQUIRING → RUNNING → WAITING_QUOTA/WAITING_BACKOFF → PAUSED → TERMINATED
                      ↓
              (permit granted)
```

#### URN Atomic Identifiers

All ETL items use hierarchical URN keys:
```
v2e::<provider>::<type>::<atomic_id>
Examples:
  v2e::nvd::cve::CVE-2024-12233
  v2e::mitre::cwe::CWE-79
  v2e::mitre::capec::CAPEC-66
```

URNs enable:
- Immutable identity across checkpoints and lookups
- Resumable workflows (checkpointed every 100 items)
- URN-validated persistence in BoltDB

#### Auto-Recovery

On service restart:
- **RUNNING providers**: Resume execution with permit re-acquisition
- **PAUSED providers**: Remain paused (manual resume required)
- **WAITING_QUOTA providers**: Retry permit requests
- **WAITING_BACKOFF providers**: Maintain state (auto-retry timer continues)
- **TERMINATED providers**: Skipped (not recovered)

#### Architecture Flow

```
Meta Service (Slave)              Broker (Master)
    │                                 │
    │ StartProvider("cve", 5)         │
    ├─ RPCRequestPermits(5) ─────────→│
    │                                 │ Allocate 5 permits
    │←────────────────────────────────┤ Response: granted=5
    │                                 │
    │ provider.OnQuotaGranted(5)      │
    │  → RUNNING state                │
    │                                 │
    │ [Execute: fetch/parse/store]    │ [Monitor: P99 latency]
    │                                 │
    │                                 │ P99 > 30ms detected!
    │←─ RPCOnQuotaUpdate(revoked=2) ──┤
    │                                 │
    │ provider.OnQuotaRevoked(2)      │
    │  → WAITING_QUOTA state          │
```

## Component Breakdown

### Core Services

- **Broker Service** ([cmd/v2broker](cmd/v2broker)): The central orchestrator responsible for:
  - Spawning and managing all subprocess services with robust supervision and restart policies
  - Routing RPC messages via a high-performance Unix Domain Sockets (UDS) transport layer
  - Utilizing `bytedance/sonic` for zero-copy JSON serialization/deserialization
  - Implementing an adaptive traffic optimizer with configurable batching, buffering, and backpressure
  - Maintaining process lifecycle, health checks, and zombie process reaping
  - Tracking comprehensive real-time message statistics and performance metrics
  - Supporting advanced logging with dual output (console + file) and configurable log levels
  - Providing dynamic configuration of performance parameters via adaptive optimization algorithms
  - **Linux-native performance optimizations**: CPU affinity binding, thread pinning, process/I/O priority tuning for deterministic low-latency message routing (see [docs/LINUX_PERFORMANCE.md](docs/LINUX_PERFORMANCE.md))

- **Access Service** ([cmd/v2access](cmd/v2access)): The REST gateway that:
  - Serves as the primary interface for the Next.js frontend
  - Exposes `/restful/rpc` endpoint for RPC forwarding
  - Translates HTTP requests to RPC calls and responses back
  - Provides health checks and basic service discovery

- **Meta Service** ([cmd/v2meta](cmd/v2meta)): The orchestration layer that:
  - Manages job scheduling and execution using go-taskflow
  - Coordinates complex multi-step operations
  - Handles session management and state persistence
  - Provides workflow control mechanisms
  - Orchestrates CVE/CWE data fetching jobs with persistent state management
  - Performs automatic CWE and CAPEC imports at startup
  - Provides memory card management (delegates to local service):
    - RPCCreateMemoryCard, RPCGetMemoryCard, RPCUpdateMemoryCard
    - RPCDeleteMemoryCard, RPCListMemoryCards

- **Local Service** ([cmd/v2local](cmd/v2local)): The data persistence layer that:
  - Manages local SQLite databases for CVE, CWE, CAPEC, and ATT&CK data
  - Provides CRUD operations for vulnerability information
  - Handles data indexing and querying
  - Implements caching mechanisms for improved performance
  - Imports ATT&CK data from XLSX files and provides access to techniques, tactics, mitigations, software, and groups
  - Supports CAPEC XML schema validation and catalog metadata retrieval
  - Offers CWE view management with storage and retrieval capabilities
  - Provides memory card storage for bookmark/knowledge management:
    - RPCCreateMemoryCard, RPCGetMemoryCard, RPCUpdateMemoryCard
    - RPCDeleteMemoryCard, RPCListMemoryCards
    - Supports TipTap JSON content, classification fields, and metadata

- **Remote Service** ([cmd/v2remote](cmd/v2remote)): The data acquisition layer that:
  - Fetches vulnerability data from external APIs (NVD, etc.)
  - Implements rate limiting and retry mechanisms
  - Handles data transformation and normalization
  - Manages API credentials and authentication

- **SysMon Service** ([cmd/v2sysmon](cmd/v2sysmon)): The system monitoring layer that:
  - Collects performance metrics and system statistics
  - Monitors resource utilization across services
  - Provides health indicators for operational awareness
  - Reports system status to the frontend

## Configuration

The system uses a hybrid configuration approach: build-time configuration via `vconfig` (ldflags) for compile-time settings, and runtime configuration via `.config` for process definitions.

### vconfig TUI

Run `./build.sh -c` to access the interactive configuration manager.

#### Logging Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `CONFIG_MIN_LOG_LEVEL` | string | `INFO` | Minimum log level (DEBUG, INFO, WARN, ERROR) |
| `CONFIG_LOGGING_DIR` | string | `./logs` | Directory for log files |
| `CONFIG_LOGGING_REFRESH` | bool | `true` | Remove log directory first for fresh logs |

#### Access Service Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `CONFIG_ACCESS_SERVERADDR` | string | `0.0.0.0:8080` | Server address for access service |
| `CONFIG_ACCESS_STATICDIR` | string | `website` | Static directory for access service |

#### Transport Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `CONFIG_PROC_UDS_BASEPATH` | string | `/tmp/v2e_uds` | Base path for subprocess UDS sockets |
| `CONFIG_BROKER_UDS_BASEPATH` | string | `/tmp/v2e_uds` | Base path for broker UDS sockets |

#### Optimizer Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `CONFIG_OPTIMIZER_BUFFER` | int | `1000` | Message channel buffer capacity |
| `CONFIG_OPTIMIZER_WORKERS` | int | `4` | Worker goroutines count |
| `CONFIG_OPTIMIZER_BATCH` | int | `1` | Message batch size |
| `CONFIG_OPTIMIZER_FLUSH` | int | `10` | Flush interval (milliseconds) |
| `CONFIG_OPTIMIZER_POLICY` | string | `drop` | Offer policy (drop, wait, reject) |

## Transport & Communication

The system uses a **UDS-only** (Unix Domain Sockets) RPC communication mechanism designed for high throughput and low latency. Legacy stdin/stdout FD pipe support has been removed - all subprocess communication now exclusively uses UDS.

### Transport Architecture

```
Frontend → Access Service → Broker → Backend Services
                      ↓           ↓
                HTTP/REST      UDS (only)
                                    ↓
                            Subprocess Services
```

### Message Flow

1. **External requests** → Access REST API (`/restful/rpc`) → Broker
2. **Broker routing** → Backend Services via UDS (exclusively)
3. **Response path** → Broker → Access Service → Frontend
4. **No direct subprocess-to-subprocess communication** is allowed
5. **UDS-only transport**: All broker-to-subprocess communication uses Unix Domain Sockets with 0600 permissions

### Message Types

| Type | Purpose |
|------|---------|
| Request | RPC call with correlation ID |
| Response | RPC response matching correlation ID |
| Event | Asynchronous notification |
| Error | Error response with details |

### Key Features

- **Message Pooling**: `sync.Pool` for reduced GC pressure
- **Zero-Copy JSON**: `bytedance/sonic` for serialization
- **Correlation IDs**: Request-response matching
- **Per-Process Statistics**: Message counts and timing

## Frontend Integration

The Next.js-based frontend ([website](website)) provides:

- **REST Gateway Interface**: Access service exposes `/restful/rpc` endpoint for frontend-backend communication
- **Sophisticated RPC Client**: Handles automatic case conversion (camelCase ↔ snake_case) and comprehensive error handling
- **Rich Component Architecture**: Tabbed interface supporting CVE, CWE, CAPEC, and system monitoring data
- **Real-time Updates**: Session control and live metrics display
- **Responsive Design**: Adaptable interface for various screen sizes and devices
- **Modern Tech Stack**: Uses Next.js 16+, React 19+, with TypeScript, Tailwind CSS, and Radix UI components
- **Data Visualization**: Recharts for performance metrics and data visualization

The frontend includes dedicated sections for:
- CVE Database browsing and management
- CWE Database and view management
- CAPEC data visualization
- ATT&CK framework data (techniques, tactics, mitigations, software, groups)
- System monitoring and performance metrics
- Session control for data fetching jobs

## Quickstart

Prerequisites: Go 1.21+, Node.js 20+, npm 10+, and basic shell tools. For macOS users, Podman is required for containerized development.

**IMPORTANT:** Always use `./build.sh` for all builds and tests. Do not use direct `go build` or `go test` commands - the wrapper handles build tags, environment setup, and proper test configuration.

### Build Script Options

| Option | Description |
|--------|-------------|
| `-c` | Run vconfig TUI to configure build options |
| `-t` | Run unit tests (excludes fuzz tests, uses `-run='^Test'`) |
| `-f` | Run fuzz tests (5 seconds per target) |
| `-m` | Run benchmarks with reporting |
| `-p` | Build and package binaries + assets |
| `-r` | Run full system: broker + all subprocesses + frontend dev server |
| `-v` | Enable verbose output |
| `-h` | Show help message |

### Common Workflows

```bash
# Configure build options (vconfig TUI)
./build.sh -c

# Run unit tests
./build.sh -t

# Run fuzz tests
./build.sh -f

# Run performance benchmarks
./build.sh -m

# Build and package everything
./build.sh -p

# Run full development environment (recommended for testing)
./build.sh -r
```

### Development Mode (`./build.sh -r`)

Starts the complete system for integration testing and development:
- Broker with all subprocesses (access, local, meta, remote, sysmon)
- Frontend dev server on http://localhost:3000
- Logs written to `.build/package/logs/`

Press Ctrl+C to stop all services cleanly.

## Development Workflow

### Containerized Development Environment

For macOS users, a containerized development environment is available via `runenv.sh`:

```bash
# Run any command in containerized environment
./runenv.sh -t  # Run unit tests in container
./runenv.sh -f  # Run fuzz tests in container
./runenv.sh -m  # Run benchmarks in container
./runenv.sh -p  # Package in container
./runenv.sh -r  # Run development mode in container
```

On Linux, the containerized environment can be used with the `USE_CONTAINER=true` environment variable:

```bash
USE_CONTAINER=true ./runenv.sh -t  # Run tests in container on Linux
```
## Job Session Management & State Machine

The meta service orchestrates CVE/CWE data fetching jobs using go-taskflow with persistent state management. Job runs are stored in BoltDB and survive service restarts. The system also performs automatic CWE and CAPEC imports at startup.

### Job States

The system supports thirteen job states with strictly defined transitions, including intermediate states for granular progress tracking:

```mermaid
stateDiagram-v2
    [*] --> Queued: Create Run
    Queued --> Initializing: Start
    Queued --> Stopped: Stop
    Initializing --> Running: Ready
    Initializing --> Failed: Setup Error
    Initializing --> Stopped: Stop
    Running --> Paused: Pause
    Running --> Completed: Finish Successfully
    Running --> Failed: Error
    Running --> Stopped: Stop
    Running --> Fetching: API Call
    Running --> Processing: Transform
    Running --> Saving: Persist
    Running --> Validating: Verify
    Fetching --> Processing: Data Received
    Fetching --> Running: Batch Complete
    Fetching --> Failed: API Error
    Fetching --> Paused: Pause During Fetch
    Processing --> Saving: Processed
    Processing --> Running: Batch Complete
    Processing --> Failed: Transform Error
    Processing --> Paused: Pause During Process
    Saving --> Running: Persisted
    Saving --> Completed: All Saved
    Saving --> Failed: Storage Error
    Saving --> Paused: Pause During Save
    Validating --> Running: Validation Passed
    Validating --> Failed: Validation Failed
    Validating --> Paused: Pause During Validate
    Paused --> Recovering: Resume
    Paused --> Stopped: Stop
    Recovering --> Running: Recovery Complete
    Recovering --> Failed: Recovery Failed
    Recovering --> Stopped: Stop During Recovery
    Completed --> [*]
    Failed --> [*]
    Stopped --> [*]
    RollingBack --> Stopped: Rollback Complete
    RollingBack --> Failed: Rollback Failed
```

**Primary State Descriptions:**

- **Queued**: Job created but not yet started
- **Initializing**: Job is being set up, resources allocated
- **Running**: Job is actively executing
- **Paused**: Job temporarily paused by user (can be resumed)
- **Completed**: Job finished successfully (all data fetched)
- **Failed**: Job encountered fatal error
- **Stopped**: Job manually stopped by user

**Intermediate Progress States:**

- **Fetching**: Actively fetching data from remote API (NVD, etc.)
- **Processing**: Transforming and normalizing fetched data
- **Saving**: Persisting processed data to database
- **Validating**: Verifying data integrity and consistency
- **Recovering**: Recovering from pause or failure state
- **RollingBack**: Rolling back partial changes after error

**Terminal States**: Completed, Failed, Stopped (cannot transition further)

### Session Persistence & Recovery

- **Single Active Run Policy**: Only one job run can be active (running or paused) at time
- **BoltDB Storage**: Job runs persist in `session.db` (configurable via `SESSION_DB_PATH` env var)
- **Auto-Recovery**: On service restart:
  - Running jobs: Automatically resumed
  - Paused jobs: Remain paused (manual resume required)
  - Terminal states: No action taken

### RPC API for Job Control

**Start Session:**
```json
{
  "method": "RPCStartSession",
  "params": {
    "session_id": "my-job-001",
    "start_index": 0,
    "results_per_batch": 100
  }
}
```

**Stop Session:**
```json
{
  "method": "RPCStopSession",
  "params": {}
}
```

**Pause Job:**
```json
{
  "method": "RPCPauseJob",
  "params": {}
}
```

**Resume Job:**
```json
{
  "method": "RPCResumeJob",
  "params": {}
}
```

**Get Status:**
```json
{
  "method": "RPCGetSessionStatus",
  "params": {}
}
```

Response includes: `state`, `session_id`, `fetched_count`, `stored_count`, `error_count`, `error_message`

### Task Orchestration with go-taskflow

The meta service uses go-taskflow to orchestrate multi-step jobs:

1. **Fetch** task: Retrieve CVE batch from remote NVD API
2. **Store** task: Save CVEs to local database

Tasks are organized in a directed acyclic graph (DAG) with dependency management, retries, and cancellation support.

### Automatic Imports

The meta service performs automatic imports at startup:
- **CWE Import**: Triggers CWE import from `assets/cwe-raw.json` after 2-second delay
- **CAPEC Import**: Checks for existing CAPEC data and imports from `assets/capec_contents_latest.xml` if not present
- **ATT&CK Import**: Local service automatically imports ATT&CK data from XLSX files found in assets directory

## Configuration Guide

The system is configured through `config.json`, which controls:

- Process definitions and startup parameters
- Logging configuration and output destinations
- Service-specific settings
- RPC communication parameters
- Performance tuning options
- Optimizer runtime parameters for message routing

The broker reads this configuration at startup and uses it to determine which subprocess services to spawn and how to configure them.

### Broker Configuration Options

The broker supports the following configuration parameters in `config.json`:

- `broker.processes`: Array of process configurations with ID, command, arguments, RPC flag, and restart policy
- `broker.log_file`: Path to log file for dual output (stdout + file)
- `broker.logs_dir`: Directory where logs are stored
- `broker.authentication`: Authentication settings for RPC endpoints
- `broker.optimizer_*`: Various optimization parameters including buffer capacity, worker count, batching, and timeouts

## Performance Characteristics

The broker-first architecture provides several performance benefits through configurable optimization parameters.

### Optimizer Configuration

Build-time configurable via vconfig (`CONFIG_OPTIMIZER_*` options):

| Parameter | Default | Description |
|-----------|---------|-------------|
| `BufferCap` | 1000 | Message channel buffer capacity |
| `NumWorkers` | 4 | Number of parallel processing workers |
| `BatchSize` | 1 | Message batch size for throughput |
| `FlushInterval` | 10ms | Maximum time before flushing batches |
| `OfferPolicy` | drop | Strategy when buffer is full (drop/wait/reject) |

### Performance Benefits

| Feature | Benefit |
|---------|---------|
| **Efficient Message Routing** | Direct process-to-process communication through broker minimizes overhead |
| **Scalable Process Management** | Broker manages dozens of subprocess services with minimal resource impact |
| **Zero-Copy JSON** | `bytedance/sonic` for JIT-optimized serialization |
| **Message Pooling** | `sync.Pool` reduces garbage collection pressure |
| **Concurrent Task Execution** | go-taskflow enables parallel execution with up to 100 concurrent goroutines |
| **Configurable Buffering** | Tunable buffer capacity for high-volume scenarios |
| **Worker Pools** | Adjustable worker count for CPU-bound processing |
| **Message Batching** | Configurable batch size and flush interval for throughput optimization |
| **Linux-Native Optimizations** | CPU affinity binding, thread pinning, RT I/O priority for deterministic performance (>40% jitter reduction) |
| **Kernel Memory Hints** | madvise hints for HTML parsing reduce page faults by >30% on large files |

### Performance Monitoring

Available metrics include:
- Message throughput statistics
- Process response times
- System resource utilization
- Error rate tracking
- Per-process message statistics
- Optimizer metrics and performance counters

## Project Layout

- **cmd/** - Service implementations
  - v2broker/ - Process manager and RPC router with message optimization
  - v2access/ - REST gateway (subprocess)
  - v2local/ - Local data storage service (CVE/CWE/CAPEC/ATT&CK)
  - v2remote/ - Remote data fetching service
  - v2meta/ - Orchestration and job control (with Taskflow)
  - v2sysmon/ - System monitoring service
- **pkg/** - Shared packages
  - proc/subprocess - Subprocess framework (stdin/stdout RPC)
  - proc/message - Optimized message handling with pooling
  - cve/taskflow - Taskflow-based job executor with persistent state
  - cve - CVE domain types and helpers
  - cwe - CWE domain types and helpers
  - capec - CAPEC domain types and helpers
  - attack - ATT&CK framework domain types and helpers
  - common - Config and logging utilities
  - broker - Broker interfaces and types
  - jsonutil - JSON utility functions
  - rpc - RPC parameter types
  - testutils - Test utilities
- **tool/vconfig** - Configuration management tool with TUI interface
- **tests/** - Integration tests (pytest)
- **website/** - Next.js frontend application
- **assets/** - Data assets (CWE raw JSON, CAPEC XML/XSD, ATT&CK XLSX files)
- **.build/** - Build artifacts and packaged distribution

## Broker Interfaces and Internal Data Structures

The broker (microkernel) is organized into three primary layers:

### 1. Core Layer (`cmd/v2broker/core`)
Central management logic responsible for process supervision and message routing.
- **Broker**: The main struct orchestrating the system.
- **Process**: Represents a managed subprocess with its lifecycle state (`ProcessInfo`) and I/O pipes.
- **ProcessInfo**: Serializable struct containing PID, status (`running`, `exited`, `failed`), command, and start/end times.
- **Router Interface**: Defines how messages are routed between processes.

### 2. Transport Layer (`cmd/v2broker/transport`)
Handles low-level communication mechanics.
- **Transport Interface**: Defines the contract for IPC.
  - `Connect() error`: Establishes the connection.
  - `Close() error`: Terminates the connection.
  - `Send(msg *proc.Message) error`: Sends a structured message.
  - `Receive() (*proc.Message, error)`: Reads a structured message.
- **UDSTransport**: High-performance implementation using Unix Domain Sockets.

### 3. Performance Layer (`cmd/v2broker/perf`)
Decoupled optimization module for high-throughput message handling.
- **Optimizer**: Manages worker pools and message batching.
- **AdaptiveOptimizer**: Monitors system load and dynamically adjusts:
  - `BufferCapacity`: Channel size.
  - `WorkerCount`: Number of concurrent processors.
  - `BatchSize` & `FlushInterval`: For throughput tuning.

### Broker Architecture Diagram

```mermaid
graph TB
    subgraph "Broker Core"
        B[Broker]
        P[Process Map]
        R[Router]
    end

    subgraph "Transport Layer"
        TI{Transport Interface}
        UDS[UDSTransport]
    end

    subgraph "Performance Layer"
        OPT[Optimizer]
        AO[Adaptive Optimizer]
        SM[System Monitor]
    end

    subgraph "Subprocesses"
        S1[Service 1]
        S2[Service 2]
    end

    B --> P
    B --> R
    B --> OPT

    OPT --> AO
    AO --> SM

    P --> TI
    TI -.-> UDS

    UDS <==> S1
    UDS <==> S2

    style B fill:#4e8cff,stroke:#333,stroke-width:2px,color:#fff
    style OPT fill:#ff9900,stroke:#333,stroke-width:2px,color:#fff
    style TI fill:#66cc99,stroke:#333,stroke-width:2px,color:#fff
```


### Adaptive Optimization Algorithms
The broker implements intelligent adaptive tuning that responds to system and application loads:
- Dynamic worker count adjustment based on CPU utilization and queue depth
- Buffer capacity scaling based on throughput patterns
- Batch size optimization based on throughput and latency
- Flush interval tuning based on latency requirements
- Offer policy adjustment based on system load conditions

## Notes and Conventions

- All subprocesses must be started and managed by the broker; never run backend subprocesses directly in production or integration tests
- Subprocesses communicate exclusively via **UDS (Unix Domain Sockets)** - stdin/stdout FD pipe support has been removed
- Configuration (process list, logging) is controlled through `config.json`
- The authoritative RPC API specification for each subprocess can be found in `cmd/*/service.md`
- All inter-service communication flows through the broker to maintain architectural integrity
- Job sessions persist across service restarts; paused jobs remain paused, running jobs auto-resume
- Only one active job run (running or paused) is allowed at a time

## Where to Look Next

- [cmd/v2broker](cmd/v2broker) — Broker implementation and message routing
- [pkg/proc/subprocess](pkg/proc/subprocess) — Helper framework for subprocesses
- [pkg/cve/taskflow](pkg/cve/taskflow) — Taskflow-based job executor
- [cmd/v2access](cmd/v2access) — REST gateway and example of using the RPC client
- [cmd/v2meta](cmd/v2meta) — Job orchestration and session management
- [website/](website/) — Next.js frontend implementation
- [tests/](tests/) — Integration tests demonstrating usage patterns

## License

MIT

## Additional Documentation

### Containerized Development

For macOS users or when isolation is required, the project includes a containerized development environment:

- **runenv.sh**: Shell script that detects the operating system and runs the build environment in a container
- **Container Image**: Uses `assets/dev.Containerfile` to create the development environment
- **Go Module Cache**: Mounts the Go module cache inside the container for faster builds
- **Cross-platform**: Works on both macOS and Linux (optional on Linux with USE_CONTAINER=true)

### Testing Methodologies

The v2e project uses a hierarchical testing system that organizes tests by complexity and resource requirements. This allows developers to run fast unit tests during development while ensuring comprehensive testing in CI.

#### Test Levels (Cumulative)

Tests are organized into three levels controlled by the `V2E_TEST_LEVEL` environment variable with **cumulative behavior**:

- **V2E_TEST_LEVEL=1**: Runs only Level 1 tests
  - Pure logic, mock-based, no external dependencies, minimal database operations
  - Fast execution (milliseconds)
  - Safe for parallel execution
  - Default level if `V2E_TEST_LEVEL` is not set
  
- **V2E_TEST_LEVEL=2**: Runs Level 1 AND Level 2 tests
  - Includes Level 1 tests
  - Plus: Database (GORM) operations with transaction isolation
  - Uses automatic transaction rollback for test isolation
  - Tests run in parallel within transactions
  - No persistent side effects
  
- **V2E_TEST_LEVEL=3**: Runs Level 1, Level 2, AND Level 3 tests (all tests)
  - Includes Level 1 and Level 2 tests
  - Plus: External APIs, E2E, and heavy integration tests
  - May require external services
  - Longer execution time
  - Used in CI for comprehensive validation

#### Writing Tests with testutils.Run()

All tests use the unified `testutils.Run()` wrapper for automatic parallelization, level filtering, and optional transaction isolation:

```go
import "github.com/cyw0ng95/v2e/pkg/testutils"

// Level 1: Pure logic test (no database)
func TestBusinessLogic(t *testing.T) {
    testutils.Run(t, testutils.Level1, "BasicCalculation", nil, func(t *testing.T, tx *gorm.DB) {
        // Test implementation (tx will be nil)
        result := Calculate(10, 20)
        if result != 30 {
            t.Errorf("Expected 30, got %d", result)
        }
    })
}

// Level 2: Database test with automatic transaction isolation
func TestDatabaseOperation(t *testing.T) {
    db := setupTestDB(t)
    
    testutils.Run(t, testutils.Level2, "InsertRecord", db, func(t *testing.T, tx *gorm.DB) {
        // All database operations use tx (transaction)
        // Automatic rollback ensures no side effects
        record := &MyRecord{Name: "test"}
        if err := tx.Create(record).Error; err != nil {
            t.Fatalf("Failed to create: %v", err)
        }
    })
}
```

#### Running Tests at Different Levels

```bash
# Run Level 1 tests only (default, fastest)
./build.sh -t

# Run Level 1 and Level 2 tests (cumulative)
V2E_TEST_LEVEL=2 ./build.sh -t

# Run all tests - Level 1, 2, and 3 (cumulative)
V2E_TEST_LEVEL=3 ./build.sh -t

# Run specific test at a specific level
V2E_TEST_LEVEL=2 go test -v ./pkg/notes/...
```

In CI, V2E_TEST_LEVEL=3 runs all tests in a single job for comprehensive coverage.

- **Unit Tests**: Standard Go unit tests with coverage reporting
  ```bash
  # Run all unit tests (excludes fuzz tests)
  ./build.sh -t
  
  # Run tests for specific packages
  go test -run='^Test' ./pkg/cve/...
  go test -run='^Test' ./cmd/v2broker/...
  
  # Run tests with coverage
  go test -run='^Test' -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
  ```

- **Fuzz Tests**: Fuzz testing for key interfaces to discover edge cases
  ```bash
  # Run all fuzz tests (5 seconds per target)
  ./build.sh -f
  
  # Run specific fuzz test with custom duration
  go test -fuzz=FuzzValidateCVEID -fuzztime=30s ./pkg/cve/remote
  go test -fuzz=FuzzSaveCVE -fuzztime=1m ./pkg/cve/local
  
  # Note: Fuzz tests automatically discover edge cases and save failing inputs to testdata/fuzz/
  ```

- **Performance Benchmarks**: Comprehensive benchmarking with statistical analysis
  ```bash
  # Run all benchmarks with full reporting
  ./build.sh -m
  
  # Run specific benchmarks
  go test -bench=BenchmarkSendMessage -benchmem ./pkg/proc/subprocess
  go test -bench=BenchmarkGetCVE -benchmem ./pkg/cve/local
  
  # Run benchmarks with memory allocation profiling
  go test -bench=. -benchmem -memprofile=mem.out ./cmd/v2broker/perf
  go tool pprof -http=:8080 mem.out
  
  # Compare benchmark results
  go test -bench=. -count=5 ./pkg/cve/local > old.txt
  # ... make changes ...
  go test -bench=. -count=5 ./pkg/cve/local > new.txt
  benchstat old.txt new.txt
  ```

- **Integration Tests**: Pytest-based integration tests in the `tests/` directory
  ```bash
  # Run integration tests (requires broker to be running)
  ./build.sh -r  # Start full system first
  # In another terminal:
  pytest tests/
  ```
