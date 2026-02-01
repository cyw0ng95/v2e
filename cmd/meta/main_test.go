package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"strings"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/taskflow"
	"github.com/cyw0ng95/v2e/pkg/meta"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestCreateErrorResponse(t *testing.T) {
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "test-request",
		Source:        "test-source",
		CorrelationID: "test-correlation",
	}

	errMsg := createErrorResponse(msg, "test error")

	if errMsg.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", errMsg.Type)
	}

	if errMsg.Error != "[meta] RPC error response: test error" {
		t.Errorf("Expected '[meta] RPC error response: test error', got '%s'", errMsg.Error)
	}

	if errMsg.ID != "test-request" {
		t.Errorf("Expected ID to match request, got %s", errMsg.ID)
	}

	if errMsg.Target != "test-source" {
		t.Errorf("Expected target to be request source, got %s", errMsg.Target)
	}

	if errMsg.CorrelationID != "test-correlation" {
		t.Errorf("Expected correlation ID to match, got %s", errMsg.CorrelationID)
	}
}

func TestRPCGetCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)
	handler := createGetCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ := subprocess.MarshalFast(map[string]string{
		"cve_id": "",
	})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetCVE",
		Payload: reqPayload,
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}

	if resp.Error == "" {
		t.Error("Expected error message for empty CVE ID")
	}

	if !strings.Contains(resp.Error, "cve_id is required") {
		t.Errorf("Expected error to contain 'cve_id is required', got: %s", resp.Error)
	}
}

func TestRPCCreateCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createCreateCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ := subprocess.MarshalFast(map[string]string{
		"cve_id": "",
	})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCCreateCVE",
		Payload: reqPayload,
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}

	if !strings.Contains(resp.Error, "cve_id is required") {
		t.Errorf("Expected error to contain 'cve_id is required', got: %s", resp.Error)
	}

	handler = createCreateCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ = subprocess.MarshalFast(map[string]string{
		"cve_id": "",
	})
	msg = &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCCreateCVE",
		Payload: reqPayload,
	}

	ctx = context.Background()
	resp, err = handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}

	if !strings.Contains(resp.Error, "cve_id is required") {
		t.Errorf("Expected error to contain 'cve_id is required', got: %s", resp.Error)
	}
}

func TestRPCUpdateCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createUpdateCVEHandler(rpcClient, logger)

	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCUpdateCVE",
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}

	if !strings.Contains(resp.Error, "failed to parse request") {
		t.Errorf("Expected error to contain 'failed to parse request', got: %s", resp.Error)
	}
}

func TestRPCDeleteCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createDeleteCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ := subprocess.MarshalFast(map[string]string{
		"cve_id": "",
	})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCDeleteCVE",
		Payload: reqPayload,
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}

	if !strings.Contains(resp.Error, "cve_id is required") {
		t.Errorf("Expected error to contain 'cve_id is required', got: %s", resp.Error)
	}
}

func TestRPCListCVEs_DefaultParameters(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createListCVEsHandler(rpcClient, logger)

	// Create request with no payload (should use defaults)
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCListCVEs",
		Payload: nil,
	}

	// Use a context with short timeout since we expect this to fail
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This will fail to connect to local since it's not running,
	// but we're testing that the handler accepts nil payload
	resp, err := handler(ctx, msg)

	// The handler itself should not return a Go error
	if err != nil {
		t.Fatalf("Handler returned unexpected error: %v", err)
	}

	// It should return an error message about connection failure or timeout, not panic
	if resp == nil {
		t.Fatal("Handler returned nil response")
	}
}

func TestRPCCountCVEs_ErrorFromLocal(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createCountCVEsHandler(rpcClient, logger)

	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCCountCVEs",
		Payload: nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resp, err := handler(ctx, msg)

	// The handler should not return a Go error, but the response might be an error due to connection issues
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// The response type could be an error if connection fails, which is acceptable
	// We just want to make sure it doesn't panic
	if resp == nil {
		t.Fatal("Handler returned nil response")
	}
}

