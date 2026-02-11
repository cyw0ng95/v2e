package core

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/cmd/v2broker/metrics"
	"github.com/cyw0ng95/v2e/cmd/v2broker/mq"
	"github.com/cyw0ng95/v2e/cmd/v2broker/perf"
	"github.com/cyw0ng95/v2e/cmd/v2broker/permits"
	"github.com/cyw0ng95/v2e/cmd/v2broker/transport"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
	subprocess "github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Spawner is a minimal interface for creating subprocesses. Implementations
// should delegate to existing broker spawn logic for this low-risk initial
// refactor.
type Spawner interface {
	Spawn(id, command string, args ...string) (*SpawnResult, error)
	SpawnRPC(id, command string, args ...string) (*SpawnResult, error)
	SpawnWithRestart(id, command string, maxRestarts int, restartDelay time.Duration, args ...string) (*SpawnResult, error)
	SpawnRPCWithRestart(id, command string, maxRestarts int, restartDelay time.Duration, args ...string) (*SpawnResult, error)
}

// SpawnResult is a lightweight DTO returned by Spawner implementations.
type SpawnResult struct {
	ID       string
	PID      int
	Command  string
	Args     []string
	Status   string
	ExitCode int
}

// Broker manages subprocesses and message passing.
type Broker struct {
	processes *sync.Map
	messages  chan *proc.Message
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	logger    *common.Logger

	bus             *mq.Bus
	metricsRegistry *metrics.Registry
	rpcEndpoints    map[string][]string
	endpointsMu     sync.RWMutex
	pendingRequests *sync.Map
	pendingMu       sync.RWMutex
	correlationSeq  uint64
	spawner         Spawner
	// optimizer optionally handles message routing asynchronously
	optimizer OptimizerInterface
	// transportManager manages communication transports for processes
	transportManager *transport.TransportManager
	// permitManager manages the global worker permit pool (Phase 2 UEE)
	permitManager *permits.PermitManager
}

// NewBroker creates a new Broker instance.
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	bus := mq.NewBus(ctx, 100)
	b := &Broker{
		processes:        &sync.Map{},
		messages:         bus.Channel(),
		ctx:              ctx,
		cancel:           cancel,
		logger:           common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
		bus:              bus,
		metricsRegistry:  metrics.NewRegistry(),
		rpcEndpoints:     make(map[string][]string),
		pendingRequests:  &sync.Map{},
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
	b.processes.Store(p.info.ID, p)
}

// ProcessCount returns the number of tracked processes.
func (b *Broker) ProcessCount() int {
	count := 0
	b.processes.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// PendingRequestCount returns the number of pending request entries.
func (b *Broker) PendingRequestCount() int {
	count := 0
	b.pendingRequests.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// AddPendingRequest registers a pending request entry. Intended for tests and benchmarks.
func (b *Broker) AddPendingRequest(correlationID string, pending *PendingRequest) {
	b.pendingRequests.Store(correlationID, pending)
}

// SetLogger sets the logger for the broker.
func (b *Broker) SetLogger(logger *common.Logger) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logger = logger
}

// SetSpawner injects a pluggable Spawner implementation.
func (b *Broker) SetSpawner(s Spawner) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spawner = s
}

// Spawner returns the currently configured Spawner (may be nil).
func (b *Broker) Spawner() Spawner {
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

// SetPermitManager attaches a permit manager to the Broker.
func (b *Broker) SetPermitManager(pm *permits.PermitManager) {
	b.mu.Lock()
	b.permitManager = pm
	b.mu.Unlock()
	if pm != nil && b.logger != nil {
		b.logger.Info("PermitManager attached")
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
	GetKernelMetrics() *perf.KernelMetrics
}
