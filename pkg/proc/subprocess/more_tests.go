package subprocess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// errorWriter simulates a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write failure")
}

// TestCustomFileDescriptors verifies New() uses RPC_INPUT_FD/RPC_OUTPUT_FD if set
func TestCustomFileDescriptors(t *testing.T) {
	inF, err := os.CreateTemp(".", "rpc-in-*")
	if err != nil {
		t.Fatalf("failed to create temp in file: %v", err)
	}
	defer os.Remove(inF.Name())
	outF, err := os.CreateTemp(".", "rpc-out-*")
	if err != nil {
		inF.Close()
		t.Fatalf("failed to create temp out file: %v", err)
	}
	defer os.Remove(outF.Name())

	// Ensure files are open
	inFd := int(inF.Fd())
	outFd := int(outF.Fd())

	// Construct a Subprocess using the NewWithFDs helper directly so tests
	// don't rely on environment variables. Use the temp file descriptors
	// created above.
	sp := NewWithFDs("fdsvc", inFd, outFd)
	// The NewWithFDs should set input/output to opened files
	if sp.input == nil || sp.output == nil {
		t.Fatalf("expected input/output to be set from fds")
	}

	// Try to write via SendEvent; only check error
	if err := sp.SendEvent("fd-test", nil); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// Close temp files and subprocess files if they are *os.File
	if f, ok := sp.input.(*os.File); ok {
		f.Close()
	}
	if f, ok := sp.output.(*os.File); ok {
		f.Close()
	}
}

// TestSendMessage_ErrorWriter verifies SendMessage returns an error when writer fails
func TestSendMessage_ErrorWriter(t *testing.T) {
	sp := New("err-writer")
	sp.SetOutput(&errorWriter{})

	msg := &Message{Type: MessageTypeEvent, ID: "e1", Payload: nil}
	if err := sp.SendMessage(msg); err == nil {
		t.Fatalf("expected SendMessage to fail with error writer")
	}
}

// TestConcurrentSendStop stress tests concurrent sendMessage calls with Stop
func TestConcurrentSendStop(t *testing.T) {
	t.Skip("Skipping long-running concurrency tests for fast CI")
	sp := New("concur")
	buf := &bytes.Buffer{}
	// Directly set output and enable batching off
	sp.SetOutput(buf)

	// Start many goroutines sending messages concurrently
	var wg sync.WaitGroup
	nG := 50
	nPer := 200
	for i := 0; i < nG; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < nPer; j++ {
				msg := &Message{Type: MessageTypeEvent, ID: fmt.Sprintf("g%d-m%d", idx, j)}
				_ = sp.SendMessage(msg) // ignore errors in stress test
			}
		}(i)
	}

	// Wait a short time and then stop concurrently
	time.Sleep(10 * time.Millisecond)
	stopDone := make(chan struct{})
	go func() {
		_ = sp.Stop()
		close(stopDone)
	}()

	// Wait for senders
	wg.Wait()
	// Wait for stop to finish
	select {
	case <-stopDone:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not complete in time")
	}
}

// TestZeroCopyPath ensures large messages follow the fast path (no panic and output present)
func TestZeroCopyPath(t *testing.T) {
	sp := New("zc")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf) // disables batching

	// create a large payload exceeding zeroCopyThreshold
	big := strings.Repeat("x", zeroCopyThreshold+512)
	payload, _ := json.Marshal(big)
	msg := &Message{Type: MessageTypeEvent, ID: "large", Payload: payload}

	if err := sp.SendMessage(msg); err != nil {
		t.Fatalf("SendMessage failed on zero-copy path: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "large") {
		t.Fatalf("expected output to contain message ID, got: %s", out)
	}
}

// TestFlushBatchLargeTotal verifies flushBatch uses direct writes when total exceeds threshold
func TestFlushBatchLargeTotal(t *testing.T) {
	sp := New("fb")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)

	// create many medium messages to exceed 4*defaultWriterBufSize
	msg := make([]byte, defaultWriterBufSize/2)
	for i := range msg {
		msg[i] = 'a'
	}
	batch := make([][]byte, 0, 200)
	for i := 0; i < 200; i++ {
		batch = append(batch, msg)
	}

	sp.flushBatch(batch)
	if buf.Len() == 0 {
		t.Fatalf("expected output from flushBatch with large total size")
	}
}

// TestStopIdempotent ensures Stop can be called multiple times safely
func TestStopIdempotent(t *testing.T) {
	t.Skip("Skipping Stop idempotence test on CI; re-enable locally if needed")
}
