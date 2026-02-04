package parser

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseDataStreamFile tests parsing a small synthetic data stream XML.
func TestParseDataStreamFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseDataStreamFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a minimal but valid SCAP data stream XML
		xmlData := `<?xml version="1.0" encoding="utf-8"?>
	<ds:data-stream-collection xmlns:ds="http://scap.nist.gov/schema/scap/source/1.2" xmlns:xccdf-1.2="http://checklists.nist.gov/xccdf/1.2" xmlns:html="http://www.w3.org/1999/xhtml" id="scap_test_collection" schematron-version="1.3">
	  <ds:data-stream id="scap_test_datastream" scap-version="1.3" use-case="OTHER" timestamp="2025-11-28T13:53:44">
	    <ds:checklists>
	      <ds:component-ref id="scap_test_cref_xccdf" xlink:href="#scap_test_comp_xccdf" xmlns:xlink="http://www.w3.org/1999/xlink"/>
	    </ds:checklists>
	  </ds:data-stream>
	  <ds:component id="scap_test_comp_xccdf" timestamp="2025-11-28T13:53:44">
	    <xccdf-1.2:Benchmark id="xccdf_org.ssgproject.content_benchmark_TEST" xsi:schemaLocation="http://checklists.nist.gov/xccdf/1.2 xccdf-1.2.xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" style="SCAP_1.2" resolved="true" xml:lang="en-US">
	      <xccdf-1.2:status date="2025-11-28">draft</xccdf-1.2:status>
	      <xccdf-1.2:title>Test Security Guide</xccdf-1.2:title>
	      <xccdf-1.2:description>This is a test security guide for unit testing.<html:br/>It contains minimal but valid XCCDF content.</xccdf-1.2:description>
	      <xccdf-1.2:version>1.0.0</xccdf-1.2:version>
	      <xccdf-1.2:Profile id="xccdf_org.ssgproject.content_profile_test_cis">
	        <xccdf-1.2:version>1.0.0</xccdf-1.2:version>
	        <xccdf-1.2:title override="true">Test CIS Profile</xccdf-1.2:title>
	        <xccdf-1.2:description override="true">This is a test CIS profile for unit testing.<html:br/>It selects specific rules.</xccdf-1.2:description>
	        <xccdf-1.2:select idref="xccdf_org.ssgproject.content_rule_test_rule_1" selected="true"/>
	        <xccdf-1.2:select idref="xccdf_org.ssgproject.content_rule_test_rule_2" selected="true"/>
	      </xccdf-1.2:Profile>
	      <xccdf-1.2:Group id="xccdf_org.ssgproject.content_group_test_system">
	        <xccdf-1.2:title>Test System Settings</xccdf-1.2:title>
	        <xccdf-1.2:description>Test group for system settings.</xccdf-1.2:description>
	        <xccdf-1.2:Group id="xccdf_org.ssgproject.content_group_test_software">
	          <xccdf-1.2:title>Test Software</xccdf-1.2:title>
	          <xccdf-1.2:description>Test group for software settings.</xccdf-1.2:description>
	          <xccdf-1.2:Rule selected="false" id="xccdf_org.ssgproject.content_rule_test_rule_1" severity="medium" weight="1.0">
	            <xccdf-1.2:title>Test Rule 1</xccdf-1.2:title>
	            <xccdf-1.2:description>This is a test rule description.<html:br/>It has multiple lines.</xccdf-1.2:description>
	            <xccdf-1.2:rationale>This is the rationale for test rule 1.</xccdf-1.2:rationale>
	            <xccdf-1.2:version>1.0</xccdf-1.2:version>
	            <xccdf-1.2:reference href="https://www.cisecurity.org/benchmark/test/">1.1.1</xccdf-1.2:reference>
	            <xccdf-1.2:reference href="http://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-53r4.pdf">CM-6(a)</xccdf-1.2:reference>
	            <xccdf-1.2:ident system="http://cce.mitre.org">CCE-12345-6</xccdf-1.2:ident>
	          </xccdf-1.2:Rule>
	          <xccdf-1.2:Rule selected="true" id="xccdf_org.ssgproject.content_rule_test_rule_2" severity="high">
	            <xccdf-1.2:title>Test Rule 2</xccdf-1.2:title>
	            <xccdf-1.2:description>This is test rule 2 description.</xccdf-1.2:description>
	            <xccdf-1.2:rationale>This is the rationale for test rule 2.</xccdf-1.2:rationale>
	            <xccdf-1.2:version>2.0</xccdf-1.2:version>
	            <xccdf-1.2:reference href="https://www.cisecurity.org/benchmark/test/">1.1.2</xccdf-1.2:reference>
	            <xccdf-1.2:ident system="http://cce.mitre.org">CCE-12346-7</xccdf-1.2:ident>
	          </xccdf-1.2:Rule>
	        </xccdf-1.2:Group>
	      </xccdf-1.2:Group>
	    </xccdf-1.2:Benchmark>
	  </ds:component>
	</ds:data-stream-collection>`

		// Parse the XML
		reader := strings.NewReader(xmlData)
		ds, benchmark, profiles, groups, rules, err := ParseDataStreamFile(reader, "ssg-test-ds.xml")

		// Validate no errors
		if err != nil {
			t.Fatalf("ParseDataStreamFile failed: %v", err)
		}

		// Test Data Stream
		if ds == nil {
			t.Fatal("Data stream is nil")
		}
		if ds.Product != "test" {
			t.Errorf("Expected product 'test', got '%s'", ds.Product)
		}
		if ds.ScapVersion != "1.3" {
			t.Errorf("Expected SCAP version '1.3', got '%s'", ds.ScapVersion)
		}
		if ds.Timestamp != "2025-11-28T13:53:44" {
			t.Errorf("Expected timestamp '2025-11-28T13:53:44', got '%s'", ds.Timestamp)
		}

		// Test Benchmark
		if benchmark == nil {
			t.Fatal("Benchmark is nil")
		}
		if benchmark.ID != "xccdf_org.ssgproject.content_benchmark_TEST" {
			t.Errorf("Expected benchmark ID 'xccdf_org.ssgproject.content_benchmark_TEST', got '%s'", benchmark.ID)
		}
		if benchmark.Title != "Test Security Guide" {
			t.Errorf("Expected benchmark title 'Test Security Guide', got '%s'", benchmark.Title)
		}
		if !strings.Contains(benchmark.Description, "test security guide") {
			t.Errorf("Expected description to contain 'test security guide', got '%s'", benchmark.Description)
		}
		if benchmark.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", benchmark.Version)
		}
		if benchmark.Status != "draft" {
			t.Errorf("Expected status 'draft', got '%s'", benchmark.Status)
		}
		if benchmark.StatusDate != "2025-11-28" {
			t.Errorf("Expected status date '2025-11-28', got '%s'", benchmark.StatusDate)
		}
		if benchmark.ProfileCount != 1 {
			t.Errorf("Expected 1 profile, got %d", benchmark.ProfileCount)
		}
		if benchmark.GroupCount != 2 {
			t.Errorf("Expected 2 groups, got %d", benchmark.GroupCount)
		}
		if benchmark.RuleCount != 2 {
			t.Errorf("Expected 2 rules, got %d", benchmark.RuleCount)
		}

		// Test Profiles
		if len(profiles) != 1 {
			t.Fatalf("Expected 1 profile, got %d", len(profiles))
		}
		profile := profiles[0]
		if profile.ID != "xccdf_org.ssgproject.content_profile_test_cis" {
			t.Errorf("Expected profile ID 'xccdf_org.ssgproject.content_profile_test_cis', got '%s'", profile.ID)
		}
		if profile.Title != "Test CIS Profile" {
			t.Errorf("Expected profile title 'Test CIS Profile', got '%s'", profile.Title)
		}
		if profile.RuleCount != 2 {
			t.Errorf("Expected 2 selected rules, got %d", profile.RuleCount)
		}
		if len(profile.SelectedRules) != 2 {
			t.Errorf("Expected 2 selected rules in array, got %d", len(profile.SelectedRules))
		}

		// Test Profile Rules
		for i, pr := range profile.SelectedRules {
			if !pr.Selected {
				t.Errorf("Profile rule %d should be selected", i)
			}
			if pr.ProfileID != profile.ID {
				t.Errorf("Profile rule %d has wrong profile ID: %s", i, pr.ProfileID)
			}
		}

		// Test Groups
		if len(groups) != 2 {
			t.Fatalf("Expected 2 groups, got %d", len(groups))
		}

		// Test top-level group
		group1 := groups[0]
		if group1.ID != "xccdf_org.ssgproject.content_group_test_system" {
			t.Errorf("Expected group ID 'xccdf_org.ssgproject.content_group_test_system', got '%s'", group1.ID)
		}
		if group1.Title != "Test System Settings" {
			t.Errorf("Expected group title 'Test System Settings', got '%s'", group1.Title)
		}
		if group1.ParentID != "" {
			t.Errorf("Expected empty parent ID for top-level group, got '%s'", group1.ParentID)
		}
		if group1.Level != 0 {
			t.Errorf("Expected level 0 for top-level group, got %d", group1.Level)
		}

		// Test nested group
		group2 := groups[1]
		if group2.ID != "xccdf_org.ssgproject.content_group_test_software" {
			t.Errorf("Expected group ID 'xccdf_org.ssgproject.content_group_test_software', got '%s'", group2.ID)
		}
		if group2.ParentID != group1.ID {
			t.Errorf("Expected parent ID '%s', got '%s'", group1.ID, group2.ParentID)
		}
		if group2.Level != 1 {
			t.Errorf("Expected level 1 for nested group, got %d", group2.Level)
		}

		// Test Rules
		if len(rules) != 2 {
			t.Fatalf("Expected 2 rules, got %d", len(rules))
		}

		// Test Rule 1
		rule1 := rules[0]
		if rule1.ID != "xccdf_org.ssgproject.content_rule_test_rule_1" {
			t.Errorf("Expected rule ID 'xccdf_org.ssgproject.content_rule_test_rule_1', got '%s'", rule1.ID)
		}
		if rule1.Title != "Test Rule 1" {
			t.Errorf("Expected rule title 'Test Rule 1', got '%s'", rule1.Title)
		}
		if !strings.Contains(rule1.Description, "test rule description") {
			t.Errorf("Expected description to contain 'test rule description', got '%s'", rule1.Description)
		}
		if rule1.Severity != "medium" {
			t.Errorf("Expected severity 'medium', got '%s'", rule1.Severity)
		}
		if rule1.Selected {
			t.Error("Rule 1 should not be selected by default")
		}
		if rule1.Weight != "1.0" {
			t.Errorf("Expected weight '1.0', got '%s'", rule1.Weight)
		}
		if rule1.Version != "1.0" {
			t.Errorf("Expected version '1.0', got '%s'", rule1.Version)
		}
		if rule1.GroupID != group2.ID {
			t.Errorf("Expected rule to belong to group '%s', got '%s'", group2.ID, rule1.GroupID)
		}

		// Test Rule 1 References
		if len(rule1.References) != 2 {
			t.Errorf("Expected 2 references for rule 1, got %d", len(rule1.References))
		} else {
			ref1 := rule1.References[0]
			if !strings.Contains(ref1.Href, "cisecurity.org") {
				t.Errorf("Expected CIS reference, got href '%s'", ref1.Href)
			}
			if ref1.RefID != "1.1.1" {
				t.Errorf("Expected ref ID '1.1.1', got '%s'", ref1.RefID)
			}

			ref2 := rule1.References[1]
			if !strings.Contains(ref2.Href, "NIST") {
				t.Errorf("Expected NIST reference, got href '%s'", ref2.Href)
			}
			if ref2.RefID != "CM-6(a)" {
				t.Errorf("Expected ref ID 'CM-6(a)', got '%s'", ref2.RefID)
			}
		}

		// Test Rule 1 Identifiers
		if len(rule1.Identifiers) != 1 {
			t.Errorf("Expected 1 identifier for rule 1, got %d", len(rule1.Identifiers))
		} else {
			ident := rule1.Identifiers[0]
			if !strings.Contains(ident.System, "cce.mitre.org") {
				t.Errorf("Expected CCE system, got '%s'", ident.System)
			}
			if ident.Identifier != "CCE-12345-6" {
				t.Errorf("Expected identifier 'CCE-12345-6', got '%s'", ident.Identifier)
			}
		}

		// Test Rule 2
		rule2 := rules[1]
		if rule2.ID != "xccdf_org.ssgproject.content_rule_test_rule_2" {
			t.Errorf("Expected rule ID 'xccdf_org.ssgproject.content_rule_test_rule_2', got '%s'", rule2.ID)
		}
		if rule2.Severity != "high" {
			t.Errorf("Expected severity 'high', got '%s'", rule2.Severity)
		}
		if !rule2.Selected {
			t.Error("Rule 2 should be selected by default")
		}
	})

}

