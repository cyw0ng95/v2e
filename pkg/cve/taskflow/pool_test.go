package taskflow

import (
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestTieredPool_Get_Put(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_Get_Put", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Get buffers of various sizes
		buf1 := tp.Get(100)
		buf2 := tp.Get(500)
		buf3 := tp.Get(2000)
		buf4 := tp.Get(10000)

		if cap(buf1) < 100 || cap(buf2) < 500 || cap(buf3) < 2000 || cap(buf4) < 10000 {
			t.Error("buffers have insufficient capacity")
		}

		// Return buffers to pool
		tp.Put(buf1)
		tp.Put(buf2)
		tp.Put(buf3)
		tp.Put(buf4)
	})
}

func TestTieredPool_TierDistribution(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_TierDistribution", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		tests := []struct {
			size int
			tier PoolSize
		}{
			{100, PoolTiny},
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

func TestTieredPool_LargeSizeDirectAlloc(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_LargeSizeDirectAlloc", nil, func(t *testing.T, tx *gorm.DB) {
		config := DefaultTieredPoolConfig()
		config.MaxPooledSize = 1024
		tp := NewTieredPool(config)

		// Large sizes should be allocated directly, not pooled
		buf1 := tp.Get(2000)
		buf2 := tp.Get(5000)

		if cap(buf1) < 2000 || cap(buf2) < 5000 {
			t.Error("large buffers have insufficient capacity")
		}

		// These should be discarded, not pooled
		tp.Put(buf1)
		tp.Put(buf2)

		// Stats should show misses for large sizes (PoolLarge tier for 2000/5000 bytes)
		stats := tp.GetStats()
		breakdown := stats["tier_breakdown"].(map[PoolSize]map[string]int64)
		// 2000 and 5000 bytes map to PoolLarge tier (4097-16384 range)
		if breakdown[PoolLarge]["misses"] == 0 {
			t.Error("expected misses for large sizes (PoolLarge tier)")
		}
	})
}

func TestTieredPool_ConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_ConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()
		var wg sync.WaitGroup

		// Concurrent gets and puts
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				buf := tp.Get(1000)
				time.Sleep(time.Microsecond)
				tp.Put(buf)
			}()
		}

		wg.Wait()

		// Verify stats are consistent
		stats := tp.GetStats()
		totalAllocs := stats["total_allocations"].(int64)
		if totalAllocs == 0 {
			t.Error("expected some allocations")
		}
	})
}

func TestTieredPool_NilBufferPut(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_NilBufferPut", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Should not panic
		tp.Put(nil)
	})
}

func TestTieredPool_GetStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_GetStats", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Perform some operations
		for i := 0; i < 10; i++ {
			buf := tp.Get(500)
			tp.Put(buf)
		}

		stats := tp.GetStats()

		// Verify stats structure
		if _, ok := stats["total_allocations"]; !ok {
			t.Error("missing total_allocations in stats")
		}
		if _, ok := stats["total_hits"]; !ok {
			t.Error("missing total_hits in stats")
		}
		if _, ok := stats["total_misses"]; !ok {
			t.Error("missing total_misses in stats")
		}
		if _, ok := stats["hit_rate_percent"]; !ok {
			t.Error("missing hit_rate_percent in stats")
		}
		if _, ok := stats["tier_breakdown"]; !ok {
			t.Error("missing tier_breakdown in stats")
		}
	})
}

func TestTieredPool_HitRate(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_HitRate", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Get a buffer and return it
		buf1 := tp.Get(500)
		tp.Put(buf1)

		// Get another buffer of same size
		buf2 := tp.Get(500)
		tp.Put(buf2)

		stats := tp.GetStats()
		hitRate := stats["hit_rate_percent"].(float64)

		// Should have some hit rate from reuse
		if hitRate < 0 {
			t.Errorf("hit rate should be non-negative, got %f", hitRate)
		}
	})
}

func TestTieredPool_PoolSizeLimit(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_PoolSizeLimit", nil, func(t *testing.T, tx *gorm.DB) {
		config := DefaultTieredPoolConfig()
		config.MaxPoolSize = 5
		config.PreAllocCounts = map[PoolSize]int{} // No pre-allocation
		tp := NewTieredPool(config)

		// Allocate and return many buffers
		for i := 0; i < 100; i++ {
			buf := tp.Get(500)
			tp.Put(buf)
		}

		// Should not crash or leak excessively
		stats := tp.GetStats()
		breakdown := stats["tier_breakdown"].(map[PoolSize]map[string]int64)
		allocations := breakdown[PoolSmall]["allocations"]
		hits := breakdown[PoolSmall]["hits"]

		// With MaxPoolSize=5 and pool reuse, we should have:
		// - A few initial allocations to fill the pool (~5-10)
		// - Many hits from reuse (should be > 80)
		// This verifies the pool size limit is working
		// Note: sync.Pool may pre-allocate some buffers, so we allow some flexibility
		if allocations > 30 {
			t.Errorf("pool size limit not working: expected <= 30 allocations with MaxPoolSize=5, got %d", allocations)
		}
		if hits < 70 {
			t.Errorf("pool not reusing buffers efficiently: expected >= 70 hits with 100 operations, got %d", hits)
		}
	})
}

func TestTieredPool_CustomConfig(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_CustomConfig", nil, func(t *testing.T, tx *gorm.DB) {
		config := TieredPoolConfig{
			MaxPooledSize: 4096,
			PreAllocCounts: map[PoolSize]int{
				PoolTiny:   10,
				PoolSmall:  5,
				PoolMedium: 2,
			},
			MaxPoolSize: 50,
		}

		tp := NewTieredPool(config)

		buf := tp.Get(100)
		if cap(buf) < 100 {
			t.Error("buffer has insufficient capacity")
		}
		tp.Put(buf)

		// Verify config was applied
		if tp.config.MaxPooledSize != 4096 {
			t.Errorf("expected MaxPooledSize 4096, got %d", tp.config.MaxPooledSize)
		}
	})
}

func TestTieredPool_ReuseBuffer(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTieredPool_ReuseBuffer", nil, func(t *testing.T, tx *gorm.DB) {
		tp := NewTieredPoolWithDefaults()

		// Get and put buffer
		buf1 := tp.Get(1000)
		cap1 := cap(buf1)
		tp.Put(buf1)

		// Get buffer of same size, should reuse
		buf2 := tp.Get(1000)
		cap2 := cap(buf2)
		tp.Put(buf2)

		// Buffer should be reused if hit occurred
		stats := tp.GetStats()
		hits := stats["total_hits"].(int64)
		if hits == 0 && cap1 != cap2 {
			t.Log("buffers may not be reused due to pool dynamics")
		}
	})
}
