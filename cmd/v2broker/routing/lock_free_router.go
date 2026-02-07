package routing

import (
	"fmt"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// LockFreeRouter provides lock-free message routing using sync.Map
type LockFreeRouter struct {
	routes *sync.Map // map[string]chan *proc.Message
}

// NewLockFreeRouter creates a new lock-free router
func NewLockFreeRouter() *LockFreeRouter {
	return &LockFreeRouter{
		routes: &sync.Map{},
	}
}

// Route performs lock-free message routing with timeout
func (r *LockFreeRouter) Route(msg *proc.Message) error {
	value, ok := r.routes.Load(msg.Target)
	if !ok {
		return fmt.Errorf("no route for target: %s", msg.Target)
	}

	ch := value.(chan *proc.Message)

	select {
	case ch <- msg:
		return nil
	default:
		return fmt.Errorf("route channel full for target: %s", msg.Target)
	}
}

// ProcessBrokerMessage routes broker control messages
func (r *LockFreeRouter) ProcessBrokerMessage(msg *proc.Message) error {
	value, ok := r.routes.Load(msg.Target)
	if !ok {
		return fmt.Errorf("no route for broker message target: %s", msg.Target)
	}

	ch := value.(chan *proc.Message)

	select {
	case ch <- msg:
		return nil
	default:
		return fmt.Errorf("broker route channel full for target: %s", msg.Target)
	}
}

// RegisterRoute registers a route for a target process
func (r *LockFreeRouter) RegisterRoute(target string, ch chan *proc.Message) {
	r.routes.Store(target, ch)
}

// UnregisterRoute removes a route for a target process
func (r *LockFreeRouter) UnregisterRoute(target string) {
	r.routes.Delete(target)
}

// ListRoutes returns all registered routes for debugging
func (r *LockFreeRouter) ListRoutes() []string {
	routes := make([]string, 0)
	r.routes.Range(func(key, value interface{}) bool {
		routes = append(routes, key.(string))
		return true
	})
	return routes
}

// GetRouteCount returns the number of registered routes
func (r *LockFreeRouter) GetRouteCount() int {
	count := 0
	r.routes.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
