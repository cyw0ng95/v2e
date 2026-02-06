package scaling

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// PredictionModel tracks prediction accuracy
type PredictionModel struct {
	mu                 sync.Mutex
	predictions        []PredictionRecord
	maxHistorySize     int
	totalPredictions   int
	correctPredictions int
	totalAbsoluteError float64
	modelType          string
	featureWeights     map[string]float64
	lastRetrained      time.Time
	retrainInterval    time.Duration
	retrainCallback    func() error
}

// PredictionRecord represents a single prediction
type PredictionRecord struct {
	ID         string
	Timestamp  time.Time
	Metric     string
	Predicted  float64
	Actual     float64
	Error      float64
	IsCorrect  bool
	Confidence float64
}

// ModelMetrics represents model performance metrics
type ModelMetrics struct {
	Accuracy          float64
	MeanAbsoluteError float64
	MeanSquaredError  float64
	RMSE              float64
	Precision         float64
	Recall            float64
	F1Score           float64
	TotalPredictions  int
	LastRetrained     time.Time
	ModelAge          time.Duration
}

// NewPredictionModel creates a new prediction model
func NewPredictionModel(modelType string, maxHistorySize int, retrainInterval time.Duration) *PredictionModel {
	return &PredictionModel{
		predictions:     make([]PredictionRecord, 0, maxHistorySize),
		maxHistorySize:  maxHistorySize,
		modelType:       modelType,
		featureWeights:  make(map[string]float64),
		retrainInterval: retrainInterval,
		lastRetrained:   time.Now(),
	}
}

// SetRetrainCallback sets the model retraining callback
func (pm *PredictionModel) SetRetrainCallback(callback func() error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.retrainCallback = callback
}

// SetFeatureWeight sets a feature weight
func (pm *PredictionModel) SetFeatureWeight(feature string, weight float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.featureWeights[feature] = weight
}

// Predict makes a prediction for a metric
func (pm *PredictionModel) Predict(metric string, features map[string]float64) (string, float64, float64, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	prediction := pm.calculatePrediction(features)
	confidence := pm.calculateConfidence(features)

	record := PredictionRecord{
		ID:         generatePredictionID(),
		Timestamp:  time.Now(),
		Metric:     metric,
		Predicted:  prediction,
		Actual:     0,
		Confidence: confidence,
	}

	pm.predictions = append(pm.predictions, record)
	if len(pm.predictions) > pm.maxHistorySize {
		pm.predictions = pm.predictions[1:]
	}

	pm.totalPredictions++

	return record.ID, prediction, confidence, nil
}

// RecordActual records the actual value and calculates error
func (pm *PredictionModel) RecordActual(predictionID string, actual float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i := range pm.predictions {
		if pm.predictions[i].ID == predictionID && pm.predictions[i].Actual == 0 {
			pm.predictions[i].Actual = actual
			pm.predictions[i].Error = math.Abs(pm.predictions[i].Predicted - actual)

			// Consider correct if error is within 10% tolerance
			tolerance := math.Abs(actual * 0.1)
			pm.predictions[i].IsCorrect = pm.predictions[i].Error <= tolerance

			if pm.predictions[i].IsCorrect {
				pm.correctPredictions++
			}

			pm.totalAbsoluteError += pm.predictions[i].Error

			break
		}
	}

	// Check if retraining is needed
	pm.checkRetraining()
}

// calculatePrediction calculates a prediction based on features and weights
func (pm *PredictionModel) calculatePrediction(features map[string]float64) float64 {
	sum := 0.0
	weightSum := 0.0

	for feature, value := range features {
		weight := pm.featureWeights[feature]
		sum += value * weight
		weightSum += math.Abs(weight)
	}

	if weightSum == 0 {
		return 0
	}

	return sum / weightSum
}

// calculateConfidence calculates prediction confidence
func (pm *PredictionModel) calculateConfidence(features map[string]float64) float64 {
	if len(pm.predictions) < 10 {
		return 0.5
	}

	recent := pm.predictions[len(pm.predictions)-10:]
	if len(recent) == 0 {
		return 0.5
	}

	var correctCount int
	for _, pred := range recent {
		if pred.Actual != 0 {
			if pred.IsCorrect {
				correctCount++
			}
		}
	}

	return float64(correctCount) / float64(len(recent))
}

// checkRetraining checks if model needs retraining
func (pm *PredictionModel) checkRetraining() {
	if pm.retrainCallback == nil {
		return
	}

	if time.Since(pm.lastRetrained) >= pm.retrainInterval {
		go func() {
			err := pm.retrainCallback()
			if err != nil {
				return
			}

			pm.mu.Lock()
			pm.lastRetrained = time.Now()
			pm.mu.Unlock()
		}()
	}
}

// GetMetrics returns current model metrics
func (pm *PredictionModel) GetMetrics() ModelMetrics {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metrics := ModelMetrics{
		TotalPredictions: pm.totalPredictions,
		LastRetrained:    pm.lastRetrained,
		ModelAge:         time.Since(pm.lastRetrained),
	}

	if pm.totalPredictions > 0 {
		metrics.Accuracy = float64(pm.correctPredictions) / float64(pm.totalPredictions)
	}

	varianceSum := 0.0
	count := 0
	for _, pred := range pm.predictions {
		if pred.Actual != 0 {
			varianceSum += pred.Error * pred.Error
			count++
		}
	}

	if count > 0 {
		metrics.MeanAbsoluteError = pm.totalAbsoluteError / float64(count)
		metrics.MeanSquaredError = varianceSum / float64(count)
		metrics.RMSE = math.Sqrt(metrics.MeanSquaredError)
	}

	return metrics
}

// ClearHistory clears prediction history
func (pm *PredictionModel) ClearHistory() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.predictions = make([]PredictionRecord, 0, pm.maxHistorySize)
	pm.totalPredictions = 0
	pm.correctPredictions = 0
	pm.totalAbsoluteError = 0
}

// Retrain triggers manual retraining
func (pm *PredictionModel) Retrain() error {
	if pm.retrainCallback == nil {
		return fmt.Errorf("no retrain callback set")
	}

	err := pm.retrainCallback()
	if err != nil {
		return err
	}

	pm.mu.Lock()
	pm.lastRetrained = time.Now()
	pm.mu.Unlock()

	return nil
}

// GetPredictions returns prediction history
func (pm *PredictionModel) GetPredictions() []PredictionRecord {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return append([]PredictionRecord{}, pm.predictions...)
}

// GetRecentPredictions returns recent predictions for a metric
func (pm *PredictionModel) GetRecentPredictions(metric string, limit int) []PredictionRecord {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	result := make([]PredictionRecord, 0, limit)
	for i := len(pm.predictions) - 1; i >= 0 && len(result) < limit; i-- {
		if pm.predictions[i].Metric == metric {
			result = append(result, pm.predictions[i])
		}
	}
	return result
}

func generatePredictionID() string {
	return fmt.Sprintf("pred-%d", time.Now().UnixNano())
}
