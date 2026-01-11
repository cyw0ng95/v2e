# v2e

A Go-based project demonstrating a multi-command structure with CVE (Common Vulnerabilities and Exposures) data fetching capabilities.

## Project Structure

This project contains multiple commands:

- `cmd/access` - RESTful API service using Gin framework
- `cmd/broker` - RPC service for managing subprocesses and process lifecycle
- `cmd/broker-stats` - RPC service for accessing broker message statistics
- `cmd/cve-remote` - RPC service for fetching CVE data from NVD API
- `cmd/cve-local` - RPC service for storing and retrieving CVE data from local database
- `cmd/cve-meta` - Backend RPC service that orchestrates CVE fetching and storage operations

And packages:

- `pkg/common` - Common utilities and configuration
- `pkg/repo` - Repository layer for external data sources (NVD CVE API)
- `pkg/proc` - Process broker for managing subprocesses and inter-process communication
- `pkg/proc/subprocess` - Common subprocess framework for message-driven subprocesses
- `pkg/cve` - CVE data types shared across packages
- `pkg/cve/remote` - Remote CVE fetching from NVD API
- `pkg/cve/local` - Local CVE storage with SQLite database

## Prerequisites

- Go 1.24 or later

## Building

To build all commands:

```bash
go build ./cmd/access
go build ./cmd/broker
go build ./cmd/broker-stats
go build ./cmd/cve-remote
go build ./cmd/cve-local
go build ./cmd/cve-meta
```

Or build a specific command:

```bash
go build -o bin/access ./cmd/access
go build -o bin/broker ./cmd/broker
go build -o bin/broker-stats ./cmd/broker-stats
go build -o bin/cve-remote ./cmd/cve-remote
go build -o bin/cve-local ./cmd/cve-local
go build -o bin/cve-meta ./cmd/cve-meta
```

## Running

### Access (RESTful API Service)

The Access service provides a RESTful API server using the Gin framework:

```bash
# Run with default settings (port 8080)
go run ./cmd/access

# Run on a custom port
go run ./cmd/access -port 3000

# Run in debug mode
go run ./cmd/access -debug

# Combine options
go run ./cmd/access -port 3000 -debug
```

**Available Endpoints:**
- `GET /health` - Health check endpoint that returns the server status

**Command Line Options:**
- `-port` - Port to listen on (default: 8080)
- `-debug` - Enable debug mode (default: false)

The Access service demonstrates the use of the Gin framework for building RESTful APIs. It can be extended with additional endpoints as needed.

### Broker (RPC Service)

The Broker service provides RPC interfaces for managing subprocesses and accessing message statistics:

```bash
# Run the service (it reads RPC requests from stdin and writes responses to stdout)
go run ./cmd/broker

# Example: Spawn a process
echo '{"type":"request","id":"RPCSpawn","payload":{"id":"my-echo","command":"echo","args":["hello","world"]}}' | go run ./cmd/broker

# Example: List all processes
echo '{"type":"request","id":"RPCListProcesses","payload":{}}' | go run ./cmd/broker

# Example: Get process info
echo '{"type":"request","id":"RPCGetProcess","payload":{"id":"my-echo"}}' | go run ./cmd/broker

# Example: Kill a process
echo '{"type":"request","id":"RPCKill","payload":{"id":"my-echo"}}' | go run ./cmd/broker

# Example: Get total message count
echo '{"type":"request","id":"RPCGetMessageCount","payload":{}}' | go run ./cmd/broker

# Example: Get detailed message statistics
echo '{"type":"request","id":"RPCGetMessageStats","payload":{}}' | go run ./cmd/broker
```

**Available RPC Interfaces:**

*Process Management:*
- `RPCSpawn` - Spawns a new subprocess with the specified command and arguments
- `RPCSpawnRPC` - Spawns a new RPC-enabled subprocess (with stdin/stdout pipes)
- `RPCGetProcess` - Gets information about a specific process by ID
- `RPCListProcesses` - Lists all managed processes
- `RPCKill` - Terminates a process by ID

*RPC Endpoint Management:*
- `RPCRegisterEndpoint` - Registers an RPC endpoint for a process
- `RPCGetEndpoints` - Gets all registered RPC endpoints for a specific process
- `RPCGetAllEndpoints` - Gets all registered RPC endpoints for all processes

*Message Statistics:*
- `RPCGetMessageCount` - Returns the total count of messages processed (sent + received)
- `RPCGetMessageStats` - Returns detailed statistics including counts by type and timestamps

