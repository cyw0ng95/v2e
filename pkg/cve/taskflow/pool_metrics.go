package taskflow

import (
	"sync"
	"time"
)

// PoolMetrics tracks comprehensive pool statistics
type PoolMetrics struct {
	mu sync.RWMutex

	// Allocation stats
	totalAllocations  map[PoolSize]int64
	activeAllocations map[PoolSize]int64

	// Performance stats
	hitCount  map[PoolSize]int64
	missCount map[PoolSize]int64

	// Size distribution
	sizeDistribution map[int]int64

	// Timing stats
	allocationTime    map[PoolSize]time.Duration
	minAllocationTime map[PoolSize]time.Duration
	maxAllocationTime map[PoolSize]time.Duration

	// Timestamps
	startTime time.Time
	lastReset time.Time
}

// PoolUtilizationStats provides utilization analysis
type PoolUtilizationStats struct {
	TotalAllocations  int64                  `json:"total_allocations"`
	ActiveAllocations int64                  `json:"active_allocations"`
	HitRate           float64                `json:"hit_rate_percent"`
	MissRate          float64                `json:"miss_rate_percent"`
	AvgAllocationTime time.Duration          `json:"avg_allocation_time"`
	TierBreakdown     map[PoolSize]TierStats `json:"tier_breakdown"`
	SizeDistribution  map[int]int64          `json:"size_distribution"`
	Uptime            time.Duration          `json:"uptime"`
}

// TierStats provides per-tier statistics
type TierStats struct {
	Allocations int64         `json:"allocations"`
	Active      int64         `json:"active"`
	Hits        int64         `json:"hits"`
	Misses      int64         `json:"misses"`
	HitRate     float64       `json:"hit_rate_percent"`
	AvgTime     time.Duration `json:"avg_allocation_time"`
	MinTime     time.Duration `json:"min_allocation_time"`
	MaxTime     time.Duration `json:"max_allocation_time"`
	Efficiency  float64       `json:"efficiency_score"` // 0-100
}

// NewPoolMetrics creates new pool metrics tracker
func NewPoolMetrics() *PoolMetrics {
	return &PoolMetrics{
		totalAllocations:  make(map[PoolSize]int64),
		activeAllocations: make(map[PoolSize]int64),
		hitCount:          make(map[PoolSize]int64),
		missCount:         make(map[PoolSize]int64),
		sizeDistribution:  make(map[int]int64),
		allocationTime:    make(map[PoolSize]time.Duration),
		minAllocationTime: make(map[PoolSize]time.Duration),
		maxAllocationTime: make(map[PoolSize]time.Duration),
		startTime:         time.Now(),
		lastReset:         time.Now(),
	}
}

// RecordAllocation records an allocation event
func (pm *PoolMetrics) RecordAllocation(tier PoolSize, size int, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.totalAllocations[tier]++
	pm.activeAllocations[tier]++
	pm.sizeDistribution[size]++
	pm.allocationTime[tier] += duration

	// Track min/max allocation times
	if pm.minAllocationTime[tier] == 0 || duration < pm.minAllocationTime[tier] {
		pm.minAllocationTime[tier] = duration
	}
	if duration > pm.maxAllocationTime[tier] {
		pm.maxAllocationTime[tier] = duration
	}
}

// RecordHit records a pool hit
func (pm *PoolMetrics) RecordHit(tier PoolSize) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.hitCount[tier]++
}

// RecordMiss records a pool miss
func (pm *PoolMetrics) RecordMiss(tier PoolSize) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.missCount[tier]++
}

// RecordRelease records an object release
func (pm *PoolMetrics) RecordRelease(tier PoolSize) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.activeAllocations[tier]--
}

