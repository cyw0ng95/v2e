package main

// Service lifecycle log constants
const (
	logWarningLoadingConfig  = "[ACCESS] Warning loading config: %v"
	logStartingAccessService = "[ACCESS] Starting access service on %s"
	logFailedStartServer     = "[ACCESS] Failed to start server: %v"
	logShuttingDown          = "[ACCESS] Shutting down access service..."
	logServerForcedShutdown  = "[ACCESS] Server forced to shutdown: %v"
	logServiceStopped        = "[ACCESS] Access service stopped"
)

// Configuration log constants
const (
	logConfigLoaded         = "[ACCESS] Config loaded successfully: RPC timeout=%ds, Shutdown timeout=%ds, Static dir=%s"
	logAddressConfigured    = "[ACCESS] Server address configured: %s"
	logRPCTimeoutConfigured = "[ACCESS] RPC timeout configured: %ds"
	logShutdownTimeoutCfg   = "[ACCESS] Shutdown timeout configured: %ds"
	logStaticDirConfigured  = "[ACCESS] Static directory configured: %s"
	logStaticDirEnvOverride = "[ACCESS] Static directory overridden via environment: %s"
)

// RPC client lifecycle log constants
const (
	logStartingRPCClient = "[ACCESS] Starting RPC client for process: %s"
	logRPCClientError    = "[ACCESS] RPC client error for process %s: %v"
	logRPCClientStopped  = "[ACCESS] RPC client stopped for process: %s"
	logRPCClientCreated  = "[ACCESS] RPC client created with timeout: %v"
	logRPCClientStarting = "[ACCESS] Starting RPC client routine"
	logRPCClientStarted  = "[ACCESS] RPC client routine started"
)

// RPC invocation log constants
const (
	logRPCInvokeStarted         = "[ACCESS] RPC invoke started: target=%s, method=%s"
	logRPCInvokeCompleted       = "[ACCESS] RPC invoke completed: target=%s, method=%s"
	logRPCPendingRequestAdded   = "[ACCESS] Added pending RPC request: correlationID=%s"
	logRPCPendingRequestRemoved = "[ACCESS] Removed pending RPC request: correlationID=%s"
	logRPCResponseReceived      = "[ACCESS] Received RPC response: correlationID=%s, type=%s"
	logRPCChannelSignal         = "[ACCESS] Signaling RPC response channel: correlationID=%s"
)

// Server operations log constants
const (
	logServerStarting          = "[ACCESS] Starting HTTP server on address: %s"
	logServerStarted           = "[ACCESS] HTTP server started successfully"
	logServerShutdownInitiated = "[ACCESS] Server shutdown initiated"
	logServerShutdownComplete  = "[ACCESS] Server shutdown completed"
	logServerShutdownForced    = "[ACCESS] Server shutdown forced"
	logHealthCheckReceived     = "[ACCESS] Health check endpoint called"
	logCORSMiddlewareAdded     = "[ACCESS] CORS middleware added to router"
	logRecoveryMiddlewareAdded = "[ACCESS] Recovery middleware added to router"
)

// HTTP request handling log constants
const (
	logHTTPRequestReceived = "[ACCESS] HTTP request received: %s %s"
	logHTTPRequestProcessed = "[ACCESS] HTTP request processed: %s %s, status=%d"
	logRequestParsingError = "[ACCESS] Error parsing request: %v"
)

// RPC forwarding log constants
const (
	logRPCForwardingStarted   = "[ACCESS] RPC forwarding request: method=%s, target=%s"
	logRPCForwardingParams    = "[ACCESS] RPC forwarding with params: %v"
	logRPCForwardingComplete  = "[ACCESS] RPC forwarding completed: method=%s, target=%s"
	logRPCForwardingError     = "[ACCESS] RPC forwarding error: %v"
	logRPCResponseParsing     = "[ACCESS] Parsing RPC response payload"
	logRPCResponseParsed      = "[ACCESS] RPC response parsed successfully"
	logRPCResponseParseError  = "[ACCESS] Error parsing RPC response: %v"
)

