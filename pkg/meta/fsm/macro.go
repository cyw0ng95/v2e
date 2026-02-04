package fsm

import (
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// MacroFSMManager implements the MacroFSM interface for high-level orchestration
type MacroFSMManager struct {
	mu          sync.RWMutex
	id          string
	state       MacroState
	providers   map[string]ProviderFSM
	storage     *storage.Store
	createdAt   time.Time
	updatedAt   time.Time
	eventChan   chan *Event
	stopChan    chan struct{}
	eventWg     sync.WaitGroup
}

// NewMacroFSMManager creates a new macro FSM manager
func NewMacroFSMManager(id string, store *storage.Store) (*MacroFSMManager, error) {
	if id == "" {
		return nil, fmt.Errorf("macro FSM id cannot be empty")
	}

	m := &MacroFSMManager{
		id:        id,
		state:     MacroBootstrapping,
		providers: make(map[string]ProviderFSM),
		storage:   store,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		eventChan: make(chan *Event, 100),
		stopChan:  make(chan struct{}),
	}

	// Try to load existing state from storage
	if store != nil {
		if err := m.loadState(); err != nil {
			// If no state exists, that's fine (new FSM)
			// Just log and continue with BOOTSTRAPPING state
		}
	}

	// Start event processing goroutine
	m.eventWg.Add(1)
	go m.processEvents()

	return m, nil
}

// GetState returns the current macro state
func (m *MacroFSMManager) GetState() MacroState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// Transition attempts to transition to a new state
func (m *MacroFSMManager) Transition(newState MacroState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate transition
	if err := ValidateMacroTransition(m.state, newState); err != nil {
		return err
	}

	oldState := m.state
	m.state = newState
	m.updatedAt = time.Now()

	// Persist state to storage
	if m.storage != nil {
		macroState := &storage.MacroFSMState{
			ID:        m.id,
			State:     storage.MacroState(m.state),
			CreatedAt: m.createdAt,
			UpdatedAt: m.updatedAt,
		}
		if err := m.storage.SaveMacroState(macroState); err != nil {
			// Rollback state on persistence failure
			m.state = oldState
			return fmt.Errorf("failed to persist state transition: %w", err)
		}
	}

	return nil
}

// HandleEvent processes an event from a provider FSM
func (m *MacroFSMManager) HandleEvent(event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Queue event for async processing
	select {
	case m.eventChan <- event:
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("event queue full, event dropped")
	}
}

// processEvents processes events from providers asynchronously
func (m *MacroFSMManager) processEvents() {
	defer m.eventWg.Done()

	for {
		select {
		case event := <-m.eventChan:
			m.handleEventInternal(event)
		case <-m.stopChan:
			return
		}
	}
}

// handleEventInternal processes a single event
func (m *MacroFSMManager) handleEventInternal(event *Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch event.Type {
	case EventProviderStarted:
		// Provider has started, might transition to ORCHESTRATING if in BOOTSTRAPPING
		if m.state == MacroBootstrapping {
			// Auto-transition to orchestrating when first provider starts
			m.state = MacroOrchestrating
			m.updatedAt = time.Now()
		}

	case EventProviderCompleted:
		// Check if all providers are completed
		allCompleted := true
		for _, provider := range m.providers {
			if provider.GetState() != ProviderTerminated {
				allCompleted = false
				break
			}
		}
		if allCompleted && m.state == MacroOrchestrating {
			// Transition to stabilizing when all providers complete
			m.state = MacroStabilizing
			m.updatedAt = time.Now()
		}

	case EventProviderFailed:
		// Could implement error handling logic here
		// For now, just track the failure

	case EventCheckpoint:
		// Provider reached a checkpoint, no macro state change needed
	}
}

// GetProviders returns all managed provider FSMs
func (m *MacroFSMManager) GetProviders() []ProviderFSM {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]ProviderFSM, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	return providers
}

// AddProvider adds a provider FSM to be managed
func (m *MacroFSMManager) AddProvider(provider ProviderFSM) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	providerID := provider.GetID()
	if _, exists := m.providers[providerID]; exists {
		return fmt.Errorf("provider %s already exists", providerID)
	}

	m.providers[providerID] = provider

	// Set event handler to bubble events to macro FSM
	provider.SetEventHandler(func(event *Event) error {
		return m.HandleEvent(event)
	})

	return nil
}

// RemoveProvider removes a provider FSM
func (m *MacroFSMManager) RemoveProvider(providerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[providerID]; !exists {
		return fmt.Errorf("provider %s not found", providerID)
	}

	delete(m.providers, providerID)
	return nil
}

// GetProvider returns a specific provider by ID
func (m *MacroFSMManager) GetProvider(providerID string) (ProviderFSM, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[providerID]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerID)
	}

	return provider, nil
}

// GetID returns the macro FSM ID
func (m *MacroFSMManager) GetID() string {
	return m.id
}

// GetProviderCount returns the number of managed providers
func (m *MacroFSMManager) GetProviderCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.providers)
}

// Stop gracefully stops the macro FSM manager
func (m *MacroFSMManager) Stop() error {
	close(m.stopChan)
	m.eventWg.Wait()

	// Transition to draining if not already
	if m.state != MacroDraining {
		return m.Transition(MacroDraining)
	}

	return nil
}

// loadState attempts to load persisted state from storage
func (m *MacroFSMManager) loadState() error {
	if m.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	state, err := m.storage.GetMacroState(m.id)
	if err != nil {
		return err
	}

	m.state = MacroState(state.State)
	m.createdAt = state.CreatedAt
	m.updatedAt = state.UpdatedAt

	return nil
}

// GetStats returns statistics about the macro FSM
func (m *MacroFSMManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"id":              m.id,
		"state":           string(m.state),
		"provider_count":  len(m.providers),
		"created_at":      m.createdAt.Format(time.RFC3339),
		"updated_at":      m.updatedAt.Format(time.RFC3339),
		"event_queue_len": len(m.eventChan),
	}

	// Count providers by state
	stateCount := make(map[string]int)
	for _, provider := range m.providers {
		state := string(provider.GetState())
		stateCount[state]++
	}
	stats["provider_states"] = stateCount

	return stats
}
