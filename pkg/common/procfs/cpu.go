package procfs

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// ReadCPUUsage returns the 1-minute load average from /proc/loadavg (fast, non-blocking)
func ReadCPUUsage() (float64, error) {
	data, err := ioutil.ReadFile("/proc/loadavg")
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
