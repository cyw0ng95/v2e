package capec

// These variables are injected at build time via ldflags
var (
	buildCAPECDBPath = "capec.db" // Default CAPEC DB path, can be overridden with -ldflags "-X capec.buildCAPECDBPath=capec.db"
)

// DefaultBuildCAPECDBPath returns the default CAPEC DB path based on build configuration
func DefaultBuildCAPECDBPath() string {
	return buildCAPECDBPath
}
