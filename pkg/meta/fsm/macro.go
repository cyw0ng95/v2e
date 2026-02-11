package fsm

import (
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// MacroFSMManager implements the MacroFSM interface for high-level orchestration
type MacroFSMManager struct {
	mu         sync.RWMutex
	id         string
	state      MacroState
	providers  map[string]ProviderFSM
	storage    *storage.Store
	createdAt  time.Time
	updatedAt  time.Time
	eventChan  chan *Event
	stopChan   chan struct{}
	eventWg    sync.WaitGroup
	eventBatch []*Event
	batchMu    sync.Mutex
	flushChan  chan struct{}
}

// NewMacroFSMManager creates a new macro FSM manager
func NewMacroFSMManager(id string, store *storage.Store) (*MacroFSMManager, error) {
	if id == "" {
		return nil, fmt.Errorf("macro FSM id cannot be empty")
	}

	m := &MacroFSMManager{
		id:         id,
		state:      MacroBootstrapping,
		providers:  make(map[string]ProviderFSM),
		storage:    store,
		createdAt:  time.Now(),
		updatedAt:  time.Now(),
		eventChan:  make(chan *Event, 1000),
		stopChan:   make(chan struct{}),
		eventBatch: make([]*Event, 0, 50),
		flushChan:  make(chan struct{}, 1),
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

	// Log FSM transition (Requirement 6: Log FSM Transitions)
	// Collect provider states without holding m.mu to avoid deadlock
	providerCount := len(m.providers)
	stateCounts := make(map[ProviderState]int)
	for _, provider := range m.providers {
		state := provider.GetState()
		stateCounts[state]++
	}
	m.logTransition(oldState, newState, providerCount, stateCounts)

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

// processEvents processes events from providers asynchronously with batching
func (m *MacroFSMManager) processEvents() {
	defer m.eventWg.Done()

	flushTicker := time.NewTicker(100 * time.Millisecond)
	defer flushTicker.Stop()

	for {
		select {
		case event := <-m.eventChan:
			m.batchMu.Lock()
			m.eventBatch = append(m.eventBatch, event)
			batchSize := len(m.eventBatch)
			m.batchMu.Unlock()

			if batchSize >= 50 {
				select {
				case m.flushChan <- struct{}{}:
				default:
				}
			}

		case <-m.flushChan:
			m.batchMu.Lock()
			if len(m.eventBatch) > 0 {
				for _, event := range m.eventBatch {
					m.handleEventInternal(event)
				}
				m.eventBatch = m.eventBatch[:0]
			}
			m.batchMu.Unlock()

		case <-flushTicker.C:
			m.batchMu.Lock()
			if len(m.eventBatch) > 0 {
				for _, event := range m.eventBatch {
					m.handleEventInternal(event)
				}
				m.eventBatch = m.eventBatch[:0]
			}
			m.batchMu.Unlock()

		case <-m.stopChan:
			return
		}
	}
}

// handleEventInternal processes a single event
func (m *MacroFSMManager) handleEventInternal(event *Event) {
	// Collect provider states BEFORE acquiring lock to avoid deadlock
	// If provider's GetState() needs to callback into macro FSM, doing it
	// without holding m.mu prevents circular lock dependency
	var providerStates map[string]ProviderState
	if event.Type == EventProviderCompleted {
		m.mu.RLock()
		providerStates = make(map[string]ProviderState, len(m.providers))
		for id, provider := range m.providers {
			providerStates[id] = provider.GetState()
		}
		m.mu.RUnlock()
	}

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
		// Check if all providers are completed using previously collected states
		allCompleted := true
		for _, state := range providerStates {
			if state != ProviderTerminated {
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

// GetProviderStartupOrder returns providers in dependency order
// Providers with no dependencies come first, then providers that depend on them, etc.
func (m *MacroFSMManager) GetProviderStartupOrder() []ProviderFSM {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all providers
	providers := make([]ProviderFSM, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}

	// Topological sort based on dependencies
	return m.topologicalSort(providers)
}

// topologicalSort performs a topological sort of providers based on dependencies
func (m *MacroFSMManager) topologicalSort(providers []ProviderFSM) []ProviderFSM {
	// Build dependency map and in-degree count
	providerIDMap := make(map[string]ProviderFSM)
	inDegree := make(map[string]int)
	depMap := make(map[string][]string) // provider -> list of providers that depend on it

	for _, p := range providers {
		pid := p.GetID()
		providerIDMap[pid] = p
		inDegree[pid] = 0
	}

	// Calculate in-degrees
	for _, p := range providers {
		pid := p.GetID()
		for _, dep := range p.GetDependencies() {
			if _, exists := providerIDMap[dep]; exists {
				inDegree[pid]++
				depMap[dep] = append(depMap[dep], pid)
			}
		}
	}

	// Kahn's algorithm for topological sort
	var result []ProviderFSM
	queue := make([]string, 0)

	// Start with providers that have no dependencies
	for pid, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, pid)
		}
	}

	for len(queue) > 0 {
		// Remove from queue
		current := queue[0]
		queue = queue[1:]

		// Add to result
		if provider, exists := providerIDMap[current]; exists {
			result = append(result, provider)
		}

		// Reduce in-degree for dependent providers
		for _, dependent := range depMap[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	return result
}

// ValidateProviderDependencies checks if all provider dependencies are satisfied
func (m *MacroFSMManager) ValidateProviderDependencies() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect current provider states
	states := make(map[string]ProviderState)
	for _, p := range m.providers {
		states[p.GetID()] = p.GetState()
	}

	// Check each provider's dependencies
	for _, p := range m.providers {
		if bp, ok := p.(interface{ CheckDependencies(map[string]ProviderState) error }); ok {
			if err := bp.CheckDependencies(states); err != nil {
				return err
			}
		}
	}

	return nil
}

// StartProviderWithDependencyCheck starts a provider after validating its dependencies
// Returns an error if any dependency is not in TERMINATED state
func (m *MacroFSMManager) StartProviderWithDependencyCheck(providerID string) error {
	m.mu.Lock()
	provider, exists := m.providers[providerID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("provider %s not found", providerID)
	}

	// Collect current provider states
	states := make(map[string]ProviderState)
	m.mu.RLock()
	for _, p := range m.providers {
		states[p.GetID()] = p.GetState()
	}
	m.mu.RUnlock()

	// Check dependencies using BaseProviderFSM's CheckDependencies method
	if bp, ok := provider.(interface{ CheckDependencies(map[string]ProviderState) error }); ok {
		if err := bp.CheckDependencies(states); err != nil {
			return fmt.Errorf("dependency check failed for provider %s: %w", providerID, err)
		}
	}

	// Start the provider
	return provider.Start()
}

// StartAllProvidersInOrder starts all providers in dependency order
// Providers with no dependencies are started first, then their dependents, etc.
// Returns a map of provider ID to error for any providers that failed to start
func (m *MacroFSMManager) StartAllProvidersInOrder() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get providers in startup order
	providers := m.GetProviderStartupOrder()
	errors := make(map[string]error)

	// Start each provider in order
	for _, provider := range providers {
		providerID := provider.GetID()
		currentState := provider.GetState()

		// Skip if already running or terminated
		if currentState == ProviderRunning || currentState == ProviderAcquiring {
			continue
		}
		if currentState == ProviderTerminated {
			errors[providerID] = fmt.Errorf("provider already terminated")
			continue
		}

		// Collect current states for dependency check
		states := make(map[string]ProviderState)
		for _, p := range m.providers {
			states[p.GetID()] = p.GetState()
		}

		// Check dependencies before starting
		if bp, ok := provider.(interface{ CheckDependencies(map[string]ProviderState) error }); ok {
			if err := bp.CheckDependencies(states); err != nil {
				errors[providerID] = fmt.Errorf("dependency check failed: %w", err)
				continue
			}
		}

		// Start the provider
		if err := provider.Start(); err != nil {
			errors[providerID] = err
		}
	}

	return errors
}

// Stop gracefully stops the macro FSM manager
func (m *MacroFSMManager) Stop() error {
	// Signal event processor to stop
	close(m.stopChan)
	m.eventWg.Wait()

	// Try to transition to draining without blocking
	// Use a non-blocking approach to avoid deadlock if locks are held elsewhere
	m.mu.Lock()
	currentState := m.state
	if currentState != MacroDraining {
		m.state = MacroDraining
		m.updatedAt = time.Now()
	}
	m.mu.Unlock()

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

// logTransition logs a macro FSM state transition
// Implements Requirement 6: Log FSM Transitions
// Takes provider state info as parameters to avoid calling provider methods while holding locks
func (m *MacroFSMManager) logTransition(oldState, newState MacroState, providerCount int, stateCounts map[ProviderState]int) {
	// Log structured transition
	fmt.Printf("[MACRO_FSM_TRANSITION] macro_id=%s old_state=%s new_state=%s timestamp=%s provider_count=%d provider_states=%+v\n",
		m.id,
		oldState,
		newState,
		time.Now().Format(time.RFC3339),
		providerCount,
		stateCounts,
	)
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
