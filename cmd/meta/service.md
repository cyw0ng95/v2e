# CVE Meta Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Orchestrates CVE fetching and storage operations by coordinating between local and remote services. Provides high-level CVE management and job control for continuous data synchronization.

## Available RPC Methods

### CVE Data Operations

#### 1. RPCGetCVE
- **Description**: Retrieves CVE data, checking local storage first, then fetching from remote if not found
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to retrieve
- **Response**:
  - `cve` (object): CVE object with all fields
  - `source` (string): "local" or "remote" indicating data source
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Not found: CVE not found in local or remote sources
  - RPC error: Failed to communicate with backend services
- **Example**:
  - **Request**: {"cve_id": "CVE-2021-44228"}
  - **Response**: {"cve": {"id": "CVE-2021-44228", ...}, "source": "local"}

#### 2. RPCCreateCVE
- **Description**: Creates a new CVE record in local storage by fetching from remote
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to create
- **Response**:
  - `success` (bool): true if created successfully
  - `cve_id` (string): ID of the created CVE
  - `cve` (object): The CVE object that was created
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Not found: CVE not found in remote sources
  - RPC error: Failed to communicate with backend services
  - Storage error: Failed to save to local storage

#### 3. RPCUpdateCVE
- **Description**: Updates an existing CVE record by refetching from remote and updating local storage
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to update
- **Response**:
  - `success` (bool): true if updated successfully
  - `cve_id` (string): ID of the updated CVE
  - `cve` (object): The updated CVE object
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Not found: CVE not found in remote sources
  - RPC error: Failed to communicate with backend services
  - Storage error: Failed to update local storage

#### 4. RPCDeleteCVE
- **Description**: Deletes a CVE record from local storage
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to delete
- **Response**:
  - `success` (bool): true if deleted successfully
  - `cve_id` (string): ID of the deleted CVE
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - RPC error: Failed to communicate with backend services
  - Storage error: Failed to delete from local storage

#### 5. RPCListCVEs
- **Description**: Lists CVE records from local storage with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `cves` ([]object): Array of CVE objects
  - `total` (int): Total number of CVEs in local storage
  - `offset` (int): The offset used
  - `limit` (int): The limit used
- **Errors**:
  - RPC error: Failed to communicate with backend services
  - Storage error: Failed to query local storage

#### 6. RPCCountCVEs
- **Description**: Counts the total number of CVEs in local storage
- **Request Parameters**: None
- **Response**:
  - `count` (int): Total number of CVEs in local storage
- **Errors**:
  - RPC error: Failed to communicate with backend services
  - Storage error: Failed to query local storage

### Job Session Control

#### 7. RPCStartSession
- **Description**: Starts a new CVE fetching session that continuously synchronizes data
- **Request Parameters**:
  - `session_id` (string, required): Unique identifier for the session
  - `start_index` (int, optional): Index to start fetching from (default: 0)
  - `results_per_batch` (int, optional): Number of results per batch (default: 100)
- **Response**:
  - `success` (bool): true if session started successfully
  - `session_id` (string): ID of the started session
  - `state` (string): Current state of the session ("running")
  - `created_at` (string): Timestamp when session was created
- **Errors**:
  - Missing session ID: `session_id` parameter is required
  - Session exists: A session is already running
  - RPC error: Failed to communicate with backend services

#### 8. RPCStopSession
- **Description**: Stops the current CVE fetching session and cleans up resources
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if session stopped successfully
  - `session_id` (string): ID of the stopped session
  - `fetched_count` (int): Number of CVEs fetched during the session
  - `stored_count` (int): Number of CVEs successfully stored during the session
  - `error_count` (int): Number of errors encountered during the session
- **Errors**:
  - No session: No active session to stop
  - RPC error: Failed to communicate with backend services

#### 9. RPCGetSessionStatus
- **Description**: Retrieves the current status of the active CVE fetching session
- **Request Parameters**: None
- **Response**:
  - `has_session` (bool): true if a session exists
  - `session_id` (string): ID of the session (if exists)
  - `state` (string): Current state of the session ("running", "paused", "stopped")
  - `start_index` (int): Index where the session started
  - `results_per_batch` (int): Number of results per batch
  - `created_at` (string): Timestamp when session was created
  - `updated_at` (string): Timestamp when session was last updated
  - `fetched_count` (int): Number of CVEs fetched during the session
  - `stored_count` (int): Number of CVEs successfully stored during the session
  - `error_count` (int): Number of errors encountered during the session
- **Errors**: None (returns empty status if no session exists)

#### 10. RPCPauseJob
- **Description**: Pauses the currently running CVE fetching job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if job paused successfully
  - `state` (string): Current state of the job ("paused")
- **Errors**:
  - No running job: No job is currently running
  - RPC error: Failed to communicate with backend services

#### 11. RPCResumeJob
- **Description**: Resumes a paused CVE fetching job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if job resumed successfully
  - `state` (string): Current state of the job ("running")
- **Errors**:
  - No paused job: No job is currently paused
  - RPC error: Failed to communicate with backend services

#### 12. RPCStartCWEViewJob
- **Description**: Starts a background job to fetch and save CWE views
- **Request Parameters**:
  - `start_index` (int, optional): Index to start fetching from (default: 0)
  - `results_per_page` (int, optional): Number of results per page (default: 100)
- **Response**:
  - `success` (bool): true if job started successfully
  - `session_id` (string): ID of the started job session
- **Errors**:
  - RPC error: Failed to communicate with backend services
  - Import error: Failed to start the import process

#### 13. RPCStopCWEViewJob
- **Description**: Stops a running CWE view job
- **Request Parameters**:
  - `session_id` (string, optional): ID of the session to stop (default: current session)
- **Response**:
  - `success` (bool): true if job stopped successfully
  - `session_id` (string): ID of the stopped job session
- **Errors**:
  - No running job: No job is currently running
  - RPC error: Failed to communicate with backend services

## Configuration
- **Session Database Path**: Configurable via `SESSION_DB_PATH` environment variable (default: "session.db")
- **RPC Timeout**: Fixed at 30 seconds for communication with other services

## Notes
- Orchestrates operations between local and remote services
- Job sessions are persistent (stored in bolt K-V database)
- Only one job session can run at a time
- Session state survives service restarts
- Uses RPC to communicate with local and remote services
- All communication is routed through the broker
- Automatically imports CWE data from "assets/cwe-raw.json" at startup
- Automatically imports CAPEC data from "assets/capec_contents_latest.xml" at startup if not already present
- Recovers running sessions after restart (auto-resumes running sessions, keeps paused sessions paused)