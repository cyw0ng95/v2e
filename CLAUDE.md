# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

v2e (Vulnerabilities Viewer Engine) is a broker-first microservices system for managing CVE, CWE, CAPEC, and ATT&CK security data. The architecture enforces a central broker pattern where `cmd/broker` is the sole process manager and message router - no subprocess-to-subprocess communication is allowed.

## Build & Development Commands

```bash
# Primary build script
./build.sh -t     # Unit tests (excludes fuzz tests, uses -run='^Test')
./build.sh -f     # Fuzz tests (5 seconds per target)
./build.sh -m     # Benchmarks with reporting
./build.sh -p     # Build and package binaries + assets
./build.sh -r     # Development mode (auto-restart on changes)
./build.sh -c     # Run vconfig TUI for configuration

# Containerized development (macOS always uses container)
./runenv.sh -t    # Run tests in container
./runenv.sh -r    # Dev mode in container

# Frontend (website/)
cd website
npm run dev       # Development server
npm run build     # Static export to out/
npm run lint      # ESLint

# Manual builds
go build ./cmd/...                              # Build all services
go test -tags="$GO_TAGS" -race ./...            # Unit tests with race detector
```

**Build Tags**: Default `GO_TAGS=CONFIG_USE_LIBXML2`. Override via environment variable.

**Version Requirements**: Go 1.21+, Node.js 20+, npm 10+. macOS requires Podman for containerized builds.

## Architecture: Broker-First Pattern

The broker is the central orchestrator. All subprocess services (`cmd/access`, `cmd/local`, `cmd/remote`, `cmd/meta`, `cmd/sysmon`) communicate exclusively via stdin/stdout RPC messages routed through the broker.

### Communication Flow
```
Frontend (Next.js) → Access Service (/restful/rpc) → Broker → Backend Services
```

### Core Rules
1. **Only broker spawns subprocesses** - never add process management to `cmd/*` services
2. **RPC-only inter-service communication** - no direct subprocess-to-subprocess interaction
3. **Subprocess I/O constraint** - services must only read stdin / write stdout for broker-controlled RPC

### Transport Layer
- **Default**: Unix Domain Sockets (UDS) with 0600 permissions
- **Fallback**: File Descriptor Pipes (configured via build-time ldflags)
- Message types: Request, Response, Event, Error (with correlation IDs)

### Key Packages
- `pkg/proc/subprocess` - Subprocess lifecycle framework, provides `SetupLogging()`, `RunWithDefaults()`, `RegisterHandler()`
- `pkg/proc/message` - Message types with `sync.Pool` optimization
- `cmd/broker/core` - Broker orchestrator, Process, Router interfaces
- `cmd/broker/transport` - UDS and FD Pipe transport implementations
- `cmd/broker/perf` - Optimizer with adaptive tuning, message batching, worker pools
- `pkg/cve/taskflow` - Taskflow-based job executor with BoltDB persistence

## Subprocess Development Pattern

All subprocesses under `cmd/*` must follow this pattern:

```go
import "github.com/cyw0ng95/v2e/pkg/proc/subprocess"

func main() {
    // 1. Setup logging with process ID
    logger := subprocess.SetupLogging("my-service")

    // 2. Create subprocess
    sp := subprocess.NewSubprocess(logger)

    // 3. Register RPC handlers
    sp.RegisterHandler("RPCMyMethod", func(params json.RawMessage) (interface{}, error) {
        // Handler implementation
    })

    // 4. Run with defaults (stdin/stdout RPC, graceful shutdown)
    subprocess.RunWithDefaults(sp, logger)
}
```

## Testing Guidelines

### Unit Tests
- Located alongside source (`*_test.go`)
- Pattern: `Test*` functions only (use `-run='^Test'` to exclude fuzz tests)
- Table-driven tests preferred
- Run with `-race` flag

### Fuzz Tests
- Pattern: `Fuzz*` functions
- Run separately via `./build.sh -f`
- 1-second duration on CI

### Benchmarks
- Pattern: `Benchmark*` functions
- Output: TSV + aggregated report
- Run via `./build.sh -m`

### Integration Tests
- Use pytest framework in `tests/` directory
- **Must** start broker/access gateway - never spawn subprocesses directly
- Test via `/restful/rpc` endpoint

### CI Configuration
`.github/workflows/test.yml` defines: unit-tests, fuzz-tests, build-and-package, performance-benchmarks jobs.

## Performance Principles

The codebase emphasizes performance optimization. Key patterns:

