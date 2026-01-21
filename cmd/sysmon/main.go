// Package sysmon implements the sysmon service for monitoring system performance.

package sysmon

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common/procfs"
)

// SysmonService represents the system monitoring service
func SysmonService() {
	log.Println("Sysmon service started")

	// Simulate periodic metrics collection
	for {
		metrics, err := collectMetrics()
		if err != nil {
			log.Printf("Failed to collect metrics: %v", err)
			continue
		}

		jsonMetrics, err := json.Marshal(metrics)
		if err != nil {
			log.Printf("Error marshaling metrics: %v", err)
			continue
		}

		// Simulate writing metrics to stdout (or broker communication)
		os.Stdout.Write(jsonMetrics)
		os.Stdout.Write([]byte("\n"))

		time.Sleep(5 * time.Second) // Collect metrics every 5 seconds
	}
}

func collectMetrics() (map[string]interface{}, error) {
	cpuUsage, err := procfs.ReadCPUUsage()
	if err != nil {
		return nil, err
	}

	memoryUsage, err := procfs.ReadMemoryUsage()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
	}, nil
}
