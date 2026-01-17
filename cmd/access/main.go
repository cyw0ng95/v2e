/*
Package main implements the access RPC service.

RPC API Specification:

Access Service
====================

Service Type: REST (HTTP/JSON)
Description: RESTful API gateway service that provides external access to the v2e system.

	Forwards RPC requests to backend services through the broker.

Available REST Endpoints:
-------------------------

 1. GET /restful/health
    Description: Health check endpoint to verify service is running
    Request Parameters: None
    Response:
    - status (string): "ok" if service is healthy
    Errors: None
    Example:
    Request:  GET /restful/health
    Response: {"status": "ok"}

 2. POST /restful/rpc
    Description: Generic RPC forwarding endpoint that routes requests to backend services
    Request Parameters:
    - method (string, required): RPC method name (e.g., "RPCGetCVE")
    - params (object, optional): Parameters to pass to the RPC method
    - target (string, optional): Target process ID (default: "broker")
    Response:
    - retcode (int): 0 for success, non-zero for errors
    - message (string): Success message or error description
    - payload (object): Response data from the backend service
    Errors:
    - Invalid JSON: retcode=400, missing or malformed request body
    - RPC timeout: retcode=500, backend service did not respond in time
    - Backend error: retcode=500, backend service returned an error
    Example:
    Request:  {"method": "RPCGetCVE", "target": "cve-meta", "params": {"cve_id": "CVE-2021-44228"}}
    Response: {"retcode": 0, "message": "success", "payload": {"id": "CVE-2021-44228", ...}}

Notes:
------
- All RPC requests are forwarded through the broker for security and routing
- Default RPC timeout is 30 seconds
- Service runs as a subprocess managed by the broker
- Uses stdin/stdout for RPC communication with broker
- External clients access via HTTP on configured address (default: 0.0.0.0:8080)
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	// DefaultRPCTimeout is the default timeout for RPC requests
	DefaultRPCTimeout = 30 * time.Second
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second
)

// RPCClient handles RPC communication with the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]chan *subprocess.Message
	mu              sync.RWMutex
	correlationSeq  uint64
	// per-client RPC timeout (configurable)
	rpcTimeout time.Duration
}

// NewRPCClient creates a new RPC client for broker communication
func NewRPCClient(processID string, rpcTimeout time.Duration) *RPCClient {
	sp := subprocess.New(processID)
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]chan *subprocess.Message),
		rpcTimeout:      rpcTimeout,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)

	return client
}

// handleResponse handles response messages from the broker
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Look up the pending request
	c.mu.Lock()
	respChan, exists := c.pendingRequests[msg.CorrelationID]
	if exists {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()

	if exists {
		select {
		case respChan <- msg:
		case <-time.After(1 * time.Second):
			// Timeout sending to channel
		}
	}

	return nil, nil // Don't send another response
}

// handleError handles error messages from the broker (treat them as responses)
func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Error messages are also valid responses
	return c.handleResponse(ctx, msg)
}

// InvokeRPC invokes an RPC method on the broker and waits for response
func (c *RPCClient) InvokeRPC(ctx context.Context, method string, params interface{}) (*subprocess.Message, error) {
	return c.InvokeRPCWithTarget(ctx, "broker", method, params)
}

// InvokeRPCWithTarget invokes an RPC method on a specific target process and waits for response
func (c *RPCClient) InvokeRPCWithTarget(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	// Generate correlation ID
	c.mu.Lock()
	c.correlationSeq++
	correlationID := fmt.Sprintf("access-rpc-%d", c.correlationSeq)
	c.mu.Unlock()

	// Create response channel
	respChan := make(chan *subprocess.Message, 1)

	// Register pending request
	c.mu.Lock()
	c.pendingRequests[correlationID] = respChan
	c.mu.Unlock()

	// Clean up on exit
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		close(respChan)
	}()

	// Create request message
	var payload json.RawMessage
	if params != nil {
		data, err := sonic.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		payload = data
	}

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            method,
		Payload:       payload,
		Target:        target,
		CorrelationID: correlationID,
	}

	// Send request to broker
	if err := c.sp.SendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}

	// Wait for response with timeout (use configured rpcTimeout)
	select {
	case response := <-respChan:
		return response, nil
	case <-time.After(c.rpcTimeout):
		return nil, fmt.Errorf("RPC timeout waiting for response")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Run starts the RPC client message processing
func (c *RPCClient) Run(ctx context.Context) error {
	return c.sp.Run()
}

func main() {
	// Load configuration
	config, err := common.LoadConfig("config.json")
	if err != nil {
		common.Error("Error loading config: %v", err)
		os.Exit(1)
	}

	// Configure service timeouts and static dir from config (with defaults)
	rpcTimeout := 30 * time.Second
	if config.Access.RPCTimeoutSeconds > 0 {
		rpcTimeout = time.Duration(config.Access.RPCTimeoutSeconds) * time.Second
	}

	shutdownTimeout := 10 * time.Second
	if config.Access.ShutdownTimeoutSeconds > 0 {
		shutdownTimeout = time.Duration(config.Access.ShutdownTimeoutSeconds) * time.Second
	}

	staticDir := "website"
	if config.Access.StaticDir != "" {
		staticDir = config.Access.StaticDir
	}

	// Set default address if not configured
	address := "0.0.0.0:8080"
	if config.Server.Address != "" {
		address = config.Server.Address
	}

	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "access"
	}

	// Create RPC client for broker communication (use configured rpc timeout)
	rpcClient := NewRPCClient(processID, rpcTimeout)

	// Start RPC client in background
	go func() {
		if err := rpcClient.Run(context.Background()); err != nil {
			common.Error("RPC client error: %v", err)
		}
	}()

	// Set Gin mode to release (minimal logging)
	gin.SetMode(gin.ReleaseMode)

	// Disable Gin's default logger to prevent stdout pollution
	gin.DefaultWriter = os.Stderr
	gin.DefaultErrorWriter = os.Stderr

	// Create Gin router without default middleware
	router := gin.New()
	// Add recovery middleware but log to stderr
	router.Use(gin.RecoveryWithWriter(os.Stderr))

	// Add CORS middleware
	router.Use(cors.Default())

	// Create RESTful API group
	restful := router.Group("/restful")
	{
		// Health check endpoint
		restful.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})

		// Generic RPC forwarding endpoint
		// POST /restful/rpc
		// Request body: {"method": "RPCMethodName", "params": {...}, "target": "process-id"}
		// Response: {"retcode": 0, "message": "success", "payload": {...}}
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
			ctx, cancel := context.WithTimeout(c.Request.Context(), rpcTimeout)
			defer cancel()

			response, err := rpcClient.InvokeRPCWithTarget(ctx, target, request.Method, request.Params)
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

	// Serve static files from configured staticDir if present (Next.js static export)
	outDir := staticDir
	if _, err := os.Stat(outDir); err == nil {
		absOut, _ := filepath.Abs(outDir)
		common.Info("[ACCESS] Serving static files from %s", absOut)

		// Use NoRoute to serve files and fallback to index.html for SPA routes.
		// Avoid registering a catch-all route which conflicts with existing API prefixes.
		router.NoRoute(func(c *gin.Context) {
			// Do not handle API routes here
			if strings.HasPrefix(c.Request.URL.Path, "/restful") {
				c.JSON(http.StatusNotFound, gin.H{"retcode": 404, "message": "not found", "payload": nil})
				return
			}

			// Clean requested path and map to filesystem
			reqPath := filepath.Clean(c.Request.URL.Path)
			if reqPath == "." || reqPath == "/" {
				c.File(filepath.Join(outDir, "index.html"))
				return
			}

			relPath := strings.TrimPrefix(reqPath, "/")
			fullPath := filepath.Join(outDir, relPath)
			if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
				c.File(fullPath)
				return
			}

			// Fallback to index.html for SPA routing
			c.File(filepath.Join(outDir, "index.html"))
		})
	} else {
		common.Info("[ACCESS] Static dir %s not found, skipping static serving", outDir)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		common.Info("[ACCESS] Starting access service on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Error("[ACCESS] Failed to start server: %v", err)
			return
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	common.Info("[ACCESS] Shutting down access service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		common.Error("[ACCESS] Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	common.Info("[ACCESS] Access service stopped\n")
}
