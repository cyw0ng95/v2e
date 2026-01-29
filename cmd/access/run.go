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
		common.Warn(LogMsgWarningLoadingConfig, err)
		os.Exit(1)
	}

	common.Info(LogMsgConfigLoaded, config.Access.RPCTimeoutSeconds, config.Access.ShutdownTimeoutSeconds, config.Access.StaticDir)

	// Configure service timeouts and static dir from config (with defaults)
	rpcTimeout := 30 * time.Second
	if config.Access.RPCTimeoutSeconds > 0 {
		rpcTimeout = time.Duration(config.Access.RPCTimeoutSeconds) * time.Second
		common.Info(LogMsgRPCTimeoutConfigured, config.Access.RPCTimeoutSeconds)
	}

	shutdownTimeout := 10 * time.Second
	if config.Access.ShutdownTimeoutSeconds > 0 {
		shutdownTimeout = time.Duration(config.Access.ShutdownTimeoutSeconds) * time.Second
		common.Info(LogMsgShutdownTimeoutConfig, config.Access.ShutdownTimeoutSeconds)
	}

	staticDir := "website"
	if config.Access.StaticDir != "" {
		staticDir = config.Access.StaticDir
		common.Info(LogMsgStaticDirConfigured, staticDir)
	}
	// Allow broker to override static dir via environment when running as subprocess
	if envStatic := os.Getenv("ACCESS_STATIC_DIR"); envStatic != "" {
		staticDir = envStatic
		common.Info(LogMsgStaticDirEnvOverride, staticDir)
	}

	// Set default address if not configured
	address := "0.0.0.0:8080"
	if config.Server.Address != "" {
		address = config.Server.Address
	}
	common.Info(LogMsgAddressConfigured, address)

	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "access"
	}

	// Create RPC client for broker communication (use configured rpc timeout)
	rpcClient := NewRPCClient(processID, rpcTimeout)
	common.Info(LogMsgRPCClientCreated, rpcTimeout)

	// Start RPC client in background
	go func() {
		common.Info(LogMsgRPCClientStarting)
		common.Info(LogMsgStartingRPCClient, processID)
		if err := rpcClient.Run(context.Background()); err != nil {
			common.Warn(LogMsgRPCClientError, processID, err)
		} else {
			common.Info(LogMsgRPCClientStopped, processID)
		}
	}()
	common.Info(LogMsgRPCClientStarted)

	// Setup router and server
	router := setupRouter(rpcClient, int(rpcTimeout.Seconds()), staticDir)

	// Create HTTP server
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		common.Info(LogMsgServerStarting, address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Error(LogMsgFailedStartServer, err)
			return
		}
		common.Info(LogMsgServerStarted)
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	common.Info("[ACCESS] Waiting for shutdown signal...")
	<-quit

	common.Info(LogMsgShuttingDown)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	common.Info(LogMsgServerShutdownInitiated)
	if err := srv.Shutdown(ctx); err != nil {
		common.Warn(LogMsgServerForcedShutdown, err)
		common.Info(LogMsgServerShutdownForced)
		os.Exit(1)
	}
	common.Info(LogMsgServerShutdownComplete)

	common.Info(LogMsgServiceStopped)
}
