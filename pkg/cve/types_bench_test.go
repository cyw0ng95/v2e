package cve

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"encoding/json"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

// Helper function to create a test CVE item with comprehensive data
func createTestCVEItem(id string) *CVEItem {
	return &CVEItem{
		ID:                    id,
		SourceID:              "nvd@nist.gov",
		Published:             NewNVDTime(time.Now()),
		LastModified:          NewNVDTime(time.Now()),
		VulnStatus:            "Analyzed",
		EvaluatorComment:      "This is a test evaluator comment",
		EvaluatorSolution:     "This is a test evaluator solution",
		EvaluatorImpact:       "This is a test evaluator impact",
		CisaExploitAdd:        "2024-01-01",
		CisaActionDue:         "2024-07-01",
		CisaRequiredAction:    "Apply updates per vendor instructions",
		CisaVulnerabilityName: "Test Vulnerability Name",
		CVETags: []CVETag{
			{
				SourceIdentifier: "nvd@nist.gov",
				Tags:             []string{"Exploited", "Active"},
			},
		},
		Descriptions: []Description{
			{
				Lang:  "en",
				Value: "This is a test description of the CVE vulnerability with detailed information about the security issue.",
			},
		},
		Metrics: &Metrics{
			CvssMetricV31: []CVSSMetricV3{
				{
					Source: "nvd@nist.gov",
					Type:   "Primary",
					CvssData: CVSSDataV3{
						Version:               "3.1",
						VectorString:          "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
						BaseScore:             10.0,
						BaseSeverity:          "CRITICAL",
						AttackVector:          "NETWORK",
						AttackComplexity:      "LOW",
						PrivilegesRequired:    "NONE",
						UserInteraction:       "NONE",
						Scope:                 "CHANGED",
						ConfidentialityImpact: "HIGH",
						IntegrityImpact:       "HIGH",
						AvailabilityImpact:    "HIGH",
					},
					ExploitabilityScore: 3.9,
					ImpactScore:         6.0,
				},
			},
			CvssMetricV2: []CVSSMetricV2{
				{
					Source: "nvd@nist.gov",
					Type:   "Primary",
					CvssData: CVSSDataV2{
						Version:               "2.0",
						VectorString:          "AV:N/AC:L/Au:N/C:C/I:C/A:C",
						BaseScore:             10.0,
						AccessVector:          "NETWORK",
						AccessComplexity:      "LOW",
						Authentication:        "NONE",
						ConfidentialityImpact: "COMPLETE",
						IntegrityImpact:       "COMPLETE",
						AvailabilityImpact:    "COMPLETE",
					},
					BaseSeverity: "HIGH",
				},
			},
		},
		Weaknesses: []Weakness{
			{
				Source: "cna@nist.gov",
				Type:   "Primary",
				Description: []Description{
					{
						Lang:  "en",
						Value: "The software does not properly restrict reading from restricted locations.",
					},
				},
			},
		},
		Configurations: []Config{
			{
				Operator: "OR",
				Nodes: []Node{
					{
						Operator: "AND",
						CPEMatch: []CPEMatch{
							{
								Vulnerable: true,
								Criteria:   "cpe:2.3:a:vendor:product:*:*:*:*:*:*:*:*",
							},
						},
					},
				},
			},
		},
		References: []Reference{
			{
				URL:    "https://example.com/reference",
				Source: "nvd@nist.gov",
				Tags:   []string{"Patch", "Vendor Advisory"},
			},
		},
		VendorComments: []VendorComment{
			{
				Organization: "Test Vendor",
				Comment:      "This is a test vendor comment about the vulnerability.",
				LastModified: NewNVDTime(time.Now()),
			},
		},
	}
}

// BenchmarkCVEItemJSONMarshal benchmarks marshaling CVE items to JSON
func BenchmarkCVEItemJSONMarshal(b *testing.B) {
	item := createTestCVEItem("CVE-2024-0001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(item)
		if err != nil {
			b.Fatalf("Failed to marshal: %v", err)
		}
	}
}

// BenchmarkCVEItemJSONUnmarshal benchmarks unmarshaling CVE items from JSON
func BenchmarkCVEItemJSONUnmarshal(b *testing.B) {
	item := createTestCVEItem("CVE-2024-0001")
	data, _ := json.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CVEItem
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal: %v", err)
		}
	}
}

// BenchmarkCVEItemSonicMarshal benchmarks marshaling CVE items using Sonic
func BenchmarkCVEItemSonicMarshal(b *testing.B) {
	item := createTestCVEItem("CVE-2024-0001")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := sonic.Marshal(item)
		if err != nil {
			b.Fatalf("Failed to marshal with sonic: %v", err)
		}
	}
}

