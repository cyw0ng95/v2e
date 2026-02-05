package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/cyw0ng95/v2e/pkg/urn"
	"gorm.io/gorm"
)

// testLogger creates a logger for tests
func testLogger() *common.Logger {
	return common.NewLogger(os.Stderr, "test", common.InfoLevel)
}

// MockRPCClient implements a mock RPC client for testing
type MockRPCClient struct {
	responses map[string]map[string]interface{}
	errors    map[string]error
	callCount map[string]int
}

func NewMockRPCClient() *MockRPCClient {
	return &MockRPCClient{
		responses: make(map[string]map[string]interface{}),
		errors:    make(map[string]error),
		callCount: make(map[string]int),
	}
}

func (m *MockRPCClient) InvokeRPC(ctx context.Context, target string, method string, params interface{}) (map[string]interface{}, error) {
	key := target + "::" + method
	m.callCount[key]++

	if err, exists := m.errors[key]; exists {
		return nil, err
	}

	if resp, exists := m.responses[key]; exists {
		return resp, nil
	}

	return nil, fmt.Errorf("no mock response for %s", key)
}

func (m *MockRPCClient) SetResponse(target, method string, response map[string]interface{}) {
	key := target + "::" + method
	m.responses[key] = response
}

func (m *MockRPCClient) SetError(target, method string, err error) {
	key := target + "::" + method
	m.errors[key] = err
}

func (m *MockRPCClient) GetCallCount(target, method string) int {
	key := target + "::" + method
	return m.callCount[key]
}

// Test 1: Provider Factory Creation
func TestProviderFactory_CreateProvider_CVE(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CreateCVEProvider", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: &rpc.Client{},
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeCVE, "test-cve-1", map[string]interface{}{
			"batch_size": 50,
		})

		if err != nil {
			t.Fatalf("Failed to create CVE provider: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider is nil")
		}

		if provider.GetID() != "test-cve-1" {
			t.Errorf("Provider ID = %s, want test-cve-1", provider.GetID())
		}

		if provider.GetType() != "cve" {
			t.Errorf("Provider type = %s, want cve", provider.GetType())
		}
	})
}

// Test 2: Provider Factory Creation - CWE
func TestProviderFactory_CreateProvider_CWE(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CreateCWEProvider", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: &rpc.Client{},
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeCWE, "test-cwe-1", map[string]interface{}{
			"file_path": "/tmp/cwe.xml",
		})

		if err != nil {
			t.Fatalf("Failed to create CWE provider: %v", err)
		}

		if provider.GetType() != "cwe" {
			t.Errorf("Provider type = %s, want cwe", provider.GetType())
		}
	})
}

// Test 3: Provider Factory Creation - CAPEC
func TestProviderFactory_CreateProvider_CAPEC(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CreateCAPECProvider", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: &rpc.Client{},
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeCAPEC, "test-capec-1", map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to create CAPEC provider: %v", err)
		}

		if provider.GetType() != "capec" {
			t.Errorf("Provider type = %s, want capec", provider.GetType())
		}
	})
}

// Test 4: Provider Factory Creation - ATT&CK
func TestProviderFactory_CreateProvider_ATTACK(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CreateATTACKProvider", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: &rpc.Client{},
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeATTACK, "test-attack-1", map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to create ATT&CK provider: %v", err)
		}

		if provider.GetType() != "attack" {
			t.Errorf("Provider type = %s, want attack", provider.GetType())
		}
	})
}

// Test 5: Provider Factory - Unsupported Type
func TestProviderFactory_CreateProvider_UnsupportedType(t *testing.T) {
	testutils.Run(t, testutils.Level1, "UnsupportedType", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: &rpc.Client{},
			Logger:    testLogger(),
		})

		_, err := factory.CreateProvider("unknown", "test-1", map[string]interface{}{})

		if err == nil {
			t.Error("Expected error for unsupported type, got nil")
		}
	})
}

// Test 6: Provider Factory - Get Supported Providers
func TestProviderFactory_GetSupportedProviders(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetSupportedProviders", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{})

		supported := factory.GetSupportedProviders()

		if len(supported) != 4 {
			t.Errorf("Supported providers count = %d, want 4", len(supported))
		}

		expectedTypes := map[ProviderType]bool{
			ProviderTypeCVE:    true,
			ProviderTypeCWE:    true,
			ProviderTypeCAPEC:  true,
			ProviderTypeATTACK: true,
		}

		for _, providerType := range supported {
			if !expectedTypes[providerType] {
				t.Errorf("Unexpected provider type: %s", providerType)
			}
		}
	})
}

