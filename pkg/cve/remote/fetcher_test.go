package remote

import (
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