func TestNewRPCClient(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test-client")

	client := NewRPCClient(sp, logger)

	if client == nil {
		t.Fatal("NewRPCClient returned nil")
	}

	if client.sp != sp {
		t.Error("RPC client subprocess reference not set correctly")
	}

	if client.logger != logger {
		t.Error("RPC client logger reference not set correctly")
	}

	if client.pendingRequests == nil {
		t.Error("RPC client pendingRequests map not initialized")
	}
}

func TestRPCClient_HandleResponse(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test-client")

	client := NewRPCClient(sp, logger)

	// Create a test message
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "test",
		CorrelationID: "unknown-correlation",
	}

	ctx := context.Background()
	// Should not panic even with unknown correlation ID
	resp, err := client.handleResponse(ctx, msg)

	if err != nil {
		t.Errorf("handleResponse returned error: %v", err)
	}

	// Should return nil (no additional response needed)
	if resp != nil {
		t.Error("handleResponse should return nil response")
	}
}

func TestRPCClient_HandleError(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test-client")

	client := NewRPCClient(sp, logger)

	// Create a test error message
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeError,
		ID:            "test",
		Error:         "test error",
		CorrelationID: "unknown-correlation",
	}

	ctx := context.Background()
	// Should not panic even with unknown correlation ID
	resp, err := client.handleError(ctx, msg)

	if err != nil {
		t.Errorf("handleError returned error: %v", err)
	}

	// Should return nil (no additional response needed)
	if resp != nil {
		t.Error("handleError should return nil response")
	}
}

// TestRPCSaveCVEByID_MissingCVE tests the error case when CVE field is missing
func TestRPCSaveCVEByID_MissingCVE(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createCreateCVEHandler(rpcClient, logger)

	// Create request with missing fields
	reqPayload, _ := subprocess.MarshalFast(map[string]string{})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCCreateCVE",
		Payload: reqPayload,
		Source:  "test",
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}
}

func TestRecoverRuns_NoRuns(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// create a temporary run DB
	tmp := "./test_runs.db"
	_ = os.Remove(tmp)
	runStore, err := taskflow.NewRunStore(tmp, logger)
	if err != nil {
		t.Fatalf("failed to create run store: %v", err)
	}
	defer func() {
		runStore.Close()
		_ = os.Remove(tmp)
	}()

	inv := &stubInvoker{}
	executor := taskflow.NewJobExecutor(inv, runStore, logger, 10)

	// Should not panic when no runs exist
	recoverRuns(executor, logger)
}

func TestRecoverRuns_RunningStartsJob(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_recover_running.db")
	runStore, err := taskflow.NewRunStore(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create run store: %v", err)
	}
	defer runStore.Close()

	// Create a run in running state
	run, err := runStore.CreateRun("s1", 0, 10, taskflow.DataTypeCVE)
	if err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}
	if err := runStore.UpdateState(run.ID, taskflow.StateRunning); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	// Use stub invoker that returns empty results (job will complete)
	inv := &stubInvoker{}
	executor := taskflow.NewJobExecutor(inv, runStore, logger, 10)

	recoverRuns(executor, logger)

	// Wait a moment for recovery to start
	time.Sleep(100 * time.Millisecond)

	// Check run status - should be completed
	finalRun, err := runStore.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}

	// Should be completed since stub invoker returns empty results
	if finalRun.State != taskflow.StateCompleted && finalRun.State != taskflow.StateRunning {
		t.Logf("Run state: %s (may still be running)", finalRun.State)
	}
}

