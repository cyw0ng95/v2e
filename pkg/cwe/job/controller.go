package job

import (
	"context"
	"time"
)

// Controller is a stub for CWE view job controller. Full implementation is planned
// in a later tier. This controller exposes Start/Stop/Status methods and is
// designed to be invoked by the meta service via RPC. It intentionally does not
// spawn processes and relies on broker-mediated RPC for inter-process calls.
type Controller struct {
	running bool
	started time.Time
}

// NewController creates a new Controller instance (stub).
func NewController() *Controller {
	return &Controller{}
}

// Start begins the job controller session. Currently a no-op that records start time.
func (c *Controller) Start(ctx context.Context, params map[string]interface{}) (string, error) {
	c.running = true
	c.started = time.Now()
	// TODO: implement job orchestration, session persistence, and broker RPC invocations.
	return "stub-session-id", nil
}

// Stop stops a running session. Currently a no-op.
func (c *Controller) Stop(ctx context.Context, sessionID string) error {
	c.running = false
	return nil
}

// Status returns a minimal status description for the session.
func (c *Controller) Status(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"running": c.running,
		"started": c.started.String(),
	}, nil
}
