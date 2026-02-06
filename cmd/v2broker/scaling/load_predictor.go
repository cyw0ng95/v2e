package scaling

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type MetricType string

const (
	MetricTypeCPU         MetricType = "cpu"
	MetricTypeMemory      MetricType = "memory"
	MetricTypeRequests    MetricType = "requests"
	MetricTypeLatency     MetricType = "latency"
	MetricTypeConnections MetricType = "connections"
	MetricTypeGoroutines  MetricType = "goroutines"
	MetricTypeDiskIO      MetricType = "disk_io"
	MetricTypeNetworkIO   MetricType = "network_io"
)

type MetricData struct {
	Timestamp time.Time
	Type      MetricType
	Value     float64
	Labels    map[string]string
}

type PredictionModel interface {
	Train(data []MetricData) error
	Predict(metricType MetricType, horizon time.Duration) ([]float64, error)
	Accuracy() float64
	Features() []string
}

type LoadPredictor struct {
	models            map[MetricType]PredictionModel
	historicalData    map[MetricType][]MetricData
	maxHistory        int
	trainingWindow    time.Duration
	predictionHorizon time.Duration
	mu                sync.RWMutex
}

type LinearRegressionModel struct {
	weights   map[string]float64
	intercept float64
	features  []string
	accuracy  float64
	trained   bool
}

type MovingAverageModel struct {
	windowSize int
	values     []float64
	accuracy   float64
	trained    bool
}

type ExponentialSmoothingModel struct {
	alpha    float64
	values   []float64
	accuracy float64
	trained  bool
}

type SeasonalDecompositionModel struct {
	trend    []float64
	seasonal []float64
	period   int
	accuracy float64
	trained  bool
}

func NewLoadPredictor(maxHistory int, trainingWindow, predictionHorizon time.Duration) *LoadPredictor {
	return &LoadPredictor{
		models:            make(map[MetricType]PredictionModel),
		historicalData:    make(map[MetricType][]MetricData),
		maxHistory:        maxHistory,
		trainingWindow:    trainingWindow,
		predictionHorizon: predictionHorizon,
	}
}

func (lp *LoadPredictor) SetModel(metricType MetricType, model PredictionModel) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	lp.models[metricType] = model
}

func (lp *LoadPredictor) AddMetric(data MetricData) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	if lp.historicalData[data.Type] == nil {
		lp.historicalData[data.Type] = make([]MetricData, 0, lp.maxHistory)
	}

	lp.historicalData[data.Type] = append(lp.historicalData[data.Type], data)

	if len(lp.historicalData[data.Type]) > lp.maxHistory {
		lp.historicalData[data.Type] = lp.historicalData[data.Type][1:]
	}
}

func (lp *LoadPredictor) Train(metricType MetricType) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	model, exists := lp.models[metricType]
	if !exists {
		return fmt.Errorf("no model set for metric type %s", metricType)
	}

	data := lp.getTrainingData(metricType)
	if len(data) < 10 {
		return fmt.Errorf("insufficient data for training: need at least 10 samples, got %d", len(data))
	}

	return model.Train(data)
}

func (lp *LoadPredictor) getTrainingData(metricType MetricType) []MetricData {
	data := lp.historicalData[metricType]
	if data == nil {
		return nil
	}

	now := time.Now()
	cutoff := now.Add(-lp.trainingWindow)

	var result []MetricData
	for _, d := range data {
		if d.Timestamp.After(cutoff) {
			result = append(result, d)
		}
	}

	return result
}

func (lp *LoadPredictor) Predict(metricType MetricType, horizon time.Duration) ([]float64, error) {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	model, exists := lp.models[metricType]
	if !exists {
		return nil, fmt.Errorf("no model set for metric type %s", metricType)
	}

	if !isModelTrained(model) {
		return nil, fmt.Errorf("model not trained for metric type %s", metricType)
	}

	return model.Predict(metricType, horizon)
}

func (lp *LoadPredictor) PredictAll(horizon time.Duration) (map[MetricType][]float64, error) {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	results := make(map[MetricType][]float64)

	for metricType, model := range lp.models {
		if isModelTrained(model) {
			predictions, err := model.Predict(metricType, horizon)
			if err != nil {
				return nil, fmt.Errorf("prediction failed for %s: %w", metricType, err)
			}
			results[metricType] = predictions
		}
	}

	return results, nil
}

func (lp *LoadPredictor) GetModelAccuracy(metricType MetricType) (float64, error) {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	model, exists := lp.models[metricType]
	if !exists {
		return 0, fmt.Errorf("no model set for metric type %s", metricType)
	}

	return model.Accuracy(), nil
}

func (lp *LoadPredictor) GetHistoricalData(metricType MetricType) []MetricData {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	data := lp.historicalData[metricType]
	result := make([]MetricData, len(data))
	copy(result, data)

	return result
}

