package main

const (
	// Service Lifecycle Log Messages
	LogMsgFailedToSetupLogging    = "[sysmon] Failed to setup logging: %v"
	LogMsgServiceStarted          = "[sysmon] Sysmon service started"
	LogMsgServiceReady            = "[sysmon] Sysmon service ready and accepting requests"
	LogMsgServiceShutdownStarting = "[sysmon] Sysmon service shutdown starting"
	LogMsgServiceShutdownComplete = "[sysmon] Sysmon service shutdown completed"

	// Configuration Log Messages
	LogMsgProcessIDConfigured    = "[sysmon] Process ID configured: %s"
	LogMsgBootstrapLoggerCreated = "[sysmon] Bootstrap logger created"
	LogMsgLoggingSetupComplete   = "[sysmon] Logging setup completed with level: %v"

	// Component Initialization Log Messages
	LogMsgSubprocessCreated    = "[sysmon] Subprocess created with ID: %s"
	LogMsgRPCHandlerRegistered = "[sysmon] RPC handler registered: %s"
	LogMsgRegisteredSysMetrics = "[sysmon] Registered RPCGetSysMetrics handler"
	LogMsgRPCClientCreated     = "[sysmon] RPC client created for broker communication"

	// RPC Handler Log Messages
	LogMsgSysMetricsInvoked     = "[sysmon] RPCGetSysMetrics handler invoked. msg.ID=%s, correlation_id=%s"
	LogMsgStartingGetSysMetrics = "[sysmon] Starting RPCGetSysMetrics request, correlation_id=%s"
	LogMsgGetSysMetricsSuccess  = "[sysmon] RPCGetSysMetrics completed successfully, correlation_id=%s"
	LogMsgGetSysMetricsFailed   = "[sysmon] RPCGetSysMetrics failed, correlation_id=%s, error: %v"
	LogMsgFailedMarshalMetrics  = "[sysmon] Failed to marshal metrics: %v"
	LogMsgReturningMetrics      = "[sysmon] Returning metrics response. correlation_id=%s"

	// Metric Collection Log Messages
	LogMsgFailedCollectMetrics        = "[sysmon] Failed to collect metrics: %v"
	LogMsgMetricCollectionStarted     = "[sysmon] Metric collection started"
	LogMsgMetricCollectionCompleted   = "[sysmon] Metric collection completed"
	LogMsgCollectedMetrics            = "[sysmon] Collected metrics: cpu_usage=%.2f, memory_usage=%.2f, load_avg=%v, uptime=%.0fs"
	LogMsgPerformanceMetricsCollected = "[sysmon] Performance metrics collected"
	LogMsgHealthCheckPerformed        = "[sysmon] Health check performed"

	// System Resource Metrics Log Messages
	LogMsgCPUUsageCollected        = "[sysmon] CPU usage collected: %.2f%%"
	LogMsgMemoryUsageCollected     = "[sysmon] Memory usage collected: %.2f%%"
	LogMsgLoadAvgCollected         = "[sysmon] Load average collected: %v"
	LogMsgUptimeCollected          = "[sysmon] Uptime collected: %.0fs"
	LogMsgDiskUsageCollected       = "[sysmon] Disk usage collected: used=%d, total=%d"
	LogMsgSwapUsageCollected       = "[sysmon] Swap usage collected: %d"
	LogMsgNetworkUsageCollected    = "[sysmon] Network usage collected: rx=%d, tx=%d"
	LogMsgResourceThresholdCrossed = "[sysmon] Resource threshold crossed: %s, value=%.2f"

	// ProcFS Operations Log Messages
	LogMsgProcFSReadStarted = "[sysmon] ProcFS read started: %s"
	LogMsgProcFSReadSuccess = "[sysmon] ProcFS read completed: %s"
	LogMsgProcFSReadFailed  = "[sysmon] ProcFS read failed: %s, error: %v"

	// System Monitoring Log Messages
	LogMsgSystemResourceMonitorStarted = "[sysmon] System resource monitoring started"
	LogMsgSystemResourceMonitorStopped = "[sysmon] System resource monitoring stopped"

	// RPC Communication Log Messages
	LogMsgRPCInvokeStarted         = "[sysmon] RPC invoke started: target=%s, method=%s"
	LogMsgRPCInvokeCompleted       = "[sysmon] RPC invoke completed: target=%s, method=%s"
	LogMsgRPCResponseReceived      = "[sysmon] RPC response received: correlationID=%s, type=%s"
	LogMsgRPCPendingRequestAdded   = "[sysmon] Added pending RPC request: correlationID=%s"
	LogMsgRPCPendingRequestRemoved = "[sysmon] Removed pending RPC request: correlationID=%s"
	LogMsgRPCChannelSignaled       = "[sysmon] RPC response channel signaled: correlationID=%s"

	// Broker Statistics Log Messages
	LogMsgFailedFetchBrokerStats  = "[sysmon] Failed to fetch message stats from broker: %v"
	LogMsgFailedUnmarshalStats    = "[sysmon] Failed to unmarshal broker message stats: %v"
	LogMsgBrokerStatsFetchStarted = "[sysmon] Broker message stats fetch started"
	LogMsgBrokerStatsFetchSuccess = "[sysmon] Broker message stats fetch succeeded"
	LogMsgBrokerStatsFetchFailed  = "[sysmon] Broker message stats fetch failed: %v"

	// Marshaling Operations Log Messages
	LogMsgMetricsMarshalingStarted = "[sysmon] Metrics marshaling started"
	LogMsgMetricsMarshalingSuccess = "[sysmon] Metrics marshaling completed successfully"
	LogMsgMetricsMarshalingFailed  = "[sysmon] Metrics marshaling failed: %v"

	// Error Handling Log Messages
	LogMsgRPCClientPanicRecovered = "[sysmon] Recovered from panic during RPC call to broker: %v"
	LogMsgBrokerUnavailable       = "[sysmon] Broker unavailable, skipping message stats fetch"
)
