//go:build libxml2

package capec

import (
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
)

// Helper function to create a test CAPEC item
func createTestCAPECItem(id int) *CAPECAttackPattern {
	return &CAPECAttackPattern{
		ID:              id,
		Name:            "Test Attack Pattern",
		Abstraction:     "Meta",
		Status:          "Stable",
		Summary:         "This is a test summary for the attack pattern",
		Description:     InnerXML{XML: "<p>This is a detailed description of the attack pattern.</p>"},
		Likelihood:      "High",
		TypicalSeverity: "Critical",
		RelatedWeaknesses: []RelatedWeakness{
			{CWEID: "CWE-79"},
			{CWEID: "CWE-89"},
			{CWEID: "CWE-434"},
		},
		Examples: []InnerXML{
			{XML: "<p>Example 1: Description of the example</p>"},
			{XML: "<p>Example 2: Another example scenario</p>"},
		},
		Mitigations: []InnerXML{
			{XML: "<p>Mitigation 1: First mitigation approach</p>"},
			{XML: "<p>Mitigation 2: Second mitigation approach</p>"},
		},
		References: []RelatedRef{
			{ExternalRef: "REF-001"},
			{ExternalRef: "REF-002"},
		},
	}
}

// BenchmarkCAPECItemJSONMarshal benchmarks marshaling CAPEC items to JSON
func BenchmarkCAPECItemJSONMarshal(b *testing.B) {
	item := createTestCAPECItem(1001)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(item)
		if err != nil {
			b.Fatalf("Failed to marshal: %v", err)
		}
	}
}

// BenchmarkCAPECItemJSONUnmarshal benchmarks unmarshaling CAPEC items from JSON
func BenchmarkCAPECItemJSONUnmarshal(b *testing.B) {
	item := createTestCAPECItem(1001)
	data, _ := json.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CAPECAttackPattern
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal: %v", err)
		}
	}
}

// BenchmarkCAPECItemSonicMarshal benchmarks marshaling CAPEC items using Sonic
func BenchmarkCAPECItemSonicMarshal(b *testing.B) {
	item := createTestCAPECItem(1001)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(item)
		if err != nil {
			b.Fatalf("Failed to marshal with sonic: %v", err)
		}
	}
}

// BenchmarkCAPECItemSonicUnmarshal benchmarks unmarshaling CAPEC items using Sonic
func BenchmarkCAPECItemSonicUnmarshal(b *testing.B) {
	item := createTestCAPECItem(1001)
	data, _ := sonic.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CAPECAttackPattern
		err := sonic.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal with sonic: %v", err)
		}
	}
}

// BenchmarkCAPECItemSonicGet benchmarks getting values from CAPEC items using Sonic AST
func BenchmarkCAPECItemSonicGet(b *testing.B) {
	item := createTestCAPECItem(1001)
	data, _ := sonic.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Parse using Sonic's ast package
		node, err := sonic.Get(data, "id")
		if err != nil {
			b.Fatalf("Failed to get id with sonic: %v", err)
		}
		_, _ = node.Int64()
	}
}

// BenchmarkCAPECItemSliceMarshal benchmarks marshaling slices of CAPEC items
func BenchmarkCAPECItemSliceMarshal(b *testing.B) {
	items := make([]CAPECAttackPattern, 100)
	for i := range items {
		items[i] = *createTestCAPECItem(i + 1000)
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

// BenchmarkCAPECItemSliceUnmarshal benchmarks unmarshaling slices of CAPEC items
func BenchmarkCAPECItemSliceUnmarshal(b *testing.B) {
	items := make([]CAPECAttackPattern, 100)
	for i := range items {
		items[i] = *createTestCAPECItem(i + 1000)
	}
	data, _ := json.Marshal(items)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []CAPECAttackPattern
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal slice: %v", err)
		}
	}
}
