
# Sysmon Service

## Service Type
RPC (stdin/stdout message passing)

## Description
The Sysmon service is responsible for monitoring system performance metrics and exposing them via RPC interfaces. It collects data such as CPU usage, memory consumption, load average, disk usage, network statistics, and swap usage using the procfs interface, and makes this information available to other services or the frontend.

## Available RPC Methods

### 1. RPCGetSysMetrics
- **Description**: Retrieves the current system performance metrics.
- **Request Parameters**: None
- **Response**:
  - `cpu_usage` (float): The percentage of CPU usage.
  - `memory_usage` (float): The percentage of memory usage.
  - `load_avg` (array): Array of load averages for 1, 5, and 15 minutes.
  - `uptime` (float): System uptime in seconds.
  - `disk_usage` (uint64): Used disk space in bytes.
  - `disk_total` (uint64): Total disk space in bytes.
  - `disk` (object): Detailed disk usage by mount point (e.g., {"/": {"used": 123456, "total": 789012}}).
  - `swap_usage` (float): The percentage of swap usage.
  - `net_rx` (uint64): Total received network traffic in bytes.
  - `net_tx` (uint64): Total transmitted network traffic in bytes.
  - `network` (object): Detailed network statistics by interface.
  - `message_stats` (object): Message statistics from the broker if available.
- **Errors**:
  - `ServiceUnavailable`: The service is unable to collect metrics at the moment.
  - `InternalError`: An unexpected error occurred while processing the request.

### 2. RPCGetDiskUsage
- **Description**: Retrieves detailed disk usage information for all mount points.
- **Request Parameters**: None
- **Response**:
  - `disk` (object): Detailed disk usage by mount point
    - Each mount point contains:
      - `used` (uint64): Used space in bytes
      - `total` (uint64): Total space in bytes
      - `used_percent` (float): Used percentage
- **Errors**:
  - `InternalError`: Failed to read disk statistics.

### 3. RPCGetNetworkStats
- **Description**: Retrieves network statistics for all interfaces.
- **Request Parameters**: None
- **Response**:
  - `network` (object): Network statistics by interface
    - Each interface contains:
      - `rx_bytes` (uint64): Received bytes
      - `tx_bytes` (uint64): Transmitted bytes
      - `rx_packets` (uint64): Received packets
      - `tx_packets` (uint64): Transmitted packets
      - `rx_errors` (uint64): Receive errors
      - `tx_errors` (uint64): Transmit errors
  - `total_rx` (uint64): Total received bytes across all interfaces
  - `total_tx` (uint64): Total transmitted bytes across all interfaces
- **Errors**:
  - `InternalError`: Failed to read network statistics.

### 4. RPCGetMemoryInfo
- **Description**: Retrieves detailed memory usage information.
- **Request Parameters**: None
- **Response**:
  - `total` (uint64): Total memory in bytes.
  - `available` (uint64): Available memory in bytes.
  - `used` (uint64): Used memory in bytes.
  - `usage_percent` (float): Memory usage percentage.
  - `buffers` (uint64): Buffer memory.
  - `cached` (uint64): Cached memory.
  - `swap_total` (uint64): Total swap space.
  - `swap_free` (uint64): Free swap space.
  - `swap_used` (uint64): Used swap space.
  - `swap_usage_percent` (float): Swap usage percentage.
- **Errors**:
  - `InternalError`: Failed to read memory information.

### 5. RPCGetLoadAverage
- **Description**: Retrieves system load averages.
- **Request Parameters**: None
- **Response**:
  - `load1` (float): Load average for the last 1 minute.
  - `load5` (float): Load average for the last 5 minutes.
  - `load15` (float): Load average for the last 15 minutes.
  - `runnable_tasks` (int): Number of runnable tasks.
  - `existing_tasks` (int): Total number of existing tasks.
- **Errors**:
  - `InternalError`: Failed to read load average.

### 6. RPCGetUptime
- **Description**: Retrieves system uptime information.
- **Request Parameters**: None
- **Response**:
  - `uptime` (float): System uptime in seconds.
  - `idle_time` (float): Total idle time in seconds.
- **Errors**: None

### 7. RPCCPUUsage
- **Description**: Retrieves detailed CPU usage statistics.
- **Request Parameters**: None
- **Response**:
  - `user` (float): CPU time spent in user mode.
  - `nice` (float): CPU time spent in nice mode.
  - `system` (float): CPU time spent in system mode.
  - `idle` (float): CPU time spent in idle mode.
  - `iowait` (float): CPU time spent waiting for I/O.
  - `irq` (float): CPU time spent servicing interrupts.
  - `softirq` (float): CPU time spent servicing soft interrupts.
  - `steal` (float): CPU time stolen by hypervisor.
  - `guest` (float): CPU time spent running a virtual CPU.
  - `guest_nice` (float): CPU time spent running a niced guest.
  - `usage_percent` (float): Overall CPU usage percentage.
- **Errors**:
  - `InternalError`: Failed to read CPU statistics.

### 8. RPCGetProcessStats
- **Description**: Retrieves process and task statistics.
- **Request Parameters**: None
- **Response**:
  - `processes_running` (int): Number of processes in running state.
  - `processes_sleeping` (int): Number of processes in sleeping state.
  - `processes_stopped` (int): Number of processes in stopped state.
  - `processes_zombie` (int): Number of zombie processes.
  - `threads` (int): Total number of threads.
  - `procs_blocked` (int): Number of processes blocked waiting for I/O.
- **Errors**:
  - `InternalError`: Failed to read process statistics.

### 9. RPCGetSystemInfo
- **Description**: Retrieves general system information.
- **Request Parameters**: None
- **Response**:
  - `hostname` (string): System hostname.
  - `kernel` (string): Kernel version.
  - `os` (string): Operating system.
  - `platform` (string): Hardware platform.
  - `processors` (int): Number of processors.
  - `boot_time` (string): Boot time timestamp.
- **Errors**: None

---

## Notes
- The Sysmon service is designed to be lightweight and efficient, ensuring minimal impact on system performance while collecting metrics.
- It adheres to the broker-first architecture, meaning it is spawned and managed by the broker process.
- All communication is broker-mediated, ensuring secure and reliable message passing.
- The service can query the broker for message statistics via RPC and include them in the response.
- Metrics collection uses the procfs interface for accurate system statistics.
- All metrics are cached for a short period to reduce system call overhead.

## Dependencies
- **Subprocess Framework**: Utilizes the `pkg/proc/subprocess` package for lifecycle management and logging.
- **Procfs Package**: Uses the `pkg/common/procfs` package to gather CPU, memory, load average, disk, network, and swap metrics.

## Configuration
- The service reads its configuration from the `config.json` file managed by the broker.
- Logging and other runtime parameters are inherited from the broker's configuration.
- Cache duration for metrics can be configured via `sysmon_cache_duration` (default: 1 second).

## Testing
- Unit tests are located in the `cmd/sysmon` directory.
- Integration tests ensure the service works correctly within the broker-first architecture.
- Use the `pytest` framework to validate the service's functionality.
