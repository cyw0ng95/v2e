// Package crossref provides cross-reference extraction logic for SSG objects.
// Extracts links between guides, tables, manifests, and data streams based on
// common identifiers (Rule IDs, CCE, Products, Profiles).
package crossref

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/ssg"
)

// ruleIDPattern matches XCCDF rule IDs like "xccdf_org.ssgproject.content_rule_aide_build_database"
var ruleIDPattern = regexp.MustCompile(`xccdf_org\.ssgproject\.content_rule_([a-z0-9_]+)`)

// ccePattern matches CCE identifiers like "CCE-80644-8"
var ccePattern = regexp.MustCompile(`CCE-\d+-\d+`)

// Extractor provides methods to extract cross-references from SSG objects.
type Extractor struct{}

// NewExtractor creates a new cross-reference extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// MetadataRuleID contains metadata for rule_id link type.
type MetadataRuleID struct {
	RuleShortID string `json:"rule_short_id"` // e.g., "aide_build_database"
	FullRuleID  string `json:"full_rule_id"`  // e.g., "xccdf_org.ssgproject.content_rule_aide_build_database"
}

// MetadataCCE contains metadata for cce link type.
type MetadataCCE struct {
	CCENumber string `json:"cce_number"` // e.g., "CCE-80644-8"
}

// MetadataProduct contains metadata for product link type.
type MetadataProduct struct {
	ProductName string `json:"product_name"` // e.g., "al2023", "rhel8"
}

// MetadataProfile contains metadata for profile_id link type.
type MetadataProfile struct {
	ProfileID   string `json:"profile_id"`   // e.g., "cis", "stig"
	ProfileName string `json:"profile_name"` // e.g., "CIS Level 1"
}

// ExtractFromGuide extracts cross-references from an SSG guide.
// Finds rule IDs in HTML content and creates product-based links.
func (e *Extractor) ExtractFromGuide(guide *ssg.SSGGuide) ([]ssg.SSGCrossReference, error) {
	var refs []ssg.SSGCrossReference

	// Extract rule IDs from HTML content
	ruleIDs := ruleIDPattern.FindAllStringSubmatch(guide.HTMLContent, -1)
	seenRules := make(map[string]bool)

	for _, match := range ruleIDs {
		if len(match) < 2 {
			continue
		}
		fullRuleID := match[0]
		shortRuleID := match[1]

		// Avoid duplicates
		if seenRules[fullRuleID] {
			continue
		}
		seenRules[fullRuleID] = true

		// Create metadata
		metadata := MetadataRuleID{
			RuleShortID: shortRuleID,
			FullRuleID:  fullRuleID,
		}
		metadataJSON, _ := json.Marshal(metadata)

		// This creates a self-reference that can be matched against
		// data streams and manifests that have the same rule ID
		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "guide",
			SourceID:   guide.ID,
			TargetType: "guide", // Will be updated when matching against other objects
			TargetID:   guide.ID,
			LinkType:   "rule_id",
			Metadata:   string(metadataJSON),
		})
	}

	// Add product-based cross-reference
	if guide.Product != "" {
		metadata := MetadataProduct{
			ProductName: guide.Product,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "guide",
			SourceID:   guide.ID,
			TargetType: "product",
			TargetID:   guide.Product,
			LinkType:   "product",
			Metadata:   string(metadataJSON),
		})
	}

	// Add profile-based cross-reference
	if guide.ProfileID != "" {
		metadata := MetadataProfile{
			ProfileID:   guide.ProfileID,
			ProfileName: guide.Title,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "guide",
			SourceID:   guide.ID,
			TargetType: "profile",
			TargetID:   guide.ProfileID,
			LinkType:   "profile_id",
			Metadata:   string(metadataJSON),
		})
	}

	return refs, nil
}

// ExtractFromTable extracts cross-references from an SSG table.
// Finds CCE identifiers in table entries and creates product-based links.
func (e *Extractor) ExtractFromTable(table *ssg.SSGTable, entries []ssg.SSGTableEntry) ([]ssg.SSGCrossReference, error) {
	var refs []ssg.SSGCrossReference
	seenCCEs := make(map[string]bool)

	// Extract CCE identifiers from table entries
	for _, entry := range entries {
		if entry.Mapping == "" {
			continue
		}

		// Check if mapping is a CCE identifier
		if ccePattern.MatchString(entry.Mapping) {
			if seenCCEs[entry.Mapping] {
				continue
			}
			seenCCEs[entry.Mapping] = true

			metadata := MetadataCCE{
				CCENumber: entry.Mapping,
			}
			metadataJSON, _ := json.Marshal(metadata)

			refs = append(refs, ssg.SSGCrossReference{
				SourceType: "table",
				SourceID:   table.ID,
				TargetType: "cce",
				TargetID:   entry.Mapping,
				LinkType:   "cce",
				Metadata:   string(metadataJSON),
			})
		}
	}

	// Add product-based cross-reference
	if table.Product != "" {
		metadata := MetadataProduct{
			ProductName: table.Product,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "table",
			SourceID:   table.ID,
			TargetType: "product",
			TargetID:   table.Product,
			LinkType:   "product",
			Metadata:   string(metadataJSON),
		})
	}

	return refs, nil
}

