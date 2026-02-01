package notes

const (
	// RPC Operations Log Messages
	LogMsgRPCError                = "RPC Error: %s"
	LogMsgRPCRequestReceived      = "RPC request received: method=%s, params=%v"
	LogMsgRPCMethodNotFound       = "RPC method not found: %s"
	LogMsgRPCInvalidParams        = "RPC invalid parameters: %v"
	LogMsgRPCInternalError        = "RPC internal error: %v"
	LogMsgRPCResponseSent         = "RPC response sent: method=%s, status=%s, duration=%v"

	// Database Operations Log Messages
	LogMsgDatabaseConnected       = "Database connected successfully"
	LogMsgDatabaseConnectionError = "Database connection error: %v"
	LogMsgDatabaseQueryExecuted   = "Database query executed: table=%s, operation=%s"
	LogMsgDatabaseTransactionBegin = "Database transaction started"
	LogMsgDatabaseTransactionCommit = "Database transaction committed"
	LogMsgDatabaseTransactionRollback = "Database transaction rolled back: %v"

	// Note Operations Log Messages
	LogMsgNoteCreated             = "Note created: id=%s, title=%s"
	LogMsgNoteUpdated             = "Note updated: id=%s, title=%s"
	LogMsgNoteDeleted             = "Note deleted: id=%s"
	LogMsgNoteRetrieved           = "Note retrieved: id=%s"
	LogMsgNoteListRetrieved       = "Note list retrieved: count=%d"
	LogMsgNoteSearchPerformed     = "Note search performed: query=%s, results=%d"

	// Tag Operations Log Messages
	LogMsgTagCreated              = "Tag created: id=%s, name=%s"
	LogMsgTagDeleted              = "Tag deleted: id=%s"
	LogMsgTagAssigned             = "Tag assigned: note_id=%s, tag_id=%s"
	LogMsgTagRemoved              = "Tag removed: note_id=%s, tag_id=%s"

	// Migration Operations Log Messages
	LogMsgMigrationStarted        = "Database migration started: version=%d"
	LogMsgMigrationCompleted      = "Database migration completed: version=%d, duration=%v"
	LogMsgMigrationFailed         = "Database migration failed: version=%d, error=%v"
	LogMsgSchemaUpToDate          = "Database schema is up to date: version=%d"
)
