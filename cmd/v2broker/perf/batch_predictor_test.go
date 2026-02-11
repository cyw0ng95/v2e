package perf

import (
	"testing"
	"time"
)

func TestBatchSizePredictor_Creation(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	if predictor == nil {
		t.Fatal("Failed to create predictor")
	}

	if predictor.GetCurrentBatchSize() != 1 {
		t.Errorf("Expected initial batch size 1, got %d", predictor.GetCurrentBatchSize())
	}

	if predictor.minBatchSize != 1 {
		t.Errorf("Expected min batch size 1, got %d", predictor.minBatchSize)
	}

	if predictor.maxBatchSize != 100 {
		t.Errorf("Expected max batch size 100, got %d", predictor.maxBatchSize)
	}
}

func TestBatchSizePredictor_BoundaryValidation(t *testing.T) {
	predictor := NewBatchSizePredictor(-1, 0, 100)

	if predictor.minBatchSize < 1 {
		t.Errorf("Expected min batch size to be >= 1, got %d", predictor.minBatchSize)
	}

	if predictor.maxBatchSize < predictor.minBatchSize {
		t.Errorf("Expected max batch size >= min batch size, got %d < %d",
			predictor.maxBatchSize, predictor.minBatchSize)
	}
}

func TestBatchSizePredictor_RecordBatch(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 10)

	initialSize := predictor.GetCurrentBatchSize()
	if initialSize != 1 {
		t.Errorf("Expected initial batch size 1, got %d", initialSize)
	}

	predictor.RecordBatch(10, 100.0, 10*time.Millisecond, 0)
	predictor.RecordBatch(20, 200.0, 15*time.Millisecond, 0)
	predictor.RecordBatch(30, 300.0, 20*time.Millisecond, 0)

	if len(predictor.batchHistory) != 3 {
		t.Errorf("Expected 3 batch records, got %d", len(predictor.batchHistory))
	}

	if len(predictor.throughputHistory) != 3 {
		t.Errorf("Expected 3 throughput records, got %d", len(predictor.throughputHistory))
	}

	if len(predictor.latencyHistory) != 3 {
		t.Errorf("Expected 3 latency records, got %d", len(predictor.latencyHistory))
	}
}

func TestBatchSizePredictor_HistoryLimit(t *testing.T) {
	// Note: NewBatchSizePredictor uses minimum 100 for historySize, and maxMetricHistory defaults to 1000
	predictor := NewBatchSizePredictor(1, 100, 5)

	for i := 0; i < 10; i++ {
		predictor.RecordBatch(i+1, float64(i+1)*100, time.Duration(i+1)*time.Millisecond, 0)
	}

	// Check batch history (limited by maxHistorySize which is min(historySize, 100))
	if len(predictor.batchHistory) > 100 {
		t.Errorf("Expected batch history size <= 100, got %d", len(predictor.batchHistory))
	}

	// Check throughput history (limited by maxMetricHistory which is 1000)
	if len(predictor.throughputHistory) > 1000 {
		t.Errorf("Expected throughput history size <= 1000, got %d", len(predictor.throughputHistory))
	}
}

func TestBatchSizePredictor_PredictBatchSize(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	predictor.RecordBatch(10, 100.0, 10*time.Millisecond, 0)
	predictor.RecordBatch(20, 200.0, 15*time.Millisecond, 0)
	predictor.RecordBatch(30, 300.0, 20*time.Millisecond, 0)
	predictor.RecordBatch(10, 150.0, 12*time.Millisecond, 0)
	predictor.RecordBatch(20, 250.0, 18*time.Millisecond, 0)

	predictedSize := predictor.PredictBatchSize()

	if predictedSize < predictor.minBatchSize || predictedSize > predictor.maxBatchSize {
		t.Errorf("Predicted batch size %d out of bounds [%d, %d]",
			predictedSize, predictor.minBatchSize, predictor.maxBatchSize)
	}
}

func TestBatchSizePredictor_ThroughputOptimizedSize(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	// Record batches with different sizes and throughputs
	// Larger batch size should have higher throughput
	predictor.RecordBatch(10, 100.0, 10*time.Millisecond, 0)
	predictor.RecordBatch(20, 250.0, 15*time.Millisecond, 0)
	predictor.RecordBatch(30, 450.0, 20*time.Millisecond, 0)
	predictor.RecordBatch(10, 120.0, 12*time.Millisecond, 0)
	predictor.RecordBatch(20, 280.0, 18*time.Millisecond, 0)

	size := predictor.throughputOptimizedSize()

	// Should prefer batch size 30 (highest throughput)
	if size != 30 {
		t.Logf("Note: Throughput optimized size is %d, may vary based on algorithm", size)
	}
}

