package subprocess

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

// TestNewInvalidFDsFallback ensures New falls back to stdio and logs a warning
// when RPC_INPUT_FD/RPC_OUTPUT_FD are invalid.
func TestNewInvalidFDsFallback(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
		w.Close()
		r.Close()
	}()

	os.Setenv("RPC_INPUT_FD", "not-a-number")
	os.Setenv("RPC_OUTPUT_FD", "also-bad")
	defer func() {
		os.Unsetenv("RPC_INPUT_FD")
		os.Unsetenv("RPC_OUTPUT_FD")
	}()

	sp := New("fd-fallback")
	// close the writer so we can read
	_ = w.Close()
	out, _ := io.ReadAll(r)
	stderrStr := string(out)

	if !strings.Contains(stderrStr, "Warning") {
		t.Fatalf("expected warning on stderr, got: %s", stderrStr)
	}

	// Ensure fallback to stdio
	if sp.input != os.Stdin || sp.output != os.Stdout {
		t.Fatalf("expected stdio fallback, got input=%T output=%T", sp.input, sp.output)
	}
}

// TestSendResponseNilPayload verifies SendResponse works with nil payloads
func TestSendResponseNilPayload(t *testing.T) {
	sp := New("resp-nil")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)

	if err := sp.SendResponse("rid-nil", nil); err != nil {
		t.Fatalf("SendResponse returned error: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatal("expected output from SendResponse")
	}

	// Ensure payload is omitted (since nil)
	if strings.Contains(out, "payload") {
		t.Fatalf("did not expect payload field in output: %s", out)
	}
}

// TestSendEventNilPayload verifies SendEvent works with nil payloads
func TestSendEventNilPayload(t *testing.T) {
	sp := New("evt-nil")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)

	if err := sp.SendEvent("eid-nil", nil); err != nil {
		t.Fatalf("SendEvent returned error: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatal("expected output from SendEvent")
	}
}

// TestSendMessage_ContextDone verifies sendMessage returns context error when context is canceled
func TestSendMessage_ContextDone(t *testing.T) {
	sp := New("ctx-test")
	// Ensure batching is enabled (do not call SetOutput which disables it)
	// Cancel the context
	sp.cancel()

	err := sp.sendMessage(&Message{Type: MessageTypeEvent, ID: "x"})
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

// TestFlushBatchJoinPath exercises the join buffer path used for moderate-sized batches
func TestFlushBatchJoinPath(t *testing.T) {
	sp := New("join-path")
	buf := &bytes.Buffer{}
	sp.SetOutput(buf)

	batch := make([][]byte, 0, 20)
	for i := 0; i < 12; i++ {
		batch = append(batch, []byte("msg"))
	}

	sp.flushBatch(batch)
	if buf.Len() == 0 {
		t.Fatalf("expected output from flushBatch join path")
	}

	// Second flush should also work (exercise buffer pool reuse)
	buf.Reset()
	sp.flushBatch(batch)
	if buf.Len() == 0 {
		t.Fatalf("expected output on second flushBatch call")
	}
}
