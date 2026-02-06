package ssg

const (
	// SSG Log Messages
	LogMsgSSGStoreCreated       = "SSG store created successfully"
	LogMsgSSGStoreClosed        = "SSG store closed"
	LogMsgParsingStarted        = "Parsing SSG data"
	LogMsgParsingCompleted      = "Parsing SSG data completed"
	LogMsgParsingFailed         = "Failed to parse SSG data: %v"
	LogMsgDatabaseQueryExecuted = "Database query executed: table=ssg, operation=%s"

	// SSG Format
	SSGIDPrefix      = "SSG-"
	SSGIDLength      = 10
	SSGVersionFormat = "v%d.%d"

	// Validation
	MaxTitleLength       = 200
	MaxDescriptionLength = 1000
	MaxReferenceCount    = 20
)
