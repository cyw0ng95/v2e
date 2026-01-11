# Copilot Instructions for v2e

## Architecture Principles

**CRITICAL: Broker-First Architecture**

The broker (`cmd/broker`) is a standalone process that boots and manages all subprocesses in the system.

**Strict Rules:**
- The broker is the ONLY process that can spawn and manage subprocesses
- The `access` service is a subprocess - it provides REST API but does NOT manage processes
- Subprocesses must NEVER embed broker logic or create their own broker instances
- All inter-process communication must go through broker-mediated RPC messages
- Never add process management capabilities to subprocesses

This architecture is fundamental to the system design. Violating it will cause circular dependencies and architectural problems.

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
