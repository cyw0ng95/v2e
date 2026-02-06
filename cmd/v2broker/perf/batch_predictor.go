package perf

import (
	"math"
	"sync"
	"time"
)

// BatchSizePredictor predicts optimal batch sizes based on historical data patterns
type BatchSizePredictor struct {
	mu sync.RWMutex

	// Historical data for pattern recognition
	batchHistory   []batchRecord
	maxHistorySize int

	// Pattern recognition parameters
	patternWindow int
	lookbackCount int

	// Performance tracking
	throughputHistory []throughputRecord
	latencyHistory    []latencyRecord
	maxMetricHistory  int

	// Adaptive parameters
	minBatchSize     int
	maxBatchSize     int
	currentBatchSize int

	// Trend detection
	currentTrend    trendDirection
	trendConfidence float64
}

type batchRecord struct {
	batchSize  int
	timestamp  time.Time
	throughput float64
	latency    time.Duration
	errorCount int
}

type throughputRecord struct {
	throughput float64
	timestamp  time.Time
	batchSize  int
}

type latencyRecord struct {
	latency   time.Duration
	timestamp time.Time
	batchSize int
}

type trendDirection int

const (
	trendStable trendDirection = iota
	trendIncreasing
	trendDecreasing
)

// NewBatchSizePredictor creates a new batch size predictor
func NewBatchSizePredictor(minBatch, maxBatch, historySize int) *BatchSizePredictor {
	if minBatch < 1 {
		minBatch = 1
	}
	if maxBatch < minBatch {
		maxBatch = minBatch * 10
	}
	if historySize < 100 {
		historySize = 100
	}

	return &BatchSizePredictor{
		batchHistory:      make([]batchRecord, 0, historySize),
		maxHistorySize:    historySize,
		patternWindow:     10,
		lookbackCount:     5,
		throughputHistory: make([]throughputRecord, 0, historySize),
		latencyHistory:    make([]latencyRecord, 0, historySize),
		maxMetricHistory:  historySize,
		minBatchSize:      minBatch,
		maxBatchSize:      maxBatch,
		currentBatchSize:  minBatch,
		currentTrend:      trendStable,
		trendConfidence:   0.0,
	}
}

// RecordBatch records a batch operation with its performance metrics
func (bsp *BatchSizePredictor) RecordBatch(batchSize int, throughput float64, latency time.Duration, errors int) {
	bsp.mu.Lock()
	defer bsp.mu.Unlock()

	now := time.Now()

	record := batchRecord{
		batchSize:  batchSize,
		timestamp:  now,
		throughput: throughput,
		latency:    latency,
		errorCount: errors,
	}

	bsp.batchHistory = append(bsp.batchHistory, record)
	if len(bsp.batchHistory) > bsp.maxHistorySize {
		bsp.batchHistory = bsp.batchHistory[1:]
	}

	bsp.throughputHistory = append(bsp.throughputHistory, throughputRecord{
		throughput: throughput,
		timestamp:  now,
		batchSize:  batchSize,
	})
	if len(bsp.throughputHistory) > bsp.maxMetricHistory {
		bsp.throughputHistory = bsp.throughputHistory[1:]
	}

	bsp.latencyHistory = append(bsp.latencyHistory, latencyRecord{
		latency:   latency,
		timestamp: now,
		batchSize: batchSize,
	})
	if len(bsp.latencyHistory) > bsp.maxMetricHistory {
		bsp.latencyHistory = bsp.latencyHistory[1:]
	}

	bsp.analyzeTrends()
}

// analyzeTrends analyzes historical data to detect performance trends
func (bsp *BatchSizePredictor) analyzeTrends() {
	if len(bsp.throughputHistory) < bsp.patternWindow {
		return
	}

	window := bsp.throughputHistory[len(bsp.throughputHistory)-bsp.patternWindow:]

	increasingCount := 0
	decreasingCount := 0
	totalChange := 0.0

	for i := 1; i < len(window); i++ {
		change := window[i].throughput - window[i-1].throughput
		totalChange += change
		if change > 0 {
			increasingCount++
		} else if change < 0 {
			decreasingCount++
		}
	}

	bsp.currentTrend = trendStable
	bsp.trendConfidence = 0.0

	if increasingCount > decreasingCount {
		bsp.currentTrend = trendIncreasing
		bsp.trendConfidence = float64(increasingCount) / float64(len(window)-1)
	} else if decreasingCount > increasingCount {
		bsp.currentTrend = trendDecreasing
		bsp.trendConfidence = float64(decreasingCount) / float64(len(window)-1)
	}
}

// PredictBatchSize predicts the optimal batch size based on historical patterns
func (bsp *BatchSizePredictor) PredictBatchSize() int {
	bsp.mu.RLock()
	defer bsp.mu.RUnlock()

	if len(bsp.batchHistory) < bsp.lookbackCount {
		return bsp.currentBatchSize
	}

	predictedSize := bsp.calculateOptimalBatchSize()

	if predictedSize < bsp.minBatchSize {
		predictedSize = bsp.minBatchSize
	} else if predictedSize > bsp.maxBatchSize {
		predictedSize = bsp.maxBatchSize
	}

	bsp.currentBatchSize = predictedSize
	return predictedSize
}