// Test 7: CVE Provider - Basic Execution
func TestCVEProvider_ExecuteBatch_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ExecuteBatchSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		// Mock successful CVE fetch
		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": []map[string]interface{}{
				{
					"cve_id":        "CVE-2024-00001",
					"description":   "Test CVE",
					"last_modified": "2024-01-01T00:00:00Z",
				},
			},
		})

		// Mock successful CVE save
		mockRPC.SetResponse("local", "RPCSaveCVE", map[string]interface{}{
			"success": true,
		})

		provider, err := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			Storage:   &storage.Store{},
			RPCClient: mockRPC,
			Logger:    testLogger(),
			BatchSize: 1,
		})

		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}

		err = provider.executeBatch()
		if err != nil {
			t.Errorf("ExecuteBatch failed: %v", err)
		}

		// Verify RPC calls
		if mockRPC.GetCallCount("remote", "RPCFetchCVEBatch") != 1 {
			t.Errorf("FetchCVEBatch call count = %d, want 1", mockRPC.GetCallCount("remote", "RPCFetchCVEBatch"))
		}

		if mockRPC.GetCallCount("local", "RPCSaveCVE") != 1 {
			t.Errorf("SaveCVE call count = %d, want 1", mockRPC.GetCallCount("local", "RPCSaveCVE"))
		}
	})
}

// Test 8: CVE Provider - Field-Level Diffing
func TestCVEProvider_DiffFields_NoChanges(t *testing.T) {
	testutils.Run(t, testutils.Level1, "DiffFieldsNoChanges", nil, func(t *testing.T, tx *gorm.DB) {
		provider := &CVEProvider{
			logger: testLogger(),
		}

		existing := map[string]interface{}{
			"cve_id":      "CVE-2024-00001",
			"description": "Test description",
			"severity":    "HIGH",
		}

		incoming := map[string]interface{}{
			"cve_id":      "CVE-2024-00001",
			"description": "Test description",
			"severity":    "HIGH",
		}

		changed := provider.diffFields(existing, incoming)

		if len(changed) != 0 {
			t.Errorf("Expected no changes, got %d changes", len(changed))
		}
	})
}

// Test 9: CVE Provider - Field-Level Diffing with Changes
func TestCVEProvider_DiffFields_WithChanges(t *testing.T) {
	testutils.Run(t, testutils.Level1, "DiffFieldsWithChanges", nil, func(t *testing.T, tx *gorm.DB) {
		provider := &CVEProvider{
			logger: testLogger(),
		}

		existing := map[string]interface{}{
			"cve_id":      "CVE-2024-00001",
			"description": "Old description",
			"severity":    "MEDIUM",
		}

		incoming := map[string]interface{}{
			"cve_id":      "CVE-2024-00001",
			"description": "New description",
			"severity":    "HIGH",
		}

		changed := provider.diffFields(existing, incoming)

		if len(changed) != 2 {
			t.Errorf("Expected 2 changes, got %d", len(changed))
		}

		if changed["description"] != "New description" {
			t.Errorf("Description not updated correctly")
		}

		if changed["severity"] != "HIGH" {
			t.Errorf("Severity not updated correctly")
		}
	})
}

// Test 10: CVE Provider - Error Threshold Check - Below Threshold
func TestCVEProvider_CheckErrorThreshold_BelowThreshold(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ErrorThresholdBelowLimit", nil, func(t *testing.T, tx *gorm.DB) {
		baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
			ID:           "test-cve",
			ProviderType: "cve",
		})

		provider := &CVEProvider{
			BaseProviderFSM:  baseFSM,
			errorCount:       5,
			totalProcessed:   100,
			failureThreshold: 0.1, // 10%
			logger:           testLogger(),
		}

		err := provider.checkErrorThreshold()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

// Test 11: CVE Provider - Error Threshold Check - Above Threshold (Auto-Pause)
func TestCVEProvider_CheckErrorThreshold_AboveThreshold(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ErrorThresholdAutoFail", nil, func(t *testing.T, tx *gorm.DB) {
		baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
			ID:           "test-cve",
			ProviderType: "cve",
		})

		provider := &CVEProvider{
			BaseProviderFSM:  baseFSM,
			errorCount:       15,
			totalProcessed:   100,
			failureThreshold: 0.1, // 10%
			logger:           testLogger(),
		}

		err := provider.checkErrorThreshold()
		if err == nil {
			t.Error("Expected error for threshold exceeded, got nil")
		}

		// Provider should transition to PAUSED
		if provider.GetState() != fsm.ProviderPaused {
			t.Errorf("Provider state = %s, want PAUSED", provider.GetState())
		}
	})
}

