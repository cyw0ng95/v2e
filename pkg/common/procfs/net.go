package procfs

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// ReadNetDevDetailed returns per-interface rx/tx bytes as a map keyed by interface name.
func ReadNetDevDetailed() (map[string]map[string]uint64, error) {
	data, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	result := make(map[string]map[string]uint64)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Inter-") || strings.HasPrefix(line, "face") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		ifName := strings.TrimSpace(parts[0])
		fields := strings.Fields(strings.TrimSpace(parts[1]))
		if len(fields) < 9 {
			continue
		}
		r, _ := strconv.ParseUint(fields[0], 10, 64) // rx_bytes
		t, _ := strconv.ParseUint(fields[8], 10, 64) // tx_bytes
		result[ifName] = map[string]uint64{"rx": r, "tx": t}
	}
	return result, nil
}

// ReadNetDev returns total rx and tx bytes across all interfaces (excluding lo)
func ReadNetDev() (rx uint64, tx uint64, err error) {
	m, err := ReadNetDevDetailed()
	if err != nil {
		return 0, 0, err
	}
	for ifName, stats := range m {
		if ifName == "lo" {
			continue
		}
		rx += stats["rx"]
		tx += stats["tx"]
	}
	return rx, tx, nil
}
