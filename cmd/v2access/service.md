
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
  - **Response**: `{"status": "ok"}`

### 2. POST /restful/rpc
- **Description**: Generic RPC forwarding endpoint that routes requests to backend services
- **Request Parameters**:
  - `method` (string, required): RPC method name (e.g., "RPCGetCVE")
  - `params` (object, optional): Parameters to pass to the RPC method
  - `target` (string, optional): Target process ID (default: "broker")
- **Response**:
  - `retcode` (int): 0 for success, non-zero for errors
  - `message` (string): Success message or error description
  - `payload` (object): Response data from backend service
- **Errors**:
  - Invalid JSON: `retcode=400`, missing or malformed request body
  - RPC timeout: `retcode=500`, backend service did not respond in time
  - Backend error: `retcode=500`, backend service returned an error
- **Example**:
  - **Request**: `{"method": "RPCGetCVE", "target": "local", "params": {"id": "CVE-2021-44228"}}`
  - **Response**: `{"retcode": 0, "message": "success", "payload": {...}}`

### 3. GET /restful/cves
- **Description**: Lists CVE records from local storage with pagination and filtering
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset (default: 0)
  - `limit` (int, optional): Pagination limit (default: 10, max: 100)
- **Response**:
  - `cves` ([]object): Array of CVE objects
  - `total` (int): Total number of CVEs
  - `offset` (int): Offset used
  - `limit` (int): Limit used
- **Example**:
  - **Request**: GET /restful/cves?offset=0&limit=10
  - **Response**: `{"cves": [...], "total": 150000, "offset": 0, "limit": 10}`

### 4. GET /restful/cves/:id
- **Description**: Retrieves a specific CVE by ID from local storage
- **Request Parameters**:
  - `id` (string, required): CVE identifier in URL path
- **Response**:
  - `cve` (object): CVE object with all fields
- **Errors**:
  - Not found: CVE not found in database
- **Example**:
  - **Request**: GET /restful/cves/CVE-2021-44228
  - **Response**: `{"cve": {...}}`

### 5. POST /restful/cves
- **Description**: Creates a new CVE record by fetching from remote and saving to local
- **Request Parameters**:
  - `id` (string, required): CVE identifier to create
- **Response**:
  - `success` (bool): true if created successfully
  - `cve` (object): The CVE object
- **Example**:
  - **Request**: POST /restful/cves {"id": "CVE-2021-44228"}
  - **Response**: `{"success": true, "cve": {...}}`

### 6. GET /restful/bookmarks
- **Description**: Lists bookmarks with optional filters
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset (default: 0)
  - `limit` (int, optional): Pagination limit (default: 10)
- **Response**:
  - `bookmarks` ([]object): Array of bookmark objects
  - `total` (int): Total number of bookmarks
  - `offset` (int): Offset used
  - `limit` (int): Limit used

### 7. POST /restful/bookmarks
- **Description**: Creates a new bookmark
- **Request Parameters**:
  - `global_item_id` (string, required): Global item identifier
  - `item_type` (string, required): Type of item (cve, cwe, capec, attack, etc.)
  - `item_id` (string, required): Item identifier
  - `title` (string, required): Bookmark title
  - `description` (string, optional): Bookmark description
- **Response**:
  - `success` (bool): true if created
  - `bookmark` (object): The created bookmark
  - `memory_card` (object): Associated memory card

### 8. GET /restful/bookmarks/:id
- **Description**: Retrieves a bookmark by ID
- **Request Parameters**:
  - `id` (int, required): Bookmark ID in URL path
- **Response**:
  - `bookmark` (object): Bookmark object

### 9. PUT /restful/bookmarks/:id
- **Description**: Updates a bookmark by ID
- **Request Parameters**:
  - `id` (int, required): Bookmark ID in URL path
  - Any updatable field (title, description)
- **Response**:
  - `success` (bool): true if updated
  - `bookmark` (object): Updated bookmark

