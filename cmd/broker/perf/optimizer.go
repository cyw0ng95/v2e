package perf

import (
	"runtime"
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
	statsSyncDone     chan struct{}

	optimizedMessages chan *proc.Message

	numWorkers int
	workerWG   sync.WaitGroup

	// Moving-window metrics
	metricsMu          sync.Mutex
	lastTotalMessages  int64
	lastStatsTimestamp time.Time
}

func New(router routing.Router) *Optimizer {
	n := runtime.NumCPU()
	if n < 4 {
		n = 4
	}
	opt := &Optimizer{
		router:            router,
		statsSyncInterval: 100 * time.Millisecond,
		optimizedMessages: make(chan *proc.Message, 1000),
		numWorkers:        n,
		statsSyncDone:     make(chan struct{}),
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
	for {
		select {
		case msg := <-o.optimizedMessages:
			if msg.Target == "broker" {
				_ = o.router.ProcessBrokerMessage(msg)
			} else {
				_ = o.router.Route(msg, msg.Source)
			}
			o.updateAtomic(msg, true)
		case <-o.statsSyncDone:
			return
		}
	}
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
func (o *Optimizer) Offer(msg *proc.Message) {
	select {
	case o.optimizedMessages <- msg:
	default:
		// drop to avoid blocking; policy can change later
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
		"go_routines":              runtime.NumGoroutine(),
	}
}

func (o *Optimizer) Stop() {
	o.statsSyncTicker.Stop()
	close(o.statsSyncDone)
	o.workerWG.Wait()
}
