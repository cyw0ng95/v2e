package scaling

import (
	"fmt"
	"sync"
	"time"
)

type ScalingDecision string

const (
	ScalingDecisionNone      ScalingDecision = "none"
	ScalingDecisionScaleUp   ScalingDecision = "scale_up"
	ScalingDecisionScaleDown ScalingDecision = "scale_down"
)

type ScalingAction struct {
	Decision       ScalingDecision
	ResourceType   string
	CurrentValue   float64
	PredictedValue float64
	Threshold      float64
	Confidence     float64
	Reason         string
	Execute        func() error
}

type ScalingConfig struct {
	MinWorkers             int
	MaxWorkers             int
	ScaleUpThreshold       float64
	ScaleDownThreshold     float64
	PredictionHorizon      time.Duration
	ConfidenceThreshold    float64
	CooldownPeriod         time.Duration
	EnableProactiveScaling bool
}

type AutoScaler struct {
	predictor      *LoadPredictor
	config         ScalingConfig
	scalingActions chan ScalingAction
	cooldowns      map[string]time.Time
	mu             sync.RWMutex
	closed         bool
	stopChan       chan struct{}
}

type ScalingMetrics struct {
	TotalScaleUps     int
	TotalScaleDowns   int
	TotalSkipped      int
	TotalPredictions  int
	AverageConfidence float64
	LastScalingTime   time.Time
}

func NewAutoScaler(predictor *LoadPredictor, config ScalingConfig) *AutoScaler {
	if config.ScaleUpThreshold <= 0 {
		config.ScaleUpThreshold = 0.8
	}

	if config.ScaleDownThreshold <= 0 {
		config.ScaleDownThreshold = 0.3
	}

	if config.PredictionHorizon <= 0 {
		config.PredictionHorizon = 5 * time.Minute
	}

	if config.ConfidenceThreshold <= 0 {
		config.ConfidenceThreshold = 0.7
	}

	if config.CooldownPeriod <= 0 {
		config.CooldownPeriod = 2 * time.Minute
	}

	return &AutoScaler{
		predictor:      predictor,
		config:         config,
		scalingActions: make(chan ScalingAction, 100),
		cooldowns:      make(map[string]time.Time),
		stopChan:       make(chan struct{}),
	}
}

func (as *AutoScaler) Start() error {
	go as.scalingLoop()
	return nil
}

func (as *AutoScaler) Stop() {
	if as.closed {
		return
	}

	as.closed = true
	close(as.stopChan)
	close(as.scalingActions)
}

func (as *AutoScaler) scalingLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			as.evaluateScaling()

		case action := <-as.scalingActions:
			as.executeAction(action)

		case <-as.stopChan:
			return
		}
	}
}

func (as *AutoScaler) evaluateScaling() {
	predictions, err := as.predictor.PredictAll(as.config.PredictionHorizon)
	if err != nil {
		return
	}

	for metricType, predictedValues := range predictions {
		if len(predictedValues) == 0 {
			continue
		}

		predictedValue := predictedValues[len(predictedValues)-1]
		decision, action := as.evaluateMetric(metricType, predictedValue)

		if decision != ScalingDecisionNone {
			as.scalingActions <- action
		}
	}
}

func (as *AutoScaler) evaluateMetric(metricType MetricType, predictedValue float64) (ScalingDecision, ScalingAction) {
	as.mu.Lock()
	defer as.mu.Unlock()

	metricName := string(metricType)

	if as.inCooldown(metricName) {
		return ScalingDecisionNone, ScalingAction{}
	}

	if predictedValue > as.config.ScaleUpThreshold {
		return ScalingDecisionScaleUp, ScalingAction{
			Decision:       ScalingDecisionScaleUp,
			ResourceType:   metricName,
			CurrentValue:   0,
			PredictedValue: predictedValue,
			Threshold:      as.config.ScaleUpThreshold,
			Confidence:     as.getConfidence(metricType),
			Reason:         fmt.Sprintf("Predicted value %.2f exceeds scale-up threshold %.2f", predictedValue, as.config.ScaleUpThreshold),
			Execute:        as.scaleUp,
		}
	}

	if predictedValue < as.config.ScaleDownThreshold {
		return ScalingDecisionScaleDown, ScalingAction{
			Decision:       ScalingDecisionScaleDown,
			ResourceType:   metricName,
			CurrentValue:   0,
			PredictedValue: predictedValue,
			Threshold:      as.config.ScaleDownThreshold,
			Confidence:     as.getConfidence(metricType),
			Reason:         fmt.Sprintf("Predicted value %.2f below scale-down threshold %.2f", predictedValue, as.config.ScaleDownThreshold),
			Execute:        as.scaleDown,
		}
	}

	return ScalingDecisionNone, ScalingAction{}
}

func (as *AutoScaler) inCooldown(resourceType string) bool {
	cooldownTime, exists := as.cooldowns[resourceType]
	if !exists {
		return false
	}

	return time.Since(cooldownTime) < as.config.CooldownPeriod
}

func (as *AutoScaler) setCooldown(resourceType string) {
	as.cooldowns[resourceType] = time.Now()
}

func (as *AutoScaler) getConfidence(metricType MetricType) float64 {
	accuracy, err := as.predictor.GetModelAccuracy(metricType)
	if err != nil {
		return 0.5
	}

	return accuracy / 100.0
}

func (as *AutoScaler) executeAction(action ScalingAction) error {
	if action.Decision == ScalingDecisionNone {
		return nil
	}

	confidence := action.Confidence
	if confidence < as.config.ConfidenceThreshold {
		return fmt.Errorf("confidence %.2f below threshold %.2f", confidence, as.config.ConfidenceThreshold)
	}

	if err := action.Execute(); err != nil {
		return fmt.Errorf("failed to execute scaling action: %w", err)
	}

	as.mu.Lock()
	as.setCooldown(action.ResourceType)
	as.mu.Unlock()

	return nil
}

func (as *AutoScaler) scaleUp() error {
	fmt.Printf("[AutoScaler] Scaling up resources")
	as.config.MinWorkers++
	return nil
}

func (as *AutoScaler) scaleDown() error {
	if as.config.MinWorkers <= as.config.MinWorkers {
		return fmt.Errorf("cannot scale below minimum workers %d", as.config.MinWorkers)
	}

	fmt.Printf("[AutoScaler] Scaling down resources")
	as.config.MinWorkers--
	return nil
}

func (as *AutoScaler) GetScalingActions() <-chan ScalingAction {
	return as.scalingActions
}

func (as *AutoScaler) GetConfig() ScalingConfig {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.config
}

func (as *AutoScaler) UpdateConfig(newConfig ScalingConfig) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.config = newConfig
}

func (as *AutoScaler) GetMetrics() ScalingMetrics {
	return ScalingMetrics{
		TotalScaleUps:     0,
		TotalScaleDowns:   0,
		TotalSkipped:      0,
		TotalPredictions:  0,
		AverageConfidence: 0.75,
		LastScalingTime:   time.Now(),
	}
}

func NewDefaultScalingConfig() ScalingConfig {
	return ScalingConfig{
		MinWorkers:             1,
		MaxWorkers:             16,
		ScaleUpThreshold:       0.8,
		ScaleDownThreshold:     0.3,
		PredictionHorizon:      5 * time.Minute,
		ConfidenceThreshold:    0.7,
		CooldownPeriod:         2 * time.Minute,
		EnableProactiveScaling: true,
	}
}
