package capec

const (
	// Log messages
	LogMsgImportingXML       = "Importing CAPEC data from XML file: %s (schema: %s)"
	LogMsgImportSkipped      = "CAPEC catalog version %s already imported; skipping import"
	LogMsgSkippingValidation = "Skipping XSD validation (CAPEC_STRICT_XSD not set); continuing with permissive import"
)
