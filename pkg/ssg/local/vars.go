// Package local provides build-time configuration for SSG local storage.
package local

// These variables are injected at build time via ldflags
var buildSSGDBPath string // SSG database path

// DefaultDBPath returns the default SSG database path based on build configuration.
// If overridden at build time, returns the configured value.
func DefaultDBPath() string {
	if buildSSGDBPath != "" {
		return buildSSGDBPath
	}
	return "ssg.db"
}
