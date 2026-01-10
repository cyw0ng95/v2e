# v2e

A basic Go-based project demonstrating a multi-command structure.

## Project Structure

This project contains multiple commands:

- `cmd/server` - A simple HTTP server
- `cmd/client` - A simple HTTP client

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

## Development

### Dependencies

This project uses Go modules for dependency management:

```bash
go mod tidy
go mod download
```

### Testing

Run tests:

```bash
go test ./...
```

## License

MIT
