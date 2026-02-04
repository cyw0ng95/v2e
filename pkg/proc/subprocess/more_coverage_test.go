package subprocess

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestNewInvalidFDsFallback ensures New falls back to stdio and logs a warning
// when RPC_INPUT_FD/RPC_OUTPUT_FD are invalid.
func TestNewInvalidFDsFallback(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewInvalidFDsFallback", nil, func(t *testing.T, tx *gorm.DB) {
		// This test ensures New() falls back to stdio. No environment flags are used
		// for configuration in the new design, so nothing to unset.

		sp := New("fd-fallback")

		// Ensure fallback to stdio when fds are not usable
		if sp.input != os.Stdin || sp.output != os.Stdout {
			t.Fatalf("expected stdio fallback, got input=%T output=%T", sp.input, sp.output)
		}
	})

}

// TestSendResponseNilPayload verifies SendResponse works with nil payloads
func TestSendResponseNilPayload(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSendResponseNilPayload", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

// TestSendEventNilPayload verifies SendEvent works with nil payloads
func TestSendEventNilPayload(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSendEventNilPayload", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}

// TestFlushBatchJoinPath exercises the join buffer path used for moderate-sized batches
func TestFlushBatchJoinPath(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestFlushBatchJoinPath", nil, func(t *testing.T, tx *gorm.DB) {
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
	})

}
