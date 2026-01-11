package remote

import (
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/go-resty/resty/v2"
)

// Fetcher handles fetching CVE data from the NVD API
type Fetcher struct {
	client  *resty.Client
	baseURL string
	apiKey  string
}

// NewFetcher creates a new CVE fetcher
func NewFetcher(apiKey string) *Fetcher {
	client := resty.New()
	client.SetTimeout(30 * time.Second)

	return &Fetcher{
		client:  client,
		baseURL: cve.NVDAPIURL,
		apiKey:  apiKey,
	}
}

// FetchCVEByID fetches a specific CVE by its ID
func (f *Fetcher) FetchCVEByID(cveID string) (*cve.CVEResponse, error) {
	if cveID == "" {
		return nil, fmt.Errorf("CVE ID cannot be empty")
	}

	req := f.client.R().
		SetResult(&cve.CVEResponse{}).
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

	result, ok := resp.Result().(*cve.CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}

// FetchCVEs fetches CVEs with optional filters
func (f *Fetcher) FetchCVEs(startIndex, resultsPerPage int) (*cve.CVEResponse, error) {
	if startIndex < 0 {
		return nil, fmt.Errorf("startIndex must be non-negative")
	}
	if resultsPerPage < 1 || resultsPerPage > 2000 {
		return nil, fmt.Errorf("resultsPerPage must be between 1 and 2000")
	}

	req := f.client.R().
		SetResult(&cve.CVEResponse{}).
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

	result, ok := resp.Result().(*cve.CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}