func (lp *LoadPredictor) ClearHistory(metricType MetricType) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.historicalData[metricType] = nil
}

func (lp *LoadPredictor) RetrainAll() error {
	lp.mu.RLock()
	defer lp.mu.RUnlock()

	var errs []error
	for metricType := range lp.models {
		if err := lp.Train(metricType); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", metricType, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple training errors: %v", errs)
	}

	return nil
}

func (lrm *LinearRegressionModel) Train(data []MetricData) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data: need at least 2 samples")
	}

	lrm.trained = false

	values := make([]float64, len(data))
	for i, d := range data {
		values[i] = d.Value
	}

	if lrm.weights == nil {
		lrm.weights = make(map[string]float64)
	}

	lrm.weights["trend"] = calculateTrend(values)
	lrm.weights["volatility"] = calculateVolatility(values)
	lrm.intercept = calculateIntercept(values, lrm.weights["trend"])

	lrm.trained = true
	lrm.accuracy = calculateAccuracy(data, lrm)

	return nil
}

func (lrm *LinearRegressionModel) Predict(metricType MetricType, horizon time.Duration) ([]float64, error) {
	if !lrm.trained {
		return nil, fmt.Errorf("model not trained")
	}

	steps := int(horizon.Seconds())
	if steps <= 0 {
		steps = 1
	}

	predictions := make([]float64, steps)
	for i := 0; i < steps; i++ {
		trend := lrm.weights["trend"]
		predictions[i] = lrm.intercept + float64(i+1)*trend
	}

	return predictions, nil
}

func (lrm *LinearRegressionModel) Accuracy() float64 {
	return lrm.accuracy
}

func (lrm *LinearRegressionModel) Features() []string {
	return lrm.features
}

func (mav *MovingAverageModel) Train(data []MetricData) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data: need at least 2 samples")
	}

	mav.values = make([]float64, len(data))
	for i, d := range data {
		mav.values[i] = d.Value
	}

	if mav.windowSize == 0 {
		mav.windowSize = int(math.Min(float64(len(data)), 10))
	}

	mav.trained = true
	mav.accuracy = calculateAccuracy(data, mav)

	return nil
}

func (mav *MovingAverageModel) Predict(metricType MetricType, horizon time.Duration) ([]float64, error) {
	if !mav.trained {
		return nil, fmt.Errorf("model not trained")
	}

	steps := int(horizon.Seconds())
	if steps <= 0 {
		steps = 1
	}

	ma := calculateMovingAverage(mav.values, mav.windowSize)
	predictions := make([]float64, steps)
	for i := 0; i < steps; i++ {
		predictions[i] = ma
	}

	return predictions, nil
}

func (mav *MovingAverageModel) Accuracy() float64 {
	return mav.accuracy
}

func (mav *MovingAverageModel) Features() []string {
	return []string{"moving_average"}
}

func (esm *ExponentialSmoothingModel) Train(data []MetricData) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data: need at least 2 samples")
	}

	esm.values = make([]float64, len(data))
	for i, d := range data {
		esm.values[i] = d.Value
	}

	if esm.alpha <= 0 || esm.alpha > 1 {
		esm.alpha = 0.3
	}

	esm.trained = true
	esm.accuracy = calculateAccuracy(data, esm)

	return nil
}

func (esm *ExponentialSmoothingModel) Predict(metricType MetricType, horizon time.Duration) ([]float64, error) {
	if !esm.trained {
		return nil, fmt.Errorf("model not trained")
	}

	steps := int(horizon.Seconds())
	if steps <= 0 {
		steps = 1
	}

	smoothed := esm.values[len(esm.values)-1]
	predictions := make([]float64, steps)
	for i := 0; i < steps; i++ {
		predictions[i] = smoothed
	}

	return predictions, nil
}

func (esm *ExponentialSmoothingModel) Accuracy() float64 {
	return esm.accuracy
}

func (esm *ExponentialSmoothingModel) Features() []string {
	return []string{"exponential_smoothing"}
}

func (sdm *SeasonalDecompositionModel) Train(data []MetricData) error {
	if len(data) < sdm.period*2 {
		return fmt.Errorf("insufficient data: need at least %d samples", sdm.period*2)
	}

	sdm.trend = make([]float64, len(data))
	sdm.seasonal = make([]float64, len(data))

	values := make([]float64, len(data))
	for i, d := range data {
		values[i] = d.Value
	}

	sdm.trend = calculateTrendComponent(values, sdm.period)
	sdm.seasonal = calculateSeasonalComponent(values, sdm.trend, sdm.period)

	sdm.trained = true
	sdm.accuracy = calculateAccuracy(data, sdm)

	return nil
}

