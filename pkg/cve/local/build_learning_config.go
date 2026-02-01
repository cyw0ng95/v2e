package local

// These variables are injected at build time via ldflags
var (
	buildLearningDBPath = "./learning.db" // Default learning DB path, can be overridden with -ldflags "-X local.buildLearningDBPath=./learning.db"
)

// DefaultBuildLearningDBPath returns the default learning DB path based on build configuration
func DefaultBuildLearningDBPath() string {
	return buildLearningDBPath
}
