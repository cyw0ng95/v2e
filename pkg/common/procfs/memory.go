package procfs

import (
	"os"
	"strconv"
	"strings"
)

// ReadMemoryUsage reads memory usage from /proc/meminfo
func ReadMemoryUsage() (float64, error) {
	f, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	var total, free float64
	foundTotal, foundFree := false, false
	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				total, _ = strconv.ParseFloat(fields[1], 64)
				foundTotal = true
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				free, _ = strconv.ParseFloat(fields[1], 64)
				foundFree = true
			}
		}
		if foundTotal && foundFree {
			break
		}
	}
	if total == 0 {
		return 0, nil
	}
	return ((total - free) / total) * 100, nil
}

// ReadSwapUsage returns swap usage percentage (0..100). If no swap, returns 0.
func ReadSwapUsage() (float64, error) {
	f, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	var total, free float64
	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, "SwapTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				total, _ = strconv.ParseFloat(fields[1], 64)
			}
		} else if strings.HasPrefix(line, "SwapFree:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				free, _ = strconv.ParseFloat(fields[1], 64)
			}
		}
		if total > 0 && free >= 0 {
			break
		}
	}
	if total == 0 {
		return 0, nil
	}
	// meminfo reports kB, convert via percent
	used := ((total - free) / total) * 100
	return used, nil
}
