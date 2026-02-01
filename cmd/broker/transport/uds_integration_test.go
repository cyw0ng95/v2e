package transport

import (
	"fmt"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// TestUDSDeterministicPath verifies that TransportManager's registered UDS
// socket path matches the subprocess deterministic path so a subprocess
// using the deterministic path will compute the same socket file name.
func TestUDSDeterministicPath(t *testing.T) {
	tm := NewTransportManager()
	// Ensure transport manager uses the subprocess default base path
	tm.SetUdsBasePath(subprocess.DefaultProcUDSBasePath())

	procID := "test-uds-svc"

	// Register a UDS transport as server; this creates the listener at the
	// deterministic path derived from the base path and process ID.
	socketPath, err := tm.RegisterUDSTransport(procID, true)
	if err != nil {
		t.Fatalf("failed to register uds transport: %v", err)
	}
	if socketPath == "" {
		t.Fatalf("expected socketPath to be set")
	}

	// Construct the path subprocess would compute
	expected := fmt.Sprintf("%s_%s.sock", subprocess.DefaultProcUDSBasePath(), procID)
	if socketPath != expected {
		t.Fatalf("socket path mismatch: got=%s want=%s", socketPath, expected)
	}

	// Cleanup: remove registered transport
	tm.UnregisterTransport(procID)
}
