package subprocess

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// messageWriter is a background goroutine that batches and writes messages
// This reduces syscalls and mutex contention for better performance
func (s *Subprocess) messageWriter() {
	defer s.wg.Done()

	// Optimized batch buffer size (Principle 12)
	// Larger batch reduces syscalls but increases latency
	batch := make([][]byte, 0, 20)
	ticker := time.NewTicker(5 * time.Millisecond) // Faster flush for lower latency (Principle 12)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// Flush any remaining messages before exiting
			if len(batch) > 0 {
				s.flushBatch(batch)
			}
			return
		case data, ok := <-s.outChan:
			if !ok {
				// Channel closed, flush and exit
				if len(batch) > 0 {
					s.flushBatch(batch)
				}
				return
			}
			batch = append(batch, data)

			// Adaptive batching: flush at 20 messages (Principle 12)
			if len(batch) >= 20 {
				s.flushBatch(batch)
				batch = batch[:0] // Reset batch
			}
		case <-ticker.C:
			// Periodic flush to avoid holding messages too long
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch writes all batched messages in a single operation
// Principle 13: Use bufio.Writer for efficient batch writing
// Principle 14: Pool bufio.Writers to reduce allocations
func (s *Subprocess) flushBatch(batch [][]byte) {
	if len(batch) == 0 {
		return
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	// Get writer from pool and reset it
	writer := writerPool.Get().(*bufio.Writer)
	writer.Reset(s.output)
	defer writerPool.Put(writer)

	for _, data := range batch {
		// Write data directly without fmt.Fprintf overhead
		if _, err := writer.Write(data); err != nil {
			// Log error but continue processing
			fmt.Fprintf(os.Stderr, "Failed to write message: %v\n", err)
			continue
		}
		// Write newline
		if err := writer.WriteByte('\n'); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write newline: %v\n", err)
		}
	}
	// Flush to ensure all data is written
	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to flush batch: %v\n", err)
	}
}
