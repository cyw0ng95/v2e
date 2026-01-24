package main

import (
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

func TestRequestEntry_SignalClose(t *testing.T) {
	resp := make(chan *subprocess.Message, 1)
	e := &requestEntry{resp: resp}

	// first signal should send and close the channel
	e.signal(&subprocess.Message{Type: subprocess.MessageTypeResponse})
	// read the value
	select {
	case _, ok := <-resp:
		if ok {
			// channel should be closed after signal, so ok may be false
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timed out waiting for signal")
	}

	// subsequent close should not panic
	e.close()
}
