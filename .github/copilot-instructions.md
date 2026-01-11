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
2. **stdin/stdout only**: Only read from stdin and write to stdout
3. **No external inputs**: Do NOT accept command-line arguments, environment variables (except PROCESS_ID and service-specific config), or network connections for control
4. **RPC handlers only**: All functionality must be exposed via RPC message handlers
5. **Broker-spawned**: Services must be designed to be spawned by the broker via `SpawnRPC`
6. **Signal handling**: Implement graceful shutdown with signal handlers (SIGINT, SIGTERM)

Example subprocess pattern:
```go
func main() {
    // Get process ID from environment
    processID := os.Getenv("PROCESS_ID")
    if processID == "" {
        processID = "my-service"
    }

    // Create subprocess instance
    sp := subprocess.New(processID)

    // Register RPC handlers
    sp.RegisterHandler("RPCMyFunction", createMyFunctionHandler())

    // Set up signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Run subprocess
    errChan := make(chan error, 1)
    go func() {
        errChan <- sp.Run()
    }()

    // Wait for completion or signal
    select {
    case err := <-errChan:
        if err != nil {
            sp.SendError("fatal", err)
            os.Exit(1)
        }
    case <-sigChan:
        sp.SendEvent("subprocess_shutdown", map[string]string{
            "id": sp.ID,
            "reason": "signal received",
        })
        sp.Stop()
    }
}
```

### Configuration Guidelines

- Use `config.json` to define broker settings and processes to spawn
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
