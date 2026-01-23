package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// runAccess contains the original implementation of main moved here for maintainability
func runAccess() {
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
	// Allow broker to override static dir via environment when running as subprocess
	if envStatic := os.Getenv("ACCESS_STATIC_DIR"); envStatic != "" {
		staticDir = envStatic
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

	// Setup router and server
	router := setupRouter(rpcClient, int(rpcTimeout.Seconds()), staticDir)

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
