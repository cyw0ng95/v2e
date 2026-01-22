# Sysmon Service

## Overview
The Sysmon service is responsible for monitoring system performance metrics and exposing them via RPC interfaces. It collects data such as CPU usage and memory consumption using the procfs interface, and makes this information available to other services or the frontend.

## Service Type
- **Type**: RPC
- **Description**: Provides system performance metrics to clients via RPC calls.

## Available RPC Methods

### RPCGetSysMetrics
- **Description**: Retrieves the current system performance metrics.
- **Request Parameters**: None
- **Response**:
  - `cpuUsage` (float): The percentage of CPU usage.
  - `memoryUsage` (float): The percentage of memory usage.
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
      "cpuUsage": 45.3,
      "memoryUsage": 67.8
    }
    ```

## Notes
- The Sysmon service is designed to be lightweight and efficient, ensuring minimal impact on system performance while collecting metrics.
- It adheres to the broker-first architecture, meaning it is spawned and managed by the broker process.
- All communication is broker-mediated, ensuring secure and reliable message passing.

## Dependencies
- **Subprocess Framework**: Utilizes the `pkg/proc/subprocess` package for lifecycle management and logging.
- **Procfs Package**: Uses the `pkg/common/procfs` package to gather CPU and memory metrics.

## Configuration
- The service reads its configuration from the `config.json` file managed by the broker.
- Logging and other runtime parameters are inherited from the broker's configuration.

## Testing
- Unit tests are located in the `cmd/sysmon` directory.
- Integration tests ensure the service works correctly within the broker-first architecture.
- Use the `pytest` framework to validate the service's functionality.