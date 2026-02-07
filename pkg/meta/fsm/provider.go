package fsm

import (
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
}

// ProviderConfig holds configuration for creating a provider FSM
type ProviderConfig struct {
	ID           string
	ProviderType string
	Storage      *storage.Store
	Executor     func() error // Custom execution logic
}

// NewBaseProviderFSM creates a new base provider FSM
func NewBaseProviderFSM(config ProviderConfig) (*BaseProviderFSM, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("provider ID cannot be empty")
	}
	if config.ProviderType == "" {
		return nil, fmt.Errorf("provider type cannot be empty")
	}

	p := &BaseProviderFSM{
		id:           config.ID,
		providerType: config.ProviderType,
		state:        ProviderIdle,
		storage:      config.Storage,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
		executor:     config.Executor,
	}

	// Try to load existing state from storage
	if config.Storage != nil {
		if err := p.loadState(); err != nil {
			// If no state exists, that's fine (new provider)
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

// GetState returns the current provider state
func (p *BaseProviderFSM) GetState() ProviderState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// Transition attempts to transition to a new state
func (p *BaseProviderFSM) Transition(newState ProviderState) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Validate transition
	if err := ValidateProviderTransition(p.state, newState); err != nil {
		return err
	}

	oldState := p.state
	p.state = newState
	p.updatedAt = time.Now()

	// Log FSM transition (Requirement 6: Log FSM Transitions)
	p.logTransition(oldState, newState, "manual")

	// Persist state to storage
	if p.storage != nil {
		providerState := providerStatePool.Get().(*storage.ProviderFSMState)
		providerState.ID = p.id
		providerState.ProviderType = p.providerType
		providerState.State = storage.ProviderState(p.state)
		providerState.LastCheckpoint = p.lastCheckpoint
		providerState.ProcessedCount = atomic.LoadInt64(&p.processedCount)
		providerState.ErrorCount = atomic.LoadInt64(&p.errorCount)
		providerState.CreatedAt = p.createdAt
		providerState.UpdatedAt = time.Now()

		if err := p.storage.SaveProviderState(providerState); err != nil {
			// Rollback state on persistence failure
			p.state = oldState
			providerStatePool.Put(providerState)
			return fmt.Errorf("failed to persist state transition: %w", err)
		}

		providerStatePool.Put(providerState)
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

		// Start execution
		go p.executeAsync()
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
			time.Sleep(retryAfter)
			// Transition back to ACQUIRING
			if p.GetState() == ProviderWaitingBackoff {
				p.Transition(ProviderAcquiring)
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
	}

	// Emit checkpoint event every 100 items
	if p.processedCount%100 == 0 {
		p.emitEvent(EventCheckpoint)
	}

	return nil
}

// GetStats returns provider statistics
func (p *BaseProviderFSM) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"id":              p.id,
		"provider_type":   p.providerType,
		"state":           string(p.state),
		"last_checkpoint": p.lastCheckpoint,
		"processed_count": p.processedCount,
		"error_count":     p.errorCount,
		"permits_held":    p.permitsHeld,
		"created_at":      p.createdAt.Format(time.RFC3339),
		"updated_at":      p.updatedAt.Format(time.RFC3339),
	}
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

	p.state = ProviderState(state.State)
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
