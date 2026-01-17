# v2e

[![CI Tests](https://github.com/cyw0ng95/v2e/actions/workflows/test.yml/badge.svg)](https://github.com/cyw0ng95/v2e/actions/workflows/test.yml)

A Go-based project demonstrating a multi-command structure with CVE (Common Vulnerabilities and Exposures) data fetching capabilities.

## Architecture Principles

**Important: Broker-First Architecture**

The broker (`cmd/broker`) is a standalone process that boots and manages all subprocesses in the system. All other services (including `access`, `remote`, `local`, `meta`, etc.) are subprocesses managed by the broker.

**Key Rules:**
- The broker is the only entry point for process management
- Subprocesses must NOT embed broker logic or create their own broker instances
- The `access` service is a subprocess like any other - it provides a REST API but does not manage processes itself
- All process spawning, lifecycle management, and inter-process communication goes through the broker
- Subprocesses communicate with each other only through broker-mediated RPC messages

This architecture ensures clean separation of concerns and prevents circular dependencies.

## Project Structure

This project contains multiple commands:

- `cmd/access` - RESTful API gateway service (the central entry point for external requests)
- `cmd/broker` - RPC service for managing subprocesses and process lifecycle
- `cmd/remote` - RPC service for fetching CVE data from NVD API
- `cmd/local` - RPC service for storing and retrieving CVE data from local database
- `cmd/meta` - Backend RPC service that orchestrates CVE fetching and storage operations

And packages:

- `pkg/common` - Common utilities and configuration
- `pkg/repo` - Repository layer for external data sources (NVD CVE API)
- `pkg/proc` - Process broker for managing subprocesses and inter-process communication
- `pkg/proc/subprocess` - Common subprocess framework for message-driven subprocesses
- `pkg/cve` - CVE data types shared across packages
- `pkg/cve/remote` - Remote CVE fetching from NVD API
- `pkg/cve/local` - Local CVE storage with SQLite database
- `pkg/cve/session` - Session management with bbolt K-V database
- `pkg/cve/job` - Job controller for continuous CVE fetch/store operations

## CVE Meta Job Control

The `meta` service provides job control capabilities for continuous CVE data fetching and storing operations. This allows you to start long-running jobs that continuously fetch CVE data from the NVD API and store it in the local database.

### Features

- **Session Management**: Create and manage job sessions with unique IDs
- **Single Session Enforcement**: Only one job session can run at a time to prevent conflicts
- **Persistent State**: Session state is stored in a bbolt K-V database and survives service restarts
- **Job Control**: Start, stop, pause, and resume continuous CVE fetching operations
- **Progress Tracking**: Monitor fetched, stored, and error counts during job execution

### RPC APIs

#### RPCStartSession
Starts a new job session for continuous CVE fetching and storing.

**Parameters:**
- `session_id` (string, required): Unique identifier for the session
- `start_index` (int, optional): Starting index for NVD API pagination (default: 0)
- `results_per_batch` (int, optional): Number of results per batch (default: 100)

**Response:**
```json
{
  "success": true,
  "session_id": "my-session",
  "state": "running",
  "created_at": "2026-01-13T02:00:00Z"
}
```

#### RPCStopSession
Stops the current session and cleans up resources.

**Response:**
```json
{
  "success": true,
  "session_id": "my-session",
  "fetched_count": 150,
  "stored_count": 145,
  "error_count": 5
}
```

#### RPCGetSessionStatus
Gets the current session status and progress.

**Response:**
```json
{
  "has_session": true,
  "session_id": "my-session",
  "state": "running",
  "start_index": 0,
  "results_per_batch": 100,
  "created_at": "2026-01-13T02:00:00Z",
  "updated_at": "2026-01-13T02:05:00Z",
  "fetched_count": 150,
  "stored_count": 145,
  "error_count": 5
}
```

#### RPCPauseJob
Pauses the running job without deleting the session.

**Response:**
```json
{
  "success": true,
  "state": "paused"
}
```

#### RPCResumeJob
Resumes a paused job from where it left off.

**Response:**
```json
{
  "success": true,
  "state": "running"
}
```

### Example Usage

Via the REST API (through the access service):

```bash
# Start a new session
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCStartSession",
    "target": "meta",
    "params": {
      "session_id": "bulk-fetch-2026",
      "start_index": 0,
      "results_per_batch": 100
    }
  }'

# Check session status
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCGetSessionStatus",
    "target": "meta",
    "params": {}
  }'

# Pause the job
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCPauseJob",
    "target": "meta",
    "params": {}
  }'

# Resume the job
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCResumeJob",
    "target": "meta",
    "params": {}
  }'

# Stop the session
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCStopSession",
    "target": "meta",
    "params": {}
  }'
```

### Session State Management

Job sessions go through the following states:

- **idle**: Session created but job not yet started
- **running**: Job is actively fetching and storing CVEs
- **paused**: Job temporarily paused, can be resumed
- **stopped**: Session stopped and deleted

The session state is persisted in `session.db` (bbolt K-V database) and includes:
- Session ID and configuration (start index, batch size)
- Progress counters (fetched, stored, error counts)
- Timestamps (created, last updated)

## Architecture and Security

### Deployment Model

This project follows a **broker-first architecture** with strict subprocess isolation:

1. **Broker is the entry point**: Users run `broker` with `config.json` as the first argument
2. **Broker spawns all subprocesses**: Includes access (REST API), remote, local, meta
3. **External users access via REST**: HTTP requests to access service (http://localhost:8080/restful/*)
4. **Access forwards to broker**: Converts REST → RPC messages sent to broker (future implementation)
5. **Broker routes to services**: Routes RPC messages to appropriate backend subprocess
6. **No direct subprocess access**: External users cannot directly communicate with backend services

**Deployment Flow:**

```
User runs:    broker config.json
              ↓
Broker spawns: access, remote, local, meta
              ↓
External user → REST API (access) → Broker → Backend Services
                                       ↓
                         Message Routing (RPCInvoke)
                                       ↓
              remote ←→ Broker ←→ local
                             ↓
                          meta
```

This architecture provides:
- **Process isolation**: Services cannot directly access each other
- **Message routing**: Broker controls all inter-process communication via RPCInvoke
- **Request-response correlation**: Broker tracks requests and matches responses using correlation IDs
- **Security**: Single point of control for spawning and managing processes
- **Monitoring**: Centralized message statistics and process management
- **No direct RPC**: External users cannot bypass access service to reach backends
- **Cross-service calls**: Services can invoke each other through the broker's message routing

### Security Principles

#### RPC-Only Communication (for Subprocess Services)

All inter-process communication between subprocess services MUST use RPC messages in JSON format:

```json
{
  "type": "request|response|event|error",
  "id": "unique-message-id",
  "payload": { ... }
}
```

Messages are exchanged via stdin/stdout pipes between subprocess services. Subprocess services never use:
- Network sockets for inter-process communication
- Files or shared memory for commands
- Direct subprocess-to-subprocess communication

#### Broker-Mediated Message Routing

The broker acts as a secure message router:

```
User/Config → Broker → [Subprocess A]
                    ↓
                    → [Subprocess B]
                    ↓
                    → [Subprocess C]
```

All messages flow through the broker, which:
- Routes messages to the correct subprocess
- Tracks message statistics
- Manages subprocess lifecycle
- Prevents unauthorized inter-process communication

#### Subprocess Isolation

Each subprocess service:

- **ONLY** reads from stdin (messages from broker)
- **ONLY** writes to stdout (messages to broker)
- **DOES NOT** accept:
  - Command-line arguments (except those passed by broker)
  - Direct network connections for control
  - File-based input for commands
  - User input from terminal

This ensures that all service control flows through the broker, preventing:
- Unauthorized access to services
- Bypassing of broker security controls
- Untracked communication between services

### Example: Secure Service Implementation

All services in `cmd/` follow this pattern:

```go
// Service only accepts PROCESS_ID from environment
processID := os.Getenv("PROCESS_ID")

// Create subprocess instance (reads from stdin, writes to stdout)
sp := subprocess.New(processID)

// Register RPC handlers for all functionality
sp.RegisterHandler("RPCMyFunction", handler)

// Run subprocess (blocks on stdin, no other input accepted)
sp.Run()
```

## Prerequisites

- Go 1.24 or later

## Building

To build all commands:

```bash
go build ./cmd/access
go build ./cmd/broker
go build ./cmd/remote
go build ./cmd/local
go build ./cmd/meta
```

Or build a specific command:

```bash
go build -o bin/access ./cmd/access
go build -o bin/broker ./cmd/broker
go build -o bin/remote ./cmd/remote
go build -o bin/local ./cmd/local
go build -o bin/meta ./cmd/meta
```

## Running

### Deployment with Broker (Recommended)

The recommended way to deploy and run services is through the broker with a configuration file:

```bash
# Build all binaries first
./build.sh -p

# Run broker with config file (automatically starts configured services)
./broker config.json

# The broker will:
# 1. Load config.json as specified
# 2. Spawn all configured subprocess services
# 3. Route RPC messages between services
# 4. Monitor and restart services as needed
```

The `config.json` file defines which services to start:

```json
{
  "server": {
    "address": "0.0.0.0:8080"
  },
  "broker": {
    "logs_dir": "./logs",
    "processes": [
      {
        "id": "access",
        "command": "./access",
        "args": [],
        "rpc": false,
        "restart": true,
        "max_restarts": -1
      },
      {
        "id": "remote",
        "command": "./remote",
        "args": [],
        "rpc": true,
        "restart": true,
        "max_restarts": -1
      },
      {
        "id": "local",
        "command": "./local",
        "args": [],
        "rpc": true,
        "restart": true,
        "max_restarts": -1
      },
      {
        "id": "meta",
        "command": "./meta",
        "args": [],
        "rpc": true,
        "restart": true,
        "max_restarts": -1
      }
    ]
  },
  "logging": {
    "level": "info",
    "dir": "./logs"
  }
}
```

**Configuration Options:**
- `logs_dir` - Directory for log files
- `processes` - Array of services to automatically spawn
  - `id` - Unique identifier for the service
  - `command` - Path to binary executable
  - `args` - Command arguments
  - `rpc` - Enable RPC communication (stdin/stdout pipes)
  - `restart` - Automatically restart on exit
  - `max_restarts` - Maximum restart attempts (-1 for unlimited)

**Security Note:** Services listed in the config are spawned by the broker and can only communicate via RPC messages routed through the broker. This ensures process isolation and prevents unauthorized access.

### Access (RESTful API Service)

The Access service is the external entry point for the v2e system. It provides a RESTful API server that external users interact with. The access service is spawned by the broker as a subprocess and communicates with the broker via RPC to fulfill requests.

**Deployment Model:**

In the correct v2e architecture:

1. **Broker is the entry point**: Run broker with `config.json` as the first argument
2. **Broker spawns all subprocesses**: Including access, remote, local, meta
3. **Access provides REST API**: External interface for the system
4. **Access forwards to broker**: Converts REST requests to RPC messages sent to broker
5. **Broker routes messages**: Routes RPC requests to appropriate backend services

**Key Features:**
- Single REST interface for external system interactions
- RESTful API schema under `/restful/` prefix
- Runs as a subprocess spawned by the broker
- Configurable listen address (default: `0.0.0.0:8080`)

**Running the System:**

```bash
# Run broker with configuration
go run ./cmd/broker config.json

# The broker will:
# 1. Load configuration from config.json
# 2. Spawn all configured subprocesses including access
# 3. Access service starts REST API server
# 4. Backend CVE services are spawned and ready
```

**Configuration:**

The access service is configured in `config.json`:

```json
{
  "server": {
    "address": "0.0.0.0:8080"
  },
  "broker": {
    "logs_dir": "logs",
    "processes": [
      {
        "id": "access",
        "command": "go",
        "args": ["run", "./cmd/access"],
        "rpc": false,
        "restart": true
      },
      {
        "id": "remote",
        "command": "go",
        "args": ["run", "./cmd/remote"],
        "rpc": true,
        "restart": true
      }
    ]
  }
}
```

**Available RESTful Endpoints:**

*Currently Implemented:*
- `GET /restful/health` - Health check endpoint
  ```bash
  curl http://localhost:8080/restful/health
  # Response: {"status":"ok"}
  ```

*Future Implementation (requires RPC forwarding):*

The following endpoints require implementing RPC message correlation and forwarding between access service and broker:

- `GET /restful/stats` - Get broker message statistics (via RPC to broker)
- `GET /restful/processes` - List all managed processes (via RPC to broker)
- `GET /restful/processes/:id` - Get process details (via RPC to broker)
- `POST /restful/rpc/:process_id/:endpoint` - Forward RPC to backend services

**Architecture:**

The Access service acts as a REST-to-RPC gateway:

```
External Clients → Access (REST API) → Broker (RPC Router)
                                         ↓
                             ┌──────────┼──────────┐
                             ↓          ↓          ↓
                         local  remote  meta
```

All external requests are handled by the Access service, which:
1. Receives HTTP requests on RESTful endpoints
2. Translates them to RPC messages (future implementation)
3. Sends RPC messages to broker via subprocess communication
4. Broker forwards messages to the appropriate backend service
5. Returns responses from backend services as JSON (via broker)

**Security:**

The Access service provides a controlled interface to the system:
- Single point of entry for external requests
- All backend services are isolated and only accessible via broker
- Access service is spawned by broker (not standalone)
- No direct RPC access to backend services from external clients
- All communication routes through broker's secure message passing

**Note:** Full RPC forwarding implementation is future work. Currently, only health check endpoint is implemented.

### CVE Remote (RPC Service)

**Production Deployment:** This service should be spawned by the broker via config.json (see "Deployment with Broker" section above).

The CVE Remote service provides RPC interfaces for fetching CVE data from the NVD API.

**Available RPC Interfaces:**
- `RPCGetCVECnt` - Returns the total count of CVEs in the NVD database
- `RPCGetCVEByID` - Fetches a specific CVE by its ID from the NVD API

**Environment Variables:**
- `NVD_API_KEY` - Optional NVD API key for higher rate limits

**Accessing the Service:**

RPC services should be accessed via the Access service's RESTful API, not directly via stdin/stdout:

```bash
# Get CVE count from NVD
curl -X POST http://localhost:8080/restful/rpc/remote/RPCGetCVECnt \
  -H "Content-Type: application/json" \
  -d '{}'

# Fetch a specific CVE by ID
curl -X POST http://localhost:8080/restful/rpc/remote/RPCGetCVEByID \
  -H "Content-Type: application/json" \
  -d '{"cve_id":"CVE-2021-44228"}'
```

**Security Note:** This service only accepts RPC messages routed through the broker. Direct invocation is not supported - all requests must go through the Access service's RESTful API.

### CVE Local (RPC Service)

**Production Deployment:** This service should be spawned by the broker via config.json (see "Deployment with Broker" section above).

The CVE Local service provides RPC interfaces for storing and retrieving CVE data from a local SQLite database.

**Available RPC Interfaces:**
- `RPCIsCVEStoredByID` - Checks if a CVE exists in the local database
- `RPCSaveCVEByID` - Saves a CVE to the local database

**Environment Variables:**
- `CVE_DB_PATH` - Path to the SQLite database file (default: `cve.db`)

**Accessing the Service:**

RPC services should be accessed via the Access service's RESTful API, not directly via stdin/stdout:

```bash
# Example: Access local service via RESTful API
curl -X POST http://localhost:8080/restful/rpc/local/RPCIsCVEStoredByID \
  -H "Content-Type: application/json" \
  -d '{"cve_id":"CVE-2021-44228"}'
```

**Security Note:** This service only accepts RPC messages routed through the broker. Direct invocation is not supported - all requests must go through the Access service's RESTful API.

### CVE Meta Service

**Production Deployment:** This service should be spawned by the broker via config.json (see "Deployment with Broker" section above).

The CVE Meta service is a backend RPC service that orchestrates CVE fetching and storage operations. It acts as a coordinator between local and remote services.

**Current RPC Interfaces:**
- `RPCGetCVE` - Retrieves CVE data (currently returns stub data, will orchestrate local and remote in the future)

**Planned RPC Interfaces:**
- `RPCFetchAndStoreCVE` - Will fetch a CVE from NVD (if not already stored locally) and save it to the database
- `RPCBatchFetchCVEs` - Will fetch and store multiple CVEs in batch mode
- `RPCGetRemoteCVECount` - Will return the total count of CVEs in the NVD database

**Environment Variables:**
- `CVE_DB_PATH` - Path to the SQLite database file (default: `cve.db`)

**Accessing the Service:**

The meta service is accessed via the Access service's RESTful API with the `target` parameter:

```bash
# Get CVE data via meta service
curl -X POST http://localhost:8080/restful/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "method": "RPCGetCVE",
    "target": "meta",
    "params": {"cve_id": "CVE-2021-44228"}
  }'

# Response format:
# {
#   "retcode": 0,
#   "message": "success",
#   "payload": {
#     "id": "CVE-2021-44228",
#     "descriptions": [...]
#   }
# }
```

**Testing:**

The meta service has comprehensive integration tests that verify functionality using only the RESTful API. All tests follow the broker-first architecture:

```bash
# Run meta integration tests
pytest tests/test_integration.py::TestCVEMetaService -v
```

Test coverage includes:
- Valid CVE ID requests
- Missing required parameters
- Empty parameter validation
- Multiple sequential requests
- Error handling and routing

**Security Note:** This service only accepts RPC messages routed through the broker. Direct invocation is not supported - all requests must go through the Access service's RESTful API with the appropriate `target` parameter.

## Development

### Subprocess Framework

The `pkg/proc/subprocess` package provides a framework for building message-driven subprocesses:

```go
import (
    "context"
    "github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Create a new subprocess
sp := subprocess.New("my-worker")

// Register message handlers
sp.RegisterHandler("request", func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
    // Parse request payload
    var req map[string]interface{}
    if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
        return nil, err
    }
    
    // Process the request and create a response
    response := &subprocess.Message{
        Type: subprocess.MessageTypeResponse,
        ID:   msg.ID,
    }
    
    return response, nil
})

// Run the subprocess (blocks until stopped)
if err := sp.Run(); err != nil {
    log.Fatal(err)
}
```

Key features:
- **Message-driven architecture**: Communication via JSON messages over stdin/stdout
- **No broker dependencies**: Subprocess code is completely independent of the broker
- **Handler-based routing**: Register handlers for different message types or IDs
- **Graceful shutdown**: Built-in signal handling and cleanup
- **Type-safe messaging**: Structured message types (Request, Response, Event, Error)
- **Security by design**: Only accepts input via stdin, no other input channels

The subprocess framework allows you to build worker processes that:
1. Are controlled by the broker through message passing
2. Can be spawned and monitored by the broker
3. Have a clear lifecycle with proper initialization and shutdown
4. Focus on business logic without worrying about process management
5. **Are isolated from external input** - only communicate via RPC messages through the broker

**Security Best Practices:**

When building a new subprocess service:
- ✅ **DO** use `subprocess.New()` and `sp.Run()` for all input handling
- ✅ **DO** register all functionality as RPC handlers
- ✅ **DO** read only from stdin (handled by the framework)
- ✅ **DO** write only to stdout (via message sending methods)
- ❌ **DO NOT** accept command-line flags for control operations
- ❌ **DO NOT** open network sockets for receiving commands
- ❌ **DO NOT** read from files or other external sources for commands
- ❌ **DO NOT** implement alternative input mechanisms

This ensures all subprocess control flows through the broker's secure message routing.

### CVE Remote Fetcher

The `pkg/cve/remote` package provides a CVE fetcher that integrates with the NVD API v2.0:

```go
import "github.com/cyw0ng95/v2e/pkg/cve/remote"

// Create a new CVE fetcher (optionally with API key for higher rate limits)
fetcher := remote.NewFetcher("")

// Fetch a specific CVE by ID
cveData, err := fetcher.FetchCVEByID("CVE-2021-44228")

// Fetch multiple CVEs with pagination
cveList, err := fetcher.FetchCVEs(0, 10)
```

For production use with higher rate limits, obtain an API key from [NVD](https://nvd.nist.gov/developers/request-an-api-key) and pass it to `NewFetcher()`.

### CVE Local Storage

The `pkg/cve/local` package includes a GORM-based ORM engine for storing CVE data in a local SQLite database:

```go
import (
    "github.com/cyw0ng95/v2e/pkg/cve"
    "github.com/cyw0ng95/v2e/pkg/cve/local"
)

// Create or open the CVE database
db, err := local.NewDB("cve.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Save a CVE to the database
cveItem := &cve.CVEItem{
    ID:           "CVE-2021-44228",
    SourceID:     "nvd@nist.gov",
    Published:    cve.NewNVDTime(time.Now()),
    LastModified: cve.NewNVDTime(time.Now()),
    VulnStatus:   "Analyzed",
    Descriptions: []cve.Description{
        {Lang: "en", Value: "Apache Log4j vulnerability"},
    },
}
err = db.SaveCVE(cveItem)

// Retrieve a CVE by ID
retrieved, err := db.GetCVE("CVE-2021-44228")

// List CVEs with pagination
cves, err := db.ListCVEs(0, 10) // offset=0, limit=10

// Get total count
count, err := db.Count()
```

The database file `cve.db` is created in the project root directory and is excluded from version control via `.gitignore`.

To create a sample database with CVE data, run the integration test:

```bash
go test ./pkg/cve/local -v -run TestCreateCVEDatabase
```

This will create `cve.db` in the project root with sample CVE records that you can inspect or download.

### Working with Remote and Local Together

You can combine the remote fetcher with local storage to build a CVE database:

```go
import (
    "github.com/cyw0ng95/v2e/pkg/cve/remote"
    "github.com/cyw0ng95/v2e/pkg/cve/local"
)

// Initialize remote fetcher and local database
fetcher := remote.NewFetcher("")
db, err := local.NewDB("cve.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Fetch CVE from NVD and save to local database
response, err := fetcher.FetchCVEByID("CVE-2021-44228")
if err != nil {
    log.Fatal(err)
}

if len(response.Vulnerabilities) > 0 {
    cveItem := response.Vulnerabilities[0].CVE
    err = db.SaveCVE(&cveItem)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Development

### Logging

The project includes a structured logging module in `pkg/common`:

```go
import "github.com/cyw0ng95/v2e/pkg/common"

// Use the default logger
common.Info("Server starting on port %d", 8080)
common.Debug("Debug information")
common.Warn("Warning message")
common.Error("Error occurred: %v", err)

// Set log level
common.SetLevel(common.DebugLevel) // DebugLevel, InfoLevel, WarnLevel, ErrorLevel

// Create a custom logger
logger := common.NewLogger(os.Stdout, "", common.InfoLevel)
logger.Info("Custom logger message")
```

### Process Broker

The `pkg/proc` package provides a process broker for managing subprocesses and inter-process communication:

```go
import "github.com/cyw0ng95/v2e/pkg/proc"

// Create a new broker
broker := proc.NewBroker()
defer broker.Shutdown()

// Spawn a subprocess
info, err := broker.Spawn("my-process", "echo", "hello", "world")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Started process %s with PID %d\n", info.ID, info.PID)

// Get process information
procInfo, err := broker.GetProcess("my-process")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Process status: %s\n", procInfo.Status)

// List all processes
processes := broker.ListProcesses()
for _, p := range processes {
    fmt.Printf("Process %s: PID=%d Status=%s\n", p.ID, p.PID, p.Status)
}

// Kill a process
err = broker.Kill("my-process")
if err != nil {
    log.Fatal(err)
}
```

#### Message Statistics

The broker tracks message statistics to help monitor message flow:

```go
import "github.com/cyw0ng95/v2e/pkg/proc"

// Create a new broker
broker := proc.NewBroker()
defer broker.Shutdown()

// Send some messages
reqMsg, _ := proc.NewRequestMessage("req-1", map[string]string{"action": "test"})
broker.SendMessage(reqMsg)

respMsg, _ := proc.NewResponseMessage("resp-1", map[string]string{"result": "ok"})
broker.SendMessage(respMsg)

// Get total message count
count := broker.GetMessageCount()
fmt.Printf("Total messages processed: %d\n", count)

// Get detailed statistics
stats := broker.GetMessageStats()
fmt.Printf("Sent: %d, Received: %d\n", stats.TotalSent, stats.TotalReceived)
fmt.Printf("Requests: %d, Responses: %d, Events: %d, Errors: %d\n",
    stats.RequestCount, stats.ResponseCount, stats.EventCount, stats.ErrorCount)
fmt.Printf("First message: %v\n", stats.FirstMessageTime)
fmt.Printf("Last message: %v\n", stats.LastMessageTime)
```

Key statistics features:
- **GetMessageCount()**: Returns the total number of messages processed (sent + received)
- **GetMessageStats()**: Returns detailed statistics including:
  - Total messages sent and received
  - Message counts by type (request, response, event, error)
  - Timestamp of first and last message
- **Thread-safe**: All statistics methods are safe for concurrent access
- **RPC Access**: Statistics can also be accessed remotely via the broker's RPC service using `RPCGetMessageCount` and `RPCGetMessageStats` handlers (see [Broker](#broker-rpc-service) section)

#### RPC Communication

The broker supports RPC-style communication with subprocesses via stdin/stdout pipes:

```go
import "github.com/cyw0ng95/v2e/pkg/proc"

// Create a new broker
broker := proc.NewBroker()
defer broker.Shutdown()

// Spawn a subprocess with RPC support (stdin/stdout pipes)
info, err := broker.SpawnRPC("remote", "go", "run", "./cmd/remote")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Started RPC process %s with PID %d\n", info.ID, info.PID)

// Send a message to the subprocess
req, _ := proc.NewRequestMessage("RPCGetCVEByID", map[string]string{
    "cve_id": "CVE-2021-44228",
})
err = broker.SendToProcess("remote", req)
if err != nil {
    log.Fatal(err)
}

// Receive the response from the subprocess
ctx := context.Background()
msg, err := broker.ReceiveMessage(ctx)
if err != nil {
    log.Fatal(err)
}

// Process the response
if msg.Type == proc.MessageTypeResponse {
    var response map[string]interface{}
    msg.UnmarshalPayload(&response)
    fmt.Printf("Received response: %v\n", response)
}
```

Key RPC features:
- **SpawnRPC**: Spawns a subprocess with stdin/stdout pipes for message passing
- **SendToProcess**: Sends a message to a specific subprocess via its stdin
- **Automatic message routing**: The broker reads messages from subprocess stdout and forwards them to the message channel
- **Broker-mediated communication**: All messages are sent through the broker, which routes them to target services

#### Message Passing

The broker supports message passing between processes:

```go
// Create and send a request message
req, _ := proc.NewRequestMessage("req-1", map[string]string{
    "action": "process_data",
    "data":   "example",
})
broker.SendMessage(req)

// Receive messages (blocking)
ctx := context.Background()
msg, err := broker.ReceiveMessage(ctx)
if err != nil {
    log.Fatal(err)
}

// Unmarshal the message payload
var payload map[string]string
msg.UnmarshalPayload(&payload)

// Different message types
respMsg, _ := proc.NewResponseMessage("resp-1", map[string]interface{}{
    "status": "success",
    "result": 42,
})

eventMsg, _ := proc.NewEventMessage("evt-1", map[string]string{
    "event": "process_completed",
})

errorMsg := proc.NewErrorMessage("err-1", errors.New("something went wrong"))
```

The broker automatically sends event messages when processes exit:

```go
// Wait for process exit events
for {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    msg, err := broker.ReceiveMessage(ctx)
    cancel()
    
    if err != nil {
        break
    }
    
    if msg.Type == proc.MessageTypeEvent {
        var event map[string]interface{}
        msg.UnmarshalPayload(&event)
        if event["event"] == "process_exited" {
            fmt.Printf("Process %s exited with code %v\n", 
                event["id"], event["exit_code"])
        }
    }
}
```

### Dependencies

This project uses Go modules for dependency management:

```bash
go mod tidy
go mod download
```

Key dependencies:
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework for building RESTful APIs
- [go-resty/resty](https://github.com/go-resty/resty) - HTTP client library for making API requests
- [GORM](https://gorm.io/) - ORM library for database operations
- [GORM SQLite Driver](https://github.com/go-gorm/sqlite) - SQLite driver for GORM

### Testing

Run unit tests:

```bash
go test ./...
# Or use build script
./build.sh -t
```

Run unit tests with coverage:

```bash
go test -cover ./...
```

### Performance Benchmarking

The project includes comprehensive performance benchmarks for critical code paths.

Run performance benchmarks:

```bash
./build.sh -m
```

This will:
- Execute all benchmark tests across the codebase
- Generate a human-readable report with timing and memory allocation data
- Save results to `.build/benchmark-report.txt` and `.build/benchmark-raw.txt`
- Highlight slowest operations and highest memory allocations

The benchmark report includes:
- System information (OS, architecture, CPU)
- Detailed benchmark results for each function
- Summary of slowest operations (top 10)
- Summary of highest memory allocations (top 10)

**Performance Optimization Workflow:**

When optimizing performance:

1. Run benchmarks to establish a baseline:
   ```bash
   ./build.sh -m
   cp .build/benchmark-report.txt .build/benchmark-baseline.txt
   ```

2. Make your optimization changes

3. Run benchmarks again to measure impact:
   ```bash
   ./build.sh -m
   ```

4. Compare results and document improvements in your commit

Example benchmark results can be downloaded from GitHub Actions artifacts after each CI run.

### Integration Testing

The project includes a pytest-based integration test framework for testing RPC services and multi-service cooperation.

#### Prerequisites

- Python 3.12 or later
- pytest

#### Installation

Install test dependencies:

```bash
pip3 install -r tests/requirements.txt
```

#### Running Integration Tests

The integration tests use a shared broker instance that spawns and manages all services. This approach mirrors real-world usage where broker coordinates multiple RPC services.

Run fast integration tests (recommended):

```bash
# Skip slow tests that make external API calls
pytest tests/ -m "not slow"
```

Run all integration tests:

```bash
pytest tests/
```

Run specific test files:

```bash
# Test broker RPC service (fast)
pytest tests/test_broker_integration.py

# Test meta service with multiple cooperating services
pytest tests/test_cve_meta_integration.py

# Run benchmark tests to measure RPC performance
pytest tests/test_benchmarks.py --benchmark-only
```

Run with verbose output:

```bash
pytest tests/ -v
```

Run tests with specific markers:

```bash
# Run only integration tests
pytest tests/ -m integration

# Run only RPC tests
pytest tests/ -m rpc

# Skip slow tests (recommended for CI/CD)
pytest tests/ -m "not slow"

# Run only slow tests with extended timeout (requires NVD API access)
pytest tests/ -m slow --timeout=300

# Run benchmark tests (fast, optimized for CI)
pytest tests/ -m "benchmark and not slow" --benchmark-only --benchmark-min-rounds=5 --benchmark-max-time=1.0 --benchmark-warmup-iterations=0

# Run all benchmark tests (including slow ones)
pytest tests/ -m benchmark --benchmark-only
```

# Run only RPC tests
pytest tests/ -m rpc

# Skip slow tests (recommended for CI/CD)
pytest tests/ -m "not slow"

# Run only slow tests with extended timeout (requires NVD API access)
pytest tests/ -m slow --timeout=300
```

**Note on Slow Tests:**
Tests marked with `@pytest.mark.slow` make calls to the external NVD API and may:
- Take significantly longer to complete (2-5 minutes)
- Fail due to rate limiting (HTTP 429 errors)
- Require network connectivity

These tests are designed to use minimal API calls (1-2 CVEs) but may still encounter issues due to NVD API availability. For CI/CD pipelines, it's recommended to run only the fast tests using `-m "not slow"`.

#### Test Architecture

The integration tests follow a broker-centric architecture that mirrors real-world usage:

1. **Shared Broker Fixture** (`broker_with_services`): A module-scoped fixture that:
   - Uses pre-built binaries from `./build.sh -p` when available, or builds them on-demand
   - Starts a broker instance
   - Uses the broker to spawn all required RPC services (remote, local)
   - Verifies services are running before tests begin
   - Automatically cleans up after tests complete

2. **Service Interaction**: Tests interact with spawned services through the broker using RPC messages, ensuring:
   - Services are managed via broker (spawn, kill, list)
   - Message routing through broker is tested
   - Real-world multi-service cooperation is validated

3. **Performance Benchmarks** (Local development only):
   - `test_benchmarks.py` contains benchmarks for all RPC endpoints
   - Uses pytest-benchmark to measure operations per second
   - Available for local testing and performance regression tracking
   - Not included in CI due to environment variability

**Important Testing Principles:**
- All tests must follow the broker-first architecture
- Services like `access` are subprocesses - they should NOT have process management endpoints
- If testing REST APIs that need process management, the REST service should communicate with broker via RPC
- Never embed broker logic in subprocess tests

Example benchmark output (local testing):
```
Name (time in us)                     Min       Max      Mean   StdDev    Median     IQR  Outliers  OPS (Kops/s)
test_benchmark_list_processes     56.1350  490.6160  102.3540  17.8387  100.0690  8.7930   315;394        9.7700
```

#### Test Structure

The integration tests are located in the `tests/` directory:

- `tests/__init__.py` - Package initialization
- `tests/conftest.py` - Shared fixtures (broker_with_services, test_binaries)
- `tests/helpers.py` - Helper utilities for RPC testing
- `tests/test_broker_integration.py` - Integration tests for broker service (5 tests)
- `tests/test_cve_meta_integration.py` - Integration tests for meta service (4 tests)
- `tests/test_benchmarks.py` - Performance benchmarks for RPC endpoints
- `tests/requirements.txt` - Python dependencies for testing
- `pytest.ini` - Pytest configuration

#### Writing Integration Tests

The test framework provides utilities and shared fixtures for testing RPC services:

```python
# Use the shared broker fixture with all services running
def test_my_feature(broker_with_services):
    broker = broker_with_services
    
    # Services are already running, interact with them via broker
    response = broker.send_request("RPCListProcesses", {})
    assert response["type"] == "response"

# Or build and manage processes manually
from tests.helpers import RPCProcess, build_go_binary

# Build a Go binary for testing
build_go_binary("./cmd/broker", "/tmp/broker")

# Start an RPC process and send requests
with RPCProcess(["/tmp/broker"], process_id="test-broker") as broker:
    # Send RPC request
    response = broker.send_request("RPCListProcesses", {})
    
    # Verify response
    assert response["type"] == "response"
    assert "payload" in response
```

Key features:
- **Automatic process management**: Processes are started and stopped automatically
- **RPC communication**: Built-in support for sending/receiving RPC messages
- **Binary building**: Helper function to build Go binaries for testing
- **Timeout handling**: Configurable timeouts for RPC requests
- **Context managers**: Clean resource management with Python context managers

#### Test Coverage

The integration tests cover:

1. **Broker Service** (`test_broker_integration.py` - 5 tests):
   - Spawning processes
   - Listing processes
   - Getting process information
   - Spawning RPC processes
   - Killing processes

2. **CVE Meta Service** (`test_cve_meta_integration.py` - 4 tests):
   - Multi-service cooperation (meta, local, remote)
   - Getting remote CVE count
   - Batch fetching CVEs
   - Service orchestration and message routing

3. **Performance Benchmarks** (`test_benchmarks.py` - local testing only):
   - RPCSpawn performance
   - RPCListProcesses performance
   - RPCGetProcess performance
   - RPCSpawnRPC performance
   - Available for local development and performance tracking

These integration tests complement the Go unit tests by verifying that multiple services can work together correctly through RPC communication.

## Continuous Integration

The project uses GitHub Actions for automated testing:

### CI Pipeline Stages

1. **Unit Tests** (Always runs first)
   - Runs Go unit tests with race detection
   - Generates code coverage reports
   - Fast feedback on basic functionality

2. **Integration Tests** (Runs after unit tests pass)
   - Tests multi-service RPC communication
   - Uses broker-managed service architecture
   - Optimized for speed with reduced startup times
   - Typical duration: ~50-60 seconds

### Benchmark Tests (Local Testing Only)

Performance benchmarks are available for local testing but not included in CI due to environment variability:

```bash
# Run benchmark tests locally
pytest tests/ -v -m "benchmark and not slow" --benchmark-only --benchmark-min-rounds=5 --benchmark-max-time=1.0 --benchmark-warmup-iterations=0
```

Benchmark tests measure RPC endpoint performance and can help track regressions during local development.

### Slow Tests (Local Testing Only)

Network-dependent tests that make external NVD API calls are available for local testing:

```bash
# Run slow tests locally (requires NVD API access)
pytest tests/ -v -m slow --timeout=300
```

These tests are not part of CI to avoid rate limiting and environment issues.

### Running CI Locally

To run the same tests that CI runs:

```bash
# Unit tests
go test -v -race -coverprofile=coverage.out ./...

# Integration tests (fast)
pip install -r tests/requirements.txt
pytest tests/ -v -m "not slow and not benchmark"
```

### Viewing CI Results

- **Coverage reports**: Available as artifacts after each run
- **Test logs**: Full output available in GitHub Actions logs

The simplified CI pipeline provides fast feedback to developers with unit and integration tests only.

## License

MIT