**Configuration File Support:**

The broker can load process configurations from `config.json`:

```json
{
  "broker": {
    "log_file": "broker.log",
    "processes": [
      {
        "id": "cve-remote",
        "command": "go",
        "args": ["run", "./cmd/cve-remote"],
        "rpc": true,
        "restart": true,
        "max_restarts": -1
      }
    ]
  }
}
```

Configuration options:
- `log_file` - Path to log file (logs will be written to both stdout and file)
- `processes` - Array of processes to automatically start
  - `id` - Unique process identifier
  - `command` - Command to execute
  - `args` - Command arguments
  - `rpc` - Whether to enable RPC communication (stdin/stdout pipes)
  - `restart` - Whether to automatically restart on exit
  - `max_restarts` - Maximum restart attempts (-1 for unlimited)

**Auto-Restart Feature:**

Processes can be configured to automatically restart on exit:

```bash
# Spawn a process with auto-restart (max 3 restarts)
echo '{"type":"request","id":"RPCSpawn","payload":{"id":"my-worker","command":"./worker","args":[],"restart":true,"max_restarts":3}}' | go run ./cmd/broker
```

The broker will monitor process exits and restart them automatically according to the configuration.

**Dual Logging:**

When a log file is configured, the broker writes logs to both stdout and the specified file simultaneously, ensuring that log messages are visible in real-time while also being persisted to disk.

**Request Format for RPCSpawn:**
```json
{
  "id": "unique-process-id",
  "command": "echo",
  "args": ["hello", "world"]
}
```

**Response Format for RPCSpawn:**
```json
{
  "id": "unique-process-id",
  "pid": 12345,
  "command": "echo",
  "status": "running"
}
```

This service can be spawned by a broker to provide remote access to process management via RPC.

### Broker Stats (RPC Service)

The Broker Stats service provides RPC interfaces for accessing message statistics from a broker instance:

```bash
# Run the service (it reads RPC requests from stdin and writes responses to stdout)
go run ./cmd/broker-stats

# Example: Get total message count
echo '{"type":"request","id":"RPCGetMessageCount","payload":{}}' | go run ./cmd/broker-stats

# Example: Get detailed message statistics
echo '{"type":"request","id":"RPCGetMessageStats","payload":{}}' | go run ./cmd/broker-stats
```

**Available RPC Interfaces:**
- `RPCGetMessageCount` - Returns the total count of messages processed (sent + received)
- `RPCGetMessageStats` - Returns detailed statistics including counts by type and timestamps

**Response Format for RPCGetMessageCount:**
```json
{
  "total_count": 42
}
```

**Response Format for RPCGetMessageStats:**
```json
{
  "total_sent": 21,
  "total_received": 21,
  "request_count": 10,
  "response_count": 10,
  "event_count": 1,
  "error_count": 1,
  "first_message_time": "2026-01-10T15:00:00Z",
  "last_message_time": "2026-01-10T15:30:00Z"
}
```

This service can be spawned by a broker to provide remote access to process management and message statistics via RPC.

### Worker

The worker is an example subprocess that demonstrates the `pkg/proc/subprocess` framework:

```bash
# Run the worker (it reads messages from stdin and writes to stdout)
go run ./cmd/worker

# Example: Send a ping message
echo '{"type":"request","id":"ping"}' | go run ./cmd/worker

# Example: Send an echo request
echo '{"type":"request","id":"req-1","payload":{"action":"echo","data":"hello"}}' | go run ./cmd/worker
```

The worker demonstrates how to build message-driven subprocesses that can be controlled by the broker.

### CVE Remote (RPC Service)

The CVE Remote service provides RPC interfaces for fetching CVE data from the NVD API:

```bash
# Run the service (it reads RPC requests from stdin and writes responses to stdout)
go run ./cmd/cve-remote

# Example: Get total CVE count from NVD
echo '{"type":"request","id":"RPCGetCVECnt","payload":{}}' | go run ./cmd/cve-remote

# Example: Fetch a specific CVE by ID
echo '{"type":"request","id":"RPCGetCVEByID","payload":{"cve_id":"CVE-2021-44228"}}' | go run ./cmd/cve-remote
```

**Available RPC Interfaces:**
- `RPCGetCVECnt` - Returns the total count of CVEs in the NVD database
- `RPCGetCVEByID` - Fetches a specific CVE by its ID from the NVD API

