package procfs

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// ReadMemoryUsage reads memory usage from /proc/meminfo
func ReadMemoryUsage() (float64, error) {
	f, err := ioutil.ReadFile("/proc/meminfo")
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
