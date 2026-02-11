package taskflow

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// LeakDetectorConfig configures leak detection
type LeakDetectorConfig struct {
	// Enabled enables leak detection
	Enabled bool
	// CheckInterval is how often to check for leaks
	CheckInterval time.Duration
	// MaxLifetime is max time before an object is considered leaked
	MaxLifetime time.Duration
	// Threshold is min number of objects before triggering leak warning
	Threshold int
}

// DefaultLeakDetectorConfig returns default configuration
func DefaultLeakDetectorConfig() LeakDetectorConfig {
	return LeakDetectorConfig{
		Enabled:       true,
		CheckInterval: 30 * time.Second,
		MaxLifetime:   5 * time.Minute,
		Threshold:     10,
	}
}

// ObjectRef tracks a pooled object reference
type ObjectRef struct {
	ID        string
	CreatedAt time.Time
	Stack     []uintptr
	Released  bool
	Tier      PoolSize
	Size      int
}

// LeakDetector tracks object lifecycle and detects leaks
type LeakDetector struct {
	config    LeakDetectorConfig
	tracked   map[string]*ObjectRef
	mu        sync.RWMutex
	started   bool
	stopChan  chan struct{}
	onLeak    func(leaked []*ObjectRef)
	stackPool sync.Pool
}

// NewLeakDetector creates a new leak detector
func NewLeakDetector(config LeakDetectorConfig) *LeakDetector {
	return &LeakDetector{
		config:   config,
		tracked:  make(map[string]*ObjectRef),
		stopChan: make(chan struct{}),
		stackPool: sync.Pool{
			New: func() interface{} {
				return make([]uintptr, 32)
			},
		},
	}
}

// NewLeakDetectorWithDefaults creates detector with default config
func NewLeakDetectorWithDefaults() *LeakDetector {
	return NewLeakDetector(DefaultLeakDetectorConfig())
}

// Track records an object allocation
func (ld *LeakDetector) Track(id string, tier PoolSize, size int) *ObjectRef {
	if !ld.config.Enabled {
		return nil
	}

	ld.mu.Lock()
	defer ld.mu.Unlock()

	ref := &ObjectRef{
		ID:        id,
		CreatedAt: time.Now(),
		Tier:      tier,
		Size:      size,
	}

	// Capture call stack for leak debugging
	stack := ld.stackPool.Get().([]uintptr)
	n := runtime.Callers(2, stack)
	ref.Stack = stack[:n:n]
	ld.stackPool.Put(stack)

	ld.tracked[id] = ref
	return ref
}

// Release marks an object as returned to pool
func (ld *LeakDetector) Release(id string) {
	if !ld.config.Enabled {
		return
	}

	ld.mu.Lock()
	defer ld.mu.Unlock()

	if ref, ok := ld.tracked[id]; ok {
		ref.Released = true
		// Don't delete immediately - keep for analysis
	}
}

// Start begins leak detection
func (ld *LeakDetector) Start() {
	if ld.config.Enabled && !ld.started {
		ld.started = true
		go ld.detectLeaks()
	}
}

// Stop stops leak detection
func (ld *LeakDetector) Stop() {
	if ld.started {
		ld.started = false
		close(ld.stopChan)
	}
}

// SetLeakCallback sets callback for leak notifications
func (ld *LeakDetector) SetLeakCallback(fn func(leaked []*ObjectRef)) {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.onLeak = fn
}

// detectLeaks periodically checks for leaked objects
func (ld *LeakDetector) detectLeaks() {
	ticker := time.NewTicker(ld.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ld.checkLeaks()
		case <-ld.stopChan:
			return
		}
	}
}

// checkLeaks scans for leaked objects
func (ld *LeakDetector) checkLeaks() {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	now := time.Now()
	var leaked []*ObjectRef

	for id, ref := range ld.tracked {
		age := now.Sub(ref.CreatedAt)

		// Consider unreleased objects older than MaxLifetime as leaks
		if !ref.Released && age > ld.config.MaxLifetime {
			leaked = append(leaked, ref)
		}

		// Clean up old released references (older than 2x MaxLifetime)
		if ref.Released && age > 2*ld.config.MaxLifetime {
			delete(ld.tracked, id)
		}
	}

	// Trigger leak notification if threshold exceeded
	if len(leaked) >= ld.config.Threshold && ld.onLeak != nil {
		ld.onLeak(leaked)
	}
}

// GetStats returns current tracking statistics
func (ld *LeakDetector) GetStats() map[string]interface{} {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	now := time.Now()
	active := 0
	leaked := 0
	leakedObjects := []*ObjectRef{}

	for _, ref := range ld.tracked {
		if !ref.Released {
			active++
			if now.Sub(ref.CreatedAt) > ld.config.MaxLifetime {
				leaked++
				leakedObjects = append(leakedObjects, ref)
			}
		}
	}

	return map[string]interface{}{
		"active_count":      active,
		"leaked_count":      leaked,
		"tracked_total":     len(ld.tracked),
		"leak_threshold":    ld.config.Threshold,
		"leaked_objects":    leakedObjects,
		"detection_enabled": ld.config.Enabled,
		"detection_running": ld.started,
	}
}

// GetStackTrace formats a stack trace for debugging
func (ld *LeakDetector) GetStackTrace(ref *ObjectRef) string {
	if ref == nil || len(ref.Stack) == 0 {
		return "no stack trace available"
	}

	frames := runtime.CallersFrames(ref.Stack)
	var buf []byte

	for {
		frame, more := frames.Next()
		buf = append(buf, fmt.Sprintf("%s\n    %s:%d\n", frame.Function, frame.File, frame.Line)...)

		if !more {
			break
		}
	}

	return string(buf)
}
