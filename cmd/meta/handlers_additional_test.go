package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/job"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// stubInvoker returns an immediate empty CVE response so runJob ends quickly
type stubInvoker struct{}

func (s *stubInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	// Return a subprocess.Message whose payload is a CVEResponse with no vulnerabilities
	resp := &subprocess.Message{
		Type: subprocess.MessageTypeResponse,
	}
	var vulns []struct {
		CVE cve.CVEItem `json:"cve"`
	}
	cr := cve.CVEResponse{Vulnerabilities: vulns}
	b, _ := sonic.Marshal(cr)
	resp.Payload = b
	return resp, nil
}

func TestCreateStartAndStopSessionHandler(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Create session manager
	tmp := t.TempDir()
	sessDB := tmp + "/sess.db"
	sm, err := session.NewManager(sessDB, logger)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer func() { sm.Close() }()

	// Create job controller with stub invoker
	inv := &stubInvoker{}
	jc := job.NewController(inv, sm, logger)

	// Create handlers
	startHandler := createStartSessionHandler(sm, jc, logger)
	stopHandler := createStopSessionHandler(sm, jc, logger)

	// Start session via handler
	req := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCStartSession"}
	b, _ := sonic.Marshal(map[string]interface{}{"session_id": "s1", "start_index": 0, "results_per_batch": 1})
	req.Payload = b

	resp, err := startHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("startHandler returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("startHandler returned nil response")
	}

	// ensure response payload indicates success
	var res map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &res); err != nil {
		t.Fatalf("failed to unmarshal start response: %v", err)
	}
	if res["success"] != true {
		t.Fatalf("expected success true, got %v", res["success"])
	}

	// Wait a short time for background job to run and stop itself
	time.Sleep(50 * time.Millisecond)

	// Now call stop handler (should succeed even if job already stopped)
	stopReq := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCStopSession"}
	stopResp, err := stopHandler(context.Background(), stopReq)
	if err != nil {
		t.Fatalf("stopHandler returned error: %v", err)
	}
	if stopResp == nil {
		t.Fatalf("stopHandler returned nil response")
	}
	var sres map[string]interface{}
	if err := sonic.Unmarshal(stopResp.Payload, &sres); err != nil {
		t.Fatalf("failed to unmarshal stop response: %v", err)
	}
	if sres["success"] != true {
		t.Fatalf("expected stop success true, got %v", sres["success"])
	}
}

func TestGetSessionStatusHandler_HasSession(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	// Create session manager
	tmp := t.TempDir()
	sessDB := tmp + "/sess2.db"
	sm, err := session.NewManager(sessDB, logger)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	defer func() { sm.Close() }()

	// Create a session via manager
	_, err = sm.CreateSession("s2", 5, 10)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	// Update progress
	if err := sm.UpdateProgress(3, 2, 1); err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	handler := createGetSessionStatusHandler(sm, logger)
	req := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetSessionStatus"}
	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("handler returned nil response")
	}
	var out map[string]interface{}
	if err := sonic.Unmarshal(resp.Payload, &out); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if has, ok := out["has_session"].(bool); !ok || !has {
		t.Fatalf("expected has_session=true, got %v", out["has_session"])
	}
	if out["fetched_count"].(float64) < 3 {
		t.Fatalf("expected fetched_count >= 3, got %v", out["fetched_count"])
	}
	if out["stored_count"].(float64) < 2 {
		t.Fatalf("expected stored_count >= 2, got %v", out["stored_count"])
	}
}
