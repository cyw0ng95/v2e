package main

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// BenchmarkBrokerSendMessage benchmarks the basic message sending performance
func BenchmarkBrokerSendMessage(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a sample message
	msg, _ := proc.NewRequestMessage("test-request", map[string]interface{}{
		"data": "test-data",
		"id":   123,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.SendMessage(msg)
	}
}

// BenchmarkBrokerSendMessageConcurrent benchmarks concurrent message sending
func BenchmarkBrokerSendMessageConcurrent(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a sample message
	msg, _ := proc.NewRequestMessage("test-request", map[string]interface{}{
		"data": "test-data",
		"id":   123,
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = broker.SendMessage(msg)
		}
	})
}

// BenchmarkBrokerProcessManagement benchmarks process management operations
func BenchmarkBrokerProcessManagement(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "test"}
	} else {
		cmd = "echo"
		args = []string{"test"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := "test-" + string(rune('a'+i%26))
		_, _ = broker.Spawn(id, cmd, args...)
		_ = broker.Kill(id)
	}
}

// BenchmarkBrokerMessageStats benchmarks stats operations
func BenchmarkBrokerMessageStats(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Send some messages to generate stats
	for i := 0; i < 10; i++ {
		msg, _ := proc.NewRequestMessage("test", nil)
		broker.SendMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.GetMessageStats()
	}
}

// BenchmarkBrokerConcurrentStats benchmarks concurrent access to stats
func BenchmarkBrokerConcurrentStats(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Send some messages to generate stats
	for i := 0; i < 10; i++ {
		msg, _ := proc.NewRequestMessage("test", nil)
		broker.SendMessage(msg)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = broker.GetMessageStats()
		}
	})
}

// BenchmarkBrokerMessageRoundTrip benchmarks a complete message round trip
func BenchmarkBrokerMessageRoundTrip(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a message
	msg, _ := proc.NewRequestMessage("test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Send message
		err := broker.SendMessage(msg)
		if err != nil {
			b.Fatal(err)
		}

		// Receive message
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, _ = broker.ReceiveMessage(ctx)
		cancel()
	}
}

// BenchmarkBrokerInvokeRPC benchmarks the RPC invocation performance
func BenchmarkBrokerInvokeRPC(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Start a simple RPC process
	info, err := broker.SpawnRPC("test-service", "sleep", "30")
	if err != nil {
		b.Skipf("Could not start test service: %v", err)
		return
	}
	defer broker.Kill(info.ID)

	// Wait for process to start
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Try to invoke RPC (this will likely timeout since sleep doesn't respond to RPC)
		_, err := broker.InvokeRPC("caller", "test-service", "NonExistentRPC", nil, 10*time.Millisecond)
		// We expect timeouts, so ignore them
		if err != nil && err.Error() != "timeout waiting for response from test-service" {
			b.Logf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkBrokerLargeMessage benchmarks handling of large messages
func BenchmarkBrokerLargeMessage(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a large payload
	largeData := make([]byte, 1024*10) // 10KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	msg, _ := proc.NewRequestMessage("large-message", map[string]interface{}{
		"data": string(largeData),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.SendMessage(msg)
	}
}

// BenchmarkBrokerManyConcurrentOperations benchmarks many concurrent operations
func BenchmarkBrokerManyConcurrentOperations(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Number of concurrent goroutines
	numGoroutines := 10
	messagesPerGoroutine := b.N / numGoroutines

	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				// Create and send a message
				msg, _ := proc.NewRequestMessage("concurrent-test", map[string]interface{}{
					"goroutine": goroutineID,
					"iteration": j,
				})
				_ = broker.SendMessage(msg)
			}
		}(i)
	}

	wg.Wait()
}

// BenchmarkPerformanceOptimizerSendMessage benchmarks the optimized message sending
func BenchmarkPerformanceOptimizerSendMessage(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	optimizer := NewPerformanceOptimizer(broker)
	defer optimizer.Stop()

	// Create a sample message
	msg, _ := proc.NewRequestMessage("test-request", map[string]interface{}{
		"data": "test-data",
		"id":   123,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = optimizer.SendMessageOptimized(msg)
	}
}

// BenchmarkBrokerMessageRouting benchmarks message routing performance
func BenchmarkBrokerMessageRouting(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Create a target process
	targetInfo, err := broker.Spawn("target-process", "sleep", "30")
	if err != nil {
		b.Skipf("Could not start target process: %v", err)
		return
	}
	defer broker.Kill(targetInfo.ID)

	// Wait for process to start
	time.Sleep(100 * time.Millisecond)

	// Create a message to route
	msg, _ := proc.NewRequestMessage("routing-test", nil)
	msg.Target = "target-process"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := broker.RouteMessage(msg, "source-process")
		if err != nil {
			// Expected since target process doesn't read from stdin
		}
	}
}