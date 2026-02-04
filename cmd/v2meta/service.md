### Memory Card Operations

#### 20. RPCCreateMemoryCard
- **Description**: Creates a new memory card (delegates to local service)
- **Request Parameters**:
  - `bookmark_id` (int, required): The bookmark ID to associate
  - `front_content` (string, required): Front/question content
  - `back_content` (string, required): Back/answer content
  - `major_class` (string, required): Major class/category
  - `minor_class` (string, required): Minor class/category
  - `status` (string, required): Status (e.g., active, archived)
  - `content` (object, required): TipTap JSON content
  - `card_type` (string, optional): Card type (basic, cloze, etc.)
  - `author` (string, optional): Author
  - `is_private` (bool, optional): Privacy flag
  - `metadata` (object, optional): Additional metadata
- **Response**:
  - `success` (bool): true if created
  - `memory_card` (object): The created memory card
- **Errors**:
  - Missing/invalid parameters
  - RPC/local service error

#### 21. RPCGetMemoryCard
- **Description**: Retrieves a memory card by ID (delegates to local service)
- **Request Parameters**:
  - `id` (int, required): Memory card ID
- **Response**:
  - `memory_card` (object): The memory card
- **Errors**:
  - Not found
  - RPC/local service error

#### 22. RPCUpdateMemoryCard
- **Description**: Updates a memory card by ID (delegates to local service)
- **Request Parameters**:
  - `id` (int, required): Memory card ID
  - Any updatable field (see Create)
- **Response**:
  - `success` (bool): true if updated
  - `memory_card` (object): The updated memory card
- **Errors**:
  - Not found
  - RPC/local service error

#### 23. RPCDeleteMemoryCard
- **Description**: Deletes a memory card by ID (delegates to local service)
- **Request Parameters**:
  - `id` (int, required): Memory card ID
- **Response**:
  - `success` (bool): true if deleted
- **Errors**:
  - Not found
  - RPC/local service error

#### 24. RPCListMemoryCards
- **Description**: Lists memory cards with optional filters and pagination (delegates to local service)
- **Request Parameters**:
  - `bookmark_id` (int, optional): Filter by bookmark
  - `major_class` (string, optional): Filter by major class
  - `minor_class` (string, optional): Filter by minor class
  - `status` (string, optional): Filter by status
  - `author` (string, optional): Filter by author
  - `is_private` (bool, optional): Filter by privacy
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `memory_cards` (array): List of memory cards
  - `total` (int): Total count
  - `offset` (int): Offset used
  - `limit` (int): Limit used
- **Errors**:
  - RPC/local service error

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

#### 8. RPCStartTypedSession
- **Description**: Starts a new typed data fetching session for CVE, CWE, CAPEC, or ATT&CK data
- **Request Parameters**:
  - `session_id` (string, required): Unique identifier for the session
  - `data_type` (string, required): Type of data to fetch - "cve", "cwe", "capec", or "attack"
  - `start_index` (int, optional): Index to start fetching from (default: 0)
  - `results_per_batch` (int, optional): Number of results per batch (default: 100)
  - `params` (object, optional): Additional parameters for the job
- **Response**:
  - `success` (bool): true if session started successfully
  - `session_id` (string): ID of the started session
  - `state` (string): Current state of the session ("running")
  - `data_type` (string): The data type being fetched
  - `created_at` (string): Timestamp when session was created
  - `start_index` (int): Index where fetching started
  - `batch_size` (int): Number of results per batch
  - `params` (object): Additional parameters for the job
- **Errors**:
  - Missing session ID: `session_id` parameter is required
  - Session exists: A session is already running
  - Invalid data type: `data_type` must be one of "cve", "cwe", "capec", or "attack"
  - RPC error: Failed to communicate with backend services

#### 9. RPCStopSession
- **Description**: Stops the current data fetching session and cleans up resources
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if session stopped successfully
  - `session_id` (string): ID of the stopped session
  - `fetched_count` (int): Number of items fetched during the session
  - `stored_count` (int): Number of items successfully stored during the session
  - `error_count` (int): Number of errors encountered during the session
- **Errors**:
  - No session: No active session to stop
  - RPC error: Failed to communicate with backend services

#### 10. RPCGetSessionStatus
- **Description**: Retrieves the current status of the active data fetching session
- **Request Parameters**: None
- **Response**:
  - `has_session` (bool): true if a session exists
  - `session_id` (string): ID of the session (if exists)
  - `state` (string): Current state of the session ("running", "paused", "stopped")
  - `data_type` (string): Type of data being fetched ("cve", "cwe", "capec", "attack")
  - `start_index` (int): Index where the session started
  - `results_per_batch` (int): Number of results per batch
  - `created_at` (string): Timestamp when session was created
  - `updated_at` (string): Timestamp when session was last updated
  - `fetched_count` (int): Number of items fetched during the session
  - `stored_count` (int): Number of items successfully stored during the session
  - `error_count` (int): Number of errors encountered during the session
  - `error_message` (string, optional): Error message if session failed
  - `progress` (object, optional): Progress details per data type
- **Errors**: None (returns empty status if no session exists)

#### 11. RPCPauseJob
- **Description**: Pauses the currently running data fetching job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if job paused successfully
  - `state` (string): Current state of the job ("paused")
- **Errors**:
  - No running job: No job is currently running
  - RPC error: Failed to communicate with backend services

#### 12. RPCResumeJob
- **Description**: Resumes a paused data fetching job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if job resumed successfully
  - `state` (string): Current state of the job ("running")
