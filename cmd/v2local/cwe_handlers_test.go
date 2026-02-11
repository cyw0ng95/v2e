package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestCWEHandlers(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCWEHandlers", nil, func(t *testing.T, tx *gorm.DB) {
		// Use t.TempDir() for cleaner cleanup
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe-test.db")
		jsonPath := filepath.Join(tempDir, "cwe-import.json")

		var buf bytes.Buffer
		logger := common.NewLogger(&buf, "", common.DebugLevel)

		store, err := cwe.NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("NewLocalCWEStore error: %v", err)
		}

		// create a small JSON file for import
		sample := []map[string]interface{}{{"ID": "CWE-1", "Name": "Test CWE"}}
		data, err := subprocess.MarshalFast(sample)
		if err != nil {
			t.Fatalf("marshal sample: %v", err)
		}
		if err := os.WriteFile(jsonPath, data, 0600); err != nil {
			t.Fatalf("write sample: %v", err)
		}

		ctx := context.Background()

		// Import
		importH := createImportCWEsHandler(store, logger)
		importReq := map[string]interface{}{"path": jsonPath}
		impResp, err := importH(ctx, &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "imp1", Payload: func() []byte { b, _ := subprocess.MarshalFast(importReq); return b }()})
		if err != nil || impResp == nil || impResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("import handler failed: err=%v resp=%v", err, impResp)
		}

		// Get by ID
		getH := createGetCWEByIDHandler(store, logger)
		getReq := map[string]interface{}{"cwe_id": "CWE-1"}
		getResp, err := getH(ctx, &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "g1", Payload: func() []byte { b, _ := subprocess.MarshalFast(getReq); return b }()})
		if err != nil || getResp == nil || getResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("get handler failed: err=%v resp=%v", err, getResp)
		}
		var got cwe.CWEItem
		if err := subprocess.UnmarshalPayload(getResp, &got); err != nil {
			t.Fatalf("unmarshal get cwe: %v", err)
		}
		if got.ID != "CWE-1" {
			t.Fatalf("expected CWE-1 got %s", got.ID)
		}

		// List
		listH := createListCWEsHandler(store, logger)
		listResp, err := listH(ctx, &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "l1"})
		if err != nil || listResp == nil || listResp.Type != subprocess.MessageTypeResponse {
			t.Fatalf("list handler failed: err=%v resp=%v", err, listResp)
		}
		var listResult map[string]interface{}
		if err := subprocess.UnmarshalPayload(listResp, &listResult); err != nil {
			t.Fatalf("unmarshal list result: %v", err)
		}
		if total, ok := listResult["total"].(float64); !ok || total < 1 {
			t.Fatalf("expected total >=1 got %v", listResult["total"])
		}
	})

}
