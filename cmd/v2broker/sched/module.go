// Package sched implements scheduler and performance optimization for the broker
package sched

import (
	"math"
	"time"
)

// Scheduler defines the interface for scheduling and optimizing operations
type Scheduler interface {
	Start()
	Stop()
	UpdateLoad(load LoadMetrics)
	GetLoadMetrics() LoadMetrics
	AdjustParameters()
}

// LoadMetrics contains system and application load metrics
type LoadMetrics struct {
	CPUUtilization       float64
	MemoryUtilization    float64
	MessageQueueDepth    int64
	MessageThroughput    float64
	AverageLatency       time.Duration
	ActiveConnections    int
	SystemLoadAvg        float64
	ProcessResourceUsage map[string]ProcessResourceMetrics
}

// ProcessResourceMetrics contains resource metrics for a specific process
type ProcessResourceMetrics struct {
	CPUUtilization      float64
	MemoryUsage         uint64
	MessageRate         float64
	AverageResponseTime time.Duration
}

// Optimizer defines the interface for performance optimization
type Optimizer interface {
	Initialize()
	Observe(metrics LoadMetrics) error
	AdjustConfiguration() error
	GetMetrics() map[string]interface{}
}

// AdaptiveOptimizer implements adaptive optimization based on system load
type AdaptiveOptimizer struct {
	currentMetrics LoadMetrics
	parameters     OptimizationParameters
	
	// Historical metrics for trend analysis
	metricsHistory []LoadMetrics
	maxHistoryLen  int
	
	// Performance tracking for gradient-based optimization
	lastPerformanceScore float64
	lastAdjustmentTime   time.Time
	adjustmentCooldown   time.Duration
	
	// Adaptive thresholds based on historical data
	adaptiveThresholds adaptiveThresholds
	
	// Moving averages for predictive tuning
	cpuMA         *movingAverage
	throughputMA  *movingAverage
	latencyMA     *movingAverage
	queueDepthMA  *movingAverage
}

// adaptiveThresholds holds dynamic thresholds that adapt to workload patterns
type adaptiveThresholds struct {
	cpuHighLoad       float64
	cpuLowLoad        float64
	throughputHigh    float64
	throughputMedium  float64
	throughputLow     float64
	latencyHigh       time.Duration
	latencyMedium     time.Duration
	queueDepthHigh    int64
}

// movingAverage implements simple moving average for metric prediction
type movingAverage struct {
	samples []float64
	window  int
	index   int
	sum     float64
	full    bool
}

// OptimizationParameters holds adjustable parameters for optimization
type OptimizationParameters struct {
	BufferCapacity int
	WorkerCount    int
	BatchSize      int
	FlushInterval  time.Duration
	OfferPolicy    string
	OfferTimeout   time.Duration
}

// NewAdaptiveOptimizer creates a new adaptive optimizer
func NewAdaptiveOptimizer() *AdaptiveOptimizer {
	return &AdaptiveOptimizer{
		parameters: OptimizationParameters{
			BufferCapacity: 1000,
			WorkerCount:    4,
			BatchSize:      1,
			FlushInterval:  10 * time.Millisecond,
			OfferPolicy:    "drop",
			OfferTimeout:   100 * time.Millisecond,
		},
		metricsHistory:     make([]LoadMetrics, 0, 100),
		maxHistoryLen:      100,
		adjustmentCooldown: 5 * time.Second,
		lastAdjustmentTime: time.Now(),
		adaptiveThresholds: adaptiveThresholds{
			cpuHighLoad:      0.9,
			cpuLowLoad:       0.7,
			throughputHigh:   1500,
			throughputMedium: 800,
			throughputLow:    300,
			latencyHigh:      20 * time.Millisecond,
			latencyMedium:    10 * time.Millisecond,
			queueDepthHigh:   500,
		},
		cpuMA:        newMovingAverage(20),
		throughputMA: newMovingAverage(20),
		latencyMA:    newMovingAverage(20),
		queueDepthMA: newMovingAverage(20),
	}
}

