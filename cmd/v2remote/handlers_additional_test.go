package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/go-resty/resty/v2"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// newTestFetcher returns a fetcher configured to talk to serverURL by rewriting
// the internal resty client and baseURL via reflection (test-only shim).
func newTestFetcher(serverURL string) *remote.Fetcher {
	f, _ := remote.NewFetcher("")
	if serverURL == "" {
		return f
	}
	client := resty.New()
	client.SetBaseURL(serverURL)

	v := reflect.ValueOf(f).Elem()
	baseURL := v.FieldByName("baseURL")
	reflect.NewAt(baseURL.Type(), unsafe.Pointer(baseURL.UnsafeAddr())).Elem().SetString(serverURL)
	clientField := v.FieldByName("client")
	reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(client))
	return f
}

func TestCreateGetCVEByIDHandler_Validation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCVEByIDHandler_Validation", nil, func(t *testing.T, tx *gorm.DB) {
		fetcher := newTestFetcher("")
		h := createGetCVEByIDHandler(fetcher)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "1"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error == "" {
			t.Fatalf("expected parse error, got %+v", resp)
		}

		payload, _ := subprocess.MarshalFast(map[string]string{"cve_id": ""})
		msg = &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "2", Payload: payload}
		resp, err = h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error != "cve_id is required" {
			t.Fatalf("expected validation error, got %+v", resp)
		}
	})

}

func TestCreateGetCVEByIDHandler_Success(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCVEByIDHandler_Success", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("cveId") != "CVE-123" {
				t.Fatalf("unexpected cveId query: %s", r.URL.RawQuery)
			}
			resp := cve.CVEResponse{TotalResults: 1, ResultsPerPage: 1, StartIndex: 0, Version: "2.0"}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher := newTestFetcher(server.URL)
		h := createGetCVEByIDHandler(fetcher)

		payload, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-123"})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "1", Payload: payload}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %+v", resp)
		}
	})

}

func TestCreateGetCVEByIDHandler_RateLimited(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCVEByIDHandler_RateLimited", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()
		fetcher := newTestFetcher(server.URL)
		h := createGetCVEByIDHandler(fetcher)

		payload, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-1"})
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "1", Payload: payload}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Error != "NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)" {
			t.Fatalf("expected rate limited error, got %+v", resp)
		}
	})

}

func TestCreateGetCVECntHandler_DefaultsAndSuccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCVECntHandler_DefaultsAndSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("startIndex") != "0" || r.URL.Query().Get("resultsPerPage") != "1" {
				t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
			}
			resp := cve.CVEResponse{TotalResults: 42, ResultsPerPage: 1, StartIndex: 0}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher := newTestFetcher(server.URL)
		h := createGetCVECntHandler(fetcher)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "cnt"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %+v", resp)
		}
		var decoded map[string]any
		if err := json.Unmarshal(resp.Payload, &decoded); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		totalResults, ok := decoded["total_results"].(float64)
		if !ok || totalResults != 42 {
			t.Fatalf("unexpected total_results: %+v", decoded)
		}
	})

}

func TestCreateGetCVECntHandler_RateLimited(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateGetCVECntHandler_RateLimited", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()
		fetcher := newTestFetcher(server.URL)
		h := createGetCVECntHandler(fetcher)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "cnt"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Error != "NVD_RATE_LIMITED: NVD API rate limit exceeded (HTTP 429)" {
			t.Fatalf("expected rate limited error, got %+v", resp)
		}
	})

}

func TestCreateFetchCVEsHandler_Defaults(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateFetchCVEsHandler_Defaults", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("startIndex") != "0" || r.URL.Query().Get("resultsPerPage") != "100" {
				t.Fatalf("unexpected defaults: %s", r.URL.RawQuery)
			}
			resp := cve.CVEResponse{TotalResults: 10, ResultsPerPage: 100, StartIndex: 0}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		fetcher := newTestFetcher(server.URL)
		h := createFetchCVEsHandler(fetcher)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "fetch"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %+v", resp)
		}
	})

}

func TestCreateFetchCVEsHandler_Error(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateFetchCVEsHandler_Error", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		fetcher := newTestFetcher(server.URL)
		h := createFetchCVEsHandler(fetcher)

		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "fetch"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler returned error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError || resp.Error == "" {
			t.Fatalf("expected error response, got %+v", resp)
		}
	})

}
