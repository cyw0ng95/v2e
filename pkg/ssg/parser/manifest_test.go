package parser

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestParseManifestFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseManifestFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a test manifest file
		testJSON := `{
	  "product_name": "test-product",
	  "rules": {},
	  "profiles": {
	    "cis": {
	      "rules": [
	        "aide_build_database",
	        "account_disable_post_pw_expiration",
	        "accounts_password_pam_minlen"
	      ]
	    },
	    "stig": {
	      "rules": [
	        "aide_build_database",
	        "auditd_data_retention_max_log_file"
	      ]
	    }
	  }
	}`

		// Write to temp file
		tmpFile, err := os.CreateTemp("", "manifest-test-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte(testJSON)); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		tmpFile.Close()

		// Parse the file
		manifest, profiles, profileRules, err := ParseManifestFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("ParseManifestFile failed: %v", err)
		}

		// Verify manifest
		if manifest == nil {
			t.Fatal("manifest is nil")
		}
		if manifest.Product != "test-product" {
			t.Errorf("manifest.Product = %v, want test-product", manifest.Product)
		}

		// Verify profiles
		if len(profiles) != 2 {
			t.Fatalf("expected 2 profiles, got %d", len(profiles))
		}

		// Check CIS profile
		var cisProfile *struct {
			ID        string
			ProfileID string
			RuleCount int
		}
		for i := range profiles {
			if profiles[i].ProfileID == "cis" {
				cisProfile = &struct {
					ID        string
					ProfileID string
					RuleCount int
				}{
					ID:        profiles[i].ID,
					ProfileID: profiles[i].ProfileID,
					RuleCount: profiles[i].RuleCount,
				}
				break
			}
		}
		if cisProfile == nil {
			t.Fatal("CIS profile not found")
		}
		if cisProfile.RuleCount != 3 {
			t.Errorf("CIS profile RuleCount = %d, want 3", cisProfile.RuleCount)
		}

		// Verify profile rules
		if len(profileRules) != 5 { // 3 from cis + 2 from stig
			t.Fatalf("expected 5 profile rules, got %d", len(profileRules))
		}

		// Count rules for CIS profile
		cisRuleCount := 0
		for _, pr := range profileRules {
			if pr.ProfileID == cisProfile.ID {
				cisRuleCount++
			}
		}
		if cisRuleCount != 3 {
			t.Errorf("expected 3 rules for CIS profile, got %d", cisRuleCount)
		}
	})
}

func TestParseManifestFile_RealFile(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestParseManifestFile_RealFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with a real manifest file from submodule if available
		manifestPath := filepath.Join("..", "..", "..", "assets", "ssg-static", "manifests", "manifest-al2023.json")

		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			t.Skip("Skipping test: manifest file not found (submodule not initialized)")
		}

		manifest, profiles, profileRules, err := ParseManifestFile(manifestPath)
		if err != nil {
			t.Fatalf("ParseManifestFile failed: %v", err)
		}

		// Verify basic structure
		if manifest == nil {
			t.Fatal("manifest is nil")
		}
		if manifest.ID != "manifest-al2023" {
			t.Errorf("manifest.ID = %v, want manifest-al2023", manifest.ID)
		}
		if manifest.Product != "al2023" {
			t.Errorf("manifest.Product = %v, want al2023", manifest.Product)
		}

		// Should have at least some profiles
		if len(profiles) == 0 {
			t.Error("expected at least one profile")
		}

		// Should have profile rules
		if len(profileRules) == 0 {
			t.Error("expected at least one profile rule")
		}

		t.Logf("Parsed manifest: %d profiles, %d total profile rules", len(profiles), len(profileRules))
	})

}

func TestParseManifestFile_InvalidFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseManifestFile_InvalidFile", nil, func(t *testing.T, tx *gorm.DB) {
		_, _, _, err := ParseManifestFile("/nonexistent/path/manifest.json")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

}

func TestParseManifestFile_InvalidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseManifestFile_InvalidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a file with invalid JSON
		tmpFile, err := os.CreateTemp("", "manifest-invalid-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write([]byte("not valid json")); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		tmpFile.Close()

		_, _, _, err = ParseManifestFile(tmpFile.Name())
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

}

func TestExtractManifestIDFromPath(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExtractManifestIDFromPath", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			path            string
			expectedID      string
			expectedProduct string
		}{
			{
				path:            "manifests/manifest-al2023.json",
				expectedID:      "manifest-al2023",
				expectedProduct: "al2023",
			},
			{
				path:            "/tmp/manifest-rhel8.json",
				expectedID:      "manifest-rhel8",
				expectedProduct: "rhel8",
			},
			{
				path:            "manifest-ubuntu2204.json",
				expectedID:      "manifest-ubuntu2204",
				expectedProduct: "ubuntu2204",
			},
		}

		for _, tt := range tests {
			id, product := extractManifestIDFromPath(tt.path)
			if id != tt.expectedID {
				t.Errorf("extractManifestIDFromPath(%q) id = %v, want %v", tt.path, id, tt.expectedID)
			}
			if product != tt.expectedProduct {
				t.Errorf("extractManifestIDFromPath(%q) product = %v, want %v", tt.path, product, tt.expectedProduct)
			}
		}
	})

}
