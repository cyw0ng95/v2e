package repo

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// NVDAPIURL is the base URL for the NVD CVE API v2.0
	NVDAPIURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"
)

// CVEResponse represents the top-level response from the NVD API
type CVEResponse struct {
	ResultsPerPage  int       `json:"resultsPerPage"`
	StartIndex      int       `json:"startIndex"`
	TotalResults    int       `json:"totalResults"`
	Format          string    `json:"format"`
	Version         string    `json:"version"`
	Timestamp       time.Time `json:"timestamp"`
	Vulnerabilities []struct {
		CVE CVEItem `json:"cve"`
	} `json:"vulnerabilities"`
}

// CVEItem represents a single CVE item from the NVD API
type CVEItem struct {
	ID          string       `json:"id"`
	SourceID    string       `json:"sourceIdentifier"`
	Published   time.Time    `json:"published"`
	LastModified time.Time   `json:"lastModified"`
	VulnStatus  string       `json:"vulnStatus"`
	Descriptions []Description `json:"descriptions"`
	Metrics     *Metrics     `json:"metrics,omitempty"`
	References  []Reference  `json:"references,omitempty"`
}

// Description represents a CVE description
type Description struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// Metrics contains CVSS metrics for a CVE
type Metrics struct {
	CvssMetricV31 []CVSSMetric `json:"cvssMetricV31,omitempty"`
	CvssMetricV30 []CVSSMetric `json:"cvssMetricV30,omitempty"`
	CvssMetricV2  []CVSSMetric `json:"cvssMetricV2,omitempty"`
}

// CVSSMetric represents CVSS scoring data
type CVSSMetric struct {
	Source              string  `json:"source"`
	Type                string  `json:"type"`
	CvssData            CVSSData `json:"cvssData,omitempty"`
	ExploitabilityScore float64 `json:"exploitabilityScore,omitempty"`
	ImpactScore         float64 `json:"impactScore,omitempty"`
}

// CVSSData contains the actual CVSS score information
type CVSSData struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`
}

// Reference represents a reference link for a CVE
type Reference struct {
	URL    string   `json:"url"`
	Source string   `json:"source,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

// CVEFetcher handles fetching CVE data from the NVD API
type CVEFetcher struct {
	client  *resty.Client
	baseURL string
	apiKey  string
}

// NewCVEFetcher creates a new CVE fetcher
func NewCVEFetcher(apiKey string) *CVEFetcher {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	
	return &CVEFetcher{
		client:  client,
		baseURL: NVDAPIURL,
		apiKey:  apiKey,
	}
}

// FetchCVEByID fetches a specific CVE by its ID
func (f *CVEFetcher) FetchCVEByID(cveID string) (*CVEResponse, error) {
	if cveID == "" {
		return nil, fmt.Errorf("CVE ID cannot be empty")
	}

	req := f.client.R().
		SetResult(&CVEResponse{}).
		SetError(&map[string]interface{}{})

	// Add API key if provided
	if f.apiKey != "" {
		req.SetHeader("apiKey", f.apiKey)
	}

	resp, err := req.Get(f.baseURL + "?cveId=" + cveID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CVE: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	result, ok := resp.Result().(*CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}

// FetchCVEs fetches CVEs with optional filters
func (f *CVEFetcher) FetchCVEs(startIndex, resultsPerPage int) (*CVEResponse, error) {
	if startIndex < 0 {
		return nil, fmt.Errorf("startIndex must be non-negative")
	}
	if resultsPerPage < 1 || resultsPerPage > 2000 {
		return nil, fmt.Errorf("resultsPerPage must be between 1 and 2000")
	}

	req := f.client.R().
		SetResult(&CVEResponse{}).
		SetError(&map[string]interface{}{}).
		SetQueryParam("startIndex", fmt.Sprintf("%d", startIndex)).
		SetQueryParam("resultsPerPage", fmt.Sprintf("%d", resultsPerPage))

	// Add API key if provided
	if f.apiKey != "" {
		req.SetHeader("apiKey", f.apiKey)
	}

	resp, err := req.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CVEs: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	result, ok := resp.Result().(*CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}
