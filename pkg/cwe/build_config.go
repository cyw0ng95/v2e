package cwe

// These variables are injected at build time via ldflags
var (
	buildCWEDBPath = "cwe.db" // Default CWE DB path, can be overridden with -ldflags "-X cwe.buildCWEDBPath=cwe.db"
)

// DefaultBuildCWEDBPath returns the default CWE DB path based on build configuration
func DefaultBuildCWEDBPath() string {
	return buildCWEDBPath
}
