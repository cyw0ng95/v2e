package cwe

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestCWEItem_JSONMarshalUnmarshal covers CWE JSON serialization edge cases.
func TestCWEItem_JSONMarshalUnmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCWEItem_JSONMarshalUnmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			item CWEItem
		}{
			{
				name: "minimal-item",
				item: CWEItem{ID: "CWE-79"},
			},
			{
				name: "item-with-name",
				item: CWEItem{ID: "CWE-79", Name: "Cross-site Scripting"},
			},
			{
				name: "unicode-name",
				item: CWEItem{ID: "CWE-89", Name: "SQL注入 - SQL-инъекция"},
			},
			{
				name: "long-name",
				item: CWEItem{ID: "CWE-78", Name: strings.Repeat("x", 500)},
			},
			{
				name: "special-chars-name",
				item: CWEItem{ID: "CWE-20", Name: "Input\nValidation\t\"Improper\""},
			},
			{
				name: "item-with-description",
				item: CWEItem{ID: "CWE-79", Name: "XSS", Description: "Description text"},
			},
			{
				name: "long-description",
				item: CWEItem{ID: "CWE-79", Name: "XSS", Description: strings.Repeat("desc ", 1000)},
			},
			{
				name: "unicode-description",
				item: CWEItem{ID: "CWE-79", Description: "跨站脚本攻击 🔒"},
			},
			{
				name: "html-in-description",
				item: CWEItem{ID: "CWE-79", Description: "<p>HTML content</p>"},
			},
			{
				name: "item-with-abstraction",
				item: CWEItem{ID: "CWE-79", Abstraction: "Variant"},
			},
			{
				name: "item-with-status",
				item: CWEItem{ID: "CWE-79", Status: "Draft"},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.item.ID {
					t.Fatalf("ID mismatch: want %s got %s", tc.item.ID, decoded.ID)
				}
			})
		}
	})

}

// TestCWEItem_IDFormats validates various CWE ID formats.
func TestCWEItem_IDFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCWEItem_IDFormats", nil, func(t *testing.T, tx *gorm.DB) {
		validIDs := []string{
			"CWE-1",
			"CWE-79",
			"CWE-89",
			"CWE-200",
			"CWE-1234",
			"CWE-9999",
		}

		for _, id := range validIDs {
			t.Run(id, func(t *testing.T) {
				item := CWEItem{ID: id}
				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != id {
					t.Fatalf("ID mismatch: want %s got %s", id, decoded.ID)
				}
			})
		}
	})

}

// TestCWEItem_AbstractionValues validates abstraction enumeration values.
func TestCWEItem_AbstractionValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCWEItem_AbstractionValues", nil, func(t *testing.T, tx *gorm.DB) {
		abstractions := []string{
			"Class",
			"Base",
			"Variant",
			"Compound",
			"",
		}

		for _, abstraction := range abstractions {
			t.Run(fmt.Sprintf("abstraction-%s", abstraction), func(t *testing.T) {
				item := CWEItem{ID: "CWE-79", Abstraction: abstraction}
				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.Abstraction != abstraction {
					t.Fatalf("Abstraction mismatch: want %s got %s", abstraction, decoded.Abstraction)
				}
			})
		}
	})

}

// TestCWEItem_StatusValues validates status enumeration values.
func TestCWEItem_StatusValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCWEItem_StatusValues", nil, func(t *testing.T, tx *gorm.DB) {
		statuses := []string{
			"Draft",
			"Incomplete",
			"Stable",
			"Deprecated",
			"Obsolete",
			"",
		}

		for _, status := range statuses {
			t.Run(fmt.Sprintf("status-%s", status), func(t *testing.T) {
				item := CWEItem{ID: "CWE-79", Status: status}
				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.Status != status {
					t.Fatalf("Status mismatch: want %s got %s", status, decoded.Status)
				}
			})
		}
	})

}