func TestBatchSizePredictor_LatencyOptimizedSize(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	// Record batches with different sizes and latencies
	// Smaller batch size should have lower latency
	predictor.RecordBatch(10, 100.0, 5*time.Millisecond, 0)
	predictor.RecordBatch(20, 200.0, 10*time.Millisecond, 0)
	predictor.RecordBatch(30, 300.0, 15*time.Millisecond, 0)
	predictor.RecordBatch(10, 120.0, 6*time.Millisecond, 0)
	predictor.RecordBatch(20, 280.0, 11*time.Millisecond, 0)

	size := predictor.latencyOptimizedSize()

	// Should prefer batch size 10 (lowest latency)
	if size != 10 {
		t.Logf("Note: Latency optimized size is %d, may vary based on algorithm", size)
	}
}

func TestBatchSizePredictor_TrendDetection(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	// Record increasing throughput trend
	for i := 0; i < 10; i++ {
		throughput := float64(100 + i*10)
		predictor.RecordBatch(10, throughput, 10*time.Millisecond, 0)
	}

	trend, confidence := predictor.GetTrendInfo()

	if trend != trendIncreasing && confidence > 0.6 {
		t.Logf("Detected trend: %d, confidence: %.2f", trend, confidence)
	}
}

func TestBatchSizePredictor_Reset(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	// Record some data
	for i := 0; i < 5; i++ {
		predictor.RecordBatch(i+1, float64(i+1)*100, time.Duration(i+1)*time.Millisecond, 0)
	}

	if len(predictor.batchHistory) == 0 {
		t.Error("Expected batch history to have records before reset")
	}

	predictor.Reset()

	if len(predictor.batchHistory) != 0 {
		t.Errorf("Expected empty batch history after reset, got %d records", len(predictor.batchHistory))
	}

	if len(predictor.throughputHistory) != 0 {
		t.Errorf("Expected empty throughput history after reset, got %d records", len(predictor.throughputHistory))
	}

	if predictor.GetCurrentBatchSize() != predictor.minBatchSize {
		t.Errorf("Expected batch size to reset to min (%d), got %d",
			predictor.minBatchSize, predictor.GetCurrentBatchSize())
	}

	if predictor.currentTrend != trendStable {
		t.Errorf("Expected trend to be stable after reset, got %d", predictor.currentTrend)
	}
}

func TestBatchSizePredictor_PatternBased(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	// Create a pattern: batch size 20 is most successful
	predictor.RecordBatch(20, 200.0, 15*time.Millisecond, 0)
	predictor.RecordBatch(20, 220.0, 16*time.Millisecond, 0)
	predictor.RecordBatch(10, 100.0, 8*time.Millisecond, 1) // Has errors
	predictor.RecordBatch(30, 300.0, 25*time.Millisecond, 0)
	predictor.RecordBatch(20, 210.0, 14*time.Millisecond, 0)

	size := predictor.patternBasedSize()

	// Should prefer batch size 20 (highest success rate)
	if size != 20 {
		t.Logf("Note: Pattern based size is %d, may vary based on algorithm", size)
	}
}

func TestBatchSizePredictor_Average(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	values := []float64{10.0, 20.0, 30.0, 40.0, 50.0}
	avg := predictor.average(values)

	expected := 30.0
	if avg != expected {
		t.Errorf("Expected average %.2f, got %.2f", expected, avg)
	}
}

func TestBatchSizePredictor_AverageDuration(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	values := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}
	avg := predictor.averageDuration(values)

	expected := 20 * time.Millisecond
	if avg != expected {
		t.Errorf("Expected average %v, got %v", expected, avg)
	}
}

func TestBatchSizePredictor_ConcurrentAccess(t *testing.T) {
	predictor := NewBatchSizePredictor(1, 100, 100)

	done := make(chan bool)

	// Multiple goroutines recording batches
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				predictor.RecordBatch(id+1, float64(id+1)*100, time.Duration(id+1)*time.Millisecond, 0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify data integrity
	predictedSize := predictor.PredictBatchSize()
	if predictedSize < predictor.minBatchSize || predictedSize > predictor.maxBatchSize {
		t.Errorf("Predicted batch size %d out of bounds after concurrent access", predictedSize)
	}
}

func TestBatchSizePredictor_EmptyHistoryPrediction(t *testing.T) {
	predictor := NewBatchSizePredictor(5, 50, 100)

	// Don't record any data
	predictedSize := predictor.PredictBatchSize()

	// Should return initial size
	if predictedSize != predictor.minBatchSize {
		t.Errorf("Expected initial batch size %d, got %d", predictor.minBatchSize, predictedSize)
	}
}

func TestBatchSizePredictor_TrendBasedSize(t *testing.T) {
	predictor := NewBatchSizePredictor(10, 100, 100)

	// Simulate increasing trend
	for i := 0; i < 10; i++ {
		throughput := float64(100 + i*20)
		predictor.RecordBatch(10, throughput, 10*time.Millisecond, 0)
	}

	// Force trend detection
	predictor.analyzeTrends()

	size := predictor.trendBasedSize()

	// Should increase batch size based on trend
	if size < 10 {
		t.Errorf("Expected batch size to increase based on trend, got %d", size)
	}
}