1. **Sonic JSON**: Use `github.com/bytedance/sonic` for zero-copy JSON (JIT-optimized)
2. **Pre-allocate slices**: Use exact capacity when final size is known
3. **Database pooling**: Enable `PrepareStmt: true` in GORM, configure `SetMaxIdleConns`, `SetMaxOpenConns`
4. **HTTP pooling**: Configure `Transport` with `MaxIdleConns`, `MaxIdleConnsPerHost`
5. **Buffer pooling**: Use `sync.Pool` for large temporary buffers
6. **Batch operations**: Use `CreateInBatches` for bulk database inserts
7. **WAL mode**: Enable `PRAGMA journal_mode=WAL`, `PRAGMA synchronous=NORMAL`
8. **Worker pools**: Use goroutine pools for parallel independent work
9. **Typed structs over maps**: Prefer typed structs for RPC params to reduce allocations
10. **Lock-free batching**: Use channels with buffered batching for high-frequency messages

See `.github/copilot-instructions.md` for detailed performance principles with benchmarks.

## Frontend (website/)

### Tech Stack
- Next.js 15+ with App Router, **Static Site Generation** (`output: 'export'`)
- Tailwind CSS v4 + shadcn/ui (Radix UI components)
- TanStack Query v5 for data fetching
- Lucide React icons

### Build Output
- `npm run build` → `out/` directory (static HTML/JS/CSS)
- Copied to `.build/package/website/` for Go service
- **All paths must be relative** (no absolute paths for Go sub-route compatibility)

### RPC Client Pattern
```typescript
// lib/rpc-client.ts
POST /restful/rpc with {method, target, params}
Automatic case conversion: camelCase ↔ snake_case
Mock mode: NEXT_PUBLIC_USE_MOCK_DATA=true
```

### Frontend Rules
- NO `next/headers` or server-side features requiring Node.js
- Dynamic routes require `generateStaticParams()` or use client-side navigation only
- TypeScript types in `lib/types.ts` mirror Go structs (camelCase naming)

## Service Documentation

Each service has a `service.md` file documenting its RPC API:
- `cmd/broker/service.md` - Broker process management and routing
- `cmd/access/service.md` - REST gateway endpoints
- `cmd/local/service.md` - CVE/CWE/CAPEC/ATT&CK data storage (50+ RPC methods)
- `cmd/remote/service.md` - External API fetching
- `cmd/meta/service.md` - Job orchestration with go-taskflow
- `cmd/sysmon/service.md` - System monitoring

**Always update `service.md` when adding RPC handlers.**

## Configuration

### Environment Variables
- `SESSION_DB_PATH` - BoltDB session storage (default: session.db)
- `CVE_DB_PATH` - SQLite CVE database
- `CWE_DB_PATH` - SQLite CWE database
- `CAPEC_DB_PATH` - SQLite CAPEC database
- `ATTACK_DB_PATH` - SQLite ATT&CK database
- `CAPEC_STRICT_XSD` - Enable XSD validation for CAPEC imports

### Broker Configuration
`config.json` controls:
- Process definitions with IDs, commands, arguments, restart policies
- Transport mode (`uds` or `fd_pipe`)
- Optimizer parameters: buffer capacity, worker count, batch size, flush interval, offer policy
- Logging configuration and output destinations

### vconfig Tool
Run `./build.sh -c` to access the TUI configuration manager for build flags and default config generation.

## Commit Workflow

Per `.github/agents/v2e-go.agent.md`:
1. **Commit at each milestone** - incremental, clean commits
2. **Exclude binaries and databases** - never commit `.db`, `.db-wal`, `.db-shm`, compiled binaries
3. **Update existing `service.md`** - only update, never create new markdown files
4. **Include tests** - table-driven unit tests and `testing.B` benchmarks for performance-critical paths

## Job Session Management

The meta service orchestrates CVE/CWE data fetching jobs using go-taskflow with BoltDB persistence:

**Job States**: Queued → Running → (Paused | Completed | Failed | Stopped)

**Single Active Run Policy**: Only one job run can be active at a time.

**Auto-Recovery**: Running jobs resume on service restart; paused jobs remain paused.

**Session Control RPCs**: `RPCStartSession`, `RPCStopSession`, `RPCPauseJob`, `RPCResumeJob`, `RPCGetSessionStatus`

## Database Migrations

Located in `tool/migrations/` with:
- `README.md` - Migration guide
- `RUNBOOK.md` - Verification and rollback procedures
- SQL migrations numbered `0001_*.sql`, `0002_*.sql`, etc.

## Key Locations

| Purpose | Location |
|---------|----------|
| Broker core | `cmd/broker/` |
| Subprocess framework | `pkg/proc/subprocess/` |
| Message handling | `pkg/proc/message/` |
| Taskflow jobs | `pkg/cve/taskflow/` |
| Frontend | `website/` |
| Service RPC specs | `cmd/*/service.md` |
| Migrations | `tool/migrations/` |
| Assets (data files) | `assets/` |