// Test 12: CVE Provider - Checkpoint Saving
func TestCVEProvider_SaveCheckpoint_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "SaveCheckpointSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
			ID:           "test-cve",
			ProviderType: "cve",
		})

		provider := &CVEProvider{
			BaseProviderFSM: baseFSM,
			logger:          testLogger(),
		}

		itemURN := urn.MustParse("v2e::nvd::cve::CVE-2024-12345")

		err := provider.SaveCheckpoint(itemURN, true, "")
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		stats := provider.GetStats()
		if stats["last_checkpoint"] != itemURN.Key() {
			t.Errorf("Checkpoint not saved correctly")
		}
	})
}

// Test 13: CVE Provider - Incremental Fetching
func TestCVEProvider_IncrementalFetching(t *testing.T) {
	testutils.Run(t, testutils.Level1, "IncrementalFetching", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		lastModDate := "2024-01-01T00:00:00Z"

		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": []map[string]interface{}{
				{
					"cve_id":        "CVE-2024-00001",
					"last_modified": lastModDate,
				},
			},
		})

		mockRPC.SetResponse("local", "RPCSaveCVE", map[string]interface{}{
			"success": true,
		})

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:               "test-cve",
			RPCClient:        mockRPC,
			Logger:           testLogger(),
			LastModStartDate: lastModDate,
		})

		err := provider.executeBatch()
		if err != nil {
			t.Errorf("ExecuteBatch failed: %v", err)
		}

		// Verify incremental fetch used lastModStartDate
		if provider.lastModStartDate != lastModDate {
			t.Errorf("LastModStartDate not set correctly")
		}
	})
}

// Test 14: CVE Provider - Get Progress
func TestCVEProvider_GetProgress(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetProgress", nil, func(t *testing.T, tx *gorm.DB) {
		baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
			ID:           "test-cve",
			ProviderType: "cve",
		})

		provider := &CVEProvider{
			BaseProviderFSM: baseFSM,
			errorCount:      10,
			totalProcessed:  200,
			batchSize:       100,
			logger:          testLogger(),
		}

		progress := provider.GetProgress()

		if progress["total_processed"].(int64) != 200 {
			t.Errorf("Total processed = %d, want 200", progress["total_processed"])
		}

		if progress["error_count"].(int64) != 10 {
			t.Errorf("Error count = %d, want 10", progress["error_count"])
		}

		errorRate := progress["error_rate"].(float64)
		expectedRate := 10.0 / 200.0
		if errorRate != expectedRate {
			t.Errorf("Error rate = %f, want %f", errorRate, expectedRate)
		}
	})
}

// Test 15: CVE Provider - Empty Batch (Completion)
func TestCVEProvider_EmptyBatch_Completion(t *testing.T) {
	testutils.Run(t, testutils.Level1, "EmptyBatchCompletion", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		// Return empty CVE list
		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": []map[string]interface{}{},
		})

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			RPCClient: mockRPC,
			Logger:    testLogger(),
		})

		err := provider.executeBatch()
		if err != nil {
			t.Errorf("ExecuteBatch should succeed with empty batch: %v", err)
		}
	})
}

// Test 16: CVE Provider - RPC Fetch Error
func TestCVEProvider_FetchError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "FetchError", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		mockRPC.SetError("remote", "RPCFetchCVEBatch", fmt.Errorf("network error"))

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			RPCClient: mockRPC,
			Logger:    testLogger(),
		})

		err := provider.executeBatch()
		if err == nil {
			t.Error("Expected error for RPC failure, got nil")
		}

		if provider.errorCount != 1 {
			t.Errorf("Error count = %d, want 1", provider.errorCount)
		}
	})
}

// Test 17: CVE Provider - Save Error
func TestCVEProvider_SaveError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "SaveError", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": []map[string]interface{}{
				{"cve_id": "CVE-2024-00001"},
			},
		})

		// Simulate save error
		mockRPC.SetError("local", "RPCGetCVE", fmt.Errorf("not found"))
		mockRPC.SetError("local", "RPCSaveCVE", fmt.Errorf("database error"))

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			RPCClient: mockRPC,
			Logger:    testLogger(),
		})

		err := provider.executeBatch()
		// Batch should complete but increment error count
		if err != nil {
			t.Errorf("Batch execution should not fail: %v", err)
		}

		if provider.errorCount == 0 {
			t.Error("Error count should be incremented for save failure")
		}
	})
}

