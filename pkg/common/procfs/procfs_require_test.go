package procfs

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadCPUUsage(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReadCPUUsage", nil, func(t *testing.T, tx *gorm.DB) {
		load, err := ReadCPUUsage()
		require.NoError(t, err)
		require.GreaterOrEqual(t, load, 0.0, "load average should be non-negative")
	})

}

func TestReadLoadAvg(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReadLoadAvg", nil, func(t *testing.T, tx *gorm.DB) {
		loads, err := ReadLoadAvg()
		require.NoError(t, err)
		require.Len(t, loads, 3, "expected 3 load average values")
		for i, v := range loads {
			require.GreaterOrEqualf(t, v, 0.0, "load[%d] should be non-negative", i)
		}
	})

}

func TestReadUptime(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReadUptime", nil, func(t *testing.T, tx *gorm.DB) {
		u, err := ReadUptime()
		require.NoError(t, err)
		require.GreaterOrEqual(t, u, 0.0, "uptime should be non-negative")
	})

}

func TestReadMemoryAndSwapUsage(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReadMemoryAndSwapUsage", nil, func(t *testing.T, tx *gorm.DB) {
		mem, err := ReadMemoryUsage()
		require.NoError(t, err)
		require.GreaterOrEqual(t, mem, 0.0)
		require.LessOrEqual(t, mem, 100.0)

		swap, err := ReadSwapUsage()
		require.NoError(t, err)
		require.GreaterOrEqual(t, swap, 0.0)
		require.LessOrEqual(t, swap, 100.0)
	})

}

func TestReadDiskUsageRoot(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReadDiskUsageRoot", nil, func(t *testing.T, tx *gorm.DB) {
		used, total, err := ReadDiskUsage("/")
		require.NoError(t, err)
		require.Greater(t, total, uint64(0), "total bytes should be > 0")
		require.LessOrEqual(t, used, total, "used must be <= total")
	})

}

func TestReadNetDevDetailedAndTotals(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReadNetDevDetailedAndTotals", nil, func(t *testing.T, tx *gorm.DB) {
		detailed, err := ReadNetDevDetailed()
		require.NoError(t, err)
		require.NotNil(t, detailed)

		rxTotal, txTotal, err := ReadNetDev()
		require.NoError(t, err)

		var rxSum, txSum uint64
		for ifName, stats := range detailed {
			if ifName == "lo" {
				continue
			}
			rxSum += stats["rx"]
			txSum += stats["tx"]
		}
		require.Equal(t, rxSum, rxTotal, "aggregated rx bytes should match ReadNetDev total")
		require.Equal(t, txSum, txTotal, "aggregated tx bytes should match ReadNetDev total")
	})

}
