package cce

const (
	// CCE Log Messages
	LogMsgCCEStoreCreated       = "CCE store created successfully"
	LogMsgCCEStoreClosed        = "CCE store closed"
	LogMsgParsingStarted        = "Parsing CCE data"
	LogMsgParsingCompleted      = "Parsing CCE data completed"
	LogMsgParsingFailed         = "Failed to parse CCE data: %v"
	LogMsgDatabaseQueryExecuted = "Database query executed: table=cce, operation=%s"

	// CCE ID Format
	CCEIDPrefix = "CCE-"
	CCEIDLength = 10

	// Validation
	MaxDescriptionLength = 500
	MaxReferenceCount    = 20
)
