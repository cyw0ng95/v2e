package subprocess

import (
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

// sonicFast is a shared instance of sonic configured for fastest parsing/marshalling.
var sonicFast = sonic.ConfigFastest

// batchJoinPool provides temporary buffers for joining batched messages
// to avoid allocating large temporary slices on each flush.
var batchJoinPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 16*1024) // 16KB initial capacity
		return &b
	},
}

// Tunable batching parameters (kept package-private)
var (
	defaultBatchSize      = 20
	defaultFlushInterval  = 5 * time.Millisecond
	defaultOutChanBufSize = 256
	defaultWriterBufSize  = 8 * 1024 // 8KB bufio.Writer size

	// zeroCopyThreshold is the minimum payload size (bytes) to attempt
	// a zero-copy direct-write path (only used when batching is disabled)
	zeroCopyThreshold = 4 * 1024 // 4KB
)
