
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

### 6. RPCGetMessageCount
- **Description**: Retrieves the total number of messages processed by the broker
- **Request Parameters**: None
- **Response**:
  - `count` (int): Total number of messages processed (sent + received)
- **Errors**: None

### 7. RPCRequestPermits
- **Description**: Requests worker permits from the broker's global pool for concurrent execution
- **Request Parameters**:
  - `provider_id` (string, required): Unique identifier for the requesting provider
  - `permit_count` (int, required): Number of permits requested (must be > 0)
- **Response**:
  - `granted` (int): Number of permits granted (may be less than requested)
  - `available` (int): Total permits still available in the pool
  - `provider_id` (string): The provider ID
- **Errors**:
  - Invalid request: Missing provider_id or permit_count <= 0
  - No permits available: No permits currently available in the pool

### 8. RPCReleasePermits
- **Description**: Returns worker permits to the broker's global pool
- **Request Parameters**:
  - `provider_id` (string, required): Unique identifier for the provider releasing permits
  - `permit_count` (int, required): Number of permits to release (must be > 0)
- **Response**:
  - `success` (bool): true if permits were released successfully
  - `available` (int): Total permits available after release
  - `provider_id` (string): The provider ID
- **Errors**:
  - Invalid request: Missing provider_id or permit_count <= 0
  - Provider not found: No permits allocated to this provider

### 9. RPCOnQuotaUpdate
- **Description**: Event broadcast from broker to providers when permits are revoked due to kernel metrics breaches
- **Request Parameters**: None (this is a broker-initiated event)
- **Response**: Not applicable (event only)
- **Event Payload**:
  - `revoked_permits` (int): Number of permits being revoked globally
  - `reason` (string): Reason for revocation (e.g., "P99 latency exceeded 50ms")
  - `kernel_metrics` (object): Current kernel performance metrics
- **Notes**: Providers should transition to WAITING_QUOTA state when receiving this event

### 10. RPCGetKernelMetrics
- **Description**: Retrieves current kernel performance metrics from the broker
- **Request Parameters**: None
- **Response**:
  - `p99_latency_ms` (float): 99th percentile message routing latency in milliseconds
  - `buffer_saturation` (float): Message buffer saturation percentage (0-100)
  - `active_workers` (int): Number of currently active workers
  - `total_permits` (int): Total permits in the global pool
  - `allocated_permits` (int): Number of permits currently allocated
  - `available_permits` (int): Number of permits available for allocation
  - `message_rate` (float): Messages per second
  - `error_rate` (float): Errors per second
- **Errors**: None

---

## Configuration
- **Log File**: Configurable via `config.json` under `broker.log_file` for dual output (stdout + file)
- **Process Management**: Processes can be configured to auto-restart with configurable max restarts
- **RPC File Descriptors**: Custom file descriptor numbers for RPC communication can be configured via `proc.rpc_input_fd`, `proc.rpc_output_fd`, `broker.rpc_input_fd`, or `broker.rpc_output_fd`

## Notes
- Uses custom file descriptors (typically fd 3 and 4) for RPC communication to avoid conflicts with stdio
- Manages subprocess lifecycles with optional auto-restart capability
- Maintains message statistics for monitoring and debugging
- Routes messages between services using a correlation ID mechanism for request-response matching
- Supports graceful shutdown of all managed processes
- Handles process restart policies with configurable limits

## Implementation Notes (2024-04)
- **Runtime FD Validity Check**: As of April 2024, all subprocesses now perform a runtime check to ensure the input/output file descriptors passed for RPC are valid (not closed or invalid). If an invalid fd is detected, the subprocess logs a fatal error and exits with code 254. This prevents cryptic errors such as `epollwait on fd N failed with 9` and improves diagnosability of broker/subprocess startup issues.

## Benchmarks
_No broker-specific benchmarks for subprocess startup or fd validation as of April 2024. Add here if/when available._