package local

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

// LearningDB represents the learning database connection
type LearningDB struct {
	db *gorm.DB
}

// NewLearningDB creates a new learning database connection
func NewLearningDB(dbPath string) (*LearningDB, error) {
	db, err := NewOptimizedDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create learning database: %w", err)
	}

	return NewLearningDBFromGorm(db.db)
}

// NewLearningDBFromGorm creates a learning database using an existing GORM connection
// This enables sharing the connection pool with the main CVE database
func NewLearningDBFromGorm(gormDB *gorm.DB) (*LearningDB, error) {
	// Auto-migrate the learning schema
	if err := gormDB.AutoMigrate(&LearningItem{}); err != nil {
		return nil, fmt.Errorf("failed to migrate learning schema: %w", err)
	}

	return &LearningDB{db: gormDB}, nil
}

// Close closes the database connection
func (ldb *LearningDB) Close() error {
	sqlDB, err := ldb.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateLearningItem creates a new learning item
func (ldb *LearningDB) CreateLearningItem(item *LearningItem) error {
	// Set default values if not provided
	if item.EaseFactor == 0 {
		item.EaseFactor = 2.5
	}
	if item.Interval == 0 {
		item.Interval = 1
	}
	if item.Status == "" {
		item.Status = LearningStatusNew
	}

	// Set initial next review date
	if item.NextReview.IsZero() {
		item.NextReview = time.Now().AddDate(0, 0, item.Interval)
	}

	return ldb.db.Create(item).Error
}

// GetLearningItem retrieves a learning item by ID
func (ldb *LearningDB) GetLearningItem(id uint) (*LearningItem, error) {
	var item LearningItem
	if err := ldb.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetLearningItemByItem retrieves a learning item by item type and ID
func (ldb *LearningDB) GetLearningItemByItem(itemType LearningItemType, itemID string) (*LearningItem, error) {
	var item LearningItem
	if err := ldb.db.Where("item_type = ? AND item_id = ?", itemType, itemID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// ListLearningItems lists learning items with optional filtering
func (ldb *LearningDB) ListLearningItems(status LearningStatus, limit, offset int) ([]LearningItem, int64, error) {
	var items []LearningItem
	var total int64

	query := ldb.db.Model(&LearningItem{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get items with pagination
	if err := query.Limit(limit).Offset(offset).Order("next_review ASC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// UpdateLearningItem updates an existing learning item
func (ldb *LearningDB) UpdateLearningItem(item *LearningItem) error {
	return ldb.db.Save(item).Error
}

// DeleteLearningItem deletes a learning item by ID
func (ldb *LearningDB) DeleteLearningItem(id uint) error {
	return ldb.db.Delete(&LearningItem{}, id).Error
}

// RateLearningItem applies spaced repetition rating to update learning progress
func (ldb *LearningDB) RateLearningItem(id uint, rating string) error {
	item, err := ldb.GetLearningItem(id)
	if err != nil {
		return err
	}

	// Parse history for adaptive algorithm
	history, err := ldb.parseHistory(item.HistoryJSON)
	if err != nil {
		history = &LearningHistory{
			Ratings:    make([]string, 0),
			Intervals:  make([]int, 0),
			Timestamps: make([]int64, 0),
			Confidence: 0.0,
		}
	}

	// Apply adaptive spaced repetition algorithm
	ldb.applyAdaptiveRating(item, rating, history)

	// Update history
	history.Ratings = append(history.Ratings, rating)
	history.Intervals = append(history.Intervals, item.Interval)
	history.Timestamps = append(history.Timestamps, time.Now().Unix())

	// Limit history size to prevent unbounded growth
	if len(history.Ratings) > 50 {
		history.Ratings = history.Ratings[len(history.Ratings)-50:]
		history.Intervals = history.Intervals[len(history.Intervals)-50:]
		history.Timestamps = history.Timestamps[len(history.Timestamps)-50:]
	}

	// Serialize history back to JSON
	historyJSON, err := sonic.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to serialize learning history: %w", err)
	}
	item.HistoryJSON = string(historyJSON)

	return ldb.UpdateLearningItem(item)
}

// applyAdaptiveRating implements an enhanced SM-2 algorithm with adaptive parameters
func (ldb *LearningDB) applyAdaptiveRating(item *LearningItem, rating string, history *LearningHistory) {
	// Calculate performance trend for dynamic adjustment
	trend := ldb.calculatePerformanceTrend(history)

	switch rating {
	case "again": // Failed to recall
		item.Repetition = 0
		item.Interval = 1
		// More aggressive ease factor reduction for poor performance
		easeAdjustment := 0.2 + (1.0-trend)*0.1
		item.EaseFactor = floatMax(1.3, item.EaseFactor-easeAdjustment)
		history.Confidence = floatMax(0.0, history.Confidence-0.3)
		item.Status = LearningStatusLearning
	case "hard": // Recalled with difficulty
		if item.Repetition == 0 {
			item.Interval = 1
		} else {
			// Adjust interval based on recent performance
			baseMultiplier := 1.2 + (trend-0.5)*0.3
			item.Interval = int(float64(item.Interval) * floatMax(1.0, baseMultiplier))
		}
		easeAdjustment := 0.15 + (1.0-trend)*0.1
		item.EaseFactor = floatMax(1.3, item.EaseFactor-easeAdjustment)
		history.Confidence = floatMax(0.0, history.Confidence-0.1)
		item.Status = LearningStatusLearning
	case "good": // Recalled correctly
		switch item.Repetition {
		case 0:
			item.Interval = 1
		case 1:
			item.Interval = 6
		default:
			// Dynamic interval calculation based on ease factor and trend
			trendMultiplier := 0.8 + trend*0.4
			item.Interval = int(float64(item.Interval) * item.EaseFactor * trendMultiplier)
		}
		item.Repetition++
		// Adaptive ease factor adjustment
		easeAdjustment := 0.1 + trend*0.05
		item.EaseFactor += easeAdjustment
		history.Confidence = floatMin(1.0, history.Confidence+0.15)
		item.Status = LearningStatusLearning
	case "easy": // Recalled easily
		switch item.Repetition {
		case 0:
			item.Interval = 1
		case 1:
			item.Interval = 6
		default:
			// Boost interval for easy recall with confidence bonus
			confidenceBonus := 1.0 + history.Confidence*0.3
			trendMultiplier := 1.0 + trend*0.5
			item.Interval = int(float64(item.Interval) * item.EaseFactor * 1.3 * confidenceBonus * trendMultiplier)
		}
		item.Repetition++
		easeAdjustment := 0.15 + trend*0.05
		item.EaseFactor += easeAdjustment
		history.Confidence = floatMin(1.0, history.Confidence+0.25)
		item.Status = LearningStatusMastered
	}

	// Update timestamps
	item.LastReviewed = time.Now()
	item.NextReview = time.Now().AddDate(0, 0, item.Interval)

	// Enhanced progress calculation using confidence and repetition
	baseProgress := float64(item.Repetition) / 10.0
	confidenceBoost := history.Confidence * 0.3
	item.Progress = floatMin(1.0, baseProgress+confidenceBoost)

	// Cap ease factor with adaptive bounds
	minEase := floatMax(1.3, 1.8-history.Confidence*0.5)
	maxEase := floatMin(3.0, 2.5+history.Confidence*0.5)
	item.EaseFactor = floatMin(maxEase, floatMax(minEase, item.EaseFactor))
}

// calculatePerformanceTrend analyzes recent rating history to determine performance trend
func (ldb *LearningDB) calculatePerformanceTrend(history *LearningHistory) float64 {
	if len(history.Ratings) < 3 {
		return 0.5 // Neutral trend for insufficient data
	}

	// Look at last 10 ratings (or all if fewer)
	recentCount := len(history.Ratings)
	if recentCount > 10 {
		recentCount = 10
	}

	startIndex := len(history.Ratings) - recentCount
	recentRatings := history.Ratings[startIndex:]

	// Convert ratings to numerical scores
	scores := make([]float64, len(recentRatings))
	for i, rating := range recentRatings {
		switch rating {
		case "again":
			scores[i] = 0.0
		case "hard":
			scores[i] = 0.3
		case "good":
			scores[i] = 0.7
		case "easy":
			scores[i] = 1.0
		}
	}

	// Calculate weighted average (more recent ratings weighted higher)
	var sum, weightSum float64
	for i, score := range scores {
		weight := float64(i + 1) // Linear weighting
		sum += score * weight
		weightSum += weight
	}

	if weightSum == 0 {
		return 0.5
	}

	return sum / weightSum
}

// parseHistory deserializes JSON history or returns empty history on error
func (ldb *LearningDB) parseHistory(historyJSON string) (*LearningHistory, error) {
	if historyJSON == "" {
		return &LearningHistory{
			Ratings:    make([]string, 0),
			Intervals:  make([]int, 0),
			Timestamps: make([]int64, 0),
			Confidence: 0.0,
		}, nil
	}

	var history LearningHistory
	if err := sonic.Unmarshal([]byte(historyJSON), &history); err != nil {
		// Return empty history on parse error
		return &LearningHistory{
			Ratings:    make([]string, 0),
			Intervals:  make([]int, 0),
			Timestamps: make([]int64, 0),
			Confidence: 0.0,
		}, nil
	}

	return &history, nil
}

// GetItemsDueForReview gets items that are due for review (next_review <= now)
func (ldb *LearningDB) GetItemsDueForReview(limit int) ([]LearningItem, error) {
	var items []LearningItem
	err := ldb.db.Where("next_review <= ? AND status IN (?, ?)",
		time.Now(), LearningStatusNew, LearningStatusLearning).
		Order("next_review ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

// GetLearningStats returns statistics about learning progress
func (ldb *LearningDB) GetLearningStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	var count int64

	// New items
	ldb.db.Model(&LearningItem{}).Where("status = ?", LearningStatusNew).Count(&count)
	stats["new"] = count

	// Learning items
	ldb.db.Model(&LearningItem{}).Where("status = ?", LearningStatusLearning).Count(&count)
	stats["learning"] = count

	// Mastered items
	ldb.db.Model(&LearningItem{}).Where("status = ?", LearningStatusMastered).Count(&count)
	stats["mastered"] = count

	// Archived items
	ldb.db.Model(&LearningItem{}).Where("status = ?", LearningStatusArchived).Count(&count)
	stats["archived"] = count

	// Items due for review
	ldb.db.Model(&LearningItem{}).Where("next_review <= ? AND status IN (?, ?)",
		time.Now(), LearningStatusNew, LearningStatusLearning).Count(&count)
	stats["due_for_review"] = count

	return stats, nil
}

// RateLearningItems applies ratings to multiple learning items in a single transaction
func (ldb *LearningDB) RateLearningItems(ids []uint, ratings []string) error {
	if len(ids) != len(ratings) {
		return fmt.Errorf("ids and ratings slices must have equal length")
	}

	tx := ldb.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, id := range ids {
		item, err := ldb.GetLearningItemWithDB(tx, id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get item %d: %w", id, err)
		}

		// Parse history
		history, err := ldb.parseHistory(item.HistoryJSON)
		if err != nil {
			history = &LearningHistory{
				Ratings:    make([]string, 0),
				Intervals:  make([]int, 0),
				Timestamps: make([]int64, 0),
				Confidence: 0.0,
			}
		}

		// Apply rating
		ldb.applyAdaptiveRating(item, ratings[i], history)

		// Update history
		history.Ratings = append(history.Ratings, ratings[i])
		history.Intervals = append(history.Intervals, item.Interval)
		history.Timestamps = append(history.Timestamps, time.Now().Unix())

		// Limit history size
		if len(history.Ratings) > 50 {
			history.Ratings = history.Ratings[len(history.Ratings)-50:]
			history.Intervals = history.Intervals[len(history.Intervals)-50:]
			history.Timestamps = history.Timestamps[len(history.Timestamps)-50:]
		}

		// Serialize history
		historyJSON, err := sonic.Marshal(history)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to serialize history for item %d: %w", id, err)
		}
		item.HistoryJSON = string(historyJSON)

		// Save updated item
		if err := tx.Save(item).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save item %d: %w", id, err)
		}
	}

	return tx.Commit().Error
}

// GetLearningItemWithDB retrieves a learning item using a specific database connection
func (ldb *LearningDB) GetLearningItemWithDB(db *gorm.DB, id uint) (*LearningItem, error) {
	var item LearningItem
	if err := db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetLearningTrends returns performance trends over time
func (ldb *LearningDB) GetLearningTrends(days int) (map[string]interface{}, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	var items []LearningItem
	if err := ldb.db.Where("last_reviewed >= ?", cutoff).Find(&items).Error; err != nil {
		return nil, err
	}

	trends := make(map[string]interface{})
	totalReviews := len(items)
	trends["total_reviews"] = totalReviews

	if totalReviews == 0 {
		trends["average_confidence"] = 0.0
		trends["success_rate"] = 0.0
		return trends, nil
	}

	var totalConfidence, successCount float64
	ratingCounts := make(map[string]int)

	for _, item := range items {
		history, err := ldb.parseHistory(item.HistoryJSON)
		if err != nil || len(history.Ratings) == 0 {
			continue
		}

		// Get last rating
		lastRating := history.Ratings[len(history.Ratings)-1]
		ratingCounts[lastRating]++

		// Count successful ratings
		if lastRating == "good" || lastRating == "easy" {
			successCount++
		}

		totalConfidence += history.Confidence
	}

	trends["average_confidence"] = totalConfidence / float64(totalReviews)
	trends["success_rate"] = successCount / float64(totalReviews)
	trends["rating_distribution"] = ratingCounts

	return trends, nil
}

// PredictMasteryTimeline predicts when an item will reach mastery
func (ldb *LearningDB) PredictMasteryTimeline(itemType LearningItemType, itemID string) (map[string]interface{}, error) {
	item, err := ldb.GetLearningItemByItem(itemType, itemID)
	if err != nil {
		return nil, err
	}

	prediction := make(map[string]interface{})
	history, err := ldb.parseHistory(item.HistoryJSON)
	if err != nil {
		history = &LearningHistory{Confidence: 0.0}
	}

	// Simple prediction model based on current trend
	trend := ldb.calculatePerformanceTrend(history)

	// Estimate days until mastery (confidence >= 0.8)
	confidenceNeeded := 0.8 - history.Confidence
	if confidenceNeeded <= 0 {
		prediction["days_to_mastery"] = 0
		prediction["already_mastered"] = true
	} else {
		// Rough estimate: assuming 0.1 confidence gain per successful review
		estimatedDays := int(confidenceNeeded / (0.1 * trend))
		prediction["days_to_mastery"] = estimatedDays
		prediction["already_mastered"] = false
	}

	prediction["current_confidence"] = history.Confidence
	prediction["performance_trend"] = trend

	return prediction, nil
}

// Helper functions for float64
func floatMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func floatMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
