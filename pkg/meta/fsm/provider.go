package fsm

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

var (
	eventPool = sync.Pool{
		New: func() interface{} {
			return &Event{
				Data: make(map[string]interface{}, 4),
			}
		},
	}

	providerStatePool = sync.Pool{
		New: func() interface{} {
			return &storage.ProviderFSMState{}
		},
	}
)

// buildBaseStats builds common statistics map with id, state, and timestamps.
// This is a shared helper for both BaseProviderFSM and MacroFSMManager GetStats().
func buildBaseStats(id string, state string, createdAt, updatedAt time.Time) map[string]interface{} {
	return map[string]interface{}{
		"id":         id,
		"state":      state,
		"created_at": createdAt.Format(time.RFC3339),
		"updated_at": updatedAt.Format(time.RFC3339),
	}
}

// BaseProviderFSM provides a base implementation of ProviderFSM
// Concrete providers (CVE, CWE, etc.) can embed this and override Execute()
type BaseProviderFSM struct {
	mu             sync.RWMutex
	id             string
	providerType   string
	state          ProviderState
	storage        *storage.Store
	eventHandler   func(*Event) error
	createdAt      time.Time
	updatedAt      time.Time
	lastCheckpoint string
	processedCount int64
	errorCount     int64
	permitsHeld    int32
	executor       func() error
	eventQueue     chan *Event
	ctx            context.Context
	cancel         context.CancelFunc

	// Common configuration for all providers
	batchSize  int
	maxRetries int
	retryDelay time.Duration

	// Dependencies: list of provider IDs that must complete before this provider can start
	dependencies []string

	// Transition strategy for state transitions (reduces code duplication)
	transitionStrategy *ProviderTransitionStrategy
}

// ProviderConfig holds configuration for creating a provider FSM
type ProviderConfig struct {
	ID           string
	ProviderType string
	Storage      *storage.Store
	Executor     func() error // Custom execution logic

	// Dependencies: list of provider IDs that must complete before this provider can start
	Dependencies []string

	// Common configuration (defaults applied if zero)
	BatchSize  int           // Default: 100
	MaxRetries int           // Default: 3
	RetryDelay time.Duration // Default: 5 * time.Second
}

// NewBaseProviderFSM creates a new base provider FSM
func NewBaseProviderFSM(config ProviderConfig) (*BaseProviderFSM, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("provider ID cannot be empty")
	}
	if config.ProviderType == "" {
		return nil, fmt.Errorf("provider type cannot be empty")
	}

	// Apply defaults for common configuration
	batchSize := config.BatchSize
	if batchSize == 0 {
		batchSize = 100
	}
	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}
	retryDelay := config.RetryDelay
	if retryDelay == 0 {
		retryDelay = 5 * time.Second
	}

	p := &BaseProviderFSM{
		id:                 config.ID,
		providerType:       config.ProviderType,
		state:              ProviderIdle,
		storage:            config.Storage,
		createdAt:          time.Now(),
		updatedAt:          time.Now(),
		executor:           config.Executor,
		batchSize:          batchSize,
		maxRetries:         maxRetries,
		retryDelay:         retryDelay,
		dependencies:       config.Dependencies,
		transitionStrategy: NewProviderTransitionStrategy(),
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Try to load existing state from storage only if it exists
	if config.Storage != nil {
		if err := p.loadStateIfExists(); err != nil {
			// Ignore errors - state might not exist for new providers
		}
	}

	return p, nil
}

// GetID returns the provider ID
func (p *BaseProviderFSM) GetID() string {
	return p.id
}

// GetType returns the provider type
func (p *BaseProviderFSM) GetType() string {
	return p.providerType
}

// GetProgress returns to provider progress metrics
func (p *BaseProviderFSM) GetProgress() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"processedCount": p.processedCount,
		"errorCount":     p.errorCount,
	}
}

// GetStorage returns the storage instance
func (p *BaseProviderFSM) GetStorage() *storage.Store {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.storage
}

// GetState returns the current provider state
func (p *BaseProviderFSM) GetState() ProviderState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// GetBatchSize returns the batch size for processing
func (p *BaseProviderFSM) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// SetBatchSize sets the batch size for processing
func (p *BaseProviderFSM) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// GetMaxRetries returns the maximum retry count
func (p *BaseProviderFSM) GetMaxRetries() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.maxRetries
}

// SetMaxRetries sets the maximum retry count
func (p *BaseProviderFSM) SetMaxRetries(retries int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxRetries = retries
}

