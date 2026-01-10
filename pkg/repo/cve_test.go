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
