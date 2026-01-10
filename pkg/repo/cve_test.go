package repo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewCVEFetcher(t *testing.T) {
	fetcher := NewCVEFetcher("")
	if fetcher == nil {
		t.Error("NewCVEFetcher should not return nil")
	}
	if fetcher.baseURL != NVDAPIURL {
		t.Errorf("Expected baseURL %s, got %s", NVDAPIURL, fetcher.baseURL)
	}
}

func TestNewCVEFetcher_WithAPIKey(t *testing.T) {
	apiKey := "test-api-key"
	fetcher := NewCVEFetcher(apiKey)
	if fetcher.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, fetcher.apiKey)
	}
}

func TestFetchCVEByID_EmptyID(t *testing.T) {
	fetcher := NewCVEFetcher("")
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
		response := CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Format:         "NVD_CVE",
			Version:        "2.0",
			Timestamp:      time.Now(),
			Vulnerabilities: []struct {
				CVE CVEItem `json:"cve"`
			}{
				{
					CVE: CVEItem{
						ID:           "CVE-2021-44228",
						SourceID:     "nvd@nist.gov",
						Published:    time.Now(),
						LastModified: time.Now(),
						VulnStatus:   "Analyzed",
						Descriptions: []Description{
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
	fetcher := NewCVEFetcher("")
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

	cve := result.Vulnerabilities[0].CVE
	if cve.ID != "CVE-2021-44228" {
		t.Errorf("Expected CVE ID CVE-2021-44228, got %s", cve.ID)
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
		response := CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Vulnerabilities: []struct {
				CVE CVEItem `json:"cve"`
			}{
				{
					CVE: CVEItem{
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
	fetcher := NewCVEFetcher(apiKey)
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
	fetcher := NewCVEFetcher("")
	fetcher.baseURL = server.URL

	// Fetch CVE - should fail
	_, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err == nil {
		t.Error("FetchCVEByID should return error when server returns 500")
	}
}

func TestFetchCVEs_InvalidParameters(t *testing.T) {
	fetcher := NewCVEFetcher("")

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
		response := CVEResponse{
			ResultsPerPage: 10,
			StartIndex:     0,
			TotalResults:   100,
			Format:         "NVD_CVE",
			Version:        "2.0",
			Timestamp:      time.Now(),
			Vulnerabilities: []struct {
				CVE CVEItem `json:"cve"`
			}{
				{
					CVE: CVEItem{
						ID:         "CVE-2021-44228",
						VulnStatus: "Analyzed",
					},
				},
				{
					CVE: CVEItem{
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
	fetcher := NewCVEFetcher("")
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
	fetcher := NewCVEFetcher("")
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
		response := CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Vulnerabilities: []struct {
				CVE CVEItem `json:"cve"`
			}{
				{
					CVE: CVEItem{
						ID: "CVE-2021-44228",
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

	fetcher := NewCVEFetcher("")
	fetcher.baseURL = server.URL

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	if len(result.Vulnerabilities) != 1 {
		t.Fatalf("Expected 1 vulnerability, got %d", len(result.Vulnerabilities))
	}

	cve := result.Vulnerabilities[0].CVE
	if cve.Metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	if len(cve.Metrics.CvssMetricV31) != 1 {
		t.Fatalf("Expected 1 CVSS v3.1 metric, got %d", len(cve.Metrics.CvssMetricV31))
	}

	metric := cve.Metrics.CvssMetricV31[0]
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
		response := CVEResponse{
			ResultsPerPage: 1,
			StartIndex:     0,
			TotalResults:   1,
			Vulnerabilities: []struct {
				CVE CVEItem `json:"cve"`
			}{
				{
					CVE: CVEItem{
						ID: "CVE-2021-44228",
						Metrics: &Metrics{
							CvssMetricV2: []CVSSMetricV2{
								{
									Source:       "nvd@nist.gov",
									Type:         "Primary",
									BaseSeverity: "HIGH",
									CvssData: CVSSDataV2{
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

	fetcher := NewCVEFetcher("")
	fetcher.baseURL = server.URL

	result, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchCVEByID failed: %v", err)
	}

	if len(result.Vulnerabilities) != 1 {
		t.Fatalf("Expected 1 vulnerability, got %d", len(result.Vulnerabilities))
	}

	cve := result.Vulnerabilities[0].CVE
	if cve.Metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	if len(cve.Metrics.CvssMetricV2) != 1 {
		t.Fatalf("Expected 1 CVSS v2.0 metric, got %d", len(cve.Metrics.CvssMetricV2))
	}

	metric := cve.Metrics.CvssMetricV2[0]
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
