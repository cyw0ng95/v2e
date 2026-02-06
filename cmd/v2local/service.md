
# CVE & CWE Local Service

## Service Type
RPC (stdin/stdout message passing)

## Description
Manages local storage and retrieval of CVE, CWE, CAPEC, ATT&CK, ASVS, SSG, CCE, bookmarks, notes, and memory cards using SQLite databases. Provides CRUD operations for CVE records and read/import operations for CWE, CAPEC, ATT&CK, and ASVS records.

## Available RPC Methods

### CVE Operations

#### X. RPCGetNotesByBookmark (alias for RPCGetNotesByBookmarkID)
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

#### 1. RPCSaveCVEByID
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

#### 2. RPCIsCVEStoredByID
- **Description**: Checks if a CVE exists in the local database
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to check
- **Response**:
  - `stored` (bool): true if CVE exists in database
  - `cve_id` (string): The queried CVE ID
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Database error: Failed to query database

#### 3. RPCGetCVEByID
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

#### 4. RPCDeleteCVEByID
- **Description**: Deletes a CVE record from the local database
- **Request Parameters**:
  - `cve_id` (string, required): CVE identifier to delete
- **Response**:
  - `success` (bool): true if deleted successfully
  - `cve_id` (string): The deleted CVE ID
- **Errors**:
  - Missing CVE ID: `cve_id` parameter is required
  - Database error: Failed to delete from database

#### 5. RPCListCVEs
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

#### 6. RPCCountCVEs
- **Description**: Counts the total number of CVEs in the local database
- **Request Parameters**: None
- **Response**:
  - `count` (int): Total number of CVEs in the database
- **Errors**:
  - Database error: Failed to query database

#### 7. RPCCreateCVE
- **Description**: Creates a new CVE record by fetching from remote NVD API and saving to local database
- **Request Parameters**:
  - `id` (string, required): CVE identifier to create (e.g., "CVE-2021-44228")
- **Response**:
  - `success` (bool): true if created successfully
  - `cve_id` (string): ID of the created CVE
  - `cve` (object): The CVE object that was created
- **Errors**:
  - Missing CVE ID: `id` parameter is required
  - Not found: CVE not found in NVD database
  - Database error: Failed to save to database
  - Remote error: Failed to fetch from NVD API
- **Example**:
  - **Request**: {"id": "CVE-2021-44228"}
  - **Response**: {"success": true, "cve_id": "CVE-2021-44228", "cve": {...}}

#### 8. RPCUpdateCVE
- **Description**: Updates an existing CVE record by refetching from remote NVD API and updating local database
- **Request Parameters**:
  - `id` (string, required): CVE identifier to update
- **Response**:
  - `success` (bool): true if updated successfully
  - `cve_id` (string): ID of the updated CVE
  - `cve` (object): The updated CVE object
- **Errors**:
  - Missing CVE ID: `id` parameter is required
  - Not found: CVE not found in NVD database
  - Database error: Failed to update database
  - Remote error: Failed to fetch from NVD API
- **Example**:
  - **Request**: {"id": "CVE-2021-44228"}
  - **Response**: {"success": true, "cve_id": "CVE-2021-44228", "cve": {...}}

### CWE Operations

#### 9. RPCGetCWEByID
- **Description**: Retrieves a CWE record from the local database
- **Request Parameters**:
  - `cwe_id` (string, required): CWE identifier to retrieve
- **Response**:
  - `cwe` (object): The CWE object with all fields
- **Errors**:
  - Missing CWE ID: `cwe_id` parameter is required
  - Not found: CWE not found in database
  - Database error: Failed to query database

#### 10. RPCListCWEs
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

#### 11. RPCImportCWEs
- **Description**: Imports CWE data from a JSON file into the local database
- **Request Parameters**:
  - `path` (string, optional): Path to the JSON file containing CWE data (default: "assets/cwe-raw.json")
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of CWEs imported
- **Errors**:
  - File error: Failed to read or parse the JSON file
  - Database error: Failed to insert CWE data into database

### CAPEC Operations

#### 12. RPCImportCAPECs
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

#### 13. RPCForceImportCAPECs
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

#### 14. RPCListCAPECs
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

#### 15. RPCGetCAPECByID
- **Description**: Retrieves a CAPEC record from the local database
- **Request Parameters**:
  - `capec_id` (string, required): CAPEC identifier to retrieve
- **Response**:
  - `capec` (object): The CAPEC object with all fields