### 10. DELETE /restful/bookmarks/:id
- **Description**: Deletes a bookmark by ID
- **Request Parameters**:
  - `id` (int, required): Bookmark ID in URL path
- **Response**:
  - `success` (bool): true if deleted

### 11. GET /restful/bookmarks/:id/notes
- **Description**: Retrieves notes for a bookmark
- **Request Parameters**:
  - `id` (int, required): Bookmark ID in URL path
- **Response**:
  - `notes` ([]object): Array of note objects

### 12. POST /restful/notes
- **Description**: Creates a new note for a bookmark
- **Request Parameters**:
  - `bookmark_id` (int, required): Associated bookmark ID
  - `content` (string, required): Note content (TipTap JSON)
- **Response**:
  - `success` (bool): true if created
  - `note` (object): The created note

### 13. GET /restful/notes/:id
- **Description**: Retrieves a note by ID
- **Request Parameters**:
  - `id` (int, required): Note ID in URL path
- **Response**:
  - `note` (object): Note object

### 14. PUT /restful/notes/:id
- **Description**: Updates a note by ID
- **Request Parameters**:
  - `id` (int, required): Note ID in URL path
  - `content` (string, optional): Updated content
- **Response**:
  - `success` (bool): true if updated
  - `note` (object): Updated note

### 15. DELETE /restful/notes/:id
- **Description**: Deletes a note by ID
- **Request Parameters**:
  - `id` (int, required): Note ID in URL path
- **Response**:
  - `success` (bool): true if deleted

### 16. POST /restful/memory-cards
- **Description**: Creates a new memory card for a bookmark
- **Request Parameters**:
  - `bookmark_id` (int, required): Associated bookmark ID
  - `front_content` (string, required): Front content
  - `back_content` (string, required): Back content
  - `major_class` (string, required): Major class
  - `minor_class` (string, required): Minor class
  - `status` (string, required): Status (active, archived)
  - `content` (object, required): TipTap JSON content
- **Response**:
  - `success` (bool): true if created
  - `memory_card` (object): The created memory card

### 17. GET /restful/memory-cards/:id
- **Description**: Retrieves a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID in URL path
- **Response**:
  - `memory_card` (object): Memory card object

### 18. PUT /restful/memory-cards/:id
- **Description**: Updates a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID in URL path
  - Any updatable field
- **Response**:
  - `success` (bool): true if updated
  - `memory_card` (object): Updated memory card

### 19. DELETE /restful/memory-cards/:id
- **Description**: Deletes a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID in URL path
- **Response**:
  - `success` (bool): true if deleted

### 20. GET /restful/memory-cards
- **Description**: Lists memory cards with optional filters
- **Request Parameters**:
  - `bookmark_id` (int, optional): Filter by bookmark
  - `major_class` (string, optional): Filter by major class
  - `minor_class` (string, optional): Filter by minor class
  - `status` (string, optional): Filter by status
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `memory_cards` ([]object): Array of memory card objects
  - `total` (int): Total count
  - `offset` (int): Offset used
  - `limit` (int): Limit used

### 21. GET /restful/cwes
- **Description**: Lists CWE records with pagination
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset (default: 0)
  - `limit` (int, optional): Pagination limit (default: 10, max: 100)
- **Response**:
  - `cwes` ([]object): Array of CWE objects
  - `total` (int): Total number of CWEs
  - `offset` (int): Offset used
  - `limit` (int): Limit used

### 22. GET /restful/cwes/:id
- **Description**: Retrieves a specific CWE by ID
- **Request Parameters**:
  - `id` (string, required): CWE identifier in URL path (e.g., "CWE-79")
- **Response**:
  - `cwe` (object): CWE object with all fields

### 23. GET /restful/cwes/views
- **Description**: Lists CWE views
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `views` ([]object): Array of CWE view objects
  - `total` (int): Total count
  - `offset` (int): Offset used
  - `limit` (int): Limit used

### 24. GET /restful/cwes/views/:id
- **Description**: Retrieves a specific CWE view by ID
- **Request Parameters**:
  - `id` (string, required): View ID in URL path
