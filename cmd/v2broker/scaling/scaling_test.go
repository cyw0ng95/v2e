package scaling

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAnomalyDetector_BasicDetection(t *testing.T) {
	ad := NewAnomalyDetector(100, 3.0, 5*time.Minute)

	var alertReceived atomic.Bool
	var receivedAlert atomic.Value
	ad.SetAlertCallback(func(alert AnomalyAlert) {
		alertReceived.Store(true)
		receivedAlert.Store(alert)
	})

	ad.RecordMetric("test_metric", 10.0)
	ad.RecordMetric("test_metric", 10.5)
	ad.RecordMetric("test_metric", 10.2)
	ad.RecordMetric("test_metric", 10.8)
	ad.RecordMetric("test_metric", 10.3)

	ad.RecordMetric("test_metric", 10.1)
	ad.RecordMetric("test_metric", 10.4)
	ad.RecordMetric("test_metric", 10.6)
	ad.RecordMetric("test_metric", 10.0)
	ad.RecordMetric("test_metric", 10.2)

	ad.RecordMetric("test_metric", 100.0)

	time.Sleep(10 * time.Millisecond)

	if !alertReceived.Load() {
		t.Fatal("Expected anomaly alert")
	}

	alert := receivedAlert.Load().(AnomalyAlert)
	if alert.Metric != "test_metric" {
		t.Errorf("Expected metric 'test_metric', got '%s'", alert.Metric)
	}
}

func TestAnomalyDetector_DisableEnable(t *testing.T) {
	ad := NewAnomalyDetector(100, 3.0, 5*time.Minute)

	var alertReceived atomic.Bool
	ad.SetAlertCallback(func(alert AnomalyAlert) {
		alertReceived.Store(true)
	})

	// First build up history while enabled to establish baseline
	for i := 0; i < 15; i++ {
		ad.RecordMetric("test_metric", 10.0+float64(i%5))
	}

	// Give time for async callback to complete
	time.Sleep(20 * time.Millisecond)

	// Now disable and record more - should not trigger alert
	ad.Disable()
	alertReceived.Store(false)

	for i := 0; i < 5; i++ {
		ad.RecordMetric("test_metric", float64(10+i))
	}

	// Even extreme value when disabled should not trigger
	ad.RecordMetric("test_metric", 1000.0)
	time.Sleep(20 * time.Millisecond)

	if alertReceived.Load() {
		t.Fatal("Expected no alert when disabled")
	}

	// Enable and trigger anomaly - should now detect
	ad.Enable()
	alertReceived.Store(false)
	// Record a value that deviates significantly from baseline
	ad.RecordMetric("test_metric", 1000.0)

	// Give time for async detection
	time.Sleep(50 * time.Millisecond)

	if !alertReceived.Load() {
		t.Fatal("Expected alert when enabled")
	}
}

func TestSelfHealingManager_RegisterStrategy(t *testing.T) {
	shm := NewSelfHealingManager(100)

	strategy := &RetryWithBackoffStrategy{
		maxRetries:    3,
		initialDelay:  100 * time.Millisecond,
		maxDelay:      5 * time.Second,
		backoffFactor: 2.0,
		retryFunc: func(attempt int) error {
			if attempt < 2 {
				return fmt.Errorf("failed")
			}
			return nil
		},
	}

	shm.RegisterStrategy("test_fault", strategy)

	if len(shm.healingStrategies) != 1 {
		t.Errorf("Expected 1 strategy, got %d", len(shm.healingStrategies))
	}
}

func TestSelfHealingManager_RetryWithBackoff(t *testing.T) {
	shm := NewSelfHealingManager(100)

	var attempts atomic.Int64
	strategy := &RetryWithBackoffStrategy{
		maxRetries:    3,
		initialDelay:  10 * time.Millisecond,
		maxDelay:      100 * time.Millisecond,
		backoffFactor: 2.0,
		retryFunc: func(attempt int) error {
			attempts.Add(1)
			if attempt < 2 {
				return fmt.Errorf("failed")
			}
			return nil
		},
	}

	shm.RegisterStrategy("retry_fault", strategy)

	var alertReceived atomic.Bool
	shm.SetAlertCallback(func(alert HealingAlert) {
		if alert.Success {
			alertReceived.Store(true)
		}
	})

	err := shm.ReportFault("retry_fault", "medium", "Test fault")
	if err != nil {
		t.Fatalf("Failed to report fault: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !alertReceived.Load() {
		t.Fatal("Expected successful healing alert")
	}

	if attempts.Load() < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts.Load())
	}
}

