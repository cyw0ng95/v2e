# CVE & CWE Local Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Manages local storage and retrieval of CVE, CWE, CAPEC, and ATT&CK data using SQLite databases. Provides CRUD operations for CVE records and read/import operations for CWE, CAPEC, and ATT&CK records.


## Available RPC Methods

### X. RPCGetNotesByBookmark (alias for RPCGetNotesByBookmarkID)
- **Description**: Retrieves all notes for a given bookmark ID (alias for `RPCGetNotesByBookmarkID`)
- **Request Parameters**:
  - `bookmark_id` (int, required): The bookmark ID to retrieve notes for
- **Response**:
  - `notes` ([]object): Array of note objects for the bookmark
- **Errors**:
  - Missing or invalid bookmark_id: `bookmark_id` parameter is required and must be a valid integer
  - Database error: Failed to query notes for the bookmark
- **Example**:
  - **Request**: {"bookmark_id": 1}
  - **Response**: {"notes": [ ... ]}

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

### 3. RPCGetCVEByID
- **Description**: Retrieves a CVE record from the local database
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to retrieve
- **Response**:
  - `cve` (object): The CVE object with all fields
  - `id` (string): The CVE ID
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Not found: CVE not found in database
  - Database error: Failed to query database

### 4. RPCDeleteCVEByID
- **Description**: Deletes a CVE record from the local database
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to delete
- **Response**:
  - `success` (bool): true if deleted successfully
  - `cve_id` (string): The deleted CVE ID
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Database error: Failed to delete from database

### 5. RPCListCVEs
- **Description**: Lists CVE records from the local database with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `cves` ([]object): Array of CVE objects
  - `total` (int): Total number of CVEs in the database
  - `offset` (int): The offset used
  - `limit` (int): The limit used
- **Errors**:
  - Database error: Failed to query database

### 6. RPCCountCVEs
- **Description**: Counts the total number of CVEs in the local database
- **Request Parameters**: None
- **Response**:
  - `count` (int): Total number of CVEs in the database
- **Errors**:
  - Database error: Failed to query database

### 7. RPCGetCWEByID
- **Description**: Retrieves a CWE record from the local database
- **Request Parameters**:
  - `cwe_id` (string, required): CWE identifier to retrieve
- **Response**:
  - `cwe` (object): The CWE object with all fields
- **Errors**:
  - Missing CWE ID: `cwe_id` parameter is required
  - Not found: CWE not found in database
  - Database error: Failed to query database

### 8. RPCListCWEs
- **Description**: Lists CWE records from the local database with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `cwes` ([]object): Array of CWE objects
  - `total` (int): Total number of CWEs in the database
  - `offset` (int): The offset used
  - `limit` (int): The limit used
- **Errors**:
  - Database error: Failed to query database

### 9. RPCImportCWEs
- **Description**: Imports CWE data from a JSON file into the local database
- **Request Parameters**:
  - `path` (string, optional): Path to the JSON file containing CWE data (default: "assets/cwe-raw.json")
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of CWEs imported
- **Errors**:
  - File error: Failed to read or parse the JSON file
  - Database error: Failed to insert CWE data into database

### 10. RPCImportCAPECs
- **Description**: Imports CAPEC data from XML file into the local database with optional XSD validation
- **Request Parameters**:
  - `path` (string, optional): Path to the XML file containing CAPEC data (default: "assets/capec_contents_latest.xml")
  - `xsd` (string, optional): Path to XSD schema file for validation (default: "assets/capec_schema_latest.xsd")
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of CAPEC entries imported
- **Errors**:
  - File error: Failed to read or parse the XML file
  - Validation error: Failed XSD validation if enabled
  - Database error: Failed to insert CAPEC data into database

### 11. RPCForceImportCAPECs
- **Description**: Forces import of CAPEC data from XML file, overwriting existing data
- **Request Parameters**:
  - `path` (string, optional): Path to the XML file containing CAPEC data (default: "assets/capec_contents_latest.xml")
  - `xsd` (string, optional): Path to XSD schema file for validation (default: "assets/capec_schema_latest.xsd")
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of CAPEC entries imported
- **Errors**:
  - File error: Failed to read or parse the XML file
  - Validation error: Failed XSD validation if enabled
  - Database error: Failed to insert CAPEC data into database

