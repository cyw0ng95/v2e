package main

import (
	"context"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/taskflow"
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
	b, _ := subprocess.MarshalFast(cr)
	resp.Payload = b
	return resp, nil
}

func TestCreateStartAndStopSessionHandler(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Create run store
	tmp := t.TempDir()
	runDB := tmp + "/runs.db"
	runStore, err := taskflow.NewRunStore(runDB, logger)
	if err != nil {
		t.Fatalf("failed to create run store: %v", err)
	}
	defer func() { runStore.Close() }()

	// Create job executor with stub invoker
	inv := &stubInvoker{}
	executor := taskflow.NewJobExecutor(inv, runStore, logger, 10)

	// Create handlers
	startHandler := createStartSessionHandler(executor, logger)
	stopHandler := createStopSessionHandler(executor, logger)

	// Start session via handler
	req := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCStartSession"}
	b, _ := subprocess.MarshalFast(map[string]interface{}{"session_id": "s1", "start_index": 0, "results_per_batch": 1})
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
	if err := subprocess.UnmarshalFast(resp.Payload, &res); err != nil {
		t.Fatalf("failed to unmarshal start response: %v", err)
	}
	if res["success"] != true {
		t.Fatalf("expected success true, got %v", res["success"])
	}

	// Don't wait for job to finish - stop it while it's still active
	stopReq := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCStopSession"}
	stopResp, err := stopHandler(context.Background(), stopReq)
	if err != nil {
		t.Fatalf("stopHandler returned error: %v", err)
	}
	if stopResp == nil {
		t.Fatalf("stopHandler returned nil response")
	}

	// Check if it's an error (job might have already completed) or success
	if stopResp.Type == subprocess.MessageTypeError {
		// Job completed before we could stop it - that's okay for this test
		t.Logf("Job completed before stop could be called: %s", stopResp.Error)
		return
	}

	var sres map[string]interface{}
	if err := subprocess.UnmarshalFast(stopResp.Payload, &sres); err != nil {
		t.Fatalf("failed to unmarshal stop response: %v", err)
	}
	if sres["success"] != true {
		t.Fatalf("expected stop success true, got %v", sres["success"])
	}
}

func TestGetSessionStatusHandler_HasSession(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)

	// Create run store
	tmp := t.TempDir()
	runDB := tmp + "/runs2.db"
	runStore, err := taskflow.NewRunStore(runDB, logger)
	if err != nil {
		t.Fatalf("failed to create run store: %v", err)
	}
	defer func() { runStore.Close() }()

	// Create job executor
	inv := &stubInvoker{}
	executor := taskflow.NewJobExecutor(inv, runStore, logger, 10)

	// Start a run to create some state
	err = executor.Start(context.Background(), "s2", 5, 10)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Update progress via store
	if err := runStore.UpdateProgress("s2", 3, 2, 1); err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	handler := createGetSessionStatusHandler(executor, logger)
	req := &subprocess.Message{Type: subprocess.MessageTypeRequest, ID: "RPCGetSessionStatus"}
	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("handler returned nil response")
	}
	var out map[string]interface{}
	if err := subprocess.UnmarshalFast(resp.Payload, &out); err != nil {
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
