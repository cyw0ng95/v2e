package perf

import (
	"sync"
	"time"
)

// KernelMetrics represents the current kernel performance metrics
type KernelMetrics struct {
	P99LatencyMs     float64 `json:"p99_latency_ms"`
	BufferSaturation float64 `json:"buffer_saturation"`
	ActiveWorkers    int     `json:"active_workers"`
	TotalPermits     int     `json:"total_permits"`
	AllocatedPermits int     `json:"allocated_permits"`
	AvailablePermits int     `json:"available_permits"`
	MessageRate      float64 `json:"message_rate"`
	ErrorRate        float64 `json:"error_rate"`
}

// MetricsCollector collects and tracks kernel performance metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// Latency tracking (rolling window)
	latencies    []time.Duration
	maxLatencies int
	latencyIndex int

	// Buffer saturation
	bufferCapacity int
	currentBuffer  int

	// Message and error rates
	messageCount int64
	errorCount   int64
	lastReset    time.Time
	windowSize   time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(maxLatencies, bufferCapacity int) *MetricsCollector {
	if maxLatencies <= 0 {
		maxLatencies = 1000 // Default to tracking last 1000 latencies
	}
	if bufferCapacity <= 0 {
		bufferCapacity = 1000
	}

	return &MetricsCollector{
		latencies:      make([]time.Duration, maxLatencies),
		maxLatencies:   maxLatencies,
		bufferCapacity: bufferCapacity,
		lastReset:      time.Now(),
		windowSize:     time.Second, // 1-second window for rate calculations
	}
}

// RecordLatency records a message processing latency
func (mc *MetricsCollector) RecordLatency(latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.latencies[mc.latencyIndex] = latency
	mc.latencyIndex = (mc.latencyIndex + 1) % mc.maxLatencies
}

// RecordMessage increments the message count
func (mc *MetricsCollector) RecordMessage() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.messageCount++
	mc.checkReset()
}

// RecordError increments the error count
func (mc *MetricsCollector) RecordError() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.errorCount++
	mc.checkReset()
}

// UpdateBuffer updates the current buffer size
func (mc *MetricsCollector) UpdateBuffer(size int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.currentBuffer = size
}

// checkReset resets counters if window has elapsed (must hold lock)
func (mc *MetricsCollector) checkReset() {
	elapsed := time.Since(mc.lastReset)
	if elapsed >= mc.windowSize {
		mc.messageCount = 0
		mc.errorCount = 0
		mc.lastReset = time.Now()
	}
}

// GetP99Latency calculates the 99th percentile latency
func (mc *MetricsCollector) GetP99Latency() time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Copy latencies for sorting
	latencies := make([]time.Duration, len(mc.latencies))
	copy(latencies, mc.latencies)

	// Filter out zero values (unfilled slots)
	var valid []time.Duration
	for _, lat := range latencies {
		if lat > 0 {
			valid = append(valid, lat)
		}
	}

	if len(valid) == 0 {
		return 0
	}

	// Sort latencies
	sortDurations(valid)

	// Calculate P99 index
	p99Index := int(float64(len(valid)) * 0.99)
	if p99Index >= len(valid) {
		p99Index = len(valid) - 1
	}

	return valid[p99Index]
}

// GetBufferSaturation calculates buffer saturation as a percentage
func (mc *MetricsCollector) GetBufferSaturation() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.bufferCapacity == 0 {
		return 0
	}

	return (float64(mc.currentBuffer) / float64(mc.bufferCapacity)) * 100.0
}

// GetMessageRate calculates messages per second
func (mc *MetricsCollector) GetMessageRate() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	elapsed := time.Since(mc.lastReset).Seconds()
	if elapsed == 0 {
		return 0
	}

	return float64(mc.messageCount) / elapsed
}

// GetErrorRate calculates errors per second
func (mc *MetricsCollector) GetErrorRate() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	elapsed := time.Since(mc.lastReset).Seconds()
	if elapsed == 0 {
		return 0
	}

	return float64(mc.errorCount) / elapsed
}

// GetMetrics returns a snapshot of all kernel metrics
func (mc *MetricsCollector) GetMetrics(activeWorkers, totalPermits, allocatedPermits, availablePermits int) *KernelMetrics {
	p99 := mc.GetP99Latency()
	p99Ms := float64(p99.Microseconds()) / 1000.0 // Convert to milliseconds

	return &KernelMetrics{
		P99LatencyMs:     p99Ms,
		BufferSaturation: mc.GetBufferSaturation(),
		ActiveWorkers:    activeWorkers,
		TotalPermits:     totalPermits,
		AllocatedPermits: allocatedPermits,
		AvailablePermits: availablePermits,
		MessageRate:      mc.GetMessageRate(),
		ErrorRate:        mc.GetErrorRate(),
	}
}

// sortDurations is a simple insertion sort for time.Duration slices
// Efficient for small to medium-sized slices typical in P99 calculation
func sortDurations(arr []time.Duration) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}
