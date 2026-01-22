package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/gin-gonic/gin"
)

// registerHandlers registers the REST endpoints on the provided router group
func registerHandlers(restful *gin.RouterGroup, rpcClient *RPCClient, rpcTimeoutSec int) {
	// Health check endpoint
	restful.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Generic RPC forwarding endpoint
	restful.POST("/rpc", func(c *gin.Context) {
		// Parse request body
		var request struct {
			Method string                 `json:"method" binding:"required"`
			Params map[string]interface{} `json:"params"`
			Target string                 `json:"target"` // Optional target process (defaults to "broker")
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"retcode": 400,
				"message": fmt.Sprintf("Invalid request: %v", err),
				"payload": nil,
			})
			return
		}

		// Default target to broker if not specified
		target := request.Target
		if target == "" {
			target = "broker"
		}

		// Forward RPC request to target process (use configured rpc timeout)
		response, err := rpcClient.InvokeRPCWithTarget(c.Request.Context(), target, request.Method, request.Params)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"retcode": 500,
				"message": fmt.Sprintf("RPC error: %v", err),
				"payload": nil,
			})
			return
		}

		// Check response type
		if response.Type == subprocess.MessageTypeError {
			c.JSON(http.StatusOK, gin.H{
				"retcode": 500,
				"message": response.Error,
				"payload": nil,
			})
			return
		}

		// Parse payload
		var payload interface{}
		if response.Payload != nil {
			if err := sonic.Unmarshal(response.Payload, &payload); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"retcode": 500,
					"message": fmt.Sprintf("Failed to parse response: %v", err),
					"payload": nil,
				})
				return
			}
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"retcode": 0,
			"message": "success",
			"payload": payload,
		})
	})
}

func TestRegisterHandlers_HealthEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	rg := r.Group("/api")
	registerHandlers(rg, nil, 0)

	// Perform request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/health", nil)
	r.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRegisterHandlers_RPCForwarding(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	rg := r.Group("/api")

	// Create a real RPCClient with a mocked subprocess
	mockSubprocess := &subprocess.Subprocess{
		ID: "mock-subprocess",
	}
	rpcClient := &RPCClient{
		sp:              mockSubprocess,
		pendingRequests: make(map[string]*requestEntry),
		rpcTimeout:      5 * time.Second,
	}
	registerHandlers(rg, rpcClient, 5)

	// Perform request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/rpc", nil)
	r.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// MockRPCClient is a mock implementation of RPCClient for testing
// Add methods as needed to simulate behavior
type MockRPCClient struct{}

func (m *MockRPCClient) InvokeRPCWithTarget(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	return &subprocess.Message{
		Type:    subprocess.MessageTypeResponse,
		Payload: []byte(`{"mock": "response"}`),
	}, nil
}

func (m *MockRPCClient) Run(ctx context.Context) error {
	return nil
}

// MockSubprocess is a mock implementation of subprocess.Subprocess for testing
// Add methods as needed to simulate behavior
type MockSubprocess struct {
	ID       string
	handlers map[string]subprocess.Handler
}

func (m *MockSubprocess) RegisterHandler(messageType string, handler subprocess.Handler) {
	if m.handlers == nil {
		m.handlers = make(map[string]subprocess.Handler)
	}
	m.handlers[messageType] = handler
}

func (m *MockSubprocess) Send(ctx context.Context, msg *subprocess.Message) error {
	return nil
}

func (m *MockSubprocess) Run(ctx context.Context) error {
	return nil
}
