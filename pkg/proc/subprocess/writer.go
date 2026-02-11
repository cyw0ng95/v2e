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
	batch := make([][]byte, 0, defaultBatchSize)
	ticker := time.NewTicker(defaultFlushInterval) // Faster flush for lower latency (Principle 12)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// Drain any currently buffered messages from outChan without blocking
			for {
				select {
				case data, ok := <-s.outChan:
					if !ok {
						if len(batch) > 0 {
							s.flushBatch(batch)
						}
						return
					}
					batch = append(batch, data)
					// flush if we reached batch size while draining
					if len(batch) >= defaultBatchSize {
						s.flushBatch(batch)
						batch = batch[:0]
					}
				default:
					if len(batch) > 0 {
						s.flushBatch(batch)
					}
					return
				}
			}
		case data, ok := <-s.outChan:
			if !ok {
				// Channel closed, flush and exit
				if len(batch) > 0 {
					s.flushBatch(batch)
				}
				return
			}
			batch = append(batch, data)

			// Adaptive batching: flush at defaultBatchSize messages (Principle 12)
			if len(batch) >= defaultBatchSize {
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

// flushBatch writes all batched messages in a single operation.
// Uses bufio.Writer for efficient batch writing (Principle 13).
// Pools bufio.Writers to reduce allocations (Principle 14).
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

	// Compute total size to decide strategy
	totalSize := 0
	for _, b := range batch {
		totalSize += len(b) + 1 // include newline
	}

	// If total size exceeds 4x writer buffer, write directly to avoid large allocations.
	// This handles the edge case of very large messages or batches.
	if totalSize >= 4*defaultWriterBufSize {
		for _, data := range batch {
			if _, err := writer.Write(data); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write message: %v\n", err)
				continue
			}
			if err := writer.WriteByte('\n'); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write newline: %v\n", err)
			}
		}
		if err := writer.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to flush batch: %v\n", err)
		}
		return
	}

	// Normal case: join batch into single buffer for efficient writing.
	// Use a pooled buffer to avoid allocating on each flush.
	bufPtr := batchJoinPool.Get().(*[]byte)
	buf := *bufPtr

	// Pre-allocate with known size to avoid reslices
	if cap(buf) < totalSize {
		buf = make([]byte, 0, totalSize)
	}
	buf = buf[:0]

	// Join all messages with newlines
	for i, b := range batch {
		buf = append(buf, b...)
		if i < len(batch)-1 {
			buf = append(buf, '\n')
		}
	}
	// Ensure trailing newline
	buf = append(buf, '\n')

	// Write the joined buffer in a single call
	if _, err := writer.Write(buf); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write batch: %v\n", err)
		batchJoinPool.Put(bufPtr)
		return
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to flush batch: %v\n", err)
	}

	// Reset and return buffer to pool for reuse
	*bufPtr = buf[:0]
	batchJoinPool.Put(bufPtr)
}
