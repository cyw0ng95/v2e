package ssg

// SSG version and URLs
const (
	// DefaultSSGVersion is the default SSG version to fetch
	DefaultSSGVersion = "0.1.79"

	// SSGReleaseURLTemplate is the GitHub release URL template
	SSGReleaseURLTemplate = "https://github.com/ComplianceAsCode/content/releases/download/v%s/scap-security-guide-%s.tar.gz"

	// SSGSHA512URLTemplate is the SHA512 checksum URL template
	SSGSHA512URLTemplate = "https://github.com/ComplianceAsCode/content/releases/download/v%s/scap-security-guide-%s.tar.gz.sha512"

	// SSGXMLPattern is the pattern for SSG DataStream XML files
	SSGXMLPattern = "ssg-*-ds.xml"

	// XCCDFNamespace is the XCCDF 1.2 namespace
	XCCDFNamespace = "http://checklists.nist.gov/xccdf/1.2"
)

// Log messages
const (
	LogMsgSSGDatabaseOpened          = "SSG database opened at %s"
	LogMsgSSGDatabasePathConfigured  = "SSG database path configured: %s"
	LogMsgSSGDocPathConfigured       = "SSG document path configured: %s"
	LogMsgFailedOpenSSGDB            = "Failed to open SSG database: %v"
	LogMsgSSGDatabaseClosing         = "Closing SSG database at %s"
	LogMsgSSGPackageDeployed         = "SSG package deployed to %s"
	LogMsgSSGPackageExtracting       = "Extracting SSG package to %s"
	LogMsgSSGPackageExtracted        = "SSG package extracted successfully"
	LogMsgSSGProfilesLoaded          = "Loaded %d SSG profiles from %s"
	LogMsgSSGRulesLoaded             = "Loaded %d SSG rules from %s"
	LogMsgSSGBenchmarkParsed         = "Parsed SSG benchmark: %s"
	LogMsgSSGParsingFile             = "Parsing SSG file: %s"
	LogMsgSSGParsingCompleted        = "SSG parsing completed successfully"
	LogMsgSSGProfileNotFound         = "SSG profile not found: %s"
	LogMsgSSGRuleNotFound            = "SSG rule not found: %s"
)
