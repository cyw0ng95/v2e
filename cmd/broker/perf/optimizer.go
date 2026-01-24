package perf

import (
	"context"
	"runtime"
	"github.com/cyw0ng95/v2e/pkg/common"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/cmd/broker/routing"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

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
}

func New(router routing.Router) *Optimizer {
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
	}
	opt.statsSyncTicker = time.NewTicker(opt.statsSyncInterval)
	for i := 0; i < opt.numWorkers; i++ {
		opt.workerWG.Add(1)
		go opt.worker(i)
	}
	return opt
}

// NewWithParams constructs an Optimizer with tunable runtime parameters.
// Pass bufferCap<=0 to use default (1000). Pass numWorkers<=0 to use CPU-based default.
// Pass statsInterval<=0 to use default (100ms).
func NewWithParams(router routing.Router, bufferCap, numWorkers int, statsInterval time.Duration, offerPolicy string, offerTimeout time.Duration, batchSize int, flushInterval time.Duration) *Optimizer {
	if bufferCap <= 0 {
		bufferCap = 1000
	}
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
		if numWorkers < 4 {
			numWorkers = 4
		}
	}
	if statsInterval <= 0 {
		statsInterval = 100 * time.Millisecond
	}
	if offerPolicy == "" {
		offerPolicy = "drop"
	}
	if batchSize <= 0 {
		batchSize = 1
	}
	if flushInterval <= 0 {
		flushInterval = 10 * time.Millisecond
	}
	ctx, cancel := context.WithCancel(context.Background())
	opt := &Optimizer{
		router:            router,
		statsSyncInterval: statsInterval,
		optimizedMessages: make(chan *proc.Message, bufferCap),
		bufferCap:         bufferCap,
		numWorkers:        numWorkers,
		offerPolicy:       offerPolicy,
		offerTimeout:      offerTimeout,
		batchSize:         batchSize,
		flushInterval:     flushInterval,
		ctx:               ctx,
		cancel:            cancel,
	}
	opt.statsSyncTicker = time.NewTicker(opt.statsSyncInterval)
	for i := 0; i < opt.numWorkers; i++ {
		opt.workerWG.Add(1)
		go opt.worker(i)
	}
	return opt
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
	for {
		// collect at least one message (blocking)
		var batch []*proc.Message
		select {
		case msg := <-o.optimizedMessages:
			batch = append(batch, msg)
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
		for _, msg := range batch {
			if msg.Target == "broker" {
				_ = o.router.ProcessBrokerMessage(msg)
			} else {
				_ = o.router.Route(msg, msg.Source)
			}
			o.updateAtomic(msg, true)
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
	o.statsSyncTicker.Stop()
	o.cancel()
	o.workerWG.Wait()
}