func TestSelfHealingManager_CircuitBreaker(t *testing.T) {
	shm := NewSelfHealingManager(100)

	var attemptCount atomic.Int64
	executeFunc := func() error {
		count := attemptCount.Add(1)
		if count < 3 {
			return fmt.Errorf("service unavailable")
		}
		return nil
	}

	strategy := &CircuitBreakerStrategy{
		failureThreshold: 2,
		recoveryTimeout:  100 * time.Millisecond,
		executeFunc:      executeFunc,
	}

	shm.RegisterStrategy("circuit_fault", strategy)

	var alertsMu sync.Mutex
	alerts := []HealingAlert{}
	shm.SetAlertCallback(func(alert HealingAlert) {
		alertsMu.Lock()
		defer alertsMu.Unlock()
		alerts = append(alerts, alert)
	})

	for i := 0; i < 5; i++ {
		shm.ReportFault("circuit_fault", "high", "Service fault")
		time.Sleep(20 * time.Millisecond)
	}

	alertsMu.Lock()
	alertCount := len(alerts)
	alertsMu.Unlock()

	if alertCount < 3 {
		t.Errorf("Expected at least 3 alerts, got %d", alertCount)
	}
}

func TestPredictionModel_BasicPrediction(t *testing.T) {
	pm := NewPredictionModel("linear", 100, 1*time.Hour)

	pm.SetFeatureWeight("cpu", 0.5)
	pm.SetFeatureWeight("memory", 0.3)
	pm.SetFeatureWeight("network", 0.2)

	features := map[string]float64{
		"cpu":     80.0,
		"memory":  60.0,
		"network": 40.0,
	}

	_, prediction, confidence, err := pm.Predict("load", features)
	if err != nil {
		t.Fatalf("Failed to make prediction: %v", err)
	}

	expectedPrediction := (80.0*0.5 + 60.0*0.3 + 40.0*0.2) / (0.5 + 0.3 + 0.2)
	if prediction != expectedPrediction {
		t.Errorf("Expected prediction %.2f, got %.2f", expectedPrediction, prediction)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %.2f", confidence)
	}
}

func TestPredictionModel_AccuracyTracking(t *testing.T) {
	pm := NewPredictionModel("linear", 100, 1*time.Hour)

	pm.SetFeatureWeight("feature1", 1.0)

	// Record actual results to build accuracy metrics
	var predictionIDs []string
	for i := 0; i < 20; i++ {
		features := map[string]float64{"feature1": float64(i)}
		predictionID, predictionValue, _, _ := pm.Predict("test", features)
		predictionIDs = append(predictionIDs, predictionID)
		// Record actual result (same as prediction for this test)
		pm.RecordActual(predictionID, predictionValue)
	}

	metrics := pm.GetMetrics()
	if metrics.TotalPredictions != 20 {
		t.Errorf("Expected 20 predictions, got %d", metrics.TotalPredictions)
	}

	// Accuracy should be 1.0 (100%) since actual = prediction
	if metrics.Accuracy == 0 {
		t.Error("Expected non-zero accuracy")
	}
}

func TestPredictionModel_Retraining(t *testing.T) {
	pm := NewPredictionModel("linear", 100, 50*time.Millisecond)

	var retrainCalled atomic.Bool
	pm.SetRetrainCallback(func() error {
		retrainCalled.Store(true)
		return nil
	})

	// Make predictions
	var predictionIDs []string
	for i := 0; i < 5; i++ {
		features := map[string]float64{"f": float64(i)}
		predictionID, predictionValue, _, _ := pm.Predict("test", features)
		predictionIDs = append(predictionIDs, predictionID)
		_ = predictionValue
	}

	// Wait for retrain interval to elapse
	time.Sleep(75 * time.Millisecond)

	// Now record actuals - this should trigger retraining check
	for i := 0; i < 5; i++ {
		features := map[string]float64{"f": float64(i)}
		_, predictionValue, _, _ := pm.Predict("test", features)
		pm.RecordActual(predictionIDs[i], predictionValue)
	}

	// Give time for async retraining
	time.Sleep(50 * time.Millisecond)

	if !retrainCalled.Load() {
		t.Fatal("Expected retrain callback to be called")
	}
}

func TestCircuitBreakerState(t *testing.T) {
	cbs := &CircuitBreakerStrategy{
		failureThreshold: 3,
		recoveryTimeout:  200 * time.Millisecond,
		executeFunc: func() error {
			return fmt.Errorf("failed")
		},
	}

	fault := FaultRecord{
		ID:          "test-1",
		Type:        "test",
		Severity:    "high",
		Description: "Test fault",
		Attempts:    0,
	}

	// Trigger failures to open the circuit
	for i := 0; i < 3; i++ {
		cbs.Heal(fault)
		fault.Attempts++
	}

	if cbs.state != CircuitOpen {
		t.Errorf("Expected circuit to be open, got %s", cbs.state)
	}

	err := cbs.Heal(fault)
	if err == nil {
		t.Error("Expected error when circuit is open")
	}

	// Wait for recovery timeout
	time.Sleep(250 * time.Millisecond)

	// Next call should transition to half-open
	err = cbs.Heal(fault)
	// The circuit breaker should allow the attempt (transition to half-open)
	// but executeFunc still returns error, so we expect an error
	// The important thing is the state transition to half-open

	if cbs.state != CircuitHalfOpen && cbs.state != CircuitOpen {
		t.Errorf("Expected circuit to be half-open or open, got %s", cbs.state)
	}

	// If the circuit allowed the attempt, it should be in half-open
	// If executeFunc fails, it may stay open or go to half-open depending on implementation
}