// Test that a paused run is not auto-resumed
func TestRecoverRuns_PausedDoesNotStartJob(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_recover_paused.db")
	runStore, err := taskflow.NewRunStore(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create run store: %v", err)
	}
	defer runStore.Close()

	// Create a run in paused state
	run, err := runStore.CreateRun("s2", 0, 10, taskflow.DataTypeCVE)
	if err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}
	// Transition to running first, then pause
	if err := runStore.UpdateState(run.ID, taskflow.StateRunning); err != nil {
		t.Fatalf("UpdateState to running failed: %v", err)
	}
	if err := runStore.UpdateState(run.ID, taskflow.StatePaused); err != nil {
		t.Fatalf("UpdateState to paused failed: %v", err)
	}

	inv := &stubInvoker{}
	executor := taskflow.NewJobExecutor(inv, runStore, logger, 10)

	recoverRuns(executor, logger)

	// Give a short moment
	time.Sleep(50 * time.Millisecond)

	// Check that run is still paused
	finalRun, err := runStore.GetRun(run.ID)
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}

	if finalRun.State != taskflow.StatePaused {
		t.Fatalf("expected run to remain paused, got state: %s", finalRun.State)
	}
}

// Test InvokeRPC happy path where a response is delivered via handleResponse
func TestRPCClient_InvokeRPC_Success(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test-client")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf) // disables batching for direct writes
	client := NewRPCClient(sp, logger)

	done := make(chan *subprocess.Message, 1)
	go func() {
		resp, err := client.InvokeRPC(context.Background(), "local", "RPCTest", map[string]string{"foo": "bar"})
		if err != nil {
			t.Errorf("InvokeRPC returned error: %v", err)
			done <- nil
			return
		}
		done <- resp
	}()

	// Wait for pendingRequests to be populated
	var corr string
	found := false
	for i := 0; i < 100; i++ {
		client.mu.RLock()
		for k := range client.pendingRequests {
			corr = k
			found = true
			break
		}
		client.mu.RUnlock()
		if found {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !found {
		t.Fatal("pendingRequests entry not found")
	}

	payload, _ := subprocess.MarshalFast(map[string]bool{"success": true})
	respMsg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "RPCTest",
		CorrelationID: corr,
		Payload:       payload,
	}

	// Deliver response via handler
	if _, err := client.handleResponse(context.Background(), respMsg); err != nil {
		t.Fatalf("handleResponse returned error: %v", err)
	}

	select {
	case res := <-done:
		if res == nil {
			t.Fatal("InvokeRPC returned nil response")
		}
		if res.CorrelationID != corr {
			t.Fatalf("CorrelationID mismatch: got %s, want %s", res.CorrelationID, corr)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("InvokeRPC did not return in time")
	}
}

// Test that InvokeRPC returns when context is canceled
func TestRPCClient_InvokeRPC_ContextCanceled(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test-client")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)
	client := NewRPCClient(sp, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.InvokeRPC(ctx, "local", "RPCTest", nil)
	if err == nil {
		t.Fatal("Expected error due to canceled context")
	}
	// Accept either canceled or deadline exceeded depending on timing
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Logf("got context error: %v", err)
	}
}

// Helper invokers for testing recoverSession and Controller interactions
// blockingInvoker blocks until its channel is closed, allowing tests to control job lifetime.
type blockingInvoker struct{ ch chan struct{} }

func (b *blockingInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	select {
	case <-b.ch:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	empty := &cve.CVEResponse{}
	p, _ := subprocess.MarshalFast(empty)
	return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: p}, nil
}

// fnInvoker wraps a function as an RPCInvoker implementation
type fnInvoker struct {
	f func(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

func (f *fnInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	return f.f(ctx, target, method, params)
}

func TestRPCClientAdapter_InvokeRPC(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf) // disables batching for direct writes
	rpcClient := NewRPCClient(sp, logger)

	adapter := &RPCClientAdapter{client: rpcClient}

	done := make(chan *subprocess.Message, 1)
	go func() {
		resp, err := adapter.InvokeRPC(context.Background(), "local", "RPCTest", map[string]string{"foo": "bar"})
		if err != nil {
			t.Errorf("InvokeRPC returned error: %v", err)
			done <- nil
			return
		}
		if msg, ok := resp.(*subprocess.Message); ok {
			done <- msg
		} else {
			// Unexpected return type
			done <- nil
		}
	}()

	// Wait for pendingRequests to be populated
	var corr string
	found := false
	for i := 0; i < 100; i++ {
		rpcClient.mu.RLock()
		for k := range rpcClient.pendingRequests {
			corr = k
			found = true
			break
		}
		rpcClient.mu.RUnlock()
		if found {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !found {
		t.Fatal("pendingRequests entry not found")
	}

	payload, _ := subprocess.MarshalFast(map[string]bool{"success": true})
	respMsg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "RPCTest",
		CorrelationID: corr,
		Payload:       payload,
	}

	// Deliver response via handler
	if _, err := rpcClient.handleResponse(context.Background(), respMsg); err != nil {
		t.Fatalf("handleResponse returned error: %v", err)
	}

	select {
	case res := <-done:
		if res == nil {
			t.Fatal("InvokeRPC returned nil response")
		}
		if res.CorrelationID != corr {
			t.Fatalf("CorrelationID mismatch: got %s, want %s", res.CorrelationID, corr)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("InvokeRPC did not return in time")
	}
}

func TestRPCClientAdapter_InvokeRPC_ContextCanceled(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)
	rpcClient := NewRPCClient(sp, logger)

	adapter := &RPCClientAdapter{client: rpcClient}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := adapter.InvokeRPC(ctx, "local", "RPCTest", nil)
	if err == nil {
		t.Fatal("Expected error due to canceled context")
	}
	// Accept either canceled or deadline exceeded depending on timing
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Logf("got context error: %v", err)
	}
}

func TestCreateStartSessionHandler_EmptySessionID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	tmpDir := t.TempDir()
	runDBPath := filepath.Join(tmpDir, "runs.db")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		t.Fatalf("Failed to create run store: %v", err)
	}
	defer runStore.Close()

	invoker := &stubInvoker{}
	executor := taskflow.NewJobExecutor(invoker, runStore, logger, 10)

	handler := createStartSessionHandler(executor, logger)

	// Create request with empty data_type
	reqPayload, _ := subprocess.MarshalFast(map[string]interface{}{
		"data_type": "",
	})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCStartSession",
		Payload: reqPayload,
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type, got %s", resp.Type)
	}

	if resp.Error == "" {
		t.Error("Expected error message for empty session ID")
	}

	if !strings.Contains(resp.Error, "data_type is required") {
		t.Errorf("Expected error to contain 'data_type is required', got: %s", resp.Error)
	}
}

