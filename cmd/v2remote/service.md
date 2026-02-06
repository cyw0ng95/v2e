
# CVE Remote Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Fetches CVE (Common Vulnerabilities and Exposures) data from the NVD (National Vulnerability Database) API. Provides remote CVE data retrieval capabilities to other services.

## Available RPC Methods

### CVE Remote Operations

#### 1. RPCGetCVEByID
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

#### 2. RPCGetCVECnt
- **Description**: Gets the total count of CVEs available in the NVD database
- **Request Parameters**: None
- **Response**:
  - `total_results` (int): Total number of CVEs in NVD
- **Errors**:
  - NVD API error: Failed to query NVD API
  - NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)

#### 3. RPCFetchCVEs
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

#### 4. RPCFetchViews
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

## Configuration
- **NVD API Key**: Configurable via `NVD_API_KEY` environment variable (optional, increases rate limits)
- **View Fetch URL**: Configurable via `VIEW_FETCH_URL` environment variable (default: "https://github.com/CWE-CAPEC/REST-API-wg/archive/refs/heads/main.zip")

## Notes
- Rate limits apply to NVD API access (requests with API key have higher limits)
- Automatically retries failed requests with exponential backoff
- Downloads and parses CWE views from GitHub repository
- Uses ZIP archive extraction to retrieve JSON files from GitHub repository
- All requests are routed through the broker for centralized management
- Service runs as a subprocess managed by the broker

---

# SSG Remote Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Manages Git repository operations for SCAP Security Guide (SSG) data. Provides clone/pull operations and file listing for SSG guides.

## Available RPC Methods

### SSG Git Operations

#### 5. RPCSSGCloneRepo
- **Description**: Clones the SSG repository to the local path
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if cloned successfully
  - `path` (string): Local path where repository was cloned
- **Errors**:
  - Repository exists: Repository already exists at the target path
  - Git error: Failed to clone repository

#### 6. RPCSSGPullRepo
- **Description**: Pulls the latest changes from the SSG repository
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if pull succeeded (or already up to date)
- **Errors**:
  - Git error: Failed to pull repository

#### 7. RPCSSGGetRepoStatus
- **Description**: Gets the current status of the SSG repository
- **Request Parameters**: None
- **Response**:
  - `commit_hash` (string): Short commit hash (7 characters)
  - `branch` (string): Current branch name
  - `is_clean` (bool): true if working tree is clean (no uncommitted changes)
- **Errors**:
  - Not found: Repository does not exist locally
  - Git error: Failed to get repository status

#### 8. RPCSSGListGuideFiles
- **Description**: Lists all SSG guide HTML files in the repository
- **Request Parameters**: None
- **Response**:
  - `files` ([]string): Array of guide filenames (e.g., "ssg-al2023-guide-cis.html")
  - `count` (int): Total number of guide files
- **Errors**:
  - Not found: Repository does not exist or guides directory is missing
  - Git error: Failed to read repository

#### 9. RPCSSGGetFilePath
- **Description**: Gets the absolute path to a file in the SSG repository
- **Request Parameters**:
  - `filename` (string, required): Relative path to file (e.g., "guides/ssg-al2023-guide-cis.html")
- **Response**:
  - `path` (string): Absolute path to the file
- **Errors**:
  - Missing filename: filename parameter is required

#### 10. RPCSSGGetGitClientStatus
- **Description**: Gets the status of the SSG Git client
- **Request Parameters**: None
- **Response**:
  - `initialized` (bool): Whether the client is initialized
  - `repo_exists` (bool): Whether the repository exists locally
  - `repo_url` (string): Repository URL
  - `repo_path` (string): Local repository path
- **Errors**: None

#### 11. RPCSSGEnsureRepo
- **Description**: Ensures the SSG repository exists, cloning if necessary
- **Request Parameters**: None
- **Response**:
  - `success` (bool): true if repository exists
  - `cloned` (bool): true if repository was cloned during this call
  - `path` (string): Local repository path
- **Errors**:
  - Git error: Failed to clone repository

## Configuration
SSG Git configuration is via build-time ldflags (see config_spec.json):
- **CONFIG_SSG_REPO_URL**: Git repository URL (default: "https://github.com/cyw0ng95/scap-security-guide-0.1.79")
- **CONFIG_SSG_REPO_PATH**: Local checkout path (default: "assets/ssg-git")

## Notes
- Uses go-git library for pure Go Git operations (no system git required)
- Repository is cloned to local path on first use
- Guide files must match `*-guide-*.html` pattern to be listed
- All requests are routed through the broker for centralized management
- Service runs as a subprocess managed by the broker