func (sdm *SeasonalDecompositionModel) Predict(metricType MetricType, horizon time.Duration) ([]float64, error) {
	if !sdm.trained {
		return nil, fmt.Errorf("model not trained")
	}

	steps := int(horizon.Seconds())
	if steps <= 0 {
		steps = 1
	}

	lastTrend := sdm.trend[len(sdm.trend)-1]
	predictions := make([]float64, steps)
	for i := 0; i < steps; i++ {
		seasonalIdx := (len(sdm.seasonal) + i) % sdm.period
		predictions[i] = lastTrend + sdm.seasonal[seasonalIdx]
	}

	return predictions, nil
}

func (sdm *SeasonalDecompositionModel) Accuracy() float64 {
	return sdm.accuracy
}

func (sdm *SeasonalDecompositionModel) Features() []string {
	return []string{"trend", "seasonal"}
}

func calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX := n * (n - 1) / 2
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for i, v := range values {
		x := float64(i)
		sumY += v
		sumXY += x * v
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	return slope
}

func calculateVolatility(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values) - 1)

	return math.Sqrt(variance)
}

func calculateIntercept(values []float64, slope float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	meanX := float64(len(values)-1) / 2
	return mean - slope*meanX
}

func calculateMovingAverage(values []float64, windowSize int) float64 {
	if len(values) == 0 {
		return 0
	}

	window := int(math.Min(float64(windowSize), float64(len(values))))
	sum := 0.0
	for i := len(values) - window; i < len(values); i++ {
		sum += values[i]
	}

	return sum / float64(window)
}

func calculateTrendComponent(values []float64, period int) []float64 {
	trend := make([]float64, len(values))

	for i := range values {
		window := int(math.Min(float64(period), float64(len(values))))
		if i >= window/2 && i < len(values)-window/2 {
			sum := 0.0
			for j := i - window/2; j <= i+window/2; j++ {
				sum += values[j]
			}
			trend[i] = sum / float64(window)
		} else {
			trend[i] = values[i]
		}
	}

	return trend
}

func calculateSeasonalComponent(values, trend []float64, period int) []float64 {
	seasonal := make([]float64, len(values))

	for i := 0; i < len(values); i++ {
		seasonal[i] = values[i] - trend[i]
	}

	avgSeasonal := make([]float64, period)
	for i := range avgSeasonal {
		sum := 0.0
		count := 0
		for j := i; j < len(seasonal); j += period {
			sum += seasonal[j]
			count++
		}
		if count > 0 {
			avgSeasonal[i] = sum / float64(count)
		}
	}

	for i := range seasonal {
		seasonal[i] = avgSeasonal[i%period]
	}

	return seasonal
}

func calculateAccuracy(data []MetricData, model PredictionModel) float64 {
	if len(data) < 2 {
		return 0
	}

	trainSize := int(float64(len(data)) * 0.8)
	trainData := data[:trainSize]
	testData := data[trainSize:]

	_ = model.Train(trainData)

	totalError := 0.0
	for _, d := range testData {
		predictions, err := model.Predict(d.Type, 1*time.Second)
		if err != nil {
			continue
		}
		if len(predictions) > 0 {
			error := math.Abs(predictions[0] - d.Value)
			totalError += error
		}
	}

	meanValue := 0.0
	for _, d := range testData {
		meanValue += d.Value
	}
	meanValue /= float64(len(testData))

	if meanValue == 0 {
		return 1.0
	}

	meanAbsoluteError := totalError / float64(len(testData))
	mape := (meanAbsoluteError / meanValue) * 100

	accuracy := math.Max(0, 100-mape)
	return accuracy
}

func isModelTrained(model PredictionModel) bool {
	lrm, ok := model.(*LinearRegressionModel)
	if ok {
		return lrm.trained
	}

	mav, ok := model.(*MovingAverageModel)
	if ok {
		return mav.trained
	}

	esm, ok := model.(*ExponentialSmoothingModel)
	if ok {
		return esm.trained
	}

	sdm, ok := model.(*SeasonalDecompositionModel)
	if ok {
		return sdm.trained
	}

	return false
}

func NewDefaultLoadPredictor() *LoadPredictor {
	lp := NewLoadPredictor(1000, 1*time.Hour, 5*time.Minute)

	lp.SetModel(MetricTypeCPU, &LinearRegressionModel{})
	lp.SetModel(MetricTypeMemory, &MovingAverageModel{windowSize: 10})
	lp.SetModel(MetricTypeRequests, &ExponentialSmoothingModel{alpha: 0.3})
	lp.SetModel(MetricTypeLatency, &SeasonalDecompositionModel{period: 60})
	lp.SetModel(MetricTypeConnections, &LinearRegressionModel{})
	lp.SetModel(MetricTypeGoroutines, &MovingAverageModel{windowSize: 5})
	lp.SetModel(MetricTypeDiskIO, &ExponentialSmoothingModel{alpha: 0.2})
	lp.SetModel(MetricTypeNetworkIO, &MovingAverageModel{windowSize: 10})

	return lp
}
