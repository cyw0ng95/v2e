package core

import (
	"context"
	"io"
	"sync"

	"github.com/cyw0ng95/v2e/cmd/v2broker/mq"
	"github.com/cyw0ng95/v2e/cmd/v2broker/transport"
	"github.com/cyw0ng95/v2e/pkg/broker"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	subprocess "github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Broker manages subprocesses and message passing.
type Broker struct {
	processes map[string]*Process
	messages  chan *proc.Message
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	logger    *common.Logger

	bus             *mq.Bus
	rpcEndpoints    map[string][]string
	endpointsMu     sync.RWMutex
	pendingRequests map[string]*PendingRequest
	pendingMu       sync.RWMutex
	correlationSeq  uint64
	spawner         broker.Spawner
	// optimizer optionally handles message routing asynchronously
	optimizer OptimizerInterface
	// transportManager manages communication transports for processes
	transportManager *transport.TransportManager
}

// NewBroker creates a new Broker instance.
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	bus := mq.NewBus(ctx, 100)
	b := &Broker{
		processes:        make(map[string]*Process),
		messages:         bus.Channel(),
		ctx:              ctx,
		cancel:           cancel,
		logger:           common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
		bus:              bus,
		rpcEndpoints:     make(map[string][]string),
		pendingRequests:  make(map[string]*PendingRequest),
		correlationSeq:   0,
		transportManager: transport.NewTransportManager(),
	}

	// Set transport error handler to log warnings
	b.transportManager.SetTransportErrorHandler(func(err error) {
		if b.logger != nil {
			b.logger.Warn("Transport background error: %v", err)
		}
	})

	// Ensure transport manager uses the same UDS base path as subprocesses
	b.transportManager.SetUdsBasePath(subprocess.DefaultProcUDSBasePath())

	// Install a default SpawnAdapter that delegates to existing spawn methods.

	return b
}

// InsertProcessForTest inserts a pre-constructed process into the broker (testing only).
func (b *Broker) InsertProcessForTest(p *Process) {
	b.mu.Lock()
	b.processes[p.info.ID] = p
	b.mu.Unlock()
}

// ProcessCount returns the number of tracked processes.
func (b *Broker) ProcessCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.processes)
}

// PendingRequestCount returns the number of pending request entries.
func (b *Broker) PendingRequestCount() int {
	b.pendingMu.RLock()
	defer b.pendingMu.RUnlock()
	return len(b.pendingRequests)
}

// AddPendingRequest registers a pending request entry. Intended for tests and benchmarks.
func (b *Broker) AddPendingRequest(correlationID string, pending *PendingRequest) {
	b.pendingMu.Lock()
	b.pendingRequests[correlationID] = pending
	b.pendingMu.Unlock()
}

// SetLogger sets the logger for the broker.
func (b *Broker) SetLogger(logger *common.Logger) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger = logger
}

// SetSpawner injects a pluggable Spawner implementation.
func (b *Broker) SetSpawner(s broker.Spawner) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spawner = s
}

// Spawner returns the currently configured Spawner (may be nil).
func (b *Broker) Spawner() broker.Spawner {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.spawner
}

// SetOptimizer attaches a performance optimizer to the Broker.
func (b *Broker) SetOptimizer(o OptimizerInterface) {
	b.mu.Lock()
	b.optimizer = o
	b.mu.Unlock()
	// attach broker logger to optimizer if possible
	if o != nil && b.logger != nil {
		o.SetLogger(b.logger)
		b.logger.Info("Optimizer attached")
	}
}

// Context returns the broker's context.
func (b *Broker) Context() context.Context {
	return b.ctx
}

// OptimizerInterface is a lightweight interface used by Broker to avoid
// importing the concrete optimizer implementation and creating an import cycle.
type OptimizerInterface interface {
	Offer(msg *proc.Message) bool
	Stop()
	Metrics() map[string]interface{}
	SetLogger(l *common.Logger)
}
