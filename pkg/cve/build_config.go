package cve

// These variables are injected at build time via ldflags
var (
	buildCVEDBPath = "cve.db" // Default CVE DB path, can be overridden with -ldflags "-X cve.buildCVEDBPath=cve.db"
)

// DefaultBuildCVEDBPath returns the default CVE DB path based on build configuration
func DefaultBuildCVEDBPath() string {
	return buildCVEDBPath
}
