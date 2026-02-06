package capec

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

// TestCAPECAttackPattern_XMLMarshalUnmarshal covers CAPEC XML serialization edge cases.
func TestCAPECAttackPattern_XMLMarshalUnmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_XMLMarshalUnmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name    string
			pattern CAPECAttackPattern
		}{
			{
				name:    "minimal-pattern",
				pattern: CAPECAttackPattern{ID: 1},
			},
			{
				name:    "pattern-with-name",
				pattern: CAPECAttackPattern{ID: 1, Name: "Buffer Overflow"},
			},
			{
				name:    "unicode-name",
				pattern: CAPECAttackPattern{ID: 2, Name: "缓冲区溢出 - переполнение буфера"},
			},
			{
				name:    "long-name",
				pattern: CAPECAttackPattern{ID: 3, Name: strings.Repeat("x", 500)},
			},
			{
				name:    "special-chars-name",
				pattern: CAPECAttackPattern{ID: 4, Name: "Attack & Defense <Test>"},
			},
			{
				name:    "pattern-with-abstraction",
				pattern: CAPECAttackPattern{ID: 5, Abstraction: "Detailed"},
			},
			{
				name:    "all-abstractions",
				pattern: CAPECAttackPattern{ID: 6, Abstraction: "Meta"},
			},
			{
				name:    "pattern-with-status",
				pattern: CAPECAttackPattern{ID: 7, Status: "Draft"},
			},
			{
				name:    "pattern-with-summary",
				pattern: CAPECAttackPattern{ID: 8, Summary: "Summary text"},
			},
			{
				name:    "long-summary",
				pattern: CAPECAttackPattern{ID: 9, Summary: strings.Repeat("summary ", 100)},
			},
			{
				name:    "unicode-summary",
				pattern: CAPECAttackPattern{ID: 10, Summary: "概要 - сводка - 🔒"},
			},
			{
				name: "pattern-with-description",
				pattern: CAPECAttackPattern{
					ID:          11,
					Description: InnerXML{XML: "<p>Description text</p>"},
				},
			},
			{
				name: "complex-description",
				pattern: CAPECAttackPattern{
					ID:          12,
					Description: InnerXML{XML: "<div><p>Para 1</p><ul><li>Item 1</li><li>Item 2</li></ul></div>"},
				},
			},
			{
				name: "description-with-cdata",
				pattern: CAPECAttackPattern{
					ID:          13,
					Description: InnerXML{XML: "<![CDATA[<script>alert('xss')</script>]]>"},
				},
			},
			{
				name: "description-with-entities",
				pattern: CAPECAttackPattern{
					ID:          14,
					Description: InnerXML{XML: "&lt;tag&gt; &amp; &quot;quoted&quot;"},
				},
			},
			{
				name:    "pattern-with-likelihood",
				pattern: CAPECAttackPattern{ID: 15, Likelihood: "High"},
			},
			{
				name:    "pattern-with-severity",
				pattern: CAPECAttackPattern{ID: 16, TypicalSeverity: "Very High"},
			},
			{
				name: "pattern-with-weaknesses",
				pattern: CAPECAttackPattern{
					ID: 17,
					RelatedWeaknesses: []RelatedWeakness{
						{CWEID: "CWE-79"},
						{CWEID: "CWE-89"},
					},
				},
			},
			{
				name: "single-weakness",
				pattern: CAPECAttackPattern{
					ID:                18,
					RelatedWeaknesses: []RelatedWeakness{{CWEID: "CWE-78"}},
				},
			},
			{
				name: "many-weaknesses",
				pattern: CAPECAttackPattern{
					ID: 19,
					RelatedWeaknesses: []RelatedWeakness{
						{CWEID: "CWE-1"}, {CWEID: "CWE-2"}, {CWEID: "CWE-3"},
						{CWEID: "CWE-4"}, {CWEID: "CWE-5"},
					},
				},
			},
			{
				name: "pattern-with-examples",
				pattern: CAPECAttackPattern{
					ID: 20,
					Examples: []InnerXML{
						{XML: "<p>Example 1</p>"},
						{XML: "<p>Example 2</p>"},
					},
				},
			},
			{
				name: "complex-examples",
				pattern: CAPECAttackPattern{
					ID: 21,
					Examples: []InnerXML{
						{XML: "<div><h3>Title</h3><pre>code block</pre></div>"},
					},
				},
			},
			{
				name: "pattern-with-mitigations",
				pattern: CAPECAttackPattern{
					ID: 22,
					Mitigations: []InnerXML{
						{XML: "<p>Mitigation 1</p>"},
						{XML: "<p>Mitigation 2</p>"},
					},
				},
			},
			{
				name: "pattern-with-references",
				pattern: CAPECAttackPattern{
					ID: 23,
					References: []RelatedRef{
						{ExternalRef: "REF-1"},
						{ExternalRef: "REF-2"},
					},
				},
			},
			{
				name: "all-fields",
				pattern: CAPECAttackPattern{
					ID:              100,
					Name:            "Full Pattern",
					Abstraction:     "Detailed",
					Status:          "Stable",
					Summary:         "Full summary",
					Description:     InnerXML{XML: "<p>Description</p>"},
					Likelihood:      "High",
					TypicalSeverity: "Very High",
					RelatedWeaknesses: []RelatedWeakness{
						{CWEID: "CWE-79"},
					},
					Examples: []InnerXML{
						{XML: "<p>Example</p>"},
					},
					Mitigations: []InnerXML{
						{XML: "<p>Mitigation</p>"},
					},
					References: []RelatedRef{
						{ExternalRef: "REF-1"},
					},
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := xml.Marshal(&tc.pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.pattern.ID {
					t.Fatalf("ID mismatch: want %d got %d", tc.pattern.ID, decoded.ID)
				}
			})
		}
	})

}

