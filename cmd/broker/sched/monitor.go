// Package sched provides system and application monitoring capabilities
package sched

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// SystemMonitor collects system-level metrics
type SystemMonitor struct {
	ctx         context.Context
	cancel      context.CancelFunc
	interval    time.Duration
	callback    func(LoadMetrics)
	
	// Atomic counters for metrics
	messageQueueDepth int64
	messageThroughput int64
	lastThroughputSample int64
	lastThroughputTime time.Time
	
	// Latency tracking
	latencySamples []time.Duration
	latencyMu      sync.RWMutex
	
	// Process resource tracking
	processMetrics map[string]*ProcessResourceMetrics
	processMetricsMu sync.RWMutex
	
	running int32
}

// NewSystemMonitor creates a new system monitor with the specified sampling interval
func NewSystemMonitor(interval time.Duration) *SystemMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SystemMonitor{
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		lastThroughputTime: time.Now(),
		latencySamples: make([]time.Duration, 0),
		processMetrics: make(map[string]*ProcessResourceMetrics),
	}
}

// SetCallback sets the callback function to receive collected metrics
func (sm *SystemMonitor) SetCallback(callback func(LoadMetrics)) {
	sm.callback = callback
}

// Start begins periodic metric collection
func (sm *SystemMonitor) Start() {
	if !atomic.CompareAndSwapInt32(&sm.running, 0, 1) {
		return // Already running
	}
	
	go sm.collectLoop()
}

// Stop stops periodic metric collection
func (sm *SystemMonitor) Stop() {
	if !atomic.CompareAndSwapInt32(&sm.running, 1, 0) {
		return // Already stopped
	}
	
	sm.cancel()
}

// collectLoop runs the periodic metric collection loop
func (sm *SystemMonitor) collectLoop() {
	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			metrics := sm.CollectMetrics()
			
			if sm.callback != nil {
				sm.callback(metrics)
			}
		case <-sm.ctx.Done():
			return
		}
	}
}

// CollectMetrics gathers all system and application metrics
func (sm *SystemMonitor) CollectMetrics() LoadMetrics {
	cpuUtil, memUtil := sm.getSystemMetrics()
	queueDepth := atomic.LoadInt64(&sm.messageQueueDepth)
	
	// Calculate throughput (messages per second)
	now := time.Now()
	lastSample := atomic.LoadInt64(&sm.lastThroughputSample)
	currentThroughput := atomic.LoadInt64(&sm.messageThroughput)
	
	timeDiff := now.Sub(sm.lastThroughputTime).Seconds()
	if timeDiff > 0 {
		calculatedThroughput = float64(currentThroughput-lastSample) / timeDiff
		
		// Update for next calculation
		atomic.StoreInt64(&sm.lastThroughputSample, currentThroughput)
		sm.lastThroughputTime = now
	}
	
	// Calculate average latency from samples
	avgLatency := sm.calculateAverageLatency()
	
	// Get process resource usage
	processUsage := sm.getProcessResourceUsage()
	
	return LoadMetrics{
		CPUUtilization:      cpuUtil,
		MemoryUtilization:   memUtil,
		MessageQueueDepth:   queueDepth,
		MessageThroughput:   calculatedThroughput,
		AverageLatency:      avgLatency,
		ActiveConnections:   len(processUsage), // Approximation
		SystemLoadAvg:       sm.getSystemLoadAverage(),
		ProcessResourceUsage: processUsage,
	}
}

// getSystemMetrics retrieves CPU and memory utilization
func (sm *SystemMonitor) getSystemMetrics() (cpuUtil, memUtil float64) {
	// In a real implementation, this would use platform-specific APIs
	// For now, we'll return dummy values
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Calculate memory utilization as a percentage of heap usage
	// This is a simplified approximation
	memUtil = float64(m.Alloc) / float64(m.Sys) * 100.0
	if memUtil > 100.0 {
		memUtil = 100.0
	}
	
	// For CPU utilization, we'd typically use platform-specific APIs
	// Returning a dummy value for now
	cpuUtil = 50.0 // Placeholder value
	
	return cpuUtil, memUtil
}

// getSystemLoadAverage returns the system load average
func (sm *SystemMonitor) getSystemLoadAverage() float64 {
	// In a real implementation, this would retrieve actual system load average
	// Returning a dummy value for now
	return 1.0
}

// UpdateMessageQueueDepth updates the message queue depth counter
func (sm *SystemMonitor) UpdateMessageQueueDepth(depth int64) {
	atomic.StoreInt64(&sm.messageQueueDepth, depth)
}

// RecordMessage increments the message throughput counter
func (sm *SystemMonitor) RecordMessage() {
	atomic.AddInt64(&sm.messageThroughput, 1)
}

// GetThroughput returns the current throughput measurement
func (sm *SystemMonitor) GetThroughput() float64 {
	return float64(atomic.LoadInt64(&sm.messageThroughput))
}

// AddLatencySample adds a latency sample for average calculation
func (sm *SystemMonitor) AddLatencySample(duration time.Duration) {
	sm.latencyMu.Lock()
	defer sm.latencyMu.Unlock()
	
	// Keep only the last N samples to prevent unbounded growth
	const maxSamples = 100
	if len(sm.latencySamples) >= maxSamples {
		sm.latencySamples = sm.latencySamples[1:] // Remove oldest sample
	}
	sm.latencySamples = append(sm.latencySamples, duration)
}

// calculateAverageLatency calculates the average latency from samples
func (sm *SystemMonitor) calculateAverageLatency() time.Duration {
	sm.latencyMu.RLock()
	defer sm.latencyMu.RUnlock()
	
	if len(sm.latencySamples) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, sample := range sm.latencySamples {
		total += sample
	}
	return total / time.Duration(len(sm.latencySamples))
}

// UpdateProcessMetrics updates metrics for a specific process
func (sm *SystemMonitor) UpdateProcessMetrics(processID string, metrics ProcessResourceMetrics) {
	sm.processMetricsMu.Lock()
	defer sm.processMetricsMu.Unlock()
	
	sm.processMetrics[processID] = &metrics
}

// getProcessResourceUsage returns a copy of process resource usage metrics
func (sm *SystemMonitor) getProcessResourceUsage() map[string]ProcessResourceMetrics {
	sm.processMetricsMu.RLock()
	defer sm.processMetricsMu.RUnlock()
	
	result := make(map[string]ProcessResourceMetrics)
	for id, metrics := range sm.processMetrics {
		if metrics != nil {
			result[id] = *metrics
		}
	}
	return result
}

// Monitorable represents an interface for components that can be monitored
type Monitorable interface {
	RegisterMonitor(monitor *SystemMonitor)
	GetMetrics() LoadMetrics
}