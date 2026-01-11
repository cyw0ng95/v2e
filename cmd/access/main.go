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