// GetRetryDelay returns the retry delay duration
func (p *BaseProviderFSM) GetRetryDelay() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.retryDelay
}

// SetRetryDelay sets the retry delay duration
func (p *BaseProviderFSM) SetRetryDelay(delay time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.retryDelay = delay
}

// Initialize sets up provider context before starting
func (p *BaseProviderFSM) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// BaseProviderFSM doesn't need context storage
	// Context management is handled by individual providers
	return nil
}

// GetStats returns provider statistics for monitoring
func (p *BaseProviderFSM) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := buildBaseStats(p.id, string(p.state), p.createdAt, p.updatedAt)
	stats["provider_type"] = p.providerType
	stats["last_checkpoint"] = p.lastCheckpoint
	stats["processed_count"] = atomic.LoadInt64(&p.processedCount)
	stats["error_count"] = atomic.LoadInt64(&p.errorCount)
	stats["permits_held"] = atomic.LoadInt32(&p.permitsHeld)
	return stats
}

// Transition attempts to transition to a new state
func (p *BaseProviderFSM) Transition(newState ProviderState) error {
	p.mu.Lock()

	oldState := p.state
	oldUpdatedAt := p.updatedAt

	// Validate transition using strategy
	if err := p.transitionStrategy.Validate(oldState, newState); err != nil {
		p.mu.Unlock()
		return err
	}

	// Update state
	p.state = newState
	now := time.Now()
	p.updatedAt = now

	// Log FSM transition (Requirement 6: Log FSM Transitions)
	p.logTransition(oldState, newState, "manual")

	// Create state for persistence (capture values while holding lock)
	providerState := providerStatePool.Get().(*storage.ProviderFSMState)
	providerState.ID = p.id
	providerState.ProviderType = p.providerType
	providerState.State = storage.ProviderState(newState)
	providerState.LastCheckpoint = p.lastCheckpoint
	providerState.ProcessedCount = atomic.LoadInt64(&p.processedCount)
	providerState.ErrorCount = atomic.LoadInt64(&p.errorCount)
	providerState.CreatedAt = p.createdAt
	providerState.UpdatedAt = now

	// Persist to storage
	if p.storage != nil {
		if err := p.storage.SaveProviderState(providerState); err != nil {
			// Rollback state and updatedAt on persistence failure
			p.state = oldState
			p.updatedAt = oldUpdatedAt
			providerStatePool.Put(providerState)
			p.mu.Unlock()
			return fmt.Errorf("failed to persist state transition: %w", err)
		}
	}

	providerStatePool.Put(providerState)
	p.mu.Unlock()
	return nil
}

// GetDependencies returns the list of provider IDs that must complete before this provider
func (p *BaseProviderFSM) GetDependencies() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to prevent external modification
	if p.dependencies == nil {
		return nil
	}
	deps := make([]string, len(p.dependencies))
	copy(deps, p.dependencies)
	return deps
}

// CheckDependencies verifies that dependency providers are in a valid state
// Valid states: IDLE (not started), ACQUIRING (starting), RUNNING (in progress), TERMINATED (completed)
// Invalid: WAITING_QUOTA, WAITING_BACKOFF, PAUSED
func (p *BaseProviderFSM) CheckDependencies(providerStates map[string]ProviderState) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.dependencies) == 0 {
		return nil
	}

	for _, depID := range p.dependencies {
		state, exists := providerStates[depID]
		if !exists {
			return fmt.Errorf("dependency provider %s not found", depID)
		}
		// Allow IDLE, ACQUIRING, RUNNING, TERMINATED - these are all valid states
		// Block on transient failure states: WAITING_QUOTA, WAITING_BACKOFF, PAUSED
		if state == ProviderWaitingQuota || state == ProviderWaitingBackoff || state == ProviderPaused {
			return fmt.Errorf("dependency provider %s is in invalid state %s (must be IDLE, ACQUIRING, RUNNING, or TERMINATED)", depID, state)
		}
	}

	return nil
}

// Start begins execution (IDLE -> ACQUIRING -> RUNNING)
func (p *BaseProviderFSM) Start() error {
	currentState := p.GetState()

	if currentState != ProviderIdle {
		return fmt.Errorf("cannot start from state %s, must be IDLE", currentState)
	}

	// Transition to ACQUIRING (waiting for permits)
	if err := p.Transition(ProviderAcquiring); err != nil {
		return err
	}

	// Emit event
	p.emitEvent(EventProviderStarted)

	return nil
}