// TestParseDataStreamFile_RealFile tests parsing a real SSG data stream file from submodule.
// This test validates parsing against actual production data.
func TestParseDataStreamFile_RealFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseDataStreamFile_RealFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Path to real SSG data stream file in submodule
		dsPath := filepath.Join("..", "..", "..", "assets", "ssg-static", "ssg-al2023-ds.xml")

		// Check if file exists (skip test if submodule not initialized)
		if _, err := os.Stat(dsPath); os.IsNotExist(err) {
			t.Skip("Skipping test: SSG submodule not initialized. Run: git submodule update --init --recursive")
		}

		// Open file
		file, err := os.Open(dsPath)
		if err != nil {
			t.Fatalf("Failed to open data stream file: %v", err)
		}
		defer file.Close()

		// Parse the file
		ds, benchmark, profiles, groups, rules, err := ParseDataStreamFile(file, "ssg-al2023-ds.xml")
		if err != nil {
			t.Fatalf("ParseDataStreamFile failed on real file: %v", err)
		}

		// Validate Data Stream
		if ds == nil {
			t.Fatal("Data stream is nil")
		}
		if ds.Product != "al2023" {
			t.Errorf("Expected product 'al2023', got '%s'", ds.Product)
		}
		t.Logf("Data Stream: ID=%s, Product=%s, SCAP Version=%s", ds.ID, ds.Product, ds.ScapVersion)

		// Validate Benchmark
		if benchmark == nil {
			t.Fatal("Benchmark is nil")
		}
		if benchmark.Title == "" {
			t.Error("Benchmark title is empty")
		}
		if benchmark.Version == "" {
			t.Error("Benchmark version is empty")
		}
		t.Logf("Benchmark: ID=%s, Title=%s, Version=%s", benchmark.ID, benchmark.Title, benchmark.Version)
		t.Logf("Benchmark Stats: %d profiles, %d groups, %d rules", benchmark.ProfileCount, benchmark.GroupCount, benchmark.RuleCount)

		// Validate Profiles
		if len(profiles) == 0 {
			t.Error("No profiles found in real data stream")
		}
		t.Logf("Found %d profiles:", len(profiles))
		for i, profile := range profiles {
			if profile.ID == "" {
				t.Errorf("Profile %d has empty ID", i)
			}
			if profile.Title == "" {
				t.Errorf("Profile %d has empty title", i)
			}
			t.Logf("  Profile %d: ID=%s, Title=%s, RuleCount=%d", i+1, profile.ID, profile.Title, profile.RuleCount)

			// Validate profile has selected rules
			if len(profile.SelectedRules) == 0 {
				t.Errorf("Profile %d has no selected rules", i)
			}
		}

		// Validate Groups
		if len(groups) == 0 {
			t.Error("No groups found in real data stream")
		}
		t.Logf("Found %d groups", len(groups))

		// Count groups by level
		levelCounts := make(map[int]int)
		for _, group := range groups {
			levelCounts[group.Level]++
			if group.ID == "" {
				t.Error("Found group with empty ID")
			}
			if group.Title == "" {
				t.Error("Found group with empty title")
			}
		}
		for level, count := range levelCounts {
			t.Logf("  Level %d: %d groups", level, count)
		}

		// Validate Rules
		if len(rules) == 0 {
			t.Error("No rules found in real data stream")
		}
		t.Logf("Found %d rules", len(rules))

		// Analyze rule statistics
		severityCounts := make(map[string]int)
		rulesWithRefs := 0
		rulesWithIdents := 0
		totalRefs := 0
		totalIdents := 0

		for _, rule := range rules {
			if rule.ID == "" {
				t.Error("Found rule with empty ID")
			}
			if rule.Title == "" {
				t.Error("Found rule with empty title")
			}

			severityCounts[rule.Severity]++

			if len(rule.References) > 0 {
				rulesWithRefs++
				totalRefs += len(rule.References)
			}
			if len(rule.Identifiers) > 0 {
				rulesWithIdents++
				totalIdents += len(rule.Identifiers)
			}
		}

		t.Logf("Rule Statistics:")
		for severity, count := range severityCounts {
			t.Logf("  Severity '%s': %d rules", severity, count)
		}
		t.Logf("  Rules with references: %d/%d (total refs: %d)", rulesWithRefs, len(rules), totalRefs)
		t.Logf("  Rules with identifiers: %d/%d (total identifiers: %d)", rulesWithIdents, len(rules), totalIdents)

		// Validate hierarchical structure
		// Check that all groups have valid benchmark ID
		for i, group := range groups {
			if group.BenchmarkID != benchmark.ID {
				t.Errorf("Group %d has incorrect benchmark ID: expected '%s', got '%s'", i, benchmark.ID, group.BenchmarkID)
			}
		}

		// Check that all rules have valid group ID
		groupIDs := make(map[string]bool)
		for _, group := range groups {
			groupIDs[group.ID] = true
		}
		for i, rule := range rules {
			if rule.GroupID == "" {
				t.Errorf("Rule %d has empty group ID", i)
			} else if !groupIDs[rule.GroupID] {
				t.Errorf("Rule %d has invalid group ID '%s' not found in groups", i, rule.GroupID)
			}
		}

		// Sample detailed validation of first rule
		if len(rules) > 0 {
			rule := rules[0]
			t.Logf("\nSample Rule (first rule):")
			t.Logf("  ID: %s", rule.ID)
			t.Logf("  Title: %s", rule.Title)
			t.Logf("  Severity: %s", rule.Severity)
			t.Logf("  Selected: %v", rule.Selected)
			t.Logf("  Version: %s", rule.Version)
			t.Logf("  Weight: %s", rule.Weight)
			t.Logf("  Description: %s", truncate(rule.Description, 100))
			t.Logf("  Rationale: %s", truncate(rule.Rationale, 100))
			t.Logf("  References: %d", len(rule.References))
			if len(rule.References) > 0 {
				t.Logf("    First ref: href=%s, id=%s", truncate(rule.References[0].Href, 50), rule.References[0].RefID)
			}
			t.Logf("  Identifiers: %d", len(rule.Identifiers))
			if len(rule.Identifiers) > 0 {
				t.Logf("    First ident: system=%s, value=%s", truncate(rule.Identifiers[0].System, 50), rule.Identifiers[0].Identifier)
			}
		}
	})

}

