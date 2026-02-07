package provider

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/cwe/provider"
	"github.com/cyw0ng95/v2e/pkg/capec/provider"
	"github.com/cyw0ng95/v2e/pkg/attack/provider"
	"github.com/cyw0ng95/v2e/pkg/ssg/provider"
	"github.com/cyw0ng95/v2e/pkg/asvs/provider"
)

	"github.com/cyw0ng95/v2e/pkg/meta/provider"
)

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry(storage *provider.Storage, logger *common.Logger) (*ProviderRegistry, error) {
	providers := make(map[string]provider.DataSourceProvider)

	// CVE Provider
	cveProvider, err := cve.NewCVEProvider(storage)
	if err != nil {
		return nil, err
	}
	providers["CVE"] = cveProvider

	// CWE Provider
	cweProvider, err := cwe.NewCWEProvider(storage)
	if err != nil {
		return nil, err
	}
	providers["CWE"] = cweProvider

	// CAPEC Provider
	capecProvider, err := capec.NewCAPECProvider(storage)
	if err != nil {
		return nil, err
	}
	providers["CAPEC"] = capecProvider

	// ATT&CK Provider
	attackProvider, err := attack.NewATTACKProvider(storage)
	if err != nil {
		return nil, err
	}
	providers["ATT&CK"] = attackProvider

	// SSG Provider
	ssgProvider, err := ssg.NewSSGProvider(storage)
	if err != nil {
		return nil, err
	}

	// ASVS Provider
	asvsProvider, err := asvs.NewASVSProvider(storage)
	if err != nil {
		return nil, err
	}

	return &ProviderRegistry{
		providers: providers,
		storage:  storage,
		logger:   logger,
	}, nil
}

// ProviderRegistry manages all data source providers
type ProviderRegistry struct {
	providers map[string]DataSourceProvider
	storage   *provider.Storage
	logger    *common.Logger
	mu       sync.RWMutex
}

// GetProvider returns a provider by ID
func (r *ProviderRegistry) GetProvider(id string) (DataSourceProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[id]
	if !ok {
		return nil, false
	}

	return provider, true
}

// GetAllProviders returns all registered providers
func (r *ProviderRegistry) GetAllProviders() []DataSourceProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]DataSourceProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}

	return providers
}

// StartProvider starts a provider by ID
func (r *ProviderRegistry) StartProvider(ctx context.Context, id string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	return provider.Start()
}

// PauseProvider pauses a provider by ID
func (r *ProviderRegistry) PauseProvider(ctx context.Context, id string) error {
	r.mu.RLock()
	defer r.mu.Unlock()

	provider, ok := r.providers[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	return provider.Pause()
}

// StopProvider stops a provider by ID
func (r *ProviderRegistry) StopProvider(ctx context.Context, id string) error {
	r.mu.RLock()
	defer r.mu.Unlock()

	provider, ok := r.providers[id]
	if !ok {
		return fmt.Errorf("provider not found: %s", id)
	}

	return provider.Stop()
}

// GetProviderStatus returns status of a provider
func (r *ProviderRegistry) GetProviderStatus(id string) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.Unlock()

	provider, ok := r.providers[id]
	if !ok {
		return nil, false, fmt.Errorf("provider not found: %s", id)
	}

	return map[string]interface{}{
		"id":     provider.GetID(),
		"type":     provider.GetType(),
		"state":    provider.GetState(),
		"progress":  provider.GetProgress(),
	}, nil
}

// GetProviderConfigs returns configuration for all providers
func (r *ProviderRegistry) GetProviderConfigs() (map[string]*provider.ProviderConfig, error) {
	r.mu.RLock()
	defer r.mu.Unlock()

	configs := make(map[string]*provider.ProviderConfig)
	for id, provider := range r.providers {
		configs[id] = provider.GetConfig()
	}

	return configs, nil
}
}

	registry, err := provider.NewProviderRegistry(nil, logger)
	if err != nil {
		return err
	}

	for _, p := range providers {
		registry.RegisterProvider(p)
	}

	logger.Info("All providers initialized successfully")
	return nil
}

// ProviderRegistry manages all data source providers
type ProviderRegistry struct {
	providers map[string]provider.DataSourceProvider
	storage   *provider.Storage
	logger    *common.Logger
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry(storage *provider.Storage, logger *common.Logger) *ProviderRegistry {
	return &ProviderRegistry{
			providers: make(map[string]provider.DataSourceProvider),
		storage:  storage,
		logger:    logger,
	}
}

// RegisterProvider registers a data source provider
func (r *ProviderRegistry) RegisterProvider(p provider provider.DataSourceProvider) error {
	id := provider.GetID()
	r.logger.Info("Registering provider: %s", id)

	r.providers[id] = provider
	r.logger.Info("Provider registered successfully: %s", id)

	return nil
}

// GetProvider returns a provider by ID
func (r *ProviderRegistry) GetProvider(id string) (provider.DataSourceProvider, bool) {
	provider, ok := r.providers[id]
	return provider, ok
}

// GetAllProviders returns all registered providers
func (r *ProviderRegistry) GetAllProviders() []provider.DataSourceProvider {
	providers := make([]provider.DataSourceProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// StartAll starts all registered providers
func (r *ProviderRegistry) StartAll(ctx context.Context) error {
	for _, provider := range r.providers {
		if err := provider.Start(); err != nil {
			r.logger.Error("Failed to start provider %s: %w", provider.GetID(), err)
			continue
		}
	}

	r.logger.Info("All providers started successfully")
	return nil
}

// StopAll stops all running providers
func (r *ProviderRegistry) StopAll(ctx context.Context) error {
	for _, provider := range r.providers {
		if err := provider.Stop(); err != nil {
			r.logger.Error("Failed to stop provider %s: %w", provider.GetID(), err)
			continue
		}
	}

	r.logger.Info("All providers stopped successfully")
	return nil
}

// GetProviderStatus returns status of all providers
func (r *ProviderRegistry) GetProviderStatus() map[string]interface{} {
	status := make(map[string]interface{})

	for id, provider := range r.providers {
		status[id] = map[string]interface{}{
			"id":       provider.GetID(),
			"type":     provider.GetType(),
			"state":    provider.GetState(),
			"progress":  provider.GetProgress(),
		}
	}

	return status
}
