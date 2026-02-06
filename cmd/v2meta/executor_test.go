package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/cyw0ng95/v2e/pkg/urn"
	"gorm.io/gorm"
)

// testLogger creates a logger for tests
func testExecutorLogger() *common.Logger {
	return common.NewLogger(os.Stderr, "test", common.InfoLevel)
}

// MockProvider implements a minimal ProviderFSM for testing
type MockProvider struct {
	*fsm.BaseProviderFSM
	executeFunc func() error
	executed    bool
	permitCount int
}

func NewMockProvider(id string) *MockProvider {
	baseFSM, _ := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           id,
		ProviderType: "mock",
	})

	return &MockProvider{
		BaseProviderFSM: baseFSM,
	}
}

func (m *MockProvider) Execute() error {
	m.executed = true
	if m.executeFunc != nil {
		return m.executeFunc()
	}
	return nil
}

func (m *MockProvider) UpdatePermits(count int) error {
	m.permitCount = count
	return nil
}

func (m *MockProvider) GetLastCheckpoint() (string, error) {
	stats := m.GetStats()
	if cp, ok := stats["last_checkpoint"].(string); ok {
		return cp, nil
	}
	return "", nil
}

func (m *MockProvider) SaveCheckpoint(checkpoint string) error {
	itemURN := urn.MustParse(checkpoint)
	return m.BaseProviderFSM.SaveCheckpoint(itemURN, true, "")
}

// MockRPCClient for executor tests
type MockExecutorRPCClient struct {
	requestPermitsResponse map[string]interface{}
	requestPermitsError    error
	releasePermitsError    error
	permitRequests         []int
	permitReleases         []int
}

func NewMockExecutorRPCClient() *MockExecutorRPCClient {
	return &MockExecutorRPCClient{
		requestPermitsResponse: make(map[string]interface{}),
		permitRequests:         make([]int, 0),
		permitReleases:         make([]int, 0),
	}
}

func (m *MockExecutorRPCClient) InvokeRPC(ctx context.Context, target string, method string, params interface{}) (*subprocess.Message, error) {
	switch method {
	case "RPCRequestPermits":
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if count, ok := paramsMap["count"].(int); ok {
				m.permitRequests = append(m.permitRequests, count)
			}
		}
		if m.requestPermitsError != nil {
			return nil, m.requestPermitsError
		}
		
		// Convert to subprocess.Message
		payloadBytes, _ := json.Marshal(m.requestPermitsResponse)
		return &subprocess.Message{
			Type:    subprocess.MessageTypeResponse,
			Payload: json.RawMessage(payloadBytes),
		}, nil
		
	case "RPCReleasePermits":
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if count, ok := paramsMap["count"].(int); ok {
				m.permitReleases = append(m.permitReleases, count)
			}
		}
		if m.releasePermitsError != nil {
			return nil, m.releasePermitsError
		}
		
		// Convert to subprocess.Message
		response := map[string]interface{}{"success": true}
		payloadBytes, _ := json.Marshal(response)
		return &subprocess.Message{
			Type:    subprocess.MessageTypeResponse,
			Payload: json.RawMessage(payloadBytes),
		}, nil
		
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// Test 1: PermitExecutor - Create Executor
func TestPermitExecutor_NewPermitExecutor(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewPermitExecutor", nil, func(t *testing.T, tx *gorm.DB) {
		executor := NewPermitExecutor(&rpc.Client{}, testExecutorLogger())

		if executor == nil {
			t.Fatal("Executor is nil")
		}

		if len(executor.activeJobs) != 0 {
			t.Errorf("Active jobs count = %d, want 0", len(executor.activeJobs))
		}
	})
}

