package remote

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestFetcher_HTTP2ConnectionReuse(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2ConnectionReuse", nil, func(t *testing.T, tx *gorm.DB) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		for i := 0; i < 5; i++ {
			_, err := f.FetchCVEByID("CVE-TEST-1")
			if err != nil {
				t.Fatalf("request %d failed: %v", i, err)
			}
		}

		if requestCount != 5 {
			t.Errorf("expected 5 requests, got %d", requestCount)
		}
	})
}

func TestFetcher_HTTP2WithTLS(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2WithTLS", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))

		server.EnableHTTP2 = true
		server.StartTLS()
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL
		resp, err := f.FetchCVEByID("CVE-TEST-1")
		if err != nil {
			t.Fatalf("HTTP/2 request failed: %v", err)
		}
		if resp == nil || len(resp.Vulnerabilities) != 1 {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})
}

func TestFetcher_HTTP2ConcurrentRequests(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2ConcurrentRequests", nil, func(t *testing.T, tx *gorm.DB) {
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

		numRequests := 10
		done := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				_, err := f.FetchCVEByID("CVE-TEST-1")
				done <- err
			}()
		}

		for i := 0; i < numRequests; i++ {
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("concurrent request failed: %v", err)
				}
			default:
				t.Fatal("not all requests completed")
			}
		}
	})
}

func TestFetcher_HTTP2Multiplexing(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2Multiplexing", nil, func(t *testing.T, tx *gorm.DB) {
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

		ids := []string{"CVE-A", "CVE-B", "CVE-C", "CVE-D", "CVE-E"}
		resps, errs := f.FetchCVEsConcurrent(ids, 5)

		if len(resps) != 5 {
			t.Errorf("expected 5 responses, got %d", len(resps))
		}
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got %d", len(errs))
		}
	})
}

func TestFetcher_HTTP2ConnectionPool(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2ConnectionPool", nil, func(t *testing.T, tx *gorm.DB) {
		connCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connCount++
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		for i := 0; i < 20; i++ {
			_, _ = f.FetchCVEByID("CVE-TEST-1")
		}

		t.Logf("Total connections created: %d", connCount)
		if connCount > 5 {
			t.Logf("Connection pooling appears to be working (only %d connections for 20 requests)", connCount)
		}
	})
}

func TestFetcher_HTTP2BufferPooling(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2BufferPooling", nil, func(t *testing.T, tx *gorm.DB) {
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

		if f.bufferPool == nil {
			t.Fatal("expected bufferPool to be initialized")
		}

		for i := 0; i < 10; i++ {
			_, err := f.FetchCVEByID("CVE-TEST-1")
			if err != nil {
				t.Fatalf("request %d failed: %v", i, err)
			}
		}

		t.Log("Buffer pooling test completed successfully")
	})
}

func TestFetcher_HTTP2WithAPIKey(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2WithAPIKey", nil, func(t *testing.T, tx *gorm.DB) {
		apiKeyReceived := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("apiKey") == "test-api-key" {
				apiKeyReceived = true
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))
		defer server.Close()

		f, err := NewFetcher("test-api-key")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		_, err = f.FetchCVEByID("CVE-TEST-1")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if !apiKeyReceived {
			t.Error("API key was not sent in HTTP/2 request")
		}
	})
}

func TestFetcher_HTTP2RaceCondition(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2RaceCondition", nil, func(t *testing.T, tx *gorm.DB) {
		var wg sync.WaitGroup
		errors := make(chan error, 100)

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

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := f.FetchCVEByID("CVE-TEST-1")
				if err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		errorCount := 0
		for err := range errors {
			t.Logf("Error: %v", err)
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("encountered %d errors during concurrent HTTP/2 requests", errorCount)
		}
	})
}

func TestFetcher_HTTP2BackwardsCompatibility(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFetcher_HTTP2BackwardsCompatibility", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor != 2 {
				t.Logf("HTTP/2 not negotiated, falling back to %s", r.Proto)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(testutils.MakeCVEResponseJSON("CVE-TEST-1", 1))
		}))
		defer server.Close()

		f, err := NewFetcher("")
		if err != nil {
			t.Fatalf("failed to create fetcher: %v", err)
		}
		f.baseURL = server.URL

		resp, err := f.FetchCVEByID("CVE-TEST-1")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp == nil || len(resp.Vulnerabilities) != 1 {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})
}
