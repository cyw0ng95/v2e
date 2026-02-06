package proc

import (
	"sync"
)

// ResponseBufferPool manages a pool of response buffers
type ResponseBufferPool struct {
	mu         sync.RWMutex
	smallPool  *sync.Pool
	mediumPool *sync.Pool
	largePool  *sync.Pool
	hugePool   *sync.Pool
	hitCount   map[BufferClass]int
	missCount  int
}

// BufferClass defines size classes for buffers
type BufferClass int

const (
	BufferClassSmall  BufferClass = iota // < 4KB
	BufferClassMedium                    // < 32KB
	BufferClassLarge                     // < 256KB
	BufferClassHuge                      // >= 256KB
)

// BufferStats tracks buffer pool statistics
type BufferStats struct {
	Hits   map[BufferClass]int
	Misses int
}

// NewResponseBufferPool creates a new response buffer pool
func NewResponseBufferPool() *ResponseBufferPool {
	return &ResponseBufferPool{
		smallPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 4096) // 4KB
				return &b
			},
		},
		mediumPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 32768) // 32KB
				return &b
			},
		},
		largePool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 262144) // 256KB
				return &b
			},
		},
		hugePool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 1048576) // 1MB
				return &b
			},
		},
		hitCount: make(map[BufferClass]int),
	}
}

// Get retrieves a buffer from the pool based on desired size
func (rbp *ResponseBufferPool) Get(size int) *[]byte {
	pool := rbp.selectPool(size)
	bufPtr := pool.Get().(*[]byte)

	rbp.mu.Lock()
	class := rbp.classifySize(size)
	rbp.hitCount[class]++
	rbp.mu.Unlock()

	return bufPtr
}

// Put returns a buffer to the pool
func (rbp *ResponseBufferPool) Put(buf *[]byte) {
	if buf == nil {
		return
	}

	size := cap(*buf)
	pool := rbp.selectPool(size)

	// Reset buffer length
	*buf = (*buf)[:0]

	pool.Put(buf)
}

// selectPool selects appropriate pool based on size
func (rbp *ResponseBufferPool) selectPool(size int) *sync.Pool {
	if size < 4096 {
		return rbp.smallPool
	} else if size < 32768 {
		return rbp.mediumPool
	} else if size < 262144 {
		return rbp.largePool
	} else {
		return rbp.hugePool
	}
}

// classifySize classifies buffer size into a class
func (rbp *ResponseBufferPool) classifySize(size int) BufferClass {
	if size < 4096 {
		return BufferClassSmall
	} else if size < 32768 {
		return BufferClassMedium
	} else if size < 262144 {
		return BufferClassLarge
	} else {
		return BufferClassHuge
	}
}

// GetStats returns buffer pool statistics
func (rbp *ResponseBufferPool) GetStats() BufferStats {
	rbp.mu.RLock()
	defer rbp.mu.RUnlock()

	stats := BufferStats{
		Hits:   make(map[BufferClass]int),
		Misses: rbp.missCount,
	}

	for k, v := range rbp.hitCount {
		stats.Hits[k] = v
	}

	return stats
}

// ResetStats clears buffer pool statistics
func (rbp *ResponseBufferPool) ResetStats() {
	rbp.mu.Lock()
	defer rbp.mu.Unlock()

	rbp.hitCount = make(map[BufferClass]int)
	rbp.missCount = 0
}

// GetPoolForMethod returns the pool for a specific hot RPC method
func (rbp *ResponseBufferPool) GetPoolForMethod(methodName string) *sync.Pool {
	// Map known hot methods to appropriate pool sizes
	// This is a simple heuristic - could be learned over time
	hotMethods := map[string]int{
		"RPCGetCVE":     32768,   // Medium
		"RPCListCVEs":   1048576, // Huge for bulk
		"RPCSearchCVEs": 1048576, // Huge
		"RPCGetCWE":     4096,    // Small
		"RPCGetCAPEC":   32768,   // Medium
	}

	if size, ok := hotMethods[methodName]; ok {
		if size < 4096 {
			return rbp.smallPool
		} else if size < 32768 {
			return rbp.mediumPool
		} else if size < 262144 {
			return rbp.largePool
		} else {
			return rbp.hugePool
		}
	}

	// Default to medium pool
	return rbp.mediumPool
}