// Test 2: PermitExecutor - Start Provider Success
func TestPermitExecutor_StartProvider_Success(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StartProviderSuccess", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-1")

		err := executor.StartProvider(provider, 5)
		if err != nil {
			t.Fatalf("Failed to start provider: %v", err)
		}

		// Wait for async execution to start
		time.Sleep(50 * time.Millisecond)

		// Check provider state
		if provider.GetState() != fsm.ProviderRunning {
			t.Errorf("Provider state = %s, want RUNNING", provider.GetState())
		}

		// Check permits
		if provider.permitCount != 5 {
			t.Errorf("Provider permits = %d, want 5", provider.permitCount)
		}

		// Check active jobs
		if len(executor.activeJobs) != 1 {
			t.Errorf("Active jobs count = %d, want 1", len(executor.activeJobs))
		}
	})
}

// Test 3: PermitExecutor - Start Provider with No Permits Available
func TestPermitExecutor_StartProvider_NoPermits(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StartProviderNoPermits", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(0), // No permits granted
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-2")

		err := executor.StartProvider(provider, 5)
		if err != nil {
			t.Fatalf("StartProvider should succeed with no permits: %v", err)
		}

		// Provider should transition to WAITING_QUOTA
		if provider.GetState() != fsm.ProviderWaitingQuota {
			t.Errorf("Provider state = %s, want WAITING_QUOTA", provider.GetState())
		}
	})
}

// Test 4: PermitExecutor - Start Already Running Provider
func TestPermitExecutor_StartProvider_AlreadyRunning(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StartProviderAlreadyRunning", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-3")

		// Start first time
		err := executor.StartProvider(provider, 5)
		if err != nil {
			t.Fatalf("Failed to start provider: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		// Try to start again
		err = executor.StartProvider(provider, 5)
		if err == nil {
			t.Error("Expected error when starting already running provider, got nil")
		}
	})
}

// Test 5: PermitExecutor - RPC Request Permits Error
func TestPermitExecutor_StartProvider_RPCError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StartProviderRPCError", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsError = fmt.Errorf("broker unavailable")

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-4")

		err := executor.StartProvider(provider, 5)
		if err == nil {
			t.Error("Expected error for RPC failure, got nil")
		}

		// Provider should be rolled back to IDLE
		if provider.GetState() != fsm.ProviderIdle {
			t.Errorf("Provider state = %s, want IDLE (rollback)", provider.GetState())
		}
	})
}

// Test 6: PermitExecutor - Invalid Permit Response
func TestPermitExecutor_StartProvider_InvalidResponse(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StartProviderInvalidResponse", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": "invalid", // String instead of number
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-5")

		err := executor.StartProvider(provider, 5)
		if err == nil {
			t.Error("Expected error for invalid response, got nil")
		}

		// Provider should be rolled back
		if provider.GetState() != fsm.ProviderIdle {
			t.Errorf("Provider state = %s, want IDLE", provider.GetState())
		}
	})
}

// Test 7: PermitExecutor - Pause Provider
func TestPermitExecutor_PauseProvider(t *testing.T) {
	testutils.Run(t, testutils.Level1, "PauseProvider", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-6")

		// Start provider
		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)

		// Pause provider
		err := executor.PauseProvider("test-provider-6")
		if err != nil {
			t.Fatalf("Failed to pause provider: %v", err)
		}

		// Check state
		if provider.GetState() != fsm.ProviderPaused {
			t.Errorf("Provider state = %s, want PAUSED", provider.GetState())
		}

		// Permits should be released
		if len(mockRPC.permitReleases) == 0 {
			t.Error("Permits should have been released on pause")
		}
	})
}

// Test 8: PermitExecutor - Pause Non-Existent Provider
func TestPermitExecutor_PauseProvider_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "PauseProviderNotFound", nil, func(t *testing.T, tx *gorm.DB) {
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testExecutorLogger())

		err := executor.PauseProvider("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent provider, got nil")
		}
	})
}

// Test 9: PermitExecutor - Pause Not Running Provider
func TestPermitExecutor_PauseProvider_NotRunning(t *testing.T) {
	testutils.Run(t, testutils.Level1, "PauseProviderNotRunning", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(0), // No permits, will be WAITING_QUOTA
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-7")

		executor.StartProvider(provider, 5)

		// Try to pause (provider is not running, in WAITING_QUOTA)
		err := executor.PauseProvider("test-provider-7")
		if err == nil {
			t.Error("Expected error when pausing non-running provider, got nil")
		}
	})
}

