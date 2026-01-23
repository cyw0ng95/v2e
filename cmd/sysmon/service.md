# Sysmon Service

## Overview
The Sysmon service is responsible for monitoring system performance metrics and exposing them via RPC interfaces. It collects data such as CPU usage, memory consumption, load average, disk usage, network statistics, and swap usage using the procfs interface, and makes this information available to other services or the frontend.

## Service Type
- **Type**: RPC
- **Description**: Provides system performance metrics to clients via RPC calls.

## Available RPC Methods

### RPCGetSysMetrics
- **Description**: Retrieves the current system performance metrics.
- **Request Parameters**: None
- **Response**:
  - `cpu_usage` (float): The percentage of CPU usage.
  - `memory_usage` (float): The percentage of memory usage.
  - `load_avg` (array): Array of load averages for 1, 5, and 15 minutes.
  - `uptime` (float): System uptime in seconds.
  - `disk_usage` (uint64): Used disk space in bytes.
  - `disk_total` (uint64): Total disk space in bytes.
  - `disk` (object): Detailed disk usage by mount point (e.g., `{"/": {"used": 123456, "total": 789012}}`).
  - `swap_usage` (float): The percentage of swap usage.
  - `net_rx` (uint64): Total received network traffic in bytes.
  - `net_tx` (uint64): Total transmitted network traffic in bytes.
  - `network` (object): Detailed network statistics by interface.
  - `message_stats` (object): Message statistics from the broker if available.
- **Errors**:
  - `ServiceUnavailable`: The service is unable to collect metrics at the moment.
  - `InternalError`: An unexpected error occurred while processing the request.
- **Example**:
  - **Request**:
    ```json
    {
      "method": "RPCGetSysMetrics",
      "params": {}
    }
    ```
  - **Response**:
    ```json
    {
      "cpu_usage": 45.3,
      "memory_usage": 67.8,
      "load_avg": [1.2, 0.8, 0.5],
      "uptime": 3600,
      "disk_usage": 1234567890,
      "disk_total": 9876543210,
      "disk": {"/": {"used": 1234567890, "total": 9876543210}},
      "swap_usage": 10.2,
      "net_rx": 543210,
      "net_tx": 987654,
      "network": {"eth0": {"rx": 271605, "tx": 493827}},
      "message_stats": {
        "total": {
          "total_sent": 100,
          "total_received": 95,
          "request_count": 50,
          "response_count": 45,
          "event_count": 20,
          "error_count": 5,
          "first_message_time": "2023-01-01T00:00:00Z",
          "last_message_time": "2023-01-01T01:00:00Z"
        },
        "per_process": {
          "local": {
            "total_sent": 30,
            "total_received": 25,
            "request_count": 15,
            "response_count": 10,
            "event_count": 5,
            "error_count": 2,
            "first_message_time": "2023-01-01T00:00:00Z",
            "last_message_time": "2023-01-01T01:00:00Z"
          }
        }
      }
    }
    ```

## Notes
- The Sysmon service is designed to be lightweight and efficient, ensuring minimal impact on system performance while collecting metrics.
- It adheres to the broker-first architecture, meaning it is spawned and managed by the broker process.
- All communication is broker-mediated, ensuring secure and reliable message passing.
- The service can query the broker for message statistics via RPC and include them in the response.
- Metrics collection uses the procfs interface for accurate system statistics.

## Dependencies
- **Subprocess Framework**: Utilizes the `pkg/proc/subprocess` package for lifecycle management and logging.
- **Procfs Package**: Uses the `pkg/common/procfs` package to gather CPU, memory, load average, disk, network, and swap metrics.

## Configuration
- The service reads its configuration from the `config.json` file managed by the broker.
- Logging and other runtime parameters are inherited from the broker's configuration.

## Testing
- Unit tests are located in the `cmd/sysmon` directory.
- Integration tests ensure the service works correctly within the broker-first architecture.
- Use the `pytest` framework to validate the service's functionality.