package subprocess

import (
	"io"
	"testing"
	"time"
)

func BenchmarkSendMessage_Disabled(b *testing.B) {
	sp := New("bench-disabled")
	// Directly set output to discard and disable batching
	sp.output = io.Discard
	sp.disableBatching = true
	defer sp.Stop()

	msg := &Message{Type: MessageTypeEvent, ID: "bench", Payload: nil}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sp.sendMessage(msg)
	}
}

func BenchmarkSendMessage_Batched(b *testing.B) {
	sp := New("bench-batched")
	// Use discard output but do not mark batching disabled
	sp.output = io.Discard
	sp.disableBatching = false

	// Start writer
	sp.wg.Add(1)
	go sp.messageWriter()
	defer func() {
		// Stop will close outChan and wait for writer
		_ = sp.Stop()
	}()

	msg := &Message{Type: MessageTypeEvent, ID: "bench", Payload: nil}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sp.sendMessage(msg)
	}
	// Give writer a moment to flush remaining messages
	time.Sleep(defaultFlushInterval * 2)
}

func BenchmarkFlushBatch_VariousSizes(b *testing.B) {
	sp := New("bench-flush")
	sp.output = io.Discard
	batches := [][]byte{}

	// small (100 bytes), medium (1KB), large (8KB)
	sizes := []int{100, 1024, 8 * 1024}
	for _, s := range sizes {
		buf := make([]byte, s)
		for i := range buf {
			buf[i] = 'a' + byte(i%26)
		}
		batches = append(batches, buf)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// create a batch of 20 messages of medium size to simulate realistic load
		batch := make([][]byte, 0, 20)
		for j := 0; j < 20; j++ {
			batch = append(batch, batches[1])
		}
		sp.flushBatch(batch)
	}
}
