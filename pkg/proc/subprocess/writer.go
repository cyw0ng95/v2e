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

	// Compute total and max sizes to decide strategy
	totalSize := 0
	maxMsgSize := 0
	for _, b := range batch {
		sz := len(b) + 1 // include newline
		totalSize += sz
		if sz > maxMsgSize {
			maxMsgSize = sz
		}
	}

	// If any single message is very large, or totalSize exceeds writer buffer,
	// write messages directly to avoid allocating a joined buffer (which copies data).
	if maxMsgSize >= defaultWriterBufSize || totalSize >= 4*defaultWriterBufSize {
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

	// For small batches, write directly to avoid allocation from joining
	if len(batch) <= 4 {
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

	// Attempt to join batch into a single buffer and write once.
	// Use a pooled buffer to avoid allocating on each flush.
	bufPtr := batchJoinPool.Get().(*[]byte)
	buf := *bufPtr
	buf = buf[:0]

	// Pre-allocate an estimate to reduce reslices
	estSize := totalSize
	if cap(buf) < estSize {
		buf = make([]byte, 0, estSize)
	}

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
		// Return the buffer to pool and continue
		batchJoinPool.Put(bufPtr)
		return
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to flush batch: %v\n", err)
	}

	// Reset and return buffer to pool (reuse underlying array)
	*bufPtr = buf[:0]
	batchJoinPool.Put(bufPtr)
}
