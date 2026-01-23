package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/local"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRPCSaveCVEByID(t *testing.T) {
	// Create a temporary database for testing
	dbPath := "/tmp/test_cve_local_save.db"
	defer os.Remove(dbPath)

	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Create handler
	handler := createSaveCVEByIDHandler(db, logger)

	// Create test CVE data
	testCVE := cve.CVEItem{
		ID:           "CVE-2021-TEST",
		SourceID:     "test@example.com",
		VulnStatus:   "Test",
		Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE"}},
	}

	// Create request message
	payload, _ := sonic.Marshal(map[string]interface{}{
		"cve": testCVE,
	})

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCSaveCVEByID",
		Payload: payload,
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// Check results
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	var result map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}

	if result["cve_id"] != "CVE-2021-TEST" {
		t.Errorf("Expected cve_id=CVE-2021-TEST, got %s", result["cve_id"])
	}

	// Verify CVE was saved
	saved, err := db.GetCVE("CVE-2021-TEST")
	if err != nil {
		t.Errorf("CVE was not saved: %v", err)
	}
	if saved.ID != "CVE-2021-TEST" {
		t.Errorf("Saved CVE ID mismatch: %s", saved.ID)
	}
}

func TestRPCIsCVEStoredByID(t *testing.T) {
	// Create a temporary database for testing
	dbPath := "/tmp/test_cve_local_check.db"
	defer os.Remove(dbPath)

	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Save a test CVE
	testCVE := &cve.CVEItem{
		ID:           "CVE-2021-EXISTS",
		SourceID:     "test@example.com",
		VulnStatus:   "Test",
		Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE"}},
	}
	db.SaveCVE(testCVE)

	// Create handler
	handler := createIsCVEStoredByIDHandler(db, logger)

	tests := []struct {
		name     string
		cveID    string
		expected bool
	}{
		{"Existing CVE", "CVE-2021-EXISTS", true},
		{"Non-existing CVE", "CVE-2021-NOTFOUND", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request message
			payload, _ := sonic.Marshal(map[string]string{
				"cve_id": tt.cveID,
			})

			msg := &subprocess.Message{
				Type:    subprocess.MessageTypeRequest,
				ID:      "RPCIsCVEStoredByID",
				Payload: payload,
			}

			// Call handler
			ctx := context.Background()
			resp, err := handler(ctx, msg)

			// Check results
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if resp.Type != subprocess.MessageTypeResponse {
				t.Errorf("Expected response type, got %s", resp.Type)
			}

			var result map[string]interface{}
			if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if result["cve_id"] != tt.cveID {
				t.Errorf("Expected cve_id=%s, got %s", tt.cveID, result["cve_id"])
			}

			if result["stored"].(bool) != tt.expected {
				t.Errorf("Expected stored=%v, got %v", tt.expected, result["stored"])
			}
		})
	}
}

