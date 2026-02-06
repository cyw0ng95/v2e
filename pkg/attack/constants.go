package attack

const (
	// Attack Pattern Operations Log Messages
	LogMsgAttackPatternRetrieved = "Attack pattern retrieved: id=%s"
	LogMsgAttackPatternNotFound  = "Attack pattern not found: id=%s"
	LogMsgAttackPatternsListed   = "Attack patterns listed: count=%d"
	LogMsgAttackPatternSearch    = "Attack pattern search performed: query=%s, results=%d"

	// Database Operations Log Messages
	LogMsgDatabaseConnected           = "Database connected successfully"
	LogMsgDatabaseConnectionError     = "Database connection error: %v"
	LogMsgDatabaseQueryExecuted       = "Database query executed: table=%s, operation=%s"
	LogMsgDatabaseTransactionBegin    = "Database transaction started"
	LogMsgDatabaseTransactionCommit   = "Database transaction committed"
	LogMsgDatabaseTransactionRollback = "Database transaction rolled back: %v"

	// Cache Operations Log Messages
	LogMsgCacheHit     = "Cache hit: key=%s"
	LogMsgCacheMiss    = "Cache miss: key=%s"
	LogMsgCacheSet     = "Cache set: key=%s, ttl=%v"
	LogMsgCacheEvicted = "Cache evicted: key=%s"
	LogMsgCacheCleared = "Cache cleared: pattern=%s"

	// Data Processing Log Messages
	LogMsgDataValidationStarted  = "Data validation started: record_count=%d"
	LogMsgDataValidationComplete = "Data validation completed: valid=%d, invalid=%d"
	LogMsgDataTransformation     = "Data transformation applied: type=%s"
	LogMsgDataDeduplication      = "Duplicate records removed: count=%d"
)
