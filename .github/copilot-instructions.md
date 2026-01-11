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

This ensures integration tests follow the same deployment model as production, where the broker is the central entry point.
