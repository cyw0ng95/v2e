package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// TestCVEHandlers_DatabaseOperations tests complete CRUD operations on CVE database
func TestCVEHandlers_DatabaseOperations(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCVEHandlers_DatabaseOperations", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cve-integration-test.db")

		logger := common.NewLogger(nil, "[TEST] ", common.ErrorLevel)

		db, err := local.NewDB(dbPath)
		if err != nil {
			t.Fatalf("NewDB error: %v", err)
		}
		defer db.Close()

		ctx := context.Background()

		// Test: Create multiple CVEs
		t.Run("CreateMultipleCVEs", func(t *testing.T) {
			createH := createCreateCVEHandler(db, logger)

			cves := []cve.CVEItem{
				{ID: "CVE-2024-1001", Descriptions: []cve.Description{{Lang: "en", Value: "Test 1"}}},
				{ID: "CVE-2024-1002", Descriptions: []cve.Description{{Lang: "en", Value: "Test 2"}}},
				{ID: "CVE-2024-1003", Descriptions: []cve.Description{{Lang: "en", Value: "Test 3"}}},
			}

			for _, item := range cves {
				payload, _ := subprocess.MarshalFast(item)
				msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "create-" + item.ID, Payload: payload}
				resp, err := createH(ctx, msg)
				if err != nil || resp == nil || resp.Type != subprocess.MessageTypeResponse {
					t.Fatalf("failed to create CVE %s: err=%v resp=%v", item.ID, err, resp)
				}
			}

			// Verify count
			countH := createCountCVEsHandler(db, logger)
			countMsg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "count"}
			resp, err := countH(ctx, countMsg)
			if err != nil {
				t.Fatalf("count failed: %v", err)
			}
			var result map[string]interface{}
			subprocess.UnmarshalPayload(resp, &result)
			if count := int(result["count"].(float64)); count != 3 {
				t.Fatalf("expected count=3, got %d", count)
			}
		})

		// Test: Update existing CVE
		t.Run("UpdateCVE", func(t *testing.T) {
			updateH := createUpdateCVEHandler(db, logger)

			updated := cve.CVEItem{
				ID: "CVE-2024-1001",
				Descriptions: []cve.Description{{Lang: "en", Value: "Updated description"}},
			}
			payload, _ := subprocess.MarshalFast(updated)
			msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "update", Payload: payload}
			resp, err := updateH(ctx, msg)
			if err != nil || resp == nil || resp.Type != subprocess.MessageTypeResponse {
				t.Fatalf("update failed: err=%v resp=%v", err, resp)
			}

			// Verify update by checking the response
			var result map[string]interface{}
			subprocess.UnmarshalPayload(resp, &result)
			if result["cve_id"] != "CVE-2024-1001" {
				t.Fatalf("update response missing cve_id")
			}
		})

		// Test: Pagination with list
		t.Run("ListPagination", func(t *testing.T) {
			listH := createListCVEsHandler(db, logger)

			// Get first page
			payload1, _ := subprocess.MarshalFast(map[string]int{"offset": 0, "limit": 2})
			msg1 := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "list1", Payload: payload1}
			resp1, err := listH(ctx, msg1)
			if err != nil {
				t.Fatalf("list page 1 failed: %v", err)
			}
			var result1 map[string]interface{}
			subprocess.UnmarshalPayload(resp1, &result1)
			cves1 := result1["cves"].([]interface{})
			if len(cves1) != 2 {
				t.Fatalf("expected 2 items on page 1, got %d", len(cves1))
			}

			// Get second page
			payload2, _ := subprocess.MarshalFast(map[string]int{"offset": 2, "limit": 2})
			msg2 := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "list2", Payload: payload2}
			resp2, err := listH(ctx, msg2)
			if err != nil {
				t.Fatalf("list page 2 failed: %v", err)
			}
			var result2 map[string]interface{}
			subprocess.UnmarshalPayload(resp2, &result2)
			cves2 := result2["cves"].([]interface{})
			if len(cves2) != 1 {
				t.Fatalf("expected 1 item on page 2, got %d", len(cves2))
			}
		})

		// Test: IsStored for non-existent CVE
		t.Run("IsStored_NotFound", func(t *testing.T) {
			isStoredH := createIsCVEStoredByIDHandler(db, logger)
			payload, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-9999-9999"})
			msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "isstored", Payload: payload}
			resp, err := isStoredH(ctx, msg)
			if err != nil {
				t.Fatalf("isStored failed: %v", err)
			}
			var result map[string]interface{}
			subprocess.UnmarshalPayload(resp, &result)
			if stored, ok := result["stored"].(bool); !ok || stored {
				t.Fatalf("expected stored=false for non-existent CVE, got %v", result)
			}
		})

		// Test: Delete multiple CVEs
		t.Run("DeleteMultipleCVEs", func(t *testing.T) {
			deleteH := createDeleteCVEHandler(db, logger)

			// Delete first CVE
			payload1, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-2024-1001"})
			msg1 := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "del1", Payload: payload1}
			resp1, err := deleteH(ctx, msg1)
			if err != nil || resp1 == nil || resp1.Type != subprocess.MessageTypeResponse {
				t.Fatalf("delete CVE-2024-1001 failed: err=%v resp=%v", err, resp1)
			}

			// Delete second CVE
			payload2, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-2024-1002"})
			msg2 := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "del2", Payload: payload2}
			resp2, err := deleteH(ctx, msg2)
			if err != nil || resp2 == nil || resp2.Type != subprocess.MessageTypeResponse {
				t.Fatalf("delete CVE-2024-1002 failed: err=%v resp=%v", err, resp2)
			}

			// Verify count decreased
			countH := createCountCVEsHandler(db, logger)
			countMsg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "count"}
			resp, err := countH(ctx, countMsg)
			if err != nil {
				t.Fatalf("count after deletes failed: %v", err)
			}
			var result map[string]interface{}
			subprocess.UnmarshalPayload(resp, &result)
			if count := int(result["count"].(float64)); count != 1 {
				t.Fatalf("expected count=1 after 2 deletes, got %d", count)
			}
		})

		// Test: Get deleted CVE returns error
		t.Run("GetDeletedCVE_ReturnsError", func(t *testing.T) {
			getH := createGetCVEByIDHandler(db, logger)
			payload, _ := subprocess.MarshalFast(map[string]string{"cve_id": "CVE-2024-1001"})
			msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "get", Payload: payload}
			resp, err := getH(ctx, msg)
			if err != nil {
				t.Fatalf("get deleted CVE failed: %v", err)
			}
			if resp.Type != subprocess.MessageTypeError {
				t.Fatalf("expected error response for deleted CVE, got type %v", resp.Type)
			}
		})
	})
}

