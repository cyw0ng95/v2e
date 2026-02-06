package cwe

// These variables are injected at build time via ldflags
var (
	buildCWERawPath = "assets/cwe-raw.json" // Default CWE raw path, can be overridden with -ldflags "-X cwe.buildCWERawPath=assets/cwe-raw.json"
)

// DefaultBuildCWERawPath returns the default CWE raw path based on build configuration
func DefaultBuildCWERawPath() string {
	return buildCWERawPath
}