// GetUtilizationStats returns comprehensive utilization statistics
func (pm *PoolMetrics) GetUtilizationStats() PoolUtilizationStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := PoolUtilizationStats{
		TierBreakdown:    make(map[PoolSize]TierStats),
		SizeDistribution: make(map[int]int64),
		Uptime:           time.Since(pm.startTime),
	}

	// Aggregate stats across all tiers
	tiers := []PoolSize{PoolTiny, PoolSmall, PoolMedium, PoolLarge, PoolHuge}

	for _, tier := range tiers {
		total := pm.totalAllocations[tier]
		active := pm.activeAllocations[tier]
		hits := pm.hitCount[tier]
		misses := pm.missCount[tier]

		stats.TotalAllocations += total
		stats.ActiveAllocations += active

		// Calculate hit rate for this tier
		var hitRate float64
		attempts := hits + misses
		if attempts > 0 {
			hitRate = float64(hits) / float64(attempts) * 100
		}

		// Calculate average allocation time
		var avgTime time.Duration
		if total > 0 {
			avgTime = pm.allocationTime[tier] / time.Duration(total)
		}

		// Calculate efficiency score (weighted combination of hit rate and speed)
		efficiency := hitRate * 0.7
		if avgTime > 0 {
			timeScore := 100.0 * (1.0 - min(1.0, float64(avgTime.Microseconds())/1000.0))
			efficiency = (efficiency + timeScore*0.3)
		}

		stats.TierBreakdown[tier] = TierStats{
			Allocations: total,
			Active:      active,
			Hits:        hits,
			Misses:      misses,
			HitRate:     hitRate,
			AvgTime:     avgTime,
			MinTime:     pm.minAllocationTime[tier],
			MaxTime:     pm.maxAllocationTime[tier],
			Efficiency:  efficiency,
		}
	}

	// Copy size distribution
	for size, count := range pm.sizeDistribution {
		stats.SizeDistribution[size] = count
	}

	// Calculate overall hit/miss rates
	totalHits := int64(0)
	totalMisses := int64(0)
	for _, tier := range tiers {
		totalHits += pm.hitCount[tier]
		totalMisses += pm.missCount[tier]
	}

	totalAttempts := totalHits + totalMisses
	if totalAttempts > 0 {
		stats.HitRate = float64(totalHits) / float64(totalAttempts) * 100
		stats.MissRate = float64(totalMisses) / float64(totalAttempts) * 100
	}

	// Calculate overall average allocation time
	if stats.TotalAllocations > 0 {
		var totalDuration time.Duration
		for _, tier := range tiers {
			totalDuration += pm.allocationTime[tier]
		}
		stats.AvgAllocationTime = totalDuration / time.Duration(stats.TotalAllocations)
	}

	return stats
}

// Reset resets all metrics
func (pm *PoolMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	tiers := []PoolSize{PoolTiny, PoolSmall, PoolMedium, PoolLarge, PoolHuge}

	for _, tier := range tiers {
		pm.totalAllocations[tier] = 0
		pm.activeAllocations[tier] = 0
		pm.hitCount[tier] = 0
		pm.missCount[tier] = 0
		pm.allocationTime[tier] = 0
		pm.minAllocationTime[tier] = 0
		pm.maxAllocationTime[tier] = 0
	}

	pm.sizeDistribution = make(map[int]int64)
	pm.lastReset = time.Now()
}

// GetUptime returns time since metrics were created
func (pm *PoolMetrics) GetUptime() time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return time.Since(pm.startTime)
}

// GetTimeSinceReset returns time since last reset
func (pm *PoolMetrics) GetTimeSinceReset() time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return time.Since(pm.lastReset)
}

// GetMostCommonSizes returns top N most common allocation sizes
func (pm *PoolMetrics) GetMostCommonSizes(n int) []struct {
	Size  int
	Count int64
} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Sort by count
	type sizeCount struct {
		Size  int
		Count int64
	}

	var sorted []sizeCount
	for size, count := range pm.sizeDistribution {
		sorted = append(sorted, sizeCount{size, count})
	}

	// Simple sort (for production, use more efficient algorithm)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Count < sorted[j].Count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Return top N
	if n > len(sorted) {
		n = len(sorted)
	}

	result := make([]struct {
		Size  int
		Count int64
	}, n)
	for i := 0; i < n; i++ {
		result[i] = struct {
			Size  int
			Count int64
		}{sorted[i].Size, sorted[i].Count}
	}

	return result
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
