package main

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/stretchr/testify/require"
)

func TestCreateAndDetectErrorResponse(t *testing.T) {
	req := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "test-request",
		Source:        "test-source",
		CorrelationID: "corr-1",
	}

	errMsg := createErrorResponse(req, "some-error")
	require.Equal(t, subprocess.MessageTypeError, errMsg.Type)
	require.Equal(t, "[meta] RPC error response: some-error", errMsg.Error)
	require.Equal(t, "test-request", errMsg.ID)
	require.Equal(t, "test-source", errMsg.Target)
	require.Equal(t, "corr-1", errMsg.CorrelationID)

	isErr, msg := isErrorResponse(errMsg)
	require.True(t, isErr)
	require.Equal(t, "[meta] RPC error response: some-error", msg)

	// Non-error message should not be detected
	nonErr := &subprocess.Message{
		Type: subprocess.MessageTypeResponse,
		ID:   "ok",
	}
	isErr2, msg2 := isErrorResponse(nonErr)
	require.False(t, isErr2)
	require.Equal(t, "", msg2)
}
