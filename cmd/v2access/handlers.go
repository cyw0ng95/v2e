package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// contextPool pools unused context objects for reuse in RPC handlers.
// This reduces allocations for the frequently created timeout contexts.
// Each pooled context includes its cancel function for proper cleanup.
var contextPool = sync.Pool{
	New: func() interface{} {
		// Return a new pooledContext struct - will be initialized on Get()
		return &pooledContext{}
	},
}

// HTTP response helpers for reducing boilerplate in handlers

// httpErrorResponse sends an error response with given code and message.
func httpErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"retcode": code,
		"message": message,
		"payload": nil,
	})
}

// httpSuccessResponse sends a success response with given payload.
func httpSuccessResponse(c *gin.Context, payload interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"retcode": 0,
		"message": "success",
		"payload": payload,
	})
}

// registerHandlers registers REST endpoints on provided router group
func registerHandlers(restful *gin.RouterGroup, rpcClient *RPCClient) {
	// Health check endpoint
	restful.GET("/health", func(c *gin.Context) {
		common.Debug(LogMsgHealthCheckReceived)
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Generic RPC forwarding endpoint
	restful.POST("/rpc", func(c *gin.Context) {
		common.Debug(LogMsgHTTPRequestReceived, c.Request.Method, c.Request.URL.Path)

		// Parse request body
		var request struct {
			Method string                 `json:"method" binding:"required"`
			Params map[string]interface{} `json:"params"`
			Target string                 `json:"target"` // Optional target process (defaults to "broker")
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			common.Warn(LogMsgRequestParsingError, err)
			httpErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
			common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, http.StatusBadRequest)
			return
		}

		// Default target to broker if not specified
		target := request.Target
		if target == "" {
			target = "broker"
		}

		common.Info(LogMsgRPCForwardingStarted, request.Method, target)
		if request.Params != nil {
			common.Debug(LogMsgRPCForwardingParams, request.Params)
		}

		// Forward RPC request to target process (use configured rpc timeout)
		requestCtx := c.Request.Context()
		common.Info(LogMsgRPCInvokeStarted, target, request.Method)
		common.Debug("RPC request context value: %v", requestCtx)

		// Check if context is already done before making RPC call
		select {
		case <-requestCtx.Done():
			err := requestCtx.Err()
			common.Error("HTTP request context already canceled before RPC call: %v", err)
			// Use appropriate status code for context cancellation
			statusCode := http.StatusRequestTimeout
			if err == context.Canceled {
				statusCode = http.StatusServiceUnavailable
			}
			httpErrorResponse(c, statusCode, fmt.Sprintf("Request context canceled: %v", err))
			return
		default:
			// Context is not done, proceed with RPC
		}

		// Create a separate context for the RPC call to avoid cancellation from HTTP context
		// This prevents the RPC call from being canceled when the HTTP client disconnects
		// Use sync.Pool to reduce allocations for frequently created timeout contexts
		rpcCtx, cancel, pc := getPooledContext(rpcClient.rpcTimeout)
		defer cancel()
		defer putPooledContext(pc)

		response, err := rpcClient.InvokeRPCWithTarget(rpcCtx, target, request.Method, request.Params)
		common.Debug(LogMsgRPCInvokeCompleted, target, request.Method)

		// Log context state after RPC call completes
		select {
		case <-requestCtx.Done():
			err := requestCtx.Err()
			common.Warn("HTTP request context canceled after RPC call: %v", err)
		default:
			// Context is still active
		}

		if err != nil {
			common.Error(LogMsgRPCForwardingError, err)
			httpErrorResponse(c, http.StatusOK, fmt.Sprintf("RPC error: %v", err))
			common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, 200)
			return
		}

		// Check response type using subprocess helper
		if isError, errMsg := subprocess.IsErrorResponse(response); isError {
			common.Warn("RPC response is an error: %s", errMsg)
			httpErrorResponse(c, http.StatusOK, errMsg)
			common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, 200)
			return
		}

		// Parse payload
		var payload interface{}
		if response.Payload != nil {
			common.Debug(LogMsgRPCResponseParsing)
			if err := subprocess.UnmarshalFast(response.Payload, &payload); err != nil {
				common.Error(LogMsgRPCResponseParseError, err)
				httpErrorResponse(c, http.StatusOK, fmt.Sprintf("Failed to parse response: %v", err))
				common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, 200)
				return
			}
			common.Debug(LogMsgRPCResponseParsed)
		}

		// Return success response
		httpSuccessResponse(c, payload)
		common.Info(LogMsgRPCForwardingComplete, request.Method, target)
		common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, http.StatusOK)
	})
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

// pooledContext holds a context with its cancel function for pooling
type pooledContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// getPooledContext retrieves or creates a new timeout context from pool
// Returns the context, its cancel function, and the pooled wrapper for cleanup
func getPooledContext(timeout time.Duration) (ctx context.Context, cancel context.CancelFunc, pc *pooledContext) {
	// Get a pooled context object
	pc = contextPool.Get().(*pooledContext)

	// Create a new context with timeout
	// Note: We cannot reuse the context itself because canceled contexts
	// are not safe to reuse. However, we reuse the pooledContext struct
	// to reduce allocations of the wrapper struct.
	pc.ctx, pc.cancel = context.WithTimeout(context.Background(), timeout)
	return pc.ctx, pc.cancel, pc
}

// putPooledContext returns the pooled wrapper to the pool for reuse
func putPooledContext(pc *pooledContext) {
	if pc != nil {
		// Reset the pooled struct before returning to pool
		// Note: The context is already canceled by the deferred cancel() call
		pc.ctx = nil
		pc.cancel = nil
		contextPool.Put(pc)
	}
}