**Environment Variables:**
- `NVD_API_KEY` - Optional NVD API key for higher rate limits

### CVE Local (RPC Service)

The CVE Local service provides RPC interfaces for storing and retrieving CVE data from a local SQLite database:

```bash
# Run the service (it reads RPC requests from stdin and writes responses to stdout)
go run ./cmd/cve-local

# Example: Check if a CVE is stored locally
echo '{"type":"request","id":"RPCIsCVEStoredByID","payload":{"cve_id":"CVE-2021-44228"}}' | go run ./cmd/cve-local

# Example: Save a CVE to local database
echo '{"type":"request","id":"RPCSaveCVEByID","payload":{"cve":{"id":"CVE-2021-44228",...}}}' | go run ./cmd/cve-local
```

**Available RPC Interfaces:**
- `RPCIsCVEStoredByID` - Checks if a CVE exists in the local database
- `RPCSaveCVEByID` - Saves a CVE to the local database

**Environment Variables:**
- `CVE_DB_PATH` - Path to the SQLite database file (default: `cve.db`)

### CVE Meta Service

The CVE Meta service is a backend RPC service that orchestrates CVE fetching and storage operations. It runs continuously and accepts RPC commands to perform batch jobs.

```bash
# Run the service (it reads RPC requests from stdin and writes responses to stdout)
go run ./cmd/cve-meta

# Example: Fetch and store a single CVE
echo '{"type":"request","id":"RPCFetchAndStoreCVE","payload":{"cve_id":"CVE-2021-44228"}}' | go run ./cmd/cve-meta

# Example: Fetch and store multiple CVEs in batch
echo '{"type":"request","id":"RPCBatchFetchCVEs","payload":{"cve_ids":["CVE-2021-44228","CVE-2024-1234"]}}' | go run ./cmd/cve-meta

# Example: Get total CVE count from NVD
echo '{"type":"request","id":"RPCGetRemoteCVECount","payload":{}}' | go run ./cmd/cve-meta
```

**Available RPC Interfaces:**
- `RPCFetchAndStoreCVE` - Fetches a CVE from NVD (if not already stored locally) and saves it to the database
- `RPCBatchFetchCVEs` - Fetches and stores multiple CVEs in batch mode
- `RPCGetRemoteCVECount` - Returns the total count of CVEs in the NVD database

**Environment Variables:**
- `CVE_DB_PATH` - Path to the SQLite database file (default: `cve.db`)

The service performs the following workflow:
1. Spawns `cve-local` and `cve-remote` services as subprocesses
2. Accepts RPC commands via stdin
3. For `RPCFetchAndStoreCVE`:
   - Checks if the CVE is already stored locally via `cve-local`
   - If not stored, fetches it from NVD via `cve-remote`
   - Saves the fetched CVE to the local database via `cve-local`
4. For `RPCBatchFetchCVEs`:
   - Processes multiple CVE IDs
   - Returns success/failure status for each CVE
5. For `RPCGetRemoteCVECount`:
   - Forwards the request to `cve-remote` and returns the total count

This demonstrates the broker-mediated RPC communication pattern where:
- The meta service acts as an orchestrator/backend service
- All communication happens via RPC messages
- The service runs continuously accepting commands
- Batch jobs can be executed efficiently

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

The subprocess framework allows you to build worker processes that:
1. Are controlled by the broker through message passing
2. Can be spawned and monitored by the broker
3. Have a clear lifecycle with proper initialization and shutdown
4. Focus on business logic without worrying about process management

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
info, err := broker.SpawnRPC("cve-remote", "go", "run", "./cmd/cve-remote")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Started RPC process %s with PID %d\n", info.ID, info.PID)

// Send a message to the subprocess
req, _ := proc.NewRequestMessage("RPCGetCVEByID", map[string]string{
    "cve_id": "CVE-2021-44228",
})
err = broker.SendToProcess("cve-remote", req)
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
```

Run unit tests with coverage:

```bash
go test -cover ./...
```

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

# Test cve-meta service with multiple cooperating services
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
   - Builds all required binaries once per test session
   - Starts a broker instance
   - Uses the broker to spawn all required RPC services (worker, cve-remote, cve-local)
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
- `tests/test_cve_meta_integration.py` - Integration tests for cve-meta service (4 tests)
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
   - Multi-service cooperation (cve-meta, cve-local, cve-remote)
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
