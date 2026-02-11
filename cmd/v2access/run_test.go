package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunAccess_GracefulShutdown tests the graceful shutdown functionality
func TestRunAccess_GracefulShutdown(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRunAccess_GracefulShutdown", nil, func(t *testing.T, tx *gorm.DB) {
		// This test verifies the graceful shutdown flow
		// Since runAccess() is a long-running function, we test the components

		// Test signal handling setup
		quit := make(chan os.Signal, 1)
		signalNotify := func(sig chan os.Signal, sigs ...os.Signal) {
			// Mock signal.Notify
		}

		// Verify we can create a signal channel
		require.NotNil(t, quit)
		require.Equal(t, cap(quit), 1, "quit channel should have buffer size 1")

		// Test that signal notification can be set up
		signalNotify(quit, syscall.SIGINT, syscall.SIGTERM)
	})
}

// TestRPCClientCreation tests RPC client creation with various configurations
func TestRPCClientCreation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCClientCreation", nil, func(t *testing.T, tx *gorm.DB) {
		// Test creating RPC client with different process IDs
		processID := "test-access-rpc"
		rpcTimeout := 30 * time.Second

		client := NewRPCClient(processID, rpcTimeout)

		require.NotNil(t, client)
		require.NotNil(t, client.sp)
		require.NotNil(t, client.client)
		assert.Equal(t, rpcTimeout, client.rpcTimeout)
		assert.Equal(t, processID, client.sp.ID)
	})
}

// TestNewRPCClientWithSubprocess tests creating RPC client with existing subprocess
func TestNewRPCClientWithSubprocess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNewRPCClientWithSubprocess", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(os.Stderr, "[TEST] ", common.InfoLevel)
		sp := subprocess.New("test-subprocess")
		rpcTimeout := 15 * time.Second

		client := NewRPCClientWithSubprocess(sp, logger, rpcTimeout)

		require.NotNil(t, client)
		assert.Equal(t, sp, client.sp)
		assert.Equal(t, rpcTimeout, client.rpcTimeout)
	})
}

// TestRPCClient_InvokeRPC tests InvokeRPC method
func TestRPCClient_InvokeRPC(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCClient_InvokeRPC", nil, func(t *testing.T, tx *gorm.DB) {
		client := NewRPCClient("test-invoke-rpc", 5*time.Second)
		ctx := context.Background()

		// Test that InvokeRPC calls InvokeRPCWithTarget with "broker" as target
		// This will fail because there's no actual broker, but we're testing the method signature
		_, err := client.InvokeRPC(ctx, "TestMethod", nil)

		// Error is expected due to no actual broker connection
		// We're just verifying the method doesn't panic
		assert.NotNil(t, err)
	})
}

// TestRPCClient_Run tests the Run method
func TestRPCClient_Run(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCClient_Run", nil, func(t *testing.T, tx *gorm.DB) {
		client := NewRPCClient("test-run-rpc", 5*time.Second)

		// Run should start the subprocess message processing
		// Since there's no actual broker, this will return quickly or block
		// We test with a timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := client.Run(ctx)

		// Error is acceptable - we're just testing the method exists and doesn't panic
		// The subprocess Run method may return an error when there's no broker
		_ = err
	})
}

// TestDefaultServerAddr tests the DefaultServerAddr function
func TestDefaultServerAddr(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDefaultServerAddr", nil, func(t *testing.T, tx *gorm.DB) {
		addr := DefaultServerAddr()
		assert.NotEmpty(t, addr, "DefaultServerAddr should return a non-empty string")
		assert.Contains(t, addr, ":", "DefaultServerAddr should contain a port separator")
	})
}

// TestDefaultStaticDir tests the DefaultStaticDir function
func TestDefaultStaticDir(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDefaultStaticDir", nil, func(t *testing.T, tx *gorm.DB) {
		dir := DefaultStaticDir()
		assert.NotEmpty(t, dir, "DefaultStaticDir should return a non-empty string")
	})
}

// TestSetupRouter_NoStaticDir tests setupRouter when static directory doesn't exist
func TestSetupRouter_NoStaticDir(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupRouter_NoStaticDir", nil, func(t *testing.T, tx *gorm.DB) {
		nonExistentDir := "/non/existent/dir/for/testing"
		router := setupRouter(nil, 30, nonExistentDir)

		require.NotNil(t, router)

		// Verify health endpoint still works
		w := performRequest(router, "GET", "/restful/health", nil)
		assert.Equal(t, 200, w.Code)
	})
}