// TestInnerXML_Formats covers InnerXML content variations.
func TestInnerXML_Formats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestInnerXML_Formats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			xml  string
		}{
			{name: "plain-text", xml: "Plain text"},
			{name: "simple-tag", xml: "<p>Paragraph</p>"},
			{name: "nested-tags", xml: "<div><p>Nested</p></div>"},
			{name: "self-closing", xml: "<br />"},
			{name: "multiple-paragraphs", xml: "<p>P1</p><p>P2</p><p>P3</p>"},
			{name: "with-attributes", xml: `<div class="container" id="main">Content</div>`},
			{name: "unicode-content", xml: "<p>你好世界 🌍</p>"},
			{name: "entities", xml: "&lt;tag&gt; &amp; &quot;quotes&quot; &#39;apostrophe&#39;"},
			{name: "cdata", xml: "<![CDATA[<script>alert('test')</script>]]>"},
			{name: "mixed-content", xml: "Text <b>bold</b> more text"},
			{name: "list", xml: "<ul><li>Item 1</li><li>Item 2</li></ul>"},
			{name: "table", xml: "<table><tr><td>Cell</td></tr></table>"},
			{name: "code-block", xml: "<pre><code>function() {}</code></pre>"},
			{name: "long-content", xml: strings.Repeat("<p>Para</p>", 50)},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				pattern := CAPECAttackPattern{
					ID:          1,
					Description: InnerXML{XML: tc.xml},
				}

				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}
			})
		}
	})

}

// TestRelatedWeakness_Formats covers CWE ID format variations.
func TestRelatedWeakness_Formats(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRelatedWeakness_Formats", nil, func(t *testing.T, tx *gorm.DB) {
		cweIDs := []string{
			"CWE-1",
			"CWE-79",
			"CWE-89",
			"CWE-200",
			"CWE-1234",
			"CWE-9999",
		}

		for _, cweID := range cweIDs {
			t.Run(cweID, func(t *testing.T) {
				pattern := CAPECAttackPattern{
					ID:                1,
					RelatedWeaknesses: []RelatedWeakness{{CWEID: cweID}},
				}

				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if len(decoded.RelatedWeaknesses) != 1 {
					t.Fatalf("Expected 1 related weakness")
				}
				if decoded.RelatedWeaknesses[0].CWEID != cweID {
					t.Fatalf("CWE ID mismatch: want %s got %s", cweID, decoded.RelatedWeaknesses[0].CWEID)
				}
			})
		}
	})

}

// TestCAPECAttackPattern_IDRanges validates various CAPEC ID values.
func TestCAPECAttackPattern_IDRanges(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_IDRanges", nil, func(t *testing.T, tx *gorm.DB) {
		ids := []int{1, 10, 100, 500, 1000, 9999}

		for _, id := range ids {
			t.Run(fmt.Sprintf("id-%d", id), func(t *testing.T) {
				pattern := CAPECAttackPattern{ID: id}
				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if decoded.ID != id {
					t.Fatalf("ID mismatch: want %d got %d", id, decoded.ID)
				}
			})
		}
	})

}