// Test 10: PermitExecutor - Resume Provider
func TestPermitExecutor_ResumeProvider(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ResumeProvider", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-8")

		// Start, pause, then resume
		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)
		executor.PauseProvider("test-provider-8")

		// Resume
		err := executor.ResumeProvider("test-provider-8")
		if err != nil {
			t.Fatalf("Failed to resume provider: %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		// Should be running again
		if provider.GetState() != fsm.ProviderRunning {
			t.Errorf("Provider state = %s, want RUNNING", provider.GetState())
		}
	})
}

// Test 11: PermitExecutor - Resume Non-Paused Provider
func TestPermitExecutor_ResumeProvider_NotPaused(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ResumeProviderNotPaused", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-9")

		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)

		// Try to resume running provider
		err := executor.ResumeProvider("test-provider-9")
		if err == nil {
			t.Error("Expected error when resuming non-paused provider, got nil")
		}
	})
}

// Test 12: PermitExecutor - Stop Provider
func TestPermitExecutor_StopProvider(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StopProvider", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-10")

		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)

		// Stop provider
		err := executor.StopProvider("test-provider-10")
		if err != nil {
			t.Fatalf("Failed to stop provider: %v", err)
		}

		// Provider should be terminated
		if provider.GetState() != fsm.ProviderTerminated {
			t.Errorf("Provider state = %s, want TERMINATED", provider.GetState())
		}

		// Should be removed from active jobs
		if len(executor.activeJobs) != 0 {
			t.Errorf("Active jobs count = %d, want 0", len(executor.activeJobs))
		}
	})
}

// Test 13: PermitExecutor - Stop Non-Existent Provider
func TestPermitExecutor_StopProvider_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "StopProviderNotFound", nil, func(t *testing.T, tx *gorm.DB) {
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testExecutorLogger())

		err := executor.StopProvider("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent provider, got nil")
		}
	})
}

// Test 14: PermitExecutor - Handle Quota Revoked
func TestPermitExecutor_HandleQuotaRevoked(t *testing.T) {
	testutils.Run(t, testutils.Level1, "HandleQuotaRevoked", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-11")

		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)

		// Revoke 3 permits
		err := executor.HandleQuotaRevoked("test-provider-11", 3)
		if err != nil {
			t.Fatalf("Failed to handle quota revocation: %v", err)
		}

		// Provider should have 2 permits left
		if provider.permitCount != 2 {
			t.Errorf("Provider permits = %d, want 2", provider.permitCount)
		}
	})
}

// Test 15: PermitExecutor - Handle Quota Revoked All Permits
func TestPermitExecutor_HandleQuotaRevoked_AllPermits(t *testing.T) {
	testutils.Run(t, testutils.Level1, "HandleQuotaRevokedAll", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("test-provider-12")

		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)

		// Revoke all 5 permits
		err := executor.HandleQuotaRevoked("test-provider-12", 5)
		if err != nil {
			t.Fatalf("Failed to handle quota revocation: %v", err)
		}

		// Provider should transition to WAITING_QUOTA
		time.Sleep(50 * time.Millisecond)
		if provider.GetState() != fsm.ProviderWaitingQuota {
			t.Errorf("Provider state = %s, want WAITING_QUOTA", provider.GetState())
		}
	})
}

// Test 16: PermitExecutor - Get Active Providers
func TestPermitExecutor_GetActiveProviders(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetActiveProviders", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())

		// Start multiple providers
		provider1 := NewMockProvider("provider-1")
		provider2 := NewMockProvider("provider-2")
		provider3 := NewMockProvider("provider-3")

		executor.StartProvider(provider1, 5)
		executor.StartProvider(provider2, 5)
		executor.StartProvider(provider3, 5)

		time.Sleep(50 * time.Millisecond)

		activeIDs := executor.GetActiveProviders()
		if len(activeIDs) != 3 {
			t.Errorf("Active providers count = %d, want 3", len(activeIDs))
		}
	})
}

