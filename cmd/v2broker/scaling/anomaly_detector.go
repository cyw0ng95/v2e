package scaling

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// AnomalyDetector detects abnormal system behavior
type AnomalyDetector struct {
	mu              sync.Mutex
	metricsHistory  []TimeSeriesPoint
	maxHistorySize  int
	alertThreshold  float64
	alertCallback   func(AnomalyAlert)
	detectionWindow time.Duration
	enabled         bool
}

// TimeSeriesPoint represents a single metric sample
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
	Metric    string
}

// AnomalyAlert represents an anomaly detection alert
type AnomalyAlert struct {
	Timestamp     time.Time
	Metric        string
	CurrentValue  float64
	ExpectedValue float64
	Deviation     float64
	Severity      string
	Description   string
}

const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(maxHistorySize int, alertThreshold float64, detectionWindow time.Duration) *AnomalyDetector {
	return &AnomalyDetector{
		metricsHistory:  make([]TimeSeriesPoint, 0, maxHistorySize),
		maxHistorySize:  maxHistorySize,
		alertThreshold:  alertThreshold,
		detectionWindow: detectionWindow,
		enabled:         true,
	}
}

// SetAlertCallback sets the callback function for alerts
func (ad *AnomalyDetector) SetAlertCallback(callback func(AnomalyAlert)) {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.alertCallback = callback
}

// RecordMetric records a new metric sample
func (ad *AnomalyDetector) RecordMetric(metric string, value float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if !ad.enabled {
		return
	}

	point := TimeSeriesPoint{
		Timestamp: time.Now(),
		Value:     value,
		Metric:    metric,
	}

	ad.metricsHistory = append(ad.metricsHistory, point)

	if len(ad.metricsHistory) > ad.maxHistorySize {
		ad.metricsHistory = ad.metricsHistory[1:]
	}

	ad.checkForAnomalies(metric, value)
}

// checkForAnomalies checks if the current metric value is anomalous
func (ad *AnomalyDetector) checkForAnomalies(metric string, currentValue float64) {
	if len(ad.metricsHistory) < 10 {
		return
	}

	recentPoints := ad.getRecentPoints(metric, 20)
	if len(recentPoints) < 10 {
		return
	}

	expected, stdDev := ad.calculateStatistics(recentPoints)
	if stdDev == 0 {
		return
	}

	zScore := math.Abs((currentValue - expected) / stdDev)

	if zScore >= ad.alertThreshold {
		severity := ad.calculateSeverity(zScore)
		description := fmt.Sprintf("%s deviates by %.2f standard deviations", metric, zScore)

		alert := AnomalyAlert{
			Timestamp:     time.Now(),
			Metric:        metric,
			CurrentValue:  currentValue,
			ExpectedValue: expected,
			Deviation:     zScore,
			Severity:      severity,
			Description:   description,
		}

		if ad.alertCallback != nil {
			go ad.alertCallback(alert)
		}
	}
}

// getRecentPoints gets recent points for a specific metric
func (ad *AnomalyDetector) getRecentPoints(metric string, count int) []TimeSeriesPoint {
	result := make([]TimeSeriesPoint, 0, count)
	for i := len(ad.metricsHistory) - 1; i >= 0 && len(result) < count; i-- {
		if ad.metricsHistory[i].Metric == metric {
			result = append(result, ad.metricsHistory[i])
		}
	}
	return result
}

// calculateStatistics calculates mean and standard deviation
func (ad *AnomalyDetector) calculateStatistics(points []TimeSeriesPoint) (mean, stdDev float64) {
	if len(points) == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, p := range points {
		sum += p.Value
	}
	mean = sum / float64(len(points))

	varianceSum := 0.0
	for _, p := range points {
		diff := p.Value - mean
		varianceSum += diff * diff
	}
	stdDev = math.Sqrt(varianceSum / float64(len(points)))

	return mean, stdDev
}

// calculateSeverity determines the severity level based on z-score
func (ad *AnomalyDetector) calculateSeverity(zScore float64) string {
	if zScore >= 5.0 {
		return SeverityCritical
	}
	if zScore >= 4.0 {
		return SeverityHigh
	}
	if zScore >= 3.0 {
		return SeverityMedium
	}
	return SeverityLow
}

// Enable enables the anomaly detector
func (ad *AnomalyDetector) Enable() {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.enabled = true
}

// Disable disables the anomaly detector
func (ad *AnomalyDetector) Disable() {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.enabled = false
}

// GetMetricsHistory returns the current metrics history
func (ad *AnomalyDetector) GetMetricsHistory() []TimeSeriesPoint {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	return append([]TimeSeriesPoint{}, ad.metricsHistory...)
}

// ClearHistory clears the metrics history
func (ad *AnomalyDetector) ClearHistory() {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.metricsHistory = make([]TimeSeriesPoint, 0, ad.maxHistorySize)
}
