package remote

import (
	"github.com/bytedance/sonic"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
)

func TestNewFetcher(t *testing.T) {
	fetcher := NewFetcher("")
	if fetcher == nil {
		t.Error("NewFetcher should not return nil")
	}
	if fetcher.baseURL != cve.NVDAPIURL {
		t.Errorf("Expected baseURL %s, got %s", cve.NVDAPIURL, fetcher.baseURL)
	}
}

func TestNewFetcher_WithAPIKey(t *testing.T) {
	apiKey := "test-api-key"
	fetcher := NewFetcher(apiKey)
	if fetcher.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, fetcher.apiKey)
	}
}

func TestFetchCVEByID_EmptyID(t *testing.T) {
	fetcher := NewFetcher("")
	_, err := fetcher.FetchCVEByID("")
	if err == nil {
		t.Error("FetchCVEByID should return error for empty CVE ID")
	}
}

func TestFetchCVEByID_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Query().Get("cveId") != "CVE-2021-44228" {
			t.Errorf("Expected cveId CVE-2021-44228, got %s", r.URL.Query().Get("cveId"))
		}

		// Return a mock response
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
								Value: "Apache Log4j2 vulnerability",
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

	// Create fetcher with mock server URL
	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	// Fetch CVE
	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Errorf("FetchCVEByID failed: %v", err)
	}

	if result == nil {
		t.Fatal("FetchCVEByID returned nil result")
	}

	if result.TotalResults != 1 {
		t.Errorf("Expected 1 result, got %d", result.TotalResults)
	}

	if len(result.Vulnerabilities) != 1 {
		t.Fatalf("Expected 1 vulnerability, got %d", len(result.Vulnerabilities))
	}

	cveItem := result.Vulnerabilities[0].CVE
	if cveItem.ID != "CVE-2021-44228" {
		t.Errorf("Expected CVE ID CVE-2021-44228, got %s", cveItem.ID)
	}
}

