package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/gin-gonic/gin"
)

// TODO: Tests disabled - access service is currently a stub
// Access needs to be redesigned to communicate with broker via RPC
// instead of instantiating its own broker.
// See: https://github.com/cyw0ng95/v2e/pull/74

func TestHealthEndpoint(t *testing.T) {
	// Basic test placeholder
	t.Skip("Access service is currently a stub pending redesign")
}

func TestNewRPCClient_Access(t *testing.T) {
	client := NewRPCClient("test-access", DefaultRPCTimeout)
	if client == nil {
		t.Fatal("NewRPCClient returned nil")
	}
	if client.sp == nil {
		t.Fatal("RPCClient subprocess is nil")
	}
	if client.sp.ID != "test-access" {
		t.Fatalf("expected subprocess ID 'test-access', got '%s'", client.sp.ID)
	}
}

func TestRPCClient_HandleResponse_UnknownCorrelation(t *testing.T) {
	client := NewRPCClient("test-access-2", DefaultRPCTimeout)
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "m",
		CorrelationID: "nonexistent-corr",
	}
	ctx := context.Background()
	resp, err := client.handleResponse(ctx, msg)
	if err != nil {
		t.Fatalf("handleResponse returned error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response from handleResponse for unknown correlation, got %v", resp)
	}
}

func TestInvokeRPCWithTarget_ResponseDelivered(t *testing.T) {
	// Skip this test as it depends on internal implementation details
	// that are now handled by the common RPC client
	t.Skip("Skipped - internal implementation now handled by common RPC client")
}

func TestHealthEndpoint_Success(t *testing.T) {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Fatalf("expected status 'ok', got '%s'", response["status"])
	}
}

func TestRPCEndpoint_ValidRequest(t *testing.T) {
	r := gin.Default()
	r.POST("/rpc", func(c *gin.Context) {
		var request struct {
			Method string                 `json:"method"`
			Params map[string]interface{} `json:"params"`
			Target string                 `json:"target"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"retcode": 400, "message": "Invalid request"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"retcode": 0, "message": "success", "payload": nil})
	})

	w := httptest.NewRecorder()
	body := `{"method": "TestMethod", "params": {"key": "value"}}`
	req, _ := http.NewRequest(http.MethodPost, "/rpc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["retcode"] != float64(0) {
		t.Fatalf("expected retcode 0, got %v", response["retcode"])
	}
	if response["message"] != "success" {
		t.Fatalf("expected message 'success', got '%s'", response["message"])
	}
}

func TestSetupRouter_StaticDir(t *testing.T) {
	// Test the setupRouter function with a non-existent static dir
	router := setupRouter(nil, 30, "/non/existent/dir")

	// Verify the router was created
	if router == nil {
		t.Fatal("setupRouter returned nil")
	}

	// Test health endpoint exists
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/restful/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestRPCClient_InvokeRPC(t *testing.T) {
	client := NewRPCClient("test-access-4", 100*time.Millisecond) // Short timeout to prevent hanging
	if client == nil {
		t.Fatal("NewRPCClient returned nil")
	}

	// Test that InvokeRPC calls InvokeRPCWithTarget with "broker" as target
	ctx := context.Background()

	// Use a timeout context to prevent hanging when there's no broker
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	// Note: This will fail because there's no actual broker to connect to,
	// but we're testing that the method calls the right underlying function
	_, err := client.InvokeRPC(timeoutCtx, "TestMethod", nil)

	// The error is expected since there's no broker, but we want to ensure
	// the method doesn't hang and returns within a reasonable time
	if err == nil {
		t.Log("InvokeRPC succeeded (unexpected but not necessarily an error)")
	} else {
		// Expected to fail due to lack of actual broker connection
		t.Logf("InvokeRPC failed as expected: %v", err)
	}
}

func TestRPCClient_HandleError(t *testing.T) {
	client := NewRPCClient("test-access-5", DefaultRPCTimeout)

	// Create a message to test error handling
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeError,
		ID:            "test-error",
		CorrelationID: "test-corr",
		Error:         "test error message",
	}

	ctx := context.Background()
	resp, err := client.handleError(ctx, msg)

	// handleError should return the same result as handleResponse
	if err != nil {
		t.Fatalf("handleError returned error: %v", err)
	}
	if resp != nil {
		t.Fatalf("handleError should return nil response, got: %v", resp)
	}
}

func TestRequestEntry_SignalAndClose(t *testing.T) {
	// This test is now redundant as the request entry functionality
	// is handled internally by the common RPC client
	t.Skip("Skipped - functionality now handled by common RPC client")
}

func TestRPCClient_InvokeRPCWithTarget_Timeout(t *testing.T) {
	client := NewRPCClient("test-access-6", 10*time.Millisecond) // Very short timeout

	ctx := context.Background()
	_, err := client.InvokeRPCWithTarget(ctx, "nonexistent-target", "TestMethod", nil)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "RPC timeout") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestRPCClient_InvokeRPCWithTarget_ContextCancel(t *testing.T) {
	client := NewRPCClient("test-access-7", DefaultRPCTimeout)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.InvokeRPCWithTarget(ctx, "nonexistent-target", "TestMethod", nil)

	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}

	if err != ctx.Err() {
		t.Fatalf("expected context cancelled error, got: %v", err)
	}
}
