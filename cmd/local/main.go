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
	"strings"

	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// importATTACKDataAtStartup automatically imports ATT&CK data from XLSX files in the assets directory at startup
func importATTACKDataAtStartup(attackStore *attack.LocalAttackStore, logger *common.Logger) {
	// Look for XLSX files in the current directory and assets subdirectory
	xlsxFiles := []string{}

	// Check for XLSX files in current directory
	if files, err := os.ReadDir("."); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
				xlsxFiles = append(xlsxFiles, file.Name())
			}
		}
	}

	// Check for XLSX files in assets/attack/ subdirectory
	if files, err := os.ReadDir("assets/attack"); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
				xlsxFiles = append(xlsxFiles, "assets/attack/"+file.Name())
			}
		}
	} else {
		// Check for XLSX files in assets/ directory if attack subdirectory doesn't exist
		if files, err := os.ReadDir("assets"); err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
					xlsxFiles = append(xlsxFiles, "assets/"+file.Name())
				}
			}
		}
	}

	// Import each XLSX file found
	for _, xlsxFile := range xlsxFiles {
		logger.Info("Attempting to import ATT&CK data from: %s", xlsxFile)

		// Check if the file exists before importing
		if _, err := os.Stat(xlsxFile); os.IsNotExist(err) {
			logger.Warn("ATT&CK XLSX file does not exist: %s", xlsxFile)
			continue
		}

		// Perform the import
		if err := attackStore.ImportFromXLSX(xlsxFile, false); err != nil {
			logger.Error("Failed to import ATT&CK data from %s: %v", xlsxFile, err)
		} else {
			logger.Info("Successfully imported ATT&CK data from: %s", xlsxFile)
		}
	}

	if len(xlsxFiles) == 0 {
		logger.Info("No ATT&CK XLSX files found for automatic import")
	}
}

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

	// Initialize CAPEC store (using CAPEC_DB_PATH env var)
	capecDBPath := os.Getenv("CAPEC_DB_PATH")
	if capecDBPath == "" {
		capecDBPath = "capec.db"
	}
	capecStore, err := capec.NewLocalCAPECStore(capecDBPath)
	if err != nil {
		logger.Error("Failed to open CAPEC database: %v", err)
		os.Exit(1)
	}

	// Initialize ATT&CK store (using ATTACK_DB_PATH env var)
	attackDBPath := os.Getenv("ATTACK_DB_PATH")
	if attackDBPath == "" {
		attackDBPath = "attack.db"
	}
	attackStore, err := attack.NewLocalAttackStore(attackDBPath)
	if err != nil {
		logger.Error("Failed to open ATT&CK database: %v", err)
		os.Exit(1)
	}

	// Import CWEs from JSON file at startup (if file exists)
	// Removed duplicate importCWEsAtStartup definition; now only in cwe_handlers.go

	// Import ATT&CK data from XLSX files at startup (if files exist)
	importATTACKDataAtStartup(attackStore, logger)

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
	sp.RegisterHandler("RPCImportCAPECs", createImportCAPECsHandler(capecStore, logger))
	sp.RegisterHandler("RPCForceImportCAPECs", createForceImportCAPECsHandler(capecStore, logger))
	sp.RegisterHandler("RPCListCAPECs", createListCAPECsHandler(capecStore, logger))
	sp.RegisterHandler("RPCGetCAPECByID", createGetCAPECByIDHandler(capecStore, logger))
	sp.RegisterHandler("RPCGetCAPECCatalogMeta", createGetCAPECCatalogMetaHandler(capecStore, logger))

	// Register CWE View handlers
	RegisterCWEViewHandlers(sp, cweStore, logger)

	// Register ATT&CK handlers
	sp.RegisterHandler("RPCImportATTACKs", createImportATTACKsHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackTechnique", createGetAttackTechniqueHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackTactic", createGetAttackTacticHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackMitigation", createGetAttackMitigationHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackSoftware", createGetAttackSoftwareHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackGroup", createGetAttackGroupHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackTechniqueByID", createGetAttackTechniqueByIDHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackTacticByID", createGetAttackTacticByIDHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackMitigationByID", createGetAttackMitigationByIDHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackSoftwareByID", createGetAttackSoftwareByIDHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackGroupByID", createGetAttackGroupByIDHandler(attackStore, logger))
	sp.RegisterHandler("RPCListAttackTechniques", createListAttackTechniquesHandler(attackStore, logger))
	sp.RegisterHandler("RPCListAttackTactics", createListAttackTacticsHandler(attackStore, logger))
	sp.RegisterHandler("RPCListAttackMitigations", createListAttackMitigationsHandler(attackStore, logger))
	sp.RegisterHandler("RPCListAttackSoftware", createListAttackSoftwareHandler(attackStore, logger))
	sp.RegisterHandler("RPCListAttackGroups", createListAttackGroupsHandler(attackStore, logger))
	sp.RegisterHandler("RPCGetAttackImportMetadata", createGetAttackImportMetadataHandler(attackStore, logger))

	logger.Info("CVE local service started")

	// Run with default lifecycle management
	subprocess.RunWithDefaults(sp, logger)
}
