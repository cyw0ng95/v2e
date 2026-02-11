package procfs

import (
	"os"
	"strconv"
	"strings"
)

// ReadCPUUsage returns the 1-minute load average from /proc/loadavg (fast, non-blocking)
func ReadCPUUsage() (float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, nil
	}
	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}
	return load, nil
}

// ReadLoadAvg returns the 1, 5 and 15 minute load averages from /proc/loadavg
func ReadLoadAvg() ([]float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return nil, nil
	}
	vals := make([]float64, 0, 3)
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			vals = append(vals, 0)
			continue
		}
		vals = append(vals, v)
	}
	return vals, nil
}

// ReadUptime returns system uptime in seconds from /proc/uptime
func ReadUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, nil
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}
