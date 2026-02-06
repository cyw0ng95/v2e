package cwe

const (
	// Job Controller Log Messages
	LogMsgJobStarted              = "CWE view job started: session_id=%s"
	LogMsgJobStopped              = "CWE view job stopped: session_id=%s"
	LogMsgJobLoopStarting         = "CWE job loop starting: start_index=%d, page_size=%d"
	LogMsgJobLoopCancelled        = "CWE job loop cancelled"
	LogMsgFetchingViews           = "Fetching views: start_index=%d, page_size=%d"
	LogMsgFailedFetchViews        = "Failed to fetch views: %v"
	LogMsgInvalidResponseType     = "Invalid response type from remote for views"
	LogMsgErrorFromRemote         = "Error from remote: %s"
	LogMsgFailedUnmarshalResponse = "Failed to unmarshal views response: %v"
	LogMsgNoMoreViews             = "No more views to fetch. Job completed."
	LogMsgFailedSaveView          = "Failed to save view %s: %v"
	LogMsgFetchedViews            = "Fetched %d views and stored %d"

	// Local Store Log Messages
	LogMsgImportingJSON         = "Importing CWE data from JSON file: %s"
	LogMsgImportSkipped         = "CWE import skipped: first and last CWE already present (IDs: %s, %s)"
	LogMsgImportCompleted       = "CWE import completed successfully: file=%s, records=%d"
	LogMsgImportValidationError = "CWE import validation failed: %v"
	LogMsgDataFileNotFound      = "CWE data file not found: %s"
	LogMsgDataFileCorrupted     = "CWE data file corrupted: %s, error=%v"

	// Database Operations Log Messages
	LogMsgDatabaseConnected           = "Database connected successfully"
	LogMsgDatabaseConnectionError     = "Database connection error: %v"
	LogMsgDatabaseQueryExecuted       = "Database query executed: table=%s, operation=%s"
	LogMsgDatabaseTransactionBegin    = "Database transaction started"
	LogMsgDatabaseTransactionCommit   = "Database transaction committed"
	LogMsgDatabaseTransactionRollback = "Database transaction rolled back: %v"
	LogMsgBulkInsertStarted           = "Bulk insert started: table=%s, records=%d"
	LogMsgBulkInsertCompleted         = "Bulk insert completed: table=%s, inserted=%d, duration=%v"

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
