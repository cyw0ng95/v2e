package core

import (
	"context"
	"io"
	"sync"

	"github.com/cyw0ng95/v2e/cmd/broker/mq"
	"github.com/cyw0ng95/v2e/pkg/broker"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// Broker manages subprocesses and message passing.
type Broker struct {
	processes       map[string]*Process
	messages        chan *proc.Message
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	logger          *common.Logger
	config          *common.Config
	bus             *mq.Bus
	rpcEndpoints    map[string][]string
	endpointsMu     sync.RWMutex
	pendingRequests map[string]*PendingRequest
	pendingMu       sync.RWMutex
	correlationSeq  uint64
	spawner         broker.Spawner
}

// NewBroker creates a new Broker instance.
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	bus := mq.NewBus(ctx, 100)
	return &Broker{
		processes:       make(map[string]*Process),
		messages:        bus.Channel(),
		ctx:             ctx,
		cancel:          cancel,
		logger:          common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
		config:          nil,
		bus:             bus,
		rpcEndpoints:    make(map[string][]string),
		pendingRequests: make(map[string]*PendingRequest),
		correlationSeq:  0,
	}
}

// InsertProcessForTest inserts a pre-constructed process into the broker (testing only).
func (b *Broker) InsertProcessForTest(p *Process) {
	b.mu.Lock()
	b.processes[p.info.ID] = p
	b.mu.Unlock()
}

// StartProcessReaderForTest starts the stdout reader goroutine for a process (testing only).
func (b *Broker) StartProcessReaderForTest(p *Process) {
	b.wg.Add(1)
	go b.readProcessMessages(p)
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

// SetConfig sets the broker-level configuration used when spawning processes.
func (b *Broker) SetConfig(cfg *common.Config) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = cfg
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

// Context returns the broker's context.
func (b *Broker) Context() context.Context {
	return b.ctx
}
