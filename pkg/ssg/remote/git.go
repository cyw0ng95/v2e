// Package remote provides Git operations for SSG data fetching.
package remote

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

// GitClient handles Git repository operations for SSG data.
type GitClient struct {
	repoURL  string
	repoPath string
}

// NewGitClient creates a new GitClient with the given repository URL and path.
func NewGitClient(repoURL, repoPath string) *GitClient {
	if repoURL == "" {
		repoURL = DefaultRepoURL()
	}
	if repoPath == "" {
		repoPath = DefaultRepoPath()
	}
	return &GitClient{
		repoURL:  repoURL,
		repoPath: repoPath,
	}
}

// Clone clones the SSG repository to the local path.
// If the repository already exists, it returns an error.
func (c *GitClient) Clone() error {
	// Check if repository already exists
	if _, err := os.Stat(c.repoPath); err == nil {
		return fmt.Errorf("repository already exists at %s", c.repoPath)
	}

	_, err := git.PlainClone(c.repoPath, false, &git.CloneOptions{
		URL:      c.repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// Pull fetches the latest changes from the remote repository.
// If the repository doesn't exist locally, it will be cloned first.
func (c *GitClient) Pull() error {
	repo, err := git.PlainOpen(c.repoPath)
	if err != nil {
		// Repository doesn't exist, clone it
		if err == git.ErrRepositoryNotExists {
			return c.Clone()
		}
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Pull the latest changes
	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil // Already up to date
		}
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

// Status returns the current status of the repository.
func (c *GitClient) Status() (*RepoStatus, error) {
	repo, err := git.PlainOpen(c.repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get commit hash
	commitHash := head.Hash().String()

	// Get working tree status
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Check if repository is clean
	isClean := status.IsClean()

	return &RepoStatus{
		CommitHash: commitHash[:7], // Short hash
		Branch:     head.Name().Short(),
		IsClean:    isClean,
	}, nil
}

// ListGuideFiles returns a list of all guide HTML files in the repository.
func (c *GitClient) ListGuideFiles() ([]string, error) {
	guidesDir := filepath.Join(c.repoPath, "guides")

	entries, err := os.ReadDir(guidesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read guides directory: %w", err)
	}

	var guideFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match *-guide-*.html pattern
		if matchGuideFilePattern(name) {
			guideFiles = append(guideFiles, name)
		}
	}

	return guideFiles, nil
}

// ListTableFiles returns a list of all table HTML files in the repository.
func (c *GitClient) ListTableFiles() ([]string, error) {
	tablesDir := filepath.Join(c.repoPath, "tables")

	entries, err := os.ReadDir(tablesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tables directory: %w", err)
	}

	var tableFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match table-*.html pattern
		if matchTableFilePattern(name) {
			tableFiles = append(tableFiles, name)
		}
	}

	return tableFiles, nil
}

// ListManifestFiles returns a list of all manifest JSON files in the repository.
func (c *GitClient) ListManifestFiles() ([]string, error) {
	manifestsDir := filepath.Join(c.repoPath, "manifests")

	entries, err := os.ReadDir(manifestsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifests directory: %w", err)
	}

	var manifestFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match manifest-*.json pattern
		if matchManifestFilePattern(name) {
			manifestFiles = append(manifestFiles, name)
		}
	}

	return manifestFiles, nil
}

// GetFilePath returns the absolute path to a file in the repository.
// Guide files are located in the "guides" subdirectory.
// Table files are located in the "tables" subdirectory.
// Manifest files are located in the "manifests" subdirectory.
func (c *GitClient) GetFilePath(filename string) string {
	// Determine subdirectory based on filename pattern
	if matchTableFilePattern(filename) {
		return filepath.Join(c.repoPath, "tables", filename)
	}
	if matchManifestFilePattern(filename) {
		return filepath.Join(c.repoPath, "manifests", filename)
	}
	return filepath.Join(c.repoPath, "guides", filename)
}

// matchGuideFilePattern checks if filename matches *-guide-*.html pattern.
func matchGuideFilePattern(filename string) bool {
	// Check for .html extension
	if len(filename) < 5 {
		return false
	}
	if filepath.Ext(filename) != ".html" {
		return false
	}

	// Check for "-guide-" in the name
	base := filename[:len(filename)-5] // Remove .html
	return contains(base, "-guide-")
}

// matchTableFilePattern checks if filename matches table-*.html pattern.
func matchTableFilePattern(filename string) bool {
	// Check for .html extension
	if len(filename) < 5 {
		return false
	}
	if filepath.Ext(filename) != ".html" {
		return false
	}

	// Check for "table-" prefix
	return len(filename) > 6 && filename[:6] == "table-"
}

// matchManifestFilePattern checks if filename matches manifest-*.json pattern.
func matchManifestFilePattern(filename string) bool {
	// Check for .json extension
	if filepath.Ext(filename) != ".json" {
		return false
	}

	// Check for "manifest-" prefix
	return len(filename) > 9 && filename[:9] == "manifest-"
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

// containsMiddle checks if substr is in the middle of s.
func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RepoStatus represents the status of a Git repository.
type RepoStatus struct {
	CommitHash string // Short commit hash
	Branch     string // Branch name
	IsClean    bool   // true if working tree is clean
}
