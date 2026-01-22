# Broker Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Central process manager and message router for the v2e system. Spawns and manages all subprocess services, routes RPC messages between services.

## Available RPC Methods

### 1. RPCSpawn
- **Description**: Spawns a new subprocess with specified command and arguments
- **Request Parameters**:
  - `id` (string, required): Unique identifier for the process
  - `command` (string, required): Command to execute
  - `args` ([]string, optional): Command arguments
- **Response**:
  - `id` (string): Process identifier
  - `pid` (int): Process ID
  - `status` (string): Process status ("running", "exited", "failed")
- **Errors**:
  - Missing ID: Process ID is required
  - Duplicate ID: Process with this ID already exists
  - Spawn failure: Failed to start the process
- **Example**:
  - **Request**: {"id": "worker-1", "command": "./worker", "args": ["--config", "config.json"]}
  - **Response**: {"id": "worker-1", "pid": 12345, "status": "running"}

### 2. RPCSpawnRPC
- **Description**: Spawns a subprocess with RPC support (stdin/stdout pipes for message passing)
- **Request Parameters**:
  - `id` (string, required): Unique identifier for the process
  - `command` (string, required): Command to execute
  - `args` ([]string, optional): Command arguments