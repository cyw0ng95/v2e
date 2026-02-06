package perf

import (
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// ServicePriority defines the priority level for service scheduling
type ServicePriority int

const (
	// PriorityLow - Background tasks, analysis, batch operations
	PriorityLow ServicePriority = 1
	// PriorityNormal - Regular ETL operations
	PriorityNormal ServicePriority = 5
	// PriorityHigh - Frontend requests, interactive queries
	PriorityHigh ServicePriority = 10
	// PriorityCritical - Broker internal operations
	PriorityCritical ServicePriority = 20
)

// ServiceType identifies the type of service
type ServiceType string

const (
	ServiceTypeAnalysis ServiceType = "analysis"
	ServiceTypeFrontend ServiceType = "frontend"
	ServiceTypeETL      ServiceType = "etl"
	ServiceTypeBroker   ServiceType = "broker"
	ServiceTypeOther    ServiceType = "other"
)

// ServiceMetrics tracks performance metrics for a service
type ServiceMetrics struct {
	ServiceID        string
	ServiceType      ServiceType
	Priority         ServicePriority
	RequestCount     int64
	TotalLatencyMs   int64
	ActiveRequests   int64
	DroppedRequests  int64
	LastRequestTime  time.Time
}

// AnalysisOptimizer provides specialized optimization for analysis service
type AnalysisOptimizer struct {
	mu             sync.RWMutex
	logger         *common.Logger
	
	// Service metrics
	services       map[string]*ServiceMetrics
	
	// Resource allocation policies
	maxAnalysisConcurrency int
	maxFrontendConcurrency int
	maxETLConcurrency      int
	
	// Conflict resolution
	conflictPolicy ConflictPolicy
	
	// Dynamic throttling
	analysisThrottle   *ThrottleState
	frontendThrottle   *ThrottleState
	etlThrottle        *ThrottleState
}

// ConflictPolicy defines how to handle resource conflicts
type ConflictPolicy string

const (
	// PolicyFrontendFirst - Frontend requests take precedence
	PolicyFrontendFirst ConflictPolicy = "frontend_first"
	// PolicyFairShare - Equal distribution among active services
	PolicyFairShare ConflictPolicy = "fair_share"
	// PolicyWeighted - Weighted by priority levels
	PolicyWeighted ConflictPolicy = "weighted"
)

// ThrottleState tracks throttling state for a service type
type ThrottleState struct {
	Enabled          bool
	CurrentLimit     int
	BaseLimit        int
	ThrottleReason   string
	ThrottleStart    time.Time
	LastAdjustment   time.Time
}

// NewAnalysisOptimizer creates a new analysis-specific optimizer
func NewAnalysisOptimizer(logger *common.Logger) *AnalysisOptimizer {
	return &AnalysisOptimizer{
		logger:                 logger,
		services:               make(map[string]*ServiceMetrics),
		maxAnalysisConcurrency: 2,  // Low concurrency for background analysis
		maxFrontendConcurrency: 10, // High concurrency for user requests
		maxETLConcurrency:      5,  // Medium concurrency for ETL
		conflictPolicy:         PolicyFrontendFirst,
		analysisThrottle: &ThrottleState{
			Enabled:      false,
			CurrentLimit: 2,
			BaseLimit:    2,
		},
		frontendThrottle: &ThrottleState{
			Enabled:      false,
			CurrentLimit: 10,
			BaseLimit:    10,
		},
		etlThrottle: &ThrottleState{
			Enabled:      false,
			CurrentLimit: 5,
			BaseLimit:    5,
		},
	}
}

// RegisterService registers a service for optimization
func (ao *AnalysisOptimizer) RegisterService(serviceID string, serviceType ServiceType, priority ServicePriority) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	ao.services[serviceID] = &ServiceMetrics{
		ServiceID:   serviceID,
		ServiceType: serviceType,
		Priority:    priority,
	}
	
	if ao.logger != nil {
		ao.logger.Info("Registered service %s (type: %s, priority: %d)", serviceID, serviceType, priority)
	}
}

