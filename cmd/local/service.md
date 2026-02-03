### 50. RPCCreateMemoryCard
- **Description**: Creates a new memory card for a bookmark with TipTap content and classification fields
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
  - Database error

### 51. RPCGetMemoryCard
- **Description**: Retrieves a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID
- **Response**:
  - `memory_card` (object): The memory card
- **Errors**:
  - Not found
  - Database error

### 52. RPCUpdateMemoryCard
- **Description**: Updates a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID
  - Any updatable field (see Create)
- **Response**:
  - `success` (bool): true if updated
  - `memory_card` (object): The updated memory card
- **Errors**:
  - Not found
  - Database error

### 53. RPCDeleteMemoryCard
- **Description**: Deletes a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID
- **Response**:
  - `success` (bool): true if deleted
- **Errors**:
  - Not found
  - Database error

### 54. RPCListMemoryCards
- **Description**: Lists memory cards with optional filters and pagination
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
  - Database error
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
  - `stored` (bool): true if CVE exists in database
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


## CWE Views (V) — Design

This section documents the CWE "View" feature and how it is implemented in the local service.

**Purpose**
- Persist and serve CWE view resources (OpenAPI `V` views) for UI and API consumers.
- Provide CRUD and paginated listing; reserve job-controller integration for future website operations.

**Storage**
- Normalized SQLite tables prefixed `cwe_*`:
  - `cwe_views` (id TEXT PK, name, type, status, objective, raw BLOB)
  - `cwe_view_members`, `cwe_view_audience`, `cwe_view_references`, `cwe_view_notes`, `cwe_view_content_history`
- Nested arrays stored in separate tables linked by `view_id`.
- `raw` JSON blob stored on `cwe_views` for forward compatibility.

**RPC Surface (local subprocess)**
- `RPCSaveCWEView` (payload: `CWEView`)
- `RPCGetCWEViewByID` (payload: `{id}`)
- `RPCListCWEViews` (payload: `{offset,limit}`)
- `RPCDeleteCWEView` (payload: `{id}`)

**Job Controller (future)**
- A `pkg/cwe/job` controller will be added in a later tier to handle long-running view-generation/import jobs.
- It will persist session/progress and invoke local RPCs via the broker.

**Testing**
- Unit tests for store methods and handlers are provided (`pkg/cwe/local_views_test.go` and `cmd/local/cwe_handlers_views_test.go`).
- Integration with website and meta job orchestration will be tested in later tiers; integration tests remain unchanged.

**Notes**
- To enable migrations, call `AutoMigrateViews(db)` (function provided in `pkg/cwe/local_views.go`) from `NewLocalCWEStore`'s `AutoMigrate` list or manually where appropriate.
- Handler registration helper `RegisterCWEViewHandlers(sp, store, logger)` is provided; add calls in `cmd/local/main.go` where `sp` is available.

---

## Notes
- Uses SQLite databases for local storage of CVE, CWE, CAPEC, ATT&CK, and SSG data
- Automatically imports ATT&CK data from XLSX files in the assets directory at startup
- Supports multiple data types (CVE, CWE, CAPEC, ATT&CK, SSG) in separate databases
- Provides comprehensive CRUD operations for all data types
- Includes pagination support for listing operations

---

# SSG Local Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Manages local storage and retrieval of SCAP Security Guide (SSG) data using SQLite database. Provides CRUD operations for SSG guides, groups, and rules with hierarchical tree structure support.

## Available RPC Methods

### 55. RPCSSGGetGuide
- **Description**: Retrieves an SSG guide by ID
- **Request Parameters**:
  - `id` (string, required): Guide identifier
- **Response**:
  - Guide object with all fields (id, product, profile_id, title, html_content, etc.)
- **Errors**:
  - Missing id: `id` parameter is required
  - Not found: Guide not found in database
  - Database error: Failed to query database

### 56. RPCSSGListGuides
- **Description**: Lists SSG guides with optional filters
- **Request Parameters**:
  - `product` (string, optional): Filter by product (e.g., "al2023", "rhel9")
  - `profile_id` (string, optional): Filter by profile ID
- **Response**:
  - `guides` (array): Array of guide objects
  - `count` (int): Number of guides returned
- **Errors**:
  - Database error: Failed to query database

### 57. RPCSSGGetTree
- **Description**: Retrieves the complete tree structure for a guide (flat groups + rules)
- **Request Parameters**:
  - `guide_id` (string, required): Guide identifier
- **Response**:
  - SSGTree object containing guide, groups array, and rules array
- **Errors**:
  - Missing guide_id: `guide_id` parameter is required
  - Not found: Guide not found in database
  - Database error: Failed to query database

### 58. RPCSSGGetTreeNode
- **Description**: Retrieves the tree structure for a guide as hierarchical TreeNode pointers
- **Request Parameters**:
  - `guide_id` (string, required): Guide identifier
- **Response**:
  - `nodes` (array): Root TreeNode pointers with nested children
  - `count` (int): Number of root nodes
- **Errors**:
  - Missing guide_id: `guide_id` parameter is required
  - Not found: Guide not found in database
  - Database error: Failed to build tree

### 59. RPCSSGGetGroup
- **Description**: Retrieves an SSG group by ID
- **Request Parameters**:
  - `id` (string, required): Group identifier
- **Response**:
  - Group object with all fields
- **Errors**:
  - Missing id: `id` parameter is required
  - Not found: Group not found in database
  - Database error: Failed to query database

### 60. RPCSSGGetChildGroups
- **Description**: Retrieves direct child groups of a parent group
- **Request Parameters**:
  - `parent_id` (string, optional): Parent group ID (empty for top-level groups)
- **Response**:
  - `groups` (array): Array of child group objects
  - `count` (int): Number of groups returned
- **Errors**:
  - Database error: Failed to query database

### 61. RPCSSGGetRule
- **Description**: Retrieves an SSG rule by ID with references
- **Request Parameters**:
  - `id` (string, required): Rule identifier
- **Response**:
  - Rule object with all fields including references array
- **Errors**:
  - Missing id: `id` parameter is required
  - Not found: Rule not found in database
  - Database error: Failed to query database

### 62. RPCSSGListRules
- **Description**: Lists SSG rules with optional filters and pagination
- **Request Parameters**:
  - `group_id` (string, optional): Filter by parent group
  - `severity` (string, optional): Filter by severity (low, medium, high)
  - `offset` (int, optional): Pagination offset (default: 0)
  - `limit` (int, optional): Pagination limit (default: 100)
- **Response**:
  - `rules` (array): Array of rule objects with references
  - `total` (int): Total number of rules matching filters
- **Errors**:
  - Database error: Failed to query database

### 63. RPCSSGGetChildRules
- **Description**: Retrieves direct child rules of a group
- **Request Parameters**:
  - `group_id` (string, required): Parent group ID
- **Response**:
  - `rules` (array): Array of child rule objects with references
  - `count` (int): Number of rules returned
- **Errors**:
  - Missing group_id: `group_id` parameter is required
  - Database error: Failed to query database

### 64. RPCSSGDeleteGuide
- **Description**: Deletes a guide and all associated groups and rules
- **Request Parameters**:
  - `id` (string, required): Guide identifier
- **Response**:
  - `success` (bool): true if deleted successfully
  - `id` (string): The deleted guide ID
- **Errors**:
  - Missing id: `id` parameter is required
  - Database error: Failed to delete from database

## Configuration
- **SSG Database Path**: Configurable via `SSG_DB_PATH` environment variable (default: "ssg.db")