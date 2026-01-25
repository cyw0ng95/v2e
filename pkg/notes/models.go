package notes

import (
	"time"

	"gorm.io/gorm"
)

// BookmarkModel stores bookmarked items with type, global ID, title, learning state, and metadata
type BookmarkModel struct {
	ID          uint           `gorm:"primaryKey"`
	GlobalItemID string        `gorm:"index;not null"` // Links to GlobalItemModel
	ItemType    string        `gorm:"index;not null"` // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	ItemID      string        `gorm:"not null"`       // Original ID from source (e.g., CVE-2021-1234)
	Title       string        `gorm:"not null"`
	Description string        
	CreatedAt   time.Time     
	UpdatedAt   time.Time     
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Learning state fields
	LearningState string `gorm:"default:'to-review'"` // to-review, learning, mastered, archived
	LastReviewed  *time.Time
	NextReview    *time.Time
	MasteryLevel  float32 `gorm:"default:0.0"` // 0.0 to 1.0

	// Relationships
	Notes []NoteModel `gorm:"foreignKey:BookmarkID"`
	History []BookmarkHistoryModel `gorm:"foreignKey:BookmarkID"`
}

// BookmarkHistoryModel tracks state changes over time
type BookmarkHistoryModel struct {
	ID          uint   `gorm:"primaryKey"`
	BookmarkID  uint   `gorm:"index;not null"`
	Action      string `gorm:"not null"` // e.g., "created", "updated", "learning_state_changed", "note_added", "deleted"
	OldValue    string // Previous state/value
	NewValue    string // New state/value
	Timestamp   time.Time
	UserID      *string // Optional user identifier if multi-user support is added later
}

// NoteModel stores user notes associated with bookmarks
type NoteModel struct {
	ID         uint   `gorm:"primaryKey"`
	BookmarkID uint   `gorm:"index;not null"`
	Content    string `gorm:"type:text;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Author     *string // Optional author field
	IsPrivate  bool    `gorm:"default:false"` // Whether the note is private to the user
}

// MemoryCardModel stores card-specific learning data
type MemoryCardModel struct {
	ID          uint   `gorm:"primaryKey"`
	BookmarkID  uint   `gorm:"index;not null"`
	Front       string `gorm:"type:text;not null"` // Question/content on front of card
	Back        string `gorm:"type:text;not null"` // Answer/explanation on back of card
	EaseFactor  float32 `gorm:"default:2.5"`       // For spaced repetition algorithm
	Interval    int     `gorm:"default:1"`         // Days until next review
	Repetition  int     `gorm:"default:0"`         // Number of successful reviews
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// LearningSessionModel tracks learning sessions and progress
type LearningSessionModel struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        *string   // Optional user identifier
	SessionStart  time.Time
	SessionEnd    *time.Time
	CardsReviewed int       // Number of cards reviewed in this session
	CardsCorrect  int       // Number of cards answered correctly
	SessionNotes  *string   // Optional notes about the session
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CrossReferenceModel stores relationships between items from different databases
type CrossReferenceModel struct {
	ID              uint   `gorm:"primaryKey"`
	SourceItemID    string `gorm:"index;not null"` // Global ID of source item
	TargetItemID    string `gorm:"index;not null"` // Global ID of target item
	SourceType      string `gorm:"not null"`       // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	TargetType      string `gorm:"not null"`       // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	RelationshipType string `gorm:"not null"`      // e.g., "related-to", "exploits", "mitigates", "similar-to"
	Strength        float32 `gorm:"default:1.0"`   // Confidence/strength of relationship (0.0 to 1.0)
	Description     *string                         // Optional description of the relationship
	CreatedAt       time.Time
}

// GlobalItemModel stores unified identifiers for items across sources
type GlobalItemModel struct {
	ID         string `gorm:"primaryKey"` // Globally unique identifier
	ItemType   string `gorm:"index;not null"` // e.g., "CVE", "CWE", "CAPEC", "ATT&CK"
	SourceID   string `gorm:"not null"`       // Original ID from source (e.g., CVE-2021-1234)
	Title      string `gorm:"not null"`
	Source     string `gorm:"not null"` // e.g., "NVD", "MITRE", "CAPEC", "MITRE_ATT&CK"
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Description *string
	// Relationships
	Bookmarks []BookmarkModel `gorm:"foreignKey:GlobalItemID"`
}