// Test 17: PermitExecutor - Register Shutdown Hook
func TestPermitExecutor_RegisterShutdownHook(t *testing.T) {
	testutils.Run(t, testutils.Level1, "RegisterShutdownHook", nil, func(t *testing.T, tx *gorm.DB) {
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testExecutorLogger())

		hookCalled := false
		executor.RegisterShutdownHook(func() error {
			hookCalled = true
			return nil
		})

		if len(executor.shutdownHooks) != 1 {
			t.Errorf("Shutdown hooks count = %d, want 1", len(executor.shutdownHooks))
		}

		// Execute shutdown
		executor.GracefulShutdown()

		if !hookCalled {
			t.Error("Shutdown hook was not called")
		}
	})
}

// Test 18: PermitExecutor - Graceful Shutdown
func TestPermitExecutor_GracefulShutdown(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GracefulShutdown", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())

		// Start providers
		provider1 := NewMockProvider("provider-1")
		provider2 := NewMockProvider("provider-2")

		executor.StartProvider(provider1, 5)
		executor.StartProvider(provider2, 5)

		time.Sleep(50 * time.Millisecond)

		// Graceful shutdown
		err := executor.GracefulShutdown()
		if err != nil {
			t.Fatalf("Graceful shutdown failed: %v", err)
		}

		// All providers should be terminated
		if provider1.GetState() != fsm.ProviderTerminated {
			t.Errorf("Provider 1 state = %s, want TERMINATED", provider1.GetState())
		}

		if provider2.GetState() != fsm.ProviderTerminated {
			t.Errorf("Provider 2 state = %s, want TERMINATED", provider2.GetState())
		}

		// All active jobs should be cleared
		if len(executor.activeJobs) != 0 {
			t.Errorf("Active jobs count = %d, want 0", len(executor.activeJobs))
		}
	})
}

// Test 19: PermitExecutor - Graceful Shutdown with Checkpoints
func TestPermitExecutor_GracefulShutdown_WithCheckpoints(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GracefulShutdownCheckpoints", nil, func(t *testing.T, tx *gorm.DB) {
		mockRPC := NewMockExecutorRPCClient()
		mockRPC.requestPermitsResponse = map[string]interface{}{
			"granted": float64(5),
		}

		executor := NewPermitExecutor(mockRPC, testExecutorLogger())
		provider := NewMockProvider("provider-1")

		// Set a checkpoint
		provider.SaveCheckpoint("v2e::nvd::cve::CVE-2024-00001")

		executor.StartProvider(provider, 5)
		time.Sleep(50 * time.Millisecond)

		// Graceful shutdown should save checkpoint
		err := executor.GracefulShutdown()
		if err != nil {
			t.Fatalf("Graceful shutdown failed: %v", err)
		}

		stats := provider.GetStats()
		if stats["last_checkpoint"] != "CVE-2024-00001" {
			t.Error("Checkpoint should be preserved on shutdown")
		}
	})
}

// Test 20: PermitExecutor - Multiple Shutdown Hooks
func TestPermitExecutor_MultipleShutdownHooks(t *testing.T) {
	testutils.Run(t, testutils.Level1, "MultipleShutdownHooks", nil, func(t *testing.T, tx *gorm.DB) {
		executor := NewPermitExecutor(NewMockExecutorRPCClient(), testExecutorLogger())

		hook1Called := false
		hook2Called := false
		hook3Called := false

		executor.RegisterShutdownHook(func() error {
			hook1Called = true
			return nil
		})

		executor.RegisterShutdownHook(func() error {
			hook2Called = true
			return nil
		})

		executor.RegisterShutdownHook(func() error {
			hook3Called = true
			return nil
		})

		executor.GracefulShutdown()

		if !hook1Called || !hook2Called || !hook3Called {
			t.Error("Not all shutdown hooks were called")
		}
	})
}