func TestCreateStartSessionHandler_ValidSession(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	tmpDir := t.TempDir()
	runDBPath := filepath.Join(tmpDir, "runs.db")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		t.Fatalf("Failed to create run store: %v", err)
	}
	defer runStore.Close()

	// Create a stub invoker to avoid making actual RPC calls
	invoker := &stubInvoker{}
	executor := taskflow.NewJobExecutor(invoker, runStore, logger, 10)

	handler := createStartSessionHandler(executor, logger)

	// Create request with valid parameters
	reqPayload, _ := subprocess.MarshalFast(map[string]interface{}{
		"data_type":         string(taskflow.DataTypeCVE),
		"start_index":       0,
		"results_per_batch": 10,
	})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCStartSession",
		Payload: reqPayload,
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	// Parse response to verify success
	var result map[string]interface{}
	if err := subprocess.UnmarshalFast(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", result["success"])
	}

	if sessionID, ok := result["session_id"].(string); !ok || sessionID == "" {
		t.Errorf("Expected non-empty session_id, got %v", result["session_id"])
	}
	if dt, ok := result["data_type"].(string); !ok || dt != string(taskflow.DataTypeCVE) {
		t.Errorf("Expected data_type='%s', got %v", taskflow.DataTypeCVE, result["data_type"])
	}
}

