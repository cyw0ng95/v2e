# Access Service

## Service Type
REST (HTTP/JSON)

## Description
RESTful API gateway service that provides external access to the v2e system. Forwards RPC requests to backend services through the broker.

## Available REST Endpoints

### 1. GET /restful/health
- **Description**: Health check endpoint to verify service is running
- **Request Parameters**: None
- **Response**:
  - `status` (string): "ok" if service is healthy
- **Errors**: None
- **Example**:
  - **Request**: GET /restful/health
  - **Response**: `{"status": "ok"}`

### 2. POST /restful/rpc
- **Description**: Generic RPC forwarding endpoint that routes requests to backend services
- **Request Parameters**:
  - `method` (string, required): RPC method name (e.g., "RPCGetCVE")
  - `params` (object, optional): Parameters to pass to the RPC method
  - `target` (string, optional): Target process ID (default: "broker")
- **Response**:
  - `retcode` (int): 0 for success, non-zero for errors
  - `message` (string): Success message or error description
  - `payload` (object): Response data from backend service
- **Errors**:
  - Invalid JSON: `retcode=400`, missing or malformed request body
  - RPC timeout: `retcode=500`, backend service did not respond in time
  - Backend error: `retcode=500`, backend service returned an error
- **Example**:
  - **Request**: `{"method": "RPCGetCVE", "target": "local", "params": {"id": "CVE-2021-44228"}}`
  - **Response**: `{"retcode": 0, "message": "success", "payload": {...}}`

## Configuration
- **RPC Timeout**: Configurable via `config.json` under `access.rpc_timeout_seconds` (default: 30 seconds)
- **Shutdown Timeout**: Configurable via `config.json` under `access.shutdown_timeout_seconds` (default: 10 seconds)
- **Static Directory**: Configurable via `config.json` under `access.static_dir` (default: "website")
- **Server Address**: Configurable via `config.json` under `server.address` (default: "0.0.0.0:8080")

## Notes
- Forwards all RPC calls to the broker for routing
- Handles authentication and request validation
- Provides RESTful API for external clients
- Returns standardized error responses
- Runs as a subprocess managed by the broker
- Uses stdin/stdout for RPC communication with the broker