// TestCWEItem_LikelihoodValues validates likelihood enumeration values.
func TestCWEItem_LikelihoodValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCWEItem_LikelihoodValues", nil, func(t *testing.T, tx *gorm.DB) {
		likelihoods := []string{
			"High",
			"Medium",
			"Low",
			"",
			"Unknown",
		}

		for _, likelihood := range likelihoods {
			t.Run(fmt.Sprintf("likelihood-%s", likelihood), func(t *testing.T) {
				item := CWEItem{ID: "CWE-79", LikelihoodOfExploit: likelihood}
				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.LikelihoodOfExploit != likelihood {
					t.Fatalf("Likelihood mismatch: want %s got %s", likelihood, decoded.LikelihoodOfExploit)
				}
			})
		}
	})

}

// TestRelatedWeakness_Formats validates related weakness edge cases.
func TestRelatedWeakness_Formats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestRelatedWeakness_Formats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name     string
			weakness RelatedWeakness
		}{
			{name: "childof", weakness: RelatedWeakness{Nature: "ChildOf", CweID: "CWE-79"}},
			{name: "parentof", weakness: RelatedWeakness{Nature: "ParentOf", CweID: "CWE-20"}},
			{name: "canprecede", weakness: RelatedWeakness{Nature: "CanPrecede", CweID: "CWE-89"}},
			{name: "requires", weakness: RelatedWeakness{Nature: "Requires", CweID: "CWE-78"}},
			{name: "with-viewid", weakness: RelatedWeakness{Nature: "ChildOf", CweID: "CWE-79", ViewID: "1000"}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				item := CWEItem{
					ID:                "CWE-1",
					RelatedWeaknesses: []RelatedWeakness{tc.weakness},
				}

				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if len(decoded.RelatedWeaknesses) != 1 {
					t.Fatalf("Expected 1 related weakness")
				}
			})
		}
	})

}

// TestApplicablePlatform_Formats validates platform edge cases.
func TestApplicablePlatform_Formats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestApplicablePlatform_Formats", nil, func(t *testing.T, tx *gorm.DB) {
		platforms := []ApplicablePlatform{
			{Type: "Language", Name: "Java", Prevalence: "Often"},
			{Type: "Language", Name: "C", Prevalence: "Often"},
			{Type: "Architecture", Name: "x86", Prevalence: "Sometimes"},
			{Type: "Operating_System", Class: "Unix", Prevalence: "Often"},
			{Type: "Technology", Name: "Web", Prevalence: "Often"},
		}

		for i, platform := range platforms {
			t.Run(fmt.Sprintf("platform-%d", i), func(t *testing.T) {
				item := CWEItem{
					ID:                  "CWE-1",
					ApplicablePlatforms: []ApplicablePlatform{platform},
				}

				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if len(decoded.ApplicablePlatforms) != 1 {
					t.Fatalf("Expected 1 platform")
				}
			})
		}
	})

}

// TestReference_Formats validates reference structure.
func TestReference_Formats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReference_Formats", nil, func(t *testing.T, tx *gorm.DB) {
		refs := []Reference{
			{ExternalReferenceID: "REF-1", URL: "https://example.com"},
			{ExternalReferenceID: "REF-2", Title: "Reference Title"},
			{ExternalReferenceID: "REF-3", Authors: []string{"Author1", "Author2"}},
			{ExternalReferenceID: "REF-4", PublicationYear: "2021", PublicationMonth: "12"},
		}

		for i, ref := range refs {
			t.Run(fmt.Sprintf("ref-%d", i), func(t *testing.T) {
				item := CWEItem{
					ID:         "CWE-1",
					References: []Reference{ref},
				}

				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CWEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if len(decoded.References) != 1 {
					t.Fatalf("Expected 1 reference")
				}
			})
		}
	})

}

// TestCWEItem_UnicodeInFields validates unicode handling.
func TestCWEItem_UnicodeInFields(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCWEItem_UnicodeInFields", nil, func(t *testing.T, tx *gorm.DB) {
		item := CWEItem{
			ID:                    "CWE-79",
			Name:                  "跨站脚本-XSS",
			Description:           "描述 - описание - 🔒",
			BackgroundDetails:     []string{"背景-Background", "详情-Details"},
			FunctionalAreas:       []string{"认证-Auth", "授权-Authz"},
			AffectedResources:     []string{"内存-Memory", "文件-File"},
			RelatedAttackPatterns: []string{"CAPEC-18-模式"},
		}

		data, err := json.Marshal(&item)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		var decoded CWEItem
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if decoded.Name != item.Name {
			t.Fatalf("Name mismatch")
		}
	})

}
