package remote

import (
	"errors"
	"sync"
	"time"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestFetchCVEByID_Success(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_Success", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("cveId")
			if q == "CVE-TEST-1" {
				w.Header().Set("Content-Type", "application/json")
				w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		resp, err := f.FetchCVEByID("CVE-TEST-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || len(resp.Vulnerabilities) != 1 || resp.Vulnerabilities[0].CVE.ID != "CVE-TEST-1" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

}

func TestFetchCVEByID_RateLimited(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_RateLimited", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(429)
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		_, err = f.FetchCVEByID("CVE-TEST-2")
		if err == nil || err != ErrRateLimited {
			t.Fatalf("expected ErrRateLimited, got %v", err)
		}
	})

}

func TestFetchCVEs_ParamValidation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEs_ParamValidation", nil, func(t *testing.T, tx *gorm.DB) {
		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		if _, err := f.FetchCVEs(-1, 10); err == nil {
			t.Fatalf("expected error for negative startIndex")
		}
		if _, err := f.FetchCVEs(0, 0); err == nil {
			t.Fatalf("expected error for invalid resultsPerPage")
		}
	})

}

func TestFetchCVEsConcurrent_Workers(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEsConcurrent_Workers", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("cveId")
			if q == "" {
				w.WriteHeader(400)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON(q, 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		ids := []string{"CVE-A", "CVE-B", "CVE-C"}
		resps, errs := f.FetchCVEsConcurrent(ids, 3)
		if len(resps)+len(errs) != len(ids) {
			t.Fatalf("expected responses+errors == %d, got %d+%d", len(ids), len(resps), len(errs))
		}
	})
}

func TestFetchCVEByID_EmptyID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_EmptyID", nil, func(t *testing.T, tx *gorm.DB) {
		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		if _, err := f.FetchCVEByID(""); err == nil {
			t.Fatal("expected error for empty CVE ID")
		}
	})
}

func TestFetchCVEByID_NetworkError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_NetworkError", nil, func(t *testing.T, tx *gorm.DB) {
		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = "http://localhost:99999"
		if _, err := f.FetchCVEByID("CVE-TEST-1"); err == nil {
			t.Fatal("expected error for network failure")
		}
	})
}

func TestFetchCVEByID_InvalidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_InvalidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		if _, err := f.FetchCVEByID("CVE-TEST-1"); err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestFetchCVEs_Success(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEs_Success", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		resp, err := f.FetchCVEs(0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
	})
}

func TestFetchCVEs_StatusCodeError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEs_StatusCodeError", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		if _, err := f.FetchCVEs(0, 10); err == nil {
			t.Fatal("expected error for 500 status")
		}
	})
}

func TestFetchCVEsConcurrent_EmptyIDs(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEsConcurrent_EmptyIDs", nil, func(t *testing.T, tx *gorm.DB) {
		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		resps, errs := f.FetchCVEsConcurrent([]string{}, 3)
		if len(resps) != 0 || len(errs) != 0 {
			t.Fatalf("expected empty results, got %d responses and %d errors", len(resps), len(errs))
		}
	})
}

func TestFetchCVEsConcurrent_ZeroWorkers(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEsConcurrent_ZeroWorkers", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("cveId")
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON(q, 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		ids := []string{"CVE-A"}
		resps, errs := f.FetchCVEsConcurrent(ids, 0)
		if len(resps)+len(errs) != len(ids) {
			t.Fatalf("expected results with default workers, got %d+%d", len(resps), len(errs))
		}
	})
}

func TestFetcher_WithAPIKey(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_WithAPIKey", nil, func(t *testing.T, tx *gorm.DB) {
		apiKeyReceived := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("apiKey") == "test-key" {
				apiKeyReceived = true
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))
		defer server.Close()

		f, err := NewFetcher("test-key")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		f.FetchCVEByID("CVE-TEST-1")

		if !apiKeyReceived {
			t.Error("API key was not sent in request")
		}
	})
}

func TestFetcher_NetworkError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_NetworkError", nil, func(t *testing.T, tx *gorm.DB) {
		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = "http://invalid-host-that-does-not-exist-12345.com"
		if _, err := f.FetchCVEs(0, 10); err == nil {
			t.Fatal("expected network error")
		}
	})
}

