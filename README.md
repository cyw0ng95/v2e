# v2e

A Go-based project demonstrating process management with CVE (Common Vulnerabilities and Exposures) data fetching capabilities.

## Project Structure

This project contains one command:

- `cmd/broker` - Process broker for managing subprocesses with lifecycle control

And packages:

- `pkg/common` - Common utilities and configuration
- `pkg/repo` - Repository layer for external data sources (NVD CVE API)
- `pkg/proc` - Process broker for managing subprocesses and inter-process communication
- `pkg/cve` - CVE data types shared across packages
- `pkg/cve/remote` - Remote CVE fetching from NVD API
- `pkg/cve/local` - Local CVE storage with SQLite database

## Prerequisites

- Go 1.24 or later

## Building

To build the broker command:

```bash
go build -o bin/broker ./cmd/broker
```

## Running

### Broker

```bash
# Run demo mode (spawns multiple example processes)
go run ./cmd/broker

# Execute a specific command
go run ./cmd/broker -cmd "echo hello world"

# Execute a command with a custom process ID
go run ./cmd/broker -id my-process -cmd "sleep 5"
```

The broker command demonstrates the process management capabilities of the `pkg/proc` package. In demo mode, it spawns multiple processes and monitors their lifecycle, showing how processes are reaped and their exit codes captured.

### Managed Process Example

An example demonstrating the ManagedProcess interface is available:

```bash
go run ./examples/managed_process
```

This example creates multiple worker processes that communicate with the broker through message passing, demonstrating:
- Process registration and lifecycle management
- Request/response message handling
- Event publishing and monitoring
- Graceful shutdown of managed processes


## Development

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

The `pkg/proc` package provides a process broker for managing subprocesses and inter-process communication. It supports two types of processes:

1. **External processes** - System commands spawned by the broker
2. **Managed processes** - Go processes with structured lifecycle and message handling

#### Managing External Processes

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

#### Managed Processes with Lifecycle Control

The broker supports managed processes that implement the `ManagedProcess` interface, providing structured lifecycle management and message handling:

```go
import (
    "context"
    "github.com/cyw0ng95/v2e/pkg/proc"
)

// Define a custom managed process
type MyProcess struct {
    *proc.BaseProcess
    // Add custom fields
}

// Create a new managed process
func NewMyProcess(id string) *MyProcess {
    return &MyProcess{
        BaseProcess: proc.NewBaseProcess(id),
    }
}

// Implement lifecycle hooks
func (p *MyProcess) Start(ctx context.Context, broker *proc.Broker) error {
    // Initialize the base process
    if err := p.BaseProcess.Start(ctx, broker); err != nil {
        return err
    }
    
    // Custom initialization logic
    p.SendEvent("process-started", map[string]string{
        "id": p.ID(),
    })
    
    return nil
}

func (p *MyProcess) Stop() error {
    // Custom cleanup logic
    p.SendEvent("process-stopping", map[string]string{
        "id": p.ID(),
    })
    
    // Stop the base process
    return p.BaseProcess.Stop()
}

func (p *MyProcess) OnMessage(msg *proc.Message) error {
    // Handle incoming messages from the broker
    switch msg.Type {
    case proc.MessageTypeRequest:
        var payload map[string]interface{}
        msg.UnmarshalPayload(&payload)
        // Process request and send response
        return p.SendResponse(msg.ID, map[string]interface{}{
            "status": "ok",
            "result": "processed",
        })
    case proc.MessageTypeEvent:
        // Handle events
        return nil
    }
    return nil
}

// Register and use the managed process
broker := proc.NewBroker()
defer broker.Shutdown()

myProc := NewMyProcess("worker-1")
err := broker.RegisterManagedProcess(myProc)
if err != nil {
    log.Fatal(err)
}

// Send messages to the managed process
msg, _ := proc.NewRequestMessage("req-1", map[string]string{
    "action": "process_data",
})
err = broker.DispatchMessage("worker-1", msg)

// List all managed processes
managedProcs := broker.ListManagedProcesses()

// Stop a managed process
err = broker.StopManagedProcess("worker-1")
```

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
- [go-resty/resty](https://github.com/go-resty/resty) - HTTP client library for making API requests
- [GORM](https://gorm.io/) - ORM library for database operations
- [GORM SQLite Driver](https://github.com/go-gorm/sqlite) - SQLite driver for GORM

### Testing

Run tests:

```bash
go test ./...
```

## License

MIT
