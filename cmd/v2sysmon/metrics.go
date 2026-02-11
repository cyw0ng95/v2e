package main

import (
	"github.com/cyw0ng95/v2e/pkg/common/procfs"
)

// metricCollector is a function that collects a specific metric and adds it to the metrics map.
type metricCollector func(m map[string]interface{}) error

// collectMetric is a helper function that executes a metric collector and handles errors.
// If the collector fails, the error is logged but the metric is simply omitted from the result.
func collectMetric(m map[string]interface{}, collector metricCollector) {
	collector(m)
}

// Metric collectors

// collectCPUMetric collects CPU usage metrics.
func collectCPUMetric(m map[string]interface{}) error {
	cpuUsage, err := procfs.ReadCPUUsage()
	if err != nil {
		return err
	}
	m["cpu_usage"] = cpuUsage
	return nil
}

// collectMemoryMetric collects memory usage metrics.
func collectMemoryMetric(m map[string]interface{}) error {
	memoryUsage, err := procfs.ReadMemoryUsage()
	if err != nil {
		return err
	}
	m["memory_usage"] = memoryUsage
	return nil
}

// collectLoadAvgMetric collects load average metrics.
func collectLoadAvgMetric(m map[string]interface{}) error {
	loadAvg, err := procfs.ReadLoadAvg()
	if err != nil {
		return err
	}
	m["load_avg"] = loadAvg
	return nil
}

// collectUptimeMetric collects system uptime metrics.
func collectUptimeMetric(m map[string]interface{}) error {
	up, err := procfs.ReadUptime()
	if err != nil {
		return err
	}
	m["uptime"] = up
	return nil
}

// collectDiskMetric collects disk usage metrics for the root filesystem.
func collectDiskMetric(m map[string]interface{}) error {
	used, total, err := procfs.ReadDiskUsage("/")
	if err != nil {
		return err
	}
	// Provide object-style disk info keyed by mount path, and keep totals for compatibility
	m["disk"] = map[string]map[string]uint64{"/": {"used": used, "total": total}}
	m["disk_usage"] = used
	m["disk_total"] = total
	return nil
}

// collectSwapMetric collects swap usage metrics.
func collectSwapMetric(m map[string]interface{}) error {
	swap, err := procfs.ReadSwapUsage()
	if err != nil {
		return err
	}
	m["swap_usage"] = swap
	return nil
}

// collectNetworkMetric collects network device metrics with totals.
func collectNetworkMetric(m map[string]interface{}) error {
	netMap, err := procfs.ReadNetDevDetailed()
	if err != nil {
		return err
	}
	// Also provide totals for compatibility
	var totalRx, totalTx uint64
	for ifName, s := range netMap {
		if ifName == "lo" {
			continue
		}
		if v, ok := s["rx"]; ok {
			totalRx += v
		}
		if v, ok := s["tx"]; ok {
			totalTx += v
		}
	}
	m["network"] = netMap
	m["net_rx"] = totalRx
	m["net_tx"] = totalTx
	return nil
}

// List of all metric collectors in order of priority.
var metricCollectors = []struct {
	name string
	fn   metricCollector
}{
	{"cpu", collectCPUMetric},
	{"memory", collectMemoryMetric},
	{"load_avg", collectLoadAvgMetric},
	{"uptime", collectUptimeMetric},
	{"disk", collectDiskMetric},
	{"swap", collectSwapMetric},
	{"network", collectNetworkMetric},
}

// requiredMetricCollectors contains metric collectors that must succeed.
// If these fail, the entire metrics collection fails.
var requiredMetricCollectors = map[string]bool{
	"cpu":    true,
	"memory": true,
}

// collectAllMetrics collects all metrics using the registered collectors.
// Returns a map of metric names to values, or an error if a required metric collection fails.
func collectAllMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	var firstErr error

	for _, mc := range metricCollectors {
		err := mc.fn(metrics)
		if err != nil {
			// Log and continue for optional metrics
			if !requiredMetricCollectors[mc.name] {
				continue
			}
			// Fail on first required metric error
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return metrics, nil
}
