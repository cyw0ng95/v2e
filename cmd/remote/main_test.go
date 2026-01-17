package main

import (
	"context"
	"github.com/bytedance/sonic"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRPCGetCVECnt(t *testing.T) {
	// Skip in short mode since this makes a real API call
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create fetcher (no API key for basic test)
	fetcher := remote.NewFetcher("")

	// Create handler
	handler := createGetCVECntHandler(fetcher)

	// Create request message with empty payload
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetCVECnt",
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

	totalResults := result["total_results"].(float64)
	if totalResults <= 0 {
		t.Errorf("Expected total_results > 0, got %f", totalResults)
	}

	t.Logf("Total CVEs in NVD: %.0f", totalResults)
}

func TestRPCGetCVEByID(t *testing.T) {
	// Skip in short mode since this makes a real API call
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create fetcher (no API key for basic test)
	fetcher := remote.NewFetcher("")

	// Create handler
	handler := createGetCVEByIDHandler(fetcher)

	// Create request message for a well-known CVE
	payload, _ := sonic.Marshal(map[string]string{
		"cve_id": "CVE-2021-44228", // Log4Shell
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

	var result map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that we got vulnerabilities
	vulns, ok := result["vulnerabilities"].([]interface{})
	if !ok {
		t.Fatal("Expected vulnerabilities array")
	}

	if len(vulns) == 0 {
		t.Error("Expected at least one vulnerability")
	}

	// Check the CVE ID
	if len(vulns) > 0 {
		vuln := vulns[0].(map[string]interface{})
		cveData := vuln["cve"].(map[string]interface{})
		cveID := cveData["id"].(string)
		if cveID != "CVE-2021-44228" {
			t.Errorf("Expected CVE-2021-44228, got %s", cveID)
		}
		t.Logf("Successfully fetched CVE: %s", cveID)
	}
}

func TestRPCGetCVEByID_InvalidID(t *testing.T) {
	// Create fetcher (no API key for basic test)
	fetcher := remote.NewFetcher("")

	// Create handler
	handler := createGetCVEByIDHandler(fetcher)

	// Create request message with empty CVE ID
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

	// Handler should not return Go error
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Should return error message
	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error message type, got %s", resp.Type)
	}

	if resp.Error == "" {
		t.Error("Expected error message for empty CVE ID")
	}

	t.Logf("Expected error for empty CVE ID: %s", resp.Error)
}

func TestRPCGetCVEByID_MissingField(t *testing.T) {
	// Create fetcher (no API key for basic test)
	fetcher := remote.NewFetcher("")

	// Create handler
	handler := createGetCVEByIDHandler(fetcher)

	// Create request message without cve_id field
	payload, _ := sonic.Marshal(map[string]string{})

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetCVEByID",
		Payload: payload,
	}

	// Call handler
	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// Handler should not return Go error
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Should return error message
	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error message type, got %s", resp.Type)
	}

	if resp.Error == "" {
		t.Error("Expected error message for missing cve_id field")
	}

	t.Logf("Expected error for missing cve_id field: %s", resp.Error)
}
