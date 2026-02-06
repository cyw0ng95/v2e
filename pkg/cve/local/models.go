package local

import (
	"time"
)

// LearningStatus represents the status of a learning item
type LearningStatus string

const (
	LearningStatusNew      LearningStatus = "new"      // Newly added item
	LearningStatusLearning LearningStatus = "learning" // Currently being studied
	LearningStatusMastered LearningStatus = "mastered" // Successfully learned
	LearningStatusArchived LearningStatus = "archived" // Archived/completed
)

// LearningHistory tracks rating performance for adaptive algorithms
type LearningHistory struct {
	Ratings    []string  `json:"ratings"`    // History of ratings ("again", "hard", "good", "easy")
	Intervals  []int     `json:"intervals"`  // History of intervals used
	Timestamps []int64   `json:"timestamps"` // Unix timestamps of ratings
	Confidence float64   `json:"confidence"` // Current confidence level (0.0-1.0)
	LastReset  time.Time `json:"last_reset"` // When repetition counter was last reset
}

// LearningItemType represents the type of item being learned
type LearningItemType string

const (
	LearningItemTypeCVE    LearningItemType = "cve"
	LearningItemTypeCWE    LearningItemType = "cwe"
	LearningItemTypeCAPEC  LearningItemType = "capec"
	LearningItemTypeAttack LearningItemType = "attack"
)

// LearningItem represents a learning progress record
type LearningItem struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	ItemType     LearningItemType `gorm:"index;not null" json:"item_type"`
	ItemID       string           `gorm:"index;not null" json:"item_id"`
	Status       LearningStatus   `gorm:"index;not null;default:'new'" json:"status"`
	Progress     float64          `gorm:"not null;default:0.0" json:"progress"` // 0.0 to 1.0
	EaseFactor   float64          `gorm:"not null;default:2.5" json:"ease_factor"`
	Interval     int              `gorm:"not null;default:1" json:"interval"` // days until next review
	Repetition   int              `gorm:"not null;default:0" json:"repetition"`
	LastReviewed time.Time        `json:"last_reviewed"`
	NextReview   time.Time        `gorm:"index" json:"next_review"`
	HistoryJSON  string           `gorm:"type:text" json:"-"` // JSON-encoded LearningHistory
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (LearningItem) TableName() string {
	return "learning_items"
}
