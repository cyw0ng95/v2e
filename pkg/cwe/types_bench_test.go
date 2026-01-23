package cwe

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
)

// Helper function to create a test CWE item
func createTestCWEItem(id string) *CWEItem {
	return &CWEItem{
		ID:          id,
		Name:        "Test CWE Item",
		Abstraction: "Class",
		Structure:   "Simple",
		Status:      "Draft",
		Description: "This is a test description for the CWE item",
		ExtendedDescription: "This is an extended description with more detailed information about the weakness and how it can affect systems.",
		LikelihoodOfExploit: "High",
		RelatedWeaknesses: []RelatedWeakness{
			{Nature: "ChildOf", CweID: "CWE-123", ViewID: "1000"},
			{Nature: "PeerOf", CweID: "CWE-456", ViewID: "1000"},
		},
		WeaknessOrdinalities: []WeaknessOrdinality{
			{Ordinality: "Primary", Description: "Primary weakness in this category"},
		},
		ApplicablePlatforms: []ApplicablePlatform{
			{Type: "Language", Name: "Java", Class: "Web Based", Prevalence: "Often"},
			{Type: "Technology", Name: "Web Applications", Class: "Application", Prevalence: "Often"},
		},
		BackgroundDetails: []string{
			"This weakness occurs in various contexts.",
			"It can lead to serious security issues.",
		},
		AlternateTerms: []AlternateTerm{
			{Term: "Alternative Term", Description: "Alternative description for this weakness"},
		},
		ModesOfIntroduction: []ModeOfIntroduction{
			{Phase: "Architecture and Design", Note: "Occurs during design phase"},
		},
		CommonConsequences: []Consequence{
			{
				Scope:      []string{"Access Control"},
				Impact:     []string{"Gain Privileges / Assume Identity"},
				Likelihood: []string{"High"},
				Note:       "Can lead to privilege escalation",
			},
		},
		DetectionMethods: []DetectionMethod{
			{
				Method:      "Automated Static Analysis",
				Description: "Automated static analysis tools can detect this weakness.",
			},
		},
		PotentialMitigations: []Mitigation{
			{
				Strategy:    "Phases: Implementation",
				Description: "Implement proper input validation to prevent this weakness.",
			},
		},
		DemonstrativeExamples: []DemonstrativeExample{
			{
				ID: "Example-1",
				Entries: []DemonstrativeEntry{
					{
						IntroText:   "Example demonstrating the weakness",
						BodyText:    "Detailed example of how the weakness manifests",
						Nature:      "Sample",
						Language:    "Java",
						ExampleCode: "// Example vulnerable code here",
					},
				},
			},
		},
		ObservedExamples: []ObservedExample{
			{
				Reference:   "Example-Ref-1",
				Description: "Real-world example of this weakness",
				Link:        "https://example.com",
			},
		},
		FunctionalAreas:   []string{"Authentication", "Authorization"},
		AffectedResources: []string{"Memory", "File"},
		TaxonomyMappings: []TaxonomyMapping{
			{
				TaxonomyName: "OWASP Top Ten",
				EntryID:      "A1",
				EntryName:    "Injection",
				MappingFit:   "Exact",
			},
		},
		RelatedAttackPatterns: []string{"CAPEC-1", "CAPEC-2"},
		References: []Reference{
			{
				ExternalReferenceID: "REF-001",
				Title:               "Security Research Paper",
				Authors:             []string{"Author 1", "Author 2"},
			},
		},
		MappingNotes: &MappingNotes{
			Usage:     "General",
			Rationale: "This mapping makes sense based on the similarity of concepts",
			Comments:  "Additional comments about the mapping",
			Reasons:   []string{"Similar concept", "Related technology"},
		},
		Notes: []Note{
			{Type: "Note Type", Note: "Additional note information"},
		},
		ContentHistory: []ContentHistory{
			{
				Type:                   "Submission",
				SubmissionName:         "Test Submission",
				SubmissionOrganization: "Test Org",
				SubmissionDate:         "2024-01-01",
			},
		},
	}
}

// BenchmarkCWEItemJSONMarshal benchmarks marshaling CWE items to JSON
func BenchmarkCWEItemJSONMarshal(b *testing.B) {
	item := createTestCWEItem("CWE-79")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(item)
		if err != nil {
			b.Fatalf("Failed to marshal: %v", err)
		}
	}
}

// BenchmarkCWEItemJSONUnmarshal benchmarks unmarshaling CWE items from JSON
func BenchmarkCWEItemJSONUnmarshal(b *testing.B) {
	item := createTestCWEItem("CWE-79")
	data, _ := json.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CWEItem
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal: %v", err)
		}
	}
}

// BenchmarkCWEItemSonicMarshal benchmarks marshaling CWE items using Sonic
func BenchmarkCWEItemSonicMarshal(b *testing.B) {
	item := createTestCWEItem("CWE-79")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(item)
		if err != nil {
			b.Fatalf("Failed to marshal with sonic: %v", err)
		}
	}
}

// BenchmarkCWEItemSonicUnmarshal benchmarks unmarshaling CWE items using Sonic
func BenchmarkCWEItemSonicUnmarshal(b *testing.B) {
	item := createTestCWEItem("CWE-79")
	data, _ := sonic.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CWEItem
		err := sonic.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal with sonic: %v", err)
		}
	}
}

// BenchmarkCWEItemSliceMarshal benchmarks marshaling slices of CWE items
func BenchmarkCWEItemSliceMarshal(b *testing.B) {
	items := make([]CWEItem, 50)
	for i := range items {
		items[i] = *createTestCWEItem("CWE-" + string(rune('0'+i)))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(items)
		if err != nil {
			b.Fatalf("Failed to marshal slice: %v", err)
		}
	}
}

// BenchmarkCWEItemSliceUnmarshal benchmarks unmarshaling slices of CWE items
func BenchmarkCWEItemSliceUnmarshal(b *testing.B) {
	items := make([]CWEItem, 50)
	for i := range items {
		items[i] = *createTestCWEItem("CWE-" + string(rune('0'+i)))
	}
	data, _ := json.Marshal(items)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []CWEItem
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal slice: %v", err)
		}
	}
}

// BenchmarkCWEItemSonicGet benchmarks getting values from CWE items using Sonic
func BenchmarkCWEItemSonicGet(b *testing.B) {
	item := createTestCWEItem("CWE-79")
	data, _ := sonic.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Parse using Sonic's get functionality
		node, err := sonic.Get(data, "id")
		if err != nil {
			b.Fatalf("Failed to get id with sonic: %v", err)
		}
		_, _ = node.String()
	}
}