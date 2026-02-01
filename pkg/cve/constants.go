package cve

const (
	// Job Controller Log Messages
	LogMsgJobStarted              = "Job started: session_id=%s, start_index=%d, batch_size=%d"
	LogMsgJobStopped              = "Job stopped"
	LogMsgJobPaused               = "Job paused"
	LogMsgJobResumed              = "Job resumed: session_id=%s"
	LogMsgJobLoopStarting         = "Job loop starting: start_index=%d, batch_size=%d"
	LogMsgJobLoopCancelled        = "Job loop cancelled"
	LogMsgFetchingBatch           = "Fetching batch: start_index=%d, batch_size=%d"
	LogMsgFailedFetchCVEs         = "Failed to fetch CVEs: %v"
	LogMsgFailedUpdateProgress    = "Failed to update progress: %v"
	LogMsgInvalidResponseType     = "Invalid response type from remote"
	LogMsgErrorFromRemote         = "Error from remote: %s"
	LogMsgFailedUnmarshalResponse = "Failed to unmarshal CVE response: %v"
	LogMsgNoMoreCVEs              = "No more CVEs to fetch. Job completed."
	LogMsgFetchedCVEs             = "Fetched %d CVEs from NVD"
	LogMsgFailedStoreCVE          = "Failed to store CVE %s: %v"
	LogMsgStoredCVEsSuccess       = "Stored %d/%d CVEs successfully"

	// Taskflow Executor Log Messages
	LogMsgTFJobStarted        = "Job started: run_id=%s, start_index=%d, batch_size=%d, data_type=%s"
	LogMsgTFJobResumed        = "Job resumed: run_id=%s"
	LogMsgTFJobPaused         = "Job paused: run_id=%s"
	LogMsgTFJobStopped        = "Job stopped: run_id=%s"
	LogMsgTFNoActiveRuns      = "No active runs to recover"
	LogMsgTFFoundRun          = "Found run to recover: id=%s, state=%s"
	LogMsgTFAutoRecover       = "Auto-recovering running job: %s"
	LogMsgTFManualResume      = "Run is %s - manual resume required"
	LogMsgTFFailedGetRun      = "Failed to get run: %v"
	LogMsgTFJobLoopStarting   = "Job loop starting: run_id=%s, start_index=%d, batch_size=%d"
	LogMsgTFJobLoopCancelled  = "Job loop cancelled: run_id=%s"
	LogMsgTFFetchingBatch     = "Fetching batch: run_id=%s, index=%d, size=%d"
	LogMsgTFSkippingStore     = "Skipping store due to fetch error: %v"
	LogMsgTFNoMoreCVEs        = "No more CVEs to fetch. Job completed: run_id=%s"
	LogMsgTFFailedStoreCVE    = "Failed to store CVE %s: %v"
	LogMsgTFStoredCVEsSuccess = "Stored %d/%d CVEs successfully"
	LogMsgTFFetchFailed       = "Fetch failed: %v"
	LogMsgTFJobCompleted      = "Job completed: run_id=%s"

	// Session Management Log Messages
	LogMsgSessionCreated         = "Session created: id=%s"
	LogMsgSessionNotFound        = "Session not found: id=%s"
	LogMsgSessionUpdated         = "Session updated: id=%s, status=%s"
	LogMsgSessionDeleted         = "Session deleted: id=%s"
	LogMsgSessionListRetrieved   = "Session list retrieved: count=%d"
	LogMsgSessionProgressUpdated = "Session progress updated: id=%s, progress=%.2f%%"

	// Database Operations Log Messages
	LogMsgDatabaseConnected           = "Database connected successfully"
	LogMsgDatabaseConnectionError     = "Database connection error: %v"
	LogMsgDatabaseQueryExecuted       = "Database query executed: table=%s, operation=%s"
	LogMsgDatabaseTransactionBegin    = "Database transaction started"
	LogMsgDatabaseTransactionCommit   = "Database transaction committed"
	LogMsgDatabaseTransactionRollback = "Database transaction rolled back: %v"

	// Remote API Operations Log Messages
	LogMsgAPIRequestSent         = "API request sent: endpoint=%s, method=%s"
	LogMsgAPIResponseReceived    = "API response received: status=%d, duration=%v"
	LogMsgAPIRateLimitHit        = "API rate limit hit, retrying in %v"
	LogMsgAPIAuthenticationError = "API authentication error: %v"
	LogMsgAPITimeout             = "API request timeout: endpoint=%s"

	// Data Processing Log Messages
	LogMsgDataValidationStarted  = "Data validation started: record_count=%d"
	LogMsgDataValidationComplete = "Data validation completed: valid=%d, invalid=%d"
	LogMsgDataTransformation     = "Data transformation applied: type=%s"
	LogMsgDataDeduplication      = "Duplicate records removed: count=%d"
	LogMsgDataSortingComplete    = "Data sorting completed: field=%s, order=%s"

	// Cache Operations Log Messages
	LogMsgCacheHit     = "Cache hit: key=%s"
	LogMsgCacheMiss    = "Cache miss: key=%s"
	LogMsgCacheSet     = "Cache set: key=%s, ttl=%v"
	LogMsgCacheEvicted = "Cache evicted: key=%s"
	LogMsgCacheCleared = "Cache cleared: pattern=%s"
)
