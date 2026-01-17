package job

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// mockRPCInvoker is a mock implementation of RPCInvoker for testing
type mockRPCInvoker struct {
	mu            sync.RWMutex
	fetchResponse *subprocess.Message
	saveResponse  *subprocess.Message
	fetchError    error
	saveError     error
	fetchCalls    int
	saveCalls     int
}

func (m *mockRPCInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if target == "cve-remote" && method == "RPCFetchCVEs" {
		m.fetchCalls++
		if m.fetchError != nil {
			return nil, m.fetchError
		}
		return m.fetchResponse, nil
	}

	if target == "cve-local" && method == "RPCSaveCVEByID" {
		m.saveCalls++
		if m.saveError != nil {
			return nil, m.saveError
		}
		return m.saveResponse, nil
	}

	return nil, errors.New("unknown RPC method")
}

func TestNewController(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_controller_session.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	mockRPC := &mockRPCInvoker{}
	controller := NewController(mockRPC, sessionManager, logger)

	if controller == nil {
		t.Fatal("NewController returned nil")
	}

	if controller.rpcInvoker != mockRPC {
		t.Error("RPC invoker not set correctly")
	}

	if controller.sessionManager != sessionManager {
		t.Error("Session manager not set correctly")
	}

	if controller.running {
		t.Error("Controller should not be running initially")
	}
}

func TestControllerStart_NoSession(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_start_no_session.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	mockRPC := &mockRPCInvoker{}
	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()
	err := controller.Start(ctx)

	if err == nil {
		t.Error("Expected error when starting without a session")
	}
}

func TestControllerStart_AlreadyRunning(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_start_already_running.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	// Create a session
	sessionManager.CreateSession("test-session", 0, 100)

	// Create mock responses for successful fetch (empty results to stop immediately)
	emptyResponse := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{},
	}
	payload, _ := sonic.Marshal(emptyResponse)
	fetchMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: payload,
	}

	mockRPC := &mockRPCInvoker{
		fetchResponse: fetchMsg,
	}
	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()
	err := controller.Start(ctx)
	if err != nil {
		t.Fatalf("First start failed: %v", err)
	}

	// Try to start again (should fail)
	err = controller.Start(ctx)
	if err != ErrJobRunning {
		t.Errorf("Expected ErrJobRunning, got %v", err)
	}

	// Clean up
	controller.Stop()
}

func TestControllerStop_NotRunning(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_stop_not_running.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	mockRPC := &mockRPCInvoker{}
	controller := NewController(mockRPC, sessionManager, logger)

	err := controller.Stop()
	if err != ErrJobNotRunning {
		t.Errorf("Expected ErrJobNotRunning, got %v", err)
	}
}

func TestControllerPause_NotRunning(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_pause_not_running.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	mockRPC := &mockRPCInvoker{}
	controller := NewController(mockRPC, sessionManager, logger)

	err := controller.Pause()
	if err != ErrJobNotRunning {
		t.Errorf("Expected ErrJobNotRunning, got %v", err)
	}
}

func TestControllerResume_NotPaused(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_resume_not_paused.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	// Create a session in idle state
	sessionManager.CreateSession("test-session", 0, 100)

	mockRPC := &mockRPCInvoker{}
	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()
	err := controller.Resume(ctx)

	if err == nil {
		t.Error("Expected error when resuming non-paused session")
	}
}

func TestControllerIsRunning(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_is_running.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 100)

	// Create mock responses for successful fetch (empty results to stop immediately)
	emptyResponse := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{},
	}
	payload, _ := sonic.Marshal(emptyResponse)
	fetchMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: payload,
	}

	mockRPC := &mockRPCInvoker{
		fetchResponse: fetchMsg,
	}
	controller := NewController(mockRPC, sessionManager, logger)

	if controller.IsRunning() {
		t.Error("Controller should not be running initially")
	}

	ctx := context.Background()
	controller.Start(ctx)

	if !controller.IsRunning() {
		t.Error("Controller should be running after Start()")
	}

	// Give the job a moment to complete (since we return empty results)
	time.Sleep(100 * time.Millisecond)

	// Job should stop itself when no more CVEs
	if controller.IsRunning() {
		t.Error("Controller should have stopped after processing empty results")
	}
}

func TestJobLoop_EmptyResults(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_job_empty.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 100)

	// Mock empty response
	emptyResponse := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{},
		TotalResults: 0,
	}
	payload, _ := sonic.Marshal(emptyResponse)
	fetchMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: payload,
	}

	mockRPC := &mockRPCInvoker{
		fetchResponse: fetchMsg,
	}
	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()
	controller.Start(ctx)

	// Wait for job to complete
	time.Sleep(200 * time.Millisecond)

	// Verify job stopped
	if controller.IsRunning() {
		t.Error("Job should have stopped after empty results")
	}

	// Verify no save calls were made (no CVEs to save)
	if mockRPC.saveCalls > 0 {
		t.Errorf("Expected 0 save calls, got %d", mockRPC.saveCalls)
	}
}

func TestJobLoop_WithResults(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_job_with_results.db")

	sessionManager, _ := session.NewManager(dbPath)
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 100)

	// Mock response with one CVE
	cveResponse := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{
			{
				CVE: cve.CVEItem{
					ID:         "CVE-2021-44228",
					SourceID:   "nvd",
					VulnStatus: "Analyzed",
				},
			},
		},
		TotalResults: 1,
	}

	// First call returns CVEs, second call returns empty (to stop the loop)
	firstPayload, _ := sonic.Marshal(cveResponse)
	firstMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: firstPayload,
	}

	emptyResponse := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{},
	}
	emptyPayload, _ := sonic.Marshal(emptyResponse)
	emptyMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: emptyPayload,
	}

	savePayload, _ := sonic.Marshal(map[string]interface{}{
		"success": true,
		"cve_id":  "CVE-2021-44228",
	})
	saveMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: savePayload,
	}

	mockRPC := &mockRPCInvoker{
		fetchResponse: firstMsg,
		saveResponse:  saveMsg,
	}

	// After first fetch, change to empty response
	go func() {
		time.Sleep(50 * time.Millisecond)
		mockRPC.mu.Lock()
		mockRPC.fetchResponse = emptyMsg
		mockRPC.mu.Unlock()
	}()

	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()
	controller.Start(ctx)

	// Wait for job to complete
	time.Sleep(2 * time.Second)

	// Verify job stopped
	if controller.IsRunning() {
		t.Error("Job should have stopped after processing all CVEs")
	}

	// Verify fetch was called
	if mockRPC.fetchCalls == 0 {
		t.Error("Expected fetch calls, got 0")
	}

	// Verify save was called
	if mockRPC.saveCalls == 0 {
		t.Error("Expected save calls, got 0")
	}

	// Verify progress was updated
	sess, _ := sessionManager.GetSession()
	if sess.FetchedCount == 0 {
		t.Error("Expected fetched count > 0")
	}
}
