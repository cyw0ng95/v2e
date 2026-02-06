package capec

const (
	// Import operation messages
	LogMsgImportingXML  = "Importing CAPEC data from XML file: %s"
	LogMsgImportSkipped = "CAPEC catalog version %s already imported; skipping import"

	// Database operation messages
	LogMsgDatabaseOpened     = "CAPEC database opened successfully: %s"
	LogMsgDatabaseClosed     = "CAPEC database closed: %s"
	LogMsgTransactionStarted = "CAPEC database transaction started"
	LogMsgTransactionFailed  = "CAPEC database transaction failed: %v"
	LogMsgTransactionSaved   = "CAPEC database transaction saved"

	// Data processing messages
	LogMsgParsingXMLStarted     = "Parsing CAPEC XML data started"
	LogMsgParsingXMLCompleted   = "Parsing CAPEC XML data completed, entries: %d"
	LogMsgProcessingEntry       = "Processing CAPEC entry: %s"
	LogMsgEntryProcessed        = "CAPEC entry processed successfully: %s"
	LogMsgEntryProcessingFailed = "Failed to process CAPEC entry %s: %v"
	LogMsgBulkInsertStarted     = "Starting bulk insert of CAPEC entries: %d items"
	LogMsgBulkInsertCompleted   = "Bulk insert completed: %d/%d entries inserted successfully"
	LogMsgBulkInsertFailed      = "Bulk insert failed: %v"

	// Query operation messages
	LogMsgQueryStarted   = "Executing CAPEC query: %s"
	LogMsgQueryCompleted = "CAPEC query completed, results: %d"
	LogMsgQueryFailed    = "CAPEC query failed: %v"
	LogMsgCacheHit       = "CAPEC cache hit for query: %s"
	LogMsgCacheMiss      = "CAPEC cache miss for query: %s"
	LogMsgCacheUpdated   = "CAPEC cache updated with %d entries"
)
