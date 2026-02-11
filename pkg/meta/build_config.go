package meta

// These variables are injected at build time via ldflags
var (
	buildSessionDBPath = "session.db" // Default session DB path, can be overridden with -ldflags "-X meta.buildSessionDBPath=session.db"
)

// DefaultBuildSessionDBPath returns the default session DB path based on build configuration
func DefaultBuildSessionDBPath() string {
	return buildSessionDBPath
}
