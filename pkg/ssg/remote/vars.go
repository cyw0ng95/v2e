// Package remote provides Git operations for SSG data fetching.
//
// These build-time variables are injected via ldflags during compilation.
package remote

// These variables are injected at build time via ldflags
var (
	buildRepoURL  string // Git repository URL for SSG data
	buildRepoPath string // Local checkout path for SSG repository
)

// DefaultRepoURL returns the default Git repository URL for SSG data.
// If overridden at build time, returns the configured value.
func DefaultRepoURL() string {
	if buildRepoURL != "" {
		return buildRepoURL
	}
	return "https://github.com/cyw0ng95/scap-security-guide-0.1.79"
}

// DefaultRepoPath returns the default local checkout path for SSG repository.
// If overridden at build time, returns the configured value.
func DefaultRepoPath() string {
	if buildRepoPath != "" {
		return buildRepoPath
	}
	return "assets/ssg-git"
}
