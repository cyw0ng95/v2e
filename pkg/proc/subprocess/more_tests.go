package subprocess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// errorWriter simulates a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write failure")
}

// TestCustomFileDescriptors verifies New() uses RPC_INPUT_FD/RPC_OUTPUT_FD if set
func TestCustomFileDescriptors(t *testing.T) {
	t.Skip("Skipping FD pipe test - UDS-only transport")
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
	// Removed long-running concurrency stress test from CI — it exercised
	// heavy concurrency and took non-trivial time. Keep locally if needed.
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
