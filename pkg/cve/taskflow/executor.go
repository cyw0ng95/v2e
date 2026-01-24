package taskflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	gotaskflow "github.com/noneback/go-taskflow"
)

// RPCInvoker is an interface for making RPC calls to other services
type RPCInvoker interface {
	InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

// JobExecutor manages task execution using go-taskflow with persistent state
type JobExecutor struct {
	rpcInvoker RPCInvoker
	runStore   *RunStore
	executor   gotaskflow.Executor
	logger     *common.Logger

	mu         sync.RWMutex
	activeRun  *JobRun
	cancelFunc context.CancelFunc
}

// NewJobExecutor creates a new job executor with Taskflow and persistent storage
func NewJobExecutor(rpcInvoker RPCInvoker, runStore *RunStore, logger *common.Logger, concurrency uint) *JobExecutor {
	return &JobExecutor{
		rpcInvoker: rpcInvoker,
		runStore:   runStore,
		executor:   gotaskflow.NewExecutor(concurrency),
		logger:     logger,
	}
}

// Start starts a new job run (enforces single active run)
func (e *JobExecutor) Start(ctx context.Context, runID string, startIndex, resultsPerBatch int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if there's already an active run
	activeRun, err := e.runStore.GetActiveRun()
	if err != nil {
		return fmt.Errorf("failed to check active run: %w", err)
	}
	if activeRun != nil {
		return fmt.Errorf("job already running: %s (state: %s)", activeRun.ID, activeRun.State)
	}

	// Create new run
	run, err := e.runStore.CreateRun(runID, startIndex, resultsPerBatch)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	// Transition to running
	if err := e.runStore.UpdateState(runID, StateRunning); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Create cancellable context
	jobCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel
	e.activeRun = run

	// Start job in background
	go e.executeJob(jobCtx, runID)

	e.logger.Info("Job started: run_id=%s, start_index=%d, batch_size=%d",
		runID, startIndex, resultsPerBatch)

	return nil
}

// Resume resumes a paused job run
func (e *JobExecutor) Resume(ctx context.Context, runID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get the run
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	// Check if it's paused
	if run.State != StatePaused {
		return fmt.Errorf("run is not paused (current state: %s)", run.State)
	}

	// Transition to running
	if err := e.runStore.UpdateState(runID, StateRunning); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	// Create cancellable context
	jobCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel
	e.activeRun = run

	// Restart job in background
	go e.executeJob(jobCtx, runID)

	e.logger.Info("Job resumed: run_id=%s", runID)

	return nil
}

// Pause pauses the running job
func (e *JobExecutor) Pause(runID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.activeRun == nil || e.activeRun.ID != runID {
		return fmt.Errorf("run not active: %s", runID)
	}

	// Get current run
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		return err
	}

	if run.State != StateRunning {
		return fmt.Errorf("run is not running (current state: %s)", run.State)
	}

	// Cancel the job context
	if e.cancelFunc != nil {
		e.cancelFunc()
		e.cancelFunc = nil
	}

	// Update state
	if err := e.runStore.UpdateState(runID, StatePaused); err != nil {
		return err
	}

	e.activeRun = nil
	e.logger.Info("Job paused: run_id=%s", runID)

	return nil
}

// Stop stops the running job
func (e *JobExecutor) Stop(runID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.activeRun == nil || e.activeRun.ID != runID {
		return fmt.Errorf("run not active: %s", runID)
	}

	// Cancel the job context
	if e.cancelFunc != nil {
		e.cancelFunc()
		e.cancelFunc = nil
	}

	// Update state
	if err := e.runStore.UpdateState(runID, StateStopped); err != nil {
		return err
	}

	e.activeRun = nil
	e.logger.Info("Job stopped: run_id=%s", runID)

	return nil
}

// GetStatus returns the current status of a run
func (e *JobExecutor) GetStatus(runID string) (*JobRun, error) {
	return e.runStore.GetRun(runID)
}

// GetActiveRun returns the currently active run (if any)
func (e *JobExecutor) GetActiveRun() (*JobRun, error) {
	return e.runStore.GetActiveRun()
}

