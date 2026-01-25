package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

var (
	// ErrJobRunning indicates a job is already running
	ErrJobRunning = errors.New("job is already running")
	// ErrJobNotRunning indicates no job is running
	ErrJobNotRunning = errors.New("job is not running")
)

// Pooled RPC parameter structs to reduce allocations in hot loops
type fetchParams struct {
	StartIndex     int `json:"start_index"`
	ResultsPerPage int `json:"results_per_page"`
}

var fetchParamsPool = sync.Pool{
	New: func() interface{} { return &fetchParams{} },
}

// RPCInvoker is an interface for making RPC calls to other services
type RPCInvoker interface {
	InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

// Controller manages continuous fetch and store operations
type Controller struct {
	rpcInvoker     RPCInvoker
	sessionManager *session.Manager
	logger         *common.Logger

	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	running    bool
}

// NewController creates a new job controller
func NewController(rpcInvoker RPCInvoker, sessionManager *session.Manager, logger *common.Logger) *Controller {
	return &Controller{
		rpcInvoker:     rpcInvoker,
		sessionManager: sessionManager,
		logger:         logger,
		running:        false,
	}
}

// Start starts the continuous fetch and store job
func (c *Controller) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return ErrJobRunning
	}

	// Create cancellable context and mark running early to avoid races
	jobCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	c.running = true

	// Get current session
	sess, err := c.sessionManager.GetSession()
	if err != nil {
		// rollback
		c.cancelFunc()
		c.cancelFunc = nil
		c.running = false
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update session state to running
	err = c.sessionManager.UpdateState(session.StateRunning)
	if err != nil {
		// rollback
		c.cancelFunc()
		c.cancelFunc = nil
		c.running = false
		return fmt.Errorf("failed to update session state: %w", err)
	}

	// Start job in background
	go c.runJob(jobCtx, sess)

	c.logger.Info("Job started: session_id=%s, start_index=%d, batch_size=%d",
		sess.ID, sess.StartIndex, sess.ResultsPerBatch)

	return nil
}

// Stop stops the running job
func (c *Controller) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return ErrJobNotRunning
	}

	// Cancel the job
	if c.cancelFunc != nil {
		c.cancelFunc()
		c.cancelFunc = nil
	}

	c.running = false

	// Update session state
	err := c.sessionManager.UpdateState(session.StateStopped)
	if err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	c.logger.Info("Job stopped")

	return nil
}

// Pause pauses the running job
func (c *Controller) Pause() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return ErrJobNotRunning
	}

	// Cancel the current job
	if c.cancelFunc != nil {
		c.cancelFunc()
		c.cancelFunc = nil
	}

	c.running = false

	// Update session state
	err := c.sessionManager.UpdateState(session.StatePaused)
	if err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	c.logger.Info("Job paused")

	return nil
}

// Resume resumes a paused job
func (c *Controller) Resume(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return ErrJobRunning
	}

	// Get current session
	sess, err := c.sessionManager.GetSession()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is paused
	if sess.State != session.StatePaused {
		return fmt.Errorf("session is not paused (current state: %s)", sess.State)
	}

	// Update session state to running
	err = c.sessionManager.UpdateState(session.StateRunning)
	if err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	// Create cancellable context
	jobCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	c.running = true

	// Restart job in background
	go c.runJob(jobCtx, sess)

	c.logger.Info("Job resumed: session_id=%s", sess.ID)

	return nil
}

// IsRunning returns whether a job is currently running
func (c *Controller) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// runJob executes the continuous fetch and store loop
func (c *Controller) runJob(ctx context.Context, sess *session.Session) {
	defer func() {
		c.mu.Lock()
		c.running = false
		c.cancelFunc = nil
		c.mu.Unlock()
	}()

	currentIndex := sess.StartIndex
	batchSize := sess.ResultsPerBatch

	c.logger.Info("Job loop starting: start_index=%d, batch_size=%d", currentIndex, batchSize)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Job loop cancelled")
			return
		default:
			// Fetch batch from NVD via remote
			c.logger.Debug("Fetching batch: start_index=%d, batch_size=%d", currentIndex, batchSize)

			// Get pooled fetch params to avoid allocating a map each iteration
			fp := fetchParamsPool.Get().(*fetchParams)
			fp.StartIndex = currentIndex
			fp.ResultsPerPage = batchSize
			result, err := c.rpcInvoker.InvokeRPC(ctx, "remote", "RPCFetchCVEs", fp)
			// reset and return to pool
			fp.StartIndex = 0
			fp.ResultsPerPage = 0
			fetchParamsPool.Put(fp)

			if err != nil {
				c.logger.Warn("Failed to fetch CVEs: %v", err)
				if err := c.sessionManager.UpdateProgress(0, 0, 1); err != nil {
					c.logger.Warn("Failed to update progress: %v", err)
				}

				// Wait before retrying
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Parse the RPC response (it's a subprocess.Message)
			msg, ok := result.(*subprocess.Message)
			if !ok {
				c.logger.Warn("Invalid response type from remote")
				if err := c.sessionManager.UpdateProgress(0, 0, 1); err != nil {
					c.logger.Warn("Failed to update progress: %v", err)
				}
				continue
			}

			// Check if it's an error message
			if msg.Type == subprocess.MessageTypeError {
				c.logger.Warn("Error from remote: %s", msg.Error)
				if err := c.sessionManager.UpdateProgress(0, 0, 1); err != nil {
					c.logger.Warn("Failed to update progress: %v", err)
				}

				// Wait before retrying
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Parse the CVE response from payload
			var response cve.CVEResponse
			if err := jsonutil.Unmarshal(msg.Payload, &response); err != nil {
				c.logger.Warn("Failed to unmarshal CVE response: %v", err)
				if err := c.sessionManager.UpdateProgress(0, 0, 1); err != nil {
					c.logger.Warn("Failed to update progress: %v", err)
				}
				continue
			}

			// Check if we have any CVEs
			if len(response.Vulnerabilities) == 0 {
				c.logger.Info("No more CVEs to fetch. Job completed.")
				c.Stop()
				return
			}

			c.logger.Info("Fetched %d CVEs from NVD", len(response.Vulnerabilities))

			// Store each CVE via local
			storedCount := int64(0)
			errorCount := int64(0)

			for _, vuln := range response.Vulnerabilities {
				params := &rpc.SaveCVEByIDParams{CVE: vuln.CVE}
				_, err := c.rpcInvoker.InvokeRPC(ctx, "local", "RPCSaveCVEByID", params)

				if err != nil {
					c.logger.Error("Failed to store CVE %s: %v", vuln.CVE.ID, err)
					errorCount++
				} else {
					storedCount++
				}
			}

			c.logger.Info("Stored %d/%d CVEs successfully", storedCount, len(response.Vulnerabilities))

			// Update progress
			if err := c.sessionManager.UpdateProgress(int64(len(response.Vulnerabilities)), storedCount, errorCount); err != nil {
				c.logger.Warn("Failed to update progress: %v", err)
			}

			// Move to next batch
			currentIndex += batchSize

			// Wait a bit to respect rate limits
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				// Continue to next iteration
			}
		}
	}
}
