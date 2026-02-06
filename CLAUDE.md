# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

v2e (Vulnerabilities Viewer Engine) is a broker-first microservices system for managing CVE, CWE, CAPEC, and ATT&CK security data. The architecture enforces a central broker pattern where `cmd/broker` is the sole process manager and message router - no subprocess-to-subprocess communication is allowed.

## Build & Development Commands

**IMPORTANT: All build and test operations MUST use `build.sh` wrapper script.** Do not use direct `go build` or `go test` commands - the wrapper handles build tags, environment setup, and proper test configuration.

```bash
# Primary build script (REQUIRED for all builds/tests)
./build.sh -t     # Unit tests (excludes fuzz tests, uses -run='^Test')
./build.sh -f     # Fuzz tests (5 seconds per target)
./build.sh -m     # Benchmarks with reporting
./build.sh -p     # Build and package binaries + assets
./build.sh -r     # Run full system: broker + all subprocesses + frontend dev server
./build.sh -c     # Run vconfig TUI for configuration

# Development and testing with ./build.sh -r
# This command starts:
#   - Broker with all subprocesses (access, local, meta, remote, sysmon)
#   - Frontend dev server (npm run dev in website/)
#   - All services run together for integration testing and log analysis
# Press Ctrl+C to stop all services cleanly

# Containerized development (macOS always uses container)
./runenv.sh -t    # Run tests in container
./runenv.sh -r    # Dev mode in container

# Frontend (website/)
cd website
npm run dev       # Development server
npm run build     # Static export to out/
npm run lint      # ESLint
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
4. **Use `build.sh` wrapper** - all builds and tests must use `./build.sh`, never direct `go build`/`go test`
5. **Document RPC APIs in `service.md`** - every RPC handler must be documented in the service's `service.md` file
6. **NO remote API calls in tests** - tests must not access NVD, GitHub, or external services (use mocks/fixtures)
7. **Tests must be fast** - unit tests should run in milliseconds; slow tests hurt developer experience
8. **NO new documentation files** - only update existing `README.md` or `cmd/*/service.md`. Never create new markdown files.

### Transport Layer
- **Transport**: Unix Domain Sockets (UDS) with 0600 permissions
- Message types: Request, Response, Event, Error (with correlation IDs)

### Key Packages
- `pkg/proc/subprocess` - Subprocess lifecycle framework, provides `SetupLogging()`, `RunWithDefaults()`, `RegisterHandler()`
- `pkg/proc/message` - Message types with `sync.Pool` optimization
- `cmd/broker/core` - Broker orchestrator, Process, Router interfaces
- `cmd/broker/transport` - UDS transport implementation
- `cmd/broker/perf` - Optimizer with adaptive tuning, message batching, worker pools
- `pkg/cve/taskflow` - Taskflow-based job executor with BoltDB persistence

## Subprocess Development Pattern

All subprocesses under `cmd/*` must follow this pattern:

```go
import "github.com/cyw0ng95/v2e/pkg/proc/subprocess"

func main() {
    logger := subprocess.SetupLogging("my-service")
    sp := subprocess.NewSubprocess(logger)

    sp.RegisterHandler("RPCMyMethod", func(params json.RawMessage) (interface{}, error) {
        // Handler implementation
    })

    subprocess.RunWithDefaults(sp, logger)
}
```

## Testing Guidelines

**CRITICAL TEST REQUIREMENTS:**

1. **NO REMOTE API ACCESS** - Tests must NOT access NVD, GitHub APIs, or any external services. Remote API calls cause flaky tests due to rate limiting, network issues, and service unavailability. Use mocks, fixtures, or local test data instead.

2. **TESTS MUST BE FAST** - Unit tests should run in milliseconds. Slow tests degrade developer experience and CI feedback. If a test is slow, consider:
   - Using in-memory databases instead of disk I/O
   - Reducing test data size
   - Mocking expensive operations
   - Moving long-running tests to a separate benchmark or integration suite

### Unit Tests
- Located alongside source (`*_test.go`)
- Pattern: `Test*` functions only (use `-run='^Test'` to exclude fuzz tests)
- Table-driven tests preferred
- Run with `-race` flag
- **NO remote API calls** - use mocks/fixtures only

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
- **NO remote API calls** - use test fixtures and mock data

### Running Full System for Testing
Use `./build.sh -r` to run the complete system for integration testing and log analysis:
- Broker spawns all subprocesses (access, local, meta, remote, sysmon)
- Frontend dev server runs on http://localhost:3000
- Logs are written to `.build/package/logs/` for analysis
- Press Ctrl+C to stop all services cleanly
- This is the recommended way to verify changes before committing

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

### Browser Testing with Playwright MCP

When testing the website frontend, use the Playwright MCP tools available in Claude Code to navigate through pages and capture bugs through console logs.

**Start the dev server first:**
```bash
./build.sh -r    # Starts broker + frontend on http://localhost:3000
```

**Common Playwright MCP testing patterns:**

1. **Navigate to a page:**
   - Use `browser_navigate` with URL (e.g., `http://localhost:3000`)

2. **Capture page state:**
   - Use `browser_snapshot` to get accessibility tree (better than screenshots for debugging)
   - Use `browser_take_screenshot` for visual verification

3. **Check console for errors:**
   - Use `browser_console_messages` with level `error` or `warning` to capture JavaScript errors
   - Look for failed RPC calls, unhandled promises, missing assets

4. **Interact with elements:**
   - Use `browser_click` with element ref from snapshot
   - Use `browser_type` to fill forms
   - Use `browser_select_option` for dropdowns

5. **Wait for conditions:**
   - Use `browser_wait_for` with text or time
   - Use `browser_evaluate` to run custom JS to check state

**Bug capture workflow:**
```
1. Navigate to page
2. Take snapshot for baseline
3. Interact (click, type, etc.)
4. Check console messages (errors/warnings)
5. Take screenshot/snapshot of broken state
6. Use browser_evaluate to inspect React state or network requests
```

**Console log analysis:**
- `error` level: JavaScript errors, failed RPC calls, missing imports
- `warning` level: Deprecated APIs, React warnings, failed asset loads
- `info/debug` level: RPC request/response logs, state changes

## Service Documentation

**CRITICAL: All RPC APIs MUST be documented in `service.md` inside each `cmd/*/` directory.**

Every service must have a `service.md` file that documents its complete RPC API specification:
- `cmd/broker/service.md` - Broker process management and routing
- `cmd/access/service.md` - REST gateway endpoints
- `cmd/local/service.md` - CVE/CWE/CAPEC/ATT&CK data storage
- `cmd/remote/service.md` - External API fetching
- `cmd/meta/service.md` - Job orchestration with go-taskflow
- `cmd/sysmon/service.md` - System monitoring

**RPC API Documentation Requirements:**
1. **Mandatory for all RPC methods** - Every `RPC*` handler must be documented
2. **Include complete specification**:
   - Method name and description
   - Request parameters (name, type, required/optional, description)
   - Response fields (name, type, description)
   - Error conditions and messages
   - Example request/response
3. **Update before or with code** - Documentation must be updated when adding or modifying RPC handlers
4. **Single source of truth** - The `service.md` file is the authoritative API specification

**Always update `service.md` when adding RPC handlers. Never commit new RPC methods without documentation.**

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
- Optimizer parameters: buffer capacity, worker count, batch size, flush interval, offer policy
- Logging configuration and output destinations

### vconfig Tool
Run `./build.sh -c` to access the TUI configuration manager for build flags and default config generation.

## Commit Workflow

Per `.github/agents/v2e-go.agent.md`:
1. **Use `build.sh` for all builds and tests** - never use direct `go build` or `go test`
2. **Commit in-time (frequently, at logical milestones)** - make incremental, clean commits after each meaningful change. Don't batch unrelated changes or wait until "everything is done."
   - Commit after adding/changing RPC handlers (with updated `service.md`)
   - Commit after adding tests (separate commit from implementation)
   - Commit after completing a logical unit of work
3. **Exclude binaries and databases** - never commit `.db`, `.db-wal`, `.db-shm`, compiled binaries
4. **Document all RPC APIs in `service.md`** - update existing `service.md` files before committing new RPC handlers
5. **Include tests** - table-driven unit tests and `testing.B` benchmarks for performance-critical paths
6. **NO flaky remote API tests** - never add tests that access NVD, GitHub, or other external services
7. **NO new documentation files** - only update existing `README.md` or `cmd/*/service.md`. Never create new markdown files (DESIGN.md, TODO.md, etc.)

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
| Runtime logs | `.build/package/logs/` |
