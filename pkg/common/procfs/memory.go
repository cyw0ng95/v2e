package procfs

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// ReadMemoryUsage reads memory usage from /proc/meminfo
func ReadMemoryUsage() (float64, error) {
	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	total := 0.0
	free := 0.0
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			if fields[0] == "MemTotal:" {
				total, _ = strconv.ParseFloat(fields[1], 64)
			} else if fields[0] == "MemAvailable:" {
				free, _ = strconv.ParseFloat(fields[1], 64)
			}
		}
	}
	if total == 0 {
		return 0, nil
	}
	return ((total - free) / total) * 100, nil
}