// Pause pauses execution (RUNNING -> PAUSED)
func (p *BaseProviderFSM) Pause() error {
	currentState := p.GetState()

	if currentState != ProviderRunning {
		return fmt.Errorf("cannot pause from state %s, must be RUNNING", currentState)
	}

	if err := p.Transition(ProviderPaused); err != nil {
		return err
	}

	// Emit event
	p.emitEvent(EventProviderPaused)

	return nil
}

// Resume resumes execution (PAUSED -> ACQUIRING -> RUNNING)
func (p *BaseProviderFSM) Resume() error {
	currentState := p.GetState()

	if currentState != ProviderPaused {
		return fmt.Errorf("cannot resume from state %s, must be PAUSED", currentState)
	}

	// Transition back to ACQUIRING to request permits again
	if err := p.Transition(ProviderAcquiring); err != nil {
		return err
	}

	// Emit event
	p.emitEvent(EventProviderResumed)

	return nil
}

// Stop terminates execution (any state -> TERMINATED)
func (p *BaseProviderFSM) Stop() error {
	if err := p.Transition(ProviderTerminated); err != nil {
		return err
	}

	// Cancel context to stop any goroutines
	if p.cancel != nil {
		p.cancel()
	}

	// Emit event
	p.emitEvent(EventProviderCompleted)

	return nil
}

// OnQuotaRevoked handles quota revocation from broker
func (p *BaseProviderFSM) OnQuotaRevoked(revokedCount int) error {
	atomic.AddInt32(&p.permitsHeld, -int32(revokedCount))
	permits := atomic.LoadInt32(&p.permitsHeld)
	if permits < 0 {
		atomic.StoreInt32(&p.permitsHeld, 0)
	}

	currentState := p.GetState()

	// If running, transition to WAITING_QUOTA
	if currentState == ProviderRunning {
		if err := p.Transition(ProviderWaitingQuota); err != nil {
			return err
		}

		// Emit event
		p.emitEvent(EventQuotaRevoked)
	}

	return nil
}

// OnQuotaGranted handles quota grant from broker
func (p *BaseProviderFSM) OnQuotaGranted(grantedCount int) error {
	atomic.AddInt32(&p.permitsHeld, int32(grantedCount))

	currentState := p.GetState()

	// If acquiring, transition to RUNNING
	if currentState == ProviderAcquiring {
		if err := p.Transition(ProviderRunning); err != nil {
			return err
		}

		// Emit event
		p.emitEvent(EventQuotaGranted)

		// Start execution with context check
		go func() {
			select {
			case <-p.ctx.Done():
				return
			default:
				p.executeAsync()
			}
		}()
	} else if currentState == ProviderWaitingQuota {
		// Retry acquisition
		if err := p.Transition(ProviderAcquiring); err != nil {
			return err
		}
	}

	return nil
}

// OnRateLimited handles rate limiting (429 errors)
func (p *BaseProviderFSM) OnRateLimited(retryAfter time.Duration) error {
	currentState := p.GetState()

	if currentState == ProviderRunning {
		if err := p.Transition(ProviderWaitingBackoff); err != nil {
			return err
		}

		// Emit event
		p.emitEvent(EventRateLimited)

		// Schedule retry after backoff
		go func() {
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(retryAfter):
				// Transition back to ACQUIRING
				if p.GetState() == ProviderWaitingBackoff {
					p.Transition(ProviderAcquiring)
				}
			}
		}()
	}

	return nil
}

// Execute performs the actual work (called when in RUNNING state)
func (p *BaseProviderFSM) Execute() error {
	if p.executor != nil {
		return p.executor()
	}
	return fmt.Errorf("no executor defined")
}

// executeAsync runs the executor in a goroutine
func (p *BaseProviderFSM) executeAsync() {
	select {
	case <-p.ctx.Done():
		return
	default:
	}
	if err := p.Execute(); err != nil {
		p.mu.Lock()
		p.errorCount++
		p.mu.Unlock()

		// Emit failure event
		p.emitEvent(EventProviderFailed)
	}
}

// SetEventHandler sets the callback for event bubbling to MacroFSM
func (p *BaseProviderFSM) SetEventHandler(handler func(*Event) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventHandler = handler
}

// emitEvent sends an event to the event handler
func (p *BaseProviderFSM) emitEvent(eventType EventType) {
	p.mu.RLock()
	handler := p.eventHandler
	p.mu.RUnlock()

	if handler != nil {
		event := eventPool.Get().(*Event)
		event.Type = eventType
		event.ProviderID = p.id
		event.Timestamp = time.Now()
		clearEventData(event.Data)

		handler(event)

		eventPool.Put(event)
	}
}

