package perf

import (
	"context"
	"fmt"
	"hash/fnv"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/cmd/v2broker/routing"
	"github.com/cyw0ng95/v2e/cmd/v2broker/sched"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	"golang.org/x/sys/unix"
)

// OptimizedRouter provides optimized message routing
type OptimizedRouter struct {
	routes map[string]chan *proc.Message
	mu     sync.RWMutex
}

// NewOptimizedRouter creates a new optimized router
func NewOptimizedRouter() *OptimizedRouter {
	return &OptimizedRouter{
		routes: make(map[string]chan *proc.Message),
	}
}

// RouteFast performs fast message routing with timeout
func (r *OptimizedRouter) RouteFast(msg *proc.Message) error {
	r.mu.RLock()
	ch, exists := r.routes[msg.Target]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no route for target: %s", msg.Target)
	}

	select {
	case ch <- msg:
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("timeout routing message")
	}
}

// WorkStealingScheduler provides work stealing for balanced load distribution
type WorkStealingScheduler struct {
	queues  []chan *proc.Message
	workers []int
	mu      sync.Mutex
}

// NewWorkStealingScheduler creates a new work stealing scheduler
func NewWorkStealingScheduler(numQueues int) *WorkStealingScheduler {
	queues := make([]chan *proc.Message, numQueues)
	for i := range queues {
		queues[i] = make(chan *proc.Message, 100) // Buffered channel
	}

	return &WorkStealingScheduler{
		queues:  queues,
		workers: make([]int, numQueues),
	}
}

// Dispatch distributes messages using work stealing algorithm
func (ws *WorkStealingScheduler) Dispatch(msg *proc.Message) {
	// Use consistent hashing to determine primary queue
	hash := fnv.New32a()
	hash.Write([]byte(msg.Target))
	idx := hash.Sum32() % uint32(len(ws.queues))

	select {
	case ws.queues[idx] <- msg:
		// Successfully added to primary queue
	default:
		// Primary queue full, try work stealing
		ws.stealWork(msg)
	}
}

// stealWork attempts to find an available queue when primary is full
func (ws *WorkStealingScheduler) stealWork(msg *proc.Message) {
	for i := range ws.queues {
		select {
		case ws.queues[i] <- msg:
			// Successfully stole a slot
			return
		default:
			// Queue is full, continue to next
		}
	}
	// All queues full, block on first queue as last resort
	ws.queues[0] <- msg
}

// Optimizer provides a modular performance optimizer decoupled from cmd/broker.
// This initial skeleton wires to interfaces so cmd/broker can adopt it incrementally.
type Optimizer struct {
	router routing.Router

	// Atomic stats to minimize lock contention
	totalSent     int64
	totalReceived int64
	requestCount  int64
	responseCount int64
	eventCount    int64
	errorCount    int64

	statsSyncInterval time.Duration
	statsSyncTicker   *time.Ticker
	ctx               context.Context
	cancel            context.CancelFunc

	optimizedMessages chan *proc.Message

	numWorkers int
	workerWG   sync.WaitGroup
	// droppedMessages counts messages dropped by Offer when the queue is full
	droppedMessages int64
	// bufferCap holds the configured capacity of optimizedMessages channel
	bufferCap int
	// offerPolicy controls enqueue behavior: "drop" (default), "block", "timeout"
	offerPolicy string
	// offerTimeout is used when offerPolicy=="timeout"
	offerTimeout time.Duration
	// dropOldest policy and batching
	// batchSize is number of messages to collect before flush (1 = immediate)
	batchSize int
	// flushInterval is the maximum wait time to collect a batch
	flushInterval time.Duration

	// Moving-window metrics
	metricsMu          sync.Mutex
	lastTotalMessages  int64
	lastStatsTimestamp time.Time
	// logger for structured logging
	logger *common.Logger

	// Adaptive optimization components
	monitor        *sched.SystemMonitor
	adaptiveOpt    *sched.AdaptiveOptimizer
	adaptationMu   sync.Mutex
	adaptationFreq time.Duration
	lastAdaptation time.Time

	// Permit integration (Phase 2 UEE)
	permitIntegration *PermitIntegration
	
	// Analysis optimizer for service-level optimization
	analysisOptimizer *AnalysisOptimizer
}

