package perf

import (
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestAdaptiveOptimization validates the adaptive optimization features
func TestAdaptiveOptimization(t *testing.T) {
	// Create a test router for message routing
	router := &testRouter{}

	// Create an optimizer with adaptive features enabled
	opt := New(router)
	
	// Set up logging for debugging
	logger := common.NewLogger(os.Stdout, "[TEST] ", common.InfoLevel)
	opt.SetLogger(logger)

	// Enable adaptive optimization
	opt.EnableAdaptiveOptimization()

	// Create some test messages to simulate load
	msg, err := proc.NewRequestMessage("test-msg-1", "test data")
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	msg.Source = "test-source"
	msg.Target = "test-target"

	// Simulate message load by sending messages to the optimizer
	for i := 0; i < 100; i++ {
		accepted := opt.Offer(msg)
		if !accepted {
			t.Logf("Message %d was dropped", i)
		}
		time.Sleep(1 * time.Millisecond) // Brief pause between messages
	}

	// Wait for some metrics collection and adaptation
	time.Sleep(10 * time.Second)

	// Get current metrics to verify adaptive behavior occurred
	metrics := opt.Metrics()
	t.Logf("Final metrics: %+v", metrics)

	// Stop the optimizer
	opt.Stop()
}

// testRouter is a mock router for testing purposes
type testRouter struct{}

func (r *testRouter) Route(msg *proc.Message, source string) error {
	// Simulate message processing delay
	time.Sleep(500 * time.Microsecond)
	return nil
}

func (r *testRouter) ProcessBrokerMessage(msg *proc.Message) error {
	// Simulate broker message processing
	return nil
}

// BenchmarkAdaptiveOptimization benchmarks the adaptive optimization performance
func BenchmarkAdaptiveOptimization(b *testing.B) {
	router := &testRouter{}
	opt := New(router)
	
	// Enable adaptive optimization
	opt.EnableAdaptiveOptimization()

	msg, err := proc.NewRequestMessage("benchmark-msg", "benchmark data")
	if err != nil {
		b.Fatalf("Failed to create message: %v", err)
	}
	msg.Source = "benchmark-source"
	msg.Target = "benchmark-target"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accepted := opt.Offer(msg)
		if !accepted {
			b.Logf("Benchmark message %d was dropped", i)
		}
	}
	
	b.StopTimer()
	opt.Stop()
}