// calculateOptimalBatchSize calculates the optimal batch size using multiple strategies
func (bsp *BatchSizePredictor) calculateOptimalBatchSize() int {
	strategies := []struct {
		name   string
		weight float64
		size   int
	}{
		{"throughput_optimized", 0.4, bsp.throughputOptimizedSize()},
		{"latency_optimized", 0.3, bsp.latencyOptimizedSize()},
		{"pattern_based", 0.2, bsp.patternBasedSize()},
		{"trend_based", 0.1, bsp.trendBasedSize()},
	}

	weightedSum := 0.0
	totalWeight := 0.0

	for _, strategy := range strategies {
		if strategy.size > 0 {
			weightedSum += float64(strategy.size) * strategy.weight
			totalWeight += strategy.weight
		}
	}

	if totalWeight == 0 {
		return bsp.currentBatchSize
	}

	return int(weightedSum / totalWeight)
}

// throughputOptimizedSize finds batch size that maximizes throughput
func (bsp *BatchSizePredictor) throughputOptimizedSize() int {
	if len(bsp.throughputHistory) < bsp.lookbackCount {
		return bsp.currentBatchSize
	}

	bestSize := bsp.currentBatchSize
	bestThroughput := 0.0

	throughputBySize := make(map[int][]float64)
	for _, record := range bsp.throughputHistory {
		throughputBySize[record.batchSize] = append(throughputBySize[record.batchSize], record.throughput)
	}

	for size, throughputs := range throughputBySize {
		avgThroughput := bsp.average(throughputs)
		if avgThroughput > bestThroughput {
			bestThroughput = avgThroughput
			bestSize = size
		}
	}

	return bestSize
}

// latencyOptimizedSize finds batch size that minimizes latency
func (bsp *BatchSizePredictor) latencyOptimizedSize() int {
	if len(bsp.latencyHistory) < bsp.lookbackCount {
		return bsp.currentBatchSize
	}

	bestSize := bsp.currentBatchSize
	bestLatency := time.Duration(math.MaxInt64)

	latencyBySize := make(map[int][]time.Duration)
	for _, record := range bsp.latencyHistory {
		latencyBySize[record.batchSize] = append(latencyBySize[record.batchSize], record.latency)
	}

	for size, latencies := range latencyBySize {
		avgLatency := bsp.averageDuration(latencies)
		if avgLatency < bestLatency {
			bestLatency = avgLatency
			bestSize = size
		}
	}

	return bestSize
}

// patternBasedSize uses pattern recognition to predict optimal batch size
func (bsp *BatchSizePredictor) patternBasedSize() int {
	if len(bsp.batchHistory) < bsp.patternWindow {
		return bsp.currentBatchSize
	}

	window := bsp.batchHistory[len(bsp.batchHistory)-bsp.patternWindow:]

	successCount := make(map[int]int)
	totalCount := make(map[int]int)

	for _, record := range window {
		totalCount[record.batchSize]++
		if record.errorCount == 0 && record.throughput > 0 {
			successCount[record.batchSize]++
		}
	}

	bestSize := bsp.currentBatchSize
	bestSuccessRate := 0.0

	for size, total := range totalCount {
		if total > 0 {
			successRate := float64(successCount[size]) / float64(total)
			if successRate > bestSuccessRate {
				bestSuccessRate = successRate
				bestSize = size
			}
		}
	}

	return bestSize
}

// trendBasedSize adjusts batch size based on detected trend
func (bsp *BatchSizePredictor) trendBasedSize() int {
	if bsp.trendConfidence < 0.6 {
		return bsp.currentBatchSize
	}

	adjustmentFactor := 1.1
	if bsp.currentTrend == trendDecreasing {
		adjustmentFactor = 0.9
	}

	newSize := int(float64(bsp.currentBatchSize) * adjustmentFactor)

	if newSize < bsp.minBatchSize {
		return bsp.minBatchSize
	}
	if newSize > bsp.maxBatchSize {
		return bsp.maxBatchSize
	}

	return newSize
}

// average calculates the average of a slice of float64 values
func (bsp *BatchSizePredictor) average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// averageDuration calculates the average of a slice of time.Duration values
func (bsp *BatchSizePredictor) averageDuration(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}
	var sum time.Duration
	for _, v := range values {
		sum += v
	}
	return sum / time.Duration(len(values))
}

// GetCurrentBatchSize returns the current predicted batch size
func (bsp *BatchSizePredictor) GetCurrentBatchSize() int {
	bsp.mu.RLock()
	defer bsp.mu.RUnlock()
	return bsp.currentBatchSize
}

// GetTrendInfo returns information about current trend
func (bsp *BatchSizePredictor) GetTrendInfo() (trendDirection, float64) {
	bsp.mu.RLock()
	defer bsp.mu.RUnlock()
	return bsp.currentTrend, bsp.trendConfidence
}

// Reset clears all historical data
func (bsp *BatchSizePredictor) Reset() {
	bsp.mu.Lock()
	defer bsp.mu.Unlock()

	bsp.batchHistory = make([]batchRecord, 0, bsp.maxHistorySize)
	bsp.throughputHistory = make([]throughputRecord, 0, bsp.maxMetricHistory)
	bsp.latencyHistory = make([]latencyRecord, 0, bsp.maxMetricHistory)
	bsp.currentBatchSize = bsp.minBatchSize
	bsp.currentTrend = trendStable
	bsp.trendConfidence = 0.0
}
