package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestCreateFetchViewsHandler_Success(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateFetchViewsHandler_Success", nil, func(t *testing.T, tx *gorm.DB) {
		entries := map[string][]byte{
			"REST-API-wg-main/json_repo/V/test.json": []byte(`{"ID":"VIEW-1","Name":"Test View"}`),
		}
		zipBytes, err := testutils.MakeZip(entries)
		if err != nil {
			t.Fatalf("MakeZip failed: %v", err)
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(zipBytes)
		}))
		defer server.Close()
		old := os.Getenv("VIEW_FETCH_URL")
		os.Setenv("VIEW_FETCH_URL", server.URL)
		defer os.Setenv("VIEW_FETCH_URL", old)
		h := createFetchViewsHandler()
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "1", Source: "test"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("expected response, got %v", resp.Type)
		}
		var payload map[string][]map[string]interface{}
		if err := json.Unmarshal(resp.Payload, &payload); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		views := payload["views"]
		if len(views) != 1 {
			t.Fatalf("expected 1 view, got %d", len(views))
		}
	})

}

func TestCreateFetchViewsHandler_HTTPError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateFetchViewsHandler_HTTPError", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer server.Close()
		old := os.Getenv("VIEW_FETCH_URL")
		os.Setenv("VIEW_FETCH_URL", server.URL)
		defer os.Setenv("VIEW_FETCH_URL", old)
		h := createFetchViewsHandler()
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "1", Source: "test"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError {
			t.Fatalf("expected error, got %v", resp.Type)
		}
	})

}

func TestCreateFetchViewsHandler_MalformedZip(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCreateFetchViewsHandler_MalformedZip", nil, func(t *testing.T, tx *gorm.DB) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("not a zip"))
		}))
		defer server.Close()
		old := os.Getenv("VIEW_FETCH_URL")
		os.Setenv("VIEW_FETCH_URL", server.URL)
		defer os.Setenv("VIEW_FETCH_URL", old)
		h := createFetchViewsHandler()
		msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "1", Source: "test"}
		resp, err := h(context.Background(), msg)
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}
		if resp.Type != subprocess.MessageTypeError {
			t.Fatalf("expected error, got %v", resp.Type)
		}
	})

}
