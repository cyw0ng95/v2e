package proc

import (
	"sync"
	"time"
)

// MessagePattern describes a pattern of message characteristics
type MessagePattern struct {
	AvgMessageSize   int64
	AvgBatchSize     int
	MessageFrequency float64 // messages per second
	Timestamp        time.Time
}

// CapacityPredictor uses historical data to predict optimal capacity
type CapacityPredictor struct {
	mu               sync.RWMutex
	patternHistory   []MessagePattern
	maxHistorySize   int
	predictionWindow time.Duration
	lastPrediction   *CapacityPrediction
	learningRate     float64
}

// CapacityPrediction represents a capacity prediction
type CapacityPrediction struct {
	OptimalBufferSize int
	OptimalBatchSize  int
	ExpectedLoad      float64
	Confidence        float64
	Timestamp         time.Time
}

// NewCapacityPredictor creates a new capacity predictor
func NewCapacityPredictor() *CapacityPredictor {
	return &CapacityPredictor{
		patternHistory:   make([]MessagePattern, 0, 100),
		maxHistorySize:   100,
		predictionWindow: 5 * time.Minute,
		learningRate:     0.1,
	}
}

// RecordMessage records message statistics for learning
func (cp *CapacityPredictor) RecordMessage(size int64, batchSize int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	now := time.Now()

	// Calculate frequency (simplified - moving average)
	frequency := 1.0
	if len(cp.patternHistory) > 0 {
		lastPattern := cp.patternHistory[len(cp.patternHistory)-1]
		timeDiff := now.Sub(lastPattern.Timestamp).Seconds()
		if timeDiff > 0 {
			frequency = 1.0 / timeDiff
		}
	}

	pattern := MessagePattern{
		AvgMessageSize:   size,
		AvgBatchSize:     batchSize,
		MessageFrequency: frequency,
		Timestamp:        now,
	}

	cp.patternHistory = append(cp.patternHistory, pattern)

	// Limit history size
	if len(cp.patternHistory) > cp.maxHistorySize {
		cp.patternHistory = cp.patternHistory[1:]
	}
}

// Predict predicts optimal capacity based on historical patterns
func (cp *CapacityPredictor) Predict() *CapacityPrediction {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if len(cp.patternHistory) == 0 {
		return cp.defaultPrediction()
	}

	now := time.Now()

	// Filter recent patterns
	var recentPatterns []MessagePattern
	for _, pattern := range cp.patternHistory {
		if now.Sub(pattern.Timestamp) <= cp.predictionWindow {
			recentPatterns = append(recentPatterns, pattern)
		}
	}

	if len(recentPatterns) == 0 {
		recentPatterns = cp.patternHistory
	}

	// Calculate averages
	var totalSize, totalBatch, totalFreq int64
	for _, pattern := range recentPatterns {
		totalSize += pattern.AvgMessageSize
		totalBatch += int64(pattern.AvgBatchSize)
		totalFreq += int64(pattern.MessageFrequency * 1000) // Convert to integer
	}

	avgSize := totalSize / int64(len(recentPatterns))
	avgBatch := int(totalBatch / int64(len(recentPatterns)))
	avgFreq := float64(totalFreq) / float64(len(recentPatterns)) / 1000.0

	// Predict optimal buffer size (with safety margin)
	optimalBufferSize := int(avgSize*2) + 1024 // Safety margin + overhead
	if optimalBufferSize > MaxMessageSize {
		optimalBufferSize = MaxMessageSize
	}

	// Predict optimal batch size
	optimalBatchSize := avgBatch
	if optimalBatchSize < 10 {
		optimalBatchSize = 10
	} else if optimalBatchSize > 1000 {
		optimalBatchSize = 1000
	}

	// Calculate confidence based on data variability
	confidence := cp.calculateConfidence(recentPatterns)

	// Smooth prediction with learning
	if cp.lastPrediction != nil {
		optimalBufferSize = int(float64(optimalBufferSize)*(1.0-cp.learningRate) +
			float64(cp.lastPrediction.OptimalBufferSize)*cp.learningRate)
		optimalBatchSize = int(float64(optimalBatchSize)*(1.0-cp.learningRate) +
			float64(cp.lastPrediction.OptimalBatchSize)*cp.learningRate)
	}

	prediction := &CapacityPrediction{
		OptimalBufferSize: optimalBufferSize,
		OptimalBatchSize:  optimalBatchSize,
		ExpectedLoad:      avgFreq,
		Confidence:        confidence,
		Timestamp:         now,
	}

	cp.lastPrediction = prediction
	return prediction
}

// calculateConfidence calculates prediction confidence based on data stability
func (cp *CapacityPredictor) calculateConfidence(patterns []MessagePattern) float64 {
	if len(patterns) < 2 {
		return 0.5
	}

	// Calculate variance
	var sizeVariance, batchVariance float64
	avgSize := 0.0
	avgBatch := 0.0

	for _, p := range patterns {
		avgSize += float64(p.AvgMessageSize)
		avgBatch += float64(p.AvgBatchSize)
	}
	avgSize /= float64(len(patterns))
	avgBatch /= float64(len(patterns))

	for _, p := range patterns {
		sizeVariance += (float64(p.AvgMessageSize) - avgSize) * (float64(p.AvgMessageSize) - avgSize)
		batchVariance += (float64(p.AvgBatchSize) - avgBatch) * (float64(p.AvgBatchSize) - avgBatch)
	}
	sizeVariance /= float64(len(patterns))
	batchVariance /= float64(len(patterns))

	// Lower variance = higher confidence
	stdDevSize := sizeVariance / avgSize
	stdDevBatch := batchVariance / avgBatch

	// Normalize confidence to 0-1 range
	confidence := 1.0 - (stdDevSize+stdDevBatch)/2.0
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// defaultPrediction returns default capacity prediction
func (cp *CapacityPredictor) defaultPrediction() *CapacityPrediction {
	return &CapacityPrediction{
		OptimalBufferSize: 32768, // 32KB default
		OptimalBatchSize:  100,
		ExpectedLoad:      100.0,
		Confidence:        0.5,
		Timestamp:         time.Now(),
	}
}

// GetLastPrediction returns the most recent prediction
func (cp *CapacityPredictor) GetLastPrediction() *CapacityPrediction {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.lastPrediction
}

// Reset clears prediction history
func (cp *CapacityPredictor) Reset() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.patternHistory = nil
	cp.lastPrediction = nil
}
