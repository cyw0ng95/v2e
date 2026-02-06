package subprocess

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestMessage_StateTransitions covers edge cases in message type and field combinations.
func TestMessage_StateTransitions(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMessage_StateTransitions", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			msg  Message
		}{
			{name: "request-empty-id", msg: Message{Type: MessageTypeRequest, ID: ""}},
			{name: "request-with-payload", msg: Message{Type: MessageTypeRequest, ID: "r1", Payload: json.RawMessage(`{"k":"v"}`)}},
			{name: "response-with-correlation", msg: Message{Type: MessageTypeResponse, ID: "resp1", CorrelationID: "corr1"}},
			{name: "event-no-correlation", msg: Message{Type: MessageTypeEvent, ID: "evt1"}},
			{name: "error-with-message", msg: Message{Type: MessageTypeError, ID: "err1", Error: "failure"}},
			{name: "request-with-source", msg: Message{Type: MessageTypeRequest, ID: "r2", Source: "proc1"}},
			{name: "response-with-target", msg: Message{Type: MessageTypeResponse, ID: "resp2", Target: "proc2"}},
			{name: "request-source-and-target", msg: Message{Type: MessageTypeRequest, ID: "r3", Source: "a", Target: "b"}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := MarshalFast(&tc.msg)
				if err != nil {
					t.Fatalf("MarshalFast failed: %v", err)
				}
				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed: %v", err)
				}
				if decoded.Type != tc.msg.Type || decoded.ID != tc.msg.ID {
					t.Fatalf("round-trip mismatch: want %+v got %+v", tc.msg, decoded)
				}
			})
		}
	})

}

// TestSubprocess_HandlerRegistry covers many handler registration combinations.
func TestSubprocess_HandlerRegistry(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_HandlerRegistry", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("test")
		var calls []string
		var mu sync.Mutex

		for i := 0; i < 60; i++ {
			handlerID := fmt.Sprintf("handler-%02d", i)
			handler := func(ctx context.Context, msg *Message) (*Message, error) {
				mu.Lock()
				calls = append(calls, msg.ID)
				mu.Unlock()
				return &Message{Type: MessageTypeResponse, ID: msg.ID}, nil
			}
			sp.RegisterHandler(handlerID, handler)
		}

		// Verify handlers are registered
		sp.mu.RLock()
		if len(sp.handlers) != 60 {
			t.Fatalf("expected 60 handlers, got %d", len(sp.handlers))
		}
		sp.mu.RUnlock()

		// Invoke each handler via HandleMessage
		for i := 0; i < 60; i++ {
			msgID := fmt.Sprintf("handler-%02d", i)
			msg := &Message{Type: MessageTypeRequest, ID: msgID}
			resp, err := sp.HandleMessage(context.Background(), msg)
			if err != nil {
				t.Fatalf("HandleMessage failed for %s: %v", msgID, err)
			}
			if resp == nil || resp.Type != MessageTypeResponse {
				t.Fatalf("unexpected response for %s: %+v", msgID, resp)
			}
		}

		mu.Lock()
		defer mu.Unlock()
		if len(calls) != 60 {
			t.Fatalf("expected 60 calls, got %d", len(calls))
		}
	})

}

// TestSubprocess_ConcurrentMessageHandling exercises concurrent handler invocations.
func TestSubprocess_ConcurrentMessageHandling(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_ConcurrentMessageHandling", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("concurrent-test")
		var counter sync.WaitGroup
		var processed int32
		var mu sync.Mutex

		handler := func(ctx context.Context, msg *Message) (*Message, error) {
			mu.Lock()
			processed++
			mu.Unlock()
			return &Message{Type: MessageTypeResponse, ID: msg.ID}, nil
		}
		sp.RegisterHandler("concurrent", handler)

		numGoroutines := 50
		counter.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				defer counter.Done()
				msg := &Message{Type: MessageTypeRequest, ID: "concurrent", CorrelationID: fmt.Sprintf("c%d", idx)}
				_, _ = sp.HandleMessage(context.Background(), msg)
			}(i)
		}
		counter.Wait()

		mu.Lock()
		defer mu.Unlock()
		if processed != int32(numGoroutines) {
			t.Fatalf("expected %d processed messages, got %d", numGoroutines, processed)
		}
	})

}

// TestSubprocess_SendMessageBatching covers batching edge cases.
func TestSubprocess_SendMessageBatching(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_SendMessageBatching", nil, func(t *testing.T, tx *gorm.DB) {
		var buf bytes.Buffer
		sp := New("batch-test")
		sp.SetOutput(&buf)

		// Send a burst of messages to exercise batching
		for i := 0; i < 30; i++ {
			msg := &Message{Type: MessageTypeEvent, ID: fmt.Sprintf("evt-%d", i)}
			if msg.Type != MessageTypeEvent {
				t.Fatalf("Expected MessageTypeEvent, got %s", msg.Type)
			}
			if err := sp.SendEvent(msg.ID, nil); err != nil {
				t.Fatalf("SendEvent failed: %v", err)
			}
		}

		// Allow message writer goroutine to flush (small sleep or synchronization)
		sp.Stop()

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		// At least one line per message plus the ready event
		if len(lines) < 30 {
			t.Fatalf("expected at least 30 lines, got %d", len(lines))
		}
	})

}

