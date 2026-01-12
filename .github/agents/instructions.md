# Copilot Instructions for v2e

## Documentation Guidelines

- Do **NOT** generate any documents other than `README.md`
- All project documentation should be consolidated in the `README.md` file
- Avoid creating additional markdown files, guides, or documentation files
- Keep the documentation simple and focused

## Project Guidelines

- This is a Go-based project using Go modules
- The project may contain multiple commands in the `cmd/` directory
- Follow Go best practices and conventions
- Use standard Go tooling (go build, go test, go mod)

## Go Code Style - Effective Go Conformance

All Go code MUST follow the guidelines from https://go.dev/doc/effective_go and standard Go conventions:

### Naming Conventions
- **Package names**: lowercase, single-word, no underscores (e.g., `proc`, `common`, not `proc_utils`)
- **Interfaces**: Use `-er` suffix for single-method interfaces (e.g., `Reader`, `Writer`, `Handler`)
- **MixedCaps**: Use `MixedCaps` or `mixedCaps` for multi-word names, never underscores
- **Getters**: Don't use `Get` prefix (e.g., `Owner()` not `GetOwner()`)
- **Acronyms**: Keep consistent case (e.g., `HTTPServer` or `httpServer`, not `HttpServer`)

### Commentary
- **Package comments**: Every package should have a package comment before the package declaration
- **Exported names**: Every exported name should have a doc comment
- **Complete sentences**: Doc comments should be complete sentences starting with the name being declared
- **Group related declarations**: Use blank lines and comments to group related declarations

