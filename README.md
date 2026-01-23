# v2e

A minimal Go-based example project that demonstrates a broker-first architecture for running multiple subprocess services that communicate via RPC messages over stdin/stdout.

Key points
- Broker-first architecture: `cmd/broker` is the single process that spawns and manages all subprocesses.
- Subprocesses (e.g., `access`, `cve-remote`, `cve-local`, `cve-meta`) expose functionality via RPC handlers and communicate only through the broker.
- Designed for local development and testing; configuration is provided via `config.json`.

Quickstart

Prerequisites: Go 1.24+ and basic shell tools.

Build all commands:

```bash
# Build binaries
go build ./cmd/...
```

Run with the broker (recommended):

```bash
# Start the broker which spawns configured subprocesses
./broker config.json
```

Run tests:

```bash
# Unit tests
./build.sh -t
```

## Live Development Workflow

To enable live development, use the `-r` option with the `build.sh` script. This option is designed to streamline the development process by automatically restarting the broker and Node.js processes whenever changes are detected in the Go source files or frontend assets.

### Usage

Run the following command from the project root:

```bash
./build.sh -r
```

### Features
- **Automatic Restart**: The broker and Node.js processes are restarted automatically on file changes.
- **Debouncing**: Prevents rapid restarts by introducing a delay between file change detection and process restarts.
- **Process Cleanup**: Ensures that all subprocesses are properly terminated before restarting.

### Notes
- Ensure that all dependencies are installed and the environment is properly configured before using the `-r` option.
- This workflow is intended for development purposes only and should not be used in production environments.

Project layout (high level)
- cmd/
  - broker/    - process manager and RPC router
  - access/    - REST gateway (subprocess)
  - cve-remote/ - fetch CVE data from remote APIs
  - cve-local/ - local SQLite storage service
  - cve-meta/  - orchestration and job control
- pkg/
  - proc/subprocess - subprocess framework (stdin/stdout RPC)
  - cve              - domain types and helpers
  - common           - config and logging utilities
- tests/            - pytest tests
- website/          - static frontend (Next.js export)

Notes and conventions
- All subprocesses must be started and managed by the broker; do not run backend subprocesses directly in production or integration tests.
- Subprocesses communicate only via JSON RPC messages over stdin/stdout.
- Configuration (process list, logging) is controlled through `config.json`.
- The authoritative RPC API specification for each subprocess can be found in the top comment of its `cmd/*/main.go` file.

Where to look next
- `cmd/broker` — broker implementation and message routing
- `pkg/proc/subprocess` — helper framework for subprocesses
- `cmd/access` — REST gateway and example of using the RPC client
- `tests/` — integration tests demonstrating usage patterns

License
- MIT