- **Errors**:
  - Missing CAPEC ID: `capec_id` parameter is required
  - Not found: CAPEC not found in database
  - Database error: Failed to query database

#### 16. RPCGetCAPECCatalogMeta
- **Description**: Retrieves metadata about the CAPEC catalog
- **Request Parameters**: None
- **Response**:
  - `version` (string): Version of the CAPEC catalog
  - `release_date` (string): Release date of the catalog
  - `total_count` (int): Total number of CAPEC entries
- **Errors**:
  - Not found: No CAPEC catalog metadata in database
  - Database error: Failed to query database

### CWE Views Operations

#### 17. RPCGetCWEViews
- **Description**: Retrieves CWE views from the local database
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `views` ([]object): Array of CWE view objects
  - `total` (int): Total number of CWE views in the database
- **Errors**:
  - Database error: Failed to query database

#### 18. RPCGetCWEViewByID
- **Description**: Retrieves a specific CWE view by ID from the local database
- **Request Parameters**:
  - `view_id` (string, required): CWE view identifier to retrieve
- **Response**:
  - `view` (object): The CWE view object with all fields
- **Errors**:
  - Missing view ID: `view_id` parameter is required
  - Not found: CWE view not found in database
  - Database error: Failed to query database

#### 19. RPCSaveCWEView
- **Description**: Saves a CWE view to the local database
- **Request Parameters**:
  - `id` (string, required): CWE view identifier
  - All view fields (name, type, status, objective, raw)
- **Response**:
  - `success` (bool): true if saved successfully
- **Errors**:
  - Missing view ID: `id` parameter is required
  - Database error: Failed to save view

#### 20. RPCDeleteCWEView
- **Description**: Deletes a CWE view from the local database
- **Request Parameters**:
  - `id` (string, required): CWE view identifier to delete
- **Response**:
  - `success` (bool): true if deleted successfully
- **Errors**:
  - Missing view ID: ` required
  - Database error: Failed to delete view

#### 21. RPCGetCWERelationships
- **id` parameter isDescription**: Retrieves relationships for a specific CWE from the local database
- **Request Parameters**:
  - `cwe_id` (string, required): CWE identifier to retrieve relationships for
  - `relationship_type` (string, optional): Type of relationship to filter (e.g., "ChildOf", "ParentOf")
- **Response**:
  - `relationships` ([]object): Array of relationship objects
- **Errors**:
  - Missing CWE ID: `cwe_id` parameter is required
  - Database error: Failed to query database

### ATT&CK Operations

#### 22. RPCImportATTACKs
- **Description**: Imports ATT&CK data from XLSX file into the local database
- **Request Parameters**:
  - `path` (string, required): Path to the XLSX file containing ATT&CK data
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of ATT&CK entries imported
- **Errors**:
  - File error: Failed to read or parse the XLSX file
  - Database error: Failed to insert ATT&CK data into database

#### 23. RPCGetAttackTechnique
- **Description**: Retrieves ATT&CK techniques with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `techniques` ([]object): Array of ATT&CK technique objects
  - `total` (int): Total number of techniques in the database
- **Errors**:
  - Database error: Failed to query database

#### 24. RPCGetAttackTactic
- **Description**: Retrieves ATT&CK tactics with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `tactics` ([]object): Array of ATT&CK tactic objects
  - `total` (int): Total number of tactics in the database
- **Errors**:
  - Database error: Failed to query database

#### 25. RPCGetAttackMitigation
- **Description**: Retrieves ATT&CK mitigations with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `mitigations` ([]object): Array of ATT&CK mitigation objects
  - `total` (int): Total number of mitigations in the database
- **Errors**:
  - Database error: Failed to query database

#### 26. RPCGetAttackSoftware
- **Description**: Retrieves ATT&CK software with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `software` ([]object): Array of ATT&CK software objects
  - `total` (int): Total number of software items in the database
- **Errors**:
  - Database error: Failed to query database

#### 27. RPCGetAttackGroup
- **Description**: Retrieves ATT&CK groups with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `groups` ([]object): Array of ATT&CK group objects
  - `total` (int): Total number of groups in the database
- **Errors**:
  - Database error: Failed to query database

#### 28. RPCGetAttackTechniqueByID
- **Description**: Retrieves a specific ATT&CK technique by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK technique identifier
- **Response**:
  - `technique` (object): The ATT&CK technique object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Technique not found in database
  - Database error: Failed to query database

