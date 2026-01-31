package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// runAccess contains the original implementation of main moved here for maintainability
func runAccess() {
	// Use common startup utility to standardize initialization
	configStruct := subprocess.StandardStartupConfig{
		DefaultProcessID: "access",
		LogPrefix:        "[ACCESS] ",
	}
	sp, logger := subprocess.StandardStartup(configStruct)

	// Load configuration
	config, err := common.LoadConfig("config.json")
	if err != nil {
		common.Warn(LogMsgWarningLoadingConfig, err)
		os.Exit(1)
	}

	logger.Info(LogMsgConfigLoaded, config.Access.RPCTimeoutSeconds, config.Access.ShutdownTimeoutSeconds, config.Access.StaticDir)

	// Configure service timeouts and static dir from config (with defaults)
	rpcTimeout := 30 * time.Second
	if config.Access.RPCTimeoutSeconds > 0 {
		rpcTimeout = time.Duration(config.Access.RPCTimeoutSeconds) * time.Second
		logger.Info(LogMsgRPCTimeoutConfigured, config.Access.RPCTimeoutSeconds)
	}

	shutdownTimeout := 10 * time.Second
	if config.Access.ShutdownTimeoutSeconds > 0 {
		shutdownTimeout = time.Duration(config.Access.ShutdownTimeoutSeconds) * time.Second
		logger.Info(LogMsgShutdownTimeoutConfig, config.Access.ShutdownTimeoutSeconds)
	}

	staticDir := "website"
	if config.Access.StaticDir != "" {
		staticDir = config.Access.StaticDir
		logger.Info(LogMsgStaticDirConfigured, staticDir)
	}
	// Allow broker to override static dir via environment when running as subprocess
	if envStatic := os.Getenv("ACCESS_STATIC_DIR"); envStatic != "" {
		staticDir = envStatic
		logger.Info(LogMsgStaticDirEnvOverride, staticDir)
	}

	// Set default address if not configured
	address := "0.0.0.0:8080"
	if config.Server.Address != "" {
		address = config.Server.Address
	}
	logger.Info(LogMsgAddressConfigured, address)

	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "access"
	}

	// Create RPC client for broker communication (use configured rpc timeout)
	rpcClient := NewRPCClientWithSubprocess(sp, logger, rpcTimeout)
	logger.Info(LogMsgRPCClientCreated, rpcTimeout)

	// Start RPC client in background
	go func() {
		logger.Info(LogMsgRPCClientStarting)
		logger.Info(LogMsgStartingRPCClient, processID)
		if err := rpcClient.Run(context.Background()); err != nil {
			logger.Warn(LogMsgRPCClientError, processID, err)
		} else {
			logger.Info(LogMsgRPCClientStopped, processID)
		}
	}()
	logger.Info(LogMsgRPCClientStarted)

	// Setup router and server
	router := setupRouter(rpcClient, int(rpcTimeout.Seconds()), staticDir)

	// Create HTTP server
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(LogMsgServerStarting, address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(LogMsgFailedStartServer, err)
			return
		}
		logger.Info(LogMsgServerStarted)
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("[ACCESS] Waiting for shutdown signal...")
	<-quit

	logger.Info(LogMsgShuttingDown)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info(LogMsgServerShutdownInitiated)
	if err := srv.Shutdown(ctx); err != nil {
		logger.Warn(LogMsgServerForcedShutdown, err)
		logger.Info(LogMsgServerShutdownForced)
		os.Exit(1)
	}
	logger.Info(LogMsgServerShutdownComplete)

	logger.Info(LogMsgServiceStopped)
}