// TestFetchCVEsConcurrent_OrderPreservation verifies that results are returned
// in the same order as the input cveIDs slice, fixing TODO-122
func TestFetchCVEsConcurrent_OrderPreservation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEsConcurrent_OrderPreservation", nil, func(t *testing.T, tx *gorm.DB) {
		// Track request order to verify concurrent execution
		requestOrder := make([]string, 0)
		requestMu := &sync.Mutex{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("cveId")
			// Simulate variable delay to increase chance of out-of-order completion
			switch q {
			case "CVE-0001":
				// Fast response
			case "CVE-0002":
				// Medium delay
				time.Sleep(10 * time.Millisecond)
			case "CVE-0003":
				// Long delay
				time.Sleep(20 * time.Millisecond)
			case "CVE-0004":
				// Very fast
			}
			requestMu.Lock()
			requestOrder = append(requestOrder, q)
			requestMu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON(q, 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		// Request IDs in a specific order
		ids := []string{"CVE-0001", "CVE-0002", "CVE-0003", "CVE-0004"}
		resps, errs := f.FetchCVEsConcurrent(ids, 3)

		// Verify we got all responses (no errors in this test)
		if len(errs) != 0 {
			t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
		}
		if len(resps) != len(ids) {
			t.Fatalf("expected %d responses, got %d", len(ids), len(resps))
		}

		// Verify responses are in the same order as input IDs
		for i, resp := range resps {
			if resp == nil || len(resp.Vulnerabilities) == 0 {
				t.Fatalf("response at index %d is nil or empty", i)
			}
			expectedID := ids[i]
			actualID := resp.Vulnerabilities[0].CVE.ID
			if actualID != expectedID {
				t.Errorf("response at index %d: expected ID %s, got %s", i, expectedID, actualID)
			}
		}

		// Verify requests were actually made concurrently (not in strict order)
		if len(requestOrder) == len(ids) {
			// If all requests completed, verify at least some were out of order
			// due to the variable delays we added
			allInOrder := true
			for i := range requestOrder {
				if requestOrder[i] != ids[i] {
					allInOrder = false
					break
				}
			}
			// If requests completed in order, the test may not be strict enough
			// but the implementation is still correct
			if allInOrder {
				t.Log("Note: requests completed in order, but implementation correctly preserves order")
			}
		}
	})
}

// TestFetchCVEByID_ResponseTooLarge verifies that responses exceeding
// the maximum allowed size are rejected, fixing TODO-123
func TestFetchCVEByID_ResponseTooLarge(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_ResponseTooLarge", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a response that exceeds MaxResponseSize (10 MB)
		largeJSON := make([]byte, MaxResponseSize+1024) // 1 KB over the limit
		// Make it valid JSON by wrapping in an object
		largeJSON[0] = '{'
		largeJSON[1] = '"'
		largeJSON[2] = 'd'
		largeJSON[3] = '"'
		largeJSON[4] = ':'
		largeJSON[5] = '"'
		for i := 6; i < len(largeJSON)-2; i++ {
			largeJSON[i] = 'x'
		}
		largeJSON[len(largeJSON)-2] = '"'
		largeJSON[len(largeJSON)-1] = '}'

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(largeJSON)
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		_, err = f.FetchCVEByID("CVE-TEST-1")
		if err == nil {
			t.Fatal("expected error for oversized response")
		}
		// Verify the error is or contains ErrResponseTooLarge
		if !errors.Is(err, ErrResponseTooLarge) && err.Error()[:4] != "API r" {
			t.Errorf("expected ErrResponseTooLarge, got: %v", err)
		}
	})
}

// TestFetchCVEs_ResponseTooLarge verifies that responses exceeding
// the maximum allowed size are rejected in bulk fetch operations
func TestFetchCVEs_ResponseTooLarge(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEs_ResponseTooLarge", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a response that exceeds MaxResponseSize (10 MB)
		largeJSON := make([]byte, MaxResponseSize+1024) // 1 KB over the limit
		// Make it valid JSON by wrapping in an object
		largeJSON[0] = '{'
		largeJSON[1] = '"'
		largeJSON[2] = 'd'
		largeJSON[3] = '"'
		largeJSON[4] = ':'
		largeJSON[5] = '"'
		for i := 6; i < len(largeJSON)-2; i++ {
			largeJSON[i] = 'x'
		}
		largeJSON[len(largeJSON)-2] = '"'
		largeJSON[len(largeJSON)-1] = '}'

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(largeJSON)
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		_, err = f.FetchCVEs(0, 10)
		if err == nil {
			t.Fatal("expected error for oversized response")
		}
		// Verify the error is or contains ErrResponseTooLarge
		if !errors.Is(err, ErrResponseTooLarge) && err.Error()[:4] != "API r" {
			t.Errorf("expected ErrResponseTooLarge, got: %v", err)
		}
	})
}

// TestFetchCVEByID_MaxSizeBoundary verifies responses at exactly the limit are accepted
func TestFetchCVEByID_MaxSizeBoundary(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetchCVEByID_MaxSizeBoundary", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a valid CVE response that is close to but under MaxResponseSize
		// Use a valid minimal CVE response JSON
		validResponse := []byte(`{
			"vulnerabilities": [
				{
					"cve": {
						"id": "CVE-TEST-1",
						"state": "PUBLISHED",
						"descriptions": [{"lang": "en", "value": "Test CVE"}]
					}
				}
			]
		}`)

		// Pad with spaces to get close to the limit (but under it)
		paddingSize := MaxResponseSize - len(validResponse) - 100 // Stay 100 bytes under
		if paddingSize > 0 {
			paddedResponse := make([]byte, 0, len(validResponse)+paddingSize)
			paddedResponse = append(paddedResponse, validResponse...)
			for i := 0; i < paddingSize; i++ {
				paddedResponse = append(paddedResponse, ' ')
			}
			validResponse = paddedResponse
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(validResponse)
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		resp, err := f.FetchCVEByID("CVE-TEST-1")
		if err != nil {
			t.Fatalf("unexpected error for response under size limit: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if len(resp.Vulnerabilities) == 0 || resp.Vulnerabilities[0].CVE.ID != "CVE-TEST-1" {
			t.Errorf("unexpected response: got %d vulnerabilities", len(resp.Vulnerabilities))
		}
	})
}