// newMovingAverage creates a new moving average calculator
func newMovingAverage(window int) *movingAverage {
	return &movingAverage{
		samples: make([]float64, window),
		window:  window,
	}
}

// add adds a new sample to the moving average
func (ma *movingAverage) add(value float64) {
	if !ma.full {
		ma.samples[ma.index] = value
		ma.sum += value
		ma.index++
		if ma.index >= ma.window {
			ma.full = true
			ma.index = 0
		}
	} else {
		ma.sum -= ma.samples[ma.index]
		ma.samples[ma.index] = value
		ma.sum += value
		ma.index = (ma.index + 1) % ma.window
	}
}

// average returns the current moving average
func (ma *movingAverage) average() float64 {
	if !ma.full {
		if ma.index == 0 {
			return 0
		}
		return ma.sum / float64(ma.index)
	}
	return ma.sum / float64(ma.window)
}

// trend returns the trend direction (positive = increasing, negative = decreasing)
func (ma *movingAverage) trend() float64 {
	if !ma.full || ma.window < 2 {
		return 0
	}
	
	// Simple linear regression slope
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(ma.window)
	
	for i := 0; i < ma.window; i++ {
		x := float64(i)
		y := ma.samples[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX)
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}
	
	slope := (n*sumXY - sumX*sumY) / denominator
	return slope
}

// Initialize initializes the optimizer
func (ao *AdaptiveOptimizer) Initialize() {
	// Initialization logic here
}

// Observe records new metrics and adjusts parameters accordingly
func (ao *AdaptiveOptimizer) Observe(metrics LoadMetrics) error {
	ao.currentMetrics = metrics
	
	// Update moving averages for trend analysis
	ao.cpuMA.add(metrics.CPUUtilization)
	ao.throughputMA.add(metrics.MessageThroughput)
	ao.latencyMA.add(float64(metrics.AverageLatency.Milliseconds()))
	ao.queueDepthMA.add(float64(metrics.MessageQueueDepth))
	
	// Store metrics history
	ao.metricsHistory = append(ao.metricsHistory, metrics)
	if len(ao.metricsHistory) > ao.maxHistoryLen {
		ao.metricsHistory = ao.metricsHistory[1:]
	}
	
	// Update adaptive thresholds based on historical data
	ao.updateAdaptiveThresholds()
	
	// Only adjust if cooldown period has passed
	if time.Since(ao.lastAdjustmentTime) >= ao.adjustmentCooldown {
		ao.AdjustConfiguration()
		ao.lastAdjustmentTime = time.Now()
	}
	
	return nil
}

// updateAdaptiveThresholds dynamically adjusts thresholds based on workload patterns
func (ao *AdaptiveOptimizer) updateAdaptiveThresholds() {
	if len(ao.metricsHistory) < 10 {
		return // Need sufficient history
	}
	
	// Calculate percentiles for adaptive thresholds
	var cpuValues, throughputValues, latencies, queueDepths []float64
	for _, m := range ao.metricsHistory {
		cpuValues = append(cpuValues, m.CPUUtilization)
		throughputValues = append(throughputValues, m.MessageThroughput)
		latencies = append(latencies, float64(m.AverageLatency.Milliseconds()))
		queueDepths = append(queueDepths, float64(m.MessageQueueDepth))
	}
	
	// Update CPU thresholds (80th and 60th percentiles)
	ao.adaptiveThresholds.cpuHighLoad = percentile(cpuValues, 0.8)
	ao.adaptiveThresholds.cpuLowLoad = percentile(cpuValues, 0.6)
	
	// Update throughput thresholds (75th, 50th, 25th percentiles)
	ao.adaptiveThresholds.throughputHigh = percentile(throughputValues, 0.75)
	ao.adaptiveThresholds.throughputMedium = percentile(throughputValues, 0.50)
	ao.adaptiveThresholds.throughputLow = percentile(throughputValues, 0.25)
	
	// Update latency thresholds (75th and 50th percentiles)
	latency75 := percentile(latencies, 0.75)
	latency50 := percentile(latencies, 0.50)
	ao.adaptiveThresholds.latencyHigh = time.Duration(latency75) * time.Millisecond
	ao.adaptiveThresholds.latencyMedium = time.Duration(latency50) * time.Millisecond
	
	// Update queue depth threshold (75th percentile)
	ao.adaptiveThresholds.queueDepthHigh = int64(percentile(queueDepths, 0.75))
}

