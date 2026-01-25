package main

const (
	// Log messages
	LogMsgStartup                   = "[meta] STARTUP: PROCESS_ID=%s SESSION_DB_PATH=%s"
	LogMsgFailedToSetupLogging      = "[meta] Failed to setup logging: %v"
	LogMsgUsingRunDBPath            = "[meta] Using run DB path: %s"
	LogMsgRunDBFileExists           = "[meta] Run DB file exists: %s"
	LogMsgRunDBFileDoesNotExist     = "[meta] Run DB file does not exist or cannot stat: %s (err=%v)"
	LogMsgCreatingRunStore          = "[meta] Creating run store..."
	LogMsgFailedToCreateRunStore    = "[meta] Failed to create run store: %v"
	LogMsgRunStoreCreated           = "[meta] Run store created successfully"
	LogMsgServiceStarted            = "[meta] CVE meta service started - orchestrates local and remote"
	LogMsgFailedToImportCWELocal    = "Failed to import CWE on local: %v"
	LogMsgCWEImportProcessFailed    = "CWE import process failed: %v"
	LogMsgFailedToQueryCAPECCatalog = "Failed to query CAPEC catalog meta on local: %v"
	LogMsgCAPECAlreadyPresent       = "CAPEC catalog already present on local; skipping automatic import"
	LogMsgFailedToImportCAPECLocal  = "Failed to import CAPEC on local: %v"
	LogMsgCAPECImportProcessFailed  = "CAPEC import process failed: %v"
	LogMsgFailedToQueryATTACKMeta   = "Failed to query ATT&CK meta on local: %v"
	LogMsgATTACKAlreadyPresent      = "ATT&CK data already present on local; skipping automatic import"
	LogMsgFailedToImportATTACKLocal = "Failed to import ATT&CK on local: %v"
	LogMsgATTACKImportProcessFailed = "ATT&CK import process failed: %v"

	// RPC Handler Logs
	LogMsgRPCStartCWEImport         = "RPCStartCWEImport: Starting CWE import job"
	LogMsgFailedToParseRequest      = "Failed to parse request: %v"
	LogMsgFailedToStartCWEImport    = "Failed to start CWE import: %v"
	LogMsgFailedToMarshalResponse   = "Failed to marshal response: %v"
	LogMsgSuccessStartCWEImport     = "RPCStartCWEImport: Successfully started CWE import job: %s"
	LogMsgRPCStartCAPECImport       = "RPCStartCAPECImport: Starting CAPEC import job"
	LogMsgFailedToStartCAPECImport  = "Failed to start CAPEC import: %v"
	LogMsgSuccessStartCAPECImport   = "RPCStartCAPECImport: Successfully started CAPEC import job: %s"
	LogMsgRPCStartATTACKImport      = "RPCStartATTACKImport: Starting ATT&CK import job"
	LogMsgFailedToStartATTACKImport = "Failed to start ATT&CK import: %v"
	LogMsgSuccessStartATTACKImport  = "RPCStartATTACKImport: Successfully started ATT&CK import job: %s"
	LogMsgProcessingGetCVEFailed    = "Processing GetCVE request failed due to malformed payload: %s"
	LogMsgCVEIDRequired             = "cve_id is required but was empty or missing"
)
