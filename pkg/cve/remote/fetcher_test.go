package remote

import (
"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestFetchCVEByID_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestFetchCVEByID_Success", nil, func(t *testing.T, tx *gorm.DB) {
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

		f := NewFetcher("")
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
	testutils.Run(t, testutils.Level1, "TestFetchCVEByID_RateLimited", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(429)
		}))
		defer server.Close()

		f := NewFetcher("")
		f.baseURL = server.URL
		_, err := f.FetchCVEByID("CVE-TEST-2")
		if err == nil || err != ErrRateLimited {
			t.Fatalf("expected ErrRateLimited, got %v", err)
		}
	})

}

func TestFetchCVEs_ParamValidation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestFetchCVEs_ParamValidation", nil, func(t *testing.T, tx *gorm.DB) {
		f := NewFetcher("")
		if _, err := f.FetchCVEs(-1, 10); err == nil {
			t.Fatalf("expected error for negative startIndex")
		}
		if _, err := f.FetchCVEs(0, 0); err == nil {
			t.Fatalf("expected error for invalid resultsPerPage")
		}
	})

}

func TestFetchCVEsConcurrent_Workers(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestFetchCVEsConcurrent_Workers", nil, func(t *testing.T, tx *gorm.DB) {
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

		f := NewFetcher("")
		f.baseURL = server.URL
		ids := []string{"CVE-A", "CVE-B", "CVE-C"}
		resps, errs := f.FetchCVEsConcurrent(ids, 3)
		if len(resps)+len(errs) != len(ids) {
			t.Fatalf("expected responses+errors == %d, got %d+%d", len(ids), len(resps), len(errs))
		}
	})

}