#### 29. RPCGetAttackTacticByID
- **Description**: Retrieves a specific ATT&CK tactic by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK tactic identifier
- **Response**:
  - `tactic` (object): The ATT&CK tactic object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Tactic not found in database
  - Database error: Failed to query database

#### 30. RPCGetAttackMitigationByID
- **Description**: Retrieves a specific ATT&CK mitigation by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK mitigation identifier
- **Response**:
  - `mitigation` (object): The ATT&CK mitigation object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Mitigation not found in database
  - Database error: Failed to query database

#### 31. RPCGetAttackSoftwareByID
- **Description**: Retrieves a specific ATT&CK software by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK software identifier
- **Response**:
  - `software` (object): The ATT&CK software object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Software not found in database
  - Database error: Failed to query database

#### 32. RPCGetAttackGroupByID
- **Description**: Retrieves a specific ATT&CK group by ID
- **Request Parameters**:
  - `id` (string, required): ATT&CK group identifier
- **Response**:
  - `group` (object): The ATT&CK group object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Group not found in database
  - Database error: Failed to query database

#### 33. RPCListAttackTechniques
- **Description**: Lists ATT&CK techniques with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `techniques` ([]object): Array of ATT&CK technique objects
  - `total` (int): Total number of techniques in the database
- **Errors**:
  - Database error: Failed to query database

#### 34. RPCListAttackTactics
- **Description**: Lists ATT&CK tactics with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `tactics` ([]object): Array of ATT&CK tactic objects
  - `total` (int): Total number of tactics in the database
- **Errors**:
  - Database error: Failed to query database

#### 35. RPCListAttackMitigations
- **Description**: Lists ATT&CK mitigations with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `mitigations` ([]object): Array of ATT&CK mitigation objects
  - `total` (int): Total number of mitigations in the database
- **Errors**:
  - Database error: Failed to query database

#### 36. RPCListAttackSoftware
- **Description**: Lists ATT&CK software with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `software` ([]object): Array of ATT&CK software objects
  - `total` (int): Total number of software items in the database
- **Errors**:
  - Database error: Failed to query database

#### 37. RPCListAttackGroups
- **Description**: Lists ATT&CK groups with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 10)
- **Response**:
  - `groups` ([]object): Array of ATT&CK group objects
  - `total` (int): Total number of groups in the database
- **Errors**:
  - Database error: Failed to query database

#### 38. RPCGetAttackImportMetadata
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

### ASVS Operations

#### 39. RPCImportASVS
- **Description**: Imports ASVS requirements from a CSV URL
- **Request Parameters**:
  - `url` (string, required): URL to the ASVS CSV file
- **Response**:
  - `success` (bool): true if imported successfully
- **Errors**:
  - Missing URL: `url` parameter is required
  - Download error: Failed to download CSV from URL
  - Parse error: Failed to parse CSV file
  - Database error: Failed to save ASVS requirements to database
- **Example**:
  - **Request**: {"url": "https://raw.githubusercontent.com/OWASP/ASVS/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv"}
  - **Response**: {"success": true}

#### 40. RPCListASVS
- **Description**: Lists ASVS requirements from the local database with pagination and filtering
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 100, max: 1000)
  - `chapter` (string, optional): Filter by chapter (e.g., "V1", "V2")
  - `level` (int, optional): Filter by ASVS level (1, 2, or 3)
- **Response**:
  - `requirements` ([]object): Array of ASVS requirement objects
  - `total` (int): Total number of requirements matching filters
  - `offset` (int): The offset used
  - `limit` (int): The limit used
- **Errors**:
  - Database error: Failed to query database
- **Example**:
  - **Request**: {"offset": 0, "limit": 10, "chapter": "V1", "level": 1}
  - **Response**: {"requirements": [...], "total": 25, "offset": 0, "limit": 10}

#### 41. RPCGetASVSByID
- **Description**: Retrieves an ASVS requirement by its ID
- **Request Parameters**:
  - `requirement_id` (string, required): ASVS requirement identifier (e.g., "1.1.1")
- **Response**:
  - ASVS requirement object with fields:
    - `requirementID` (string): Requirement identifier
    - `chapter` (string): Chapter identifier (e.g., "V1")
    - `section` (string): Section name
    - `description` (string): Requirement description
    - `level1` (bool): Applies to Level 1
    - `level2` (bool): Applies to Level 2
    - `level3` (bool): Applies to Level 3
    - `cwe` (string, optional): Related CWE identifiers