// Test 18: CVE Provider - Missing CVE ID
func TestCVEProvider_MissingCVEID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "MissingCVEID", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": []map[string]interface{}{
				{"description": "No CVE ID"}, // Missing cve_id
			},
		})

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			RPCClient: mockRPC,
			Logger:    testLogger(),
		})

		err := provider.executeBatch()
		if err != nil {
			t.Errorf("Should handle missing CVE ID gracefully: %v", err)
		}

		if provider.errorCount == 0 {
			t.Error("Error count should be incremented for missing CVE ID")
		}
	})
}

// Test 19: CVE Provider - Update Existing CVE
func TestCVEProvider_UpdateExistingCVE(t *testing.T) {
	testutils.Run(t, testutils.Level1, "UpdateExistingCVE", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		// Mock existing CVE
		existingCVE, _ := json.Marshal(map[string]interface{}{
			"cve_id":      "CVE-2024-00001",
			"description": "Old description",
		})
		
		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": []map[string]interface{}{
				{
					"cve_id":      "CVE-2024-00001",
					"description": "New description",
				},
			},
		})

		mockRPC.SetResponse("local", "RPCGetCVE", map[string]interface{}{
			"result": string(existingCVE),
		})

		mockRPC.SetResponse("local", "RPCUpdateCVE", map[string]interface{}{
			"success": true,
		})

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			RPCClient: mockRPC,
			Logger:    testLogger(),
		})

		err := provider.executeBatch()
		if err != nil {
			t.Errorf("ExecuteBatch failed: %v", err)
		}
	})
}

// Test 20: CVE Provider - Default Configuration
func TestCVEProvider_DefaultConfiguration(t *testing.T) {
	testutils.Run(t, testutils.Level1, "DefaultConfiguration", nil, func(t *testing.T, tx *gorm.DB) {
		provider, err := NewCVEProvider(CVEProviderConfig{
			ID:        "test-cve",
			RPCClient: NewMockRPCClient(),
			Logger:    testLogger(),
			// No batch size or thresholds specified
		})

		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}

		if provider.batchSize != 100 {
			t.Errorf("Default batch size = %d, want 100", provider.batchSize)
		}

		if provider.checkpointInterval != 100 {
			t.Errorf("Default checkpoint interval = %d, want 100", provider.checkpointInterval)
		}

		if provider.failureThreshold != 0.1 {
			t.Errorf("Default failure threshold = %f, want 0.1", provider.failureThreshold)
		}
	})
}

// Test 21: CVE Provider - Custom Configuration
func TestCVEProvider_CustomConfiguration(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CustomConfiguration", nil, func(t *testing.T, tx *gorm.DB) {
		provider, err := NewCVEProvider(CVEProviderConfig{
			ID:                 "test-cve",
			RPCClient:          NewMockRPCClient(),
			Logger:             testLogger(),
			BatchSize:          50,
			CheckpointInterval: 25,
			FailureThreshold:   0.05,
		})

		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}

		if provider.batchSize != 50 {
			t.Errorf("Batch size = %d, want 50", provider.batchSize)
		}

		if provider.checkpointInterval != 25 {
			t.Errorf("Checkpoint interval = %d, want 25", provider.checkpointInterval)
		}

		if provider.failureThreshold != 0.05 {
			t.Errorf("Failure threshold = %f, want 0.05", provider.failureThreshold)
		}
	})
}

// Test 22: CWE Provider - Creation with File Path
func TestCWEProvider_CreationWithFilePath(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CWECreationWithFilePath", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: NewMockRPCClient(),
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeCWE, "test-cwe", map[string]interface{}{
			"file_path":           "/tmp/cwe.xml",
			"batch_size":          200,
			"failure_threshold":   0.15,
			"checkpoint_interval": 50,
		})

		if err != nil {
			t.Fatalf("Failed to create CWE provider: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider is nil")
		}
	})
}

// Test 23: CAPEC Provider - Default Options
func TestCAPECProvider_DefaultOptions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CAPECDefaultOptions", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: NewMockRPCClient(),
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeCAPEC, "test-capec", map[string]interface{}{})

		if err != nil {
			t.Fatalf("Failed to create CAPEC provider: %v", err)
		}

		if provider.GetID() != "test-capec" {
			t.Errorf("Provider ID = %s, want test-capec", provider.GetID())
		}
	})
}

// Test 24: ATT&CK Provider - Custom Batch Size
func TestATTACKProvider_CustomBatchSize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ATTACKCustomBatchSize", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: NewMockRPCClient(),
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeATTACK, "test-attack", map[string]interface{}{
			"batch_size": 500,
		})

		if err != nil {
			t.Fatalf("Failed to create ATT&CK provider: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider is nil")
		}
	})
}

