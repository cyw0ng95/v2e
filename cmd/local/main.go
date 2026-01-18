/*
Package main implements the local RPC service.

RPC API Specification:

CVE & CWE Local Service
=======================

Service Type: RPC (stdin/stdout message passing)
Description: Manages local storage and retrieval of CVE and CWE data using SQLite databases.
Provides CRUD operations for CVE records and read/import operations for CWE records.

Available RPC Methods:
----------------------

 1. RPCSaveCVEByID
    Description: Saves a CVE record to the local database
    Request Parameters:
    - cve (object, required): CVE object to save (must include id field)
    Response:
    - success (bool): true if saved successfully
    - cve_id (string): ID of the saved CVE
    Errors:
    - Missing CVE data: cve parameter is required
    - Invalid CVE: CVE object is missing required fields
    - Database error: Failed to save to database
    Example:
    Request:  {"cve": {"id": "CVE-2021-44228", "descriptions": [...], ...}}
    Response: {"success": true, "cve_id": "CVE-2021-44228"}

 2. RPCIsCVEStoredByID
    Description: Checks if a CVE exists in the local database
    Request Parameters:
    - cve_id (string, required): CVE identifier to check
    Response:
    - exists (bool): true if CVE exists in database
    - cve_id (string): The queried CVE ID
    Errors:
    - Missing CVE ID: cve_id parameter is required
    - Database error: Failed to query database
    Example:
    Request:  {"cve_id": "CVE-2021-44228"}
    Response: {"exists": true, "cve_id": "CVE-2021-44228"}

 3. RPCGetCVEByID
    Description: Retrieves a CVE record from the local database
    Request Parameters:
    - cve_id (string, required): CVE identifier to retrieve
    Response:
    - cve (object): CVE object with all fields
    Errors:
    - Missing CVE ID: cve_id parameter is required
    - Not found: CVE not found in database
    - Database error: Failed to query database
    Example:
    Request:  {"cve_id": "CVE-2021-44228"}
    Response: {"cve": {"id": "CVE-2021-44228", "descriptions": [...], ...}}

 4. RPCDeleteCVEByID
    Description: Deletes a CVE record from the local database
    Request Parameters:
    - cve_id (string, required): CVE identifier to delete
    Response:
    - success (bool): true if deleted successfully
    - cve_id (string): ID of the deleted CVE
    Errors:
    - Missing CVE ID: cve_id parameter is required
    - Not found: CVE not found in database
    - Database error: Failed to delete from database
    Example:
    Request:  {"cve_id": "CVE-2021-44228"}
    Response: {"success": true, "cve_id": "CVE-2021-44228"}

 5. RPCListCVEs
    Description: Lists CVE records with pagination support
    Request Parameters:
    - offset (int, optional): Starting offset for pagination (default: 0)
    - limit (int, optional): Maximum number of records to return (default: 10)
    Response:
    - cves ([]object): Array of CVE objects
    - offset (int): Starting offset used
    - limit (int): Limit used
    - total (int): Total number of CVEs in database
    Errors:
    - Database error: Failed to query database
    Example:
    Request:  {"offset": 0, "limit": 10}
    Response: {"cves": [...], "offset": 0, "limit": 10, "total": 150}

 6. RPCCountCVEs
    Description: Gets the total count of CVEs in the local database
    Request Parameters: None
    Response:
    - count (int): Total number of CVE records
    Errors:
    - Database error: Failed to query database
    Example:
    Request:  {}
    Response: {"count": 150}

 7. RPCGetCWEByID
    Description: Retrieves a CWE record from the local database
    Request Parameters:
    - cwe_id (string, required): CWE identifier to retrieve
    Response:
    - cwe (object): CWE object with all fields
    Errors:
    - Missing CWE ID: cwe_id parameter is required
    - Not found: CWE not found in database
    - Database error: Failed to query database
    Example:
    Request:  {"cwe_id": "CWE-79"}
    Response: {"ID": "CWE-79", "Name": "Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')", ...}

 8. RPCListCWEs
    Description: Lists CWE records with pagination support
    Request Parameters:
    - offset (int, optional): Starting offset for pagination (default: 0)
    - limit (int, optional): Maximum number of records to return (default: 100)
    Response:
    - cwes ([]object): Array of CWE objects
    - offset (int): Starting offset used
    - limit (int): Limit used
    - total (int): Total number of CWEs in database
    Errors:
    - Database error: Failed to query database
    Example:
    Request:  {"offset": 0, "limit": 100}
    Response: {"cwes": [...], "offset": 0, "limit": 100, "total": 1200}

 9. RPCImportCWEs
    Description: Imports CWE records from a JSON file into the local database
    Request Parameters:
    - path (string, required): Path to the JSON file containing CWE records
    Response:
    - success (bool): true if import succeeded
    Errors:
    - Missing path: path parameter is required
    - File error: Failed to open or parse file
    - Database error: Failed to import records
    Example:
    Request:  {"path": "assets/cwe-raw.json"}
    Response: {"success": true}

Notes:
------
- Uses SQLite databases for local CVE and CWE storage
- Database paths configured via CVE_DB_PATH and CWE_DB_PATH environment variables (defaults: cve.db, cwe.db)
- Supports GORM for ORM operations
- Service runs as a subprocess managed by the broker
- All requests are routed through the broker via RPC
*/
package main

import (
	"fmt"
	"os"

	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func main() {
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "local"
	}

	// Set up logging using common subprocess framework
	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	// Get database path from environment or use default
	dbPath := os.Getenv("CVE_DB_PATH")
	if dbPath == "" {
		dbPath = "cve.db"
	}

	// Create or open the database
	db, err := local.NewDB(dbPath)
	if err != nil {
		logger.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize CWE store (using a separate DB file or the same as CVE)
	cweDBPath := os.Getenv("CWE_DB_PATH")
	if cweDBPath == "" {
		cweDBPath = "cwe.db"
	}
	cweStore, err := cwe.NewLocalCWEStore(cweDBPath)
	if err != nil {
		logger.Error("Failed to open CWE database: %v", err)
		os.Exit(1)
	}

	// Import CWEs from JSON file at startup (if file exists)
	// Removed duplicate importCWEsAtStartup definition; now only in cwe_handlers.go

	// Create subprocess instance
	sp := subprocess.New(processID)

	// Register RPC handlers
	sp.RegisterHandler("RPCSaveCVEByID", createSaveCVEByIDHandler(db, logger))
	sp.RegisterHandler("RPCIsCVEStoredByID", createIsCVEStoredByIDHandler(db, logger))
	sp.RegisterHandler("RPCGetCVEByID", createGetCVEByIDHandler(db, logger))
	sp.RegisterHandler("RPCDeleteCVEByID", createDeleteCVEByIDHandler(db, logger))
	sp.RegisterHandler("RPCListCVEs", createListCVEsHandler(db, logger))
	sp.RegisterHandler("RPCCountCVEs", createCountCVEsHandler(db, logger))
	sp.RegisterHandler("RPCGetCWEByID", createGetCWEByIDHandler(cweStore, logger))
	sp.RegisterHandler("RPCListCWEs", createListCWEsHandler(cweStore, logger))
	sp.RegisterHandler("RPCImportCWEs", createImportCWEsHandler(cweStore, logger))

	logger.Info("CVE local service started")

	// Run with default lifecycle management
	subprocess.RunWithDefaults(sp, logger)
}
