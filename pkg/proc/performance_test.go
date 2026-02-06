package proc

import (
	"testing"
	"time"
)

func TestCapacityPredictor_RecordMessage(t *testing.T) {
	cp := NewCapacityPredictor()

	// Record some messages
	cp.RecordMessage(1024, 10)
	cp.RecordMessage(2048, 20)
	cp.RecordMessage(4096, 30)

	if len(cp.patternHistory) != 3 {
		t.Errorf("Expected 3 patterns, got %d", len(cp.patternHistory))
	}
}

func TestCapacityPredictor_Predict(t *testing.T) {
	cp := NewCapacityPredictor()

	// Record enough data to make a prediction
	for i := 0; i < 10; i++ {
		cp.RecordMessage(1024, 100)
	}

	prediction := cp.Predict()

	if prediction.OptimalBufferSize <= 0 {
		t.Errorf("Expected positive buffer size, got %d", prediction.OptimalBufferSize)
	}
	if prediction.OptimalBatchSize <= 0 {
		t.Errorf("Expected positive batch size, got %d", prediction.OptimalBatchSize)
	}
	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", prediction.Confidence)
	}
}

func TestCapacityPredictor_Learning(t *testing.T) {
	cp := NewCapacityPredictor()

	// Record initial data
	for i := 0; i < 5; i++ {
		cp.RecordMessage(1024, 10)
	}

	prediction1 := cp.Predict()

	// Record more data with different sizes
	for i := 0; i < 5; i++ {
		cp.RecordMessage(8192, 50)
	}

	prediction2 := cp.Predict()

	// Second prediction should be different due to new data
	if prediction1.OptimalBufferSize == prediction2.OptimalBufferSize {
		t.Error("Expected prediction to change with new data")
	}
}

func TestCapacityPredictor_Reset(t *testing.T) {
	cp := NewCapacityPredictor()

	cp.RecordMessage(1024, 10)
	cp.Predict()

	cp.Reset()

	if cp.lastPrediction != nil {
		t.Error("Expected lastPrediction to be nil after reset")
	}
}

func TestResponseBufferPool_GetPut(t *testing.T) {
	pool := NewResponseBufferPool()

	// Get buffer
	buf := pool.Get(1024)
	if buf == nil {
		t.Error("Expected non-nil buffer")
	}

	// Put buffer back
	pool.Put(buf)

	// Get again
	buf2 := pool.Get(1024)
	if buf2 == nil {
		t.Error("Expected non-nil buffer")
	}
}

func TestResponseBufferPool_SizeClasses(t *testing.T) {
	pool := NewResponseBufferPool()

	// Test small buffer
	smallBuf := pool.Get(2048)
	if cap(*smallBuf) < 2048 {
		t.Error("Small buffer capacity too small")
	}
	pool.Put(smallBuf)

	// Test medium buffer
	mediumBuf := pool.Get(32768)
	if cap(*mediumBuf) < 32768 {
		t.Error("Medium buffer capacity too small")
	}
	pool.Put(mediumBuf)

	// Test large buffer
	largeBuf := pool.Get(262144)
	if cap(*largeBuf) < 262144 {
		t.Error("Large buffer capacity too small")
	}
	pool.Put(largeBuf)

	// Test huge buffer - request size that falls in huge class (>= 262144)
	// Use a fresh pool to avoid buffer reuse issues
	hugePool := NewResponseBufferPool()
	hugeBuf := hugePool.Get(262145)
	// Should get huge pool with capacity 1048576
	if cap(*hugeBuf) < 262145 {
		t.Errorf("Huge buffer capacity %d too small for request %d", cap(*hugeBuf), 262145)
	}
	hugePool.Put(hugeBuf)
}

func TestResponseBufferPool_Stats(t *testing.T) {
	pool := NewResponseBufferPool()

	// Get some buffers
	for i := 0; i < 10; i++ {
		buf := pool.Get(1024)
		pool.Put(buf)
	}

	stats := pool.GetStats()

	if len(stats.Hits) == 0 {
		t.Error("Expected hit statistics")
	}
}

func TestResponseBufferPool_GetPoolForMethod(t *testing.T) {
	pool := NewResponseBufferPool()

	// Test known hot method
	pool1 := pool.GetPoolForMethod("RPCGetCVE")
	if pool1 == nil {
		t.Error("Expected non-nil pool")
	}

	// Test unknown method (should return medium pool)
	pool2 := pool.GetPoolForMethod("UnknownMethod")
	if pool2 == nil {
		t.Error("Expected non-nil pool")
	}
}

