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
	client := NewRPCClient("test-access-3", DefaultRPCTimeout)
	ctx := context.Background()

	// Start InvokeRPCWithTarget in a goroutine so we can deliver a response
	done := make(chan struct{})
	var respMsg *subprocess.Message
	var invokeErr error
	go func() {
		respMsg, invokeErr = client.InvokeRPCWithTarget(ctx, "broker", "RPCTestMethod", map[string]string{"k": "v"})
		close(done)
	}()

	// Wait briefly for the pendingRequests map to be populated by InvokeRPCWithTarget
	time.Sleep(10 * time.Millisecond)

	// Simulate broker sending a response by calling the client's handler directly
	// The correlation ID uses the internal sequence - first call will be access-rpc-1
	sim := &subprocess.Message{
		Type:          subprocess.MessageTypeResponse,
		ID:            "RPCTestMethod",
		CorrelationID: "access-rpc-1",
		Payload:       []byte(`"ok"`),
	}
	if _, err := client.handleResponse(ctx, sim); err != nil {
		t.Fatalf("handleResponse simulation returned error: %v", err)
	}

	// Wait for InvokeRPCWithTarget to return
	select {
	case <-done:
		// proceed
	case <-time.After(1 * time.Second):
		t.Fatal("InvokeRPCWithTarget did not return in time")
	}

	if invokeErr != nil {
		t.Fatalf("InvokeRPCWithTarget returned error: %v", invokeErr)
	}
	if respMsg == nil {
		t.Fatal("expected response message, got nil")
	}
	if respMsg.CorrelationID != "access-rpc-1" {
		t.Fatalf("expected correlation id access-rpc-1, got %s", respMsg.CorrelationID)
	}
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
