package perf

import (
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/cmd/v2broker/core"
)

// RevocationThreshold defines when permits should be revoked
type RevocationThreshold struct {
	P99LatencyMs     float64       // P99 latency threshold in milliseconds
	BufferSaturation float64       // Buffer saturation threshold (0-100)
	CheckInterval    time.Duration // How often to check metrics
}

// DefaultRevocationThreshold returns recommended threshold values
func DefaultRevocationThreshold() RevocationThreshold {
	return RevocationThreshold{
		P99LatencyMs:     30.0,               // Revoke at 30ms P99
		BufferSaturation: 80.0,               // Revoke at 80% buffer saturation
		CheckInterval:    5 * time.Second,    // Check every 5 seconds
	}
}

// PermitIntegration adds permit management capabilities to the Optimizer
type PermitIntegration struct {
	permitManager *core.PermitManager
	metrics       *MetricsCollector
	threshold     RevocationThreshold
	
	// Event callbacks
	onQuotaUpdate func(revokedPermits int, reason string, metrics *KernelMetrics)
}

// SetPermitManager attaches a PermitManager to the optimizer
func (o *Optimizer) SetPermitManager(pm *core.PermitManager) {
	o.permitIntegration = &PermitIntegration{
		permitManager: pm,
		threshold:     DefaultRevocationThreshold(),
	}
}

// SetMetricsCollector attaches a MetricsCollector to the optimizer
func (o *Optimizer) SetMetricsCollector(mc *MetricsCollector) {
	if o.permitIntegration == nil {
		o.permitIntegration = &PermitIntegration{
			threshold: DefaultRevocationThreshold(),
		}
	}
	o.permitIntegration.metrics = mc
}

// SetRevocationThreshold sets custom revocation thresholds
func (o *Optimizer) SetRevocationThreshold(threshold RevocationThreshold) {
	if o.permitIntegration == nil {
		o.permitIntegration = &PermitIntegration{
			threshold: threshold,
		}
	} else {
		o.permitIntegration.threshold = threshold
	}
}

// SetQuotaUpdateCallback sets the callback for quota update events
func (o *Optimizer) SetQuotaUpdateCallback(callback func(revokedPermits int, reason string, metrics *KernelMetrics)) {
	if o.permitIntegration == nil {
		o.permitIntegration = &PermitIntegration{
			threshold: DefaultRevocationThreshold(),
		}
	}
	o.permitIntegration.onQuotaUpdate = callback
}

// StartRevocationMonitor starts monitoring kernel metrics and revoking permits when thresholds are breached
func (o *Optimizer) StartRevocationMonitor() {
	if o.permitIntegration == nil || o.permitIntegration.permitManager == nil || o.permitIntegration.metrics == nil {
		if o.logger != nil {
			o.logger.Warn("Cannot start revocation monitor: PermitManager or MetricsCollector not set")
		}
		return
	}

	if o.logger != nil {
		o.logger.Info("Starting permit revocation monitor (P99 threshold: %.1fms, buffer threshold: %.1f%%)",
			o.permitIntegration.threshold.P99LatencyMs,
			o.permitIntegration.threshold.BufferSaturation)
	}

	go o.revocationMonitorLoop()
}

// revocationMonitorLoop continuously monitors metrics and revokes permits when needed
func (o *Optimizer) revocationMonitorLoop() {
	ticker := time.NewTicker(o.permitIntegration.threshold.CheckInterval)
	defer ticker.Stop()

	consecutiveBreaches := 0
	const breachThreshold = 2 // Require 2 consecutive breaches before revoking

	for {
		select {
		case <-ticker.C:
			o.checkAndRevokePermits(&consecutiveBreaches, breachThreshold)
		case <-o.ctx.Done():
			if o.logger != nil {
				o.logger.Info("Revocation monitor stopping")
			}
			return
		}
	}
}