### 12. RPCListCAPECs
- **Description**: Lists CAPEC records from the local database with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `capecs` ([]object): Array of CAPEC objects
  - `total` (int): Total number of CAPECs in the database
  - `offset` (int): The offset used
  - `limit` (int): The limit used
- **Errors**:
  - Database error: Failed to query database

### 13. RPCGetCAPECByID
- **Description**: Retrieves a CAPEC record from the local database
- **Request Parameters**:
  - `capec_id` (string, required): CAPEC identifier to retrieve
- **Response**:
  - `capec` (object): The CAPEC object with all fields
- **Errors**:
  - Missing CAPEC ID: `capec_id` parameter is required
  - Not found: CAPEC not found in database
  - Database error: Failed to query database

### 14. RPCGetCAPECCatalogMeta
- **Description**: Retrieves metadata about the CAPEC catalog
- **Request Parameters**: None
- **Response**:
  - `version` (string): Version of the CAPEC catalog
  - `release_date` (string): Release date of the catalog
  - `total_count` (int): Total number of CAPEC entries
- **Errors**:
  - Not found: No CAPEC catalog metadata in database
  - Database error: Failed to query database

### 15. RPCGetCWEViews
- **Description**: Retrieves CWE views from the local database
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `views` ([]object): Array of CWE view objects
  - `total` (int): Total number of CWE views in the database
- **Errors**:
  - Database error: Failed to query database

### 16. RPCGetCWEViewByID
- **Description**: Retrieves a specific CWE view by ID from the local database
- **Request Parameters**:
  - `view_id` (string, required): CWE view identifier to retrieve
- **Response**:
  - `view` (object): The CWE view object with all fields
- **Errors**:
  - Missing view ID: `view_id` parameter is required
  - Not found: CWE view not found in database
  - Database error: Failed to query database

### 17. RPCGetCWERelationships
- **Description**: Retrieves relationships for a specific CWE from the local database
- **Request Parameters**:
  - `cwe_id` (string, required): CWE identifier to retrieve relationships for
  - `relationship_type` (string, optional): Type of relationship to filter (e.g., "ChildOf", "ParentOf")
- **Response**:
  - `relationships` ([]object): Array of relationship objects
- **Errors**:
  - Missing CWE ID: `cwe_id` parameter is required
  - Database error: Failed to query database

### 18. RPCImportATTACKs
- **Description**: Imports ATT&CK data from XLSX file into the local database
- **Request Parameters**:
  - `path` (string, required): Path to the XLSX file containing ATT&CK data
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of ATT&CK entries imported
- **Errors**:
  - File error: Failed to read or parse the XLSX file
  - Database error: Failed to insert ATT&CK data into database

### 19. RPCGetAttackTechnique
- **Description**: Retrieves ATT&CK techniques with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `techniques` ([]object): Array of ATT&CK technique objects
  - `total` (int): Total number of techniques in the database
- **Errors**:
  - Database error: Failed to query database

### 20. RPCGetAttackTactic
- **Description**: Retrieves ATT&CK tactics with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `tactics` ([]object): Array of ATT&CK tactic objects
  - `total` (int): Total number of tactics in the database
- **Errors**:
  - Database error: Failed to query database

### 21. RPCGetAttackMitigation
- **Description**: Retrieves ATT&CK mitigations with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `mitigations` ([]object): Array of ATT&CK mitigation objects
  - `total` (int): Total number of mitigations in the database
- **Errors**:
  - Database error: Failed to query database

### 22. RPCGetAttackSoftware
- **Description**: Retrieves ATT&CK software with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `software` ([]object): Array of ATT&CK software objects
  - `total` (int): Total number of software items in the database
- **Errors**:
  - Database error: Failed to query database

### 23. RPCGetAttackGroup
- **Description**: Retrieves ATT&CK groups with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `groups` ([]object): Array of ATT&CK group objects
  - `total` (int): Total number of groups in the database