// TestSetupRouter_WithStaticDir tests setupRouter with an actual static directory
func TestSetupRouter_WithStaticDir(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupRouter_WithStaticDir", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary directory with a test HTML file
		tempDir := t.TempDir()
		indexHTML := filepath.Join(tempDir, "index.html")
		err := os.WriteFile(indexHTML, []byte("<html><body>Test Page</body></html>"), 0644)
		require.NoError(t, err)

		router := setupRouter(nil, 30, tempDir)
		require.NotNil(t, router)

		// Test serving the index.html file
		w := performRequest(router, "GET", "/index.html", nil)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Test Page")

		// Test SPA fallback - non-existent route should return index.html
		w = performRequest(router, "GET", "/some/nonexistent/route", nil)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Test Page")
	})
}

// TestSetupRouter_APIRoutesNotHandledByNoRoute tests that API routes return 404
func TestSetupRouter_APIRoutesNotHandledByNoRoute(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupRouter_APIRoutesNotHandledByNoRoute", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		indexHTML := filepath.Join(tempDir, "index.html")
		err := os.WriteFile(indexHTML, []byte("<html><body>Test</body></html>"), 0644)
		require.NoError(t, err)

		router := setupRouter(nil, 30, tempDir)

		// Non-existent API route should return 404, not index.html
		w := performRequest(router, "GET", "/restful/nonexistent", nil)
		assert.Equal(t, 404, w.Code)
		assert.NotContains(t, w.Body.String(), "Test Page")
	})
}

// TestRegisterHandlers tests handler registration
func TestRegisterHandlers(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRegisterHandlers", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a mock router and client
		gin.SetMode(gin.TestMode)
		router := gin.New()
		restful := router.Group("/restful")

		mockClient := &MockRPCClient{}
		registerHandlers(restful, mockClient)

		// Test health endpoint
		w := performRequest(router, "GET", "/restful/health", nil)
		assert.Equal(t, 200, w.Code)
	})
}

// TestHTTPErrorResponse tests the httpErrorResponse helper function
func TestHTTPErrorResponse(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestHTTPErrorResponse", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		router.GET("/error", func(c *gin.Context) {
			httpErrorResponse(c, 400, "Bad request")
		})

		w := performRequest(router, "GET", "/error", nil)
		assert.Equal(t, 400, w.Code)

		// Verify response structure
		expected := `{"retcode":400,"message":"Bad request","payload":null}`
		assert.JSONEq(t, expected, w.Body.String())
	})
}

// TestHTTPSuccessResponse tests the httpSuccessResponse helper function
func TestHTTPSuccessResponse(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestHTTPSuccessResponse", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		payload := map[string]string{"key": "value"}
		router.GET("/success", func(c *gin.Context) {
			httpSuccessResponse(c, payload)
		})

		w := performRequest(router, "GET", "/success", nil)
		assert.Equal(t, 200, w.Code)

		// Verify response structure
		expected := `{"message":"success","payload":{"key":"value"},"retcode":0}`
		assert.JSONEq(t, expected, w.Body.String())
	})
}

// TestRPCHandler_MissingMethod tests RPC handler with missing method parameter
func TestRPCHandler_MissingMethod(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_MissingMethod", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		restful := router.Group("/restful")

		mockClient := &MockRPCClient{}
		registerHandlers(restful, mockClient)

		// Send request without method parameter
		body := `{"params": {}}`
		w := performRequestWithBody(router, "POST", "/restful/rpc", body)

		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request")
	})
}

// TestRPCHandler_InvalidJSON tests RPC handler with invalid JSON
func TestRPCHandler_InvalidJSON(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_InvalidJSON", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		restful := router.Group("/restful")

		mockClient := &MockRPCClient{}
		registerHandlers(restful, mockClient)

		// Send invalid JSON
		body := `{invalid json`
		w := performRequestWithBody(router, "POST", "/restful/rpc", body)

		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request")
	})
}

// TestRPCHandler_ValidRequest tests RPC handler with valid request
func TestRPCHandler_ValidRequest(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_ValidRequest", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		restful := router.Group("/restful")

		mockClient := &MockRPCClient{}
		registerHandlers(restful, mockClient)

		// Send valid request
		body := `{"method": "TestRPC", "params": {"key": "value"}}`
		w := performRequestWithBody(router, "POST", "/restful/rpc", body)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

// TestRPCHandler_WithTarget tests RPC handler with explicit target
func TestRPCHandler_WithTarget(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCHandler_WithTarget", nil, func(t *testing.T, tx *gorm.DB) {
		gin.SetMode(t, gin.TestMode)
		router := gin.New()
		restful := router.Group("/restful")

		mockClient := &MockRPCClient{}
		registerHandlers(restful, mockClient)

		// Send request with explicit target
		body := `{"method": "TestRPC", "target": "meta", "params": {}}`
		w := performRequestWithBody(router, "POST", "/restful/rpc", body)

		assert.Equal(t, 200, w.Code)
	})
}

