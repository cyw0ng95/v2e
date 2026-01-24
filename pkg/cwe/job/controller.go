package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

var (
	ErrJobRunning    = errors.New("job is already running")
	ErrJobNotRunning = errors.New("job is not running")
)

// RPCInvoker is an interface for making RPC calls to other services
type RPCInvoker interface {
	InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

// Controller manages background jobs that fetch and persist CWE views
type Controller struct {
	rpcInvoker RPCInvoker
	logger     *common.Logger

	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	running    bool
	started    time.Time
}

// NewController creates a new CWE job controller
func NewController(rpcInvoker RPCInvoker, logger *common.Logger) *Controller {
	return &Controller{
		rpcInvoker: rpcInvoker,
		logger:     logger,
	}
}

// Start launches the background job to fetch and store CWE views.
// Returns a session ID string which is a timestamp-based identifier.
func (c *Controller) Start(ctx context.Context, params map[string]interface{}) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return "", ErrJobRunning
	}

	jobCtx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	c.running = true
	c.started = time.Now()

	// Launch background worker
	go c.runJob(jobCtx, params)

	sessionID := fmt.Sprintf("cwe-session-%d", time.Now().UnixNano())
	c.logger.Info("CWE view job started: session_id=%s", sessionID)
	return sessionID, nil
}

// Stop cancels the running job
func (c *Controller) Stop(ctx context.Context, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return ErrJobNotRunning
	}
	if c.cancelFunc != nil {
		c.cancelFunc()
		c.cancelFunc = nil
	}
	c.running = false
	c.logger.Info("CWE view job stopped: session_id=%s", sessionID)
	return nil
}

// Status returns a minimal status map for the job session
func (c *Controller) Status(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]interface{}{
		"running": c.running,
		"started": c.started.String(),
	}, nil
}

// IsRunning reports whether a job is active
func (c *Controller) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// runJob executes the fetch-and-store loop. It expects the remote service to
// provide `RPCFetchViews` which returns a subprocess.Message whose payload is
// JSON with a top-level `views` array.
func (c *Controller) runJob(ctx context.Context, params map[string]interface{}) {
	defer func() {
		c.mu.Lock()
		c.running = false
		c.cancelFunc = nil
		c.mu.Unlock()
	}()

	startIndex := 0
	pageSize := 100
	if v, ok := params["start_index"].(int); ok {
		startIndex = v
	}
	if v, ok := params["results_per_page"].(int); ok {
		pageSize = v
	}

	c.logger.Info("CWE job loop starting: start_index=%d, page_size=%d", startIndex, pageSize)

	current := startIndex

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("CWE job loop cancelled")
			return
		default:
			c.logger.Debug("Fetching views: start_index=%d, page_size=%d", current, pageSize)

			// Invoke remote to fetch views
			result, err := c.rpcInvoker.InvokeRPC(ctx, "remote", "RPCFetchViews", map[string]interface{}{
				"start_index":      current,
				"results_per_page": pageSize,
			})
			if err != nil {
				c.logger.Error("Failed to fetch views: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			msg, ok := result.(*subprocess.Message)
			if !ok {
				c.logger.Error("Invalid response type from remote for views")
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			if msg.Type == subprocess.MessageTypeError {
				c.logger.Error("Error from remote: %s", msg.Error)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Unmarshal payload into views
			var resp struct {
				Views []cwe.CWEView `json:"views"`
			}
			if err := jsonutil.Unmarshal(msg.Payload, &resp); err != nil {
				c.logger.Error("Failed to unmarshal views response: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			if len(resp.Views) == 0 {
				c.logger.Info("No more views to fetch. Job completed.")
				// stop gracefully
				c.mu.Lock()
				if c.cancelFunc != nil {
					c.cancelFunc()
					c.cancelFunc = nil
				}
				c.running = false
				c.mu.Unlock()
				return
			}

			stored := 0
			for _, v := range resp.Views {
				// Save via local RPC
				_, err := c.rpcInvoker.InvokeRPC(ctx, "local", "RPCSaveCWEView", v)
				if err != nil {
					c.logger.Error("Failed to save view %s: %v", v.ID, err)
				} else {
					stored++
				}
			}

			c.logger.Info("Fetched %d views and stored %d", len(resp.Views), stored)

			current += pageSize

			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}
}