// percentile calculates the p-th percentile of values (p in [0, 1])
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Simple percentile implementation (not sorting to avoid modifying original)
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	// Bubble sort (simple for small datasets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}

// AdjustConfiguration adjusts optimization parameters based on current metrics
func (ao *AdaptiveOptimizer) AdjustConfiguration() error {
	// Get trend information from moving averages
	cpuTrend := ao.cpuMA.trend()
	throughputTrend := ao.throughputMA.trend()
	latencyTrend := ao.latencyMA.trend()
	queueDepthTrend := ao.queueDepthMA.trend()
	
	// Calculate performance score (higher is better)
	performanceScore := ao.calculatePerformanceScore()
	
	// Store previous parameters for gradient descent
	prevWorkerCount := ao.parameters.WorkerCount
	prevBufferCapacity := ao.parameters.BufferCapacity
	prevBatchSize := ao.parameters.BatchSize
	
	// === INTELLIGENT WORKER COUNT ADJUSTMENT ===
	ao.adjustWorkerCount(cpuTrend, queueDepthTrend)
	
	// === INTELLIGENT BUFFER CAPACITY ADJUSTMENT ===
	ao.adjustBufferCapacity(throughputTrend, queueDepthTrend)
	
	// === INTELLIGENT BATCH SIZE ADJUSTMENT ===
	ao.adjustBatchSize(throughputTrend, latencyTrend)
	
	// === INTELLIGENT FLUSH INTERVAL ADJUSTMENT ===
	ao.adjustFlushInterval(latencyTrend)
	
	// === INTELLIGENT OFFER POLICY ADJUSTMENT ===
	ao.adjustOfferPolicy()
	
	// Gradient-based optimization: if performance degraded, revert partially
	if performanceScore < ao.lastPerformanceScore*0.95 {
		// Performance dropped significantly, revert halfway to previous values
		ao.parameters.WorkerCount = (ao.parameters.WorkerCount + prevWorkerCount) / 2
		ao.parameters.BufferCapacity = (ao.parameters.BufferCapacity + prevBufferCapacity) / 2
		ao.parameters.BatchSize = (ao.parameters.BatchSize + prevBatchSize) / 2
	}
	
	ao.lastPerformanceScore = performanceScore
	
	return nil
}

// calculatePerformanceScore computes a composite performance score
func (ao *AdaptiveOptimizer) calculatePerformanceScore() float64 {
	// Normalize metrics to [0, 1] range
	throughputScore := math.Min(ao.currentMetrics.MessageThroughput/2000.0, 1.0)
	
	latencyScore := 1.0
	if ao.currentMetrics.AverageLatency > 0 {
		latencyScore = 1.0 / (1.0 + float64(ao.currentMetrics.AverageLatency.Milliseconds())/10.0)
	}
	
	queueScore := 1.0
	if ao.currentMetrics.MessageQueueDepth > 0 {
		queueScore = 1.0 / (1.0 + float64(ao.currentMetrics.MessageQueueDepth)/100.0)
	}
	
	cpuScore := 1.0 - math.Abs(ao.currentMetrics.CPUUtilization-0.75)/0.75
	if cpuScore < 0 {
		cpuScore = 0
	}
	
	// Weighted combination (throughput and latency are most important)
	return 0.4*throughputScore + 0.3*latencyScore + 0.2*queueScore + 0.1*cpuScore
}