- **Response**:
  - `view` (object): CWE view object

### 25. GET /restful/cwes/:id/relationships
- **Description**: Retrieves relationships for a CWE
- **Request Parameters**:
  - `id` (string, required): CWE ID in URL path
  - `relationship_type` (string, optional): Filter by relationship type
- **Response**:
  - `relationships` ([]object): Array of relationship objects

### 26. GET /restful/capecs
- **Description**: Lists CAPEC records with pagination
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset (default: 0)
  - `limit` (int, optional): Pagination limit (default: 10, max: 100)
- **Response**:
  - `capecs` ([]object): Array of CAPEC objects
  - `total` (int): Total number of CAPECs
  - `offset` (int): Offset used
  - `limit` (int): Limit used

### 27. GET /restful/capecs/:id
- **Description**: Retrieves a specific CAPEC by ID
- **Request Parameters**:
  - `id` (string, required): CAPEC identifier in URL path
- **Response**:
  - `capec` (object): CAPEC object with all fields

### 28. GET /restful/capecs/meta
- **Description**: Retrieves CAPEC catalog metadata
- **Request Parameters**: None
- **Response**:
  - `version` (string): Catalog version
  - `release_date` (string): Release date
  - `total_count` (int): Total entries

### 29. GET /restful/attack/techniques
- **Description**: Lists ATT&CK techniques with pagination
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `techniques` ([]object): Array of technique objects
  - `total` (int): Total count

### 30. GET /restful/attack/tactics
- **Description**: Lists ATT&CK tactics with pagination
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `tactics` ([]object): Array of tactic objects
  - `total` (int): Total count

### 31. GET /restful/attack/mitigations
- **Description**: Lists ATT&CK mitigations with pagination
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `mitigations` ([]object): Array of mitigation objects
  - `total` (int): Total count

### 32. GET /restful/attack/techniques/:id
- **Description**: Retrieves a specific ATT&CK technique by ID
- **Request Parameters**:
  - `id` (string, required): Technique ID in URL path
- **Response**:
  - `technique` (object): Technique object

### 33. GET /restful/asvs
- **Description**: Lists ASVS requirements with pagination and filtering
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit (max: 1000)
  - `chapter` (string, optional): Filter by chapter
  - `level` (int, optional): Filter by ASVS level
- **Response**:
  - `requirements` ([]object): Array of ASVS requirement objects
  - `total` (int): Total matching requirements
  - `offset` (int): Offset used
  - `limit` (int): Limit used

### 34. GET /restful/asvs/:id
- **Description**: Retrieves a specific ASVS requirement by ID
- **Request Parameters**:
  - `id` (string, required): Requirement ID in URL path (e.g., "1.1.1")
- **Response**:
  - `requirement` (object): ASVS requirement object

### 35. GET /restful/session/status
- **Description**: Gets the current status of the data fetching session
- **Request Parameters**: None
- **Response**:
  - `has_session` (bool): Whether a session exists
  - `session_id` (string): Session ID
  - `state` (string): Session state
  - `data_type` (string): Type of data being fetched
  - `progress` (object): Progress details

### 36. POST /restful/session/start
- **Description**: Starts a new data fetching session
- **Request Parameters**:
  - `session_id` (string, optional): Unique session ID
  - `data_type` (string, optional): Type of data (cve, cwe, capec, attack)
  - `start_index` (int, optional): Starting index
  - `results_per_batch` (int, optional): Batch size
- **Response**:
  - `success` (bool): true if started
  - `session_id` (string): Session ID
  - `state` (string): Session state

### 37. POST /restful/session/stop
- **Description**: Stops the current data fetching session
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if stopped
  - `fetched_count` (int): Items fetched
  - `stored_count` (int): Items stored
  - `error_count` (int): Errors encountered

### 38. POST /restful/session/pause
- **Description**: Pauses the current data fetching session
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if paused
  - `state` (string): New state

