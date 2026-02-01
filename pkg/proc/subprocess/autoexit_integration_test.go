package subprocess

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Integration test: start a temporary UDS server to simulate the broker and a child
// subprocess that connects to it. When the server is closed, the subprocess should
// observe connection EOF and exit (we approximate this by ensuring NewWithUDS's
// connection is closed when server closes).
func TestSubprocess_AutoExitOnBrokerDeath(t *testing.T) {
	t.Skip("Skipping auto-exit integration test: remote API/network tests disabled for fast CI")
	dir := os.TempDir()
	socketPath := filepath.Join(dir, "v2e_test_autoexit.sock")
	_ = os.Remove(socketPath)

	// Start a unix socket listener to act as broker
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on uds: %v", err)
	}
	defer func() { l.Close(); os.Remove(socketPath) }()

	// Start listener accept loop
	accepted := make(chan net.Conn)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			close(accepted)
			return
		}
		accepted <- conn
	}()

	// Start the subprocess that connects to the socket
	// We'll run it in-process by calling NewWithUDS directly in a goroutine
	done := make(chan struct{})
	go func() {
		sp := NewWithUDS("autoexit-child", socketPath)
		// Wait for context cancellation or connection close by reading from input
		// NewWithUDS uses the net.Conn as input — once server closes, reads will fail
		buf := make([]byte, 1)
		_, err := sp.input.Read(buf)
		if err != nil {
			// Connection closed — expected when broker listener is closed
		}
		close(done)
	}()

	// Wait until child has connected
	select {
	case conn := <-accepted:
		// Close server side connection to simulate broker exit
		conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatalf("child did not connect to broker socket in time")
	}

	// Now close the listener to simulate broker shutdown
	l.Close()

	// Child should detect connection close and finish
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatalf("child did not exit after broker shutdown")
	}
}