func clearEventData(data map[string]interface{}) {
	for k := range data {
		delete(data, k)
	}
}

// SaveCheckpoint saves a checkpoint with URN
func (p *BaseProviderFSM) SaveCheckpoint(itemURN *urn.URN, success bool, errorMsg string) error {
	if itemURN == nil {
		return fmt.Errorf("URN cannot be nil")
	}

	p.mu.Lock()
	p.lastCheckpoint = itemURN.Key()
	p.processedCount++
	if !success {
		p.errorCount++
	}
	p.mu.Unlock()

	// Save to storage
	if p.storage != nil {
		checkpoint := &storage.Checkpoint{
			URN:          itemURN.Key(),
			ProviderID:   p.id,
			ProcessedAt:  time.Now(),
			Success:      success,
			ErrorMessage: errorMsg,
		}
		if err := p.storage.SaveCheckpoint(checkpoint); err != nil {
			return fmt.Errorf("failed to save checkpoint: %w", err)
		}

		// Persist updated stats to storage for recovery
		p.persistStats()
	}

	// Emit checkpoint event every 100 items
	if atomic.LoadInt64(&p.processedCount)%100 == 0 {
		p.emitEvent(EventCheckpoint)
	}

	return nil
}

// persistStats saves the current provider stats to storage without changing state
func (p *BaseProviderFSM) persistStats() {
	if p.storage == nil {
		return
	}

	providerState := providerStatePool.Get().(*storage.ProviderFSMState)
	providerState.ID = p.id
	providerState.ProviderType = p.providerType
	providerState.State = storage.ProviderState(p.state)
	providerState.LastCheckpoint = p.lastCheckpoint
	providerState.ProcessedCount = atomic.LoadInt64(&p.processedCount)
	providerState.ErrorCount = atomic.LoadInt64(&p.errorCount)
	providerState.CreatedAt = p.createdAt
	providerState.UpdatedAt = time.Now()

	_ = p.storage.SaveProviderState(providerState)

	providerStatePool.Put(providerState)
}

// loadState attempts to load persisted state from storage
func (p *BaseProviderFSM) loadState() error {
	if p.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	state, err := p.storage.GetProviderState(p.id)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.state = ProviderState(state.State)
	p.lastCheckpoint = state.LastCheckpoint
	p.processedCount = state.ProcessedCount
	p.errorCount = state.ErrorCount
	p.createdAt = state.CreatedAt
	p.updatedAt = state.UpdatedAt
	p.mu.Unlock()

	return nil
}

// LoadState explicitly loads the provider's state from storage
// This is used when recovering a provider after a restart
func (p *BaseProviderFSM) LoadState() error {
	return p.loadState()
}

// loadStateIfExists loads state only if it exists in storage
func (p *BaseProviderFSM) loadStateIfExists() error {
	if p.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	// Check if state exists first
	state, err := p.storage.GetProviderState(p.id)
	if err != nil {
		// State doesn't exist - this is fine for new providers
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = ProviderState(state.State)
	// If loaded state is a transient state (ACQUIRING, waiting states), reset to IDLE
	// since the operation didn't complete (system likely crashed mid-operation)
	if p.state == ProviderAcquiring || p.state == ProviderWaitingQuota ||
		p.state == ProviderWaitingBackoff {
		p.state = ProviderIdle
	}
	p.lastCheckpoint = state.LastCheckpoint
	p.processedCount = state.ProcessedCount
	p.errorCount = state.ErrorCount
	p.createdAt = state.CreatedAt
	p.updatedAt = state.UpdatedAt

	return nil
}

// logTransition logs an FSM state transition
// Implements Requirement 6: Log FSM Transitions
func (p *BaseProviderFSM) logTransition(oldState, newState ProviderState, trigger string) {
	// Get current checkpoint/URN if available
	checkpoint := p.lastCheckpoint
	if checkpoint == "" {
		checkpoint = "none"
	}

	// Log structured transition
	// Note: This uses fmt.Printf for now, but should use common.Logger when available
	// The logger can be passed via ProviderConfig if needed
	fmt.Printf("[FSM_TRANSITION] provider_id=%s provider_type=%s old_state=%s new_state=%s trigger=%s urn=%s timestamp=%s processed=%d errors=%d\n",
		p.id,
		p.providerType,
		oldState,
		newState,
		trigger,
		checkpoint,
		time.Now().Format(time.RFC3339),
		p.processedCount,
		p.errorCount,
	)
}
