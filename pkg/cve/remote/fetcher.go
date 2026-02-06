package remote

import (
	"errors"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"github.com/go-resty/resty/v2"
)

// ErrRateLimited is returned when the NVD API returns a 429 status
var ErrRateLimited = errors.New("NVD API rate limit exceeded")

// Fetcher handles fetching CVE data from the NVD API
type Fetcher struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	// bufferPool reuses temporary byte slices for response bodies
	bufferPool *sync.Pool
}

// NewFetcher creates a new CVE fetcher
func NewFetcher(apiKey string) *Fetcher {
	client := resty.New()
	client.SetTimeout(30 * time.Second)

	// Configure HTTP/2 transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		MaxConnsPerHost:     50,
		DisableCompression:  false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2: true,
	}

	// Configure HTTP/2 specific settings
	if err := http2.ConfigureTransport(transport); err != nil {
		panic(fmt.Sprintf("failed to configure HTTP/2: %v", err))
	}

	client.SetTransport(transport)

	return &Fetcher{
		client:  client,
		baseURL: cve.NVDAPIURL,
		apiKey:  apiKey,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 32*1024) // 32KB initial capacity
				return &b
			},
		},
	}
}

// FetchCVEByID fetches a specific CVE by its ID
func (f *Fetcher) FetchCVEByID(cveID string) (*cve.CVEResponse, error) {
	if cveID == "" {
		return nil, fmt.Errorf("CVE ID cannot be empty")
	}

	req := f.client.R()
	if f.apiKey != "" {
		req.SetHeader("apiKey", f.apiKey)
	}

	resp, err := req.Get(f.baseURL + "?cveId=" + cveID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CVE: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 429 {
			return nil, ErrRateLimited
		}
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	// Prefer using sonic for faster unmarshalling on hot paths
	body := resp.Body()
	var result cve.CVEResponse
	if err := jsonutil.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CVE response: %w", err)
	}

	return &result, nil
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
		SetQueryParam("startIndex", fmt.Sprintf("%d", startIndex)).
		SetQueryParam("resultsPerPage", fmt.Sprintf("%d", resultsPerPage))
	if f.apiKey != "" {
		req.SetHeader("apiKey", f.apiKey)
	}

	resp, err := req.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CVEs: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 429 {
			return nil, ErrRateLimited
		}
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	body := resp.Body()
	var result cve.CVEResponse
	if err := jsonutil.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CVE response: %w", err)
	}

	return &result, nil
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
