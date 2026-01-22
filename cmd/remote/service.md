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