package taskflow

import (
	"sync"
)

// PoolSize represents the size tier for objects
type PoolSize int

const (
	PoolTiny   PoolSize = iota // 0-256 bytes
	PoolSmall                  // 257-1024 bytes
	PoolMedium                 // 1025-4096 bytes
	PoolLarge                  // 4097-16384 bytes
	PoolHuge                   // 16385+ bytes
)

// TieredPoolConfig configures the tiered pool
type TieredPoolConfig struct {
	// MaxPooledSize is the maximum size that will be pooled (objects larger are allocated directly)
	MaxPooledSize int
	// PreAllocCounts is the number of objects to preallocate for each tier
	PreAllocCounts map[PoolSize]int
	// MaxPoolSize is the maximum number of objects to keep per tier
	MaxPoolSize int
}

// DefaultTieredPoolConfig returns default configuration
func DefaultTieredPoolConfig() TieredPoolConfig {
	return TieredPoolConfig{
		MaxPooledSize: 64 * 1024, // 64KB
		PreAllocCounts: map[PoolSize]int{
			PoolTiny:   100,
			PoolSmall:  50,
			PoolMedium: 25,
			PoolLarge:  10,
			PoolHuge:   5,
		},
		MaxPoolSize: 100,
	}
}

// TieredPool implements hierarchical object pooling with size-based pools
type TieredPool struct {
	config TieredPoolConfig
	pools  [5]*sync.Pool
	stats  *PoolStats
	mu     sync.RWMutex
}

// PoolStats tracks pool utilization metrics
type PoolStats struct {
	allocations map[PoolSize]int64
	hits        map[PoolSize]int64
	misses      map[PoolSize]int64
}

// NewTieredPool creates a new tiered object pool
func NewTieredPool(config TieredPoolConfig) *TieredPool {
	tp := &TieredPool{
		config: config,
		pools:  [5]*sync.Pool{},
		stats: &PoolStats{
			allocations: make(map[PoolSize]int64),
			hits:        make(map[PoolSize]int64),
			misses:      make(map[PoolSize]int64),
		},
	}

	// Initialize pools for each tier
	tiers := []PoolSize{PoolTiny, PoolSmall, PoolMedium, PoolLarge, PoolHuge}
	for _, tier := range tiers {
		tp.pools[tier] = &sync.Pool{
			New: func() interface{} {
				return tp.newObjectForTier(tier)
			},
		}
	}

	return tp
}

// NewTieredPoolWithDefaults creates a pool with default configuration
func NewTieredPoolWithDefaults() *TieredPool {
	return NewTieredPool(DefaultTieredPoolConfig())
}

// Get obtains a byte slice of the requested size
func (tp *TieredPool) Get(size int) []byte {
	tier := tp.sizeToTier(size)

	// If size exceeds max pooled size, allocate directly
	if size > tp.config.MaxPooledSize {
		tp.recordMiss(tier)
		return make([]byte, size)
	}

	// Try to get from pool
	if obj := tp.pools[tier].Get(); obj != nil {
		buf := obj.([]byte)
		if cap(buf) >= size {
			tp.recordHit(tier)
			return buf[:size]
		}
		// Buffer too small, put back and allocate new
		tp.pools[tier].Put(buf)
		tp.recordMiss(tier)
	}

	tp.recordMiss(tier)
	return make([]byte, size)
}

// Put returns a byte slice to the pool
func (tp *TieredPool) Put(buf []byte) {
	if buf == nil {
		return
	}

	size := cap(buf)
	tier := tp.sizeToTier(size)

	// Don't pool if too large
	if size > tp.config.MaxPooledSize {
		return
	}

	// Check pool size limit (approximate via sampling)
	tp.mu.RLock()
	count := tp.stats.allocations[tier] - tp.stats.hits[tier]
	tp.mu.RUnlock()

	if count > int64(tp.config.MaxPoolSize) {
		return
	}

	tp.pools[tier].Put(buf)
}

// GetStats returns pool utilization statistics
func (tp *TieredPool) GetStats() map[string]interface{} {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	totalAllocs := int64(0)
	totalHits := int64(0)
	totalMisses := int64(0)

	for _, tier := range []PoolSize{PoolTiny, PoolSmall, PoolMedium, PoolLarge, PoolHuge} {
		totalAllocs += tp.stats.allocations[tier]
		totalHits += tp.stats.hits[tier]
		totalMisses += tp.stats.misses[tier]
	}

	hitRate := 0.0
	if totalAllocs > 0 {
		hitRate = float64(totalHits) / float64(totalAllocs) * 100
	}

	return map[string]interface{}{
		"total_allocations": totalAllocs,
		"total_hits":        totalHits,
		"total_misses":      totalMisses,
		"hit_rate_percent":  hitRate,
		"tier_breakdown": map[PoolSize]map[string]int64{
			PoolTiny: {
				"allocations": tp.stats.allocations[PoolTiny],
				"hits":        tp.stats.hits[PoolTiny],
				"misses":      tp.stats.misses[PoolTiny],
			},
			PoolSmall: {
				"allocations": tp.stats.allocations[PoolSmall],
				"hits":        tp.stats.hits[PoolSmall],
				"misses":      tp.stats.misses[PoolSmall],
			},
			PoolMedium: {
				"allocations": tp.stats.allocations[PoolMedium],
				"hits":        tp.stats.hits[PoolMedium],
				"misses":      tp.stats.misses[PoolMedium],
			},
			PoolLarge: {
				"allocations": tp.stats.allocations[PoolLarge],
				"hits":        tp.stats.hits[PoolLarge],
				"misses":      tp.stats.misses[PoolLarge],
			},
			PoolHuge: {
				"allocations": tp.stats.allocations[PoolHuge],
				"hits":        tp.stats.hits[PoolHuge],
				"misses":      tp.stats.misses[PoolHuge],
			},
		},
	}
}

// sizeToTier maps size to appropriate tier
func (tp *TieredPool) sizeToTier(size int) PoolSize {
	switch {
	case size <= 256:
		return PoolTiny
	case size <= 1024:
		return PoolSmall
	case size <= 4096:
		return PoolMedium
	case size <= 16384:
		return PoolLarge
	default:
		return PoolHuge
	}
}

// newObjectForTier creates a new buffer for the given tier
func (tp *TieredPool) newObjectForTier(tier PoolSize) interface{} {
	var size int
	switch tier {
	case PoolTiny:
		size = 256
	case PoolSmall:
		size = 1024
	case PoolMedium:
		size = 4096
	case PoolLarge:
		size = 16384
	case PoolHuge:
		size = tp.config.MaxPooledSize
	default:
		size = 256
	}

	tp.recordAllocation(tier)
	return make([]byte, 0, size)
}

// recordAllocation records an allocation event
func (tp *TieredPool) recordAllocation(tier PoolSize) {
	tp.mu.Lock()
	tp.stats.allocations[tier]++
	tp.mu.Unlock()
}

// recordHit records a pool hit
func (tp *TieredPool) recordHit(tier PoolSize) {
	tp.mu.Lock()
	tp.stats.hits[tier]++
	tp.mu.Unlock()
}

// recordMiss records a pool miss
func (tp *TieredPool) recordMiss(tier PoolSize) {
	tp.mu.Lock()
	tp.stats.misses[tier]++
	tp.mu.Unlock()
}
