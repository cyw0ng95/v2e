// Package parser provides parsers for SSG data files (HTML guides, tables, JSON manifests, XML data streams).
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ssg "github.com/cyw0ng95/v2e/pkg/ssg"
)

// ManifestData represents the raw structure of a manifest JSON file.
type ManifestData struct {
	ProductName string                     `json:"product_name"`
	Rules       map[string]interface{}     `json:"rules"` // Empty in current manifests
	Profiles    map[string]ManifestProfile `json:"profiles"`
}

// ManifestProfile represents a profile within a manifest.
type ManifestProfile struct {
	Rules []string `json:"rules"`
}

// ParseManifestFile parses an SSG JSON manifest file and extracts manifest, profiles, and profile rules.
// Returns the manifest metadata, list of profiles, and list of profile-rule associations.
func ParseManifestFile(path string) (*ssg.SSGManifest, []ssg.SSGProfile, []ssg.SSGProfileRule, error) {
	// Read JSON content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var manifestData ManifestData
	if err := json.Unmarshal(data, &manifestData); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract manifest ID and product from filename
	manifestID, product := extractManifestIDFromPath(path)

	// Use product from JSON if available, otherwise from filename
	if manifestData.ProductName != "" {
		product = manifestData.ProductName
	}

	// Create manifest
	manifest := &ssg.SSGManifest{
		ID:      manifestID,
		Product: product,
	}

	// Create profiles and profile rules
	var profiles []ssg.SSGProfile
	var profileRules []ssg.SSGProfileRule

	for profileID, profileData := range manifestData.Profiles {
		// Create profile
		profile := ssg.SSGProfile{
			ID:         fmt.Sprintf("%s:%s", product, profileID),
			ManifestID: manifestID,
			Product:    product,
			ProfileID:  profileID,
			RuleCount:  len(profileData.Rules),
		}
		profiles = append(profiles, profile)

		// Create profile rules
		for _, ruleShortID := range profileData.Rules {
			profileRule := ssg.SSGProfileRule{
				ProfileID:   profile.ID,
				RuleShortID: ruleShortID,
			}
			profileRules = append(profileRules, profileRule)
		}
	}

	return manifest, profiles, profileRules, nil
}

// extractManifestIDFromPath extracts the manifest ID and product from the file path.
// Example: "manifest-al2023.json" -> ("manifest-al2023", "al2023")
func extractManifestIDFromPath(path string) (id, product string) {
	filename := filepath.Base(path)
	id = strings.TrimSuffix(filename, ".json")

	// Extract product from filename pattern: manifest-{product}.json
	if strings.HasPrefix(id, "manifest-") {
		product = strings.TrimPrefix(id, "manifest-")
	}

	return id, product
}
