package subprocess

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

func TestNewDefaults(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNewDefaults", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("svc1")
		if sp == nil {
			t.Fatal("New returned nil")
		}
		if sp.ID != "svc1" {
			t.Fatalf("expected ID svc1, got %s", sp.ID)
		}
		if sp.outChan == nil {
			t.Fatal("outChan should be initialized")
		}
	})

}

func TestSetOutputAndSendMessage_Disabled(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetOutputAndSendMessage_Disabled", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-out")
		buf := &bytes.Buffer{}
		sp.SetOutput(buf)

		msg := &Message{Type: MessageTypeEvent, ID: "evt-1", Payload: nil}
		if err := sp.SendMessage(msg); err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		got := buf.String()
		if !strings.Contains(got, "evt-1") || !strings.Contains(got, "event") {
			t.Fatalf("unexpected output: %s", got)
		}
	})

}

func TestSendMessage_BatchedWriter(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendMessage_BatchedWriter", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-batch")
		buf := &bytes.Buffer{}
		sp.SetInput(strings.NewReader(""))
		// enable batching (SetOutput disables it), so set output directly
		sp.mu.Lock()
		sp.output = buf
		sp.disableBatching = false
		sp.mu.Unlock()

		// start writer
		sp.wg.Add(1)
		go sp.messageWriter()

		msg := &Message{Type: MessageTypeEvent, ID: "b1", Payload: nil}
		for i := 0; i < 50; i++ {
			if err := sp.sendMessage(msg); err != nil {
				t.Fatalf("sendMessage failed: %v", err)
			}
		}

		// stop and wait for writer to flush
		if err := sp.Stop(); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		// Wait up to 500ms for writer to flush to buffer (avoid flakiness)
		deadline := time.After(500 * time.Millisecond)
		for {
			out := buf.String()
			if bytes.Count([]byte(out), []byte("\n")) > 0 {
				break
			}
			select {
			case <-deadline:
				t.Fatalf("expected output lines, got none (buffer empty)")
			default:
				time.Sleep(5 * time.Millisecond)
			}
		}
	})

}

func TestSendMessage_BatchedWriter_ImmediateClose(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendMessage_BatchedWriter_ImmediateClose", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-batch-close")
		buf := &bytes.Buffer{}
		sp.SetInput(strings.NewReader(""))
		sp.mu.Lock()
		sp.output = buf
		sp.disableBatching = false
		sp.mu.Unlock()

		sp.wg.Add(1)
		go sp.messageWriter()

		msg := &Message{Type: MessageTypeEvent, ID: "close1", Payload: nil}
		// send fewer than batch size and immediately stop
		for i := 0; i < 5; i++ {
			if err := sp.sendMessage(msg); err != nil {
				t.Fatalf("sendMessage failed: %v", err)
			}
		}

		if err := sp.Stop(); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		out := buf.String()
		if bytes.Count([]byte(out), []byte("\n")) == 0 {
			t.Fatalf("expected output after immediate close, got none")
		}
	})

}

func TestSendMessage_BatchedWriter_LargePayload(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendMessage_BatchedWriter_LargePayload", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-batch-large")
		buf := &bytes.Buffer{}
		sp.SetInput(strings.NewReader(""))
		sp.mu.Lock()
		sp.output = buf
		sp.disableBatching = false
		sp.mu.Unlock()

		sp.wg.Add(1)
		go sp.messageWriter()

		// Create a large payload exceeding defaultWriterBufSize
		// Use valid JSON string data instead of raw 'z' characters
		largeString := strings.Repeat("x", defaultWriterBufSize*3)
		largePayload := map[string]string{
			"data": largeString,
		}

		largeJSON, err := json.Marshal(largePayload)
		if err != nil {
			t.Fatalf("Failed to marshal large payload: %v", err)
		}

		msg := &Message{Type: MessageTypeEvent, ID: "large1", Payload: largeJSON}
		if err := sp.sendMessage(msg); err != nil {
			t.Fatalf("sendMessage failed: %v", err)
		}

		if err := sp.Stop(); err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "large1") {
			t.Fatalf("expected large payload message in output, got: %s", out)
		}
	})

}

func TestRun_InvalidJSON_SendsParseError(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRun_InvalidJSON_SendsParseError", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-run")
		in := strings.NewReader("not-a-json\n")
		sp.SetInput(in)
		buf := &bytes.Buffer{}
		sp.SetOutput(buf)

		// run will return after processing input
		if err := sp.Run(); err != nil {
			t.Fatalf("Run returned error: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "parse-error") {
			t.Fatalf("expected parse-error in output, got: %s", out)
		}
	})

}

