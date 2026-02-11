package main

const (
	// Service Lifecycle Log Messages
	LogMsgWarningLoadingConfig  = "[ACCESS] Warning loading config: %v"
	LogMsgStartingAccessService = "[ACCESS] Starting access service on %s"
	LogMsgFailedStartServer     = "[ACCESS] Failed to start server: %v"
	LogMsgShuttingDown          = "[ACCESS] Shutting down access service..."
	LogMsgServerForcedShutdown  = "[ACCESS] Server forced to shutdown: %v"
	LogMsgServiceStopped        = "[ACCESS] Access service stopped"

	// Configuration Log Messages
	LogMsgConfigLoaded          = "[ACCESS] Config loaded successfully: RPC timeout=%ds, Shutdown timeout=%ds, Static dir=%s"
	LogMsgAddressConfigured     = "[ACCESS] Server address configured: %s"
	LogMsgRPCTimeoutConfigured  = "[ACCESS] RPC timeout configured: %ds"
	LogMsgShutdownTimeoutConfig = "[ACCESS] Shutdown timeout configured: %ds"
	LogMsgStaticDirConfigured   = "[ACCESS] Static directory configured: %s"
	LogMsgStaticDirEnvOverride  = "[ACCESS] Static directory overridden via environment: %s"

	// RPC Client Log Messages
	LogMsgStartingRPCClient        = "[ACCESS] Starting RPC client for process: %s"
	LogMsgRPCClientError           = "[ACCESS] RPC client error for process %s: %v"
	LogMsgRPCClientStopped         = "[ACCESS] RPC client stopped for process: %s"
	LogMsgRPCClientCreated         = "[ACCESS] RPC client created with timeout: %v"
	LogMsgRPCClientStarting        = "[ACCESS] Starting RPC client routine"
	LogMsgRPCClientStarted         = "[ACCESS] RPC client routine started"
	LogMsgRPCInvokeStarted         = "[ACCESS] RPC invoke started: target=%s, method=%s"
	LogMsgRPCInvokeCompleted       = "[ACCESS] RPC invoke completed: target=%s, method=%s"
	LogMsgRPCPendingRequestAdded   = "[ACCESS] Added pending RPC request: correlationID=%s"
	LogMsgRPCPendingRequestRemoved = "[ACCESS] Removed pending RPC request: correlationID=%s"
	LogMsgRPCResponseReceived      = "[ACCESS] Received RPC response: correlationID=%s, type=%s"
	LogMsgRPCChannelSignal         = "[ACCESS] Signaling RPC response channel: correlationID=%s"

	// Server Operations Log Messages
	LogMsgServerStarting          = "[ACCESS] Starting HTTP server on address: %s"
	LogMsgServerStarted           = "[ACCESS] HTTP server started successfully"
	LogMsgServerShutdownInitiated = "[ACCESS] Server shutdown initiated"
	LogMsgServerShutdownComplete  = "[ACCESS] Server shutdown completed"
	LogMsgServerShutdownForced    = "[ACCESS] Server shutdown forced"
	LogMsgHealthCheckReceived     = "[ACCESS] Health check endpoint called"
	LogMsgCORSMiddlewareAdded     = "[ACCESS] CORS middleware added to router"
	LogMsgRecoveryMiddlewareAdded = "[ACCESS] Recovery middleware added to router"

	// HTTP Request Handling Log Messages
	LogMsgHTTPRequestReceived  = "[ACCESS] HTTP request received: %s %s"
	LogMsgHTTPRequestProcessed = "[ACCESS] HTTP request processed: %s %s, status=%d"
	LogMsgRequestParsingError  = "[ACCESS] Error parsing request: %v"
	LogMsgRequestParamMissing  = "[ACCESS] Required request parameter missing: %v"

	// RPC Forwarding Log Messages
	LogMsgRPCForwardingStarted  = "[ACCESS] RPC forwarding request: method=%s, target=%s"
	LogMsgRPCForwardingParams   = "[ACCESS] RPC forwarding with params: %v"
	LogMsgRPCForwardingComplete = "[ACCESS] RPC forwarding completed: method=%s, target=%s"
	LogMsgRPCForwardingError    = "[ACCESS] RPC forwarding error: %v"
	LogMsgRPCResponseParsing    = "[ACCESS] Parsing RPC response payload"
	LogMsgRPCResponseParsed     = "[ACCESS] RPC response parsed successfully"
	LogMsgRPCResponseParseError = "[ACCESS] Error parsing RPC response: %v"

	// Static File Serving Log Messages
	LogMsgStaticFileServing  = "[ACCESS] Serving static files from directory: %s"
	LogMsgStaticFileNotFound = "[ACCESS] Static file not found, serving index.html for SPA: %s"
	LogMsgStaticFileServed   = "[ACCESS] Static file served: %s"
	LogMsgStaticDirNotFound  = "[ACCESS] Static directory not found, skipping static file serving: %s"
	LogMsgStaticFallbackSPA  = "[ACCESS] SPA fallback triggered, serving index.html for route: %s"

	// Rate Limiting Log Messages
	LogMsgRateLimiterStarted  = "[ACCESS] Rate limiter started: max_tokens=%d, refill_interval=%v"
	LogMsgRateLimitExceeded   = "[ACCESS] Rate limit exceeded for client: %s"
)
