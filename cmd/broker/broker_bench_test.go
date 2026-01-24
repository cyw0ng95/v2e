package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// BenchmarkBrokerSendMessage benchmarks the basic message sending performance
func BenchmarkBrokerSendMessage(b *testing.B) {
	broker := NewBroker()
	defer func() {
		_ = broker.Shutdown()
	}()

	// Start a background goroutine to drain the broker messages channel
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-broker.messages:
				// discard
			}
		}
	}()
	defer close(stop)

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

	// Drain messages to avoid blocking when channel fills
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-broker.messages:
			}
		}
	}()
	defer close(stop)

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

// BenchmarkBrokerLargeMessage benchmarks handling of large messages
func BenchmarkBrokerLargeMessage(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Drain messages to avoid blocking
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-broker.messages:
			}
		}
	}()
	defer close(stop)

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

	// Drain messages to avoid blocking when many messages are sent
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-broker.messages:
			}
		}
	}()
	defer close(stop)

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

func BenchmarkGenerateCorrelationID(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.GenerateCorrelationID()
	}
}

// BenchmarkGetMessageCount micro-benchmark for GetMessageCount
func BenchmarkGetMessageCount(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.GetMessageCount()
	}
}

// BenchmarkGetAllEndpoints micro-benchmark for GetAllEndpoints
func BenchmarkGetAllEndpoints(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Register many endpoints
	for i := 0; i < 100; i++ {
		broker.RegisterEndpoint("proc1", fmt.Sprintf("ep-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.GetAllEndpoints()
	}
}

// BenchmarkSendMessageNoAlloc sends the same pre-allocated message repeatedly
func BenchmarkSendMessageNoAlloc(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Drainer
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-broker.messages:
			}
		}
	}()
	defer close(stop)

	msg, _ := proc.NewRequestMessage("bench-noalloc", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.SendMessage(msg)
	}
}

// BenchmarkRegisterEndpoint measures performance of registering endpoints
func BenchmarkRegisterEndpoint(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broker.RegisterEndpoint("proc-register", fmt.Sprintf("ep-%d", i))
	}
}

// BenchmarkGetEndpoints measures performance of fetching endpoints for a process
func BenchmarkGetEndpoints(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	// Pre-register endpoints
	for i := 0; i < 1000; i++ {
		broker.RegisterEndpoint("proc-get", fmt.Sprintf("ep-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.GetEndpoints("proc-get")
	}
}

// BenchmarkRouteResponseToPending benchmarks routing a response to a pending request
func BenchmarkRouteResponseToPending(b *testing.B) {
	broker := NewBroker()
	defer broker.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a unique correlation ID per iteration
		corr := broker.GenerateCorrelationID()

		// Create buffered channel and register pending request
		ch := make(chan *proc.Message, 1)
		broker.pendingMu.Lock()
		broker.pendingRequests[corr] = &PendingRequest{
			SourceProcess: "bench-source",
			ResponseChan:  ch,
			Timestamp:     time.Now(),
		}
		broker.pendingMu.Unlock()

		// Build response message matching the correlation ID
		resp, _ := proc.NewResponseMessage("RPCMethod", nil)
		resp.CorrelationID = corr
		resp.Source = "proc"
		resp.Target = "bench-source"

		// Route the response and read from the channel
		if err := broker.RouteMessage(resp, "proc"); err != nil {
			b.Fatal(err)
		}

		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			b.Fatal("timeout waiting for routed response")
		}
	}
}
