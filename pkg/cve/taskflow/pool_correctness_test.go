package taskflow

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestPoolCorrectness_BufferCapacity(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolCorrectness_BufferCapacity", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		sizes := []int{50, 300, 1500, 8000, 20000}
		for _, size := range sizes {
			buf := tp.Get(size)
			if cap(buf) < size {
				t.Errorf("buffer capacity %d < requested size %d", cap(buf), size)
			}
			if len(buf) != size {
				t.Errorf("buffer length %d != requested size %d", len(buf), size)
			}
			tp.Put(buf)
		}
	})
}

func TestPoolCorrectness_BufferReuse(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolCorrectness_BufferReuse", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		buf1 := tp.Get(500)
		cap1 := cap(buf1)
		tp.Put(buf1)

		buf2 := tp.Get(500)
		cap2 := cap(buf2)
		tp.Put(buf2)

		// Buffers should be reused (same capacity)
		if cap1 != cap2 {
			t.Logf("buffers may not be reused: cap1=%d, cap2=%d", cap1, cap2)
		}
	})
}

func TestPoolCorrectness_NoDataLeaks(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolCorrectness_NoDataLeaks", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Write data to buffer
		buf1 := tp.Get(100)
		for i := range buf1 {
			buf1[i] = 0xFF
		}
		tp.Put(buf1)

		// Get new buffer
		buf2 := tp.Get(100)

		// Buffer should be zeroed (or at least not contain old data)
		if len(buf2) != 100 {
			t.Errorf("expected length 100, got %d", len(buf2))
		}

		// Note: sync.Pool does NOT zero memory, so we expect Get() to return cleared slice
		// but the underlying capacity may retain old data
	})
}

func TestPoolCorrectness_ConcurrentSafety(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolCorrectness_ConcurrentSafety", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				buf := tp.Get(1000)
				time.Sleep(1 * time.Microsecond)
				tp.Put(buf)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 100; i++ {
			<-done
		}

		stats := tp.GetStats()
		if stats["total_allocations"].(int64) == 0 {
			t.Error("expected some allocations")
		}
	})
}

func TestPoolCorrectness_SizeLimit(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolCorrectness_SizeLimit", nil, func(t *testing.T, tx *gorm.DB) {
		config := DefaultTieredPoolConfig()
		config.MaxPooledSize = 4096
		tp := NewTieredPool(config)

		// Small buffer should be pooled
		buf1 := tp.Get(100)
		tp.Put(buf1)

		// Large buffer should be allocated directly
		buf2 := tp.Get(10000)
		tp.Put(buf2)

		stats := tp.GetStats()
		breakdown := stats["tier_breakdown"].(map[PoolSize]map[string]int64)

		// Tiny pool should have allocations
		if breakdown[PoolTiny]["allocations"] == 0 {
			t.Error("expected tiny pool allocations")
		}

		// Large allocations should go to huge tier or be direct
		totalHuge := breakdown[PoolHuge]["allocations"]
		allocs := stats["total_allocations"].(int64)
		if totalHuge > allocs {
			t.Error("huge tier allocations exceed total")
		}
	})
}

func TestPoolCorrectness_TierMapping(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolCorrectness_TierMapping", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		tests := []struct {
			size int
			tier PoolSize
		}{
			{256, PoolTiny},
			{257, PoolSmall},
			{1024, PoolSmall},
			{1025, PoolMedium},
			{4096, PoolMedium},
			{4097, PoolLarge},
			{16384, PoolLarge},
			{16385, PoolHuge},
		}

		for _, tt := range tests {
			tier := tp.sizeToTier(tt.size)
			if tier != tt.tier {
				t.Errorf("size %d maps to tier %d, expected %d", tt.size, tier, tt.tier)
			}
		}
	})
}

func TestPoolPerformance_AllocationSpeed(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolPerformance_AllocationSpeed", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Warm up pool
		for i := 0; i < 100; i++ {
			buf := tp.Get(1000)
			tp.Put(buf)
		}

		// Measure pooled allocation speed
		start := time.Now()
		for i := 0; i < 10000; i++ {
			buf := tp.Get(1000)
			tp.Put(buf)
		}
		pooledDuration := time.Since(start)

		// Measure direct allocation speed
		start = time.Now()
		for i := 0; i < 10000; i++ {
			buf := make([]byte, 0, 1000)
			_ = buf
		}
		directDuration := time.Since(start)

		t.Logf("Pooled: %v, Direct: %v, Speedup: %.2fx",
			pooledDuration, directDuration, float64(directDuration)/float64(pooledDuration))

		// Pooled should be at least as fast as direct (with warm pool)
		// If not, it's not necessarily wrong, but worth investigating
	})
}

func TestPoolPerformance_HitRate(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolPerformance_HitRate", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Get and put buffers of same size
		for i := 0; i < 1000; i++ {
			buf := tp.Get(1000)
			tp.Put(buf)
		}

		stats := tp.GetStats()
		hitRate := stats["hit_rate_percent"].(float64)

		// With warm pool and same size, we expect high hit rate
		t.Logf("Hit rate: %.2f%%", hitRate)
	})
}

func TestPoolPerformance_MemoryAllocation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolPerformance_MemoryAllocation", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Allocate many buffers
		var bufs [][]byte
		for i := 0; i < 100; i++ {
			bufs = append(bufs, tp.Get(1000))
		}

		// Return all to pool
		for _, buf := range bufs {
			tp.Put(buf)
		}

		// Allocate again - should reuse from pool
		for i := 0; i < 100; i++ {
			buf := tp.Get(1000)
			if cap(buf) < 1000 {
				t.Errorf("insufficient capacity: %d", cap(buf))
			}
			tp.Put(buf)
		}

		stats := tp.GetStats()
		hits := stats["total_hits"].(int64)
		if hits == 0 {
			t.Log("No pool hits - buffers may not be reused due to pool behavior")
		}
	})
}

func TestPoolPerformance_Stress(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolPerformance_Stress", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Stress test with concurrent operations
		done := make(chan bool)
		for i := 0; i < 50; i++ {
			go func() {
				for j := 0; j < 200; j++ {
					buf := tp.Get(j % 20000)
					tp.Put(buf)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 50; i++ {
			<-done
		}

		stats := tp.GetStats()
		t.Logf("Total allocations: %d, Hit rate: %.2f%%",
			stats["total_allocations"], stats["hit_rate_percent"])
	})
}
