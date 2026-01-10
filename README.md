# v2e

A Go-based project demonstrating a multi-command structure with CVE (Common Vulnerabilities and Exposures) data fetching capabilities.

## Project Structure

This project contains multiple commands:

- `cmd/server` - A simple HTTP server with CVE API integration
- `cmd/client` - A simple HTTP client

And packages:

- `pkg/common` - Common utilities and configuration
- `pkg/repo` - Repository layer for external data sources (NVD CVE API)

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

#### API Endpoints

- `GET /` - Server information
- `GET /cve/{cve-id}` - Fetch CVE data from the National Vulnerability Database (NVD)

Example:
```bash
curl http://localhost:8080/cve/CVE-2021-44228
```

### Client

```bash
go run ./cmd/client [url]
```

If no URL is provided, it will connect to `http://localhost:8080` by default.

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

## Development

### Dependencies

This project uses Go modules for dependency management:

```bash
go mod tidy
go mod download
```

Key dependencies:
- [go-resty/resty](https://github.com/go-resty/resty) - HTTP client library for making API requests

### Testing

Run tests:

```bash
go test ./...
```

## License

MIT
