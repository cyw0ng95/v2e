/*
Package main implements the local RPC service.

Refer to service.md for the RPC API Specification and available methods.

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