func (o *Optimizer) EnableAdaptiveOptimization() {
	// Set up callback to receive metrics from the monitor
	o.monitor.SetCallback(func(metrics sched.LoadMetrics) {
		// Update adaptive optimizer with new metrics
		err := o.adaptiveOpt.Observe(metrics)
		if err != nil && o.logger != nil {
			o.logger.Warn("Error observing metrics: %v", err)
		}

		// Check if it's time to adapt parameters
		o.adaptationMu.Lock()
		if time.Since(o.lastAdaptation) >= o.adaptationFreq {
			err := o.adaptiveOpt.AdjustConfiguration()
			if err != nil && o.logger != nil {
				o.logger.Warn("Error adjusting configuration: %v", err)
			}
			o.lastAdaptation = time.Now()

			// Apply the adjusted parameters to the optimizer
			o.applyAdaptedParameters()
		}
		o.adaptationMu.Unlock()
	})

	// Start the monitor
	o.monitor.Start()
}

func (o *Optimizer) applyAdaptedParameters() {
	metrics := o.adaptiveOpt.GetMetrics()

	if bufferCap, ok := metrics["buffer_capacity"].(int); ok {
		if bufferCap != o.bufferCap {
			// Note: We can't easily change channel capacity at runtime
			// This would require recreating the channel, which is complex
			if o.logger != nil {
				o.logger.Info("Buffer capacity change suggested: %d -> %d", o.bufferCap, bufferCap)
			}
		}
	}

	if workerCount, ok := metrics["worker_count"].(int); ok {
		if workerCount != o.numWorkers {
			if o.logger != nil {
				o.logger.Info("Adjusting worker count: %d -> %d", o.numWorkers, workerCount)
			}

			// Adjust worker count by adding or removing workers
			o.adjustWorkerCount(workerCount)
		}
	}

	if batchSize, ok := metrics["batch_size"].(int); ok {
		o.batchSize = batchSize
		if o.logger != nil {
			o.logger.Info("Adjusted batch size to: %d", batchSize)
		}
	}

	if flushInterval, ok := metrics["flush_interval"].(time.Duration); ok {
		o.flushInterval = flushInterval
		if o.logger != nil {
			o.logger.Info("Adjusted flush interval to: %v", flushInterval)
		}
	}
}

func (o *Optimizer) adjustWorkerCount(newCount int) {
	currentCount := o.numWorkers

	if newCount > currentCount {
		// Add more workers
		for i := currentCount; i < newCount; i++ {
			o.workerWG.Add(1)
			go o.worker(i)
		}
		o.numWorkers = newCount
	} else if newCount < currentCount {
		// Reducing workers is complex and potentially unsafe
		// For now, we'll just log that we'd like to reduce
		if o.logger != nil {
			o.logger.Info("Would like to reduce worker count: %d -> %d, but reducing workers is not implemented", currentCount, newCount)
		}
		// In a production system, you'd need a more sophisticated approach
		// to safely shut down worker goroutines
	}
}

// setProcessPriority sets the broker process to high priority (-10)
// to ensure message routing is not starved by other processes.
func setProcessPriority() {
	// Set process priority to -10 (high priority)
	// PRIO_PROCESS with pid 0 means current process
	err := unix.Setpriority(unix.PRIO_PROCESS, 0, -10)
	if err != nil {
		// Log but don't fail - this requires CAP_SYS_NICE capability
		// In production, the broker should run with appropriate permissions
	}
}

// setCPUAffinity binds a worker goroutine to a specific CPU core.
// This reduces cache misses and context switch overhead.
func setCPUAffinity(workerID int) {
	numCPU := runtime.NumCPU()
	if numCPU <= 1 {
		return // No point in pinning on single-core systems
	}

	// Distribute workers across available CPUs
	cpu := workerID % numCPU

	var cpuSet unix.CPUSet
	cpuSet.Zero()
	cpuSet.Set(cpu)

	// Get current thread ID (LWP)
	// Note: unix.Gettid() returns the thread ID
	tid := unix.Gettid()

	// Set CPU affinity for this thread
	err := unix.SchedSetaffinity(tid, &cpuSet)
	if err != nil {
		// Log but don't fail - this requires appropriate permissions
		// In production, the broker should run with CAP_SYS_NICE
	}
}

