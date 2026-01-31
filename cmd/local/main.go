/*
Package main implements the local RPC service.

Refer to service.md for the RPC API Specification and available
methods.
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
	"os"
	"strconv"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/notes"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// importATTACKDataAtStartup automatically imports ATT&CK data from XLSX files in the assets directory at startup
func importATTACKDataAtStartup(attackStore *attack.LocalAttackStore, logger *common.Logger) {
	logger.Info("Starting ATT&CK data import at startup")
	// Look for XLSX files in the current directory and assets subdirectory
	xlsxFiles := []string{}

	logger.Info(LogMsgLookingForXLSXFiles, ".")
	// Check for XLSX files in current directory
	if files, err := os.ReadDir("."); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
				xlsxFiles = append(xlsxFiles, file.Name())
			}
		}
	}

	logger.Info(LogMsgLookingForXLSXFiles, "assets/attack")
	// Check for XLSX files in assets/attack/ subdirectory
	if files, err := os.ReadDir("assets/attack"); err == nil {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
				xlsxFiles = append(xlsxFiles, "assets/attack/"+file.Name())
			}
		}
	} else {
		logger.Info(LogMsgLookingForXLSXFiles, "assets")
		// Check for XLSX files in assets/ directory if attack subdirectory doesn't exist
		if files, err := os.ReadDir("assets"); err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xlsx") {
					xlsxFiles = append(xlsxFiles, "assets/"+file.Name())
				}
			}
		}
	}

	logger.Info(LogMsgFoundXLSXFiles, xlsxFiles)

	// Import each XLSX file found
	for i, xlsxFile := range xlsxFiles {
		logger.Info(LogMsgStartingImportProcess, xlsxFile)
		logger.Info("Starting ATT&CK import process (%d/%d): %s", i+1, len(xlsxFiles), xlsxFile)

		// Check if the file exists before importing
		logger.Info(LogMsgCheckingFileExistence, xlsxFile)
		if _, err := os.Stat(xlsxFile); os.IsNotExist(err) {
			logger.Warn(LogMsgFileDoesNotExist, xlsxFile)
			continue
		} else {
			logger.Info(LogMsgFileExists, xlsxFile)
		}

		// Perform the import
		logger.Info("Beginning ATT&CK import from XLSX file: %s", xlsxFile)
		if err := attackStore.ImportFromXLSX(xlsxFile, false); err != nil {
			logger.Warn(LogMsgImportProcessFailed, xlsxFile, err)
			logger.Error("ATT&CK import failed for file %s: %v", xlsxFile, err)
		} else {
			logger.Info(LogMsgImportProcessCompleted, xlsxFile)
			logger.Info("ATT&CK import completed successfully for file: %s", xlsxFile)
		}
	}

	if len(xlsxFiles) == 0 {
		logger.Info(LogMsgNoXLSXFilesForImport)
		logger.Info("No XLSX files found for ATT&CK import")
	} else {
		logger.Info("ATT&CK import process completed for %d files", len(xlsxFiles))
	}

	logger.Info("ATT&CK data import at startup completed")
}

func main() {
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "local"
	}
	common.Info(LogMsgProcessIDConfigured, processID)

	// Use a bootstrap logger for initial messages before the full logging system is ready
	bootstrapLogger := common.NewLogger(os.Stderr, "", common.InfoLevel)
	common.Info(LogMsgBootstrapLoggerCreated)

	// Use subprocess package for logging to ensure build-time log level from .config is used
	logLevel := subprocess.DefaultBuildLogLevel()
	logger, err := subprocess.SetupLogging(processID, common.DefaultLogsDir, logLevel)
	if err != nil {
		bootstrapLogger.Error(LogMsgFailedSetupLogging, err)
		os.Exit(1)
	}
	common.Info(LogMsgLoggingSetupComplete, logLevel)

	// Get database path from environment or use default
	dbPath := os.Getenv("CVE_DB_PATH")
	if dbPath == "" {
		dbPath = "cve.db"
	}
	logger.Info(LogMsgDatabasePathConfigured, dbPath)

	// Create or open the database
	db, err := local.NewDB(dbPath)
	if err != nil {
		logger.Error(LogMsgFailedOpenDB, err)
		os.Exit(1)
	}
	logger.Info(LogMsgDatabaseOpened, dbPath)
	defer func() {
		logger.Info(LogMsgDatabaseClosing, dbPath)
		db.Close()
	}()

	// Initialize CWE store (using a separate DB file or the same as CVE)
	cweDBPath := os.Getenv("CWE_DB_PATH")
	if cweDBPath == "" {
		cweDBPath = "cwe.db"
	}
	logger.Info(LogMsgCWEDatabasePathConfigured, cweDBPath)
	cweStore, err := cwe.NewLocalCWEStore(cweDBPath)
	if err != nil {
		logger.Error(LogMsgFailedOpenCWEDB, err)
		os.Exit(1)
	}
	logger.Info(LogMsgCWEDatabaseOpened, cweDBPath)

	// Initialize CAPEC store (using CAPEC_DB_PATH env var)
	capecDBPath := os.Getenv("CAPEC_DB_PATH")
	if capecDBPath == "" {
		capecDBPath = "capec.db"
	}
	logger.Info(LogMsgCAPECDatabasePathConfigured, capecDBPath)
	capecStore, err := capec.NewLocalCAPECStore(capecDBPath)
	if err != nil {
		logger.Error(LogMsgFailedOpenCAPECDB, err)
		os.Exit(1)
	}
	logger.Info(LogMsgCAPECDatabaseOpened, capecDBPath)

	// Initialize ATT&CK store (using ATTACK_DB_PATH env var)
	attackDBPath := os.Getenv("ATTACK_DB_PATH")
	if attackDBPath == "" {
		attackDBPath = "attack.db"
	}
	logger.Info(LogMsgATTACKDatabasePathConfigured, attackDBPath)
	attackStore, err := attack.NewLocalAttackStore(attackDBPath)
	if err != nil {
		logger.Error(LogMsgFailedOpenATTACKDB, err)
		os.Exit(1)
	}
	logger.Info(LogMsgATTACKDatabaseOpened, attackDBPath)

	// Import CWEs from JSON file at startup (if file exists)
	// Removed duplicate importCWEsAtStartup definition; now only in cwe_handlers.go

	// Import ATT&CK data from XLSX files at startup (if files exist)
	logger.Info(LogMsgImportATTACKAtStartup)
	importATTACKDataAtStartup(attackStore, logger)
	logger.Info(LogMsgImportATTACKStartupCompleted)

	// Log completion of all startup activities
	logger.Info("All startup import processes completed")

	// Initialize notes service and run migrations
	logger.Info("Initializing notes service and running migrations...")
	notesServiceContainer := notes.NewServiceContainer(db.GormDB())
	// Run the notes table migrations to ensure tables exist
	if err := notes.MigrateNotesTables(db.GormDB()); err != nil {
		logger.Error("Failed to migrate notes tables: %v", err)
		os.Exit(1)
	}
	logger.Info("Notes service initialized and migrations completed")

	// Create subprocess instance
	var sp *subprocess.Subprocess

	// Check if we're running as an RPC subprocess with file descriptors
	if os.Getenv("BROKER_PASSING_RPC_FDS") == "1" {
		// Use file descriptors 3 and 4 for RPC communication
		inputFD := 3
		outputFD := 4

		// Allow environment override for file descriptors
		if val := os.Getenv("RPC_INPUT_FD"); val != "" {
			if fd, err := strconv.Atoi(val); err == nil {
				inputFD = fd
			}
		}
		if val := os.Getenv("RPC_OUTPUT_FD"); val != "" {
			if fd, err := strconv.Atoi(val); err == nil {
				outputFD = fd
			}
		}

		sp = subprocess.NewWithFDs(processID, inputFD, outputFD)
	} else {
		// Use default stdin/stdout for non-RPC mode
		sp = subprocess.New(processID)
	}

	logger.Info(LogMsgSubprocessCreated, processID)

	// Register RPC handlers
	logger.Info("Registering RPC handlers...")
	sp.RegisterHandler("RPCSaveCVEByID", createSaveCVEByIDHandler(db, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCSaveCVEByID")
	sp.RegisterHandler("RPCIsCVEStoredByID", createIsCVEStoredByIDHandler(db, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCIsCVEStoredByID")
	sp.RegisterHandler("RPCGetCVEByID", createGetCVEByIDHandler(db, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCVEByID")
	sp.RegisterHandler("RPCDeleteCVEByID", createDeleteCVEByIDHandler(db, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCDeleteCVEByID")
	sp.RegisterHandler("RPCListCVEs", createListCVEsHandler(db, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListCVEs")
	sp.RegisterHandler("RPCCountCVEs", createCountCVEsHandler(db, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCCountCVEs")
	sp.RegisterHandler("RPCGetCWEByID", createGetCWEByIDHandler(cweStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCWEByID")
	sp.RegisterHandler("RPCListCWEs", createListCWEsHandler(cweStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListCWEs")
	sp.RegisterHandler("RPCImportCWEs", createImportCWEsHandler(cweStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCImportCWEs")
	sp.RegisterHandler("RPCImportCAPECs", createImportCAPECsHandler(capecStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCImportCAPECs")
	sp.RegisterHandler("RPCForceImportCAPECs", createForceImportCAPECsHandler(capecStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCForceImportCAPECs")
	sp.RegisterHandler("RPCListCAPECs", createListCAPECsHandler(capecStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListCAPECs")
	sp.RegisterHandler("RPCGetCAPECByID", createGetCAPECByIDHandler(capecStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCAPECByID")
	sp.RegisterHandler("RPCGetCAPECCatalogMeta", createGetCAPECCatalogMetaHandler(capecStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCAPECCatalogMeta")

	// Register CWE View handlers
	RegisterCWEViewHandlers(sp, cweStore, logger)
	logger.Info("CWE View handlers registered")

	// Register ATT&CK handlers
	sp.RegisterHandler("RPCImportATTACKs", createImportATTACKsHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCImportATTACKs")
	sp.RegisterHandler("RPCGetAttackTechnique", createGetAttackTechniqueHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackTechnique")
	sp.RegisterHandler("RPCGetAttackTactic", createGetAttackTacticHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackTactic")
	sp.RegisterHandler("RPCGetAttackMitigation", createGetAttackMitigationHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackMitigation")
	sp.RegisterHandler("RPCGetAttackSoftware", createGetAttackSoftwareHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackSoftware")
	sp.RegisterHandler("RPCGetAttackGroup", createGetAttackGroupHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackGroup")
	sp.RegisterHandler("RPCGetAttackTechniqueByID", createGetAttackTechniqueByIDHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackTechniqueByID")
	sp.RegisterHandler("RPCGetAttackTacticByID", createGetAttackTacticByIDHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackTacticByID")
	sp.RegisterHandler("RPCGetAttackMitigationByID", createGetAttackMitigationByIDHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackMitigationByID")
	sp.RegisterHandler("RPCGetAttackSoftwareByID", createGetAttackSoftwareByIDHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackSoftwareByID")
	sp.RegisterHandler("RPCGetAttackGroupByID", createGetAttackGroupByIDHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackGroupByID")
	sp.RegisterHandler("RPCListAttackTechniques", createListAttackTechniquesHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListAttackTechniques")
	sp.RegisterHandler("RPCListAttackTactics", createListAttackTacticsHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListAttackTactics")
	sp.RegisterHandler("RPCListAttackMitigations", createListAttackMitigationsHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListAttackMitigations")
	sp.RegisterHandler("RPCListAttackSoftware", createListAttackSoftwareHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListAttackSoftware")
	sp.RegisterHandler("RPCListAttackGroups", createListAttackGroupsHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListAttackGroups")
	sp.RegisterHandler("RPCGetAttackImportMetadata", createGetAttackImportMetadataHandler(attackStore, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetAttackImportMetadata")

	// Register Notes service handlers
	notes.NewRPCHandlers(notesServiceContainer, sp, logger)
	logger.Info("Notes service handlers registered")

	logger.Info(LogMsgServiceStarting, processID)
	logger.Info(LogMsgServiceStarted)
	logger.Info(LogMsgServiceReady)

	// Run with default lifecycle management
	logger.Info("Starting subprocess with default lifecycle management")
	logger.Info("Local service entering main loop, ready to handle requests")
	subprocess.RunWithDefaults(sp, logger)
	logger.Info(LogMsgServiceShutdownStarting)
	logger.Info(LogMsgServiceShutdownComplete)
}
