// Package parser provides HTML parsing tests for SSG guide files.
package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/cyw0ng95/v2e/pkg/ssg"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestExtractIDFromPath(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExtractIDFromPath", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name        string
			path        string
			wantID      string
			wantProduct string
			wantShortID string
		}{
			{
				name:        "AL2023 CIS guide",
				path:        "guides/ssg-al2023-guide-cis.html",
				wantID:      "ssg-al2023-guide-cis",
				wantProduct: "al2023",
				wantShortID: "cis",
			},
			{
				name:        "AL2023 CIS Server Level 1",
				path:        "guides/ssg-al2023-guide-cis_server_l1.html",
				wantID:      "ssg-al2023-guide-cis_server_l1",
				wantProduct: "al2023",
				wantShortID: "cis_server_l1",
			},
			{
				name:        "RHEL9 STIG",
				path:        "guides/ssg-rhel9-guide-stig.html",
				wantID:      "ssg-rhel9-guide-stig",
				wantProduct: "rhel9",
				wantShortID: "stig",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, product, shortID := extractIDFromPath(tt.path)
				if id != tt.wantID {
					t.Errorf("extractIDFromPath() id = %v, want %v", id, tt.wantID)
				}
				if product != tt.wantProduct {
					t.Errorf("extractIDFromPath() product = %v, want %v", product, tt.wantProduct)
				}
				if shortID != tt.wantShortID {
					t.Errorf("extractIDFromPath() shortID = %v, want %v", shortID, tt.wantShortID)
				}
			})
		}
	})

}

func TestExtractShortID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExtractShortID", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name        string
			fullID      string
			elementType string
			want        string
		}{
			{
				name:        "group system",
				fullID:      "xccdf_org.ssgproject.content_group_system",
				elementType: "group",
				want:        "system",
			},
			{
				name:        "rule package aide",
				fullID:      "xccdf_org.ssgproject.content_rule_package_aide_installed",
				elementType: "rule",
				want:        "package_aide_installed",
			},
			{
				name:        "no match",
				fullID:      "some_other_id",
				elementType: "group",
				want:        "some_other_id",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := extractShortID(tt.fullID, tt.elementType); got != tt.want {
					t.Errorf("extractShortID() = %v, want %v", got, tt.want)
				}
			})
		}
	})

}

func TestNormalizeParentID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNormalizeParentID", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name     string
			parentID string
			want     string
		}{
			{
				name:     "with children prefix",
				parentID: "children-xccdf_org.ssgproject.content_group_system",
				want:     "xccdf_org.ssgproject.content_group_system",
			},
			{
				name:     "without prefix",
				parentID: "xccdf_org.ssgproject.content_group_system",
				want:     "xccdf_org.ssgproject.content_group_system",
			},
			{
				name:     "empty",
				parentID: "",
				want:     "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := normalizeParentID(tt.parentID); got != tt.want {
					t.Errorf("normalizeParentID() = %v, want %v", got, tt.want)
				}
			})
		}
	})

}

func TestParseReferences(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseReferences", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name     string
			html     string
			ruleID   string
			wantLen  int
			wantFirst ssg.SSGReference
		}{
			{
				name:   "single reference",
				html:   `<table class="identifiers"><tr><td><a href="https://example.com/nist">nist</a></td><td>CM-6(a)</td></tr></table>`,
				ruleID: "test-rule",
				wantLen: 1,
				wantFirst: ssg.SSGReference{
					RuleID: "test-rule",
					Href:   "https://example.com/nist",
					Label:  "nist",
					Value:  "CM-6(a)",
				},
			},
			{
				name:   "multiple references",
				html:   `<table class="identifiers"><tr><td><a href="https://example.com/nist">nist</a></td><td>CM-6(a)</td></tr><tr><td><a href="https://example.com/cis">cis</a></td><td>1.3.1</td></tr></table>`,
				ruleID: "test-rule",
				wantLen: 2,
				wantFirst: ssg.SSGReference{
					RuleID: "test-rule",
					Href:   "https://example.com/nist",
					Label:  "nist",
					Value:  "CM-6(a)",
				},
			},
			{
				name:    "no references",
				html:    `<table class="identifiers"></table>`,
				ruleID:  "test-rule",
				wantLen: 0,
			},
			{
				name:    "empty row",
				html:    `<table class="identifiers"><tr><td></td><td></td></tr></table>`,
				ruleID:  "test-rule",
				wantLen: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
				if err != nil {
					t.Fatalf("Failed to parse HTML: %v", err)
				}

				refs := parseReferences(doc.Selection, tt.ruleID)
				if len(refs) != tt.wantLen {
					t.Errorf("parseReferences() got %d references, want %d", len(refs), tt.wantLen)
				}

				if tt.wantLen > 0 {
					if refs[0] != tt.wantFirst {
						t.Errorf("parseReferences() first = %+v, want %+v", refs[0], tt.wantFirst)
					}
				}
			})
		}
	})

}