// adjustWorkerCount intelligently adjusts worker count based on trends
func (ao *AdaptiveOptimizer) adjustWorkerCount(cpuTrend, queueDepthTrend float64) {
	currentCPU := ao.currentMetrics.CPUUtilization
	currentQueueDepth := ao.currentMetrics.MessageQueueDepth
	
	// Predictive adjustment based on trends
	if cpuTrend < -0.01 && queueDepthTrend > 0.5 {
		// CPU usage decreasing but queue growing - add workers
		ao.parameters.WorkerCount += 2
	} else if cpuTrend > 0.02 && currentCPU > ao.adaptiveThresholds.cpuHighLoad {
		// CPU usage increasing and already high - reduce workers
		if ao.parameters.WorkerCount > 2 {
			ao.parameters.WorkerCount -= 1
		}
	} else if currentCPU < ao.adaptiveThresholds.cpuLowLoad && currentQueueDepth > ao.adaptiveThresholds.queueDepthHigh {
		// Low CPU but high queue depth - add workers
		ao.parameters.WorkerCount += 1
	} else if currentCPU > ao.adaptiveThresholds.cpuHighLoad {
		// High CPU - reduce workers if possible
		if ao.parameters.WorkerCount > 2 {
			ao.parameters.WorkerCount -= 1
		}
	}
	
	// Apply constraints with adaptive upper bound
	maxWorkers := 32
	if ao.currentMetrics.CPUUtilization > 0.95 {
		maxWorkers = ao.parameters.WorkerCount // Don't increase if already overloaded
	}
	
	if ao.parameters.WorkerCount > maxWorkers {
		ao.parameters.WorkerCount = maxWorkers
	} else if ao.parameters.WorkerCount < 1 {
		ao.parameters.WorkerCount = 1
	}
}

// adjustBufferCapacity intelligently adjusts buffer capacity based on trends
func (ao *AdaptiveOptimizer) adjustBufferCapacity(throughputTrend, queueDepthTrend float64) {
	currentThroughput := ao.currentMetrics.MessageThroughput
	avgThroughput := ao.throughputMA.average()
	
	// Use adaptive thresholds instead of fixed values
	if currentThroughput > ao.adaptiveThresholds.throughputHigh || throughputTrend > 10.0 {
		// High or increasing throughput - increase buffer
		ao.parameters.BufferCapacity = int(avgThroughput * 1.5)
	} else if currentThroughput > ao.adaptiveThresholds.throughputMedium {
		// Medium throughput - moderate buffer
		ao.parameters.BufferCapacity = int(avgThroughput * 1.2)
	} else if currentThroughput > ao.adaptiveThresholds.throughputLow {
		// Low throughput - smaller buffer
		ao.parameters.BufferCapacity = int(avgThroughput * 1.0)
	} else {
		// Very low throughput - minimal buffer
		ao.parameters.BufferCapacity = 500
	}
	
	// Adjust for queue depth trends
	if queueDepthTrend > 1.0 {
		// Queue growing - increase buffer capacity
		ao.parameters.BufferCapacity = int(float64(ao.parameters.BufferCapacity) * 1.2)
	}
	
	// Apply constraints
	if ao.parameters.BufferCapacity > 5000 {
		ao.parameters.BufferCapacity = 5000
	} else if ao.parameters.BufferCapacity < 500 {
		ao.parameters.BufferCapacity = 500
	}
}

// adjustBatchSize intelligently adjusts batch size based on trends
func (ao *AdaptiveOptimizer) adjustBatchSize(throughputTrend, latencyTrend float64) {
	currentThroughput := ao.currentMetrics.MessageThroughput
	currentLatency := ao.currentMetrics.AverageLatency
	
	// Balance throughput vs latency
	if currentThroughput > ao.adaptiveThresholds.throughputHigh && currentLatency < ao.adaptiveThresholds.latencyMedium {
		// High throughput with low latency - increase batching
		ao.parameters.BatchSize = 20
	} else if currentThroughput > ao.adaptiveThresholds.throughputMedium && currentLatency < ao.adaptiveThresholds.latencyHigh {
		// Moderate throughput - moderate batching
		ao.parameters.BatchSize = 10
	} else if latencyTrend > 0.5 {
		// Latency increasing - reduce batch size
		if ao.parameters.BatchSize > 1 {
			ao.parameters.BatchSize = ao.parameters.BatchSize / 2
		}
	} else if throughputTrend < -5.0 && ao.parameters.BatchSize > 1 {
		// Throughput decreasing - try smaller batches
		ao.parameters.BatchSize = ao.parameters.BatchSize / 2
	} else {
		// Default - small batches for responsiveness
		ao.parameters.BatchSize = 5
	}
	
	// Apply constraints
	if ao.parameters.BatchSize > 50 {
		ao.parameters.BatchSize = 50
	} else if ao.parameters.BatchSize < 1 {
		ao.parameters.BatchSize = 1
	}
}

