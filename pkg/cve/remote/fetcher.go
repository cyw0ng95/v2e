package remote

import (
	"errors"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"github.com/go-resty/resty/v2"
)

// ErrRateLimited is returned when the NVD API returns a 429 status
var ErrRateLimited = errors.New("NVD API rate limit exceeded")

// ErrResponseTooLarge is returned when the API response body exceeds the maximum allowed size
var ErrResponseTooLarge = errors.New("API response body exceeds maximum allowed size")

// RateLimitError wraps ErrRateLimited with retry-after information
type RateLimitError struct {
	RetryAfter time.Duration
}

// Error returns the error message
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("NVD API rate limit exceeded, retry after %v", e.RetryAfter)
}

// Unwrap returns the underlying error
func (e *RateLimitError) Unwrap() error {
	return ErrRateLimited
}

// parseRetryAfter parses the Retry-After header from an HTTP response.
// Returns the duration to wait before retrying, or a default duration if the header is invalid.
func parseRetryAfter(resp *resty.Response) time.Duration {
	// Default retry-after duration if header is not present or invalid
	const defaultRetryAfter = 5 * time.Second

	retryAfterHeader := resp.Header().Get("Retry-After")
	if retryAfterHeader == "" {
		return defaultRetryAfter
	}

	// Try to parse as seconds (integer)
	if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
		duration := time.Duration(seconds) * time.Second
		// Limit the retry-after to a reasonable maximum (1 hour)
		if duration > time.Hour {
			return time.Hour
		}
		if duration < time.Second {
			return time.Second
		}
		return duration
	}

	// Try to parse as HTTP-date (e.g., "Wed, 21 Oct 2015 07:28:00 GMT")
	if t, err := http.ParseTime(retryAfterHeader); err == nil {
		duration := time.Until(t)
		// Handle cases where the time is in the past
		if duration <= 0 {
			return time.Second
		}
		// Limit the retry-after to a reasonable maximum (1 hour)
		if duration > time.Hour {
			return time.Hour
		}
		return duration
	}

	// If parsing fails, return default
	return defaultRetryAfter
}

const (
	// MaxResponseSize is the maximum allowed response body size (10 MB)
	// This prevents OOM issues when the API returns malicious or malformed large responses
	MaxResponseSize = 10 * 1024 * 1024
)

// Fetcher handles fetching CVE data from the NVD API
type Fetcher struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	// bufferPool reuses temporary byte slices for response bodies
	bufferPool *sync.Pool
}

// NewFetcher creates a new CVE fetcher
func NewFetcher(apiKey string) (*Fetcher, error) {
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
		return nil, fmt.Errorf("failed to configure HTTP/2: %w", err)
	}

	client.SetTransport(transport)

	return &Fetcher{
		client:  client,
		baseURL: cve.NVDAPIURL,
		apiKey:  apiKey,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 32*1024)
				return &b
			},
		},
	}, nil
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
			return nil, &RateLimitError{RetryAfter: parseRetryAfter(resp)}
		}
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	// Check response body size to prevent OOM on malicious large responses
	body := resp.Body()
	if len(body) > MaxResponseSize {
		return nil, fmt.Errorf("%w: got %d bytes, max %d bytes", ErrResponseTooLarge, len(body), MaxResponseSize)
	}

	// Prefer using sonic for faster unmarshalling on hot paths
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
			return nil, &RateLimitError{RetryAfter: parseRetryAfter(resp)}
		}
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	// Check response body size to prevent OOM on malicious large responses
	body := resp.Body()
	if len(body) > MaxResponseSize {
		return nil, fmt.Errorf("%w: got %d bytes, max %d bytes", ErrResponseTooLarge, len(body), MaxResponseSize)
	}

	var result cve.CVEResponse
	if err := jsonutil.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CVE response: %w", err)
	}

	return &result, nil
}

// FetchCVEsConcurrent fetches multiple CVE IDs concurrently using a worker pool.
// Results are returned in the same order as the input cveIDs slice.
// Principle 11: Worker pool pattern for parallel processing
func (f *Fetcher) FetchCVEsConcurrent(cveIDs []string, workers int) ([]*cve.CVEResponse, []error) {
	if workers <= 0 {
		workers = 5 // Default worker count
	}

	if len(cveIDs) == 0 {
		return nil, nil
	}

	// jobIndex preserves the original index for result ordering
	type jobIndex struct {
		index int
		cveID string
	}

	// resultWithIndex preserves the index for ordered results
	type resultWithIndex struct {
		index int
		resp  *cve.CVEResponse
		err   error
	}

	jobs := make(chan jobIndex, len(cveIDs))
	results := make(chan resultWithIndex, len(cveIDs))

	// Start worker pool
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				resp, err := f.FetchCVEByID(job.cveID)
				results <- resultWithIndex{
					index: job.index,
					resp:  resp,
					err:   err,
				}
			}
		}()
	}

	// Send jobs with their original indices
	for i, cveID := range cveIDs {
		jobs <- jobIndex{index: i, cveID: cveID}
	}
	close(jobs)

	// Wait for all workers to finish, then close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	responses := make([]*cve.CVEResponse, len(cveIDs))
	errs := make([]error, len(cveIDs))

	for result := range results {
		if result.err != nil {
			errs[result.index] = result.err
		} else {
			responses[result.index] = result.resp
		}
	}

	// Filter out nil responses and nil errors for cleaner return values
	var filteredResponses []*cve.CVEResponse
	var filteredErrors []error
	for i := range cveIDs {
		if errs[i] != nil {
			filteredErrors = append(filteredErrors, errs[i])
		} else if responses[i] != nil {
			filteredResponses = append(filteredResponses, responses[i])
		}
	}

	return filteredResponses, filteredErrors
}
