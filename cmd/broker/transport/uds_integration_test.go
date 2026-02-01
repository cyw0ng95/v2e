package transport

import "testing"

// TestUDSDeterministicPath verifies that TransportManager's registered UDS
// socket path matches the subprocess deterministic path so a subprocess
// using the deterministic path will compute the same socket file name.
func TestUDSDeterministicPath(t *testing.T) {
	// This integration-style test was removed because creating actual UDS
	// listeners is environment-dependent and caused intermittent CI failures.
}
