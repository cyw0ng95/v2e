package procfs

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// ReadCPUUsage reads CPU usage from /proc/stat
func ReadCPUUsage() (float64, error) {
	data, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "cpu" {
			total := 0.0
			for _, value := range fields[1:] {
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return 0, err
				}
				total += v
			}
			return total, nil
		}
	}
	return 0, nil
}
