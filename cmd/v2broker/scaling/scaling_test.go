package scaling

import (
	"fmt"
	"testing"
	"time"
)

func TestAnomalyDetector_BasicDetection(t *testing.T) {
	ad := NewAnomalyDetector(100, 3.0, 5*time.Minute)

	alertReceived := false
	var receivedAlert AnomalyAlert
	ad.SetAlertCallback(func(alert AnomalyAlert) {
		alertReceived = true
		receivedAlert = alert
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

	if !alertReceived {
		t.Fatal("Expected anomaly alert")
	}

	if receivedAlert.Metric != "test_metric" {
		t.Errorf("Expected metric 'test_metric', got '%s'", receivedAlert.Metric)
	}
}

func TestAnomalyDetector_DisableEnable(t *testing.T) {
	ad := NewAnomalyDetector(100, 3.0, 5*time.Minute)

	alertReceived := false
	ad.SetAlertCallback(func(alert AnomalyAlert) {
		alertReceived = true
	})

	ad.Disable()

	for i := 0; i < 10; i++ {
		ad.RecordMetric("test_metric", float64(10+i))
	}

	ad.RecordMetric("test_metric", 1000.0)

	if alertReceived {
		t.Fatal("Expected no alert when disabled")
	}

	ad.Enable()
	ad.RecordMetric("test_metric", 1000.0)

	if !alertReceived {
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

	attempts := 0
	strategy := &RetryWithBackoffStrategy{
		maxRetries:    3,
		initialDelay:  10 * time.Millisecond,
		maxDelay:      100 * time.Millisecond,
		backoffFactor: 2.0,
		retryFunc: func(attempt int) error {
			attempts++
			if attempt < 2 {
				return fmt.Errorf("failed")
			}
			return nil
		},
	}

	shm.RegisterStrategy("retry_fault", strategy)

	alertReceived := false
	shm.SetAlertCallback(func(alert HealingAlert) {
		if alert.Success {
			alertReceived = true
		}
	})

	err := shm.ReportFault("retry_fault", "medium", "Test fault")
	if err != nil {
		t.Fatalf("Failed to report fault: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !alertReceived {
		t.Fatal("Expected successful healing alert")
	}

	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}
}

func TestSelfHealingManager_CircuitBreaker(t *testing.T) {
	shm := NewSelfHealingManager(100)

	attemptCount := 0
	executeFunc := func() error {
		attemptCount++
		if attemptCount < 3 {
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

	alerts := []HealingAlert{}
	shm.SetAlertCallback(func(alert HealingAlert) {
		alerts = append(alerts, alert)
	})

	for i := 0; i < 5; i++ {
		shm.ReportFault("circuit_fault", "high", "Service fault")
		time.Sleep(20 * time.Millisecond)
	}

	if len(alerts) < 3 {
		t.Errorf("Expected at least 3 alerts, got %d", len(alerts))
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

	for i := 0; i < 20; i++ {
		features := map[string]float64{"feature1": float64(i)}
		_, _, _, _ = pm.Predict("test", features)
		time.Sleep(1 * time.Millisecond)
	}

	metrics := pm.GetMetrics()
	if metrics.TotalPredictions != 20 {
		t.Errorf("Expected 20 predictions, got %d", metrics.TotalPredictions)
	}

	if metrics.Accuracy == 0 {
		t.Error("Expected non-zero accuracy")
	}
}

func TestPredictionModel_Retraining(t *testing.T) {
	pm := NewPredictionModel("linear", 100, 100*time.Millisecond)

	retrainCalled := false
	pm.SetRetrainCallback(func() error {
		retrainCalled = true
		return nil
	})

	_, _, _, _ = pm.Predict("test", map[string]float64{"f": 1.0})
	time.Sleep(150 * time.Millisecond)

	if !retrainCalled {
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

	time.Sleep(250 * time.Millisecond)

	err = cbs.Heal(fault)
	if err != nil {
		t.Errorf("Expected circuit to transition to half-open: %v", err)
	}

	if cbs.state != CircuitHalfOpen {
		t.Errorf("Expected circuit to be half-open, got %s", cbs.state)
	}
}