### Error Handling
- **Check all errors**: Never ignore errors, handle them explicitly
- **Error messages**: Should be lowercase and not end with punctuation (unless it's a question mark or exclamation point)
- **Custom errors**: Use `fmt.Errorf` with `%w` for error wrapping
- **Error returns**: Put error as the last return value

### Control Structures
- **No parentheses**: Don't use parentheses in if, for, switch statements unless necessary for precedence
- **Short variable declarations**: Use `:=` in if statements when appropriate
- **switch**: Prefer switch over if-else chains
- **Goroutines**: Use goroutines for concurrent operations, don't forget to handle cleanup

### Functions and Methods
- **Receiver names**: Should be short, consistent, and reflect the type (e.g., `func (b *Broker)`)
- **Multiple return values**: Use when appropriate, especially for returning values and errors
- **Defer**: Use defer for cleanup operations (close files, unlock mutexes, etc.)
- **Named result parameters**: Use sparingly, mainly for documentation in complex functions

### Interfaces
- **Small interfaces**: Prefer small, focused interfaces (1-3 methods)
- **Accept interfaces, return structs**: Functions should accept interfaces and return concrete types
- **Empty interface**: Use `interface{}` or `any` sparingly, prefer specific types

### Concurrency
- **Channels**: Use channels to communicate between goroutines
- **sync.Mutex**: Use for protecting shared state
- **sync.WaitGroup**: Use for waiting for goroutines to complete
- **Context**: Use context.Context for cancellation and timeouts
- **Don't communicate by sharing memory**: Share memory by communicating (use channels)

### Data Structures and Memory
- **Composite literals**: Use composite literals for initialization
- **Zero values**: Design types to work correctly when zero-valued
- **new vs make**: Use `new` for zero-valued allocation, `make` for slices, maps, channels
- **Slices**: Prefer slices over arrays for flexibility
- **Maps**: Initialize maps with `make` before use

### Performance Guidelines
- **Avoid allocations**: Reuse objects when possible (e.g., sync.Pool)
- **String building**: Use strings.Builder for efficient string concatenation
- **Byte slices**: Keep data as []byte instead of converting to string unnecessarily
- **Defer overhead**: Be aware defer has small overhead, avoid in tight loops
- **Buffer reuse**: Reuse buffers to reduce GC pressure

### Testing
- **Test files**: Name test files with `_test.go` suffix
- **Test functions**: Start with `Test` prefix (e.g., `TestBrokerSpawn`)
- **Table-driven tests**: Use for testing multiple cases
- **Examples**: Write example functions for documentation

### Code Organization
- **Package structure**: Organize code into focused packages
- **Internal packages**: Use `internal/` for private packages
- **cmd/**: Put command-line programs in `cmd/` directory
- **pkg/**: Put reusable library code in `pkg/` directory

## Architecture and Security Guidelines

### Deployment Model

- The broker is the **central entry point** for deployment
- Users deploy the package by running the broker with a config file (`config.json`)
- The broker spawns and manages all subprocess services
- All subprocesses are started by the broker, not directly by users

### Message Passing and Security

- **RPC-only communication**: All inter-process communication MUST use RPC messages
- **Broker-mediated routing**: All messages MUST be routed through the broker to ensure security
- **No direct communication**: Subprocesses MUST NOT communicate directly with each other
- **No external input**: Subprocesses MUST NOT accept any input other than stdin pipeline from the broker

### Subprocess Implementation Rules

When implementing a new subprocess service:

1. **Use the subprocess framework**: Always use `pkg/proc/subprocess` package
2. **Use common logging**: Call `subprocess.SetupLogging(processID)` to initialize logging from config.json
3. **Use common lifecycle**: Call `subprocess.RunWithDefaults(sp, logger)` for standard signal handling and error handling
4. **stdin/stdout only**: Only read from stdin and write to stdout
5. **No external inputs**: Do NOT accept command-line arguments, environment variables (except PROCESS_ID and service-specific config), or network connections for control
6. **RPC handlers only**: All functionality must be exposed via RPC message handlers
7. **Broker-spawned**: Services must be designed to be spawned by the broker via `SpawnRPC`

Example subprocess pattern:
```go
func main() {
    // Get process ID from environment
    processID := os.Getenv("PROCESS_ID")
    if processID == "" {
        processID = "my-service"
    }

    // Set up logging using common subprocess framework
    logger, err := subprocess.SetupLogging(processID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
        os.Exit(1)
    }

    // Create subprocess instance
    sp := subprocess.New(processID)

    // Register RPC handlers
    sp.RegisterHandler("RPCMyFunction", createMyFunctionHandler())

    logger.Info("My service started")

    // Run with default lifecycle management
    subprocess.RunWithDefaults(sp, logger)
}
```

### Configuration Guidelines

- Use `config.json` to define broker settings, logging, and processes to spawn
- Logging configuration in `config.json` applies to all subprocesses via the subprocess framework
- All process configurations should be in the broker config
- Subprocesses should only use environment variables for service-specific settings (e.g., database paths, API keys)

## Testing Guidelines

- When adding new features, always consider adding test cases
- Unit tests should be written in Go using the standard `testing` package
- Integration tests should be written in Python using pytest
- Test cases should cover:
  - Normal operation paths
  - Error handling and edge cases
  - Integration between multiple services (for RPC-based features)
  - Security constraints (e.g., subprocesses only accepting stdin input)
- Run `./build.sh -t` to execute unit tests
- Run `./build.sh -i` to execute integration tests
- Test case generation should be comprehensive and follow existing patterns in the codebase

### Integration Test Constraints

**IMPORTANT**: All integration tests MUST follow the broker-first architecture:

1. **Start broker first**: The broker (or access service which embeds a broker) must be started before any subprocess
2. **No direct subprocess testing**: Do NOT start or test subprocesses directly without going through the broker
3. **Use access service as gateway**: Integration tests should use the access REST API as the primary entry point
4. **Broker spawns subprocesses**: Let the broker spawn and manage all subprocess services via configuration or REST API
5. **RESTful testing only**: All RPC tests for backend services (like cve-meta) MUST use the RESTful API endpoint `/restful/rpc` with the `target` parameter

#### Testing Backend Services via RESTful API

When testing backend RPC services (cve-meta, cve-local, cve-remote), use the access service's generic RPC endpoint:

```python
# Example: Testing cve-meta service
access = AccessClient()
response = access.rpc_call(
    method="RPCGetCVE",
    target="cve-meta",  # Target the specific backend service
    params={"cve_id": "CVE-2021-44228"}
)

# Verify standardized response format
assert response["retcode"] == 0
assert response["message"] == "success"
assert response["payload"] is not None
```

The access service routes the request as follows:
1. External test → REST API (`POST /restful/rpc`)
2. Access service → Broker (via RPC with `target` field)
3. Broker → Backend service (e.g., cve-meta)
4. Response flows back through the same chain

This ensures integration tests follow the same deployment model as production, where the broker is the central entry point.