// BenchmarkCVEItemSonicUnmarshal benchmarks unmarshaling CVE items using Sonic
func BenchmarkCVEItemSonicUnmarshal(b *testing.B) {
	item := createTestCVEItem("CVE-2024-0001")
	data, _ := sonic.Marshal(item)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CVEItem
		err := sonic.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal with sonic: %v", err)
		}
	}
}

// BenchmarkCVEResponseJSONMarshal benchmarks marshaling CVE responses to JSON
func BenchmarkCVEResponseJSONMarshal(b *testing.B) {
	response := CVEResponse{
		ResultsPerPage: 10,
		StartIndex:     0,
		TotalResults:   100,
		Format:         "NVD_CVE",
		Version:        "2.0",
		Timestamp:      NewNVDTime(time.Now()),
	}

	// Add 10 vulnerabilities to the response
	for i := 0; i < 10; i++ {
		cveItem := createTestCVEItem("CVE-2024-000" + string(rune('1'+i)))
		response.Vulnerabilities = append(response.Vulnerabilities, struct {
			CVE CVEItem `json:"cve"`
		}{CVE: *cveItem})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(response)
		if err != nil {
			b.Fatalf("Failed to marshal response: %v", err)
		}
	}
}

// BenchmarkCVEResponseJSONUnmarshal benchmarks unmarshaling CVE responses from JSON
func BenchmarkCVEResponseJSONUnmarshal(b *testing.B) {
	response := CVEResponse{
		ResultsPerPage: 10,
		StartIndex:     0,
		TotalResults:   100,
		Format:         "NVD_CVE",
		Version:        "2.0",
		Timestamp:      NewNVDTime(time.Now()),
	}

	// Add 10 vulnerabilities to the response
	for i := 0; i < 10; i++ {
		cveItem := createTestCVEItem("CVE-2024-000" + string(rune('1'+i)))
		response.Vulnerabilities = append(response.Vulnerabilities, struct {
			CVE CVEItem `json:"cve"`
		}{CVE: *cveItem})
	}

	data, _ := json.Marshal(response)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result CVEResponse
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal response: %v", err)
		}
	}
}

// BenchmarkNVDTimeJSONMarshal benchmarks marshaling NVDTime
func BenchmarkNVDTimeJSONMarshal(b *testing.B) {
	nvdTime := NewNVDTime(time.Now())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(nvdTime)
		if err != nil {
			b.Fatalf("Failed to marshal NVDTime: %v", err)
		}
	}
}

// BenchmarkNVDTimeJSONUnmarshal benchmarks unmarshaling NVDTime
func BenchmarkNVDTimeJSONUnmarshal(b *testing.B) {
	nvdTime := NewNVDTime(time.Now())
	data, _ := json.Marshal(nvdTime)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result NVDTime
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal NVDTime: %v", err)
		}
	}
}

// BenchmarkCVEItemSliceMarshal benchmarks marshaling slices of CVE items
func BenchmarkCVEItemSliceMarshal(b *testing.B) {
	items := make([]CVEItem, 50)
	for i := range items {
		cveItem := createTestCVEItem("CVE-2024-00" + string(rune('1'+i%9)) + string(rune('0'+i/9%10)))
		items[i] = *cveItem
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

// BenchmarkCVEItemSliceUnmarshal benchmarks unmarshaling slices of CVE items
func BenchmarkCVEItemSliceUnmarshal(b *testing.B) {
	items := make([]CVEItem, 50)
	for i := range items {
		cveItem := createTestCVEItem("CVE-2024-00" + string(rune('1'+i%9)) + string(rune('0'+i/9%10)))
		items[i] = *cveItem
	}
	data, _ := json.Marshal(items)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []CVEItem
		err := json.Unmarshal(data, &result)
		if err != nil {
			b.Fatalf("Failed to unmarshal slice: %v", err)
		}
	}
}

// BenchmarkCVSSMetricV3Processing benchmarks processing CVSS V3 metrics
func BenchmarkCVSSMetricV3Processing(b *testing.B) {
	cvssMetric := CVSSMetricV3{
		Source: "nvd@nist.gov",
		Type:   "Primary",
		CvssData: CVSSDataV3{
			Version:               "3.1",
			VectorString:          "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
			BaseScore:             10.0,
			BaseSeverity:          "CRITICAL",
			AttackVector:          "NETWORK",
			AttackComplexity:      "LOW",
			PrivilegesRequired:    "NONE",
			UserInteraction:       "NONE",
			Scope:                 "CHANGED",
			ConfidentialityImpact: "HIGH",
			IntegrityImpact:       "HIGH",
			AvailabilityImpact:    "HIGH",
		},
		ExploitabilityScore: 3.9,
		ImpactScore:         6.0,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate processing of CVSS metrics
		_ = cvssMetric.Source
		_ = cvssMetric.Type
		_ = cvssMetric.CvssData.BaseScore
		_ = cvssMetric.CvssData.BaseSeverity
		_ = cvssMetric.ExploitabilityScore
		_ = cvssMetric.ImpactScore
	}
}
