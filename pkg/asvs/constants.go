package asvs

const (
	// ASVS Log Messages
	LogMsgASVSStoreCreated      = "ASVS store created successfully"
	LogMsgASVSStoreClosed       = "ASVS store closed"
	LogMsgCSVImportStarted      = "Importing ASVS data from CSV file: %s"
	LogMsgCSVImportCompleted    = "ASVS import completed: %d requirements imported"
	LogMsgCSVImportFailed       = "Failed to import ASVS data: %v"
	LogMsgCSVFileNotFound       = "CSV file not found: %s"
	LogMsgCSVReadFailed         = "Failed to read CSV file: %v"
	LogMsgCSVValidationFailed   = "CSV validation failed: %v"
	LogMsgRequirementFound      = "Requirement found: ID=%s"
	LogMsgRequirementNotFound   = "Requirement not found: ID=%s"
	LogMsgRequirementsListed    = "Listed %d ASVS requirements"
	LogMsgRequirementsByCWE     = "Found %d requirements matching CWE ID: %s"
	LogMsgRequirementsCounted   = "Total ASVS requirements: %d"
	LogMsgDatabaseQueryExecuted = "Database query executed: table=asvs_requirements, operation=%s"
)
