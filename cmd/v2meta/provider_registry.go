package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
)

// ProviderRegistry manages all data source providers
type ProviderRegistry struct {
	providers map[string]provider.DataSourceProvider
	logger    *common.Logger
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry(logger *common.Logger) *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]provider.DataSourceProvider),
		logger:    logger,
	}
}

// RegisterProvider registers a data source provider
func (r *ProviderRegistry) RegisterProvider(p provider.DataSourceProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.GetID()
	r.logger.Info("Registering provider: %s", id)
	r.providers[id] = p
	r.logger.Info("Provider registered successfully: %s", id)

	return nil
}

// GetProvider returns a provider by ID
func (r *ProviderRegistry) GetProvider(id string) (provider.DataSourceProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	return p, ok
}

// GetAllProviders returns all registered providers
func (r *ProviderRegistry) GetAllProviders() []provider.DataSourceProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]provider.DataSourceProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}

	return providers
}

// GetProviderStatus returns status of a provider
func (r *ProviderRegistry) GetProviderStatus(id string) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", id)
	}

	return map[string]interface{}{
		"id":       p.GetID(),
		"type":     p.GetType(),
		"state":    p.GetState(),
		"progress": p.GetProgress(),
	}, nil
}

// StartProvider starts a provider by ID
func (r *ProviderRegistry) StartProvider(ctx context.Context, id string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	return p.Start()
}

// PauseProvider pauses a provider by ID
func (r *ProviderRegistry) PauseProvider(ctx context.Context, id string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	return p.Pause()
}

// StopProvider stops a provider by ID
func (r *ProviderRegistry) StopProvider(ctx context.Context, id string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	return p.Stop()
}

// GetProviderConfigs returns configuration for all providers
func (r *ProviderRegistry) GetProviderConfigs() (map[string]*provider.ProviderConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configs := make(map[string]*provider.ProviderConfig)
	for id, p := range r.providers {
		configs[id] = p.GetConfig()
	}

	return configs, nil
}

// GetProviderStatuses returns status of all providers
func (r *ProviderRegistry) GetProviderStatuses() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]interface{})
	for id, p := range r.providers {
		status[id] = map[string]interface{}{
			"id":       p.GetID(),
			"type":     p.GetType(),
			"state":    p.GetState(),
			"progress": p.GetProgress(),
		}
	}

	return status
}

// StartAll starts all registered providers
func (r *ProviderRegistry) StartAll(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.providers {
		if err := p.Start(); err != nil {
			r.logger.Error("Failed to start provider %s: %v", p.GetID(), err)
			continue
		}
	}

	r.logger.Info("All providers started successfully")
	return nil
}

// StopAll stops all running providers
func (r *ProviderRegistry) StopAll(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.providers {
		if err := p.Stop(); err != nil {
			r.logger.Error("Failed to stop provider %s: %v", p.GetID(), err)
			continue
		}
	}

	r.logger.Info("All providers stopped successfully")
	return nil
}
