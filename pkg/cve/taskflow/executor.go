package taskflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	gotaskflow "github.com/noneback/go-taskflow"
)

// RPCInvoker is an interface for making RPC calls to other services
type RPCInvoker interface {
	InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

// JobExecutor manages task execution using go-taskflow with persistent state
type JobExecutor struct {
	rpcInvoker           RPCInvoker
	runStore             *RunStore
	executor             gotaskflow.Executor
	logger               *common.Logger
	remoteCircuitBreaker *CircuitBreaker
	localCircuitBreaker  *CircuitBreaker
	tieredPool           *TieredPool
	poolMetrics          *PoolMetrics

	mu         sync.RWMutex
	activeRun  *JobRun
	cancelFunc context.CancelFunc
	doneChan   chan struct{} // Signals when executeJob goroutine completes
}

// NewJobExecutor creates a new job executor with Taskflow and persistent storage
func NewJobExecutor(rpcInvoker RPCInvoker, runStore *RunStore, logger *common.Logger, concurrency uint) *JobExecutor {
	// Create circuit breakers: 5 failures triggers open, 60s to reset
	remoteCB := NewCircuitBreaker(5, 60*time.Second)
	localCB := NewCircuitBreaker(10, 30*time.Second)

	// Initialize tiered pool for message processing
	tp := NewTieredPoolWithDefaults()
	metrics := NewPoolMetrics()

	return &JobExecutor{
		rpcInvoker:           rpcInvoker,
		runStore:             runStore,
		executor:             gotaskflow.NewExecutor(concurrency),
		logger:               logger,
		remoteCircuitBreaker: remoteCB,
		localCircuitBreaker:  localCB,
		tieredPool:           tp,
		poolMetrics:          metrics,
	}
}

// Start starts a new CVE job run (enforces single active run)
func (e *JobExecutor) Start(ctx context.Context, runID string, startIndex, resultsPerBatch int) error {
	return e.StartTyped(ctx, runID, startIndex, resultsPerBatch, DataTypeCVE)
}

// StartTyped starts a new job run with a specific data type (enforces single active run)
func (e *JobExecutor) StartTyped(ctx context.Context, runID string, startIndex, resultsPerBatch int, dataType DataType) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// First check in-memory activeRun (faster)
	if e.activeRun != nil {
		return fmt.Errorf("job already running: %s (state: %s)", e.activeRun.ID, e.activeRun.State)
	}

	// Double-check persisted store for active runs
	activeRun, err := e.runStore.GetActiveRun()
	if err != nil {
		return fmt.Errorf("failed to check active run: %w", err)
	}
	if activeRun != nil {
		return fmt.Errorf("job already running: %s (state: %s)", activeRun.ID, activeRun.State)
	}

	// Create new run
	run, err := e.runStore.CreateRun(runID, startIndex, resultsPerBatch, dataType)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	// Transition to running with validation
	if err := e.transitionStateLocked(runID, StateQueued, StateRunning); err != nil {
		return fmt.Errorf("failed to transition to running: %w", err)
	}

	// Set activeRun and doneChan BEFORE starting goroutine (prevents race)
	e.activeRun = run
	e.doneChan = make(chan struct{})

	// Create cancellable context
	jobCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel

	// Start job in background (lock is released after defer, but activeRun is already set)
	go e.executeJob(jobCtx, runID)

	e.logger.Info(cve.LogMsgTFJobStarted,
		runID, startIndex, resultsPerBatch, dataType)

	return nil
}

// Resume resumes a paused job run
func (e *JobExecutor) Resume(ctx context.Context, runID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate no active run (prevents double-resume)
	if e.activeRun != nil {
		return fmt.Errorf("cannot resume: another job is active: %s", e.activeRun.ID)
	}

	// Get and validate the run
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	if run.State != StatePaused {
		return fmt.Errorf("run is not paused (current state: %s)", run.State)
	}

	// Transition with validation
	if err := e.transitionStateLocked(runID, StatePaused, StateRunning); err != nil {
		return err
	}

	// Set active run before starting goroutine
	e.activeRun = run
	e.doneChan = make(chan struct{})

	// Create cancellable context
	jobCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel

	// Start job in background
	go e.executeJob(jobCtx, runID)

	e.logger.Info(cve.LogMsgTFJobResumed, runID)

	return nil
}

