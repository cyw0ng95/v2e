# CVE Remote Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Fetches CVE (Common Vulnerabilities and Exposures) data from the NVD (National Vulnerability Database) API. Provides remote CVE data retrieval capabilities to other services.

## Available RPC Methods

### 1. RPCGetCVEByID
- **Description**: Fetches a specific CVE by its ID from the NVD API
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier (e.g., "CVE-2021-44228")
- **Response**:
  - `vulnerabilities` ([]object): Array of vulnerability objects (typically one)
    - Each vulnerability contains:
      - `cve` (object): CVE item with id, descriptions, metrics, references, etc.
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - NVD API error: Failed to fetch from NVD API
  - Not found: CVE not found in NVD database
  - NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)
- **Example**:
  - **Request**: {"cve_id": "CVE-2021-44228"}
  - **Response**: {"vulnerabilities": [{"cve": {"id": "CVE-2021-44228", "descriptions": [...], ...}}]}

### 2. RPCGetCVECnt
- **Description**: Gets the total count of CVEs available in the NVD database
- **Request Parameters**: None
- **Response**:
  - `total_results` (int): Total number of CVEs in NVD
- **Errors**:
  - NVD API error: Failed to query NVD API
  - NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)

### 3. RPCFetchCVEs
- **Description**: Fetches multiple CVEs from the NVD API with pagination
- **Request Parameters**:
  - `start_index` (int, optional): Index to start fetching from (default: 0)
  - `results_per_page` (int, optional): Number of results per page (default: 100)
- **Response**:
  - `vulnerabilities` ([]object): Array of vulnerability objects
  - `total_results` (int): Total number of CVEs available in NVD
  - `result_count` (int): Number of CVEs returned in this response
- **Errors**:
  - NVD API error: Failed to query NVD API
  - NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)
- **Example**:
  - **Request**: {"start_index": 0, "results_per_page": 10}
  - **Response**: {"vulnerabilities": [...], "total_results": 180000, "result_count": 10}

### 4. RPCFetchViews
- **Description**: Fetches CWE views from the GitHub repository
- **Request Parameters**:
  - `start_index` (int, optional): Index to start fetching from (default: 0)
  - `results_per_page` (int, optional): Number of results per page (default: 100)
- **Response**:
  - `views` ([]object): Array of CWE view objects
- **Errors**:
  - Archive download error: Failed to download the GitHub archive
  - HTTP error: Unexpected HTTP status when downloading archive
  - Archive read error: Failed to read the downloaded archive
  - Archive parse error: Failed to open the zip archive
- **Example**:
  - **Request**: {"start_index": 0, "results_per_page": 10}
  - **Response**: {"views": [...]}

### 5. RPCFetchSSGPackage
- **Description**: Downloads SSG (SCAP Security Guide) package from GitHub and verifies its SHA512 checksum
- **Request Parameters**:
  - `version` (string, optional): SSG version to fetch (default: "0.1.79")
- **Response**:
  - `package_data` (bytes): The tar.gz package data
  - `sha512` (string): SHA512 checksum
  - `verified` (bool): Whether checksum verification passed
  - `version` (string): The version that was fetched
- **Errors**:
  - Download failed: Failed to download SSG package from GitHub
  - Verification failed: SHA512 checksum mismatch
  - Network error: HTTP request failed
- **Example**:
  - **Request**: {"version": "0.1.79"}
  - **Response**: {"package_data": <binary>, "sha512": "abc123...", "verified": true, "version": "0.1.79"}

## Configuration
- **NVD API Key**: Configurable via `NVD_API_KEY` environment variable (optional, increases rate limits)
- **View Fetch URL**: Configurable via `VIEW_FETCH_URL` environment variable (default: "https://github.com/CWE-CAPEC/REST-API-wg/archive/refs/heads/main.zip")

## Notes
- Rate limits apply to NVD API access (requests with API key have higher limits)
- Automatically retries failed requests with exponential backoff
- Downloads and parses CWE views from GitHub repository
- Uses ZIP archive extraction to retrieve JSON files from GitHub repository
- SSG packages are fetched from ComplianceAsCode/content GitHub releases
- SHA512 checksums are verified to ensure package integrity
- All requests are routed through the broker for centralized management
- Service runs as a subprocess managed by the broker