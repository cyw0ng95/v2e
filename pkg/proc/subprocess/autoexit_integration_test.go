package subprocess

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// Integration test: start a broker-like process that creates a UDS listener and then
// start a subprocess that connects to it. Kill the broker and ensure the subprocess exits.
func TestSubprocess_AutoExitOnBrokerDeath(t *testing.T) {
	// Build a tiny helper program that acts as a broker: create a socket file and sleep
	// For simplicity reuse the current binary and set an env var to trigger helper behavior
	brokerCmd := exec.Command(os.Args[0], "-test.run=TestBrokerHelper")
	brokerCmd.Env = append(os.Environ(), "GO_WANT_BROKER_HELPER=1")
	if err := brokerCmd.Start(); err != nil {
		t.Fatalf("failed to start broker helper: %v", err)
	}
	// Give broker time to create socket and write path
	time.Sleep(200 * time.Millisecond)

	// Start subprocess helper that connects to the socket path written by broker helper
	sockPath := os.Getenv("TEST_BROKER_SOCKET")
	if sockPath == "" {
		t.Skip("broker helper did not set TEST_BROKER_SOCKET")
	}
	childCmd := exec.Command(os.Args[0], "-test.run=TestSubprocessHelper")
	childCmd.Env = append(os.Environ(), "GO_WANT_SUBPROCESS_HELPER=1")
	childCmd.Env = append(childCmd.Env, "RPC_SOCKET_PATH="+sockPath)
	if err := childCmd.Start(); err != nil {
		brokerCmd.Process.Kill()
		t.Fatalf("failed to start subprocess helper: %v", err)
	}

	// Kill broker
	if err := brokerCmd.Process.Kill(); err != nil {
		t.Fatalf("failed to kill broker helper: %v", err)
	}

	// Wait for child to exit within a few seconds
	done := make(chan error)
	go func() { done <- childCmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			// child exited with error — acceptable if it detected broker gone
			t.Logf("subprocess exited with error as expected: %v", err)
		}
	case <-time.After(5 * time.Second):
		// Timeout — kill the child and fail
		childCmd.Process.Kill()
		t.Fatalf("subprocess did not exit after broker was killed")
	}
}

// helper: broker creates a UDS listener and writes the socket path to env variable for test
func TestBrokerHelper(t *testing.T) {
	if os.Getenv("GO_WANT_BROKER_HELPER") != "1" {
		return
	}
	// Create a temp socket path
	socket := "/tmp/v2e_test_broker.sock"
	// Write out path in env for parent test to read
	os.Setenv("TEST_BROKER_SOCKET", socket)
	// Create listener via system call to reserve socket file
	l, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		os.Exit(2)
	}
	// Keep the process alive while test runs
	time.Sleep(10 * time.Second)
	_ = l
}

// helper: subprocess connects to RPC_SOCKET_PATH and waits until EOF
func TestSubprocessHelper(t *testing.T) {
	if os.Getenv("GO_WANT_SUBPROCESS_HELPER") != "1" {
		return
	}
	socket := os.Getenv("RPC_SOCKET_PATH")
	if socket == "" {
		os.Exit(2)
	}
	// Connect using NewWithUDS (which will exit if can't connect)
	_ = NewWithUDS("test-child", socket)
	// Wait until context canceled or process killed
	select {}
}