// Pause pauses the running job
func (e *JobExecutor) Pause(runID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// First validate run exists in store
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	// Validate state
	if run.State != StateRunning {
		return fmt.Errorf("run is not running (current state: %s)", run.State)
	}

	// Then verify we own this run
	if e.activeRun == nil || e.activeRun.ID != runID {
		return fmt.Errorf("run not active: %s", runID)
	}

	// Cancel the job context
	if e.cancelFunc != nil {
		e.cancelFunc()
		e.cancelFunc = nil
	}

	// Transition with validation
	if err := e.transitionStateLocked(runID, StateRunning, StatePaused); err != nil {
		return err
	}

	// Wait for goroutine to finish (with timeout)
	select {
	case <-e.doneChan:
		// OK, goroutine finished
	case <-time.After(10 * time.Second):
		e.logger.Warn("Pause: goroutine did not finish within timeout")
	}

	e.activeRun = nil
	e.doneChan = nil
	e.logger.Info(cve.LogMsgTFJobPaused, runID)

	return nil
}

// Stop stops the running job
func (e *JobExecutor) Stop(runID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// First validate run exists in store
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	// Allow stopping from running or paused state
	if run.State != StateRunning && run.State != StatePaused {
		return fmt.Errorf("run cannot be stopped from state: %s", run.State)
	}

	// For running jobs, verify we own it and cancel
	if run.State == StateRunning {
		if e.activeRun == nil || e.activeRun.ID != runID {
			return fmt.Errorf("run not active: %s", runID)
		}

		// Cancel the job context
		if e.cancelFunc != nil {
			e.cancelFunc()
			e.cancelFunc = nil
		}

		// Wait for goroutine to finish
		select {
		case <-e.doneChan:
		case <-time.After(10 * time.Second):
			e.logger.Warn("Stop: goroutine did not finish within timeout")
		}
	}

	// Transition to stopped
	fromState := run.State
	if err := e.transitionStateLocked(runID, fromState, StateStopped); err != nil {
		return err
	}

	e.activeRun = nil
	e.doneChan = nil
	e.logger.Info(cve.LogMsgTFJobStopped, runID)

	return nil
}

// GetPoolStats returns pool utilization statistics
func (e *JobExecutor) GetPoolStats() map[string]interface{} {
	if e.tieredPool == nil || e.poolMetrics == nil {
		return nil
	}

	poolStats := e.tieredPool.GetStats()
	metricsStats := e.poolMetrics.GetUtilizationStats()

	return map[string]interface{}{
		"pool_stats":    poolStats,
		"metrics_stats": metricsStats,
	}
}

// GetStatus returns the current status of a run
func (e *JobExecutor) GetStatus(runID string) (*JobRun, error) {
	return e.runStore.GetRun(runID)
}

// transitionStateLocked performs validated state transition (caller must hold lock)
func (e *JobExecutor) transitionStateLocked(runID string, from, to JobState) error {
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		return err
	}

	if run.State != from {
		return fmt.Errorf("expected state %s, got %s", from, run.State)
	}

	if !run.State.CanTransitionTo(to) {
		return fmt.Errorf("invalid state transition: %s -> %s", run.State, to)
	}

	return e.runStore.UpdateState(runID, to)
}

// GetActiveRun returns the currently active run (if any)
func (e *JobExecutor) GetActiveRun() (*JobRun, error) {
	// First check in-memory active run to avoid races with persistence
	e.mu.RLock()
	if e.activeRun != nil {
		// attempt to return the persisted run so counters reflect latest updates
		runID := e.activeRun.ID
		e.mu.RUnlock()

		run, err := e.runStore.GetRun(runID)
		if err == nil && run != nil {
			return run, nil
		}

		// fall back to a copy of the in-memory run
		e.mu.RLock()
		runCopy := *e.activeRun
		e.mu.RUnlock()
		return &runCopy, nil
	}
	e.mu.RUnlock()

	// Fall back to persisted store
	return e.runStore.GetActiveRun()
}

// GetLatestRun returns the most recently updated run from the store
func (e *JobExecutor) GetLatestRun() (*JobRun, error) {
	return e.runStore.GetLatestRun()
}