func TestRPCGetCVEByID(t *testing.T) {
	// Create a temporary database for testing
	dbPath := "/tmp/test_cve_local_get.db"
	defer os.Remove(dbPath)

	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Save a test CVE
	testCVE := &cve.CVEItem{
		ID:           "CVE-2021-GETTEST",
		SourceID:     "test@example.com",
		VulnStatus:   "Test",
		Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE for Get"}},
	}
	db.SaveCVE(testCVE)

	// Create handler
	handler := createGetCVEByIDHandler(db, logger)

	t.Run("Get existing CVE", func(t *testing.T) {
		// Create request message
		payload, _ := sonic.Marshal(map[string]string{
			"cve_id": "CVE-2021-GETTEST",
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCGetCVEByID",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", resp.Type)
		}

		var result cve.CVEItem
		if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if result.ID != "CVE-2021-GETTEST" {
			t.Errorf("Expected CVE-2021-GETTEST, got %s", result.ID)
		}
	})

	t.Run("Get non-existing CVE", func(t *testing.T) {
		// Create request message
		payload, _ := sonic.Marshal(map[string]string{
			"cve_id": "CVE-2021-NOTFOUND",
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCGetCVEByID",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeError {
			t.Errorf("Expected error type, got %s", resp.Type)
		}
	})

	t.Run("Empty CVE ID", func(t *testing.T) {
		// Create request message
		payload, _ := sonic.Marshal(map[string]string{
			"cve_id": "",
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCGetCVEByID",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeError {
			t.Errorf("Expected error type, got %s", resp.Type)
		}

		if resp.Error == "" {
			t.Error("Expected error message for empty CVE ID")
		}
	})
}

func TestRPCDeleteCVEByID(t *testing.T) {
	// Create a temporary database for testing
	dbPath := "/tmp/test_cve_local_delete.db"
	defer os.Remove(dbPath)

	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Create handler
	handler := createDeleteCVEByIDHandler(db, logger)

	t.Run("Delete existing CVE", func(t *testing.T) {
		// Save a test CVE first
		testCVE := &cve.CVEItem{
			ID:           "CVE-2021-DELETETEST",
			SourceID:     "test@example.com",
			VulnStatus:   "Test",
			Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE for Delete"}},
		}
		db.SaveCVE(testCVE)

		// Create request message
		payload, _ := sonic.Marshal(map[string]string{
			"cve_id": "CVE-2021-DELETETEST",
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCDeleteCVEByID",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", resp.Type)
		}

		var result map[string]interface{}
		if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !result["success"].(bool) {
			t.Error("Expected success=true")
		}

		// Verify CVE was deleted
		_, err = db.GetCVE("CVE-2021-DELETETEST")
		if err == nil {
			t.Error("CVE should have been deleted")
		}
	})

	t.Run("Delete non-existing CVE", func(t *testing.T) {
		// Create request message
		payload, _ := sonic.Marshal(map[string]string{
			"cve_id": "CVE-2021-NOTEXIST",
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCDeleteCVEByID",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results - should return error
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeError {
			t.Errorf("Expected error type, got %s", resp.Type)
		}
	})

	t.Run("Empty CVE ID", func(t *testing.T) {
		// Create request message
		payload, _ := sonic.Marshal(map[string]string{
			"cve_id": "",
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCDeleteCVEByID",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeError {
			t.Errorf("Expected error type, got %s", resp.Type)
		}

		if resp.Error == "" {
			t.Error("Expected error message for empty CVE ID")
		}
	})
}

func TestRPCListCVEs(t *testing.T) {
	// Test constants
	const testCVECount = 15

	// Create a temporary database for testing
	dbPath := "/tmp/test_cve_local_list.db"
	defer os.Remove(dbPath)

	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Save test CVEs
	for i := 1; i <= testCVECount; i++ {
		testCVE := &cve.CVEItem{
			ID:           fmt.Sprintf("CVE-2021-LIST-%02d", i),
			SourceID:     "test@example.com",
			VulnStatus:   "Test",
			Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE"}},
		}
		db.SaveCVE(testCVE)
	}

	// Create handler
	handler := createListCVEsHandler(db, logger)

	t.Run("List with default pagination", func(t *testing.T) {
		// Create request message with no payload (use defaults)
		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCListCVEs",
			Payload: nil,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", resp.Type)
		}

		var result map[string]interface{}
		if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		cves := result["cves"].([]interface{})
		total := int64(result["total"].(float64))

		if len(cves) != 10 {
			t.Errorf("Expected 10 CVEs (default limit), got %d", len(cves))
		}

		if total != testCVECount {
			t.Errorf("Expected total=%d, got %d", testCVECount, total)
		}
	})

	t.Run("List with custom pagination", func(t *testing.T) {
		// Create request message with custom offset and limit
		payload, _ := sonic.Marshal(map[string]int{
			"offset": 5,
			"limit":  5,
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCListCVEs",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", resp.Type)
		}

		var result map[string]interface{}
		if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		cves := result["cves"].([]interface{})
		total := int64(result["total"].(float64))

		if len(cves) != 5 {
			t.Errorf("Expected 5 CVEs, got %d", len(cves))
		}

		if total != testCVECount {
			t.Errorf("Expected total=%d, got %d", testCVECount, total)
		}
	})

	t.Run("List with offset beyond total", func(t *testing.T) {
		// Create request message with offset beyond total
		payload, _ := sonic.Marshal(map[string]int{
			"offset": 20,
			"limit":  10,
		})

		msg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCListCVEs",
			Payload: payload,
		}

		// Call handler
		ctx := context.Background()
		resp, err := handler(ctx, msg)

		// Check results
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if resp.Type != subprocess.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", resp.Type)
		}

		var result map[string]interface{}
		if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		cves := result["cves"].([]interface{})
		total := int64(result["total"].(float64))

		if len(cves) != 0 {
			t.Errorf("Expected 0 CVEs, got %d", len(cves))
		}

		if total != testCVECount {
			t.Errorf("Expected total=%d, got %d", testCVECount, total)
		}
	})
}

func TestImportATTACKDataAtStartup(t *testing.T) {
	// Create a temporary database for testing
	attackDBPath := "/tmp/test_attack_store_startup.db"
	defer os.Remove(attackDBPath)

	attackStore, err := attack.NewLocalAttackStore(attackDBPath)
	if err != nil {
		t.Fatalf("Failed to create attack store: %v", err)
	}

	// Create logger
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Test the function with a non-existent directory
	// This should not crash and should log appropriately
	importATTACKDataAtStartup(attackStore, logger)

	// The function should complete without panicking
	t.Log("importATTACKDataAtStartup completed without crashing")
}

func TestMainFunctionInitialization(t *testing.T) {
	// This test ensures that the main function can initialize all components properly
	// without actually running the main loop

	// Test that we can create all the necessary components without error
	dbPath := "/tmp/test_init_cve.db"
	cweDBPath := "/tmp/test_init_cwe.db"
	capecDBPath := "/tmp/test_init_capec.db"
	attackDBPath := "/tmp/test_init_attack.db"
	
	defer os.Remove(dbPath)
	defer os.Remove(cweDBPath)
	defer os.Remove(capecDBPath)
	defer os.Remove(attackDBPath)

	// Test CVE DB creation
	db, err := local.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create CVE database: %v", err)
	}
	defer db.Close()

	// Test CWE store creation
	cweStore, err := cwe.NewLocalCWEStore(cweDBPath)
	if err != nil {
		t.Fatalf("Failed to create CWE store: %v", err)
	}

	// Test CAPEC store creation
	capecStore, err := capec.NewLocalCAPECStore(capecDBPath)
	if err != nil {
		t.Fatalf("Failed to create CAPEC store: %v", err)
	}

	// Test ATT&CK store creation
	attackStore, err := attack.NewLocalAttackStore(attackDBPath)
	if err != nil {
		t.Fatalf("Failed to create ATT&CK store: %v", err)
	}

	// Verify all stores are created properly
	if db == nil {
		t.Fatal("CVE database is nil")
	}
	if cweStore == nil {
		t.Fatal("CWE store is nil")
	}
	if capecStore == nil {
		t.Fatal("CAPEC store is nil")
	}
	if attackStore == nil {
		t.Fatal("ATT&CK store is nil")
	}

	t.Log("All stores initialized successfully")
}
