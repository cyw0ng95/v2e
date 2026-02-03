# v2e

A sophisticated Go-based system that demonstrates a broker-first architecture for orchestrating multiple subprocess services that communicate via RPC messages over stdin/stdout. The system provides a comprehensive CVE (Common Vulnerabilities and Exposures) management platform with integrated CWE (Common Weakness Enumeration), CAPEC (Common Attack Pattern Enumeration and Classification), ATT&CK (Adversarial Tactics, Techniques, and Common Knowledge) framework, and SSG (SCAP Security Guide) data handling.

## Executive Summary

The v2e project implements a broker-first architecture where `cmd/broker` serves as the central process manager that spawns, monitors, and manages all subprocess services. This design enforces a strict communication pattern where all inter-service communication flows through the broker, preventing direct subprocess-to-subprocess interaction. The architecture ensures clean separation of concerns while maintaining robust message routing and process lifecycle management.

Key architectural principles:
- **Centralized Process Management**: The broker is the sole orchestrator of all subprocess services
- **Enforced Communication Pattern**: All inter-service communication occurs through broker routing
- **RPC-Based Messaging**: Services communicate via structured JSON RPC messages over stdin/stdout
- **Comprehensive Data Handling**: Integrated CVE, CWE, CAPEC, ATT&CK, and SSG data management
- **Frontend Integration**: A Next.js-based web application provides user interface access
- **Performance Monitoring**: Built-in metrics collection and system monitoring capabilities
- **Message Optimization**: Asynchronous message routing with configurable buffering and batching
- **Adaptive Optimization**: Dynamic performance tuning based on workload with adaptive algorithms
- **Enhanced Configuration**: Advanced configuration management via vconfig tool with TUI interface
- **Cross-Platform Support**: Containerized development environment for macOS with Linux support

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

## Component Breakdown

### Core Services

- **Broker Service** ([cmd/broker](cmd/broker)): The central orchestrator responsible for:
  - Spawning and managing all subprocess services with robust supervision and restart policies
  - Routing RPC messages via a high-performance Unix Domain Sockets (UDS) transport layer
  - Utilizing `bytedance/sonic` for zero-copy JSON serialization/deserialization
  - Implementing an adaptive traffic optimizer with configurable batching, buffering, and backpressure
  - Maintaining process lifecycle, health checks, and zombie process reaping
  - Tracking comprehensive real-time message statistics and performance metrics
  - Supporting advanced logging with dual output (console + file) and configurable log levels
  - Providing dynamic configuration of performance parameters via adaptive optimization algorithms

- **Access Service** ([cmd/access](cmd/access)): The REST gateway that:
  - Serves as the primary interface for the Next.js frontend
  - Exposes `/restful/rpc` endpoint for RPC forwarding
  - Translates HTTP requests to RPC calls and responses back
  - Provides health checks and basic service discovery

- **Meta Service** ([cmd/meta](cmd/meta)): The orchestration layer that:
  - Manages job scheduling and execution using go-taskflow
  - Coordinates complex multi-step operations
  - Handles session management and state persistence
  - Provides workflow control mechanisms
  - Orchestrates CVE/CWE data fetching jobs with persistent state management
  - Performs automatic CWE and CAPEC imports at startup
  - Provides memory card management (delegates to local service):
    - RPCCreateMemoryCard, RPCGetMemoryCard, RPCUpdateMemoryCard
    - RPCDeleteMemoryCard, RPCListMemoryCards

- **Local Service** ([cmd/local](cmd/local)): The data persistence layer that:
  - Manages local SQLite databases for CVE, CWE, CAPEC, and ATT&CK data
  - Provides CRUD operations for vulnerability information
  - Handles data indexing and querying
  - Implements caching mechanisms for improved performance
  - Imports ATT&CK data from XLSX files and provides access to techniques, tactics, mitigations, software, and groups
  - Supports CAPEC XML schema validation and catalog metadata retrieval
  - Manages SSG (SCAP Security Guide) data with file-based storage and XML parsing
  - Offers CWE view management with storage and retrieval capabilities
  - Provides memory card storage for bookmark/knowledge management:
    - RPCCreateMemoryCard, RPCGetMemoryCard, RPCUpdateMemoryCard
    - RPCDeleteMemoryCard, RPCListMemoryCards
    - Supports TipTap JSON content, classification fields, and metadata

- **Remote Service** ([cmd/remote](cmd/remote)): The data acquisition layer that:
  - Fetches vulnerability data from external APIs (NVD, etc.)
  - Implements rate limiting and retry mechanisms
  - Handles data transformation and normalization
  - Manages API credentials and authentication

- **SysMon Service** ([cmd/sysmon](cmd/sysmon)): The system monitoring layer that:
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
- **Rich Component Architecture**: Tabbed interface supporting CVE, CWE, CAPEC, ATT&CK, SSG, and system monitoring data
- **Real-time Updates**: Session control and live metrics display
- **Responsive Design**: Adaptable interface for various screen sizes and devices
- **Modern Tech Stack**: Uses Next.js 16+, React 19+, with TypeScript, Tailwind CSS, and Radix UI components
- **Data Visualization**: Recharts for performance metrics and data visualization

The frontend includes dedicated sections for:
- CVE Database browsing and management
- CWE Database and view management
- CAPEC data visualization
- ATT&CK framework integration
- SSG (SCAP Security Guide) profiles and rules
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
  - broker/ - Process manager and RPC router with message optimization
  - access/ - REST gateway (subprocess)
  - local/ - Local data storage service (CVE/CWE/CAPEC/ATT&CK)
  - remote/ - Remote data fetching service
  - meta/ - Orchestration and job control (with Taskflow)
  - sysmon/ - System monitoring service
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

### 1. Core Layer (`cmd/broker/core`)
Central management logic responsible for process supervision and message routing.
- **Broker**: The main struct orchestrating the system.
- **Process**: Represents a managed subprocess with its lifecycle state (`ProcessInfo`) and I/O pipes.
- **ProcessInfo**: Serializable struct containing PID, status (`running`, `exited`, `failed`), command, and start/end times.
- **Router Interface**: Defines how messages are routed between processes.

### 2. Transport Layer (`cmd/broker/transport`)
Handles low-level communication mechanics.
- **Transport Interface**: Defines the contract for IPC.
  - `Connect() error`: Establishes the connection.
  - `Close() error`: Terminates the connection.
  - `Send(msg *proc.Message) error`: Sends a structured message.
  - `Receive() (*proc.Message, error)`: Reads a structured message.
- **UDSTransport**: High-performance implementation using Unix Domain Sockets.

### 3. Performance Layer (`cmd/broker/perf`)
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

- [cmd/broker](cmd/broker) — Broker implementation and message routing
- [pkg/proc/subprocess](pkg/proc/subprocess) — Helper framework for subprocesses
- [pkg/cve/taskflow](pkg/cve/taskflow) — Taskflow-based job executor
- [cmd/access](cmd/access) — REST gateway and example of using the RPC client
- [cmd/meta](cmd/meta) — Job orchestration and session management
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

- **Unit Tests**: Standard Go unit tests with coverage reporting
- **Fuzz Tests**: Fuzz testing for key interfaces to discover edge cases
- **Performance Benchmarks**: Comprehensive benchmarking with statistical analysis
- **Integration Tests**: Pytest-based integration tests in the `tests/` directory
