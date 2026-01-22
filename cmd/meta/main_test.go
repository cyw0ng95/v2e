package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/job"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// These tests focus on error handling and input validation
// Integration tests for full RPC flows are in the tests/ directory

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

	if errMsg.Error != "test error" {
		t.Errorf("Expected 'test error', got '%s'", errMsg.Error)
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
	reqPayload, _ := sonic.Marshal(map[string]string{
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

	if resp.Error != "cve_id is required" {
		t.Errorf("Expected 'cve_id is required', got '%s'", resp.Error)
	}
}

func TestRPCGetCVE_InvalidPayload(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createGetCVEHandler(rpcClient, logger)

	// Create request with invalid JSON payload
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCGetCVE",
		Payload: []byte("invalid json"),
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
		t.Error("Expected error message for invalid payload")
	}
}

func TestRPCCreateCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createCreateCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ := sonic.Marshal(map[string]string{
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

	if resp.Error != "cve_id is required" {
		t.Errorf("Expected 'cve_id is required', got '%s'", resp.Error)
	}
}

func TestRPCUpdateCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createUpdateCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ := sonic.Marshal(map[string]string{
		"cve_id": "",
	})
	msg := &subprocess.Message{
		Type:    subprocess.MessageTypeRequest,
		ID:      "RPCUpdateCVE",
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

	if resp.Error != "cve_id is required" {
		t.Errorf("Expected 'cve_id is required', got '%s'", resp.Error)
	}
}

func TestRPCDeleteCVE_EmptyCVEID(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	sp := subprocess.New("test")
	rpcClient := NewRPCClient(sp, logger)

	handler := createDeleteCVEHandler(rpcClient, logger)

	// Create request with empty CVE ID
	reqPayload, _ := sonic.Marshal(map[string]string{
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

	if resp.Error != "cve_id is required" {
		t.Errorf("Expected 'cve_id is required', got '%s'", resp.Error)
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
	reqPayload, _ := sonic.Marshal(map[string]string{})
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

func TestRecoverSession_NoSession(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// create a temporary session DB to avoid interfering with real data
	tmp := "./test_session.db"
	_ = os.Remove(tmp)
	sm, err := session.NewManager(tmp, logger)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer func() {
		sm.Close()
		_ = os.Remove(tmp)
	}()

	// Should not panic when no session exists
	recoverSession(nil, sm, logger)
}

func TestRecoverSession_RunningStartsJob(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_recover_running.db")
	sm, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer sm.Close()

	_, err = sm.CreateSession("s1", 0, 10)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if err := sm.UpdateState(session.StateRunning); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	// Blocking invoker keeps the job running until we close the channel
	inv := &blockingInvoker{ch: make(chan struct{})}
	controller := job.NewController(inv, sm, logger)

	recoverSession(controller, sm, logger)

	// Wait for controller to start
	started := false
	for i := 0; i < 100; i++ {
		if controller.IsRunning() {
			started = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !started {
		t.Fatal("expected controller to be started during recovery")
	}

	// Unblock invoker to allow the job to finish
	close(inv.ch)

	// Wait for job to stop
	for i := 0; i < 200; i++ {
		if !controller.IsRunning() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// Test that a paused session is not auto-resumed
func TestRecoverSession_PausedDoesNotStartJob(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_recover_paused.db")
	sm, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer sm.Close()

	_, err = sm.CreateSession("s2", 0, 10)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if err := sm.UpdateState(session.StatePaused); err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	ch := make(chan struct{})
	// Use a minimal invoker that would block if called
	invHandler := func(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
		select {
		case <-ch:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		// Return empty result so job will finish after unblocking
		empty := &cve.CVEResponse{}
		p, _ := sonic.Marshal(empty)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: p}, nil
	}

	controller := job.NewController(&fnInvoker{f: invHandler}, sm, logger)

	recoverSession(controller, sm, logger)

	// Give a short moment to ensure Start would have been called if it were
	time.Sleep(50 * time.Millisecond)

	if controller.IsRunning() {
		t.Fatal("controller should not be running for paused session")
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

	payload, _ := sonic.Marshal(map[string]bool{"success": true})
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
	p, _ := sonic.Marshal(empty)
	return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: p}, nil
}

// fnInvoker wraps a function as an RPCInvoker implementation
type fnInvoker struct {
	f func(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

func (f *fnInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	return f.f(ctx, target, method, params)
}
