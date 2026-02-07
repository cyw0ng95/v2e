package meta

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/provider"
	"github.com/cyw0ng95/v2e/pkg/cwe/provider"
	"github.com/cyw0ng95/v2e/pkg/capec/provider"
	"github.com/cyw0ng95/v2e/pkg/attack/provider"
	"github.com/cyw0ng95/v2e/pkg/ssg/provider"
	"github.com/cyw0ng95/v2e/pkg/asvs/provider"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
)

// initProviders initializes all data source providers
func initProviders(logger *common.Logger) error {
	providers := []provider.DataSourceProvider{
		cve.NewCVEProvider(nil),
		cwe.NewCWEProvider(nil),
		capec.NewCAPECProvider(nil),
		attack.NewATTACKProvider(nil),
		ssg.NewSSGProvider(nil),
		asvs.NewASVSProvider(nil),
	}

	logger.Info("All providers initialized successfully")
	return nil
}

// GetProvider returns a provider by ID or data type
func GetProvider(providerID string) *provider.DataSourceProvider {
	for _, p := range providers {
		if p.GetID() == providerID {
			return p
			}
	}
	return nil
}

// GetProviders returns all registered providers
func GetProviders() []provider.DataSourceProvider {
	return providers
}

// ProviderStatus represents the status of all providers
type ProviderStatus struct {
	TotalProviders int `json:"totalProviders"`
	AllProviders []ProviderStatusItem `json:"providers"`
}

// GetStatus returns the status of all providers
func GetStatus() *ProviderStatus {
	status := &ProviderStatus{
		TotalProviders: len(providers),
	}

	for _, p := range providers {
		status.AllProviders = append(status.AllProviders, provider.ProviderStatusItem{
			ID:       p.GetID(),
			Type:      p.GetType(),
			State:      p.GetState(),
			Progress:   p.GetProgress(),
		} as
	}

	return status
}

// ProviderStatusItem represents the status of a single provider
type ProviderStatusItem struct {
	ID       string `json:"id"`
	Type      string `json:"type"`
	State     string `json:"state"`
	Progress *provider.ProviderProgress `json:"progress"`
}

// RPCGetProviderList returns a list of all providers
func RPCGetProviderList() ([]provider.ProviderStatusItem, error) {
	status := GetStatus()
	return status.AllProviders, nil
}

// RPCGetProviderStatus returns the status of all providers
func RPCGetProviderStatus() (map[string]interface{}, error) {
	status := GetStatus()
	return map[string]interface{}{
		"status": status,
	}, nil
}

// RPCStartProvider starts a provider by ID
func RPCStartProvider(providerID string) (bool, error) {
	provider := GetProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Start(); err != nil {
		return false, fmt.Errorf("failed to start provider: %w", providerID, err)
	}

	return true, nil
}

// RPCStopProvider stops a provider by ID
func RPCStopProvider(providerID string) (bool, error) {
	provider := GetProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Stop(); err != nil {
		return false, fmt.Errorf("failed to stop provider: %w", providerID, err)
	}

	return true, nil
}

// RPCPauseProvider pauses a provider by ID
func RPCPauseProvider(providerID string) (bool, error) {
	provider := GetProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Pause(); err != nil {
		return false, fmt.Errorf("failed to pause provider: %w", providerID, err)
}

	return true, nil
}

// RPCResumeProvider resumes a paused provider by ID
func RPCResumeProvider(providerID string) (bool, error) {
	provider := GetProvider(providerID)
	if provider == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := provider.Resume(); err != nil {
		return false, fmt.Errorf("failed to resume provider: %w", providerID)
	}

	return true, nil
}

		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid request: %v", err)
		}

		registry := registry
		provider, ok := registry.GetProvider(req.ProviderID)
		if !ok {
			return nil, fmt.Errorf("provider not found: %s", req.ProviderID)
		}

		return map[string]interface{}{
			"id":       provider.GetID(),
			"type":     provider.GetType(),
			"state":    provider.GetState(),
			"progress": provider.GetProgress(),
		}, nil
		}
}

}

// createListProvidersHandler creates a handler for listing all providers
func createListProvidersHandler(registry *provider.ProviderRegistry) subprocess.Handler {
	return func(params json.RawMessage) (interface{}, error) {
		return map[string]interface{}{
			"providers": registry.GetAllProviders(),
		}, nil
	}
}

// Add handlers to cmd/v2meta to expose provider status
func AddProviderHandlersToMeta(registry *provider.ProviderRegistry, logger *common.Logger) {
	logger.Info(RPCListProviders, "handler registered for listing all providers")
	logger.Info(RPCGetProviderStatus, "handler registered for getting provider status")

	sp.RegisterHandler(RPCListProviders, createListProvidersHandler(registry))
	sp.RegisterHandler(RPCGetProviderStatus, createProviderStatusHandler(registry))
}