func TestFetchCVEByID_WithAPIKey(t *testing.T) {
	apiKey := "test-api-key-123"

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the API key header
		if r.Header.Get("apiKey") != apiKey {
			t.Errorf("Expected API key header %s, got %s", apiKey, r.Header.Get("apiKey"))
		}

		// Return a mock response
		response := cve.CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Vulnerabilities: []struct {
				CVE cve.CVEItem `json:"cve"`
			}{
				{
					CVE: cve.CVEItem{
						ID: "CVE-2021-44228",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create fetcher with API key and mock server URL
	fetcher := NewFetcher(apiKey)
	fetcher.baseURL = server.URL

	// Fetch CVE
	_, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Errorf("FetchCVEByID failed: %v", err)
	}
}

func TestFetchCVEByID_ServerError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create fetcher with mock server URL
	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	// Fetch CVE - should fail
	_, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err == nil {
		t.Error("FetchCVEByID should return error when server returns 500")
	}
}

func TestFetchCVEs_InvalidParameters(t *testing.T) {
	fetcher := NewFetcher("")

	// Test negative startIndex
	_, err := fetcher.FetchCVEs(-1, 10)
	if err == nil {
		t.Error("FetchCVEs should return error for negative startIndex")
	}

	// Test resultsPerPage < 1
	_, err = fetcher.FetchCVEs(0, 0)
	if err == nil {
		t.Error("FetchCVEs should return error for resultsPerPage < 1")
	}

	// Test resultsPerPage > 2000
	_, err = fetcher.FetchCVEs(0, 2001)
	if err == nil {
		t.Error("FetchCVEs should return error for resultsPerPage > 2000")
	}
}

func TestFetchCVEs_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the query parameters
		startIndex := r.URL.Query().Get("startIndex")
		resultsPerPage := r.URL.Query().Get("resultsPerPage")

		if startIndex != "0" {
			t.Errorf("Expected startIndex 0, got %s", startIndex)
		}
		if resultsPerPage != "10" {
			t.Errorf("Expected resultsPerPage 10, got %s", resultsPerPage)
		}

		// Return a mock response with multiple CVEs
		response := cve.CVEResponse{
			ResultsPerPage: 10,
			StartIndex:     0,
			TotalResults:   100,
			Format:         "NVD_CVE",
			Version:        "2.0",
			Timestamp:      cve.NewNVDTime(time.Now()),
			Vulnerabilities: []struct {
				CVE cve.CVEItem `json:"cve"`
			}{
				{
					CVE: cve.CVEItem{
						ID:         "CVE-2021-44228",
						VulnStatus: "Analyzed",
					},
				},
				{
					CVE: cve.CVEItem{
						ID:         "CVE-2021-45046",
						VulnStatus: "Analyzed",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create fetcher with mock server URL
	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	// Fetch CVEs
	result, err := fetcher.FetchCVEs(0, 10)
	if err != nil {
		t.Errorf("FetchCVEs failed: %v", err)
	}

	if result == nil {
		t.Fatal("FetchCVEs returned nil result")
	}

	if result.TotalResults != 100 {
		t.Errorf("Expected 100 total results, got %d", result.TotalResults)
	}

	if len(result.Vulnerabilities) != 2 {
		t.Errorf("Expected 2 vulnerabilities, got %d", len(result.Vulnerabilities))
	}
}

func TestFetchCVEs_ServerError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	// Create fetcher with mock server URL
	fetcher := NewFetcher("")
	fetcher.baseURL = server.URL

	// Fetch CVEs - should fail
	_, err := fetcher.FetchCVEs(0, 10)
	if err == nil {
		t.Error("FetchCVEs should return error when server returns 400")
	}
}

func TestCVSSV3_FullFields(t *testing.T) {
	// Create a mock server that returns a CVE with full CVSS v3.1 data
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
						ID: "CVE-2021-44228",
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
										TemporalScore:         9.8,
										TemporalSeverity:      "CRITICAL",
										ExploitCodeMaturity:   "FUNCTIONAL",
										RemediationLevel:      "OFFICIAL_FIX",
										ReportConfidence:      "CONFIRMED",
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	if len(result.Vulnerabilities) != 1 {
		t.Fatalf("Expected 1 vulnerability, got %d", len(result.Vulnerabilities))
	}

	cveItem := result.Vulnerabilities[0].CVE
	if cveItem.Metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	if len(cveItem.Metrics.CvssMetricV31) != 1 {
		t.Fatalf("Expected 1 CVSS v3.1 metric, got %d", len(cveItem.Metrics.CvssMetricV31))
	}

	metric := cveItem.Metrics.CvssMetricV31[0]
	if metric.Source != "nvd@nist.gov" {
		t.Errorf("Expected source nvd@nist.gov, got %s", metric.Source)
	}
	if metric.Type != "Primary" {
		t.Errorf("Expected type Primary, got %s", metric.Type)
	}
	if metric.ExploitabilityScore != 3.9 {
		t.Errorf("Expected exploitability score 3.9, got %f", metric.ExploitabilityScore)
	}
	if metric.ImpactScore != 6.0 {
		t.Errorf("Expected impact score 6.0, got %f", metric.ImpactScore)
	}

	cvss := metric.CvssData
	if cvss.Version != "3.1" {
		t.Errorf("Expected version 3.1, got %s", cvss.Version)
	}
	if cvss.BaseScore != 10.0 {
		t.Errorf("Expected base score 10.0, got %f", cvss.BaseScore)
	}
	if cvss.BaseSeverity != "CRITICAL" {
		t.Errorf("Expected base severity CRITICAL, got %s", cvss.BaseSeverity)
	}
	if cvss.AttackVector != "NETWORK" {
		t.Errorf("Expected attack vector NETWORK, got %s", cvss.AttackVector)
	}
	if cvss.AttackComplexity != "LOW" {
		t.Errorf("Expected attack complexity LOW, got %s", cvss.AttackComplexity)
	}
	if cvss.PrivilegesRequired != "NONE" {
		t.Errorf("Expected privileges required NONE, got %s", cvss.PrivilegesRequired)
	}
	if cvss.UserInteraction != "NONE" {
		t.Errorf("Expected user interaction NONE, got %s", cvss.UserInteraction)
	}
	if cvss.Scope != "CHANGED" {
		t.Errorf("Expected scope CHANGED, got %s", cvss.Scope)
	}
	if cvss.ConfidentialityImpact != "HIGH" {
		t.Errorf("Expected confidentiality impact HIGH, got %s", cvss.ConfidentialityImpact)
	}
	if cvss.IntegrityImpact != "HIGH" {
		t.Errorf("Expected integrity impact HIGH, got %s", cvss.IntegrityImpact)
	}
	if cvss.AvailabilityImpact != "HIGH" {
		t.Errorf("Expected availability impact HIGH, got %s", cvss.AvailabilityImpact)
	}
	if cvss.TemporalScore != 9.8 {
		t.Errorf("Expected temporal score 9.8, got %f", cvss.TemporalScore)
	}
	if cvss.TemporalSeverity != "CRITICAL" {
		t.Errorf("Expected temporal severity CRITICAL, got %s", cvss.TemporalSeverity)
	}
	if cvss.ExploitCodeMaturity != "FUNCTIONAL" {
		t.Errorf("Expected exploit code maturity FUNCTIONAL, got %s", cvss.ExploitCodeMaturity)
	}
	if cvss.RemediationLevel != "OFFICIAL_FIX" {
		t.Errorf("Expected remediation level OFFICIAL_FIX, got %s", cvss.RemediationLevel)
	}
	if cvss.ReportConfidence != "CONFIRMED" {
		t.Errorf("Expected report confidence CONFIRMED, got %s", cvss.ReportConfidence)
	}
}

func TestCVSSV2_FullFields(t *testing.T) {
	// Create a mock server that returns a CVE with full CVSS v2.0 data
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
						ID: "CVE-2021-44228",
						Metrics: &cve.Metrics{
							CvssMetricV2: []cve.CVSSMetricV2{
								{
									Source:       "nvd@nist.gov",
									Type:         "Primary",
									BaseSeverity: "HIGH",
									CvssData: cve.CVSSDataV2{
										Version:               "2.0",
										VectorString:          "AV:N/AC:M/Au:N/C:C/I:C/A:C",
										BaseScore:             9.3,
										AccessVector:          "NETWORK",
										AccessComplexity:      "MEDIUM",
										Authentication:        "NONE",
										ConfidentialityImpact: "COMPLETE",
										IntegrityImpact:       "COMPLETE",
										AvailabilityImpact:    "COMPLETE",
										TemporalScore:         8.5,
										Exploitability:        "FUNCTIONAL",
										RemediationLevel:      "OFFICIAL_FIX",
										ReportConfidence:      "CONFIRMED",
									},
									ExploitabilityScore:     8.6,
									ImpactScore:             10.0,
									AcInsufInfo:             false,
									ObtainAllPrivilege:      false,
									ObtainUserPrivilege:     false,
									ObtainOtherPrivilege:    false,
									UserInteractionRequired: false,
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	if len(result.Vulnerabilities) != 1 {
		t.Fatalf("Expected 1 vulnerability, got %d", len(result.Vulnerabilities))
	}

	cveItem := result.Vulnerabilities[0].CVE
	if cveItem.Metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	if len(cveItem.Metrics.CvssMetricV2) != 1 {
		t.Fatalf("Expected 1 CVSS v2.0 metric, got %d", len(cveItem.Metrics.CvssMetricV2))
	}

	metric := cveItem.Metrics.CvssMetricV2[0]
	if metric.Source != "nvd@nist.gov" {
		t.Errorf("Expected source nvd@nist.gov, got %s", metric.Source)
	}
	if metric.Type != "Primary" {
		t.Errorf("Expected type Primary, got %s", metric.Type)
	}
	if metric.BaseSeverity != "HIGH" {
		t.Errorf("Expected base severity HIGH, got %s", metric.BaseSeverity)
	}
	if metric.ExploitabilityScore != 8.6 {
		t.Errorf("Expected exploitability score 8.6, got %f", metric.ExploitabilityScore)
	}
	if metric.ImpactScore != 10.0 {
		t.Errorf("Expected impact score 10.0, got %f", metric.ImpactScore)
	}
	if metric.AcInsufInfo != false {
		t.Errorf("Expected acInsufInfo false, got %t", metric.AcInsufInfo)
	}
	if metric.ObtainAllPrivilege != false {
		t.Errorf("Expected obtainAllPrivilege false, got %t", metric.ObtainAllPrivilege)
	}
	if metric.UserInteractionRequired != false {
		t.Errorf("Expected userInteractionRequired false, got %t", metric.UserInteractionRequired)
	}

	cvss := metric.CvssData
	if cvss.Version != "2.0" {
		t.Errorf("Expected version 2.0, got %s", cvss.Version)
	}
	if cvss.BaseScore != 9.3 {
		t.Errorf("Expected base score 9.3, got %f", cvss.BaseScore)
	}
	if cvss.AccessVector != "NETWORK" {
		t.Errorf("Expected access vector NETWORK, got %s", cvss.AccessVector)
	}
	if cvss.AccessComplexity != "MEDIUM" {
		t.Errorf("Expected access complexity MEDIUM, got %s", cvss.AccessComplexity)
	}
	if cvss.Authentication != "NONE" {
		t.Errorf("Expected authentication NONE, got %s", cvss.Authentication)
	}
	if cvss.ConfidentialityImpact != "COMPLETE" {
		t.Errorf("Expected confidentiality impact COMPLETE, got %s", cvss.ConfidentialityImpact)
	}
	if cvss.IntegrityImpact != "COMPLETE" {
		t.Errorf("Expected integrity impact COMPLETE, got %s", cvss.IntegrityImpact)
	}
	if cvss.AvailabilityImpact != "COMPLETE" {
		t.Errorf("Expected availability impact COMPLETE, got %s", cvss.AvailabilityImpact)
	}
	if cvss.TemporalScore != 8.5 {
		t.Errorf("Expected temporal score 8.5, got %f", cvss.TemporalScore)
	}
	if cvss.Exploitability != "FUNCTIONAL" {
		t.Errorf("Expected exploitability FUNCTIONAL, got %s", cvss.Exploitability)
	}
	if cvss.RemediationLevel != "OFFICIAL_FIX" {
		t.Errorf("Expected remediation level OFFICIAL_FIX, got %s", cvss.RemediationLevel)
	}
	if cvss.ReportConfidence != "CONFIRMED" {
		t.Errorf("Expected report confidence CONFIRMED, got %s", cvss.ReportConfidence)
	}
}

// TestCVEItem_ExtendedFields tests the new extended fields in CVEItem
func TestCVEItem_ExtendedFields(t *testing.T) {
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
						ID:                    "CVE-2021-44228",
						EvaluatorComment:      "Critical vulnerability",
						EvaluatorSolution:     "Update to latest version",
						EvaluatorImpact:       "Remote code execution",
						CisaExploitAdd:        "2021-12-10",
						CisaActionDue:         "2021-12-24",
						CisaRequiredAction:    "Apply updates",
						CisaVulnerabilityName: "Log4Shell",
						CVETags: []cve.CVETag{
							{
								SourceIdentifier: "nvd@nist.gov",
								Tags:             []string{"disputed"},
							},
						},
						Descriptions: []cve.Description{
							{Lang: "en", Value: "Test description"},
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	cveItem := result.Vulnerabilities[0].CVE
	if cveItem.EvaluatorComment != "Critical vulnerability" {
		t.Errorf("Expected evaluator comment 'Critical vulnerability', got %s", cveItem.EvaluatorComment)
	}
	if cveItem.EvaluatorSolution != "Update to latest version" {
		t.Errorf("Expected evaluator solution 'Update to latest version', got %s", cveItem.EvaluatorSolution)
	}
	if cveItem.EvaluatorImpact != "Remote code execution" {
		t.Errorf("Expected evaluator impact 'Remote code execution', got %s", cveItem.EvaluatorImpact)
	}
	if cveItem.CisaExploitAdd != "2021-12-10" {
		t.Errorf("Expected CISA exploit add '2021-12-10', got %s", cveItem.CisaExploitAdd)
	}
	if cveItem.CisaActionDue != "2021-12-24" {
		t.Errorf("Expected CISA action due '2021-12-24', got %s", cveItem.CisaActionDue)
	}
	if cveItem.CisaRequiredAction != "Apply updates" {
		t.Errorf("Expected CISA required action 'Apply updates', got %s", cveItem.CisaRequiredAction)
	}
	if cveItem.CisaVulnerabilityName != "Log4Shell" {
		t.Errorf("Expected CISA vulnerability name 'Log4Shell', got %s", cveItem.CisaVulnerabilityName)
	}
	if len(cveItem.CVETags) != 1 {
		t.Fatalf("Expected 1 CVE tag, got %d", len(cveItem.CVETags))
	}
	if cveItem.CVETags[0].SourceIdentifier != "nvd@nist.gov" {
		t.Errorf("Expected source identifier 'nvd@nist.gov', got %s", cveItem.CVETags[0].SourceIdentifier)
	}
	if len(cveItem.CVETags[0].Tags) != 1 || cveItem.CVETags[0].Tags[0] != "disputed" {
		t.Errorf("Expected tag 'disputed', got %v", cveItem.CVETags[0].Tags)
	}
}

// TestWeakness tests the Weakness object
func TestWeakness(t *testing.T) {
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
						ID: "CVE-2021-44228",
						Weaknesses: []cve.Weakness{
							{
								Source: "nvd@nist.gov",
								Type:   "Primary",
								Description: []cve.Description{
									{Lang: "en", Value: "CWE-502"},
								},
							},
						},
						Descriptions: []cve.Description{
							{Lang: "en", Value: "Test description"},
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	cveItem := result.Vulnerabilities[0].CVE
	if len(cveItem.Weaknesses) != 1 {
		t.Fatalf("Expected 1 weakness, got %d", len(cveItem.Weaknesses))
	}
	weakness := cveItem.Weaknesses[0]
	if weakness.Source != "nvd@nist.gov" {
		t.Errorf("Expected source 'nvd@nist.gov', got %s", weakness.Source)
	}
	if weakness.Type != "Primary" {
		t.Errorf("Expected type 'Primary', got %s", weakness.Type)
	}
	if len(weakness.Description) != 1 || weakness.Description[0].Value != "CWE-502" {
		t.Errorf("Expected description 'CWE-502', got %v", weakness.Description)
	}
}

// TestConfiguration tests the Configuration object
func TestConfiguration(t *testing.T) {
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
						ID: "CVE-2021-44228",
						Configurations: []cve.Config{
							{
								Operator: "AND",
								Negate:   false,
								Nodes: []cve.Node{
									{
										Operator: "OR",
										Negate:   false,
										CPEMatch: []cve.CPEMatch{
											{
												Vulnerable:            true,
												Criteria:              "cpe:2.3:a:apache:log4j:2.0:*:*:*:*:*:*:*",
												MatchCriteriaID:       "12345678-1234-1234-1234-123456789012",
												VersionStartIncluding: "2.0",
												VersionEndExcluding:   "2.15.0",
											},
										},
									},
								},
							},
						},
						Descriptions: []cve.Description{
							{Lang: "en", Value: "Test description"},
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	cveItem := result.Vulnerabilities[0].CVE
	if len(cveItem.Configurations) != 1 {
		t.Fatalf("Expected 1 configuration, got %d", len(cveItem.Configurations))
	}

	config := cveItem.Configurations[0]
	if config.Operator != "AND" {
		t.Errorf("Expected operator 'AND', got %s", config.Operator)
	}
	if config.Negate != false {
		t.Errorf("Expected negate false, got %t", config.Negate)
	}
	if len(config.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(config.Nodes))
	}

	node := config.Nodes[0]
	if node.Operator != "OR" {
		t.Errorf("Expected node operator 'OR', got %s", node.Operator)
	}
	if len(node.CPEMatch) != 1 {
		t.Fatalf("Expected 1 CPE match, got %d", len(node.CPEMatch))
	}

	cpeMatch := node.CPEMatch[0]
	if !cpeMatch.Vulnerable {
		t.Errorf("Expected vulnerable true, got %t", cpeMatch.Vulnerable)
	}
	if cpeMatch.Criteria != "cpe:2.3:a:apache:log4j:2.0:*:*:*:*:*:*:*" {
		t.Errorf("Expected criteria 'cpe:2.3:a:apache:log4j:2.0:*:*:*:*:*:*:*', got %s", cpeMatch.Criteria)
	}
	if cpeMatch.MatchCriteriaID != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("Expected match criteria ID '12345678-1234-1234-1234-123456789012', got %s", cpeMatch.MatchCriteriaID)
	}
	if cpeMatch.VersionStartIncluding != "2.0" {
		t.Errorf("Expected version start including '2.0', got %s", cpeMatch.VersionStartIncluding)
	}
	if cpeMatch.VersionEndExcluding != "2.15.0" {
		t.Errorf("Expected version end excluding '2.15.0', got %s", cpeMatch.VersionEndExcluding)
	}
}

// TestVendorComment tests the VendorComment object
func TestVendorComment(t *testing.T) {
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
						ID: "CVE-2021-44228",
						VendorComments: []cve.VendorComment{
							{
								Organization: "Apache Software Foundation",
								Comment:      "Fixed in version 2.15.0",
								LastModified: cve.NewNVDTime(time.Date(2021, 12, 10, 0, 0, 0, 0, time.UTC)),
							},
						},
						Descriptions: []cve.Description{
							{Lang: "en", Value: "Test description"},
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	cveItem := result.Vulnerabilities[0].CVE
	if len(cveItem.VendorComments) != 1 {
		t.Fatalf("Expected 1 vendor comment, got %d", len(cveItem.VendorComments))
	}

	comment := cveItem.VendorComments[0]
	if comment.Organization != "Apache Software Foundation" {
		t.Errorf("Expected organization 'Apache Software Foundation', got %s", comment.Organization)
	}
	if comment.Comment != "Fixed in version 2.15.0" {
		t.Errorf("Expected comment 'Fixed in version 2.15.0', got %s", comment.Comment)
	}
	expectedTime := time.Date(2021, 12, 10, 0, 0, 0, 0, time.UTC)
	if !comment.LastModified.Equal(expectedTime) {
		t.Errorf("Expected last modified %v, got %v", expectedTime, comment.LastModified)
	}
}

// TestCVSSV40_FullFields tests CVSS v4.0 with full fields
func TestCVSSV40_FullFields(t *testing.T) {
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
						ID: "CVE-2021-44228",
						Metrics: &cve.Metrics{
							CvssMetricV40: []cve.CVSSMetricV40{
								{
									Source: "nvd@nist.gov",
									Type:   "Primary",
									CvssData: cve.CVSSDataV40{
										Version:                           "4.0",
										VectorString:                      "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N",
										BaseScore:                         10.0,
										BaseSeverity:                      "CRITICAL",
										AttackVector:                      "NETWORK",
										AttackComplexity:                  "LOW",
										AttackRequirements:                "NONE",
										PrivilegesRequired:                "NONE",
										UserInteraction:                   "NONE",
										VulnConfidentialityImpact:         "HIGH",
										VulnIntegrityImpact:               "HIGH",
										VulnAvailabilityImpact:            "HIGH",
										SubConfidentialityImpact:          "NONE",
										SubIntegrityImpact:                "NONE",
										SubAvailabilityImpact:             "NONE",
										ExploitMaturity:                   "ATTACKED",
										ConfidentialityRequirement:        "HIGH",
										IntegrityRequirement:              "HIGH",
										AvailabilityRequirement:           "HIGH",
										ModifiedAttackVector:              "NETWORK",
										ModifiedAttackComplexity:          "LOW",
										ModifiedAttackRequirements:        "NONE",
										ModifiedPrivilegesRequired:        "NONE",
										ModifiedUserInteraction:           "NONE",
										ModifiedVulnConfidentialityImpact: "HIGH",
										ModifiedVulnIntegrityImpact:       "HIGH",
										ModifiedVulnAvailabilityImpact:    "HIGH",
										Safety:                            "NOT_DEFINED",
										Automatable:                       "YES",
										Recovery:                          "NOT_DEFINED",
										ValueDensity:                      "NOT_DEFINED",
										VulnerabilityResponseEffort:       "NOT_DEFINED",
										ProviderUrgency:                   "NOT_DEFINED",
									},
								},
							},
						},
						Descriptions: []cve.Description{
							{Lang: "en", Value: "Test description"},
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

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	cveItem := result.Vulnerabilities[0].CVE
	if cveItem.Metrics == nil {
		t.Fatal("Metrics should not be nil")
	}
	if len(cveItem.Metrics.CvssMetricV40) != 1 {
		t.Fatalf("Expected 1 CVSS v4.0 metric, got %d", len(cveItem.Metrics.CvssMetricV40))
	}

	metric := cveItem.Metrics.CvssMetricV40[0]
	if metric.Source != "nvd@nist.gov" {
		t.Errorf("Expected source 'nvd@nist.gov', got %s", metric.Source)
	}
	if metric.Type != "Primary" {
		t.Errorf("Expected type 'Primary', got %s", metric.Type)
	}

	cvss := metric.CvssData
	if cvss.Version != "4.0" {
		t.Errorf("Expected version '4.0', got %s", cvss.Version)
	}
	if cvss.BaseScore != 10.0 {
		t.Errorf("Expected base score 10.0, got %f", cvss.BaseScore)
	}
	if cvss.BaseSeverity != "CRITICAL" {
		t.Errorf("Expected base severity 'CRITICAL', got %s", cvss.BaseSeverity)
	}
	if cvss.AttackVector != "NETWORK" {
		t.Errorf("Expected attack vector 'NETWORK', got %s", cvss.AttackVector)
	}
	if cvss.AttackComplexity != "LOW" {
		t.Errorf("Expected attack complexity 'LOW', got %s", cvss.AttackComplexity)
	}
	if cvss.AttackRequirements != "NONE" {
		t.Errorf("Expected attack requirements 'NONE', got %s", cvss.AttackRequirements)
	}
	if cvss.PrivilegesRequired != "NONE" {
		t.Errorf("Expected privileges required 'NONE', got %s", cvss.PrivilegesRequired)
	}
	if cvss.UserInteraction != "NONE" {
		t.Errorf("Expected user interaction 'NONE', got %s", cvss.UserInteraction)
	}
	if cvss.VulnConfidentialityImpact != "HIGH" {
		t.Errorf("Expected vulnerable confidentiality impact 'HIGH', got %s", cvss.VulnConfidentialityImpact)
	}
	if cvss.VulnIntegrityImpact != "HIGH" {
		t.Errorf("Expected vulnerable integrity impact 'HIGH', got %s", cvss.VulnIntegrityImpact)
	}
	if cvss.VulnAvailabilityImpact != "HIGH" {
		t.Errorf("Expected vulnerable availability impact 'HIGH', got %s", cvss.VulnAvailabilityImpact)
	}
	if cvss.SubConfidentialityImpact != "NONE" {
		t.Errorf("Expected subsequent confidentiality impact 'NONE', got %s", cvss.SubConfidentialityImpact)
	}
	if cvss.SubIntegrityImpact != "NONE" {
		t.Errorf("Expected subsequent integrity impact 'NONE', got %s", cvss.SubIntegrityImpact)
	}
	if cvss.SubAvailabilityImpact != "NONE" {
		t.Errorf("Expected subsequent availability impact 'NONE', got %s", cvss.SubAvailabilityImpact)
	}
	if cvss.ExploitMaturity != "ATTACKED" {
		t.Errorf("Expected exploit maturity 'ATTACKED', got %s", cvss.ExploitMaturity)
	}
	if cvss.Automatable != "YES" {
		t.Errorf("Expected automatable 'YES', got %s", cvss.Automatable)
	}
}
