package rpc

import (
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestClient_Signal(t *testing.T) {
	t.Run("Signal_SendMessageOnce", func(t *testing.T) {
		entry := &RequestEntry{
			resp: make(chan *subprocess.Message, 1),
		}

		msg := &subprocess.Message{
			CorrelationID: "test-id",
		}

		entry.Signal(msg)

		select {
		case receivedMsg := <-entry.resp:
			if receivedMsg != msg {
				t.Error("Expected to receive same message")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for message")
		}
	})

	t.Run("Signal_SendMessageOnlyOnce", func(t *testing.T) {
		entry := &RequestEntry{
			resp: make(chan *subprocess.Message, 1),
		}

		msg1 := &subprocess.Message{CorrelationID: "test-id-1"}
		msg2 := &subprocess.Message{CorrelationID: "test-id-2"}

		go func() {
			entry.Signal(msg1)
			time.Sleep(10 * time.Millisecond)
			entry.Signal(msg2)
		}()

		time.Sleep(50 * time.Millisecond)

		select {
		case <-entry.resp:
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for message")
		}
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("Close_CloseChannel", func(t *testing.T) {
		entry := &RequestEntry{
			resp: make(chan *subprocess.Message, 1),
		}

		entry.Close()

		select {
		case _, ok := <-entry.resp:
			if ok {
				t.Error("Expected channel to be closed")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for close")
		}
	})

	t.Run("Close_CloseOnlyOnce", func(t *testing.T) {
		entry := &RequestEntry{
			resp: make(chan *subprocess.Message, 1),
		}

		go func() {
			entry.Close()
			time.Sleep(10 * time.Millisecond)
			entry.Close()
		}()

		time.Sleep(50 * time.Millisecond)

		select {
		case _, ok := <-entry.resp:
			if ok {
				t.Error("Expected channel to be closed")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for close")
		}
	})
}

func TestClient_GetDefaultTimeout(t *testing.T) {
	t.Run("GetDefaultTimeout_ReturnsTimeout", func(t *testing.T) {
		timeout := GetDefaultTimeout()
		if timeout <= 0 {
			t.Errorf("Expected positive timeout, got: %v", timeout)
		}
	})
}
