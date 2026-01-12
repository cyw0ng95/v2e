package remote

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/go-resty/resty/v2"
)

// ErrRateLimited is returned when the NVD API returns a 429 status
var ErrRateLimited = errors.New("NVD API rate limit exceeded")

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
	
	// Enable HTTP/2 and connection pooling for better performance
	client.SetTransport(&http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false, // Enable compression
	})

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
		// Check for rate limiting
		if resp.StatusCode() == 429 {
			return nil, ErrRateLimited
		}
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
		// Check for rate limiting
		if resp.StatusCode() == 429 {
			return nil, ErrRateLimited
		}
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	result, ok := resp.Result().(*cve.CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}

// FetchCVEsConcurrent fetches multiple CVE IDs concurrently using a worker pool
// Principle 11: Worker pool pattern for parallel processing
func (f *Fetcher) FetchCVEsConcurrent(cveIDs []string, workers int) ([]*cve.CVEResponse, []error) {
if workers <= 0 {
workers = 5 // Default worker count
}

// Channels for job distribution and result collection
jobs := make(chan string, len(cveIDs))
results := make(chan *cve.CVEResponse, len(cveIDs))
errors := make(chan error, len(cveIDs))

// Start worker pool
for w := 0; w < workers; w++ {
go func() {
for cveID := range jobs {
resp, err := f.FetchCVEByID(cveID)
if err != nil {
errors <- err
} else {
results <- resp
}
}
}()
}

// Send jobs
for _, cveID := range cveIDs {
jobs <- cveID
}
close(jobs)

// Collect results
var responses []*cve.CVEResponse
var errs []error
for i := 0; i < len(cveIDs); i++ {
select {
case resp := <-results:
responses = append(responses, resp)
case err := <-errors:
errs = append(errs, err)
}
}

return responses, errs
}
