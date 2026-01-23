package main

import (
	"context"
	"github.com/bytedance/sonic"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

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


func TestRPCGetCVEByID_EmptyString(t *testing.T) {
	// Create fetcher
	fetcher := remote.NewFetcher("")

	// Create handler
	handler := createGetCVEByIDHandler(fetcher)

	// Create request message with empty string CVE ID
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

func TestMalformedPayloadScenarios(t *testing.T) {
	// Test various malformed payloads for all handlers
	
	// Test 1: Invalid JSON structure
	invalidJSON := []byte("{malformed json")
	
	// Test RPCGetCVEByID with invalid JSON
	fetcher := remote.NewFetcher("")
	handler1 := createGetCVEByIDHandler(fetcher)
	msg1 := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetCVEByID",
		Payload: invalidJSON,
	}
	ctx := context.Background()
	resp1, err := handler1(ctx, msg1)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	if resp1.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error response for malformed JSON, got %s", resp1.Type)
	}
	t.Logf("RPCGetCVEByID correctly handled malformed JSON: %s", resp1.Error)
		// Test 2: Valid JSON but wrong field types
	wrongTypePayload, _ := sonic.Marshal(map[string]int{
		"cve_id": 12345, // Should be string
	})
	msg4 := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetCVEByID",
		Payload: wrongTypePayload,
	}
	resp4, err := handler1(ctx, msg4)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	// This might succeed or fail depending on how sonic handles type conversion
	t.Logf("RPCGetCVEByID handled wrong field type: %s", resp4.Type)
}