// checkAndRevokePermits checks kernel metrics and revokes permits if thresholds are breached
func (o *Optimizer) checkAndRevokePermits(consecutiveBreaches *int, breachThreshold int) {
	stats := o.permitIntegration.permitManager.GetStats()
	activeWorkers := o.numWorkers
	totalPermits := stats["total_permits"].(int)
	allocatedPermits := stats["allocated_permits"].(int)
	availablePermits := stats["available_permits"].(int)

	// Get kernel metrics
	kernelMetrics := o.permitIntegration.metrics.GetMetrics(
		activeWorkers,
		totalPermits,
		allocatedPermits,
		availablePermits,
	)

	// Check if thresholds are breached
	p99Breached := kernelMetrics.P99LatencyMs > o.permitIntegration.threshold.P99LatencyMs
	bufferBreached := kernelMetrics.BufferSaturation > o.permitIntegration.threshold.BufferSaturation

	if p99Breached || bufferBreached {
		*consecutiveBreaches++

		if o.logger != nil {
			reasons := ""
			if p99Breached {
				reasons += "P99 latency exceeded threshold"
			}
			if bufferBreached {
				if reasons != "" {
					reasons += ", "
				}
				reasons += "buffer saturation exceeded threshold"
			}
			o.logger.Warn("Kernel metrics breach detected (%s): P99=%.2fms (threshold: %.1fms), Buffer=%.1f%% (threshold: %.1f%%) [breach %d/%d]",
				reasons,
				kernelMetrics.P99LatencyMs,
				o.permitIntegration.threshold.P99LatencyMs,
				kernelMetrics.BufferSaturation,
				o.permitIntegration.threshold.BufferSaturation,
				*consecutiveBreaches,
				breachThreshold)
		}

		// Only revoke after consecutive breaches to avoid flapping
		if *consecutiveBreaches >= breachThreshold {
			o.revokePermitsWithReason(kernelMetrics, p99Breached, bufferBreached)
			*consecutiveBreaches = 0 // Reset after revocation
		}
	} else {
		// Reset breach counter if metrics are healthy
		if *consecutiveBreaches > 0 {
			*consecutiveBreaches = 0
			if o.logger != nil {
				o.logger.Debug("Kernel metrics returned to normal: P99=%.2fms, Buffer=%.1f%%",
					kernelMetrics.P99LatencyMs,
					kernelMetrics.BufferSaturation)
			}
		}
	}
}

// revokePermitsWithReason revokes permits and broadcasts the event
func (o *Optimizer) revokePermitsWithReason(kernelMetrics *KernelMetrics, p99Breached, bufferBreached bool) {
	// Calculate how many permits to revoke (20% of allocated)
	stats := o.permitIntegration.permitManager.GetStats()
	allocatedPermits := stats["allocated_permits"].(int)
	
	if allocatedPermits == 0 {
		if o.logger != nil {
			o.logger.Debug("No permits to revoke (none allocated)")
		}
		return
	}

	revokeCount := (allocatedPermits * 20) / 100
	if revokeCount < 1 {
		revokeCount = 1 // Always revoke at least 1
	}

	// Build reason string
	reason := "Kernel metrics breach: "
	if p99Breached {
		reason += "P99 latency exceeded threshold (%.2fms > %.1fms)"
		reason = fmt.Sprintf(reason, kernelMetrics.P99LatencyMs, o.permitIntegration.threshold.P99LatencyMs)
	}
	if bufferBreached {
		if p99Breached {
			reason += ", "
		}
		reason += "buffer saturation exceeded threshold (%.1f%% > %.1f%%)"
		reason = fmt.Sprintf(reason, kernelMetrics.BufferSaturation, o.permitIntegration.threshold.BufferSaturation)
	}

	// Revoke permits
	revocations := o.permitIntegration.permitManager.RevokePermits(revokeCount)

	if o.logger != nil {
		o.logger.Warn("Revoked %d permits from %d providers due to kernel metrics breach",
			revokeCount, len(revocations))
		for providerID, count := range revocations {
			o.logger.Info("  - Provider %s: %d permits revoked", providerID, count)
		}
	}

	// Call quota update callback if set
	if o.permitIntegration.onQuotaUpdate != nil {
		o.permitIntegration.onQuotaUpdate(revokeCount, reason, kernelMetrics)
	}
}

// GetKernelMetrics returns current kernel performance metrics
func (o *Optimizer) GetKernelMetrics() *KernelMetrics {
	if o.permitIntegration == nil || o.permitIntegration.metrics == nil {
		return &KernelMetrics{}
	}

	stats := o.permitIntegration.permitManager.GetStats()
	return o.permitIntegration.metrics.GetMetrics(
		o.numWorkers,
		stats["total_permits"].(int),
		stats["allocated_permits"].(int),
		stats["available_permits"].(int),
	)
}