// RecordRequest records a service request for metrics
func (ao *AnalysisOptimizer) RecordRequest(serviceID string, latencyMs int64) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	if metrics, exists := ao.services[serviceID]; exists {
		metrics.RequestCount++
		metrics.TotalLatencyMs += latencyMs
		metrics.LastRequestTime = time.Now()
	}
}

// RecordActiveRequest increments active request count
func (ao *AnalysisOptimizer) RecordActiveRequest(serviceID string) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	if metrics, exists := ao.services[serviceID]; exists {
		metrics.ActiveRequests++
	}
}

// RecordCompletedRequest decrements active request count
func (ao *AnalysisOptimizer) RecordCompletedRequest(serviceID string) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	if metrics, exists := ao.services[serviceID]; exists {
		if metrics.ActiveRequests > 0 {
			metrics.ActiveRequests--
		}
	}
}

// ShouldThrottle determines if a service should be throttled
func (ao *AnalysisOptimizer) ShouldThrottle(serviceID string) (bool, string) {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	
	metrics, exists := ao.services[serviceID]
	if !exists {
		return false, ""
	}
	
	// Get throttle state for service type
	var throttle *ThrottleState
	switch metrics.ServiceType {
	case ServiceTypeAnalysis:
		throttle = ao.analysisThrottle
	case ServiceTypeFrontend:
		throttle = ao.frontendThrottle
	case ServiceTypeETL:
		throttle = ao.etlThrottle
	default:
		return false, ""
	}
	
	// Check if service is over concurrency limit
	if metrics.ActiveRequests >= int64(throttle.CurrentLimit) {
		return true, "concurrency limit reached"
	}
	
	// Check if throttling is enabled
	if throttle.Enabled {
		return true, throttle.ThrottleReason
	}
	
	return false, ""
}

// DetectConflict detects resource conflicts between services
func (ao *AnalysisOptimizer) DetectConflict() (bool, []string) {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	
	// Count active services by type
	activeAnalysis := int64(0)
	activeFrontend := int64(0)
	activeETL := int64(0)
	
	for _, metrics := range ao.services {
		if metrics.ActiveRequests > 0 {
			switch metrics.ServiceType {
			case ServiceTypeAnalysis:
				activeAnalysis += metrics.ActiveRequests
			case ServiceTypeFrontend:
				activeFrontend += metrics.ActiveRequests
			case ServiceTypeETL:
				activeETL += metrics.ActiveRequests
			}
		}
	}
	
	// Conflict detection based on policy
	var conflicts []string
	
	// High frontend load with active analysis
	if activeFrontend >= 5 && activeAnalysis > 0 {
		conflicts = append(conflicts, "frontend_analysis_conflict")
	}
	
	// High ETL load with active analysis
	if activeETL >= 3 && activeAnalysis > 0 {
		conflicts = append(conflicts, "etl_analysis_conflict")
	}
	
	// All services active at high concurrency
	if activeFrontend >= 3 && activeETL >= 2 && activeAnalysis > 0 {
		conflicts = append(conflicts, "all_services_conflict")
	}
	
	return len(conflicts) > 0, conflicts
}

// ResolveConflict resolves resource conflicts by adjusting service priorities
func (ao *AnalysisOptimizer) ResolveConflict(conflicts []string) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	if ao.logger != nil {
		ao.logger.Warn("Resolving resource conflicts: %v (policy: %s)", conflicts, ao.conflictPolicy)
	}
	
	switch ao.conflictPolicy {
	case PolicyFrontendFirst:
		ao.resolveFrontendFirst(conflicts)
	case PolicyFairShare:
		ao.resolveFairShare(conflicts)
	case PolicyWeighted:
		ao.resolveWeighted(conflicts)
	}
}

