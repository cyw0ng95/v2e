package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ProcessMetrics struct {
	PID       int       `json:"pid"`
	ID        string    `json:"id"`
	VmRSS     uint64    `json:"vmrss"`
	VmSize    uint64    `json:"vmsize"`
	Threads   int       `json:"threads"`
	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
}

func ReadProcessMetrics(pid int, processID string) (*ProcessMetrics, error) {
	// Read /proc/{pid}/status
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/%d/status: %w", pid, err)
	}

	metrics := &ProcessMetrics{
		PID:       pid,
		ID:        processID,
		Timestamp: time.Now(),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					metrics.VmRSS = val
				}
			}
		} else if strings.HasPrefix(line, "VmSize:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					metrics.VmSize = val
				}
			}
		} else if strings.HasPrefix(line, "Threads:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.Atoi(fields[1])
				if err == nil {
					metrics.Threads = val
				}
			}
		} else if strings.HasPrefix(line, "State:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				metrics.State = fields[1]
			}
		}
	}

	return metrics, nil
}

func ReadProcessMetricsByID(processID string) (*ProcessMetrics, error) {
	// Find PID from process name in /proc
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pidStr := entry.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Read cmdline to find process name
		cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err != nil {
			continue
		}

		// cmdline is null-separated, find our process
		processName := strings.ReplaceAll(string(cmdline), "\x00", " ")
		if strings.Contains(processName, processID) || strings.Contains(processName, "v2"+processID) {
			return ReadProcessMetrics(pid, processID)
		}
	}

	return nil, fmt.Errorf("process %s not found", processID)
}
