package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	cveprovider "github.com/cyw0ng95/v2e/pkg/cve/provider"
	cweprovider "github.com/cyw0ng95/v2e/pkg/cwe/provider"
	capecprovider "github.com/cyw0ng95/v2e/pkg/capec/provider"
	attackprovider "github.com/cyw0ng95/v2e/pkg/attack/provider"
	ssgprovider "github.com/cyw0ng95/v2e/pkg/ssg/provider"
	asvsprovider "github.com/cyw0ng95/v2e/pkg/asvs/provider"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Global registry of data source providers
var providers []provider.DataSourceProvider

// initProviders initializes all data source providers
func initProviders(logger *common.Logger) error {
	var providerList []provider.DataSourceProvider

	// CVE Provider
	cveProv, err := cveprovider.NewCVEProvider("")
	if err != nil {
		return fmt.Errorf("failed to create CVE provider: %w", err)
	}
	providerList = append(providerList, cveProv)

	// CWE Provider
	cweProv, err := cweprovider.NewCWEProvider("")
	if err != nil {
		return fmt.Errorf("failed to create CWE provider: %w", err)
	}
	providerList = append(providerList, cweProv)

	// CAPEC Provider
	capecProv, err := capecprovider.NewCAPECProvider("")
	if err != nil {
		return fmt.Errorf("failed to create CAPEC provider: %w", err)
	}
	providerList = append(providerList, capecProv)

	// ATT&CK Provider
	attackProv, err := attackprovider.NewATTACKProvider("")
	if err != nil {
		return fmt.Errorf("failed to create ATT&CK provider: %w", err)
	}
	providerList = append(providerList, attackProv)

	// SSG Provider
	ssgProv, err := ssgprovider.NewSSGProvider("")
	if err != nil {
		return fmt.Errorf("failed to create SSG provider: %w", err)
	}
	providerList = append(providerList, ssgProv)

	// ASVS Provider
	asvsProv, err := asvsprovider.NewASVSProvider("")
	if err != nil {
		return fmt.Errorf("failed to create ASVS provider: %w", err)
	}
	providerList = append(providerList, asvsProv)

	providers = providerList

	// Initialize each provider
	for _, p := range providers {
		if err := p.Initialize(context.Background()); err != nil {
			return fmt.Errorf("failed to initialize provider %s: %w", p.GetID(), err)
		}
	}

	logger.Info("All providers initialized successfully")
	return nil
}

// GetProvider returns a provider by ID or data type
func GetProvider(providerID string) provider.DataSourceProvider {
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
	TotalProviders int                      `json:"totalProviders"`
	AllProviders   []ProviderStatusItem     `json:"providers"`
}

// GetStatus returns the status of all providers
func GetStatus() *ProviderStatus {
	status := &ProviderStatus{
		TotalProviders: len(providers),
	}

	for _, p := range providers {
		status.AllProviders = append(status.AllProviders, ProviderStatusItem{
			ID:       p.GetID(),
			Type:     p.GetType(),
			State:    string(p.GetState()),
			Progress: p.GetProgress(),
		})
	}

	return status
}

// ProviderStatusItem represents the status of a single provider
type ProviderStatusItem struct {
	ID       string                      `json:"id"`
	Type     string                      `json:"type"`
	State    string                      `json:"state"`
	Progress *provider.ProviderProgress  `json:"progress"`
}

// RPCGetProviderList returns a list of all providers
func RPCGetProviderList() ([]ProviderStatusItem, error) {
	status := GetStatus()
	return status.AllProviders, nil
}

// RPCStartProvider starts a provider by ID
func RPCStartProvider(providerID string) (bool, error) {
	p := GetProvider(providerID)
	if p == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := p.Start(); err != nil {
		return false, fmt.Errorf("failed to start provider: %w", err)
	}

	return true, nil
}

// RPCStopProvider stops a provider by ID
func RPCStopProvider(providerID string) (bool, error) {
	p := GetProvider(providerID)
	if p == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := p.Stop(); err != nil {
		return false, fmt.Errorf("failed to stop provider: %w", err)
	}

	return true, nil
}

// RPCPauseProvider pauses a provider by ID
func RPCPauseProvider(providerID string) (bool, error) {
	p := GetProvider(providerID)
	if p == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := p.Pause(); err != nil {
		return false, fmt.Errorf("failed to pause provider: %w", err)
	}

	return true, nil
}

// RPCResumeProvider resumes a paused provider by ID
func RPCResumeProvider(providerID string) (bool, error) {
	p := GetProvider(providerID)
	if p == nil {
		return false, fmt.Errorf("provider not found: %s", providerID)
	}

	if err := p.Resume(); err != nil {
		return false, fmt.Errorf("failed to resume provider: %w", err)
	}

	return true, nil
}

// createListProvidersHandler creates a handler for listing all providers
func createListProvidersHandler(registry *ProviderRegistry) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		providers := registry.GetAllProviders()
		providerStatuses := make([]map[string]interface{}, 0, len(providers))
		for _, p := range providers {
			providerStatuses = append(providerStatuses, map[string]interface{}{
				"id":       p.GetID(),
				"type":     p.GetType(),
				"state":    p.GetState(),
				"progress": p.GetProgress(),
			})
		}
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"providers": providerStatuses,
		})
	}
}
