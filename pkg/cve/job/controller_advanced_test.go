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
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

// TestConcurrentJobControl tests multiple concurrent job control commands
func TestConcurrentJobControl(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_concurrent.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	// Create session first
	sessionManager.CreateSession("test-session", 0, 100)

	// Create a response that won't complete immediately
	// to give us time to test concurrent pause commands
	cves := []struct {
		CVE cve.CVEItem `json:"cve"`
	}{
		{CVE: cve.CVEItem{ID: "CVE-2021-00001"}},
	}
	response := &cve.CVEResponse{Vulnerabilities: cves}
	payload, _ := sonic.Marshal(response)
	msg := &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}

	mockRPC := &mockRPCInvoker{
		fetchResponse: msg,
		saveResponse:  &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: []byte(`{"success":true}`)},
	}
	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()
	controller.Start(ctx)
	time.Sleep(300 * time.Millisecond) // Give job time to actually start

	// Test concurrent pause commands
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	pauseSuccessCount := 0
	var mu sync.Mutex

	// Try to pause 5 times concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := controller.Pause()
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				pauseSuccessCount++
			} else if err != ErrJobNotRunning {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for unexpected errors
	for err := range errors {
		t.Errorf("Unexpected error from concurrent pause: %v", err)
	}

	// At least one pause should have succeeded
	if pauseSuccessCount == 0 {
		t.Error("Expected at least one pause to succeed")
	}

	// Verify session state (should be paused or stopped)
	sess, _ := sessionManager.GetSession()
	if sess.State != session.StatePaused && sess.State != session.StateStopped {
		t.Errorf("Expected session to be paused or stopped, got %s", sess.State)
	}
}

// TestConcurrentStartAttempts tests multiple concurrent start attempts
func TestConcurrentStartAttempts(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_concurrent_start.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	// Create session
	sessionManager.CreateSession("test-session", 0, 100)

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

	// Use a mock that delays before returning to keep the job running
	// long enough for all concurrent Start() calls to execute
	mockRPC := &mockRPCInvokerWithDelay{
		fetchResponse: emptyMsg,
		delay:         200 * time.Millisecond,
	}
	controller := NewController(mockRPC, sessionManager, logger)

	ctx := context.Background()

	// Try to start 10 times concurrently
	var wg sync.WaitGroup
	successCount := 0
	alreadyRunningCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := controller.Start(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successCount++
			} else if err == ErrJobRunning {
				alreadyRunningCount++
			}
		}()
	}

	wg.Wait()

	// Only one should succeed, rest should get ErrJobRunning
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful start, got %d", successCount)
	}

	if alreadyRunningCount != 9 {
		t.Errorf("Expected 9 'already running' errors, got %d", alreadyRunningCount)
	}

	// Clean up
	controller.Stop()
}

// TestJobDataIntegrity verifies data integrity during job execution
func TestJobDataIntegrity(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_data_integrity.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 10)

	// Create multiple CVEs to test data integrity
	cves := []struct {
		CVE cve.CVEItem `json:"cve"`
	}{
		{CVE: cve.CVEItem{ID: "CVE-2021-00001", SourceID: "nvd", VulnStatus: "Analyzed"}},
		{CVE: cve.CVEItem{ID: "CVE-2021-00002", SourceID: "nvd", VulnStatus: "Analyzed"}},
		{CVE: cve.CVEItem{ID: "CVE-2021-00003", SourceID: "nvd", VulnStatus: "Analyzed"}},
	}

	response := &cve.CVEResponse{
		Vulnerabilities: cves,
		TotalResults:    3,
	}
	payload, _ := sonic.Marshal(response)
	fetchMsg := &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: payload,
	}

	// Track which CVEs were saved
	savedCVEs := make(map[string]bool)
	var saveMutex sync.Mutex

	// Create a custom mock that tracks saves
	customMock := &mockRPCInvokerWithTracking{
		fetchResponse: fetchMsg,
		savedCVEs:     savedCVEs,
		saveMutex:     &saveMutex,
	}

	controller := NewController(customMock, sessionManager, logger)

	ctx := context.Background()
	controller.Start(ctx)

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify all CVEs were saved
	saveMutex.Lock()
	defer saveMutex.Unlock()

	for _, cveStruct := range cves {
		if !savedCVEs[cveStruct.CVE.ID] {
			t.Errorf("CVE %s was not saved", cveStruct.CVE.ID)
		}
	}

	// Verify progress tracking
	sess, _ := sessionManager.GetSession()
	if sess.FetchedCount < int64(len(cves)) {
		t.Errorf("Expected fetched count >= %d, got %d", len(cves), sess.FetchedCount)
	}

	controller.Stop()
}

