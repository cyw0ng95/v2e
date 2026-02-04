package subprocess

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bytedance/sonic"
)

// BenchmarkNew benchmarks creating a new subprocess instance
func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New("test-process")
	}
}

// BenchmarkRegisterHandler benchmarks registering a message handler
func BenchmarkRegisterHandler(b *testing.B) {
	sp := New("test-process")
	handler := func(ctx context.Context, msg *Message) (*Message, error) {
		return &Message{
			Type: MessageTypeResponse,
			ID:   msg.ID,
		}, nil
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handlerName := fmt.Sprintf("TestHandler_%d", i)
		sp.RegisterHandler(handlerName, handler)
	}
}

// BenchmarkSendMessage benchmarks sending a message
func BenchmarkSendMessage(b *testing.B) {
	var buf bytes.Buffer
	sp := New("test-process")
	sp.SetOutput(&buf)

	msg := &Message{
		Type:    MessageTypeRequest,
		ID:      "test-msg",
		Payload: json.RawMessage(`{"test":"value"}`),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = sp.SendMessage(msg)
	}
}

// BenchmarkSendResponse benchmarks sending a response message
func BenchmarkSendResponse(b *testing.B) {
	var buf bytes.Buffer
	sp := New("test-process")
	sp.SetOutput(&buf)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = sp.SendResponse("test-id", map[string]string{"result": "ok"})
	}
}

// BenchmarkMessageMarshal benchmarks marshaling a message
func BenchmarkMessageMarshal(b *testing.B) {
	msg := &Message{
		Type:    MessageTypeRequest,
		ID:      "test-id",
		Payload: json.RawMessage(`{"test":"value"}`),
		Source:  "process-a",
		Target:  "process-b",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sonic.Marshal(msg)
	}
}

// BenchmarkMessageUnmarshal benchmarks unmarshaling a message
func BenchmarkMessageUnmarshal(b *testing.B) {
	data := []byte(`{"type":"request","id":"test-id","payload":{"test":"value"},"source":"process-a","target":"process-b"}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg Message
		_ = sonic.Unmarshal(data, &msg)
	}
}

// BenchmarkMessageRoundTrip benchmarks a full message send/receive cycle
func BenchmarkMessageRoundTrip(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create message
		msg := &Message{
			Type:    MessageTypeRequest,
			ID:      "test-id",
			Payload: json.RawMessage(`{"test":"value"}`),
		}

		// Marshal
		data, _ := sonic.Marshal(msg)

		// Unmarshal
		var decoded Message
		_ = sonic.Unmarshal(data, &decoded)
	}
}

// BenchmarkHandlerRegistration benchmarks handler lookup performance
func BenchmarkHandlerRegistration(b *testing.B) {
	sp := New("test-process")

	// Register multiple handlers
	for i := 0; i < 100; i++ {
		handlerName := fmt.Sprintf("Handler_%d", i)
		sp.RegisterHandler(handlerName, func(ctx context.Context, msg *Message) (*Message, error) {
			return &Message{
				Type: MessageTypeResponse,
				ID:   msg.ID,
			}, nil
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handlerName := fmt.Sprintf("NewHandler_%d", 100+i)
		sp.RegisterHandler(handlerName, func(ctx context.Context, msg *Message) (*Message, error) {
			return &Message{
				Type: MessageTypeResponse,
				ID:   msg.ID,
			}, nil
		})
	}
}

// BenchmarkSendEvent benchmarks sending event messages
func BenchmarkSendEvent(b *testing.B) {
	var buf bytes.Buffer
	sp := New("test-process")
	sp.SetOutput(&buf)

	payload := map[string]interface{}{
		"status": "running",
		"count":  42,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = sp.SendEvent("test-event", payload)
	}
}

// BenchmarkSendError benchmarks sending error messages
func BenchmarkSendError(b *testing.B) {
	var buf bytes.Buffer
	sp := New("test-process")
	sp.SetOutput(&buf)

	testErr := context.Canceled

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = sp.SendError("test-error", testErr)
	}
}

// BenchmarkLargeMessageSend benchmarks sending large messages
func BenchmarkLargeMessageSend(b *testing.B) {
	var buf bytes.Buffer
	sp := New("test-process")
	sp.SetOutput(&buf)

	// Create a large payload (100KB)
	largeData := make([]byte, 100*1024)
	for i := range largeData {
		largeData[i] = byte('A' + i%26)
	}
	largePayload := map[string]string{
		"data": string(largeData),
	}

	msg := &Message{
		Type:    MessageTypeRequest,
		ID:      "large-msg",
		Payload: json.RawMessage(`{}`),
	}
	msg.Payload, _ = sonic.Marshal(largePayload)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = sp.SendMessage(msg)
	}
}

// BenchmarkConcurrentSend benchmarks concurrent message sending
func BenchmarkConcurrentSend(b *testing.B) {
	var buf bytes.Buffer
	sp := New("test-process")
	sp.SetOutput(&buf)

	msg := &Message{
		Type: MessageTypeRequest,
		ID:   "concurrent-msg",
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = sp.SendMessage(msg)
		}
	})
}
