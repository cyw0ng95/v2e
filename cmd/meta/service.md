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
- **Description**: Creates a new CVE record in local storage
- **Request Parameters**:
  - `cve` (object, required): CVE object to create
- **Response**:
  - `success` (bool): true if created successfully
  - `cve_id` (string): ID of the created CVE