- **Errors**:
  - Missing requirement ID: `requirement_id` parameter is required
  - Not found: ASVS requirement not found in database
  - Database error: Failed to query database
- **Example**:
  - **Request**: {"requirement_id": "1.1.1"}
  - **Response**: {"requirementID": "1.1.1", "chapter": "V1", "section": "Architecture", "description": "...", "level1": true, "level2": true, "level3": true, "cwe": "CWE-1127"}

### CCE Operations

#### 42. RPCGetCCE
- **Description**: Retrieves a CCE record by ID
- **Request Parameters**:
  - `id` (string, required): CCE identifier (e.g., "CCE-12345-0")
- **Response**:
  - `cce` (object): CCE object with all fields
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: CCE not found in database
  - Database error: Failed to query database

#### 43. RPCSaveCCE
- **Description**: Saves a CCE record to the local database
- **Request Parameters**:
  - `cce` (object, required): CCE object to save
- **Response**:
  - `success` (bool): true if saved successfully
- **Errors**:
  - Missing CCE data: `cce` parameter is required
  - Database error: Failed to save to database

#### 44. RPCListCCEs
- **Description**: Lists CCE records with pagination
- **Request Parameters**:
  - `offset` (int, optional): Offset for pagination (default: 0)
  - `limit` (int, optional): Limit for pagination (default: 100)
- **Response**:
  - `cces` ([]object): Array of CCE objects
  - `total` (int): Total number of CCEs
  - `offset` (int): The offset used
  - `limit` (int): The limit used
- **Errors**:
  - Database error: Failed to query database

#### 45. RPCCountCCEs
- **Description**: Counts the total number of CCEs in the database
- **Request Parameters**: None
- **Response**:
  - `count` (int): Total number of CCEs
- **Errors**:
  - Database error: Failed to query database

#### 46. RPCDeleteCCE
- **Description**: Deletes a CCE record from the database
- **Request Parameters**:
  - `id` (string, required): CCE identifier to delete
- **Response**:
  - `success` (bool): true if deleted successfully
- **Errors**:
  - Missing ID: `id` parameter is required
  - Database error: Failed to delete from database

#### 47. RPCImportCCEs
- **Description**: Imports CCE data from a JSON file into the local database
- **Request Parameters**:
  - `path` (string, optional): Path to the JSON file containing CCE data (default: "assets/cce-5.0-2023-06-08.json")
- **Response**:
  - `success` (bool): true if import was successful
  - `count` (int): Number of CCEs imported
- **Errors**:
  - File error: Failed to read or parse the JSON file
  - Database error: Failed to insert CCE data into database

### Bookmark Operations

#### 48. RPCCreateBookmark
- **Description**: Creates a new bookmark with associated memory card
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
- **Errors**:
  - Missing required fields: global_item_id, item_type, item_id, title are required
  - Database error: Failed to create bookmark

#### 49. RPCGetBookmark
- **Description**: Retrieves a bookmark by ID
- **Request Parameters**:
  - `id` (int, required): Bookmark ID
- **Response**:
  - `bookmark` (object): The bookmark object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Bookmark not found
  - Database error: Failed to query database

#### 50. RPCUpdateBookmark
- **Description**: Updates a bookmark by ID
- **Request Parameters**:
  - `id` (int, required): Bookmark ID
  - `title` (string, optional): Updated title
  - `description` (string, optional): Updated description
- **Response**:
  - `success` (bool): true if updated
  - `bookmark` (object): Updated bookmark
- **Errors**:
  - Not found: Bookmark not found
  - Database error: Failed to update database

#### 51. RPCDeleteBookmark
- **Description**: Deletes a bookmark by ID
- **Request Parameters**:
  - `id` (int, required): Bookmark ID
- **Response**:
  - `success` (bool): true if deleted
- **Errors**:
  - Not found: Bookmark not found
  - Database error: Failed to delete from database

#### 52. RPCListBookmarks
- **Description**: Lists bookmarks with optional filters and pagination
- **Request Parameters**:
  - `offset` (int, optional): Pagination offset (default: 0)
  - `limit` (int, optional): Pagination limit (default: 10)
- **Response**:
  - `bookmarks` ([]object): Array of bookmark objects
  - `total` (int): Total number of bookmarks
  - `offset` (int): Offset used
  - `limit` (int): Limit used
- **Errors**:
  - Database error: Failed to query database

