package proc

import (
	"sync"
)

// ResponseBufferPool manages a pool of response buffers
type ResponseBufferPool struct {
	mu         sync.RWMutex
	smallPool  []bufferEntry
	mediumPool []bufferEntry
	largePool  []bufferEntry
	hugePool   []bufferEntry
	smallNext  int
	mediumNext int
	largeNext  int
	hugeNext   int
	hitCount   map[BufferClass]int
	missCount  map[BufferClass]int
}

// bufferEntry represents a pooled buffer with availability flag
type bufferEntry struct {
	buffer    []byte
	available bool // true if buffer is available in pool
}

// BufferClass defines size classes for buffers
type BufferClass int

const (
	BufferClassSmall  BufferClass = iota // < 4KB
	BufferClassMedium                    // < 32KB
	BufferClassLarge                     // < 256KB
	BufferClassHuge                      // >= 256KB
)

const (
	maxPoolSize = 1024
)

// BufferStats tracks buffer pool statistics
type BufferStats struct {
	Hits   map[BufferClass]int
	Misses map[BufferClass]int
}

// NewResponseBufferPool creates a new response buffer pool
func NewResponseBufferPool() *ResponseBufferPool {
	poolSize := 16
	return &ResponseBufferPool{
		smallPool:  make([]bufferEntry, poolSize),
		mediumPool: make([]bufferEntry, poolSize),
		largePool:  make([]bufferEntry, poolSize),
		hugePool:   make([]bufferEntry, poolSize),
		hitCount:   make(map[BufferClass]int),
		missCount:  make(map[BufferClass]int),
	}
}

// Get retrieves a buffer from the pool based on desired size
func (rbp *ResponseBufferPool) Get(size int) *[]byte {
	rbp.mu.Lock()
	defer rbp.mu.Unlock()

	class := rbp.classifySize(size)
	pool, nextIdx := rbp.getPoolAndIndex(class)

	// Search for available buffer
	for i := 0; i < len(pool); i++ {
		idx := (nextIdx + i) % len(pool)
		if pool[idx].available {
			// Hit: found available buffer
			pool[idx].available = false
			rbp.updateNextIndex(class, idx+1)
			rbp.hitCount[class]++
			return &pool[idx].buffer
		}
	}

	// Miss: no available buffer, allocate new
	rbp.missCount[class]++
	newBuf := rbp.newBuffer(size)
	return &newBuf
}

// Put returns a buffer to the pool
func (rbp *ResponseBufferPool) Put(buf *[]byte) {
	if buf == nil {
		return
	}

	capacity := cap(*buf)
	class := rbp.classifySize(capacity)

	rbp.mu.Lock()
	defer rbp.mu.Unlock()

	pool, _ := rbp.getPoolAndIndex(class)

	// Reset buffer and mark as available
	*buf = (*buf)[:0]
	entry := bufferEntry{buffer: *buf, available: true}

	// Try to find an empty slot to reuse
	for i := 0; i < len(pool); i++ {
		if !pool[i].available {
			pool[i] = entry
			return
		}
	}

	// No empty slot, check if pool is at max capacity
	if len(pool) >= maxPoolSize {
		return
	}
	pool = append(pool, entry)
	rbp.setPool(class, pool)
}

// getPoolAndIndex returns the pool slice and next index for a class
func (rbp *ResponseBufferPool) getPoolAndIndex(class BufferClass) ([]bufferEntry, int) {
	switch class {
	case BufferClassSmall:
		return rbp.smallPool, rbp.smallNext
	case BufferClassMedium:
		return rbp.mediumPool, rbp.mediumNext
	case BufferClassLarge:
		return rbp.largePool, rbp.largeNext
	case BufferClassHuge:
		return rbp.hugePool, rbp.hugeNext
	default:
		return rbp.smallPool, rbp.smallNext
	}
}

// setPool updates the pool slice for a class
func (rbp *ResponseBufferPool) setPool(class BufferClass, pool []bufferEntry) {
	switch class {
	case BufferClassSmall:
		rbp.smallPool = pool
		return
	case BufferClassMedium:
		rbp.mediumPool = pool
		return
	case BufferClassLarge:
		rbp.largePool = pool
		return
	case BufferClassHuge:
		rbp.hugePool = pool
		return
	}
}

// updateNextIndex updates the next index for a class
func (rbp *ResponseBufferPool) updateNextIndex(class BufferClass, idx int) {
	switch class {
	case BufferClassSmall:
		rbp.smallNext = idx
	case BufferClassMedium:
		rbp.mediumNext = idx
	case BufferClassLarge:
		rbp.largeNext = idx
	case BufferClassHuge:
		rbp.hugeNext = idx
	}
}

// newBuffer creates a new buffer with appropriate capacity
func (rbp *ResponseBufferPool) newBuffer(size int) []byte {
	var capacity int
	if size < 4096 {
		capacity = 4096
	} else if size < 32768 {
		capacity = 32768
	} else if size < 262144 {
		capacity = 262144
	} else {
		capacity = 1048576
	}
	return make([]byte, 0, capacity)
}

// classifySize classifies buffer size into a class
func (rbp *ResponseBufferPool) classifySize(size int) BufferClass {
	if size <= 4096 {
		return BufferClassSmall
	} else if size <= 32768 {
		return BufferClassMedium
	} else if size <= 262144 {
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
		Misses: make(map[BufferClass]int),
	}

	for k, v := range rbp.hitCount {
		stats.Hits[k] = v
	}
	for k, v := range rbp.missCount {
		stats.Misses[k] = v
	}

	return stats
}

// ResetStats clears buffer pool statistics
func (rbp *ResponseBufferPool) ResetStats() {
	rbp.mu.Lock()
	defer rbp.mu.Unlock()

	rbp.hitCount = make(map[BufferClass]int)
	rbp.missCount = make(map[BufferClass]int)
}

// GetPoolForMethod returns a pool for a specific hot RPC method
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
		return &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, size)
				return &b
			},
		}
	}

	// Default to medium pool
	return &sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, 32768)
			return &b
		},
	}
}
