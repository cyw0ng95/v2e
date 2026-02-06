package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/asvs"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestASVSHandlers(t *testing.T) {
	// Create temp sqlite DB
	f, err := os.CreateTemp("", "asvs-test-*.db")
	if err != nil {
		t.Fatalf("temp db: %v", err)
	}
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	var buf bytes.Buffer
	logger := common.NewLogger(&buf, "", common.DebugLevel)

	store, err := asvs.NewLocalASVSStore(dbPath)
	if err != nil {
		t.Fatalf("NewLocalASVSStore error: %v", err)
	}

	ctx := context.Background()

	// Test ImportASVS handler with a mock CSV server
	t.Run("ImportASVS", func(t *testing.T) {
		// Create a mock HTTP server that serves a test CSV
		csvContent := `Requirement ID,Chapter,Section,Description,L1,L2,L3,CWE
1.1.1,V1,Architecture,Test requirement 1,x,x,x,CWE-1127
1.1.2,V1,Architecture,Test requirement 2,,x,x,CWE-79
2.1.1,V2,Authentication,Test requirement 3,x,x,,CWE-521
`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(csvContent))
		}))
		defer server.Close()

		importH := createImportASVSHandler(store, logger)
		importReq := map[string]interface{}{"url": server.URL}
		impResp, err := importH(ctx, &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "imp1",
			Payload: func() []byte { b, _ := subprocess.MarshalFast(importReq); return b }(),
		})
		if err != nil || impResp == nil || impResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("import handler failed: err=%v resp=%v", err, impResp)
		}

		var importResult map[string]interface{}
		if err := subprocess.UnmarshalPayload(impResp, &importResult); err != nil {
			t.Fatalf("unmarshal import result: %v", err)
		}
		if success, ok := importResult["success"].(bool); !ok || !success {
			t.Fatalf("expected success=true got %v", importResult["success"])
		}
	})

	// Test GetASVSByID handler
	t.Run("GetASVSByID", func(t *testing.T) {
		getH := createGetASVSByIDHandler(store, logger)
		getReq := map[string]interface{}{"requirement_id": "1.1.1"}
		getResp, err := getH(ctx, &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "g1",
			Payload: func() []byte { b, _ := subprocess.MarshalFast(getReq); return b }(),
		})
		if err != nil || getResp == nil || getResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("get handler failed: err=%v resp=%v", err, getResp)
		}

		var got asvs.ASVSRequirement
		if err := subprocess.UnmarshalPayload(getResp, &got); err != nil {
			t.Fatalf("unmarshal get asvs: %v", err)
		}
		if got.RequirementID != "1.1.1" {
			t.Fatalf("expected requirement ID 1.1.1 got %s", got.RequirementID)
		}
		if got.Chapter != "V1" {
			t.Fatalf("expected chapter V1 got %s", got.Chapter)
		}
		if !got.Level1 || !got.Level2 || !got.Level3 {
			t.Fatalf("expected all levels to be true")
		}
	})

	// Test ListASVS handler
	t.Run("ListASVS", func(t *testing.T) {
		listH := createListASVSHandler(store, logger)
		listResp, err := listH(ctx, &subprocess.Message{
			Type: subprocess.MessageTypeRequest,
			ID:   "l1",
		})
		if err != nil || listResp == nil || listResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("list handler failed: err=%v resp=%v", err, listResp)
		}

		var listResult map[string]interface{}
		if err := subprocess.UnmarshalPayload(listResp, &listResult); err != nil {
			t.Fatalf("unmarshal list result: %v", err)
		}
		if total, ok := listResult["total"].(float64); !ok || total < 3 {
			t.Fatalf("expected total >=3 got %v", listResult["total"])
		}
	})

	// Test ListASVS with filters
	t.Run("ListASVS_WithFilters", func(t *testing.T) {
		listH := createListASVSHandler(store, logger)

		// Filter by chapter
		listReq := map[string]interface{}{"chapter": "V1", "offset": 0, "limit": 10}
		listResp, err := listH(ctx, &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "l2",
			Payload: func() []byte { b, _ := subprocess.MarshalFast(listReq); return b }(),
		})
		if err != nil || listResp == nil || listResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("list handler with chapter filter failed: err=%v resp=%v", err, listResp)
		}

		var listResult map[string]interface{}
		if err := subprocess.UnmarshalPayload(listResp, &listResult); err != nil {
			t.Fatalf("unmarshal list result: %v", err)
		}
		if total, ok := listResult["total"].(float64); !ok || total != 2 {
			t.Fatalf("expected total=2 for V1 got %v", listResult["total"])
		}

		// Filter by level
		listReq = map[string]interface{}{"level": 1, "offset": 0, "limit": 10}
		listResp, err = listH(ctx, &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "l3",
			Payload: func() []byte { b, _ := subprocess.MarshalFast(listReq); return b }(),
		})
		if err != nil || listResp == nil || listResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("list handler with level filter failed: err=%v resp=%v", err, listResp)
		}

		if err := subprocess.UnmarshalPayload(listResp, &listResult); err != nil {
			t.Fatalf("unmarshal list result: %v", err)
		}
		if total, ok := listResult["total"].(float64); !ok || total != 2 {
			t.Fatalf("expected total=2 for level 1 got %v", listResult["total"])
		}
	})

	// Test error cases
	t.Run("GetASVSByID_NotFound", func(t *testing.T) {
		getH := createGetASVSByIDHandler(store, logger)
		getReq := map[string]interface{}{"requirement_id": "999.999.999"}
		getResp, err := getH(ctx, &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "g2",
			Payload: func() []byte { b, _ := subprocess.MarshalFast(getReq); return b }(),
		})
		if err != nil || getResp == nil || getResp.Type != subprocess.MessageTypeError {
			t.Fatalf("expected error for non-existent requirement")
		}
	})

	t.Run("ImportASVS_InvalidURL", func(t *testing.T) {
		importH := createImportASVSHandler(store, logger)
		importReq := map[string]interface{}{"url": "http://invalid-url-that-does-not-exist-12345.com/test.csv"}
		impResp, err := importH(ctx, &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "imp2",
			Payload: func() []byte { b, _ := subprocess.MarshalFast(importReq); return b }(),
		})
		if err != nil || impResp == nil || impResp.Type != subprocess.MessageTypeError {
			t.Fatalf("expected error for invalid URL")
		}
	})
}