func TestParseGuideFile(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestParseGuideFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Find the actual SSG static assets
		guidePath := "../../../assets/ssg-static/guides/ssg-al2023-guide-cis.html"
		if _, err := os.Stat(guidePath); os.IsNotExist(err) {
			t.Skip("SSG static assets not found, skipping integration test")
		}

		guide, groups, rules, err := ParseGuideFile(guidePath)
		if err != nil {
			t.Fatalf("ParseGuideFile() error = %v", err)
		}

		// Check guide
		if guide.ID == "" {
			t.Error("guide.ID should not be empty")
		}
		if guide.Product != "al2023" {
			t.Errorf("guide.Product = %v, want al2023", guide.Product)
		}
		if guide.ShortID != "cis" {
			t.Errorf("guide.ShortID = %v, want cis", guide.ShortID)
		}
		if guide.Title == "" {
			t.Error("guide.Title should not be empty")
		}
		if len(guide.HTMLContent) == 0 {
			t.Error("guide.HTMLContent should not be empty")
		}

		// Check that we got some groups and rules
		if len(groups) == 0 {
			t.Error("expected at least one group")
		}
		if len(rules) == 0 {
			t.Error("expected at least one rule")
		}

		// Verify group structure
		groupFound := false
		for _, g := range groups {
			if g.ID != "" && g.GuideID == guide.ID {
				groupFound = true
				if g.Title == "" {
					t.Errorf("group %s should have a title", g.ID)
				}
				break
			}
		}
		if !groupFound {
			t.Error("expected at least one valid group")
		}

		// Verify rule structure
		ruleFound := false
		for _, r := range rules {
			if r.ID != "" && r.GuideID == guide.ID {
				ruleFound = true
				if r.Title == "" {
					t.Errorf("rule %s should have a title", r.ID)
				}
				if r.GroupID == "" {
					t.Errorf("rule %s should have a GroupID", r.ID)
				}
				if r.Severity == "" {
					t.Errorf("rule %s should have a severity", r.ID)
				}
				break
			}
		}
		if !ruleFound {
			t.Error("expected at least one valid rule")
		}

		// Verify tree structure (parent-child relationships)
		// Find a rule and check its parent group exists
		for _, r := range rules {
			if r.GroupID != "" {
				groupFound := false
				for _, g := range groups {
					if g.ID == r.GroupID {
						groupFound = true
						break
					}
				}
				if !groupFound {
					t.Errorf("rule %s has GroupID %s but no matching group found", r.ID, r.GroupID)
				}
				break
			}
		}
	})

}

func TestParseIndexGuide(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseIndexGuide", nil, func(t *testing.T, tx *gorm.DB) {
		// Test parsing the index guide (smaller file)
		guidePath := "../../../assets/ssg-static/guides/ssg-al2023-guide-index.html"
		if _, err := os.Stat(guidePath); os.IsNotExist(err) {
			t.Skip("SSG index guide not found, skipping test")
		}

		guide, groups, rules, err := ParseGuideFile(guidePath)
		if err != nil {
			t.Fatalf("ParseGuideFile() error = %v", err)
		}

		// Index guides may have fewer groups/rules
		if guide.ID == "" {
			t.Error("guide.ID should not be empty")
		}

		t.Logf("Parsed index guide: ID=%s, Groups=%d, Rules=%d", guide.ID, len(groups), len(rules))
	})

}

