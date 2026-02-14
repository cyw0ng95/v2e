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

	bus                 *mq.Bus
	metricsRegistry     *metrics.Registry
	processMetricsStore *ProcessMetricsStore
	rpcEndpoints        map[string][]string
	endpointsMu         sync.RWMutex
	pendingRequests     *sync.Map
	pendingMu           sync.RWMutex
	correlationSeq      uint64
	spawner             Spawner
	// optimizer optionally handles message routing asynchronously
	optimizer OptimizerInterface
	// transportManager manages communication transports for processes
	transportManager *transport.TransportManager
	// permitManager manages the global worker permit pool (Phase 2 UEE)
	permitManager *permits.PermitManager
	// drainPeriod is the time to wait for in-flight requests to complete during shutdown
	drainPeriod time.Duration
	// healthCheckTicker triggers periodic health checks
	healthCheckTicker *time.Ticker
	// healthCheckWg waits for health check goroutine to finish
	healthCheckWg sync.WaitGroup
}

// NewBroker creates a new Broker instance.
func NewBroker() *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	bus := mq.NewBus(ctx, 100)
	b := &Broker{
		processes:           &sync.Map{},
		messages:            bus.Channel(),
		ctx:                 ctx,
		cancel:              cancel,
		logger:              common.NewLogger(io.Discard, "[BROKER] ", common.InfoLevel),
		bus:                 bus,
		metricsRegistry:     metrics.NewRegistry(),
		processMetricsStore: NewProcessMetricsStore(),
		rpcEndpoints:        make(map[string][]string),
		pendingRequests:     &sync.Map{},
		correlationSeq:      0,
		transportManager:    transport.NewTransportManager(),
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

// SetDrainPeriod sets the drain period for graceful shutdown.
// The drain period is the maximum time to wait for in-flight requests to complete
// before forcibly shutting down processes.
func (b *Broker) SetDrainPeriod(drainPeriod time.Duration) {
	b.mu.Lock()
	b.drainPeriod = drainPeriod
	b.mu.Unlock()
	if b.logger != nil {
		b.logger.Info("Drain period set to: %v", drainPeriod)
	}
}

// GetDrainPeriod returns the configured drain period.
func (b *Broker) GetDrainPeriod() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.drainPeriod
}

// HasPendingRequests returns true if there are any pending RPC requests being processed.
func (b *Broker) HasPendingRequests() bool {
	return b.PendingRequestCount() > 0
}

// StartHealthMonitoring starts the health monitoring goroutine that periodically
// checks the health of all subprocesses with health check enabled.
func (b *Broker) StartHealthMonitoring() {
	b.mu.Lock()
	if b.healthCheckTicker != nil {
		b.mu.Unlock()
		return // Already running
	}
	// Use a default interval of 30 seconds for health checks
	b.healthCheckTicker = time.NewTicker(30 * time.Second)
	b.mu.Unlock()

	b.healthCheckWg.Add(1)
	go b.healthCheckLoop()
	b.logger.Info("Health monitoring started")
}

// StopHealthMonitoring stops the health monitoring goroutine.
func (b *Broker) StopHealthMonitoring() {
	b.mu.Lock()
	if b.healthCheckTicker != nil {
		b.healthCheckTicker.Stop()
		b.healthCheckTicker = nil
	}
	b.mu.Unlock()
	b.healthCheckWg.Wait()
	b.logger.Info("Health monitoring stopped")
}

// healthCheckLoop runs periodic health checks on all subprocesses.
func (b *Broker) healthCheckLoop() {
	defer b.healthCheckWg.Done()

	for {
		var ticker *time.Ticker
		b.mu.Lock()
		if b.healthCheckTicker == nil {
			b.mu.Unlock()
			return
		}
		ticker = b.healthCheckTicker
		b.mu.Unlock()

		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.performHealthChecks()
		}
	}
}

// performHealthChecks checks the health of all subprocesses with health check enabled.
func (b *Broker) performHealthChecks() {
	b.processes.Range(func(key, value interface{}) bool {
		processID := key.(string)
		proc := value.(*Process)

		proc.mu.Lock()
		healthCheckEnabled := proc.restartConfig != nil && proc.restartConfig.HealthCheckEnabled
		proc.mu.Unlock()

		if healthCheckEnabled {
			healthy := b.performHealthCheck(processID)
			if !healthy {
				b.handleUnhealthyProcess(processID, proc)
			}
		}
		return true
	})
}

// performHealthCheck performs a health check on a single process.
// It returns true if the process is healthy, false otherwise.
func (b *Broker) performHealthCheck(processID string) bool {
	value, exists := b.processes.Load(processID)
	if !exists {
		return false
	}

	proc := value.(*Process)
	proc.mu.Lock()
	status := proc.info.Status
	pid := proc.info.PID
	// Note: healthCheckTimeout is captured for future use with ping-based health checks
	// Currently we only check process status and transport connectivity
	_ = proc.restartConfig != nil && proc.restartConfig.HealthCheckTimeout > 0
	proc.mu.Unlock()

	// Check if process is still running
	if status != ProcessStatusRunning {
		b.logger.Warn("Process %s (pid=%d) is not running (status=%s)", processID, pid, status)
		return false
	}

	// Check if process is still alive by looking it up
	// On Unix systems, we can send signal 0 to check if process is alive
	// For simplicity, we just check if the transport is still connected
	if b.transportManager != nil {
		if !b.transportManager.IsTransportConnected(processID) {
			b.logger.Warn("Process %s (pid=%d) transport is not connected", processID, pid)
			return false
		}
	}

	return true
}

// handleUnhealthyProcess handles a process that failed health check.
func (b *Broker) handleUnhealthyProcess(processID string, proc *Process) {
	proc.mu.Lock()
	if proc.restartConfig == nil {
		proc.mu.Unlock()
		return
	}

	proc.restartConfig.consecutiveFailures++
	threshold := proc.restartConfig.UnhealthyThreshold
	if threshold <= 0 {
		threshold = 3 // Default threshold
	}

	consecutiveFailures := proc.restartConfig.consecutiveFailures
	proc.mu.Unlock()

	b.logger.Warn("Process %s failed health check (%d/%d)", processID, consecutiveFailures, threshold)

	if consecutiveFailures >= threshold {
		b.logger.Error("Process %s exceeded unhealthy threshold, restarting", processID)
		// Reset consecutive failures counter
		proc.mu.Lock()
		if proc.restartConfig != nil {
			proc.restartConfig.consecutiveFailures = 0
		}
		proc.mu.Unlock()

		// Kill the process to trigger restart
		_ = b.Kill(processID)
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
