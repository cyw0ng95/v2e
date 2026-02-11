# Access Service

## Service Type
REST (HTTP/JSON)

## Description
REST-to-RPC gateway service that provides external HTTP access to the v2e system. All business operations are forwarded to backend services via a unified RPC endpoint.

## Architecture

The v2access service is a lightweight gateway that:
- Accepts HTTP requests from clients (frontend, external APIs)
- Converts them to RPC messages
- Routes them through the broker to backend services
- Returns RPC responses as HTTP responses

**Backend Services** (implement actual business logic):
- **v2local** - CVE, CWE, CAPEC, ATT&CK, ASVS, SSG, CCE, bookmarks, notes, memory cards, GLC graphs
- **v2remote** - Remote data fetching from NVD, GitHub, etc.
- **v2meta** - Job orchestration, session management, ETL coordination
- **v2sysmon** - System performance monitoring
- **v2broker** - Process management and message routing

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
- **Description**: Unified RPC endpoint that forwards all requests to backend services
- **Request Parameters**:
  - `method` (string, required): RPC method name (e.g., "RPCGetCVE", "RPCCreateBookmark")
  - `params` (object, optional): Parameters to pass to the RPC method
  - `target` (string, optional): Target process ID (default: "broker")
    - Common targets: "local", "remote", "meta", "sysmon", "broker"
- **Response**:
  - `retcode` (int): 0 for success, non-zero for errors
  - `message` (string): Success message or error description
  - `payload` (object): Response data from backend service
- **Errors**:
  - Invalid JSON: `retcode=400`, missing or malformed request body
  - RPC timeout: `retcode=500`, backend service did not respond in time
  - Backend error: `retcode=500`, backend service returned an error
- **Examples**:
  - **Get CVE**: `{"method": "RPCGetCVE", "target": "local", "params": {"cve_id": "CVE-2021-44228"}}`
  - **List CVEs**: `{"method": "RPCListCVEs", "target": "local", "params": {"offset": 0, "limit": 10}}`
  - **Create Bookmark**: `{"method": "RPCCreateBookmark", "target": "local", "params": {"global_item_id": "...", "title": "..."}}`

## Backend RPC Methods Reference

All available RPC methods are documented in the respective service documentation:

### v2local - Local Data Storage
**File**: `cmd/v2local/service.md`

**RPC Methods**: 100+ methods for:
- CVE operations: RPCGetCVE, RPCCreateCVE, RPCUpdateCVE, RPCDeleteCVE, RPCListCVEs, etc.
- CWE operations: RPCGetCWE, RPCListCWEs, RPCImportCWEs, etc.
- CAPEC operations: RPCGetCAPEC, RPCListCAPECs, RPCImportCAPECs, etc.
- ATT&CK operations: RPCGetAttackTechnique, RPCListAttackTechniques, etc.
- ASVS operations: RPCListASVS, RPCGetASVSByID, RPCImportASVS, etc.
- SSG operations: RPCSSGGetGuide, RPCSSGListGuides, RPCSSGImportGuide, etc.
- CCE operations: RPCGetCCE, RPCListCCEs, RPCImportCCEs, etc.
- Bookmark operations: RPCCreateBookmark, RPCGetBookmark, RPCUpdateBookmark, RPCDeleteBookmark, RPCListBookmarks
- Note operations: RPCAddNote, RPCGetNote, RPCUpdateNote, RPCDeleteNote, RPCGetNotesByBookmark
- Memory Card operations: RPCCreateMemoryCard, RPCGetMemoryCard, RPCUpdateMemoryCard, RPCDeleteMemoryCard, RPCListMemoryCards, RPCRateMemoryCard
- GLC operations: RPCGLCGraphCreate, RPCGLCGraphGet, RPCGLCVersionGet, RPCGLCPresetCreate, etc.

### v2remote - Remote Data Fetching
**File**: `cmd/v2remote/service.md`

**RPC Methods**:
- RPCGetCVEByID - Fetch CVE from NVD API
- RPCGetCVECnt - Get total CVE count from NVD
- RPCFetchCVEs - Fetch multiple CVEs with pagination
- RPCFetchViews - Fetch CWE views from GitHub
- SSG Git operations: RPCSSGCloneRepo, RPCSSGPullRepo, RPCSSGListFiles

### v2meta - Job Orchestration
**File**: `cmd/v2meta/service.md`

**RPC Methods**:
- CVE orchestration: RPCGetCVE, RPCCreateCVE, RPCUpdateCVE, RPCDeleteCVE, RPCListCVEs, RPCCountCVEs
- Session control: RPCStartSession, RPCStartTypedSession, RPCStopSession, RPCGetSessionStatus, RPCPauseJob, RPCResumeJob
- ETL management: RPCGetEtlTree, RPCStartProvider, RPCPauseProvider, RPCStopProvider
- SSG import jobs: RPCSSGStartImportJob, RPCSSGStopImportJob, RPCSSGPauseImportJob, RPCSSGResumeImportJob, RPCSSGGetImportStatus
- Performance: RPCGetKernelMetrics, RPCUpdatePerformancePolicy

### v2sysmon - System Monitoring
**File**: `cmd/v2sysmon/service.md`

**RPC Methods**:
- RPCGetSysMetrics - Get system performance metrics (CPU, memory, disk, network, load averages)

### v2broker - Process Management
**File**: `cmd/v2broker/service.md`

**RPC Methods**:
- RPCGetMessageStats - Get message statistics for broker and all processes
- RPCGetMessageCount - Get total message count
- RPCRequestPermits - Request worker permits from global pool
- RPCReleasePermits - Release worker permits back to pool
- RPCGetKernelMetrics - Get broker performance metrics

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
- Acts as a REST-to-RPC gateway only - no business logic implemented
- All RPC calls are forwarded to the broker for routing to backend services
- Runs as a subprocess managed by the broker
- Uses stdin/stdout for RPC communication with the broker
- Serves static files from configured static directory (Next.js export output)
- Returns standardized error responses
