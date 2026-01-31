package cwe

// These variables are injected at build time via ldflags
var (
	buildViewURL = "https://github.com/CWE-CAPEC/REST-API-wg/archive/refs/heads/main.zip" // Default view URL, can be overridden with -ldflags "-X cwe.buildViewURL=https://example.com/view"
)

// DefaultBuildViewURL returns the default view URL based on build configuration
func DefaultBuildViewURL() string {
	return buildViewURL
}