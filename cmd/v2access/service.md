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
  - **Response**: {"status": "ok"}

### 2. POST /restful/rpc
- **Description**: Generic RPC forwarding endpoint that routes requests to backend services
- **Request Parameters**:
  - `method` (string, required): RPC method name (e.g., "RPCGetCVE")
  - `params` (object, optional): Parameters to pass to the RPC method
  - `target` (string, optional): Target process ID (default: "broker")
- **Response**:
  - `retcode` (int): 0 for success, non-zero for errors
  - `message` (string): Success message or error description
  - `payload` (object): Response data from the backend service
- **Errors**:
  - Invalid JSON: `retcode=400`, missing or malformed request body
  - RPC timeout: `retcode=500`, backend service did not respond in time
  - Backend error: `retcode=500`, backend service returned an error

## Configuration
- **RPC Timeout**: Configurable via `config.json` under `access.rpc_timeout_seconds` (default: 30 seconds)
- **Shutdown Timeout**: Configurable via `config.json` under `access.shutdown_timeout_seconds` (default: 10 seconds)
- **Static Directory**: Configurable via `config.json` under `access.static_dir` (default: "website")
- **Server Address**: Configurable via `config.json` under `server.address` (default: "0.0.0.0:8080")

# Access Service

## Service Type
RPC/REST (HTTP + message passing)

## Description
Acts as the main entry point for all external API requests. Handles authentication, request routing, and response formatting. Forwards RPC calls to the correct backend service and returns results to the client.

## Available RPC Methods

### 1. RPCEcho
- **Description**: Returns the input string for testing connectivity
- **Request Parameters**:
  - `message` (string, required): The message to echo
- **Response**:
  - `message` (string): The echoed message
- **Errors**:
  - None

### 2. RPCProxy
- **Description**: Proxies an RPC call to a target backend service
- **Request Parameters**:
  - `target` (string, required): Target service name
  - `method` (string, required): RPC method to call
  - `params` (object, optional): Parameters for the RPC call
- **Response**:
  - `result` (object): The result from the backend service
- **Errors**:
  - Invalid target or method: Target service or method does not exist
  - RPC error: Error returned from backend service

### 3. RPCGetServiceList
- **Description**: Returns a list of available backend services
- **Request Parameters**: None
- **Response**:
  - `services` ([]string): List of service names
- **Errors**:
  - None

### 4. RPCGetServiceSpec
- **Description**: Returns the API specification for a given service
- **Request Parameters**:
  - `service` (string, required): Service name
- **Response**:
  - `spec` (object): The API specification for the service
- **Errors**:
  - Service not found: The specified service does not exist

### 5. RPCGetHealth
- **Description**: Returns health status of the access service
- **Request Parameters**: None
- **Response**:
  - `status` (string): Health status (e.g., "ok")
- **Errors**:
  - None

### 6. RPCGetVersion
- **Description**: Returns the version of the access service
- **Request Parameters**: None
- **Response**:
  - `version` (string): Version string
- **Errors**:
  - None

---

## Notes
- Forwards all RPC calls to the broker for routing
- Handles authentication and request validation
- Provides RESTful API for external clients
- Returns standardized error responses