// TestParseDataStreamFile_InvalidFile tests error handling for invalid input.
func TestParseDataStreamFile_InvalidFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseDataStreamFile_InvalidFile", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name    string
			data    string
			wantErr bool
		}{
			{
				name:    "Invalid XML",
				data:    "not xml at all",
				wantErr: true,
			},
			{
				name:    "Empty XML",
				data:    "<?xml version=\"1.0\"?><root></root>",
				wantErr: true,
			},
			{
				name:    "No data stream",
				data:    "<?xml version=\"1.0\"?><ds:data-stream-collection xmlns:ds=\"http://scap.nist.gov/schema/scap/source/1.2\"></ds:data-stream-collection>",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				reader := strings.NewReader(tt.data)
				_, _, _, _, _, err := ParseDataStreamFile(reader, "ssg-test-ds.xml")
				if (err != nil) != tt.wantErr {
					t.Errorf("ParseDataStreamFile() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

}

// TestExtractProduct tests product extraction from filenames.
func TestExtractProduct(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExtractProduct", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			filename string
			want     string
			wantErr  bool
		}{
			{"ssg-al2023-ds.xml", "al2023", false},
			{"ssg-rhel8-ds.xml", "rhel8", false},
			{"ssg-rhel9-ds.xml", "rhel9", false},
			{"invalid-filename.xml", "", true},
			{"ssg-al2023.xml", "", true},
			{"al2023-ds.xml", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.filename, func(t *testing.T) {
				got, err := extractProduct(tt.filename)
				if (err != nil) != tt.wantErr {
					t.Errorf("extractProduct() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("extractProduct() = %v, want %v", got, tt.want)
				}
			})
		}
	})

}

// truncate truncates a string to maxLen and adds "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
