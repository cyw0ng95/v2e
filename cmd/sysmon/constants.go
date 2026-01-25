package main

const (
	// Log messages
	LogMsgFailedToSetupLogging      = "[sysmon] Failed to setup logging: %v"
	LogMsgRegisteredSysMetrics      = "[sysmon] Registered RPCGetSysMetrics handler"
	LogMsgServiceStarted            = "[sysmon] Sysmon service started"
	LogMsgSysMetricsInvoked         = "[sysmon] RPCGetSysMetrics handler invoked. msg.ID=%s, correlation_id=%s"
	LogMsgFailedCollectMetrics      = "[sysmon] Failed to collect metrics: %v"
	LogMsgFailedFetchBrokerStats    = "[sysmon] Failed to fetch message stats from broker: %v"
	LogMsgFailedUnmarshalStats      = "[sysmon] Failed to unmarshal broker message stats: %v"
	LogMsgCollectedMetrics          = "[sysmon] Collected metrics: cpu_usage=%.2f, memory_usage=%.2f, load_avg=%v, uptime=%.0fs"
	LogMsgFailedMarshalMetrics      = "[sysmon] Failed to marshal metrics: %v"
	LogMsgReturningMetrics          = "[sysmon] Returning metrics response. correlation_id=%s"
)
