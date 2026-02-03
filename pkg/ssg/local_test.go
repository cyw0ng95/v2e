package ssg

import (
	"testing"
)

func TestConstants(t *testing.T) {
	// Test that constants are defined
	if DefaultSSGVersion == "" {
		t.Error("DefaultSSGVersion should not be empty")
	}

	if SSGReleaseURLTemplate == "" {
		t.Error("SSGReleaseURLTemplate should not be empty")
	}

	if SSGSHA512URLTemplate == "" {
		t.Error("SSGSHA512URLTemplate should not be empty")
	}

	if SSGXMLPattern == "" {
		t.Error("SSGXMLPattern should not be empty")
	}

	if XCCDFNamespace == "" {
		t.Error("XCCDFNamespace should not be empty")
	}
}

func TestCountRules(t *testing.T) {
	// Create a test benchmark
	benchmark := &SSGBenchmark{
		Rules: []SSGRule{
			{ID: "rule1"},
			{ID: "rule2"},
		},
		Groups: []SSGGroup{
			{
				Rules: []SSGRule{
					{ID: "rule3"},
				},
				Groups: []SSGGroup{
					{
						Rules: []SSGRule{
							{ID: "rule4"},
							{ID: "rule5"},
						},
					},
				},
			},
		},
	}

	count := countRules(benchmark)
	expected := 5 // 2 + 1 + 2

	if count != expected {
		t.Errorf("countRules() = %d, want %d", count, expected)
	}
}

func TestMatchesFilters(t *testing.T) {
	rule := SSGRule{
		ID:       "test_rule",
		Severity: "high",
	}

	tests := []struct {
		name    string
		filters map[string]string
		want    bool
	}{
		{
			name:    "no filters",
			filters: map[string]string{},
			want:    true,
		},
		{
			name:    "matching severity",
			filters: map[string]string{"severity": "high"},
			want:    true,
		},
		{
			name:    "non-matching severity",
			filters: map[string]string{"severity": "low"},
			want:    false,
		},
		{
			name:    "empty severity filter",
			filters: map[string]string{"severity": ""},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesFilters(rule, tt.filters); got != tt.want {
				t.Errorf("matchesFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindRuleInGroup(t *testing.T) {
	group := &SSGGroup{
		Rules: []SSGRule{
			{ID: "rule1"},
			{ID: "rule2"},
		},
		Groups: []SSGGroup{
			{
				Rules: []SSGRule{
					{ID: "rule3"},
				},
			},
		},
	}

	tests := []struct {
		name   string
		ruleID string
		found  bool
	}{
		{
			name:   "find top-level rule",
			ruleID: "rule1",
			found:  true,
		},
		{
			name:   "find nested rule",
			ruleID: "rule3",
			found:  true,
		},
		{
			name:   "rule not found",
			ruleID: "rule999",
			found:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := findRuleInGroup(group, tt.ruleID)
			if (rule != nil) != tt.found {
				t.Errorf("findRuleInGroup() found = %v, want %v", rule != nil, tt.found)
			}
			if tt.found && rule != nil && rule.ID != tt.ruleID {
				t.Errorf("findRuleInGroup() returned rule with ID %s, want %s", rule.ID, tt.ruleID)
			}
		})
	}
}