// RecoverRuns attempts to recover runs left in running state after restart
func (e *JobExecutor) RecoverRuns(ctx context.Context) error {
	activeRun, err := e.runStore.GetActiveRun()
	if err != nil {
		return err
	}

	if activeRun == nil {
		e.logger.Info(cve.LogMsgTFNoActiveRuns)
		return nil
	}

	e.logger.Info(cve.LogMsgTFFoundRun, activeRun.ID, activeRun.State)

	// Only auto-recover running jobs (paused jobs stay paused)
	if activeRun.State == StateRunning {
		e.logger.Info(cve.LogMsgTFAutoRecover, activeRun.ID)
		return e.Resume(ctx, activeRun.ID)
	}

	e.logger.Info(cve.LogMsgTFManualResume, activeRun.State)
	return nil
}

// executeJob runs the actual fetch-and-store loop using Taskflow
func (e *JobExecutor) executeJob(ctx context.Context, runID string) {
	// Signal completion when done (Pause/Stop will wait for this)
	defer func() {
		e.mu.Lock()
		if e.doneChan != nil {
			close(e.doneChan)
		}
		e.mu.Unlock()
	}()

	// Get run details
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		e.logger.Error(cve.LogMsgTFFailedGetRun, err)
		e.runStore.SetError(runID, fmt.Sprintf("failed to get run: %v", err))
		return
	}

	currentIndex := run.StartIndex
	batchSize := run.ResultsPerBatch

	e.logger.Info(cve.LogMsgTFJobLoopStarting,
		runID, currentIndex, batchSize)

	// Create Taskflow DAG for fetch-and-store loop
	// Each iteration is a simple linear flow: fetch -> store
	for {
		select {
		case <-ctx.Done():
			e.logger.Info(cve.LogMsgTFJobLoopCancelled, runID)
			// Clear activeRun only on cancellation (Pause/Stop handle their own cleanup)
			e.mu.Lock()
			e.activeRun = nil
			e.cancelFunc = nil
			e.mu.Unlock()
			return
		default:
			tf := gotaskflow.NewTaskFlow(fmt.Sprintf("cve-batch-%d", currentIndex))

			var fetchedVulns []struct {
				CVE cve.CVEItem `json:"cve"`
			}
			var fetchErr error

			// Task 1: Fetch batch from remote
			fetchTask := tf.NewTask("fetch", func() {
				e.logger.Debug(cve.LogMsgTFFetchingBatch, runID, currentIndex, batchSize)

				params := &rpc.FetchCVEsParams{
					StartIndex:     currentIndex,
					ResultsPerPage: batchSize,
				}
				result, err := e.rpcInvoker.InvokeRPC(ctx, "remote", "RPCFetchCVEs", params)

				if err != nil {
					fetchErr = err
					return
				}

				// Parse the RPC response (it's a subprocess.Message)
				msg, ok := result.(*subprocess.Message)
				if !ok {
					fetchErr = fmt.Errorf("invalid response type from remote")
					return
				}

				// Check if it's an error message
				if msg.Type == subprocess.MessageTypeError {
					rpcErr := fmt.Errorf("error from remote: %s", msg.Error)
					// Check for rate limit - should be handled differently
					if isRateLimitError(rpcErr) {
						e.logger.Warn("Rate limit detected, will retry with backoff")
						fetchErr = rpcErr
						return
					}
					fetchErr = rpcErr
					return
				}

				// Parse the CVE response from payload
				var response cve.CVEResponse
				if err := jsonutil.Unmarshal(msg.Payload, &response); err != nil {
					fetchErr = fmt.Errorf("failed to unmarshal CVE response: %w", err)
					return
				}

				fetchedVulns = response.Vulnerabilities
			})

			// Task 2: Store batch to local
			storeTask := tf.NewTask("store", func() {
				if fetchErr != nil {
					e.logger.Warn(cve.LogMsgTFSkippingStore, fetchErr)
					return
				}

				if len(fetchedVulns) == 0 {
					e.logger.Info(cve.LogMsgTFNoMoreCVEs, runID)
					e.runStore.UpdateState(runID, StateCompleted)
					return
				}

				// Store each CVE with retry logic
				storedCount := int64(0)
				errorCount := int64(0)
				maxRetries := 3

				for _, vuln := range fetchedVulns {
					params := &rpc.SaveCVEByIDParams{CVE: vuln.CVE}
					var lastErr error

					// Retry failed saves up to maxRetries times
					for attempt := 0; attempt < maxRetries; attempt++ {
						_, err := e.rpcInvoker.InvokeRPC(ctx, "local", "RPCSaveCVEByID", params)

						if err == nil {
							storedCount++
							lastErr = nil
							break
						}

						lastErr = err
						if attempt < maxRetries-1 {
							// Exponential backoff before retry
							backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
							e.logger.Debug(cve.LogMsgTFFailedStoreCVE, vuln.CVE.ID, err)
							e.logger.Debug("Retrying save for %s after %v (attempt %d/%d)", vuln.CVE.ID, backoff, attempt+1, maxRetries)
							select {
							case <-ctx.Done():
								e.logger.Warn("Context cancelled while retrying save for %s", vuln.CVE.ID)
								break
							case <-time.After(backoff):
							}
						}
					}

					if lastErr != nil {
						e.logger.Warn(cve.LogMsgTFFailedStoreCVE, vuln.CVE.ID, lastErr)
						errorCount++
					}
				}

				e.logger.Info(cve.LogMsgTFStoredCVEsSuccess, storedCount, len(fetchedVulns))

				// Update progress
				e.runStore.UpdateProgress(runID, int64(len(fetchedVulns)), storedCount, errorCount)
			})

			// Define task dependency: fetch must complete before store
			fetchTask.Precede(storeTask)

			// Execute the taskflow
			e.executor.Run(tf).Wait()

			// Check if we should continue
			if fetchErr != nil {
				e.logger.Warn(cve.LogMsgTFFetchFailed, fetchErr)
				e.runStore.UpdateProgress(runID, 0, 0, 1)

				// Check if error is unrecoverable
				if shouldGiveUp(fetchErr) {
					e.logger.Error("Job failed after unrecoverable error: %v", fetchErr)
					e.runStore.UpdateState(runID, StateFailed)
					e.runStore.SetError(runID, fetchErr.Error())
					// Clear activeRun on failure
					e.mu.Lock()
					e.activeRun = nil
					e.cancelFunc = nil
					e.mu.Unlock()
					return
				}

				// Get error count for backoff calculation
				retryCount := 0
				if run, err := e.runStore.GetRun(runID); err == nil && run != nil {
					retryCount = int(run.ErrorCount)
				}

				// Calculate appropriate backoff based on error type
				backoff := calculateBackoff(fetchErr, retryCount)

				e.logger.Info("Retrying after %v (error: %v)", backoff, fetchErr)

				// Wait before retrying
				select {
				case <-ctx.Done():
					// Clear activeRun on cancellation during backoff wait
					e.mu.Lock()
					e.activeRun = nil
					e.cancelFunc = nil
					e.mu.Unlock()
					return
				case <-time.After(backoff):
					continue
				}
			}

			if len(fetchedVulns) == 0 {
				// Job completed naturally
				e.logger.Info(cve.LogMsgTFJobCompleted, runID)
				e.runStore.UpdateState(runID, StateCompleted)
				// Clear activeRun on completion
				e.mu.Lock()
				e.activeRun = nil
				e.cancelFunc = nil
				e.mu.Unlock()
				return
			}

			// Move to next batch
			currentIndex += batchSize

			// Rate limiting
			select {
			case <-ctx.Done():
				// Clear activeRun on cancellation during rate limit wait
				e.mu.Lock()
				e.activeRun = nil
				e.cancelFunc = nil
				e.mu.Unlock()
				return
			case <-time.After(1 * time.Second):
			}
		}
	}
}

// isRateLimitError checks if an error is related to API rate limiting
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "too many requests")
}

// shouldGiveUp determines if an error is unrecoverable and the job should fail
func shouldGiveUp(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Context cancellation is unrecoverable
	if strings.Contains(errStr, "context canceled") ||
		strings.Contains(errStr, "context deadline exceeded") {
		return true
	}
	return false
}

// calculateBackoff determines the appropriate wait time before retry
func calculateBackoff(err error, retryCount int) time.Duration {
	if err == nil {
		return 5 * time.Second
	}
	// Rate limits need longer backoff
	if isRateLimitError(err) {
		return 30 * time.Second
	}
	// Exponential backoff for other errors: 2^retryCount seconds, max 60s
	backoff := time.Duration(1<<uint(retryCount)) * time.Second
	if backoff > 60*time.Second {
		backoff = 60 * time.Second
	}
	return backoff
}