// TestSubprocess_StateIsolation ensures each subprocess instance has isolated state.
func TestSubprocess_StateIsolation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_StateIsolation", nil, func(t *testing.T, tx *gorm.DB) {
		sp1 := New("sp1")
		sp2 := New("sp2")

		sp1.RegisterHandler("h1", func(ctx context.Context, msg *Message) (*Message, error) {
			return &Message{Type: MessageTypeResponse, ID: "h1-resp"}, nil
		})
		sp2.RegisterHandler("h2", func(ctx context.Context, msg *Message) (*Message, error) {
			return &Message{Type: MessageTypeResponse, ID: "h2-resp"}, nil
		})

		// sp1 should not have h2
		msg := &Message{Type: MessageTypeRequest, ID: "h2"}
		_, err := sp1.HandleMessage(context.Background(), msg)
		if err == nil {
			t.Fatalf("expected sp1 to not have h2 handler")
		}

		// sp2 should not have h1
		msg = &Message{Type: MessageTypeRequest, ID: "h1"}
		_, err = sp2.HandleMessage(context.Background(), msg)
		if err == nil {
			t.Fatalf("expected sp2 to not have h1 handler")
		}
	})

}

// TestSubprocess_ContextCancellation ensures context cancellation stops handlers.
func TestSubprocess_ContextCancellation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_ContextCancellation", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("cancel-test")
		var called bool
		var mu sync.Mutex

		handler := func(ctx context.Context, msg *Message) (*Message, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				mu.Lock()
				called = true
				mu.Unlock()
				return &Message{Type: MessageTypeResponse, ID: msg.ID}, nil
			}
		}
		sp.RegisterHandler("test", handler)

		// Cancel context before calling handler
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		msg := &Message{Type: MessageTypeRequest, ID: "test"}
		_, err := sp.HandleMessage(ctx, msg)
		if err == nil {
			t.Fatalf("expected context cancellation error")
		}

		mu.Lock()
		defer mu.Unlock()
		if called {
			t.Fatalf("handler should not have executed after cancellation")
		}
	})

}

// TestSubprocess_MessageTypeRoutingPriority verifies type-based handler fallback.
func TestSubprocess_MessageTypeRoutingPriority(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_MessageTypeRoutingPriority", nil, func(t *testing.T, tx *gorm.DB) {
		sp := New("routing-test")
		var typeHandlerCalled, idHandlerCalled bool

		sp.RegisterHandler(string(MessageTypeResponse), func(ctx context.Context, msg *Message) (*Message, error) {
			typeHandlerCalled = true
			return nil, nil
		})
		sp.RegisterHandler("specific-id", func(ctx context.Context, msg *Message) (*Message, error) {
			idHandlerCalled = true
			return nil, nil
		})

		// Response messages should prioritize type handler
		msg := &Message{Type: MessageTypeResponse, ID: "unknown-id"}
		_, _ = sp.HandleMessage(context.Background(), msg)
		if !typeHandlerCalled {
			t.Fatalf("expected type handler to be called for response message")
		}

		// Request messages should prioritize ID handler
		typeHandlerCalled = false
		idHandlerCalled = false
		msg = &Message{Type: MessageTypeRequest, ID: "specific-id"}
		_, _ = sp.HandleMessage(context.Background(), msg)
		if !idHandlerCalled {
			t.Fatalf("expected ID handler to be called for request message")
		}
	})

}

// TestSubprocess_DisableBatchingMode verifies synchronous message sending.
func TestSubprocess_DisableBatchingMode(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSubprocess_DisableBatchingMode", nil, func(t *testing.T, tx *gorm.DB) {
		var buf bytes.Buffer
		sp := New("no-batch-test")
		sp.SetOutput(&buf)
		sp.disableBatching = true

		msg := &Message{Type: MessageTypeEvent, ID: "sync-evt"}
		if err := sp.SendMessage(msg); err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		// In non-batching mode, message should be written immediately
		output := buf.String()
		if !strings.Contains(output, "sync-evt") {
			t.Fatalf("expected immediate output, got: %s", output)
		}
	})

}

// TestMessage_MarshalUnmarshalEdgeCases covers JSON edge cases.
func TestMessage_MarshalUnmarshalEdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMessage_MarshalUnmarshalEdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			msg  Message
		}{
			{name: "nil-payload", msg: Message{Type: MessageTypeRequest, ID: "r1", Payload: nil}},
			{name: "empty-json-payload", msg: Message{Type: MessageTypeRequest, ID: "r2", Payload: json.RawMessage(`{}`)}},
			{name: "nested-json-payload", msg: Message{Type: MessageTypeRequest, ID: "r3", Payload: json.RawMessage(`{"nested":{"key":"val"}}`)}},
			{name: "unicode-error", msg: Message{Type: MessageTypeError, ID: "e1", Error: "错误信息"}},
			{name: "long-id", msg: Message{Type: MessageTypeRequest, ID: strings.Repeat("x", 256)}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := MarshalFast(&tc.msg)
				if err != nil {
					t.Fatalf("MarshalFast failed: %v", err)
				}
				var decoded Message
				if err := UnmarshalFast(data, &decoded); err != nil {
					t.Fatalf("UnmarshalFast failed: %v", err)
				}
			})
		}
	})

}
