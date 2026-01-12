package remote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
)

// BenchmarkNewFetcher benchmarks creating a new fetcher instance
func BenchmarkNewFetcher(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewFetcher("")
	}
}

// BenchmarkFetchCVEByID benchmarks fetching a single CVE by ID
func BenchmarkFetchCVEByID(b *testing.B) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := cve.CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Format:         "NVD_CVE",
			Version:        "2.0",
			Timestamp:      cve.NewNVDTime(time.Now()),
			Vulnerabilities: []struct {
				CVE cve.CVEItem `json:"cve"`
			}{
				{
					CVE: cve.CVEItem{
						ID:           "CVE-2021-44228",
						SourceID:     "nvd@nist.gov",
						Published:    cve.NewNVDTime(time.Now()),
						LastModified: cve.NewNVDTime(time.Now()),
						VulnStatus:   "Analyzed",
						Descriptions: []cve.Description{
							{
								Lang:  "en",
								Value: "Apache Log4j2 vulnerability allows remote code execution",
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fetcher.FetchCVEByID("CVE-2021-44228")
	}
}

// BenchmarkFetchCVEs benchmarks fetching multiple CVEs
func BenchmarkFetchCVEs(b *testing.B) {
	// Create a mock server that returns multiple CVEs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vulnerabilities := make([]struct {
			CVE cve.CVEItem `json:"cve"`
		}, 10)

		for i := 0; i < 10; i++ {
			vulnerabilities[i] = struct {
				CVE cve.CVEItem `json:"cve"`
			}{
				CVE: cve.CVEItem{
					ID:           "CVE-2021-0000" + string(rune('0'+i)),
					VulnStatus:   "Analyzed",
					Published:    cve.NewNVDTime(time.Now()),
					LastModified: cve.NewNVDTime(time.Now()),
					Descriptions: []cve.Description{
						{
							Lang:  "en",
							Value: "Test CVE description",
						},
					},
				},
			}
		}

		response := cve.CVEResponse{
			ResultsPerPage:  10,
			StartIndex:      0,
			TotalResults:    100,
			Format:          "NVD_CVE",
			Version:         "2.0",
			Timestamp:       cve.NewNVDTime(time.Now()),
			Vulnerabilities: vulnerabilities,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fetcher.FetchCVEs(0, 10)
	}
}

// BenchmarkFetchCVEWithCVSS benchmarks fetching CVE with CVSS metrics
func BenchmarkFetchCVEWithCVSS(b *testing.B) {
	// Create a mock server with full CVSS data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := cve.CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Vulnerabilities: []struct {
				CVE cve.CVEItem `json:"cve"`
			}{
				{
					CVE: cve.CVEItem{
						ID:           "CVE-2021-44228",
						Published:    cve.NewNVDTime(time.Now()),
						LastModified: cve.NewNVDTime(time.Now()),
						Descriptions: []cve.Description{
							{Lang: "en", Value: "Test"},
						},
						Metrics: &cve.Metrics{
							CvssMetricV31: []cve.CVSSMetricV3{
								{
									Source: "nvd@nist.gov",
									Type:   "Primary",
									CvssData: cve.CVSSDataV3{
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
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fetcher.FetchCVEByID("CVE-2021-44228")
	}
}

// BenchmarkFetchCVEParseJSON benchmarks JSON parsing of CVE response
func BenchmarkFetchCVEParseJSON(b *testing.B) {
	// Pre-create a JSON response
	response := cve.CVEResponse{
		ResultsPerPage: 1,
		StartIndex:     0,
		TotalResults:   1,
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{
			{
				CVE: cve.CVEItem{
					ID:           "CVE-2021-44228",
					Published:    cve.NewNVDTime(time.Now()),
					LastModified: cve.NewNVDTime(time.Now()),
					Descriptions: []cve.Description{
						{Lang: "en", Value: "Test description"},
					},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(response)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result cve.CVEResponse
		_ = json.Unmarshal(jsonData, &result)
	}
}

// BenchmarkFetchCVEsLargeResponse benchmarks handling larger response (100 CVEs)
func BenchmarkFetchCVEsLargeResponse(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vulnerabilities := make([]struct {
			CVE cve.CVEItem `json:"cve"`
		}, 100)

		for i := 0; i < 100; i++ {
			vulnerabilities[i] = struct {
				CVE cve.CVEItem `json:"cve"`
			}{
				CVE: cve.CVEItem{
					ID:           "CVE-2021-0000" + string(rune('0'+i%10)),
					VulnStatus:   "Analyzed",
					Published:    cve.NewNVDTime(time.Now()),
					LastModified: cve.NewNVDTime(time.Now()),
					Descriptions: []cve.Description{
						{Lang: "en", Value: "Test CVE description with some detailed information"},
					},
					Metrics: &cve.Metrics{
						CvssMetricV31: []cve.CVSSMetricV3{
							{
								Source: "nvd@nist.gov",
								Type:   "Primary",
								CvssData: cve.CVSSDataV3{
									Version:      "3.1",
									BaseScore:    7.5,
									BaseSeverity: "HIGH",
								},
							},
						},
					},
				},
			}
		}

		response := cve.CVEResponse{
			ResultsPerPage:  100,
			StartIndex:      0,
			TotalResults:    1000,
			Vulnerabilities: vulnerabilities,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fetcher.FetchCVEs(0, 100)
	}
}
