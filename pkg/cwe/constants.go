package cwe

const (
	// Log messages
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

	// Local Store Log messages
	LogMsgImportingJSON = "Importing CWE data from JSON file: %s"
	LogMsgImportSkipped = "CWE import skipped: first and last CWE already present (IDs: %s, %s)"
)
