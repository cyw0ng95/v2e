
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
