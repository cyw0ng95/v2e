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

// Message statistics endpoints
// TODO: These currently return mock data. When RPC forwarding is implemented
// in the access service (issue #74), these should forward requests to the
// broker's RPCGetMessageStats and RPCGetMessageCount handlers.

// Get message statistics from broker
restful.GET("/stats/messages", func(c *gin.Context) {
// TODO: Forward RPC request to broker's RPCGetMessageStats handler
// For now, return placeholder response for integration testing
c.JSON(http.StatusOK, gin.H{
"total_sent":       0,
"total_received":   0,
"request_count":    0,
"response_count":   0,
"event_count":      0,
"error_count":      0,
"first_message_time": nil,
"last_message_time":  nil,
"note": "This endpoint will be fully functional when RPC forwarding is implemented (issue #74)",
})
})

// Get message count from broker
restful.GET("/stats/message-count", func(c *gin.Context) {
// TODO: Forward RPC request to broker's RPCGetMessageCount handler
// For now, return placeholder response for integration testing
c.JSON(http.StatusOK, gin.H{
"count": 0,
"note": "This endpoint will be fully functional when RPC forwarding is implemented (issue #74)",
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
