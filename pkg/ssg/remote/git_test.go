// Package remote provides unit tests for Git operations.
package remote

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGitClient(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		repoPath string
	}{
		{
			name:     "default values",
			repoURL:  "",
			repoPath: "",
		},
		{
			name:     "custom values",
			repoURL:  "https://github.com/custom/repo",
			repoPath: "/custom/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGitClient(tt.repoURL, tt.repoPath)
			if client == nil {
				t.Fatal("NewGitClient returned nil")
			}

			// Verify defaults are set
			if tt.repoURL == "" && client.repoURL == "" {
				t.Error("expected default repoURL to be set")
			}
			if tt.repoPath == "" && client.repoPath == "" {
				t.Error("expected default repoPath to be set")
			}
		})
	}
}

func TestDefaultRepoURL(t *testing.T) {
	url := DefaultRepoURL()
	if url == "" {
		t.Error("DefaultRepoURL returned empty string")
	}
	if url != "https://github.com/cyw0ng95/scap-security-guide-0.1.79" {
		t.Errorf("DefaultRepoURL = %s, want default", url)
	}
}

func TestDefaultRepoPath(t *testing.T) {
	path := DefaultRepoPath()
	if path == "" {
		t.Error("DefaultRepoPath returned empty string")
	}
	if path != "assets/ssg-git" {
		t.Errorf("DefaultRepoPath = %s, want default", path)
	}
}

func TestMatchGuideFilePattern(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "valid guide file",
			filename: "ssg-al2023-guide-cis.html",
			want:     true,
		},
		{
			name:     "valid guide file with underscores",
			filename: "ssg-alinux3-guide-pci-dss.html",
			want:     true,
		},
		{
			name:     "index guide",
			filename: "ssg-al2023-guide-index.html",
			want:     true,
		},
		{
			name:     "non-guide HTML file",
			filename: "other.html",
			want:     false,
		},
		{
			name:     "non-HTML file",
			filename: "ssg-al2023-ds.xml",
			want:     false,
		},
		{
			name:     "empty filename",
			filename: "",
			want:     false,
		},
		{
			name:     "guide without html extension",
			filename: "ssg-al2023-guide-cis",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGuideFilePattern(tt.filename)
			if got != tt.want {
				t.Errorf("matchGuideFilePattern(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestGitClient_GetFilePath(t *testing.T) {
	client := NewGitClient("", "/test/repo")
	want := filepath.Join("/test/repo", "guides", "test.html")
	got := client.GetFilePath("guides/test.html")
	if got != want {
		t.Errorf("GetFilePath() = %s, want %s", got, want)
	}
}

// TestGitClient_Clone_Error tests Clone with various error conditions.
func TestGitClient_Clone_Error(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	client := NewGitClient("", filepath.Join(tempDir, "test-repo"))

	// Test cloning to an invalid URL
	client.repoURL = "invalid-url-that-does-not-exist"
	err := client.Clone()
	if err == nil {
		t.Error("expected error when cloning invalid URL, got nil")
	}
}

// TestRepoStatus validation.
func TestRepoStatus(t *testing.T) {
	status := &RepoStatus{
		CommitHash: "abc1234",
		Branch:     "main",
		IsClean:    true,
	}

	if status.CommitHash != "abc1234" {
		t.Errorf("CommitHash = %s, want abc1234", status.CommitHash)
	}
	if status.Branch != "main" {
		t.Errorf("Branch = %s, want main", status.Branch)
	}
	if !status.IsClean {
		t.Error("IsClean = false, want true")
	}
}

// TestListGuideFiles tests listing guide files in a directory structure.
func TestListGuideFiles_EmptyDir(t *testing.T) {
	// Create empty temp directory
	tempDir := t.TempDir()
	client := NewGitClient("", tempDir)

	// Create guides directory but leave it empty
	guidesDir := filepath.Join(tempDir, "guides")
	if err := os.Mkdir(guidesDir, 0755); err != nil {
		t.Fatalf("failed to create guides directory: %v", err)
	}

	files, err := client.ListGuideFiles()
	if err != nil {
		t.Errorf("ListGuideFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("ListGuideFiles() = %v, want empty list", files)
	}
}

// TestListGuideFiles_NoGuidesDir tests when guides directory doesn't exist.
func TestListGuideFiles_NoGuidesDir(t *testing.T) {
	// Create empty temp directory (no guides subdirectory)
	tempDir := t.TempDir()
	client := NewGitClient("", tempDir)

	_, err := client.ListGuideFiles()
	if err == nil {
		t.Error("expected error when guides directory doesn't exist, got nil")
	}
}

// TestListGuideFiles_WithMixedFiles tests listing with mixed file types.
func TestListGuideFiles_WithMixedFiles(t *testing.T) {
	// Create temp directory with guides subdirectory
	tempDir := t.TempDir()
	guidesDir := filepath.Join(tempDir, "guides")
	if err := os.Mkdir(guidesDir, 0755); err != nil {
		t.Fatalf("failed to create guides directory: %v", err)
	}

	// Create test files
	testFiles := []string{
		"ssg-al2023-guide-cis.html",
		"ssg-al2023-guide-index.html",
		"ssg-al2023-ds.xml",          // Not a guide
		"other.html",                  // Not a guide pattern
	}

	for _, name := range testFiles {
		path := filepath.Join(guidesDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}

	client := NewGitClient("", tempDir)
	files, err := client.ListGuideFiles()
	if err != nil {
		t.Errorf("ListGuideFiles() error = %v", err)
	}

	// Should only return guide files
	expectedCount := 2 // Only the two guide files
	if len(files) != expectedCount {
		t.Errorf("ListGuideFiles() count = %d, want %d", len(files), expectedCount)
	}
}