// BenchmarkCapacityPredictor_Predict benchmarks prediction performance
func BenchmarkCapacityPredictor_Predict(b *testing.B) {
	cp := NewCapacityPredictor()

	// Pre-populate with data
	for i := 0; i < 100; i++ {
		cp.RecordMessage(1024, 100)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cp.Predict()
	}
}

// BenchmarkCapacityPredictor_RecordMessage benchmarks recording performance
func BenchmarkCapacityPredictor_RecordMessage(b *testing.B) {
	cp := NewCapacityPredictor()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cp.RecordMessage(1024, 100)
	}
}

// BenchmarkResponseBufferPool_Get benchmarks buffer retrieval
func BenchmarkResponseBufferPool_Get(b *testing.B) {
	pool := NewResponseBufferPool()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get(1024)
		pool.Put(buf)
	}
}

// BenchmarkResponseBufferPool_GetLarge benchmarks large buffer retrieval
func BenchmarkResponseBufferPool_GetLarge(b *testing.B) {
	pool := NewResponseBufferPool()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get(524288)
		pool.Put(buf)
	}
}

// BenchmarkResponseBufferPool_Concurrent benchmarks concurrent access
func BenchmarkResponseBufferPool_Concurrent(b *testing.B) {
	pool := NewResponseBufferPool()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get(1024)
			pool.Put(buf)
		}
	})
}

// TestCapacityPredictionAccuracy tests prediction accuracy
func TestCapacityPredictionAccuracy(t *testing.T) {
	cp := NewCapacityPredictor()

	// Record messages with known pattern
	for i := 0; i < 50; i++ {
		cp.RecordMessage(1024, 100)
		time.Sleep(time.Millisecond)
	}

	prediction := cp.Predict()

	t.Logf("Prediction:")
	t.Logf("  Buffer Size: %d", prediction.OptimalBufferSize)
	t.Logf("  Batch Size: %d", prediction.OptimalBatchSize)
	t.Logf("  Expected Load: %.2f msg/s", prediction.ExpectedLoad)
	t.Logf("  Confidence: %.2f", prediction.Confidence)

	// Based on recorded data, make assertions
	if prediction.OptimalBufferSize < 1024*2 {
		t.Errorf("Buffer size too small based on history: %d", prediction.OptimalBufferSize)
	}
	if prediction.Confidence > 1.0 || prediction.Confidence < 0.0 {
		t.Errorf("Invalid confidence: %f", prediction.Confidence)
	}
}

// TestCapacityPredictor_AdaptiveLearning tests adaptive learning
func TestCapacityPredictor_AdaptiveLearning(t *testing.T) {
	cp := NewCapacityPredictor()

	// Phase 1: Record small messages
	for i := 0; i < 20; i++ {
		cp.RecordMessage(1024, 10)
	}

	prediction1 := cp.Predict()
	t.Logf("Phase 1 - Buffer: %d, Batch: %d", prediction1.OptimalBufferSize, prediction1.OptimalBatchSize)

	// Phase 2: Record large messages (pattern change)
	for i := 0; i < 20; i++ {
		cp.RecordMessage(8192, 200)
	}

	prediction2 := cp.Predict()
	t.Logf("Phase 2 - Buffer: %d, Batch: %d", prediction2.OptimalBufferSize, prediction2.OptimalBatchSize)

	// Phase 3: Record mixed messages
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			cp.RecordMessage(1024, 10)
		} else {
			cp.RecordMessage(8192, 200)
		}
	}

	prediction3 := cp.Predict()
	t.Logf("Phase 3 - Buffer: %d, Batch: %d", prediction3.OptimalBufferSize, prediction3.OptimalBatchSize)

	// Predictions should adapt to pattern changes
	if prediction1.OptimalBufferSize == prediction2.OptimalBufferSize {
		t.Error("Expected prediction to adapt to pattern change")
	}
}

// TestResponseBufferPool_ResetStats tests stats reset
func TestResponseBufferPool_ResetStats(t *testing.T) {
	pool := NewResponseBufferPool()

	// Generate some activity
	for i := 0; i < 10; i++ {
		buf := pool.Get(1024)
		pool.Put(buf)
	}

	stats1 := pool.GetStats()
	if len(stats1.Hits) == 0 {
		t.Error("Expected hit statistics before reset")
	}

	// Reset stats
	pool.ResetStats()

	stats2 := pool.GetStats()
	if len(stats2.Hits) != 0 {
		t.Error("Expected no hit statistics after reset")
	}
}
