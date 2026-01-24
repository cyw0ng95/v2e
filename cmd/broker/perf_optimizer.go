package main

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// BrokerPerformanceOptimizer provides performance enhancements for the broker.
// It's placed in cmd/broker and named to make its purpose explicit.
type BrokerPerformanceOptimizer struct {
	broker *Broker

	// Optimized statistics that reduce lock contention
	atomicStats struct {
		totalSent     int64
		totalReceived int64
		requestCount  int64
		responseCount int64
		eventCount    int64
		errorCount    int64
	}

	// Reduce frequency of stats synchronization to lower lock contention
	statsSyncInterval time.Duration
	statsSyncTicker   *time.Ticker
	statsSyncDone     chan struct{}

	// Pool for frequently allocated objects to reduce GC pressure
	messagePool sync.Pool
	processPool sync.Pool

	// Optimized message channel with larger buffer
	optimizedMessages chan *proc.Message

	// Worker pool for processing messages asynchronously
	numWorkers int
	workerWG   sync.WaitGroup
}

// NewBrokerPerformanceOptimizer creates a new performance optimizer for the broker.
func NewBrokerPerformanceOptimizer(b *Broker) *BrokerPerformanceOptimizer {
	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4 // Minimum workers for good performance
	}

	opt := &BrokerPerformanceOptimizer{
		broker:            b,
		statsSyncInterval: 100 * time.Millisecond, // Sync stats every 100ms instead of on every message
		numWorkers:        numWorkers,
		optimizedMessages: make(chan *proc.Message, 1000), // Increased buffer from 100 to 1000
	}

	// Initialize pools
	opt.messagePool.New = func() interface{} {
		return &proc.Message{}
	}

	opt.processPool.New = func() interface{} {
		return &Process{}
	}

	// Start stats sync ticker
	opt.statsSyncTicker = time.NewTicker(opt.statsSyncInterval)
	opt.statsSyncDone = make(chan struct{})

	// Start worker pool
	opt.startWorkerPool()

	return opt
}

// startWorkerPool starts the worker pool for asynchronous message processing
func (opt *BrokerPerformanceOptimizer) startWorkerPool() {
	for i := 0; i < opt.numWorkers; i++ {
		opt.workerWG.Add(1)
		go opt.messageProcessor(i)
	}
}

// messageProcessor handles message processing in worker goroutines
func (opt *BrokerPerformanceOptimizer) messageProcessor(workerID int) {
	defer opt.workerWG.Done()

	for {
		select {
		case msg := <-opt.optimizedMessages:
			// Process the message through the broker's routing
			if msg.Target == "broker" {
				opt.broker.ProcessMessage(msg)
			} else {
				// Route to appropriate destination
				var sourceProcess string
				if msg.Source != "" {
					sourceProcess = msg.Source
				} else {
					sourceProcess = "internal"
				}
				opt.broker.RouteMessage(msg, sourceProcess)
			}

			// Update atomic stats (non-blocking)
			opt.updateAtomicStats(msg, true)

		case <-opt.broker.ctx.Done():
			return
		}
	}
}

// updateAtomicStats updates statistics atomically to reduce lock contention
func (opt *BrokerPerformanceOptimizer) updateAtomicStats(msg *proc.Message, isSent bool) {
	if isSent {
		atomic.AddInt64(&opt.atomicStats.totalSent, 1)
	} else {
		atomic.AddInt64(&opt.atomicStats.totalReceived, 1)
	}

	switch msg.Type {
	case proc.MessageTypeRequest:
		atomic.AddInt64(&opt.atomicStats.requestCount, 1)
	case proc.MessageTypeResponse:
		atomic.AddInt64(&opt.atomicStats.responseCount, 1)
	case proc.MessageTypeEvent:
		atomic.AddInt64(&opt.atomicStats.eventCount, 1)
	case proc.MessageTypeError:
		atomic.AddInt64(&opt.atomicStats.errorCount, 1)
	}
}