// setIOPriority sets the I/O priority for the current thread to real-time class.
// This ensures disk I/O operations don't block the hot path of message routing.
func setIOPriority() {
	// I/O priority class and priority level
	// IOPRIO_CLASS_RT = 1 (Real-Time)
	// Priority 0 is highest within RT class
	const (
		IOPRIO_CLASS_SHIFT = 13
		IOPRIO_CLASS_RT    = 1
		IOPRIO_PRIO_VALUE  = 0
	)

	// Construct ioprio value: (class << IOPRIO_CLASS_SHIFT) | prio
	ioprio := (IOPRIO_CLASS_RT << IOPRIO_CLASS_SHIFT) | IOPRIO_PRIO_VALUE

	// Set I/O priority for current thread
	// IOPRIO_WHO_PROCESS = 1, pid = 0 means current thread
	const IOPRIO_WHO_PROCESS = 1
	_, _, errno := unix.Syscall(unix.SYS_IOPRIO_SET, IOPRIO_WHO_PROCESS, 0, uintptr(ioprio))
	if errno != 0 {
		// Log but don't fail - this requires CAP_SYS_ADMIN capability
	}
}

func New(router routing.Router) *Optimizer {
	// Set process priority to high (-10) for better scheduling
	setProcessPriority()

	n := runtime.NumCPU()
	if n < 4 {
		n = 4
	}
	ctx, cancel := context.WithCancel(context.Background())
	opt := &Optimizer{
		router:            router,
		statsSyncInterval: 100 * time.Millisecond,
		optimizedMessages: make(chan *proc.Message, 1000),
		bufferCap:         1000,
		numWorkers:        n,
		ctx:               ctx,
		cancel:            cancel,
		monitor:           sched.NewSystemMonitor(5 * time.Second),
		adaptiveOpt:       sched.NewAdaptiveOptimizer(),
		adaptationFreq:    10 * time.Second,
		lastAdaptation:    time.Now(),
	}
	opt.statsSyncTicker = time.NewTicker(opt.statsSyncInterval)
	for i := 0; i < opt.numWorkers; i++ {
		opt.workerWG.Add(1)
		go opt.worker(i)
	}
	return opt
}

// Config holds the configuration for the Optimizer.
type Config struct {
	BufferCap      int
	NumWorkers     int
	StatsInterval  time.Duration
	OfferPolicy    string
	OfferTimeout   time.Duration
	BatchSize      int
	FlushInterval  time.Duration
	AdaptationFreq time.Duration
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	n := runtime.NumCPU()
	if n < 4 {
		n = 4
	}
	return Config{
		BufferCap:      1000,
		NumWorkers:     n,
		StatsInterval:  100 * time.Millisecond,
		OfferPolicy:    "drop",
		OfferTimeout:   0,
		BatchSize:      1,
		FlushInterval:  10 * time.Millisecond,
		AdaptationFreq: 10 * time.Second,
	}
}

// NewWithConfig constructs an Optimizer with the provided configuration.
func NewWithConfig(router routing.Router, cfg Config) *Optimizer {
	defaults := DefaultConfig()

	if cfg.BufferCap <= 0 {
		cfg.BufferCap = defaults.BufferCap
	}
	if cfg.NumWorkers <= 0 {
		cfg.NumWorkers = defaults.NumWorkers
	}
	if cfg.StatsInterval <= 0 {
		cfg.StatsInterval = defaults.StatsInterval
	}
	if cfg.OfferPolicy == "" {
		cfg.OfferPolicy = defaults.OfferPolicy
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaults.BatchSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = defaults.FlushInterval
	}
	if cfg.AdaptationFreq <= 0 {
		cfg.AdaptationFreq = defaults.AdaptationFreq
	}

	// Set process priority to high (-10) for better scheduling
	setProcessPriority()

	ctx, cancel := context.WithCancel(context.Background())
	opt := &Optimizer{
		router:            router,
		statsSyncInterval: cfg.StatsInterval,
		optimizedMessages: make(chan *proc.Message, cfg.BufferCap),
		bufferCap:         cfg.BufferCap,
		numWorkers:        cfg.NumWorkers,
		offerPolicy:       cfg.OfferPolicy,
		offerTimeout:      cfg.OfferTimeout,
		batchSize:         cfg.BatchSize,
		flushInterval:     cfg.FlushInterval,
		ctx:               ctx,
		cancel:            cancel,
		monitor:           sched.NewSystemMonitor(5 * time.Second),
		adaptiveOpt:       sched.NewAdaptiveOptimizer(),
		adaptationFreq:    cfg.AdaptationFreq,
		lastAdaptation:    time.Now(),
	}
	opt.statsSyncTicker = time.NewTicker(opt.statsSyncInterval)
	for i := 0; i < opt.numWorkers; i++ {
		opt.workerWG.Add(1)
		go opt.worker(i)
	}
	return opt
}