// mockRPCInvokerWithTracking is a mock that tracks saved CVEs
type mockRPCInvokerWithTracking struct {
	fetchResponse *subprocess.Message
	savedCVEs     map[string]bool
	saveMutex     *sync.Mutex
}

func (m *mockRPCInvokerWithTracking) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	if target == "remote" && method == "RPCFetchCVEs" {
		// Return empty after first call
		if len(m.savedCVEs) > 0 {
			emptyResp := &cve.CVEResponse{Vulnerabilities: []struct {
				CVE cve.CVEItem `json:"cve"`
			}{}}
			emptyPayload, _ := sonic.Marshal(emptyResp)
			return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: emptyPayload}, nil
		}
		return m.fetchResponse, nil
	}

	if target == "local" && method == "RPCSaveCVEByID" {
		// Accept either a map-based param (legacy) or typed RPC params
		switch p := params.(type) {
		case map[string]interface{}:
			if cveData, exists := p["cve"]; exists {
				if cveItem, ok := cveData.(cve.CVEItem); ok {
					m.saveMutex.Lock()
					m.savedCVEs[cveItem.ID] = true
					m.saveMutex.Unlock()
				}
			}
		case rpc.SaveCVEByIDParams:
			m.saveMutex.Lock()
			m.savedCVEs[p.CVE.ID] = true
			m.saveMutex.Unlock()
		case *rpc.SaveCVEByIDParams:
			if p != nil {
				m.saveMutex.Lock()
				m.savedCVEs[p.CVE.ID] = true
				m.saveMutex.Unlock()
			}
		}

		savePayload, _ := sonic.Marshal(map[string]interface{}{"success": true})
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: savePayload}, nil
	}

	return nil, errors.New("unknown RPC method")
}

// TestJobErrorHandling tests error scenarios during job execution
func TestJobErrorHandling(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_error_handling.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 10)

	// Test RPC fetch error
	mockRPC := &mockRPCInvoker{
		fetchError: errors.New("network error"),
	}

	controller := NewController(mockRPC, sessionManager, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	controller.Start(ctx)

	// Wait for error handling
	time.Sleep(1 * time.Second)

	// Verify error was counted
	sess, _ := sessionManager.GetSession()
	if sess.ErrorCount == 0 {
		t.Error("Expected error count > 0 after RPC failure")
	}

	controller.Stop()
}

// TestJobStateTransitions tests all valid state transitions
func TestJobStateTransitions(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_state_transitions.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 10)

	// Create a response with CVEs so job doesn't complete immediately
	cves := []struct {
		CVE cve.CVEItem `json:"cve"`
	}{
		{CVE: cve.CVEItem{ID: "CVE-2021-00001"}},
	}
	response := &cve.CVEResponse{Vulnerabilities: cves}
	payload, _ := sonic.Marshal(response)
	msg := &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}

	mockRPC := &mockRPCInvoker{
		fetchResponse: msg,
		saveResponse:  &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: []byte(`{"success":true}`)},
	}

	controller := NewController(mockRPC, sessionManager, logger)
	ctx := context.Background()

	// Test: idle -> running
	sess, _ := sessionManager.GetSession()
	if sess.State != session.StateIdle {
		t.Errorf("Initial state should be idle, got %s", sess.State)
	}

	controller.Start(ctx)
	time.Sleep(200 * time.Millisecond) // Give job time to actually start

	sess, _ = sessionManager.GetSession()
	if sess.State != session.StateRunning {
		t.Errorf("State should be running after start, got %s", sess.State)
	}

	// Test: running -> paused
	controller.Pause()
	time.Sleep(200 * time.Millisecond)

	sess, _ = sessionManager.GetSession()
	if sess.State != session.StatePaused {
		t.Errorf("State should be paused after pause, got %s", sess.State)
	}

	// Test: paused -> running
	controller.Resume(ctx)
	time.Sleep(200 * time.Millisecond)

	sess, _ = sessionManager.GetSession()
	if sess.State != session.StateRunning {
		t.Errorf("State should be running after resume, got %s", sess.State)
	}

	// Test: running -> stopped
	controller.Stop()
	time.Sleep(100 * time.Millisecond)

	sess, _ = sessionManager.GetSession()
	if sess.State != session.StateStopped {
		t.Errorf("State should be stopped after stop, got %s", sess.State)
	}
}