// Test 25: Provider Factory - Options Type Assertions
func TestProviderFactory_OptionsTypeAssertions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "OptionsTypeAssertions", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   &storage.Store{},
			RPCClient: NewMockRPCClient(),
			Logger:    testLogger(),
		})

		// Test with wrong type for batch_size (should ignore invalid type)
		provider, err := factory.CreateProvider(ProviderTypeCVE, "test-cve", map[string]interface{}{
			"batch_size": "invalid", // String instead of int
		})

		if err != nil {
			t.Fatalf("Provider creation should succeed despite invalid option: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider is nil")
		}
	})
}

// Test 26: DeepEqual - Simple Values
func TestDeepEqual_SimpleValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "DeepEqualSimple", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name     string
			a        interface{}
			b        interface{}
			expected bool
		}{
			{"Equal strings", "test", "test", true},
			{"Different strings", "test1", "test2", false},
			{"Equal ints", 123, 123, true},
			{"Different ints", 123, 456, false},
			{"Nil values", nil, nil, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := deepEqual(tt.a, tt.b)
				if result != tt.expected {
					t.Errorf("deepEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})
}

// Test 27: CVE Provider - Checkpoint Interval
func TestCVEProvider_CheckpointInterval(t *testing.T) {
	testutils.Run(t, testutils.Level1, "CheckpointInterval", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockRPCClient()
		
		// Create batch of CVEs that triggers checkpoint
		cves := make([]map[string]interface{}, 10)
		for i := 0; i < 10; i++ {
			cves[i] = map[string]interface{}{
				"cve_id": fmt.Sprintf("CVE-2024-%05d", i+1),
			}
		}

		mockRPC.SetResponse("remote", "RPCFetchCVEBatch", map[string]interface{}{
			"cves": cves,
		})

		mockRPC.SetResponse("local", "RPCSaveCVE", map[string]interface{}{
			"success": true,
		})

		provider, _ := NewCVEProvider(CVEProviderConfig{
			ID:                 "test-cve",
			RPCClient:          mockRPC,
			Logger:             testLogger(),
			CheckpointInterval: 5, // Checkpoint every 5 items
		})

		err := provider.executeBatch()
		if err != nil {
			t.Errorf("ExecuteBatch failed: %v", err)
		}

		// Should have processed 10 items
		if provider.totalProcessed != 10 {
			t.Errorf("Total processed = %d, want 10", provider.totalProcessed)
		}
	})
}

// Test 28: CVE Provider - Zero Error Rate
func TestCVEProvider_ZeroErrorRate(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ZeroErrorRate", nil, func(t *testing.T, tx *gorm.DB) {
		baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
			ID:           "test-cve",
			ProviderType: "cve",
		})

		provider := &CVEProvider{
			BaseProviderFSM: baseFSM,
			errorCount:      0,
			totalProcessed:  0, // No processing yet
			logger:          testLogger(),
		}

		err := provider.checkErrorThreshold()
		if err != nil {
			t.Errorf("Should not error with no processing: %v", err)
		}
	})
}

// Test 29: CVE Provider - High Error Rate
func TestCVEProvider_HighErrorRate(t *testing.T) {
	testutils.Run(t, testutils.Level1, "HighErrorRate", nil, func(t *testing.T, tx *gorm.DB) {
		baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
			ID:           "test-cve",
			ProviderType: "cve",
		})

		provider := &CVEProvider{
			BaseProviderFSM:  baseFSM,
			errorCount:       50,
			totalProcessed:   100,
			failureThreshold: 0.1, // 10%
			logger:           testLogger(),
		}

		err := provider.checkErrorThreshold()
		if err == nil {
			t.Error("Expected error for 50% error rate, got nil")
		}
	})
}

// Test 30: Provider Factory - Nil Storage
func TestProviderFactory_NilStorage(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NilStorage", nil, func(t *testing.T, tx *gorm.DB) {
		factory := NewProviderFactory(FactoryConfig{
			Storage:   nil, // Nil storage
			RPCClient: NewMockRPCClient(),
			Logger:    testLogger(),
		})

		provider, err := factory.CreateProvider(ProviderTypeCVE, "test-cve", map[string]interface{}{})

		// Should still create provider (storage is optional for some operations)
		if err != nil {
			t.Fatalf("Provider creation should succeed with nil storage: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider is nil")
		}
	})
}