// NewWithParams constructs an Optimizer with tunable runtime parameters.
// Deprecated: Use NewWithConfig instead.
func NewWithParams(router routing.Router, bufferCap, numWorkers int, statsInterval time.Duration, offerPolicy string, offerTimeout time.Duration, batchSize int, flushInterval time.Duration) *Optimizer {
	return NewWithConfig(router, Config{
		BufferCap:     bufferCap,
		NumWorkers:    numWorkers,
		StatsInterval: statsInterval,
		OfferPolicy:   offerPolicy,
		OfferTimeout:  offerTimeout,
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
	})
}

func (o *Optimizer) worker(id int) {
	defer o.workerWG.Done()
	defer func() {
		if r := recover(); r != nil {
			if o.logger != nil {
				o.logger.Error("optimizer worker %d panic: %v", id, r)
			}
		}
	}()

	// Linux-native optimizations for deterministic performance
	// 1. Lock this goroutine to an OS thread to prevent migration
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 2. Set CPU affinity to bind this worker to a specific core
	setCPUAffinity(id)

	// 3. Set I/O priority to real-time class for workers doing persistent writes
	setIOPriority()

	for {
		// collect at least one message (blocking)
		var batch []*proc.Message
		select {
		case msg := <-o.optimizedMessages:
			batch = append(batch, msg)

			// Record message arrival in monitor
			if o.monitor != nil {
				o.monitor.RecordMessage()
			}
		case <-o.ctx.Done():
			return
		}

		// collect up to batchSize-1 more messages, waiting up to flushInterval
		if o.batchSize > 1 {
			deadline := time.NewTimer(o.flushInterval)
		collectLoop:
			for len(batch) < o.batchSize {
				select {
				case msg := <-o.optimizedMessages:
					batch = append(batch, msg)

					// Record message arrival in monitor
					if o.monitor != nil {
						o.monitor.RecordMessage()
					}
					if len(batch) >= o.batchSize {
						break collectLoop
					}
				case <-deadline.C:
					// flush what's collected so far
					break collectLoop
				case <-o.ctx.Done():
					deadline.Stop()
					return
				}
			}
			if !deadline.Stop() {
				select {
				case <-deadline.C:
				default:
				}
			}
		}

		// process batch
		startTime := time.Now()
		for _, msg := range batch {
			if msg.Target == "broker" {
				_ = o.router.ProcessBrokerMessage(msg)
			} else {
				_ = o.router.Route(msg, msg.Source)
			}
			o.updateAtomic(msg, true)
		}
		processingDuration := time.Since(startTime)

		// Record processing time in monitor if available
		if o.monitor != nil {
			// Calculate average latency per message
			if len(batch) > 0 {
				avgProcessingTime := processingDuration / time.Duration(len(batch))
				o.monitor.AddLatencySample(avgProcessingTime)
			}
		}
	}
}

// SetLogger attaches a logger to the optimizer for runtime logging.
func (o *Optimizer) SetLogger(l *common.Logger) {
	o.logger = l
}

func (o *Optimizer) updateAtomic(msg *proc.Message, sent bool) {
	if sent {
		atomic.AddInt64(&o.totalSent, 1)
	} else {
		atomic.AddInt64(&o.totalReceived, 1)
	}
	switch msg.Type {
	case proc.MessageTypeRequest:
		atomic.AddInt64(&o.requestCount, 1)
	case proc.MessageTypeResponse:
		atomic.AddInt64(&o.responseCount, 1)
	case proc.MessageTypeEvent:
		atomic.AddInt64(&o.eventCount, 1)
	case proc.MessageTypeError:
		atomic.AddInt64(&o.errorCount, 1)
	}
}