- **Errors**:
  - Database error: Failed to query database

### 24. RPCGetAttackTechniqueByID
- **Description**: Retrieves a specific ATT&CK technique by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK technique identifier
- **Response**:
  - `technique` (object): The ATT&CK technique object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Technique not found in database
  - Database error: Failed to query database

### 25. RPCGetAttackTacticByID
- **Description**: Retrieves a specific ATT&CK tactic by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK tactic identifier
- **Response**:
  - `tactic` (object): The ATT&CK tactic object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Tactic not found in database
  - Database error: Failed to query database

### 26. RPCGetAttackMitigationByID
- **Description**: Retrieves a specific ATT&CK mitigation by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK mitigation identifier
- **Response**:
  - `mitigation` (object): The ATT&CK mitigation object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Mitigation not found in database
  - Database error: Failed to query database

### 27. RPCGetAttackSoftwareByID
- **Description**: Retrieves a specific ATT&CK software by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK software identifier
- **Response**:
  - `software` (object): The ATT&CK software object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Software not found in database
  - Database error: Failed to query database

### 28. RPCGetAttackGroupByID
- **Description**: Retrieves a specific ATT&CK group by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK group identifier
- **Response**:
  - `group` (object): The ATT&CK group object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Group not found in database
  - Database error: Failed to query database

### 29. RPCListAttackTechniques
- **Description**: Lists ATT&CK techniques with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `techniques` ([]object): Array of ATT&CK technique objects
  - `total` (int): Total number of techniques in the database
- **Errors**:
  - Database error: Failed to query database

### 30. RPCListAttackTactics
- **Description**: Lists ATT&CK tactics with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `tactics` ([]object): Array of ATT&CK tactic objects
  - `total` (int): Total number of tactics in the database
- **Errors**:
  - Database error: Failed to query database

### 31. RPCListAttackMitigations
- **Description**: Lists ATT&CK mitigations with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `mitigations` ([]object): Array of ATT&CK mitigation objects
  - `total` (int): Total number of mitigations in the database
- **Errors**:
  - Database error: Failed to query database

### 32. RPCListAttackSoftware
- **Description**: Lists ATT&CK software with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `software` ([]object): Array of ATT&CK software objects
  - `total` (int): Total number of software items in the database
- **Errors**:
  - Database error: Failed to query database

### 33. RPCListAttackGroups
- **Description**: Lists ATT&CK groups with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `groups` ([]object): Array of ATT&CK group objects
  - `total` (int): Total number of groups in the database
- **Errors**:
  - Database error: Failed to query database

### 34. RPCGetAttackImportMetadata
- **Description**: Retrieves metadata about the ATT&CK import
- **Request Parameters**: None
- **Response**:
  - `import_date` (string): Date of the last import
  - `total_techniques` (int): Total number of techniques imported
  - `total_tactics` (int): Total number of tactics imported
  - `total_mitigations` (int): Total number of mitigations imported
  - `total_software` (int): Total number of software items imported
  - `total_groups` (int): Total number of groups imported
- **Errors**:
  - Not found: No ATT&CK import metadata in database

## Configuration
- **CVE Database Path**: Configurable via `CVE_DB_PATH` environment variable (default: "cve.db")
- **CWE Database Path**: Configurable via `CWE_DB_PATH` environment variable (default: "cwe.db")
- **CAPEC Database Path**: Configurable via `CAPEC_DB_PATH` environment variable (default: "capec.db")
- **ATT&CK Database Path**: Configurable via `ATTACK_DB_PATH` environment variable (default: "attack.db")
- **CAPEC Strict XSD Validation**: Enabled via `CAPEC_STRICT_XSD` environment variable (default: disabled)

## Notes
- Uses SQLite databases for local storage of CVE, CWE, CAPEC, and ATT&CK data
- Automatically imports ATT&CK data from XLSX files in the assets directory at startup
- Supports multiple data types (CVE, CWE, CAPEC, ATT&CK) in separate databases
- Provides comprehensive CRUD operations for all data types
- Includes pagination support for listing operations