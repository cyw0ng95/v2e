package taskflow

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestPoolMetrics_RecordAllocation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_RecordAllocation", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)

		stats := pm.GetUtilizationStats()
		if stats.TotalAllocations != 1 {
			t.Errorf("expected 1 allocation, got %d", stats.TotalAllocations)
		}
	})
}

func TestPoolMetrics_HitMissTracking(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_HitMissTracking", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordHit(PoolSmall)
		pm.RecordHit(PoolSmall)
		pm.RecordMiss(PoolSmall)

		stats := pm.GetUtilizationStats()
		tierStats := stats.TierBreakdown[PoolSmall]

		if tierStats.Hits != 2 {
			t.Errorf("expected 2 hits, got %d", tierStats.Hits)
		}
		if tierStats.Misses != 1 {
			t.Errorf("expected 1 miss, got %d", tierStats.Misses)
		}
		if tierStats.HitRate != 66.66666666666666 {
			t.Errorf("expected 66.67%% hit rate, got %f", tierStats.HitRate)
		}
	})
}

func TestPoolMetrics_ActiveTracking(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_ActiveTracking", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)
		pm.RecordAllocation(PoolMedium, 5000, 200*time.Microsecond)

		stats := pm.GetUtilizationStats()
		if stats.ActiveAllocations != 2 {
			t.Errorf("expected 2 active, got %d", stats.ActiveAllocations)
		}

		pm.RecordRelease(PoolSmall)
		stats = pm.GetUtilizationStats()
		if stats.ActiveAllocations != 1 {
			t.Errorf("expected 1 active after release, got %d", stats.ActiveAllocations)
		}
	})
}

func TestPoolMetrics_SizeDistribution(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_SizeDistribution", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 100, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 100, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 500, 100*time.Microsecond)

		stats := pm.GetUtilizationStats()
		if stats.SizeDistribution[100] != 2 {
			t.Errorf("expected 2 allocations of size 100, got %d", stats.SizeDistribution[100])
		}
		if stats.SizeDistribution[500] != 1 {
			t.Errorf("expected 1 allocation of size 500, got %d", stats.SizeDistribution[500])
		}
	})
}

func TestPoolMetrics_TimingStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_TimingStats", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		durations := []time.Duration{100 * time.Microsecond, 200 * time.Microsecond, 300 * time.Microsecond}
		for _, d := range durations {
			pm.RecordAllocation(PoolSmall, 1000, d)
		}

		stats := pm.GetUtilizationStats()
		tierStats := stats.TierBreakdown[PoolSmall]

		if tierStats.MinTime != 100*time.Microsecond {
			t.Errorf("expected min 100µs, got %v", tierStats.MinTime)
		}
		if tierStats.MaxTime != 300*time.Microsecond {
			t.Errorf("expected max 300µs, got %v", tierStats.MaxTime)
		}
		expectedAvg := 200 * time.Microsecond
		if tierStats.AvgTime != expectedAvg {
			t.Errorf("expected avg %v, got %v", expectedAvg, tierStats.AvgTime)
		}
	})
}

func TestPoolMetrics_Reset(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_Reset", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)
		pm.RecordHit(PoolSmall)
		pm.RecordMiss(PoolSmall)

		pm.Reset()

		stats := pm.GetUtilizationStats()
		if stats.TotalAllocations != 0 {
			t.Errorf("expected 0 allocations after reset, got %d", stats.TotalAllocations)
		}
		if stats.ActiveAllocations != 0 {
			t.Errorf("expected 0 active after reset, got %d", stats.ActiveAllocations)
		}
	})
}

func TestPoolMetrics_Uptime(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_Uptime", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		uptime := pm.GetUptime()
		if uptime <= 0 {
			t.Error("expected positive uptime")
		}

		time.Sleep(10 * time.Millisecond)

		newUptime := pm.GetUptime()
		if newUptime <= uptime {
			t.Error("uptime should increase over time")
		}
	})
}

func TestPoolMetrics_TimeSinceReset(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_TimeSinceReset", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)

		time.Sleep(10 * time.Millisecond)

		pm.Reset()

		timeSinceReset := pm.GetTimeSinceReset()
		if timeSinceReset <= 0 {
			t.Error("expected positive time since reset")
		}
	})
}

func TestPoolMetrics_GetMostCommonSizes(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_GetMostCommonSizes", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 100, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 100, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 100, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 500, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 500, 100*time.Microsecond)
		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)

		common := pm.GetMostCommonSizes(3)

		if len(common) != 3 {
			t.Fatalf("expected 3 results, got %d", len(common))
		}

		if common[0].Size != 100 || common[0].Count != 3 {
			t.Errorf("expected size 100 with count 3, got size %d with count %d", common[0].Size, common[0].Count)
		}
	})
}

func TestPoolMetrics_EfficiencyScore(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_EfficiencyScore", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		// High hit rate = high efficiency
		for i := 0; i < 90; i++ {
			pm.RecordHit(PoolSmall)
		}
		for i := 0; i < 10; i++ {
			pm.RecordMiss(PoolSmall)
		}
		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)

		stats := pm.GetUtilizationStats()
		tierStats := stats.TierBreakdown[PoolSmall]

		if tierStats.Efficiency < 60 {
			t.Errorf("expected efficiency > 60%%, got %f", tierStats.Efficiency)
		}
	})
}

func TestPoolMetrics_ConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_ConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)
				pm.RecordHit(PoolSmall)
				pm.RecordRelease(PoolSmall)
			}()
		}
		wg.Wait()

		stats := pm.GetUtilizationStats()
		if stats.TotalAllocations == 0 {
			t.Error("expected some allocations")
		}
	})
}

func TestPoolMetrics_OverallStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPoolMetrics_OverallStats", nil, func(t *testing.T, tx *gorm.DB) {
		pm := NewPoolMetrics()

		pm.RecordAllocation(PoolSmall, 1000, 100*time.Microsecond)
		pm.RecordAllocation(PoolMedium, 5000, 200*time.Microsecond)
		pm.RecordHit(PoolSmall)
		pm.RecordHit(PoolMedium)
		pm.RecordMiss(PoolSmall)

		stats := pm.GetUtilizationStats()

		if stats.TotalAllocations != 2 {
			t.Errorf("expected 2 total allocations, got %d", stats.TotalAllocations)
		}
		// 2 hits out of 3 total attempts = 66.67% (use tolerance for float comparison)
		expectedHitRate := 100.0 * 2.0 / 3.0
		if math.Abs(stats.HitRate-expectedHitRate) > 0.01 {
			t.Errorf("expected %.2f%% hit rate, got %f", expectedHitRate, stats.HitRate)
		}
		// 1 miss out of 3 total attempts = 33.33%
		expectedMissRate := 100.0 * 1.0 / 3.0
		if math.Abs(stats.MissRate-expectedMissRate) > 0.01 {
			t.Errorf("expected %.2f%% miss rate, got %f", expectedMissRate, stats.MissRate)
		}
	})
}
