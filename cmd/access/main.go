package main

import (
"context"
"fmt"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/cyw0ng95/v2e/pkg/common"
"github.com/gin-gonic/gin"
)

const (
// DefaultRPCTimeout is the default timeout for RPC requests
DefaultRPCTimeout = 30 * time.Second
// DefaultShutdownTimeout is the default timeout for graceful shutdown
DefaultShutdownTimeout = 10 * time.Second
)

func main() {
// Load configuration
config, err := common.LoadConfig("config.json")
if err != nil {
fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
os.Exit(1)
}

// Set default address if not configured
address := "0.0.0.0:8080"
if config.Server.Address != "" {
address = config.Server.Address
}

// Set up logger
common.SetLevel(common.InfoLevel)

// Set Gin mode
gin.SetMode(gin.ReleaseMode)

// Create Gin router
router := gin.Default()

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
// Request body: {"method": "RPCMethodName", "params": {...}}
// Response: {"retcode": 0, "message": "success", "payload": {...}}
restful.POST("/rpc", func(c *gin.Context) {
// Parse request body
var request struct {
Method string                 `json:"method" binding:"required"`
Params map[string]interface{} `json:"params"`
}

if err := c.ShouldBindJSON(&request); err != nil {
c.JSON(http.StatusBadRequest, gin.H{
"retcode": 400,
"message": fmt.Sprintf("Invalid request: %v", err),
"payload": nil,
})
return
}

// TODO: Forward RPC request to broker
// When RPC forwarding is implemented (issue #74), this will:
// 1. Create RPC message with request.Method and request.Params
// 2. Send to broker via stdin
// 3. Wait for response from broker via stdout
// 4. Return response in standardized format

// For now, handle known RPC methods with placeholder data
var retcode int
var message string
var payload interface{}

switch request.Method {
case "RPCGetMessageStats":
retcode = 0
message = "success"
payload = gin.H{
"total_sent":         0,
"total_received":     0,
"request_count":      0,
"response_count":     0,
"event_count":        0,
"error_count":        0,
"first_message_time": nil,
"last_message_time":  nil,
}
case "RPCGetMessageCount":
retcode = 0
message = "success"
payload = gin.H{
"count": 0,
}
default:
retcode = 404
message = fmt.Sprintf("Unknown RPC method: %s", request.Method)
payload = nil
}

c.JSON(http.StatusOK, gin.H{
"retcode": retcode,
"message": message,
"payload": payload,
})
})

// TODO: Process management endpoints removed
// Access service needs to be redesigned to communicate with broker via RPC
// instead of creating its own broker instance.
// The access service is a subprocess and should not instantiate broker.
// See: https://github.com/cyw0ng95/v2e/pull/74
}

// Create HTTP server
srv := &http.Server{
Addr:    address,
Handler: router,
}

// Start server in a goroutine
go func() {
common.Info("Starting access service on %s", address)
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
common.Error("Failed to start server: %v", err)
os.Exit(1)
}
}()

// Wait for interrupt signal to gracefully shutdown the server
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

common.Info("Shutting down access service...")

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
common.Error("Server forced to shutdown: %v", err)
os.Exit(1)
}

common.Info("Access service stopped")
}