// adjustFlushInterval intelligently adjusts flush interval based on trends
func (ao *AdaptiveOptimizer) adjustFlushInterval(latencyTrend float64) {
	currentLatency := ao.currentMetrics.AverageLatency
	
	if currentLatency > ao.adaptiveThresholds.latencyHigh || latencyTrend > 1.0 {
		// High or increasing latency - flush more frequently
		ao.parameters.FlushInterval = 2 * time.Millisecond
	} else if currentLatency > ao.adaptiveThresholds.latencyMedium {
		// Medium latency - moderate flush interval
		ao.parameters.FlushInterval = 5 * time.Millisecond
	} else {
		// Low latency - can batch more aggressively
		ao.parameters.FlushInterval = 10 * time.Millisecond
	}
}

// adjustOfferPolicy intelligently adjusts offer policy based on system state
func (ao *AdaptiveOptimizer) adjustOfferPolicy() {
	currentCPU := ao.currentMetrics.CPUUtilization
	currentMem := ao.currentMetrics.MemoryUtilization
	currentQueueDepth := ao.currentMetrics.MessageQueueDepth
	
	if currentCPU > 0.95 || currentMem > 90 {
		// Under high load - use blocking policy to slow down ingestion
		ao.parameters.OfferPolicy = "block"
		ao.parameters.OfferTimeout = 10 * time.Millisecond
	} else if currentQueueDepth > ao.adaptiveThresholds.queueDepthHigh {
		// High queue depth - use drop policy to prevent overload
		ao.parameters.OfferPolicy = "drop_oldest"
		ao.parameters.OfferTimeout = 5 * time.Millisecond
	} else if currentCPU > ao.adaptiveThresholds.cpuHighLoad {
		// Moderately high CPU - use timeout policy
		ao.parameters.OfferPolicy = "timeout"
		ao.parameters.OfferTimeout = 50 * time.Millisecond
	} else {
		// Normal conditions - use timeout policy with longer timeout
		ao.parameters.OfferPolicy = "timeout"
		ao.parameters.OfferTimeout = 100 * time.Millisecond
	}
}

// GetMetrics returns current optimization metrics
func (ao *AdaptiveOptimizer) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"buffer_capacity":         ao.parameters.BufferCapacity,
		"worker_count":            ao.parameters.WorkerCount,
		"batch_size":              ao.parameters.BatchSize,
		"flush_interval":          ao.parameters.FlushInterval,
		"offer_policy":            ao.parameters.OfferPolicy,
		"offer_timeout":           ao.parameters.OfferTimeout,
		"cpu_utilization":         ao.currentMetrics.CPUUtilization,
		"queue_depth":             ao.currentMetrics.MessageQueueDepth,
		"throughput":              ao.currentMetrics.MessageThroughput,
		"performance_score":       ao.lastPerformanceScore,
		"cpu_trend":               ao.cpuMA.trend(),
		"throughput_trend":        ao.throughputMA.trend(),
		"latency_trend":           ao.latencyMA.trend(),
		"queue_depth_trend":       ao.queueDepthMA.trend(),
		"cpu_avg":                 ao.cpuMA.average(),
		"throughput_avg":          ao.throughputMA.average(),
		"adaptive_cpu_high":       ao.adaptiveThresholds.cpuHighLoad,
		"adaptive_cpu_low":        ao.adaptiveThresholds.cpuLowLoad,
		"adaptive_throughput_high": ao.adaptiveThresholds.throughputHigh,
	}
}
