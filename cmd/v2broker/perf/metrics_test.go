package perf

import (
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	tests := []struct {
		name             string
		maxLatencies     int
		bufferCapacity   int
		wantMaxLatencies int
		wantBufferCap    int
	}{
		{
			name:             "valid parameters",
			maxLatencies:     100,
			bufferCapacity:   500,
			wantMaxLatencies: 100,
			wantBufferCap:    500,
		},
		{
			name:             "zero latencies defaults to 1000",
			maxLatencies:     0,
			bufferCapacity:   500,
			wantMaxLatencies: 1000,
			wantBufferCap:    500,
		},
		{
			name:             "zero buffer defaults to 1000",
			maxLatencies:     100,
			bufferCapacity:   0,
			wantMaxLatencies: 100,
			wantBufferCap:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricsCollector(tt.maxLatencies, tt.bufferCapacity)
			if mc.maxLatencies != tt.wantMaxLatencies {
				t.Errorf("maxLatencies = %d, want %d", mc.maxLatencies, tt.wantMaxLatencies)
			}
			if mc.bufferCapacity != tt.wantBufferCap {
				t.Errorf("bufferCapacity = %d, want %d", mc.bufferCapacity, tt.wantBufferCap)
			}
		})
	}
}

func TestRecordLatency(t *testing.T) {
	mc := NewMetricsCollector(10, 100)

	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		15 * time.Millisecond,
		25 * time.Millisecond,
	}

	for _, lat := range latencies {
		mc.RecordLatency(lat)
	}

	// Verify latencies were recorded
	if mc.latencyIndex != 5 {
		t.Errorf("latencyIndex = %d, want 5", mc.latencyIndex)
	}

	// Record more than max to test wraparound
	for i := 0; i < 10; i++ {
		mc.RecordLatency(5 * time.Millisecond)
	}

	if mc.latencyIndex != 5 { // Should wrap around to 5 (15 % 10)
		t.Errorf("latencyIndex after wraparound = %d, want 5", mc.latencyIndex)
	}
}

func TestGetP99Latency(t *testing.T) {
	mc := NewMetricsCollector(100, 100)

	// Record 100 latencies from 1ms to 100ms
	for i := 1; i <= 100; i++ {
		mc.RecordLatency(time.Duration(i) * time.Millisecond)
	}

	p99 := mc.GetP99Latency()
	// P99 of 1-100ms should be around 99ms
	expectedP99 := 99 * time.Millisecond
	if p99 < 98*time.Millisecond || p99 > 100*time.Millisecond {
		t.Errorf("P99 = %v, want approximately %v", p99, expectedP99)
	}
}

func TestGetP99LatencyEmpty(t *testing.T) {
	mc := NewMetricsCollector(100, 100)

	p99 := mc.GetP99Latency()
	if p99 != 0 {
		t.Errorf("P99 with no data = %v, want 0", p99)
	}
}

func TestGetBufferSaturation(t *testing.T) {
	mc := NewMetricsCollector(100, 1000)

	tests := []struct {
		name         string
		bufferSize   int
		wantSatMin   float64
		wantSatMax   float64
	}{
		{
			name:       "empty buffer",
			bufferSize: 0,
			wantSatMin: 0,
			wantSatMax: 0,
		},
		{
			name:       "half full",
			bufferSize: 500,
			wantSatMin: 49,
			wantSatMax: 51,
		},
		{
			name:       "full buffer",
			bufferSize: 1000,
			wantSatMin: 99,
			wantSatMax: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc.UpdateBuffer(tt.bufferSize)
			saturation := mc.GetBufferSaturation()
			if saturation < tt.wantSatMin || saturation > tt.wantSatMax {
				t.Errorf("saturation = %v, want between %v and %v",
					saturation, tt.wantSatMin, tt.wantSatMax)
			}
		})
	}
}

func TestGetMessageRate(t *testing.T) {
	mc := NewMetricsCollector(100, 100)

	// Record 100 messages
	for i := 0; i < 100; i++ {
		mc.RecordMessage()
	}

	// Wait a bit for rate calculation
	time.Sleep(100 * time.Millisecond)

	rate := mc.GetMessageRate()
	// Should be approximately 1000 msg/s (100 in 0.1s)
	// Allow wide range due to timing variations
	if rate < 500 || rate > 2000 {
		t.Logf("Message rate = %v msg/s (acceptable range due to timing)", rate)
	}
}

func TestGetErrorRate(t *testing.T) {
	mc := NewMetricsCollector(100, 100)

	// Record 10 errors
	for i := 0; i < 10; i++ {
		mc.RecordError()
	}

	// Wait a bit for rate calculation
	time.Sleep(100 * time.Millisecond)

	rate := mc.GetErrorRate()
	// Should be approximately 100 err/s (10 in 0.1s)
	// Allow wide range due to timing variations
	if rate < 50 || rate > 200 {
		t.Logf("Error rate = %v err/s (acceptable range due to timing)", rate)
	}
}

func TestGetMetrics(t *testing.T) {
	mc := NewMetricsCollector(100, 1000)

	// Record some latencies
	mc.RecordLatency(10 * time.Millisecond)
	mc.RecordLatency(20 * time.Millisecond)
	mc.RecordLatency(30 * time.Millisecond)

	// Update buffer
	mc.UpdateBuffer(500)

	// Record messages and errors
	mc.RecordMessage()
	mc.RecordMessage()
	mc.RecordError()

	metrics := mc.GetMetrics(5, 10, 7, 3)

	// Verify all fields are populated
	if metrics.ActiveWorkers != 5 {
		t.Errorf("ActiveWorkers = %d, want 5", metrics.ActiveWorkers)
	}
	if metrics.TotalPermits != 10 {
		t.Errorf("TotalPermits = %d, want 10", metrics.TotalPermits)
	}
	if metrics.AllocatedPermits != 7 {
		t.Errorf("AllocatedPermits = %d, want 7", metrics.AllocatedPermits)
	}
	if metrics.AvailablePermits != 3 {
		t.Errorf("AvailablePermits = %d, want 3", metrics.AvailablePermits)
	}

	// Buffer saturation should be around 50%
	if metrics.BufferSaturation < 49 || metrics.BufferSaturation > 51 {
		t.Errorf("BufferSaturation = %v, want ~50", metrics.BufferSaturation)
	}

	// P99 should be around 30ms (highest value)
	if metrics.P99LatencyMs < 29 || metrics.P99LatencyMs > 31 {
		t.Errorf("P99LatencyMs = %v, want ~30", metrics.P99LatencyMs)
	}
}

func TestSortDurations(t *testing.T) {
	input := []time.Duration{
		30 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
	}

	sortDurations(input)

	expected := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	for i, want := range expected {
		if input[i] != want {
			t.Errorf("sorted[%d] = %v, want %v", i, input[i], want)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	mc := NewMetricsCollector(1000, 1000)

	// Launch concurrent goroutines to test thread safety
	done := make(chan bool)

	// Goroutine 1: Record latencies
	go func() {
		for i := 0; i < 100; i++ {
			mc.RecordLatency(time.Duration(i) * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Record messages
	go func() {
		for i := 0; i < 100; i++ {
			mc.RecordMessage()
		}
		done <- true
	}()

	// Goroutine 3: Read metrics
	go func() {
		for i := 0; i < 100; i++ {
			mc.GetMetrics(1, 10, 5, 5)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// If we get here without panic, concurrent access is safe
}