func TestHandleMessage_NoHandler(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestHandleMessage_NoHandler", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-handle")
		buf := &bytes.Buffer{}
		sp.SetOutput(buf)

		// ensure wg count so defer s.wg.Done() is safe
		sp.wg.Add(1)
		sp.handleMessage(&Message{Type: MessageTypeRequest, ID: "missing"})
		// wait a bit for send; this small sleep is acceptable because it
		// only allows a background send to complete and is unlikely to
		// cause CI instability.
		time.Sleep(10 * time.Millisecond)
		out := buf.String()
		if !strings.Contains(out, "no handler found") {
			t.Fatalf("expected no handler message, got: %s", out)
		}
	})

}

func TestRegisterHandlerAndResponse(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRegisterHandlerAndResponse", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-reg")
		buf := &bytes.Buffer{}
		sp.SetOutput(buf)

		sp.RegisterHandler("echo", func(_ context.Context, msg *Message) (*Message, error) {
			return &Message{Type: MessageTypeResponse, ID: msg.ID, Payload: json.RawMessage(`"ok"`)}, nil
		})

		sp.wg.Add(1)
		sp.handleMessage(&Message{Type: MessageTypeRequest, ID: "echo"})
		// allow time for response
		time.Sleep(10 * time.Millisecond)
		out := buf.String()
		if !strings.Contains(out, "response") || !strings.Contains(out, "echo") {
			t.Fatalf("unexpected response output: %s", out)
		}
	})

}

func TestSetupLoggingCreatesFile(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupLoggingCreatesFile", nil, func(t *testing.T, tx *gorm.DB) {
		proc := "test-logger"
		logger, err := SetupLogging(proc, common.DefaultLogsDir, common.InfoLevel)
		if err != nil {
			t.Fatalf("SetupLogging failed: %v", err)
		}
		if logger == nil {
			t.Fatalf("expected logger")
		}
		// Ensure log file exists
		path := "./logs/" + proc + ".log"
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("log file not created: %s", path)
			} else {
				t.Fatalf("stat error: %v", err)
			}
		}
		_ = logger
		// cleanup
		_ = os.Remove(path)
	})

}

func TestSetupSignalHandlerCapacity(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSetupSignalHandlerCapacity", nil, func(t *testing.T, tx *gorm.DB) {
		sig := SetupSignalHandler()
		if sig == nil {
			t.Fatal("expected non-nil channel")
		}
		if cap(sig) != 1 {
			t.Fatalf("expected channel capacity 1, got %d", cap(sig))
		}
		// unregister to avoid leaking notifications
		signal := syscall.SIGTERM
		// do not actually send signal to the process in test
		_ = signal
	})

}

func TestFlushBatchDirectAndJoinPaths(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFlushBatchDirectAndJoinPaths", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("bench-flush")
		buf := &bytes.Buffer{}
		sp.SetOutput(buf)

		// small batch
		small := make([][]byte, 0, 3)
		small = append(small, []byte("a"), []byte("b"), []byte("c"))
		sp.flushBatch(small)
		if buf.Len() == 0 {
			t.Fatalf("expected output for small batch")
		}
		buf.Reset()

		// large single message path
		large := make([]byte, defaultWriterBufSize*2)
		for i := range large {
			large[i] = 'x'
		}
		sp.flushBatch([][]byte{large})
		if buf.Len() == 0 {
			t.Fatalf("expected output for large batch")
		}
	})

}

func TestSendResponseEventErrorOutput(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSendResponseEventErrorOutput", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test-send-types")
		buf := &bytes.Buffer{}
		sp.SetOutput(buf)

		if err := sp.SendResponse("rid", map[string]string{"v": "1"}); err != nil {
			t.Fatalf("SendResponse failed: %v", err)
		}
		if err := sp.SendEvent("eid", "payload"); err != nil {
			t.Fatalf("SendEvent failed: %v", err)
		}
		if err := sp.SendError("errid", io.ErrClosedPipe); err != nil {
			t.Fatalf("SendError failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "rid") || !strings.Contains(out, "eid") || !strings.Contains(out, "errid") {
			t.Fatalf("unexpected output: %s", out)
		}
	})

}

// Unmarshal tests are covered in subprocess_test.go; no duplicates here.