// syncStats periodically syncs atomic stats to the broker's stats
func (opt *BrokerPerformanceOptimizer) syncStats() {
	go func() {
		for {
			select {
			case <-opt.statsSyncTicker.C:
				// Copy atomic values to broker stats
				opt.broker.statsMu.Lock()

				// Update broker's stats from atomic counters
				opt.broker.stats.TotalSent = atomic.LoadInt64(&opt.atomicStats.totalSent)
				opt.broker.stats.TotalReceived = atomic.LoadInt64(&opt.atomicStats.totalReceived)
				opt.broker.stats.RequestCount = atomic.LoadInt64(&opt.atomicStats.requestCount)
				opt.broker.stats.ResponseCount = atomic.LoadInt64(&opt.atomicStats.responseCount)
				opt.broker.stats.EventCount = atomic.LoadInt64(&opt.atomicStats.eventCount)
				opt.broker.stats.ErrorCount = atomic.LoadInt64(&opt.atomicStats.errorCount)

				// Set timestamp if not already set
				if opt.broker.stats.FirstMessageTime.IsZero() {
					opt.broker.stats.FirstMessageTime = time.Now()
				}
				opt.broker.stats.LastMessageTime = time.Now()

				opt.broker.statsMu.Unlock()

			case <-opt.statsSyncDone:
				return
			}
		}
	}()
}

// SendMessageOptimized sends a message using optimized pathways
func (opt *BrokerPerformanceOptimizer) SendMessageOptimized(msg *proc.Message) error {
	// Use the optimized channel with larger buffer
	select {
	case opt.optimizedMessages <- msg:
		// Update stats atomically (non-blocking)
		opt.updateAtomicStats(msg, true)
		return nil
	case <-opt.broker.ctx.Done():
		return context.Canceled
	}
}

// GetOptimizedStats returns performance-optimized statistics
func (opt *BrokerPerformanceOptimizer) GetOptimizedStats() MessageStats {
	// Return a snapshot with latest atomic values
	opt.broker.statsMu.RLock()
	defer opt.broker.statsMu.RUnlock()

	return MessageStats{
		TotalSent:        atomic.LoadInt64(&opt.atomicStats.totalSent),
		TotalReceived:    atomic.LoadInt64(&opt.atomicStats.totalReceived),
		RequestCount:     atomic.LoadInt64(&opt.atomicStats.requestCount),
		ResponseCount:    atomic.LoadInt64(&opt.atomicStats.responseCount),
		EventCount:       atomic.LoadInt64(&opt.atomicStats.eventCount),
		ErrorCount:       atomic.LoadInt64(&opt.atomicStats.errorCount),
		FirstMessageTime: opt.broker.stats.FirstMessageTime,
		LastMessageTime:  time.Now(), // Most recent timestamp
	}
}

// Start begins the optimization services
func (opt *BrokerPerformanceOptimizer) Start() {
	opt.syncStats()

	// Optionally replace the broker's message channel with the optimized one
	// This would require careful coordination to avoid breaking existing functionality
	go func() {
		// Drain the original channel to the optimized one
		for {
			select {
			case msg := <-opt.broker.messages:
				select {
				case opt.optimizedMessages <- msg:
					// Message transferred successfully
				case <-opt.broker.ctx.Done():
					return
				}
			case <-opt.broker.ctx.Done():
				return
			}
		}
	}()
}

// Stop shuts down the optimization services
func (opt *BrokerPerformanceOptimizer) Stop() {
	opt.statsSyncTicker.Stop()
	close(opt.statsSyncDone)

	// Wait for workers to finish
	opt.workerWG.Wait()
}

// ApplyToBroker applies performance optimizations to the broker
func (opt *BrokerPerformanceOptimizer) ApplyToBroker() {
	// We can't directly modify unexported fields, so we provide wrapper methods
	// The optimization works by intercepting message flows

	opt.Start()
}

// GetPerformanceMetrics returns performance metrics
func (opt *BrokerPerformanceOptimizer) GetPerformanceMetrics() map[string]interface{} {
	stats := opt.GetOptimizedStats()

	return map[string]interface{}{
		"total_messages_processed": stats.TotalSent + stats.TotalReceived,
		"messages_per_second":      float64(stats.TotalSent+stats.TotalReceived) / opt.statsSyncInterval.Seconds(),
		"message_channel_buffer":   cap(opt.optimizedMessages),
		"active_workers":           opt.numWorkers,
		"gc_runs":                  "unavailable",
		"go_routines":              runtime.NumGoroutine(),
	}
}