// Offer allows non-blocking enqueue to optimized queue.
// Offer attempts a non-blocking enqueue and returns whether the message was accepted.
func (o *Optimizer) Offer(msg *proc.Message) bool {
	// Update queue depth in monitor
	if o.monitor != nil {
		queueDepth := int64(cap(o.optimizedMessages)) - int64(len(o.optimizedMessages))
		o.monitor.UpdateMessageQueueDepth(queueDepth)
	}

	switch o.offerPolicy {
	case "block":
		// blocking send
		o.optimizedMessages <- msg
		return true
	case "timeout":
		// try to send within timeout
		if o.offerTimeout <= 0 {
			// treat zero as immediate drop
			select {
			case o.optimizedMessages <- msg:
				return true
			default:
				atomic.AddInt64(&o.droppedMessages, 1)
				return false
			}
		}
		timer := time.NewTimer(o.offerTimeout)
		defer timer.Stop()
		select {
		case o.optimizedMessages <- msg:
			return true
		case <-timer.C:
			atomic.AddInt64(&o.droppedMessages, 1)
			return false
		}
	case "drop_oldest":
		// remove oldest message if possible, then enqueue
		select {
		case o.optimizedMessages <- msg:
			return true
		default:
			// try to remove one oldest
			select {
			case <-o.optimizedMessages:
				atomic.AddInt64(&o.droppedMessages, 1)
			default:
			}
			// attempt to enqueue again
			select {
			case o.optimizedMessages <- msg:
				return true
			default:
				atomic.AddInt64(&o.droppedMessages, 1)
				return false
			}
		}
	default:
		// default: drop (non-blocking)
		select {
		case o.optimizedMessages <- msg:
			return true
		default:
			atomic.AddInt64(&o.droppedMessages, 1)
			return false
		}
	}
}

func (o *Optimizer) Metrics() map[string]interface{} {
	total := atomic.LoadInt64(&o.totalSent) + atomic.LoadInt64(&o.totalReceived)
	o.metricsMu.Lock()
	now := time.Now()
	var mps float64
	if !o.lastStatsTimestamp.IsZero() {
		dt := now.Sub(o.lastStatsTimestamp).Seconds()
		if dt > 0 {
			delta := float64(total - o.lastTotalMessages)
			mps = delta / dt
		}
	}
	o.lastTotalMessages = total
	o.lastStatsTimestamp = now
	o.metricsMu.Unlock()
	return map[string]interface{}{
		"total_messages_processed": total,
		"messages_per_second":      mps,
		"message_channel_buffer":   cap(o.optimizedMessages),
		"active_workers":           o.numWorkers,
		"dropped_messages":         atomic.LoadInt64(&o.droppedMessages),
		"go_routines":              runtime.NumGoroutine(),
	}
}

func (o *Optimizer) Stop() {
	// Stop the monitor if it exists
	if o.monitor != nil {
		o.monitor.Stop()
	}

	// Cancel the context to stop the main processing loop
	o.cancel()

	// Wait for all workers to finish
	o.workerWG.Wait()

	// Stop the stats sync ticker
	if o.statsSyncTicker != nil {
		o.statsSyncTicker.Stop()
	}
}

// SetAnalysisOptimizer attaches an AnalysisOptimizer to the broker
func (o *Optimizer) SetAnalysisOptimizer(ao *AnalysisOptimizer) {
o.analysisOptimizer = ao
if o.logger != nil {
o.logger.Info("Analysis optimizer attached to broker")
}
}

// GetAnalysisOptimizer returns the analysis optimizer
func (o *Optimizer) GetAnalysisOptimizer() *AnalysisOptimizer {
return o.analysisOptimizer
}

// StartConflictMonitor starts monitoring for service conflicts
func (o *Optimizer) StartConflictMonitor() {
if o.analysisOptimizer == nil {
if o.logger != nil {
o.logger.Warn("Cannot start conflict monitor: AnalysisOptimizer not set")
}
return
}

if o.logger != nil {
o.logger.Info("Starting service conflict monitor")
}

go o.conflictMonitorLoop()
}

// conflictMonitorLoop monitors for service conflicts and resolves them
func (o *Optimizer) conflictMonitorLoop() {
ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds
defer ticker.Stop()

for {
select {
case <-ticker.C:
if hasConflict, conflicts := o.analysisOptimizer.DetectConflict(); hasConflict {
o.analysisOptimizer.ResolveConflict(conflicts)
} else {
// Clear throttles if no conflicts detected
o.analysisOptimizer.ClearThrottles()
}
case <-o.ctx.Done():
if o.logger != nil {
o.logger.Info("Conflict monitor stopping")
}
return
}
}
}
