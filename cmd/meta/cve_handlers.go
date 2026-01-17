package main

import (
	"context"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createErrorResponse creates a properly formatted error response message
func createErrorResponse(msg *subprocess.Message, errorMsg string) *subprocess.Message {
	return &subprocess.Message{
		Type:          subprocess.MessageTypeError,
		ID:            msg.ID,
		Error:         errorMsg,
		CorrelationID: msg.CorrelationID,
		Target:        msg.Source,
	}
}

// isErrorResponse checks if an RPC response is an error and returns the error if so
func isErrorResponse(response *subprocess.Message) (bool, string) {
	if response.Type == subprocess.MessageTypeError {
		return true, response.Error
	}
	return false, ""
}

// createGetCVEHandler creates a handler that retrieves CVE data
func createGetCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createCreateCVEHandler creates a handler that fetches CVE from NVD and saves locally
func createCreateCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createUpdateCVEHandler creates a handler that refetches CVE from NVD and updates local storage
func createUpdateCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createDeleteCVEHandler creates a handler that deletes CVE from local storage
func createDeleteCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createListCVEsHandler creates a handler that lists CVEs from local storage
func createListCVEsHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createCountCVEsHandler creates a handler that counts CVEs in local storage
func createCountCVEsHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createStartSessionHandler creates a handler that starts a new job session
func createStartSessionHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createStopSessionHandler creates a handler that stops the current session
func createStopSessionHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createGetSessionStatusHandler creates a handler that returns the current session status
func createGetSessionStatusHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createPauseJobHandler creates a handler that pauses the running job
func createPauseJobHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}

// createResumeJobHandler creates a handler that resumes a paused job
func createResumeJobHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Handler logic...
	}
}
