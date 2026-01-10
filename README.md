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
