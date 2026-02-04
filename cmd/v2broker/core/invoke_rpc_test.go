package core

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"
)

func TestInvokeRPC_SuccessRoundTrip(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestInvokeRPC_SuccessRoundTrip", nil, func(t *testing.T, tx *gorm.DB) {
		t.Skip("Skipping test - UDS-only transport does not use stdin/stdout pipes for RPC communication")

		// This test was designed for stdin/stdout pipe-based communication.
		// With UDS-only transport, RPC messages are sent via Unix Domain Sockets,
		// which requires a different testing approach using actual UDS connections.
		// TODO: Implement UDS-based RPC round-trip test if needed.
	})

}
