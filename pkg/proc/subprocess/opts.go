package subprocess

import (
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/jsonutil"
)

// sonicFast is a shared instance of sonic configured for fastest parsing/marshalling.
// MarshalFast marshals a value using the shared fastest configuration.
func MarshalFast(v interface{}) ([]byte, error) {
	return jsonutil.Marshal(v)
}

// UnmarshalFast unmarshals data using the shared fastest configuration.
func UnmarshalFast(data []byte, v interface{}) error {
	return jsonutil.Unmarshal(data, v)
}

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
