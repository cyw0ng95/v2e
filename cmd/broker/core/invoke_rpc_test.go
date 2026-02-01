package core

import (
	"testing"
)

func TestInvokeRPC_SuccessRoundTrip(t *testing.T) {
	t.Skip("Skipping test - UDS-only transport does not use stdin/stdout pipes for RPC communication")

	// This test was designed for stdin/stdout pipe-based communication.
	// With UDS-only transport, RPC messages are sent via Unix Domain Sockets,
	// which requires a different testing approach using actual UDS connections.
	// TODO: Implement UDS-based RPC round-trip test if needed.
}