### Note Operations

#### 53. RPCAddNote
- **Description**: Creates a new note for a bookmark
- **Request Parameters**:
  - `bookmark_id` (int, required): Associated bookmark ID
  - `content` (string, required): Note content (TipTap JSON)
- **Response**:
  - `success` (bool): true if created
  - `note` (object): The created note
- **Errors**:
  - Missing required fields: bookmark_id and content are required
  - Not found: Bookmark not found
  - Database error: Failed to create note

#### 54. RPCGetNote
- **Description**: Retrieves a note by ID
- **Request Parameters**:
  - `id` (int, required): Note ID
- **Response**:
  - `note` (object): The note object
- **Errors**:
  - Missing ID: `id` parameter is required
  - Not found: Note not found
  - Database error: Failed to query database

#### 55. RPCUpdateNote
- **Description**: Updates a note by ID
- **Request Parameters**:
  - `id` (int, required): Note ID
  - `content` (string, optional): Updated content
- **Response**:
  - `success` (bool): true if updated
  - `note` (object): Updated note
- **Errors**:
  - Not found: Note not found
  - Database error: Failed to update database

#### 56. RPCDeleteNote
- **Description**: Deletes a note by ID
- **Request Parameters**:
  - `id` (int, required): Note ID
- **Response**:
  - `success` (bool): true if deleted
- **Errors**:
  - Not found: Note not found
  - Database error: Failed to delete from database

#### 57. RPCGetNotesByBookmarkID
- **Description**: Retrieves all notes for a given bookmark ID
- **Request Parameters**:
  - `bookmark_id` (int, required): The bookmark ID to retrieve notes for
- **Response**:
  - `notes` ([]object): Array of note objects for the bookmark
- **Errors**:
  - Missing or invalid bookmark_id
  - Database error: Failed to query notes for the bookmark

### Memory Card Operations

#### 58. RPCCreateMemoryCard
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

#### 59. RPCGetMemoryCard
- **Description**: Retrieves a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID
- **Response**:
  - `memory_card` (object): The memory card
- **Errors**:
  - Not found
  - Database error

#### 60. RPCUpdateMemoryCard
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

#### 61. RPCDeleteMemoryCard
- **Description**: Deletes a memory card by ID
- **Request Parameters**:
  - `id` (int, required): Memory card ID
- **Response**:
  - `success` (bool): true if deleted
- **Errors**:
  - Not found
  - Database error

#### 62. RPCListMemoryCards
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

### SSG Operations

#### 63. RPCSSGImportGuide
- **Description**: Imports an SSG guide from a file
- **Request Parameters**:
  - `path` (string, required): Path to the guide file
- **Response**:
  - `success` (bool): true if imported successfully
  - `guide` (object): The imported guide
- **Errors**:
  - Missing path: `path` parameter is required
  - File error: Failed to read or parse the file
  - Database error: Failed to save to database

#### 64. RPCSSGImportTable
- **Description**: Imports an SSG table from a file
- **Request Parameters**:
  - `path` (string, required): Path to the table file
- **Response**:
  - `success` (bool): true if imported successfully
  - `table` (object): The imported table
- **Errors**:
  - Missing path
  - File error
  - Database error

#### 65. RPCSSGGetGuide
- **Description**: Retrieves an SSG guide by ID
- **Request Parameters**:
  - `id` (string, required): Guide identifier
- **Response**:
  - `guide` (object): The guide object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 66. RPCSSGListGuides
- **Description**: Lists all SSG guides
- **Request Parameters**: None
- **Response**:
  - `guides` (array): Array of guide objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 67. RPCSSGListTables
- **Description**: Lists all SSG tables
- **Request Parameters**: None
- **Response**:
  - `tables` (array): Array of table objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 68. RPCSSGGetTable
- **Description**: Retrieves an SSG table by ID
- **Request Parameters**:
  - `id` (string, required): Table identifier
- **Response**:
  - `table` (object): The table object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 69. RPCSSGGetTableEntries
- **Description**: Retrieves entries for an SSG table
- **Request Parameters**:
  - `table_id` (string, required): Table identifier
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `entries` (array): Array of table entry objects
  - `total` (int): Total count
- **Errors**:
  - Missing table_id
  - Database error

#### 70. RPCSSGGetTree
- **Description**: Retrieves the full SSG tree structure
- **Request Parameters**: None
- **Response**:
  - `tree` (object): Root tree node with nested children
