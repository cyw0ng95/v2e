package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func makeMsgWithPayload(t *testing.T, payload interface{}) *subprocess.Message {
	t.Helper()
	data, err := subprocess.MarshalFast(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "1",
		Payload: data,
		Source:  "test",
		Target:  "local",
	}
}

func TestCVEHandlers(t *testing.T) {
	// temp sqlite DB
	f, err := os.CreateTemp("", "cve-test-*.db")
	if err != nil {
		t.Fatalf("temp db: %v", err)
	}
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	// logger with buffer
	var buf bytes.Buffer
	logger := common.NewLogger(&buf, "", common.DebugLevel)

	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	defer db.Close()

	// Handlers
	saveH := createSaveCVEByIDHandler(db, logger)
	isStoredH := createIsCVEStoredByIDHandler(db, logger)
	getH := createGetCVEByIDHandler(db, logger)
	deleteH := createDeleteCVEByIDHandler(db, logger)
	listH := createListCVEsHandler(db, logger)
	countH := createCountCVEsHandler(db, logger)

	ctx := context.Background()

	// Create minimal CVE
	item := cve.CVEItem{
		ID:           "CVE-TEST-1",
		Descriptions: []cve.Description{{Lang: "en", Value: "test"}},
	}

	// Save
	saveReq := map[string]interface{}{"cve": item}
	resp, err := saveH(ctx, makeMsgWithPayload(t, saveReq))
	if err != nil || resp == nil || resp.Type != subprocess.MessageTypeResponse {
		t.Fatalf("save handler failed: err=%v resp=%v", err, resp)
	}
	var saveResult map[string]interface{}
	if err := subprocess.UnmarshalPayload(resp, &saveResult); err != nil {
		t.Fatalf("unmarshal save result: %v", err)
	}

	// IsStored
	isReq := map[string]interface{}{"cve_id": item.ID}
	isResp, err := isStoredH(ctx, makeMsgWithPayload(t, isReq))
	if err != nil || isResp == nil || isResp.Type != subprocess.MessageTypeResponse {
		t.Fatalf("isStored handler failed: err=%v resp=%v", err, isResp)
	}
	var isResult map[string]interface{}
	if err := subprocess.UnmarshalPayload(isResp, &isResult); err != nil {
		t.Fatalf("unmarshal isStored result: %v", err)
	}
	if stored, ok := isResult["stored"].(bool); !ok || !stored {
		t.Fatalf("expected stored=true, got: %v", isResult)
	}

	// Get
	getReq := map[string]interface{}{"cve_id": item.ID}
	getResp, err := getH(ctx, makeMsgWithPayload(t, getReq))
	if err != nil || getResp == nil || getResp.Type != subprocess.MessageTypeResponse {
		t.Fatalf("get handler failed: err=%v resp=%v", err, getResp)
	}
	var got cve.CVEItem
	if err := subprocess.UnmarshalPayload(getResp, &got); err != nil {
		t.Fatalf("unmarshal get result: %v", err)
	}
	if got.ID != item.ID {
		t.Fatalf("expected id %s got %s", item.ID, got.ID)
	}

	// Count
	countResp, err := countH(ctx, &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "c2"})
	if err != nil || countResp == nil || countResp.Type != subprocess.MessageTypeResponse {
		t.Fatalf("count handler failed: err=%v resp=%v", err, countResp)
	}
	var countRes map[string]interface{}
	if err := subprocess.UnmarshalPayload(countResp, &countRes); err != nil {
		t.Fatalf("unmarshal count result: %v", err)
	}

	// List
	listResp, err := listH(ctx, &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "l1"})
	if err != nil || listResp == nil || listResp.Type != subprocess.MessageTypeResponse {
		t.Fatalf("list handler failed: err=%v resp=%v", err, listResp)
	}
	var listRes map[string]interface{}
	if err := subprocess.UnmarshalPayload(listResp, &listRes); err != nil {
		t.Fatalf("unmarshal list result: %v", err)
	}

	// Delete
	delReq := map[string]interface{}{"cve_id": item.ID}
	delResp, err := deleteH(ctx, makeMsgWithPayload(t, delReq))
	if err != nil || delResp == nil || delResp.Type != subprocess.MessageTypeResponse {
		t.Fatalf("delete handler failed: err=%v resp=%v", err, delResp)
	}

	// Confirm deleted
	isResp2, err := isStoredH(ctx, makeMsgWithPayload(t, isReq))
	if err != nil || isResp2 == nil || isResp2.Type != subprocess.MessageTypeResponse {
		t.Fatalf("isStored after delete failed: err=%v resp=%v", err, isResp2)
	}
	var isResult2 map[string]interface{}
	if err := subprocess.UnmarshalPayload(isResp2, &isResult2); err != nil {
		t.Fatalf("unmarshal isStored result after delete: %v", err)
	}
	if stored, ok := isResult2["stored"].(bool); !ok || stored {
		t.Fatalf("expected stored=false after delete, got: %v", isResult2)
	}

	_ = saveResult
	_ = countRes
	_ = listRes
}
