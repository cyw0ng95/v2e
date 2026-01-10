# v2e

A Go-based project demonstrating a multi-command structure with CVE (Common Vulnerabilities and Exposures) data fetching capabilities.

## Project Structure

This project contains multiple commands:

- `cmd/server` - A simple HTTP server
- `cmd/client` - A simple HTTP client
- `cmd/broker` - Process broker demo for managing subprocesses

And packages:

- `pkg/common` - Common utilities and configuration
- `pkg/repo` - Repository layer for external data sources (NVD CVE API)
- `pkg/proc` - Process broker for managing subprocesses and inter-process communication

## Prerequisites

- Go 1.24 or later

## Building

To build all commands:

```bash
go build ./cmd/server
go build ./cmd/client
```

Or build a specific command:

```bash
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

## Running

### Configuration

Both the server and client support optional configuration via a `config.json` file in the current directory. If the file doesn't exist, default values will be used.

A sample configuration file is provided as `config.json.example`. You can copy it to `config.json` and modify as needed:

```bash
cp config.json.example config.json
```

Example `config.json`:

```json
{
  "server": {
    "address": ":8080"
  },
  "client": {
    "url": "http://localhost:8080"
  }
}
```

Configuration options:
- `server.address`: The address for the server to listen on (default: `:8080`)
- `client.url`: The default URL for the client to connect to (default: `http://localhost:8080`)

Note: Command line arguments take precedence over configuration file values.

### Server

```bash
go run ./cmd/server
```

The server will start on port 8080.

### Client

```bash
go run ./cmd/client [url]
```

If no URL is provided, it will connect to `http://localhost:8080` by default.

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


## Development

### CVE Fetcher

The `pkg/repo` package provides a CVE fetcher that integrates with the NVD API v2.0:

```go
import "github.com/cyw0ng95/v2e/pkg/repo"

// Create a new CVE fetcher (optionally with API key for higher rate limits)
fetcher := repo.NewCVEFetcher("")

// Fetch a specific CVE by ID
cveData, err := fetcher.FetchCVEByID("CVE-2021-44228")

// Fetch multiple CVEs with pagination
cveList, err := fetcher.FetchCVEs(0, 10)
```

For production use with higher rate limits, obtain an API key from [NVD](https://nvd.nist.gov/developers/request-an-api-key) and pass it to `NewCVEFetcher()`.

### CVE Database

The project includes a GORM-based ORM engine for storing CVE data in a local SQLite database:

```go
import "github.com/cyw0ng95/v2e/pkg/repo"

// Create or open the CVE database
db, err := repo.NewDB("cve.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Save a CVE to the database
cve := &repo.CVEItem{
    ID:           "CVE-2021-44228",
    SourceID:     "nvd@nist.gov",
    Published:    time.Now(),
    LastModified: time.Now(),
    VulnStatus:   "Analyzed",
    Descriptions: []repo.Description{
        {Lang: "en", Value: "Apache Log4j vulnerability"},
    },
}
err = db.SaveCVE(cve)

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
go test ./pkg/repo -v -run TestCreateCVEDatabase
```

This will create `cve.db` in the project root with sample CVE records that you can inspect or download.

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