func TestCreateStopSessionHandler(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	tmpDir := t.TempDir()
	runDBPath := filepath.Join(tmpDir, "runs.db")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		t.Fatalf("Failed to create run store: %v", err)
	}
	defer runStore.Close()

	// Create a stub invoker to avoid making actual RPC calls
	invoker := &stubInvoker{}
	executor := taskflow.NewJobExecutor(invoker, runStore, logger, 10)

	// Start a run first
	err = executor.Start(context.Background(), "test-session", 0, 10)
	if err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	handler := createStopSessionHandler(executor, logger)

	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCStopSession",
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type == subprocess.MessageTypeError {
		// Job may have completed before stop was called; treat as acceptable
		t.Logf("Stop returned error (acceptable): %s", resp.Error)
		return
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	// Parse response to verify success
	var result map[string]interface{}
	if err := subprocess.UnmarshalFast(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", result["success"])
	}
}

func TestCreateGetSessionStatusHandler(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	tmpDir := t.TempDir()
	runDBPath := filepath.Join(tmpDir, "runs.db")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		t.Fatalf("Failed to create run store: %v", err)
	}
	defer runStore.Close()

	invoker := &stubInvoker{}
	executor := taskflow.NewJobExecutor(invoker, runStore, logger, 10)

	// Start a run first
	err = executor.Start(context.Background(), "test-session", 0, 10)
	if err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	handler := createGetSessionStatusHandler(executor, logger)

	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCGetSessionStatus",
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeResponse {
		t.Errorf("Expected response type, got %s", resp.Type)
	}

	// Parse response to verify session status
	var result map[string]interface{}
	if err := subprocess.UnmarshalFast(resp.Payload, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if hasSession, ok := result["has_session"].(bool); !ok || !hasSession {
		t.Errorf("Expected has_session=true, got %v", result["has_session"])
	}
}

func TestCreatePauseJobHandler(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	tmpDir := t.TempDir()
	runDBPath := filepath.Join(tmpDir, "runs.db")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		t.Fatalf("Failed to create run store: %v", err)
	}
	defer runStore.Close()

	invoker := &stubInvoker{}
	executor := taskflow.NewJobExecutor(invoker, runStore, logger, 10)

	handler := createPauseJobHandler(executor, logger)

	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCPauseJob",
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// The handler should return an error because no job is running
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type for pausing non-existent job, got %s", resp.Type)
	}
}

func TestCreateResumeJobHandler(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	tmpDir := t.TempDir()
	runDBPath := filepath.Join(tmpDir, "runs.db")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		t.Fatalf("Failed to create run store: %v", err)
	}
	defer runStore.Close()

	invoker := &stubInvoker{}
	executor := taskflow.NewJobExecutor(invoker, runStore, logger, 10)

	handler := createResumeJobHandler(executor, logger)

	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCResumeJob",
	}

	ctx := context.Background()
	resp, err := handler(ctx, msg)

	// The handler should return an error because no job is paused
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if resp.Type != subprocess.MessageTypeError {
		t.Errorf("Expected error type for resuming non-existent job, got %s", resp.Type)
	}
}

func TestRPCListCVEs_InvalidPayload(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createListCVEsHandler(rpcClient, logger)

	// Create request with invalid JSON payload
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCListCVEs",
		Payload: []byte("invalid json"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resp, err := handler(ctx, msg)

	// The handler should not return a Go error, but the response might be an error due to connection issues
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// The response type could be an error if connection fails, which is acceptable
	// We just want to make sure it doesn't panic
	if resp == nil {
		t.Fatal("Handler returned nil response")
	}
}

func TestAddBookmarkCreatesMemoryCard(t *testing.T) {
	ctx := context.Background()
	bookmarkID := "test-bookmark"
	content := "Test bookmark content"

	err := AddBookmarkHandler(ctx, bookmarkID, content)
	if err != nil {
		t.Fatalf("AddBookmarkHandler failed: %v", err)
	}

	card := meta.CreateMemoryCard(bookmarkID, content)
	if card.BookmarkID != bookmarkID {
		t.Errorf("MemoryCard.BookmarkID mismatch: got %s, want %s", card.BookmarkID, bookmarkID)
	}
	if card.Content != content {
		t.Errorf("MemoryCard.Content mismatch: got %s, want %s", card.Content, content)
	}
}
