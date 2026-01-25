package main

const (
	// Log messages
	LogMsgWarningLoadingConfig    = "[ACCESS] Warning loading config: %v"
	LogMsgStartingRPCClient       = "[ACCESS] Starting RPC client for process: %s"
	LogMsgRPCClientError          = "[ACCESS] RPC client error for process %s: %v"
	LogMsgRPCClientStopped        = "[ACCESS] RPC client stopped for process: %s"
	LogMsgStartingAccessService   = "[ACCESS] Starting access service on %s"
	LogMsgFailedStartServer       = "[ACCESS] Failed to start server: %v"
	LogMsgShuttingDown            = "[ACCESS] Shutting down access service..."
	LogMsgServerForcedShutdown    = "[ACCESS] Server forced to shutdown: %v"
	LogMsgServiceStopped          = "[ACCESS] Access service stopped"
)