// resolveFrontendFirst throttles analysis/ETL in favor of frontend
func (ao *AnalysisOptimizer) resolveFrontendFirst(conflicts []string) {
	for _, conflict := range conflicts {
		switch conflict {
		case "frontend_analysis_conflict":
			// Throttle analysis to minimum
			ao.analysisThrottle.Enabled = true
			ao.analysisThrottle.CurrentLimit = 1
			ao.analysisThrottle.ThrottleReason = "frontend priority"
			ao.analysisThrottle.ThrottleStart = time.Now()
			
			if ao.logger != nil {
				ao.logger.Info("Throttling analysis service for frontend priority")
			}
			
		case "etl_analysis_conflict":
			// Throttle analysis when ETL is active
			ao.analysisThrottle.Enabled = true
			ao.analysisThrottle.CurrentLimit = 1
			ao.analysisThrottle.ThrottleReason = "etl priority"
			ao.analysisThrottle.ThrottleStart = time.Now()
			
		case "all_services_conflict":
			// Pause analysis completely
			ao.analysisThrottle.Enabled = true
			ao.analysisThrottle.CurrentLimit = 0
			ao.analysisThrottle.ThrottleReason = "all services active"
			ao.analysisThrottle.ThrottleStart = time.Now()
			
			if ao.logger != nil {
				ao.logger.Warn("Pausing analysis service due to high system load")
			}
		}
	}
}

// resolveFairShare distributes resources equally
func (ao *AnalysisOptimizer) resolveFairShare(conflicts []string) {
	// Equal distribution: each service gets 1/3 of base limits
	ao.analysisThrottle.CurrentLimit = ao.analysisThrottle.BaseLimit / 3
	ao.frontendThrottle.CurrentLimit = ao.frontendThrottle.BaseLimit / 3
	ao.etlThrottle.CurrentLimit = ao.etlThrottle.BaseLimit / 3
	
	if ao.logger != nil {
		ao.logger.Info("Applying fair share policy: analysis=%d, frontend=%d, etl=%d",
			ao.analysisThrottle.CurrentLimit,
			ao.frontendThrottle.CurrentLimit,
			ao.etlThrottle.CurrentLimit)
	}
}

// resolveWeighted distributes resources by priority weight
func (ao *AnalysisOptimizer) resolveWeighted(conflicts []string) {
	// Frontend: 50%, ETL: 30%, Analysis: 20%
	ao.analysisThrottle.CurrentLimit = (ao.analysisThrottle.BaseLimit * 20) / 100
	ao.frontendThrottle.CurrentLimit = (ao.frontendThrottle.BaseLimit * 50) / 100
	ao.etlThrottle.CurrentLimit = (ao.etlThrottle.BaseLimit * 30) / 100
	
	if ao.analysisThrottle.CurrentLimit < 1 {
		ao.analysisThrottle.CurrentLimit = 1
	}
	
	if ao.logger != nil {
		ao.logger.Info("Applying weighted policy: analysis=%d, frontend=%d, etl=%d",
			ao.analysisThrottle.CurrentLimit,
			ao.frontendThrottle.CurrentLimit,
			ao.etlThrottle.CurrentLimit)
	}
}

// ClearThrottles resets all throttling when conflicts are resolved
func (ao *AnalysisOptimizer) ClearThrottles() {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	ao.analysisThrottle.Enabled = false
	ao.analysisThrottle.CurrentLimit = ao.analysisThrottle.BaseLimit
	ao.frontendThrottle.Enabled = false
	ao.frontendThrottle.CurrentLimit = ao.frontendThrottle.BaseLimit
	ao.etlThrottle.Enabled = false
	ao.etlThrottle.CurrentLimit = ao.etlThrottle.BaseLimit
	
	if ao.logger != nil {
		ao.logger.Info("Cleared all service throttles")
	}
}

// GetMetrics returns current service metrics
func (ao *AnalysisOptimizer) GetMetrics() map[string]*ServiceMetrics {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	
	// Return a copy to prevent concurrent modification
	result := make(map[string]*ServiceMetrics)
	for k, v := range ao.services {
		metrics := *v
		result[k] = &metrics
	}
	return result
}

// SetConflictPolicy changes the conflict resolution policy
func (ao *AnalysisOptimizer) SetConflictPolicy(policy ConflictPolicy) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	
	ao.conflictPolicy = policy
	if ao.logger != nil {
		ao.logger.Info("Conflict policy changed to: %s", policy)
	}
}