### 39. POST /restful/session/resume
- **Description**: Resumes a paused data fetching session
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if resumed
  - `state` (string): New state

### 40. GET /restful/etl/tree
- **Description**: Retrieves the hierarchical ETL tree showing Macro FSM and all Provider FSMs
- **Request Parameters**: None
- **Response**:
  - `macro_fsm` (object): Macro FSM state
  - `providers` (array): Provider FSM states
  - `total_checkpoints` (int): Total checkpoints

### 41. GET /restful/ssg/guides
- **Description**: Lists SSG guides
- **Request Parameters**: None
- **Response**:
  - `guides` ([]object): Array of guide objects
  - `total` (int): Total count

### 42. GET /restful/ssg/guides/:id
- **Description**: Retrieves a specific SSG guide by ID
- **Request Parameters**:
  - `id` (string, required): Guide ID in URL path
- **Response**:
  - `guide` (object): Guide object

### 43. POST /restful/ssg/import/start
- **Description**: Starts an SSG import job
- **Request Parameters**:
  - `run_id` (string, optional): Unique run ID
- **Response**:
  - `success` (bool): true if started
  - `run_id` (string): Job run ID

### 44. POST /restful/ssg/import/stop
- **Description**: Stops the current SSG import job
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if stopped

### 45. GET /restful/ssg/import/status
- **Description**: Gets the current status of the SSG import job
- **Request Parameters**: None
- **Response**:
  - `id` (string): Job run ID
  - `state` (string): Job state
  - `progress` (object): Progress details

### 46. GET /restful/sysmetrics
- **Description**: Retrieves current system performance metrics
- **Request Parameters**: None
- **Response**:
  - `cpu_usage` (float): CPU usage percentage
  - `memory_usage` (float): Memory usage percentage
  - `load_avg` ([]float): Load averages
  - `disk_usage` (uint64): Used disk space
  - `disk_total` (uint64): Total disk space
  - `swap_usage` (float): Swap usage percentage
  - `net_rx` (uint64): Network received bytes
  - `net_tx` (uint64): Network transmitted bytes

## Configuration
- **RPC Timeout**: Configurable via `config.json` under `access.rpc_timeout_seconds` (default: 30 seconds)
- **Shutdown Timeout**: Configurable via `config.json` under `access.shutdown_timeout_seconds` (default: 10 seconds)
- **Static Directory**: Configurable via `config.json` under `access.static_dir` (default: "website")
- **Server Address**: Configurable via `config.json` under `server.address` (default: "0.0.0.0:8080")

## Security Features

### Rate Limiting
All REST endpoints (except `/restful/health`) are protected by token-bucket rate limiting to prevent DoS attacks:
- **Per-Client Limit**: 50 requests per client with 1 request/second refill rate
- **Client Identification**: Based on IP address (supports X-Forwarded-For and X-Real-IP headers)
- **Trusted Proxies**: Localhost (127.0.0.1, ::1) bypasses rate limiting
- **Excluded Paths**: `/restful/health` is excluded from rate limiting
- **Response Headers**: Rate limited requests return HTTP 429 with:
  - `X-RateLimit-Limit`: Maximum tokens allowed
  - `X-RateLimit-Refill`: Refill interval in seconds
  - `Retry-After`: Suggested retry delay

### Input Validation
- All RPC requests are validated for required fields and data types
- CVE IDs must match format `CVE-YYYY-NNNN...` (e.g., CVE-2021-44228)
- CWE IDs must match format `CWE-NNN` (e.g., CWE-79)
- CAPEC IDs must match format `CAPEC-NNN` (e.g., CAPEC-123)
- Pagination parameters are validated to be within acceptable ranges
- File paths are validated to prevent directory traversal attacks

## Notes
- Forwards all RPC calls to the broker for routing
- Handles authentication and request validation
- Provides RESTful API for external clients
- Returns standardized error responses
- Runs as a subprocess managed by the broker
- Uses stdin/stdout for RPC communication with the broker
- All REST endpoints are documented above with their parameters and responses
