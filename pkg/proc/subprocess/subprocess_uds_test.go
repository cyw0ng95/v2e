package subprocess

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewWithUDS_ConnectSuccess creates a temporary UDS listener and ensures NewWithUDS connects successfully.
func TestNewWithUDS_ConnectSuccess(t *testing.T) {
	dir := os.TempDir()
	socketPath := filepath.Join(dir, "v2e_test_sock.sock")
	// Ensure old socket removed
	os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on uds socket: %v", err)
	}
	defer func() {
		l.Close()
		os.Remove(socketPath)
	}()

	// Accept in background
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			// keep the connection open briefly
			go func(c net.Conn) { time.Sleep(500 * time.Millisecond); c.Close() }(conn)
		}
	}()

	sp := NewWithUDS("test-uds", socketPath)
	if sp == nil {
		t.Fatalf("NewWithUDS returned nil subprocess")
	}
	// Ensure input/output are set
	if sp.input == nil || sp.output == nil {
		t.Fatalf("expected input/output to be set for UDS subprocess")
	}
}
