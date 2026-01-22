# CVE & CWE Local Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Manages local storage and retrieval of CVE and CWE data using SQLite databases. Provides CRUD operations for CVE records and read/import operations for CWE records.

## Available RPC Methods

### 1. RPCSaveCVEByID
- **Description**: Saves a CVE record to the local database
- **Request Parameters**:
  - `cve` (object, required): CVE object to save (must include id field)
- **Response**:
  - `success` (bool): true if saved successfully
  - `cve_id` (string): ID of the saved CVE
- **Errors**:
  - Missing CVE data: `cve` parameter is required
  - Invalid CVE: CVE object is missing required fields
  - Database error: Failed to save to database
- **Example**:
  - **Request**: {"cve": {"id": "CVE-2021-44228", "descriptions": [...], ...}}
  - **Response**: {"success": true, "cve_id": "CVE-2021-44228"}

### 2. RPCIsCVEStoredByID
- **Description**: Checks if a CVE exists in the local database
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to check
- **Response**:
  - `exists` (bool): true if CVE exists in database
  - `cve_id` (string): The queried CVE ID
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Database error: Failed to query database