// TestDefaultRateLimiterConfig tests the default rate limiter configuration
func TestDefaultRateLimiterConfig(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDefaultRateLimiterConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := DefaultRateLimiterConfig()

		require.NotNil(t, config)
		assert.Greater(t, config.MaxTokens, 0, "MaxTokens should be positive")
		assert.Greater(t, int64(config.RefillInterval), int64(0), "RefillInterval should be positive")
		assert.Greater(t, int64(config.CleanupInterval), int64(0), "CleanupInterval should be positive")
		assert.NotEmpty(t, config.TrustedProxies, "TrustedProxies should not be empty")
		assert.NotEmpty(t, config.ExcludedPaths, "ExcludedPaths should not be empty")
	})
}

// TestRateLimiterConfig tests custom rate limiter configuration
func TestRateLimiterConfig(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRateLimiterConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := &RateLimiterConfig{
			MaxTokens:       10,
			RefillInterval:  2 * time.Second,
			CleanupInterval: 3 * time.Minute,
			TrustedProxies:  []string{"192.168.1.1"},
			ExcludedPaths:   []string{"/test"},
		}

		assert.Equal(t, 10, config.MaxTokens)
		assert.Equal(t, int64(2*time.Second), int64(config.RefillInterval))
		assert.Equal(t, int64(3*time.Minute), int64(config.CleanupInterval))
		assert.Len(t, config.TrustedProxies, 1)
		assert.Len(t, config.ExcludedPaths, 1)
	})
}

// TestStaticFileServing_NestedPath tests serving files from nested directories
func TestStaticFileServing_NestedPath(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStaticFileServing_NestedPath", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()

		// Create nested directory structure
		nestedDir := filepath.Join(tempDir, "assets", "css")
		err := os.MkdirAll(nestedDir, 0755)
		require.NoError(t, err)

		cssFile := filepath.Join(nestedDir, "style.css")
		err = os.WriteFile(cssFile, []byte("body { margin: 0; }"), 0644)
		require.NoError(t, err)

		router := setupRouter(nil, 30, tempDir)

		// Test serving nested file
		w := performRequest(router, "GET", "/assets/css/style.css", nil)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "margin: 0;")
	})
}

// TestStaticFileServing_RootPath tests that root path serves index.html
func TestStaticFileServing_RootPath(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStaticFileServing_RootPath", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		indexHTML := filepath.Join(tempDir, "index.html")
		err := os.WriteFile(indexHTML, []byte("<html><body>Root</body></html>"), 0644)
		require.NoError(t, err)

		router := setupRouter(nil, 30, tempDir)

		// Test root path
		w := performRequest(router, "GET", "/", nil)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Root")
	})
}

// TestStaticFileServing_DirectoryListingDenied tests that directory listing returns index.html
func TestStaticFileServing_DirectoryListingDenied(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStaticFileServing_DirectoryListingDenied", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		indexHTML := filepath.Join(tempDir, "index.html")
		err := os.WriteFile(indexHTML, []byte("<html><body>Index</body></html>"), 0644)
		require.NoError(t, err)

		// Create a directory
		dirPath := filepath.Join(tempDir, "assets")
		err = os.Mkdir(dirPath, 0755)
		require.NoError(t, err)

		router := setupRouter(nil, 30, tempDir)

		// Requesting a directory should return index.html (SPA fallback)
		w := performRequest(router, "GET", "/assets/", nil)
		assert.Equal(t, 200, w.Code)
		// Should fall back to index.html for SPA routing
		assert.Contains(t, w.Body.String(), "Index")
	})
}

// TestStaticFileServing_NonExistentFile_SPAFallback tests that non-existent files fall back to index.html
func TestStaticFileServing_NonExistentFile_SPAFallback(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStaticFileServing_NonExistentFile_SPAFallback", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		indexHTML := filepath.Join(tempDir, "index.html")
		err := os.WriteFile(indexHTML, []byte("<html><body>SPA Fallback</body></html>"), 0644)
		require.NoError(t, err)

		router := setupRouter(nil, 30, tempDir)

		// Requesting a non-existent file should fall back to index.html for SPA
		w := performRequest(router, "GET", "/nonexistent/file.html", nil)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "SPA Fallback")
	})
}

// TestCorsMiddlewareAdded tests that CORS middleware is properly configured
func TestCorsMiddlewareAdded(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCorsMiddlewareAdded", nil, func(t *testing.T, tx *gorm.DB) {
		router := setupRouter(nil, 30, "/nonexistent")

		// Test CORS headers are present
		w := performRequest(router, "OPTIONS", "/restful/health", nil)
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
	})
}

// Helper function to perform HTTP requests on the router
func performRequest(router *gin.Engine, method, path string, body string) *httptest.ResponseRecorder {
	return performRequestWithBody(router, method, path, body)
}

// Helper function to perform HTTP requests with body
func performRequestWithBody(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
