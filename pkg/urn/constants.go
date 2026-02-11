package urn

const (
	// URN Format
	URNPrefix    = "urn:"
	URNNamespace = "cyw0ng95"
	URNDelimiter = ":"

	// Supported Types
	URNTypeCVE    = "cve"
	URNTypeCWE    = "cwe"
	URNTypeCAPEC  = "capec"
	URNTypeATTACK = "attack"

	// Validation
	MaxURNLength = 256
	MinURNLength = 10
)
