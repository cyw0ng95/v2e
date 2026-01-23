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
- **Description**: Spawns a subprocess with RPC support (custom file descriptors for message passing)
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

### 3. RPCSpawnWithRestart
- **Description**: Spawns a subprocess with auto-restart capability
- **Request Parameters**:
  - `id` (string, required): Unique identifier for the process
  - `command` (string, required): Command to execute
  - `max_restarts` (int, optional): Maximum number of restart attempts (-1 for unlimited, default: -1)
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
  - **Request**: {"id": "worker-1", "command": "./worker", "max_restarts": 5, "args": ["--config", "config.json"]}
  - **Response**: {"id": "worker-1", "pid": 12345, "status": "running"}

### 4. RPCSpawnRPCWithRestart
- **Description**: Spawns a subprocess with RPC support and auto-restart capability
- **Request Parameters**:
  - `id` (string, required): Unique identifier for the process
  - `command` (string, required): Command to execute
  - `max_restarts` (int, optional): Maximum number of restart attempts (-1 for unlimited, default: -1)
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
  - **Request**: {"id": "worker-1", "command": "./worker", "max_restarts": 5, "args": ["--config", "config.json"]}
  - **Response**: {"id": "worker-1", "pid": 12345, "status": "running"}

### 5. RPCGetMessageStats
- **Description**: Retrieves message statistics for the broker and all managed processes
- **Request Parameters**: None
- **Response**:
  - `total` (object): Overall message statistics for the broker
    - `total_sent` (int): Total messages sent by the broker
    - `total_received` (int): Total messages received by the broker
    - `request_count` (int): Number of request messages processed
    - `response_count` (int): Number of response messages processed
    - `event_count` (int): Number of event messages processed
    - `error_count` (int): Number of error messages processed
    - `first_message_time` (string): Time of first message (RFC3339 format)
    - `last_message_time` (string): Time of last message (RFC3339 format)
  - `per_process` (object): Message statistics broken down by process ID
- **Errors**: None
- **Example**:
  - **Request**: {}
  - **Response**: {"total": {"total_sent": 100, "total_received": 95, ...}, "per_process": {"local": {...}, "remote": {...}}}

### 6. RPCGetMessageCount
- **Description**: Retrieves the total number of messages processed by the broker
- **Request Parameters**: None
- **Response**:
  - `count` (int): Total number of messages processed (sent + received)
- **Errors**: None
- **Example**:
  - **Request**: {}
  - **Response**: {"count": 195}

## Configuration
- **Log File**: Configurable via `config.json` under `broker.log_file` for dual output (stdout + file)
- **Log Level**: Configurable via `config.json` under `logging.level` (debug, info, warn, error)
- **Process Management**: Processes can be configured to auto-restart with configurable max restarts
- **RPC File Descriptors**: Custom file descriptor numbers for RPC communication can be configured via `proc.rpc_input_fd`, `proc.rpc_output_fd`, `broker.rpc_input_fd`, or `broker.rpc_output_fd`

## Notes
- Uses custom file descriptors (typically fd 3 and 4) for RPC communication to avoid conflicts with stdio
- Manages subprocess lifecycles with optional auto-restart capability
- Maintains message statistics for monitoring and debugging
- Routes messages between services using a correlation ID mechanism for request-response matching
- Supports graceful shutdown of all managed processes
- Handles process restart policies with configurable limits