// TestJobPauseResumeMultipleTimes tests multiple pause/resume cycles
func TestJobPauseResumeMultipleTimes(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_pause_resume_multiple.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 10)

	// Create a response with CVEs so job doesn't complete
	cves := []struct {
		CVE cve.CVEItem `json:"cve"`
	}{
		{CVE: cve.CVEItem{ID: "CVE-2021-00001"}},
	}
	response := &cve.CVEResponse{Vulnerabilities: cves}
	payload, _ := sonic.Marshal(response)
	msg := &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}

	mockRPC := &mockRPCInvoker{
		fetchResponse: msg,
		saveResponse:  &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: []byte(`{"success":true}`)},
	}

	controller := NewController(mockRPC, sessionManager, logger)
	ctx := context.Background()

	controller.Start(ctx)
	time.Sleep(200 * time.Millisecond)

	// Pause and resume 5 times
	for i := 0; i < 5; i++ {
		err := controller.Pause()
		if err != nil {
			t.Errorf("Pause iteration %d failed: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)

		sess, _ := sessionManager.GetSession()
		if sess.State != session.StatePaused {
			t.Errorf("Iteration %d: expected paused state, got %s", i, sess.State)
		}

		err = controller.Resume(ctx)
		if err != nil {
			t.Errorf("Resume iteration %d failed: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)

		sess, _ = sessionManager.GetSession()
		if sess.State != session.StateRunning {
			t.Errorf("Iteration %d: expected running state, got %s", i, sess.State)
		}
	}

	controller.Stop()
}

// TestJobProgressTracking tests accurate progress tracking
func TestJobProgressTracking(t *testing.T) {
	logger := common.NewLogger(os.Stderr, "test", common.InfoLevel)
	dbPath := filepath.Join(t.TempDir(), "test_progress_tracking.db")

	sessionManager, err := session.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	defer sessionManager.Close()

	sessionManager.CreateSession("test-session", 0, 10)

	// Create batches of CVEs
	batch1 := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{
			{CVE: cve.CVEItem{ID: "CVE-2021-00001"}},
			{CVE: cve.CVEItem{ID: "CVE-2021-00002"}},
		},
	}

	batch2 := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{
			{CVE: cve.CVEItem{ID: "CVE-2021-00003"}},
		},
	}

	emptyBatch := &cve.CVEResponse{
		Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{},
	}

	payload1, _ := sonic.Marshal(batch1)
	payload2, _ := sonic.Marshal(batch2)
	emptyPayload, _ := sonic.Marshal(emptyBatch)

	// Create a custom mock that returns different batches
	customMock := &mockRPCInvokerWithBatches{
		batches: []*subprocess.Message{
			{Type: subprocess.MessageTypeResponse, Payload: payload1},
			{Type: subprocess.MessageTypeResponse, Payload: payload2},
			{Type: subprocess.MessageTypeResponse, Payload: emptyPayload},
		},
	}

	controller := NewController(customMock, sessionManager, logger)
	ctx := context.Background()

	controller.Start(ctx)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify progress
	sess, _ := sessionManager.GetSession()

	// Should have fetched 3 CVEs total (2 + 1)
	if sess.FetchedCount != 3 {
		t.Errorf("Expected fetched count 3, got %d", sess.FetchedCount)
	}

	// Should have stored 3 CVEs
	if sess.StoredCount != 3 {
		t.Errorf("Expected stored count 3, got %d", sess.StoredCount)
	}

	// Should have no errors
	if sess.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", sess.ErrorCount)
	}
}

// mockRPCInvokerWithBatches returns different batches on each call
type mockRPCInvokerWithBatches struct {
	batches   []*subprocess.Message
	callCount int
	mu        sync.Mutex
}

func (m *mockRPCInvokerWithBatches) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	if target == "remote" && method == "RPCFetchCVEs" {
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.callCount < len(m.batches) {
			msg := m.batches[m.callCount]
			m.callCount++
			return msg, nil
		}
		// Return empty after all batches
		emptyResp := &cve.CVEResponse{Vulnerabilities: []struct {
			CVE cve.CVEItem `json:"cve"`
		}{}}
		emptyPayload, _ := sonic.Marshal(emptyResp)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: emptyPayload}, nil
	}

	if target == "local" && method == "RPCSaveCVEByID" {
		savePayload, _ := sonic.Marshal(map[string]interface{}{"success": true})
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: savePayload}, nil
	}

	return nil, errors.New("unknown method")
}

// mockRPCInvokerWithDelay is a mock that delays before returning fetch responses
// This is used to keep jobs running long enough for concurrency tests
type mockRPCInvokerWithDelay struct {
	fetchResponse *subprocess.Message
	delay         time.Duration
}

func (m *mockRPCInvokerWithDelay) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	if target == "remote" && method == "RPCFetchCVEs" {
		// Delay to keep the job running
		time.Sleep(m.delay)
		return m.fetchResponse, nil
	}

	if target == "local" && method == "RPCSaveCVEByID" {
		savePayload, _ := sonic.Marshal(map[string]interface{}{"success": true})
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: savePayload}, nil
	}

	return nil, errors.New("unknown method")
}
