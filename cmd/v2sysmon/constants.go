package main

const (
	// Service Lifecycle Log Messages
	LogMsgServiceStarted          = "[sysmon] Sysmon service started"
	LogMsgServiceReady            = "[sysmon] Sysmon service ready and accepting requests"
	LogMsgServiceShutdownStarting = "[sysmon] Sysmon service shutdown starting"
	LogMsgServiceShutdownComplete = "[sysmon] Sysmon service shutdown completed"

	// Component Initialization Log Messages
	LogMsgRPCHandlerRegistered = "[sysmon] RPC handler registered: %s"
	LogMsgRegisteredSysMetrics = "[sysmon] Registered RPCGetSysMetrics handler"
	LogMsgRPCClientCreated     = "[sysmon] RPC client created for broker communication"

	// RPC Handler Log Messages
	LogMsgSysMetricsInvoked     = "[sysmon] RPCGetSysMetrics handler invoked. msg.ID=%s, correlation_id=%s"
	LogMsgStartingGetSysMetrics = "[sysmon] Starting RPCGetSysMetrics request, correlation_id=%s"
	LogMsgGetSysMetricsSuccess  = "[sysmon] RPCGetSysMetrics completed successfully, correlation_id=%s"
	LogMsgGetSysMetricsFailed   = "[sysmon] RPCGetSysMetrics failed, correlation_id=%s, error: %v"
	LogMsgReturningMetrics      = "[sysmon] Returning metrics response. correlation_id=%s"

	// Metric Collection Log Messages
	LogMsgFailedCollectMetrics      = "[sysmon] Failed to collect metrics: %v"
	LogMsgMetricCollectionStarted   = "[sysmon] Metric collection started"
	LogMsgMetricCollectionCompleted = "[sysmon] Metric collection completed"
	LogMsgCollectedMetrics           = "[sysmon] Collected metrics: cpu_usage=%.2f, memory_usage=%.2f, load_avg=%v, uptime=%.0fs"

	// Broker Statistics Log Messages
	LogMsgFailedFetchBrokerStats  = "[sysmon] Failed to fetch message stats from broker: %v"
	LogMsgFailedUnmarshalStats    = "[sysmon] Failed to unmarshal broker message stats: %v"
	LogMsgBrokerStatsFetchStarted = "[sysmon] Broker message stats fetch started"
	LogMsgBrokerStatsFetchSuccess = "[sysmon] Broker message stats fetch succeeded"
	LogMsgBrokerStatsFetchFailed  = "[sysmon] Broker message stats fetch failed: %v"

	// Marshaling Operations Log Messages
	LogMsgMetricsMarshalingStarted = "[sysmon] Metrics marshaling started"
	LogMsgMetricsMarshalingSuccess = "[sysmon] Metrics marshaling completed successfully"

	// Error Handling Log Messages
	LogMsgRPCClientPanicRecovered = "[sysmon] Recovered from panic during RPC call to broker: %v"
	LogMsgBrokerUnavailable       = "[sysmon] Broker unavailable, skipping message stats fetch"
)
