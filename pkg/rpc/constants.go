package rpc

const (
	// Log Messages
	LogMsgFailedMarshalParams  = "failed to marshal params: %w"
	LogMsgFailedSendRequest    = "failed to send RPC request: %w"
	LogMsgRPCTimeout           = "RPC timeout waiting for response from %s"
	LogMsgFoundPendingRequest  = "Found pending request for correlation ID: %s, signaling response"
	LogMsgPendingRequestNotFound = "No pending request found for correlation ID: %s"
	LogMsgFoundPendingError    = "Found pending error for correlation ID: %s, signaling error"
	LogMsgSendingRPCRequest    = "Sending RPC request to %s: method=%s, correlation_id=%s"
	LogMsgReceivedResponse     = "Received response for correlation ID: %s"
	LogMsgReceivedError        = "Received error for correlation ID: %s"

	// Correlation ID Format
	CorrelationIDFormat = "rpc-%s-%d-%d"
)