// TestCVEHandlers_EdgeCases tests edge cases and error conditions
func TestCVEHandlers_EdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCVEHandlers_EdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cve-edge-test.db")

		logger := common.NewLogger(nil, "[TEST] ", common.ErrorLevel)

		db, err := local.NewDB(dbPath)
		if err != nil {
			t.Fatalf("NewDB error: %v", err)
		}
		defer db.Close()

		ctx := context.Background()

		// Test: Create CVE with empty ID
		t.Run("CreateCVE_EmptyID", func(t *testing.T) {
			createH := createCreateCVEHandler(db, logger)
			item := cve.CVEItem{Descriptions: []cve.Description{{Lang: "en", Value: "Test"}}}
			payload, _ := subprocess.MarshalFast(item)
			msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "create", Payload: payload}
			resp, err := createH(ctx, msg)
			if err != nil {
				t.Fatalf("handler returned error: %v", err)
			}
			if resp.Type != subprocess.MessageTypeError {
				t.Fatalf("expected error for empty CVE ID, got type %v", resp.Type)
			}
		})

		// Test: Create duplicate CVE
		t.Run("CreateDuplicateCVE", func(t *testing.T) {
			createH := createCreateCVEHandler(db, logger)
			item := cve.CVEItem{ID: "CVE-DUP-1", Descriptions: []cve.Description{{Lang: "en", Value: "Test"}}}

			// First create
			payload1, _ := subprocess.MarshalFast(item)
			msg1 := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "create1", Payload: payload1}
			resp1, err := createH(ctx, msg1)
			if err != nil || resp1 == nil || resp1.Type != subprocess.MessageTypeResponse {
				t.Fatalf("first create failed: err=%v resp=%v", err, resp1)
			}

			// Duplicate create (should succeed with update)
			item.Descriptions[0].Value = "Updated"
			payload2, _ := subprocess.MarshalFast(item)
			msg2 := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "create2", Payload: payload2}
			resp2, err := createH(ctx, msg2)
			if err != nil {
				t.Fatalf("duplicate create handler error: %v", err)
			}
			// Duplicate should update existing record
			if resp2.Type != subprocess.MessageTypeResponse {
				t.Logf("duplicate create response type: %v, error: %v", resp2.Type, resp2.Error)
			}
		})

		// Test: List with large offset
		t.Run("List_LargeOffset", func(t *testing.T) {
			listH := createListCVEsHandler(db, logger)
			payload, _ := subprocess.MarshalFast(map[string]int{"offset": 1000, "limit": 10})
			msg := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "list", Payload: payload}
			resp, err := listH(ctx, msg)
			if err != nil {
				t.Fatalf("list with large offset failed: %v", err)
			}
			var result map[string]interface{}
			subprocess.UnmarshalPayload(resp, &result)
			cves := result["cves"].([]interface{})
			if len(cves) != 0 {
				t.Fatalf("expected empty list for large offset, got %d items", len(cves))
			}
		})
	})
}
