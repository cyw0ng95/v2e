package job

import (
	"context"
	"errors"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// mockCWERPCInvoker simulates RPC calls for CWE job testing.
type mockCWERPCInvoker struct {
	mu              sync.Mutex
	callCount       int
	shouldFail      bool
	failAfterN      int
	returnEmptyData bool
}

func (m *mockCWERPCInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldFail || (m.failAfterN > 0 && m.callCount > m.failAfterN) {
		return nil, errors.New("mock CWE RPC failure")
	}

	if m.returnEmptyData {
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: []byte(`{"views":[]}`)}, nil
	}

	view := cwe.CWEView{ID: "1000", Name: "Research Concepts"}
	payload := struct {
		Views []cwe.CWEView `json:"views"`
	}{Views: []cwe.CWEView{view}}

	data, _ := subprocess.MarshalFast(payload)
	return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: data}, nil
}

// TestCWEController_StartStop ensures basic CWE job lifecycle.
func TestCWEController_StartStop(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCWEController_StartStop", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testCWEWriter{t}, "test", common.ErrorLevel)
		invoker := &mockCWERPCInvoker{}
		ctrl := NewController(invoker, logger)

		if ctrl.IsRunning() {
			t.Fatalf("controller should not be running initially")
		}

		params := map[string]interface{}{"start_index": 0, "results_per_page": 50}
		sessionID, err := ctrl.Start(context.Background(), params)
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}

		if sessionID == "" {
			t.Fatalf("expected non-empty session ID")
		}

		if !ctrl.IsRunning() {
			t.Fatalf("controller should be running after Start")
		}

		if err := ctrl.Stop(context.Background(), sessionID); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		if ctrl.IsRunning() {
			t.Fatalf("controller should not be running after Stop")
		}
	})

}

// TestCWEController_ConcurrentStarts ensures only one Start succeeds.
func TestCWEController_ConcurrentStarts(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCWEController_ConcurrentStarts", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(testCWEWriter{t}, "test", common.ErrorLevel)
		invoker := &mockCWERPCInvoker{}
		ctrl := NewController(invoker, logger)

		params := map[string]interface{}{"start_index": 0, "results_per_page": 50}
		numGoroutines := 20
		var wg sync.WaitGroup
		successes := 0
		var mu sync.Mutex

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				_, err := ctrl.Start(context.Background(), params)
				if err == nil {
					mu.Lock()
					successes++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		mu.Lock()
		defer mu.Unlock()
		if successes != 1 {
			t.Fatalf("expected exactly 1 Start to succeed, got %d", successes)
		}

		ctrl.Stop(context.Background(), "test-session")
	})

}

// testCWEWriter adapts testing.T to io.Writer for logger output.
type testCWEWriter struct{ t *testing.T }

func (w testCWEWriter) Write(p []byte) (int, error) {
	w.t.Logf("%s", string(p))
	return len(p), nil
}