- **Errors**:
  - No paused job: No job is currently paused
  - RPC error: Failed to communicate with backend services

#### 13. RPCStartCWEViewJob
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

#### 14. RPCStopCWEViewJob
- **Description**: Stops a running CWE view job
- **Request Parameters**:
  - `session_id` (string, optional): ID of the session to stop (default: current session)
- **Response**:
  - `success` (bool): true if job stopped successfully
  - `session_id` (string): ID of the stopped job session
- **Errors**:
  - No running job: No job is currently running
  - RPC error: Failed to communicate with backend services

---

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

---

# SSG Meta Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Orchestrates SSG (SCAP Security Guide) import jobs by coordinating between remote and local services. Pulls SSG repository, lists guide files, and imports HTML guides into the local database.

## Available RPC Methods

#### 15. RPCSSGStartImportJob
- **Description**: Starts a new SSG import job that pulls repository and imports all guides
- **Request Parameters**:
  - `run_id` (string, optional): Unique identifier for the run (auto-generated if not provided)
- **Response**:
  - `success` (bool): true if job started successfully
  - `run_id` (string): ID of the started job
- **Errors**:
  - Job running: An import job is already running
  - RPC error: Failed to communicate with backend services

#### 16. RPCSSGStopImportJob
- **Description**: Stops the currently running SSG import job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if job stopped successfully
- **Errors**:
  - No running job: No import job is currently running
  - RPC error: Failed to communicate with backend services

#### 17. RPCSSGPauseImportJob
- **Description**: Pauses the currently running SSG import job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if job paused successfully
- **Errors**:
  - No running job: No import job is currently running
  - Not running: Job is not in running state

#### 18. RPCSSGResumeImportJob
- **Description**: Resumes a paused SSG import job
- **Request Parameters**:
  - `run_id` (string, required): ID of the job to resume
- **Response**:
  - `success` (bool): true if job resumed successfully
- **Errors**:
  - No paused job: No import job is currently paused
  - RPC error: Failed to communicate with backend services

#### 19. RPCSSGGetImportStatus
- **Description**: Gets the current status of the SSG import job
- **Request Parameters**: None
- **Response**:
  - `id` (string): Job run ID
  - `data_type` (string): Type of data ("ssg")
  - `state` (string): Current state ("queued", "running", "paused", "completed", "failed", "stopped")
  - `started_at` (string): Timestamp when job started
  - `completed_at` (string, optional): Timestamp when job completed
  - `error` (string, optional): Error message if job failed
  - `progress` (object): Progress details
    - `total_guides` (int): Total number of guides to import
    - `processed_guides` (int): Number of guides successfully imported
    - `failed_guides` (int): Number of guides that failed to import
    - `current_file` (string): Currently processing file
  - `metadata` (object, optional): Additional metadata
- **Errors**:
  - No active job: No import job is currently running

## Notes
- SSG import job workflow: 1) Pull repository, 2) List guide files, 3) Import each guide
- Job state is in-memory only (not persisted across restarts)
- Supports pause/resume during import process
- Only one SSG import job can run at a time

---

# ETL Engine Monitoring

## Service Type
RPC (stdin/stdout message passing)

## Description
Provides observability into the Unified ETL Engine's hierarchical FSM structure and permit allocations.

## Available RPC Methods

#### 20. RPCGetEtlTree
- **Description**: Retrieves the hierarchical ETL tree showing Macro FSM and all Provider FSMs
- **Request Parameters**: None
- **Response**:
  - `macro_fsm` (object): Macro FSM state
    - `id` (string): Macro FSM identifier
    - `state` (string): Current macro state ("BOOTSTRAPPING", "ORCHESTRATING", "STABILIZING", "DRAINING")
    - `created_at` (string): Creation timestamp
    - `updated_at` (string): Last update timestamp
  - `providers` (array): List of Provider FSM states
    - `id` (string): Provider FSM identifier
    - `provider_type` (string): Type of provider ("cve", "cwe", "capec", "attack")
    - `state` (string): Current provider state ("IDLE", "ACQUIRING", "RUNNING", "WAITING_QUOTA", "WAITING_BACKOFF", "PAUSED", "TERMINATED")
    - `last_checkpoint` (string): URN of last processed item
    - `processed_count` (int): Number of items processed
    - `error_count` (int): Number of errors encountered
    - `created_at` (string): Creation timestamp
    - `updated_at` (string): Last update timestamp
    - `permits_allocated` (int): Number of permits currently allocated to this provider
  - `total_checkpoints` (int): Total number of checkpoints saved
- **Errors**:
  - No active ETL session: No ETL orchestration is currently active

#### 21. RPCGetProviderCheckpoints
- **Description**: Retrieves checkpoints for a specific provider
- **Request Parameters**:
  - `provider_id` (string, required): Provider FSM identifier
  - `limit` (int, optional): Maximum number of checkpoints to return (default: 100)
  - `success_only` (bool, optional): Only return successful checkpoints (default: false)
- **Response**:
  - `checkpoints` (array): List of checkpoints
    - `urn` (string): Full URN of the checkpoint
    - `provider_id` (string): Provider FSM identifier
    - `processed_at` (string): When the checkpoint was created
    - `success` (bool): Whether processing was successful
    - `error_message` (string, optional): Error message if not successful
  - `total` (int): Total number of checkpoints for this provider
- **Errors**:
  - Provider not found: No provider with the given ID exists

## Notes
- ETL tree provides real-time view of orchestration hierarchy
- Checkpoints are stored every 100 items for resilience
- Provider states persist across service restarts
- Macro FSM coordinates all provider FSMs