// TestCAPECAttackPattern_AbstractionLevels validates abstraction values.
func TestCAPECAttackPattern_AbstractionLevels(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_AbstractionLevels", nil, func(t *testing.T, tx *gorm.DB) {
		abstractions := []string{
			"Meta",
			"Standard",
			"Detailed",
			"",
		}

		for _, abstraction := range abstractions {
			t.Run(fmt.Sprintf("abstraction-%s", abstraction), func(t *testing.T) {
				pattern := CAPECAttackPattern{ID: 1, Abstraction: abstraction}
				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if decoded.Abstraction != abstraction {
					t.Fatalf("Abstraction mismatch: want %s got %s", abstraction, decoded.Abstraction)
				}
			})
		}
	})

}

// TestCAPECAttackPattern_StatusValues validates status values.
func TestCAPECAttackPattern_StatusValues(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_StatusValues", nil, func(t *testing.T, tx *gorm.DB) {
		statuses := []string{
			"Draft",
			"Stable",
			"Deprecated",
			"Obsolete",
			"",
		}

		for _, status := range statuses {
			t.Run(fmt.Sprintf("status-%s", status), func(t *testing.T) {
				pattern := CAPECAttackPattern{ID: 1, Status: status}
				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if decoded.Status != status {
					t.Fatalf("Status mismatch: want %s got %s", status, decoded.Status)
				}
			})
		}
	})

}

// TestCAPECAttackPattern_LikelihoodValues validates likelihood values.
func TestCAPECAttackPattern_LikelihoodValues(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_LikelihoodValues", nil, func(t *testing.T, tx *gorm.DB) {
		likelihoods := []string{
			"High",
			"Medium",
			"Low",
			"",
		}

		for _, likelihood := range likelihoods {
			t.Run(fmt.Sprintf("likelihood-%s", likelihood), func(t *testing.T) {
				pattern := CAPECAttackPattern{ID: 1, Likelihood: likelihood}
				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if decoded.Likelihood != likelihood {
					t.Fatalf("Likelihood mismatch: want %s got %s", likelihood, decoded.Likelihood)
				}
			})
		}
	})

}

// TestCAPECAttackPattern_SeverityValues validates severity values.
func TestCAPECAttackPattern_SeverityValues(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_SeverityValues", nil, func(t *testing.T, tx *gorm.DB) {
		severities := []string{
			"Very High",
			"High",
			"Medium",
			"Low",
			"Very Low",
			"",
		}

		for _, severity := range severities {
			t.Run(fmt.Sprintf("severity-%s", severity), func(t *testing.T) {
				pattern := CAPECAttackPattern{ID: 1, TypicalSeverity: severity}
				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if decoded.TypicalSeverity != severity {
					t.Fatalf("Severity mismatch: want %s got %s", severity, decoded.TypicalSeverity)
				}
			})
		}
	})

}

// TestCAPECAttackPattern_ArraySizes validates various array sizes.
func TestCAPECAttackPattern_ArraySizes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCAPECAttackPattern_ArraySizes", nil, func(t *testing.T, tx *gorm.DB) {
		sizes := []int{0, 1, 5, 10, 50}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
				pattern := CAPECAttackPattern{ID: 1}

				pattern.RelatedWeaknesses = make([]RelatedWeakness, size)
				for i := 0; i < size; i++ {
					pattern.RelatedWeaknesses[i] = RelatedWeakness{CWEID: fmt.Sprintf("CWE-%d", i+1)}
				}

				pattern.Examples = make([]InnerXML, size)
				for i := 0; i < size; i++ {
					pattern.Examples[i] = InnerXML{XML: fmt.Sprintf("<p>Example %d</p>", i+1)}
				}

				data, err := xml.Marshal(&pattern)
				if err != nil {
					t.Fatalf("xml.Marshal failed: %v", err)
				}

				var decoded CAPECAttackPattern
				if err := xml.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("xml.Unmarshal failed: %v", err)
				}

				if len(decoded.RelatedWeaknesses) != size {
					t.Fatalf("RelatedWeaknesses size mismatch: want %d got %d", size, len(decoded.RelatedWeaknesses))
				}
			})
		}
	})

}