- **Errors**:
  - Database error

#### 71. RPCSSGGetTreeNode
- **Description**: Retrieves a specific tree node by guide ID
- **Request Parameters**:
  - `guide_id` (string, required): Guide identifier
- **Response**:
  - `nodes` (array): Root TreeNode pointers with nested children
  - `count` (int): Number of root nodes
- **Errors**:
  - Missing guide_id
  - Not found
  - Database error

#### 72. RPCSSGGetGroup
- **Description**: Retrieves an SSG group by ID
- **Request Parameters**:
  - `id` (string, required): Group identifier
- **Response**:
  - Group object with all fields
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 73. RPCSSGGetChildGroups
- **Description**: Retrieves direct child groups of a parent group
- **Request Parameters**:
  - `parent_id` (string, optional): Parent group ID (empty for top-level groups)
- **Response**:
  - `groups` (array): Array of child group objects
  - `count` (int): Number of groups returned
- **Errors**:
  - Database error

#### 74. RPCSSGGetRule
- **Description**: Retrieves an SSG rule by ID with references
- **Request Parameters**:
  - `id` (string, required): Rule identifier
- **Response**:
  - Rule object with all fields including references array
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 75. RPCSSGListRules
- **Description**: Lists SSG rules with optional filters and pagination
- **Request Parameters**:
  - `group_id` (string, optional): Filter by parent group
  - `severity` (string, optional): Filter by severity (low, medium, high)
  - `offset` (int, optional): Pagination offset
  - `limit` (int, optional): Pagination limit
- **Response**:
  - `rules` (array): Array of rule objects with references
  - `total` (int): Total number of rules matching filters
- **Errors**:
  - Database error

#### 76. RPCSSGGetChildRules
- **Description**: Retrieves direct child rules of a group
- **Request Parameters**:
  - `group_id` (string, required): Parent group ID
- **Response**:
  - `rules` (array): Array of child rule objects with references
  - `count` (int): Number of rules returned
- **Errors**:
  - Missing group_id
  - Database error

#### 77. RPCSSGImportManifest
- **Description**: Imports an SSG manifest from a file
- **Request Parameters**:
  - `path` (string, required): Path to the manifest file
- **Response**:
  - `success` (bool): true if imported successfully
  - `manifest` (object): The imported manifest
- **Errors**:
  - Missing path
  - File error
  - Database error

#### 78. RPCSSGListManifests
- **Description**: Lists all SSG manifests
- **Request Parameters**: None
- **Response**:
  - `manifests` (array): Array of manifest objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 79. RPCSSGGetManifest
- **Description**: Retrieves an SSG manifest by ID
- **Request Parameters**:
  - `id` (string, required): Manifest identifier
- **Response**:
  - `manifest` (object): The manifest object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 80. RPCSSGListProfiles
- **Description**: Lists all SSG profiles
- **Request Parameters**: None
- **Response**:
  - `profiles` (array): Array of profile objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 81. RPCSSGGetProfile
- **Description**: Retrieves an SSG profile by ID
- **Request Parameters**:
  - `id` (string, required): Profile identifier
- **Response**:
  - `profile` (object): The profile object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 82. RPCSSGGetProfileRules
- **Description**: Retrieves rules for an SSG profile
- **Request Parameters**:
  - `profile_id` (string, required): Profile identifier
- **Response**:
  - `rules` (array): Array of rule objects
  - `count` (int): Number of rules
- **Errors**:
  - Missing profile_id
  - Database error

#### 83. RPCSSGImportDataStream
- **Description**: Imports an SSG datastream from a file
- **Request Parameters**:
  - `path` (string, required): Path to the datastream file
- **Response**:
  - `success` (bool): true if imported successfully
  - `datastream` (object): The imported datastream
- **Errors**:
  - Missing path
  - File error
  - Database error

#### 84. RPCSSGListDataStreams
- **Description**: Lists all SSG datastreams
- **Request Parameters**: None
- **Response**:
  - `datastreams` (array): Array of datastream objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 85. RPCSSGGetDataStream
- **Description**: Retrieves an SSG datastream by ID
- **Request Parameters**:
  - `id` (string, required): Datastream identifier
- **Response**:
  - `datastream` (object): The datastream object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 86. RPCSSGListDSProfiles
- **Description**: Lists all SSG datastream profiles
- **Request Parameters**:
  - `datastream_id` (string, optional): Filter by datastream
