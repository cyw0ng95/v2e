package core

import (
	"encoding/json"
	"fmt"

	"github.com/cyw0ng95/v2e/cmd/v2broker/permits"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// HandleRPCRequestPermits handles the RPCRequestPermits RPC request.
// This allows meta service (or other services) to request worker permits from the broker's global pool.
func (b *Broker) HandleRPCRequestPermits(reqMsg *proc.Message) (*proc.Message, error) {
	if b.permitManager == nil {
		return nil, fmt.Errorf("permit manager not initialized")
	}

	// Parse request parameters
	var params struct {
		ProviderID  string `json:"provider_id"`
		PermitCount int    `json:"permit_count"`
	}

	if err := json.Unmarshal(reqMsg.Payload, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request parameters: %w", err)
	}

	// Validate parameters
	if params.ProviderID == "" {
		return nil, fmt.Errorf("provider_id is required")
	}
	if params.PermitCount <= 0 {
		return nil, fmt.Errorf("permit_count must be greater than 0")
	}

	// Request permits from manager
	resp, err := b.permitManager.RequestPermits(&permits.PermitRequest{
		ProviderID:  params.ProviderID,
		PermitCount: params.PermitCount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request permits: %w", err)
	}

	// Build response
	responseData := map[string]interface{}{
		"granted":     resp.Granted,
		"available":   resp.Available,
		"provider_id": resp.ProviderID,
	}

	respMsg, err := proc.NewResponseMessage(reqMsg.ID, responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to create response message: %w", err)
	}

	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source
	respMsg.CorrelationID = reqMsg.CorrelationID

	b.logger.Debug("Handled RPCRequestPermits: provider=%s requested=%d granted=%d available=%d",
		params.ProviderID, params.PermitCount, resp.Granted, resp.Available)

	return respMsg, nil
}

// HandleRPCReleasePermits handles the RPCReleasePermits RPC request.
// This allows services to return worker permits to the broker's global pool.
func (b *Broker) HandleRPCReleasePermits(reqMsg *proc.Message) (*proc.Message, error) {
	if b.permitManager == nil {
		return nil, fmt.Errorf("permit manager not initialized")
	}

	// Parse request parameters
	var params struct {
		ProviderID  string `json:"provider_id"`
		PermitCount int    `json:"permit_count"`
	}

	if err := json.Unmarshal(reqMsg.Payload, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request parameters: %w", err)
	}

	// Validate parameters
	if params.ProviderID == "" {
		return nil, fmt.Errorf("provider_id is required")
	}
	if params.PermitCount <= 0 {
		return nil, fmt.Errorf("permit_count must be greater than 0")
	}

	// Release permits
	resp, err := b.permitManager.ReleasePermits(params.ProviderID, params.PermitCount)
	if err != nil {
		return nil, fmt.Errorf("failed to release permits: %w", err)
	}

	// Build response
	responseData := map[string]interface{}{
		"success":     true,
		"available":   resp.Available,
		"provider_id": resp.ProviderID,
	}

	respMsg, err := proc.NewResponseMessage(reqMsg.ID, responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to create response message: %w", err)
	}

	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source
	respMsg.CorrelationID = reqMsg.CorrelationID

	b.logger.Debug("Handled RPCReleasePermits: provider=%s released=%d available=%d",
		params.ProviderID, params.PermitCount, resp.Available)

	return respMsg, nil
}

// HandleRPCGetKernelMetrics handles the RPCGetKernelMetrics RPC request.
// This returns current kernel performance metrics (P99 latency, buffer saturation, etc.)
func (b *Broker) HandleRPCGetKernelMetrics(reqMsg *proc.Message) (*proc.Message, error) {
	if b.optimizer == nil {
		return nil, fmt.Errorf("optimizer not initialized")
	}

	// Get kernel metrics from optimizer
	metrics := b.optimizer.GetKernelMetrics()

	// Build response with all metrics
	responseData := map[string]interface{}{
		"p99_latency_ms":    metrics.P99LatencyMs,
		"buffer_saturation": metrics.BufferSaturation,
		"active_workers":    metrics.ActiveWorkers,
		"total_permits":     metrics.TotalPermits,
		"allocated_permits": metrics.AllocatedPermits,
		"available_permits": metrics.AvailablePermits,
		"message_rate":      metrics.MessageRate,
		"error_rate":        metrics.ErrorRate,
	}

	respMsg, err := proc.NewResponseMessage(reqMsg.ID, responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to create response message: %w", err)
	}

	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source
	respMsg.CorrelationID = reqMsg.CorrelationID

	b.logger.Debug("Handled RPCGetKernelMetrics: P99=%.2fms buffer=%.1f%% permits=%d/%d",
		metrics.P99LatencyMs, metrics.BufferSaturation,
		metrics.AllocatedPermits, metrics.TotalPermits)

	return respMsg, nil
}

// SendQuotaUpdateEvent broadcasts an RPCOnQuotaUpdate event to specified providers
// This is called when permits are revoked due to kernel metrics breaches
func (b *Broker) SendQuotaUpdateEvent(providerIDs []string, revokedPermits int, reason string, metrics map[string]interface{}) error {
	eventData := map[string]interface{}{
		"revoked_permits": revokedPermits,
		"reason":          reason,
		"kernel_metrics":  metrics,
	}

	eventMsg, err := proc.NewEventMessage("RPCOnQuotaUpdate", eventData)
	if err != nil {
		return fmt.Errorf("failed to create quota update event: %w", err)
	}

	eventMsg.Source = "broker"

	// Send event to each affected provider
	for _, providerID := range providerIDs {
		eventMsg.Target = providerID
		if err := b.SendToProcess(providerID, eventMsg); err != nil {
			b.logger.Warn("Failed to send quota update event to %s: %v", providerID, err)
		} else {
			b.logger.Debug("Sent quota update event to %s: revoked=%d", providerID, revokedPermits)
		}
	}

	return nil
}