// RecoverRuns attempts to recover runs left in running state after restart
func (e *JobExecutor) RecoverRuns(ctx context.Context) error {
	activeRun, err := e.runStore.GetActiveRun()
	if err != nil {
		return err
	}

	if activeRun == nil {
		e.logger.Info("No active runs to recover")
		return nil
	}

	e.logger.Info("Found run to recover: id=%s, state=%s", activeRun.ID, activeRun.State)

	// Only auto-recover running jobs (paused jobs stay paused)
	if activeRun.State == StateRunning {
		e.logger.Info("Auto-recovering running job: %s", activeRun.ID)
		return e.Resume(ctx, activeRun.ID)
	}

	e.logger.Info("Run is %s - manual resume required", activeRun.State)
	return nil
}

// executeJob runs the actual fetch-and-store loop using Taskflow
func (e *JobExecutor) executeJob(ctx context.Context, runID string) {
	defer func() {
		e.mu.Lock()
		e.activeRun = nil
		e.cancelFunc = nil
		e.mu.Unlock()
	}()

	// Get run details
	run, err := e.runStore.GetRun(runID)
	if err != nil {
		e.logger.Error("Failed to get run: %v", err)
		e.runStore.SetError(runID, fmt.Sprintf("failed to get run: %v", err))
		return
	}

	currentIndex := run.StartIndex
	batchSize := run.ResultsPerBatch

	e.logger.Info("Job loop starting: run_id=%s, start_index=%d, batch_size=%d",
		runID, currentIndex, batchSize)

	// Create Taskflow DAG for fetch-and-store loop
	// Each iteration is a simple linear flow: fetch -> store
	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Job loop cancelled: run_id=%s", runID)
			return
		default:
			tf := gotaskflow.NewTaskFlow(fmt.Sprintf("cve-batch-%d", currentIndex))

			var fetchedVulns []struct {
				CVE cve.CVEItem `json:"cve"`
			}
			var fetchErr error

			// Task 1: Fetch batch from remote
			fetchTask := tf.NewTask("fetch", func() {
				e.logger.Debug("Fetching batch: run_id=%s, index=%d, size=%d", runID, currentIndex, batchSize)

				result, err := e.rpcInvoker.InvokeRPC(ctx, "remote", "RPCFetchCVEs", map[string]interface{}{
					"start_index":      currentIndex,
					"results_per_page": batchSize,
				})

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
					fetchErr = fmt.Errorf("error from remote: %s", msg.Error)
					return
				}

				// Parse the CVE response from payload
				var response cve.CVEResponse
				if err := sonic.Unmarshal(msg.Payload, &response); err != nil {
					fetchErr = fmt.Errorf("failed to unmarshal CVE response: %w", err)
					return
				}

				fetchedVulns = response.Vulnerabilities
			})

			// Task 2: Store batch to local
			storeTask := tf.NewTask("store", func() {
				if fetchErr != nil {
					e.logger.Error("Skipping store due to fetch error: %v", fetchErr)
					return
				}

				if len(fetchedVulns) == 0 {
					e.logger.Info("No more CVEs to fetch. Job completed: run_id=%s", runID)
					e.runStore.UpdateState(runID, StateCompleted)
					return
				}

				// Store each CVE
				storedCount := int64(0)
				errorCount := int64(0)

				for _, vuln := range fetchedVulns {
					_, err := e.rpcInvoker.InvokeRPC(ctx, "local", "RPCSaveCVEByID", map[string]interface{}{
						"cve": vuln.CVE,
					})

					if err != nil {
						e.logger.Error("Failed to store CVE %s: %v", vuln.CVE.ID, err)
						errorCount++
					} else {
						storedCount++
					}
				}

				e.logger.Info("Stored %d/%d CVEs successfully", storedCount, len(fetchedVulns))

				// Update progress
				e.runStore.UpdateProgress(runID, int64(len(fetchedVulns)), storedCount, errorCount)
			})

			// Define task dependency: fetch must complete before store
			fetchTask.Precede(storeTask)

			// Execute the taskflow
			e.executor.Run(tf).Wait()

			// Check if we should continue
			if fetchErr != nil {
				e.logger.Error("Fetch failed: %v", fetchErr)
				e.runStore.UpdateProgress(runID, 0, 0, 1)

				// Wait before retrying
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			if len(fetchedVulns) == 0 {
				// Job completed naturally
				e.logger.Info("Job completed: run_id=%s", runID)
				e.runStore.UpdateState(runID, StateCompleted)
				return
			}

			// Move to next batch
			currentIndex += batchSize

			// Rate limiting
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}
}