- **Response**:
  - `profiles` (array): Array of datastream profile objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 87. RPCSSGGetDSProfile
- **Description**: Retrieves an SSG datastream profile by ID
- **Request Parameters**:
  - `id` (string, required): Profile identifier
- **Response**:
  - `profile` (object): The profile object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 88. RPCSSGGetDSProfileRules
- **Description**: Retrieves rules for an SSG datastream profile
- **Request Parameters**:
  - `profile_id` (string, required): Profile identifier
- **Response**:
  - `rules` (array): Array of rule objects
  - `count` (int): Number of rules
- **Errors**:
  - Missing profile_id
  - Database error

#### 89. RPCSSGListDSGroups
- **Description**: Lists SSG datastream groups
- **Request Parameters**:
  - `datastream_id` (string, optional): Filter by datastream
- **Response**:
  - `groups` (array): Array of group objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 90. RPCSSGListDSRules
- **Description**: Lists SSG datastream rules
- **Request Parameters**:
  - `datastream_id` (string, optional): Filter by datastream
  - `group_id` (string, optional): Filter by group
- **Response**:
  - `rules` (array): Array of rule objects
  - `total` (int): Total count
- **Errors**:
  - Database error

#### 91. RPCSSGGetDSRule
- **Description**: Retrieves an SSG datastream rule by ID
- **Request Parameters**:
  - `id` (string, required): Rule identifier
- **Response**:
  - `rule` (object): The rule object
- **Errors**:
  - Missing id
  - Not found
  - Database error

#### 92. RPCSSGGetCrossReferences
- **Description**: Retrieves cross-references for a given SSG object
- **Request Parameters**:
  - `source_type` (string, optional): Type of source object ("guide", "table", "manifest", "datastream")
  - `source_id` (string, optional): Source object identifier
  - `target_type` (string, optional): Type of target object
  - `target_id` (string, optional): Target object identifier
  - `limit` (int, optional): Maximum number of results
  - `offset` (int, optional): Pagination offset
- **Response**:
  - `cross_references` (array): Array of cross-reference objects
  - `count` (int): Number of cross-references returned
- **Errors**:
  - Missing parameters: Must provide either source_type/source_id or target_type/target_id
  - Database error

#### 93. RPCSSGFindRelatedObjects
- **Description**: Finds all objects related to a given SSG object via cross-references
- **Request Parameters**:
  - `object_type` (string, required): Type of object ("guide", "table", "manifest", "datastream")
  - `object_id` (string, required): Object identifier
  - `link_type` (string, optional): Filter by link type ("rule_id", "cce", "product", "profile_id")
  - `limit` (int, optional): Maximum number of results
  - `offset` (int, optional): Pagination offset
- **Response**:
  - `related_objects` (array): Array of cross-reference objects
  - `count` (int): Number of related objects returned
- **Errors**:
  - Missing object_type or object_id
  - Database error

## Configuration
- **CVE Database Path**: Configurable via `CVE_DB_PATH` environment variable (default: "cve.db")
- **CWE Database Path**: Configurable via `CWE_DB_PATH` environment variable (default: "cwe.db")
- **CAPEC Database Path**: Configurable via `CAPEC_DB_PATH` environment variable (default: "capec.db")
- **ATT&CK Database Path**: Configurable via `ATTACK_DB_PATH` environment variable (default: "attack.db")
- **ASVS Database Path**: Configurable via `ASVS_DB_PATH` environment variable (default: "asvs.db")
- **CAPEC Strict XSD Validation**: Enabled via `CAPEC_STRICT_XSD` environment variable (default: disabled)
- **SSG Database Path**: Configurable via `SSG_DB_PATH` environment variable (default: "ssg.db")

## Notes
- Uses SQLite databases for local storage of CVE, CWE, CAPEC, ATT&CK, ASVS, SSG, CCE, bookmarks, notes, and memory cards
- Automatically imports ATT&CK data from XLSX files in the assets directory at startup
- Supports multiple data types (CVE, CWE, CAPEC, ATT&CK, ASVS, SSG, CCE) in separate databases
- Provides comprehensive CRUD operations for all data types
- ASVS data can be imported from the official OWASP ASVS v5.0.0 CSV file on GitHub
- Includes pagination support for listing operations
- SSG data is read-only after import (no update or delete operations)
- Cross-references enable navigation between related SSG objects based on rule IDs, CCE identifiers, products, and profile IDs