// Static file serving log constants
const (
	logStaticFileServing  = "[ACCESS] Serving static files from directory: %s"
	logStaticFileNotFound = "[ACCESS] Static file not found, serving index.html for SPA: %s"
	logStaticFileServed   = "[ACCESS] Static file served: %s"
	logStaticDirNotFound  = "[ACCESS] Static directory not found, skipping static file serving: %s"
	logStaticFallbackSPA  = "[ACCESS] SPA fallback triggered, serving index.html for route: %s"
)

// Rate limiting log constants
const (
	logRateLimiterStarted = "[ACCESS] Rate limiter started: max_tokens=%d, refill_interval=%v"
	logRateLimitExceeded  = "[ACCESS] Rate limit exceeded for client: %s"
)

// Public API - exported constants for backward compatibility
// Deprecated: Use the internal namespaced constants directly
const (
	// Service Lifecycle
	LogMsgWarningLoadingConfig  = logWarningLoadingConfig
	LogMsgStartingAccessService = logStartingAccessService
	LogMsgFailedStartServer     = logFailedStartServer
	LogMsgShuttingDown          = logShuttingDown
	LogMsgServerForcedShutdown  = logServerForcedShutdown
	LogMsgServiceStopped        = logServiceStopped

	// Configuration
	LogMsgConfigLoaded         = logConfigLoaded
	LogMsgAddressConfigured    = logAddressConfigured
	LogMsgRPCTimeoutConfigured = logRPCTimeoutConfigured
	LogMsgShutdownTimeoutConfig = logShutdownTimeoutCfg
	LogMsgStaticDirConfigured  = logStaticDirConfigured
	LogMsgStaticDirEnvOverride = logStaticDirEnvOverride

	// RPC Client
	LogMsgStartingRPCClient = logStartingRPCClient
	LogMsgRPCClientError    = logRPCClientError
	LogMsgRPCClientStopped  = logRPCClientStopped
	LogMsgRPCClientCreated  = logRPCClientCreated
	LogMsgRPCClientStarting = logRPCClientStarting
	LogMsgRPCClientStarted  = logRPCClientStarted

	// RPC Invocation
	LogMsgRPCInvokeStarted         = logRPCInvokeStarted
	LogMsgRPCInvokeCompleted       = logRPCInvokeCompleted
	LogMsgRPCPendingRequestAdded   = logRPCPendingRequestAdded
	LogMsgRPCPendingRequestRemoved = logRPCPendingRequestRemoved
	LogMsgRPCResponseReceived      = logRPCResponseReceived
	LogMsgRPCChannelSignal         = logRPCChannelSignal

	// Server Operations
	LogMsgServerStarting          = logServerStarting
	LogMsgServerStarted           = logServerStarted
	LogMsgServerShutdownInitiated = logServerShutdownInitiated
	LogMsgServerShutdownComplete  = logServerShutdownComplete
	LogMsgServerShutdownForced    = logServerShutdownForced
	LogMsgHealthCheckReceived     = logHealthCheckReceived
	LogMsgCORSMiddlewareAdded     = logCORSMiddlewareAdded
	LogMsgRecoveryMiddlewareAdded = logRecoveryMiddlewareAdded

	// HTTP Request Handling
	LogMsgHTTPRequestReceived  = logHTTPRequestReceived
	LogMsgHTTPRequestProcessed = logHTTPRequestProcessed
	LogMsgRequestParsingError  = logRequestParsingError

	// RPC Forwarding
	LogMsgRPCForwardingStarted   = logRPCForwardingStarted
	LogMsgRPCForwardingParams    = logRPCForwardingParams
	LogMsgRPCForwardingComplete  = logRPCForwardingComplete
	LogMsgRPCForwardingError     = logRPCForwardingError
	LogMsgRPCResponseParsing     = logRPCResponseParsing
	LogMsgRPCResponseParsed      = logRPCResponseParsed
	LogMsgRPCResponseParseError  = logRPCResponseParseError

	// Static File Serving
	LogMsgStaticFileServing  = logStaticFileServing
	LogMsgStaticFileNotFound = logStaticFileNotFound
	LogMsgStaticFileServed   = logStaticFileServed
	LogMsgStaticDirNotFound  = logStaticDirNotFound
	LogMsgStaticFallbackSPA  = logStaticFallbackSPA

	// Rate Limiting
	LogMsgRateLimiterStarted = logRateLimiterStarted
	LogMsgRateLimitExceeded  = logRateLimitExceeded
)
