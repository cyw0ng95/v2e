package main

const (
	// Service Lifecycle Log Messages
	LogMsgStartup                 = "[meta] STARTUP: PROCESS_ID=%s SESSION_DB_PATH=%s"
	LogMsgFailedToSetupLogging    = "[meta] Failed to setup logging: %v"
	LogMsgServiceStarted          = "[meta] CVE meta service started - orchestrates local and remote"
	LogMsgServiceReady            = "[meta] Meta service ready and accepting requests"
	LogMsgServiceShutdownStarting = "[meta] Meta service shutdown starting"
	LogMsgServiceShutdownComplete = "[meta] Meta service shutdown completed"

	// Configuration Log Messages
	LogMsgProcessIDConfigured    = "[meta] Process ID configured: %s"
	LogMsgBootstrapLoggerCreated = "[meta] Bootstrap logger created"
	LogMsgLoggingSetupComplete   = "[meta] Logging setup completed with level: %v"
	LogMsgRunDBPathConfigured    = "[meta] Run DB path configured: %s"
	LogMsgRunDBPathDefaultUsed   = "[meta] Run DB path default used: %s"
	LogMsgRunDBStatInfo          = "[meta] Run DB stat info: size=%d, modTime=%v"
	LogMsgUsingRunDBPath         = "[meta] Using run DB path: %s"
	LogMsgRunDBFileExists        = "[meta] Run DB file exists: %s"
	LogMsgRunDBFileDoesNotExist  = "[meta] Run DB file does not exist or cannot stat: %s (err=%v)"

	// Storage Operations Log Messages
	LogMsgCreatingRunStore       = "[meta] Creating run store..."
	LogMsgFailedToCreateRunStore = "[meta] Failed to create run store: %v"
	LogMsgRunStoreCreated        = "[meta] Run store created successfully"
	LogMsgRunStoreOpening        = "[meta] Opening run store at path: %s"
	LogMsgRunStoreOpened         = "[meta] Run store opened successfully"
	LogMsgRunStoreClosing        = "[meta] Closing run store"

	// Component Initialization Log Messages
	LogMsgSubprocessCreated       = "[meta] Subprocess created with ID: %s"
	LogMsgRPCClientCreated        = "[meta] RPC client created"
	LogMsgRPCAdapterCreated       = "[meta] RPC adapter created"
	LogMsgJobExecutorCreated      = "[meta] Job executor created with concurrency: %d"
	LogMsgCWEJobControllerCreated = "[meta] CWE job controller created"
	LogMsgRPCHandlerRegistered    = "[meta] RPC handler registered: %s"

	// Import Operations Log Messages
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

	// RPC Handler Log Messages
	LogMsgProcessingGetCVEFailed = "Processing GetCVE request failed due to malformed payload: %s"
	LogMsgCVEIDRequired          = "cve_id is required but was empty or missing"

	// Job Management Log Messages
	LogMsgRunRecoveryStarted    = "[meta] Run recovery started"
	LogMsgRunRecoveryCompleted  = "[meta] Run recovery completed"
	LogMsgRunRecoveryFailed     = "[meta] Run recovery failed: %v"
	LogMsgCWEImportTriggered    = "[meta] CWE import triggered on local with path: %s"
	LogMsgCWEImportSkipped      = "[meta] CWE import skipped, timeout exceeded"
	LogMsgCAPECImportTriggered  = "[meta] CAPEC import triggered on local with path: %s"
	LogMsgCAPECImportSkipped    = "[meta] CAPEC import skipped, already present"
	LogMsgATTACKImportTriggered = "[meta] ATT&CK import triggered on local with path: %s"
	LogMsgATTACKImportSkipped   = "[meta] ATT&CK import skipped, timeout exceeded"
	LogMsgCWERecoverRunsCalled  = "[meta] RecoverRuns called to recover any running jobs"
	LogMsgCWERecoverRunsSuccess = "[meta] RecoverRuns completed successfully"
	LogMsgCWERecoverRunsError   = "[meta] RecoverRuns encountered error: %v"

	// CVE Operations Log Messages
	LogMsgStartingGetCVE      = "[meta] Starting GetCVE request for CVE ID: %s"
	LogMsgGetCVESuccessLocal  = "[meta] GetCVE found in local storage: %s"
	LogMsgGetCVEFetchRemote   = "[meta] GetCVE not found locally, fetching from remote: %s"
	LogMsgGetCVESuccessRemote = "[meta] GetCVE fetched from remote, saving to local: %s"
	LogMsgGetCVEFetchFailed   = "[meta] GetCVE failed to fetch from remote: %s"
	LogMsgGetCVENotFound      = "[meta] GetCVE not found in either local or remote: %s"
	LogMsgStartingCreateCVE   = "[meta] Starting CreateCVE request for CVE ID: %s"
	LogMsgCreateCVESuccess    = "[meta] CreateCVE completed successfully: %s"
	LogMsgCreateCVEFailed     = "[meta] CreateCVE failed: %v"
	LogMsgStartingUpdateCVE   = "[meta] Starting UpdateCVE request for CVE ID: %s"
	LogMsgUpdateCVESuccess    = "[meta] UpdateCVE completed successfully: %s"
	LogMsgUpdateCVEFailed     = "[meta] UpdateCVE failed: %v"
	LogMsgStartingDeleteCVE   = "[meta] Starting DeleteCVE request for CVE ID: %s"
	LogMsgDeleteCVESuccess    = "[meta] DeleteCVE completed successfully: %s"
	LogMsgDeleteCVEFailed     = "[meta] DeleteCVE failed: %v"
	LogMsgStartingListCVEs    = "[meta] Starting ListCVEs request with offset: %d, limit: %d"
	LogMsgListCVEsSuccess     = "[meta] ListCVEs completed successfully, returned: %d records"
	LogMsgListCVEsFailed      = "[meta] ListCVEs failed: %v"
	LogMsgStartingCountCVEs   = "[meta] Starting CountCVEs request"
	LogMsgCountCVESuccess     = "[meta] CountCVEs completed successfully, total: %d"
	LogMsgCountCVEFailed      = "[meta] CountCVEs failed: %v"

	// Session Management Log Messages
	LogMsgStartingDataPopulation  = "[meta] Starting data population for type: %s"
	LogMsgDataPopulationSuccess   = "[meta] Data population completed successfully for type: %s, session: %s"
	LogMsgDataPopulationFailed    = "[meta] Data population failed for type: %s, error: %v"
	LogMsgStartingSession         = "[meta] Starting new session with type: %s, index: %d, batch: %d"
	LogMsgSessionStarted          = "[meta] Session started successfully: %s"
	LogMsgSessionStartFailed      = "[meta] Session start failed: %v"
	LogMsgSessionStatusRequested  = "[meta] Session status requested"
	LogMsgSessionStatusReturned   = "[meta] Session status returned: has_session=%t"
	LogMsgSessionPauseRequested   = "[meta] Session pause requested"
	LogMsgSessionPaused           = "[meta] Session paused successfully"
	LogMsgSessionPauseFailed      = "[meta] Session pause failed: %v"
	LogMsgSessionResumeRequested  = "[meta] Session resume requested"
	LogMsgSessionResumed          = "[meta] Session resumed successfully"
	LogMsgSessionResumeFailed     = "[meta] Session resume failed: %v"
	LogMsgSessionStopRequested    = "[meta] Session stop requested"
	LogMsgSessionStopped          = "[meta] Session stopped successfully"
	LogMsgSessionStopFailed       = "[meta] Session stop failed: %v"
	LogMsgStartingTypedSession    = "[meta] Starting typed session: id=%s, type=%s, index=%d, batch=%d"
	LogMsgTypedSessionStarted     = "[meta] Typed session started successfully: %s"
	LogMsgTypedSessionStartFailed = "[meta] Typed session start failed: %v"

	// CWE View Job Log Messages
	LogMsgCWEViewJobStartRequested = "[meta] CWE view job start requested"
	LogMsgCWEViewJobStarted        = "[meta] CWE view job started: %s"
	LogMsgCWEViewJobStartFailed    = "[meta] CWE view job start failed: %v"
	LogMsgCWEViewJobStopRequested  = "[meta] CWE view job stop requested for session: %s"
	LogMsgCWEViewJobStopped        = "[meta] CWE view job stopped successfully: %s"
	LogMsgCWEViewJobStopFailed     = "[meta] CWE view job stop failed: %v"

	// RPC Communication Log Messages
	LogMsgRPCInvokeStarted           = "[meta] RPC invoke started: target=%s, method=%s"
	LogMsgRPCInvokeCompleted         = "[meta] RPC invoke completed: target=%s, method=%s"
	LogMsgRPCResponseReceived        = "[meta] RPC response received: correlationID=%s, type=%s"
	LogMsgRPCPendingRequestAdded     = "[meta] Added pending RPC request: correlationID=%s"
	LogMsgRPCPendingRequestRemoved   = "[meta] Removed pending RPC request: correlationID=%s"
	LogMsgRPCChannelSignaled         = "[meta] RPC response channel signaled: correlationID=%s"
	LogMsgRPCSendMessageStarted      = "[meta] About to send RPC message: type=%s, id=%s, target=%s, correlationID=%s"
	LogMsgRPCSendMessageSuccess      = "[meta] Successfully sent RPC message: type=%s, id=%s, target=%s, correlationID=%s"
	LogMsgRPCSendMessageFailed       = "[meta] Failed to send RPC message: type=%s, id=%s, target=%s, correlationID=%s, error=%v"
	LogMsgRPCHandlerCalled           = "[meta] RPC handler called: %s"
	LogMsgRPCRequestReceived         = "[meta] RPC request received: type=%s, id=%s, source=%s, correlationID=%s"
	LogMsgRPCResponseSent            = "[meta] RPC response sent: type=%s, id=%s, target=%s, correlationID=%s"
	LogMsgRPCErrorSent               = "[meta] RPC error response sent: type=%s, id=%s, target=%s, correlationID=%s, error=%s"
	LogMsgRPCClientHandlerRegistered = "[meta] RPC client handler registered: %s"

	// Subprocess Operations Log Messages
	LogMsgSubprocessRunStarted   = "[meta] Subprocess run started"
	LogMsgSubprocessRunCompleted = "[meta] Subprocess run completed"
	LogMsgSubprocessRunError     = "[meta] Subprocess run error: %v"

	// Broker Connection Log Messages
	LogMsgBrokerConnectionAttempt = "[meta] Attempting connection to broker"
	LogMsgBrokerConnected         = "[meta] Connected to broker successfully"
	LogMsgBrokerConnectionFailed  = "[meta] Failed to connect to broker: %v"
)