// ExtractFromManifest extracts cross-references from an SSG manifest.
// Creates profile-based and product-based links.
func (e *Extractor) ExtractFromManifest(manifest *ssg.SSGManifest, profiles []ssg.SSGProfile) ([]ssg.SSGCrossReference, error) {
	var refs []ssg.SSGCrossReference

	// Extract profile-based cross-references
	for _, profile := range profiles {
		metadata := MetadataProfile{
			ProfileID:   profile.ProfileID,
			ProfileName: profile.Description,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "manifest",
			SourceID:   manifest.ID,
			TargetType: "profile",
			TargetID:   profile.ProfileID,
			LinkType:   "profile_id",
			Metadata:   string(metadataJSON),
		})
	}

	// Add product-based cross-reference
	if manifest.Product != "" {
		metadata := MetadataProduct{
			ProductName: manifest.Product,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "manifest",
			SourceID:   manifest.ID,
			TargetType: "product",
			TargetID:   manifest.Product,
			LinkType:   "product",
			Metadata:   string(metadataJSON),
		})
	}

	return refs, nil
}

// ExtractFromDataStream extracts cross-references from an SSG data stream.
// Finds rule IDs, CCE identifiers, profiles, and product information.
func (e *Extractor) ExtractFromDataStream(ds *ssg.SSGDataStream, benchmark *ssg.SSGBenchmark,
	profiles []ssg.SSGDSProfile, rules []ssg.SSGDSRule, identifiers []ssg.SSGDSRuleIdentifier) ([]ssg.SSGCrossReference, error) {

	var refs []ssg.SSGCrossReference
	seenRules := make(map[string]bool)
	seenCCEs := make(map[string]bool)

	// Extract rule IDs from rules
	for _, rule := range rules {
		if rule.RuleID == "" {
			continue
		}

		// Extract short rule ID from full XCCDF ID
		matches := ruleIDPattern.FindStringSubmatch(rule.RuleID)
		if len(matches) < 2 {
			continue
		}

		fullRuleID := matches[0]
		shortRuleID := matches[1]

		if seenRules[fullRuleID] {
			continue
		}
		seenRules[fullRuleID] = true

		metadata := MetadataRuleID{
			RuleShortID: shortRuleID,
			FullRuleID:  fullRuleID,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "datastream",
			SourceID:   ds.ID,
			TargetType: "rule",
			TargetID:   fullRuleID,
			LinkType:   "rule_id",
			Metadata:   string(metadataJSON),
		})
	}

	// Extract CCE identifiers from rule identifiers
	for _, identifier := range identifiers {
		if identifier.System != "https://nvd.nist.gov/cce/index.cfm" {
			continue
		}

		cceID := strings.TrimSpace(identifier.IdentifierID)
		if !ccePattern.MatchString(cceID) {
			continue
		}

		if seenCCEs[cceID] {
			continue
		}
		seenCCEs[cceID] = true

		metadata := MetadataCCE{
			CCENumber: cceID,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "datastream",
			SourceID:   ds.ID,
			TargetType: "cce",
			TargetID:   cceID,
			LinkType:   "cce",
			Metadata:   string(metadataJSON),
		})
	}

	// Extract profile-based cross-references
	for _, profile := range profiles {
		if profile.ProfileID == "" {
			continue
		}

		metadata := MetadataProfile{
			ProfileID:   profile.ProfileID,
			ProfileName: profile.Title,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "datastream",
			SourceID:   ds.ID,
			TargetType: "profile",
			TargetID:   profile.ProfileID,
			LinkType:   "profile_id",
			Metadata:   string(metadataJSON),
		})
	}

	// Add product-based cross-reference
	if ds.Product != "" {
		metadata := MetadataProduct{
			ProductName: ds.Product,
		}
		metadataJSON, _ := json.Marshal(metadata)

		refs = append(refs, ssg.SSGCrossReference{
			SourceType: "datastream",
			SourceID:   ds.ID,
			TargetType: "product",
			TargetID:   ds.Product,
			LinkType:   "product",
			Metadata:   string(metadataJSON),
		})
	}

	return refs, nil
}

// MatchCrossReferences finds matching cross-references between objects.
// This can be used to create bidirectional links after all objects are imported.
func (e *Extractor) MatchCrossReferences(refs []ssg.SSGCrossReference) ([]ssg.SSGCrossReference, error) {
	// Group references by link type and target ID
	byLinkAndTarget := make(map[string][]ssg.SSGCrossReference)

	for _, ref := range refs {
		key := fmt.Sprintf("%s:%s", ref.LinkType, ref.TargetID)
		byLinkAndTarget[key] = append(byLinkAndTarget[key], ref)
	}

	// Create bidirectional links
	var matched []ssg.SSGCrossReference

	for _, group := range byLinkAndTarget {
		// For each pair of references with the same link type and target
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				ref1 := group[i]
				ref2 := group[j]

				// Skip self-references
				if ref1.SourceType == ref2.SourceType && ref1.SourceID == ref2.SourceID {
					continue
				}

				// Create bidirectional link from ref1 to ref2
				matched = append(matched, ssg.SSGCrossReference{
					SourceType: ref1.SourceType,
					SourceID:   ref1.SourceID,
					TargetType: ref2.SourceType,
					TargetID:   ref2.SourceID,
					LinkType:   ref1.LinkType,
					Metadata:   ref1.Metadata,
				})

				// Create bidirectional link from ref2 to ref1
				matched = append(matched, ssg.SSGCrossReference{
					SourceType: ref2.SourceType,
					SourceID:   ref2.SourceID,
					TargetType: ref1.SourceType,
					TargetID:   ref1.SourceID,
					LinkType:   ref2.LinkType,
					Metadata:   ref2.Metadata,
				})
			}
		}
	}

	return matched, nil
}