func TestParseMultipleGuides(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestParseMultipleGuides", nil, func(t *testing.T, tx *gorm.DB) {
		// Test parsing multiple guide files to ensure robustness
		guidesDir := "../../../assets/ssg-static/guides"
		files, err := os.ReadDir(guidesDir)
		if err != nil {
			t.Skip("Cannot read guides directory, skipping test")
		}

		// Test up to 3 guide files
		count := 0
		for _, file := range files {
			if count >= 3 {
				break
			}
			if filepath.Ext(file.Name()) != ".html" {
				continue
			}
			if !strings.HasPrefix(file.Name(), "ssg-") {
				continue
			}

			guidePath := filepath.Join(guidesDir, file.Name())
			guide, groups, rules, err := ParseGuideFile(guidePath)
			if err != nil {
				t.Logf("Warning: Failed to parse %s: %v", file.Name(), err)
				continue
			}

			if guide.ID == "" {
				t.Errorf("%s: guide.ID should not be empty", file.Name())
			}
			if guide.Product == "" {
				t.Errorf("%s: guide.Product should not be empty", file.Name())
			}

			t.Logf("Parsed %s: ID=%s, Groups=%d, Rules=%d", file.Name(), guide.ID, len(groups), len(rules))
			count++
		}

		if count == 0 {
			t.Skip("No guide files found to test")
		}
	})

}

func TestParseTreeNodeStructure(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestParseTreeNodeStructure", nil, func(t *testing.T, tx *gorm.DB) {
		guidePath := "../../../assets/ssg-static/guides/ssg-al2023-guide-cis.html"
		if _, err := os.Stat(guidePath); os.IsNotExist(err) {
			t.Skip("SSG static assets not found, skipping test")
		}

		_, groups, _, err := ParseGuideFile(guidePath)
		if err != nil {
			t.Fatalf("ParseGuideFile() error = %v", err)
		}

		// Build a map for quick lookup
		groupMap := make(map[string]ssg.SSGGroup)
		for _, g := range groups {
			groupMap[g.ID] = g
		}

		// Check tree consistency
		rootCount := 0
		maxDepth := 0

		for _, g := range groups {
			// A group is a root if ParentID is empty OR points to a benchmark (not a group)
			isRoot := g.ParentID == "" || strings.Contains(g.ParentID, "content_benchmark")
			if isRoot {
				rootCount++
			} else {
				// Verify parent exists (unless parent is a benchmark)
				if !strings.Contains(g.ParentID, "content_benchmark") {
					if _, exists := groupMap[g.ParentID]; !exists {
						t.Errorf("Group %s has ParentID %s but parent not found", g.ID, g.ParentID)
					}
				}
			}
			if g.Level > maxDepth {
				maxDepth = g.Level
			}
		}

		t.Logf("Tree structure: %d root groups, max depth %d", rootCount, maxDepth)

		if rootCount == 0 {
			t.Error("Expected at least one root group")
		}
		if maxDepth == 0 && len(groups) > 1 {
			t.Error("Expected multi-level tree but all groups are at level 0")
		}
	})

}

// ============================================================================
// Benchmark Tests
// ============================================================================

// BenchmarkParseGuideFile_RealFile benchmarks parsing a real SSG HTML guide file.
// This measures the performance of HTML parsing, tree structure extraction,
// and node traversal. The current implementation has O(N^2) complexity due to
// repeated full-document searches in parseGroupFromNode and parseRuleFromNode.
func BenchmarkParseGuideFile_RealFile(b *testing.B) {
	// Path to real SSG guide file in submodule
	guidePath := filepath.Join("..", "..", "..", "assets", "ssg-static", "guides", "ssg-al2023-guide-cis.html")

	// Check if file exists (skip benchmark if submodule not initialized)
	if _, err := os.Stat(guidePath); os.IsNotExist(err) {
		b.Skip("Skipping benchmark: SSG submodule not initialized. Run: git submodule update --init --recursive")
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _, err := ParseGuideFile(guidePath)
		if err != nil {
			b.Fatalf("ParseGuideFile failed: %v", err)
		}
	}
}
