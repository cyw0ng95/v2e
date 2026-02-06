package permits

import (
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// PermitManager manages the global pool of worker permits
// This is the Master component that controls resource allocation
type PermitManager struct {
	mu               sync.RWMutex
	totalPermits     int
	availablePermits int
	allocations      map[string]int // provider_id -> allocated permits
	logger           *common.Logger
	
	// Metrics
	totalRequests  int64
	totalGrants    int64
	totalReleases  int64
	totalRevocations int64
}

// PermitRequest represents a request for worker permits
type PermitRequest struct {
	ProviderID   string
	PermitCount  int
}

// PermitResponse represents the response to a permit request
type PermitResponse struct {
	Granted      int
	Available    int
	ProviderID   string
}

// PermitAllocation represents current permit allocation
type PermitAllocation struct {
	ProviderID      string
	AllocatedCount  int
	AllocatedAt     time.Time
}

// NewPermitManager creates a new permit manager with the specified total permits
func NewPermitManager(totalPermits int, logger *common.Logger) *PermitManager {
	if totalPermits <= 0 {
		totalPermits = 10 // Default to 10 permits
	}
	
	return &PermitManager{
		totalPermits:     totalPermits,
		availablePermits: totalPermits,
		allocations:      make(map[string]int),
		logger:           logger,
	}
}

// RequestPermits attempts to allocate permits to a provider
// Returns the number of permits granted (may be less than requested)
func (pm *PermitManager) RequestPermits(req *PermitRequest) (*PermitResponse, error) {
	if req.ProviderID == "" {
		return nil, fmt.Errorf("provider_id is required")
	}
	if req.PermitCount <= 0 {
		return nil, fmt.Errorf("permit_count must be greater than 0")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.totalRequests++

	// Grant as many permits as available (up to the requested amount)
	granted := req.PermitCount
	if granted > pm.availablePermits {
		granted = pm.availablePermits
	}

	if granted > 0 {
		pm.availablePermits -= granted
		pm.allocations[req.ProviderID] += granted
		pm.totalGrants++
		
		if pm.logger != nil {
			pm.logger.Debug("Granted %d permits to %s (requested: %d, available: %d)", 
				granted, req.ProviderID, req.PermitCount, pm.availablePermits)
		}
	} else {
		if pm.logger != nil {
			pm.logger.Debug("No permits available for %s (requested: %d)", 
				req.ProviderID, req.PermitCount)
		}
	}

	return &PermitResponse{
		Granted:    granted,
		Available:  pm.availablePermits,
		ProviderID: req.ProviderID,
	}, nil
}

// ReleasePermits returns permits from a provider to the pool
func (pm *PermitManager) ReleasePermits(providerID string, count int) (*PermitResponse, error) {
	if providerID == "" {
		return nil, fmt.Errorf("provider_id is required")
	}
	if count <= 0 {
		return nil, fmt.Errorf("permit_count must be greater than 0")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	allocated, exists := pm.allocations[providerID]
	if !exists {
		return nil, fmt.Errorf("no permits allocated to provider %s", providerID)
	}

	// Can't release more than allocated
	if count > allocated {
		count = allocated
	}

	pm.availablePermits += count
	pm.allocations[providerID] -= count
	if pm.allocations[providerID] == 0 {
		delete(pm.allocations, providerID)
	}
	pm.totalReleases++

	if pm.logger != nil {
		pm.logger.Debug("Released %d permits from %s (available: %d)", 
			count, providerID, pm.availablePermits)
	}

	return &PermitResponse{
		Granted:    count,
		Available:  pm.availablePermits,
		ProviderID: providerID,
	}, nil
}

// RevokePermits forcibly revokes permits across all providers
// This is called by the broker when kernel metrics breach thresholds
// Returns a map of provider_id -> revoked_count
func (pm *PermitManager) RevokePermits(count int) map[string]int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if count <= 0 || len(pm.allocations) == 0 {
		return nil
	}

	revocations := make(map[string]int)
	remaining := count

	// Distribute revocations proportionally across providers
	totalAllocated := 0
	for _, allocated := range pm.allocations {
		totalAllocated += allocated
	}

	for providerID, allocated := range pm.allocations {
		if remaining <= 0 {
			break
		}

		// Revoke proportionally (at least 1 if there's any allocation)
		toRevoke := (allocated * count) / totalAllocated
		if toRevoke == 0 && allocated > 0 && remaining > 0 {
			toRevoke = 1
		}
		if toRevoke > remaining {
			toRevoke = remaining
		}
		if toRevoke > allocated {
			toRevoke = allocated
		}

		pm.allocations[providerID] -= toRevoke
		if pm.allocations[providerID] == 0 {
			delete(pm.allocations, providerID)
		}
		pm.availablePermits += toRevoke
		revocations[providerID] = toRevoke
		remaining -= toRevoke
		pm.totalRevocations++

		if pm.logger != nil {
			pm.logger.Warn("Revoked %d permits from %s (kernel metrics breach)", 
				toRevoke, providerID)
		}
	}

	return revocations
}

// GetStats returns current permit manager statistics
func (pm *PermitManager) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	allocatedCount := 0
	for _, count := range pm.allocations {
		allocatedCount += count
	}

	return map[string]interface{}{
		"total_permits":      pm.totalPermits,
		"available_permits":  pm.availablePermits,
		"allocated_permits":  allocatedCount,
		"active_providers":   len(pm.allocations),
		"total_requests":     pm.totalRequests,
		"total_grants":       pm.totalGrants,
		"total_releases":     pm.totalReleases,
		"total_revocations":  pm.totalRevocations,
	}
}

// GetAllocations returns current permit allocations per provider
func (pm *PermitManager) GetAllocations() map[string]int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	allocations := make(map[string]int, len(pm.allocations))
	for k, v := range pm.allocations {
		allocations[k] = v
	}
	return allocations
}

// SetTotalPermits updates the total permit pool size
// This can be used for dynamic scaling based on system load
func (pm *PermitManager) SetTotalPermits(total int) error {
	if total <= 0 {
		return fmt.Errorf("total permits must be greater than 0")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Calculate currently allocated permits
	allocatedCount := 0
	for _, count := range pm.allocations {
		allocatedCount += count
	}

	// Ensure new total can accommodate current allocations
	if total < allocatedCount {
		return fmt.Errorf("cannot reduce total permits below currently allocated (%d)", allocatedCount)
	}

	oldTotal := pm.totalPermits
	pm.totalPermits = total
	pm.availablePermits = total - allocatedCount

	if pm.logger != nil {
		pm.logger.Info("Updated total permits from %d to %d (available: %d)", 
			oldTotal, total, pm.availablePermits)
	}

	return nil
}
